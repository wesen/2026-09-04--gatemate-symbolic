"""Two-pass assembler for the Laboratory 1 tagged stack evaluator.

Produces build/<name>.hex (one 20-bit word per line, 5 lowercase hex digits,
zero-padded to ROM depth), .lst (listing), and .sym.json (label map). No
`eval` anywhere (MATE-16 discipline).

Syntax:
    ; comment to end of line
    label:              optional, defines the current address
    MNEMONIC            opcode with no immediate (ADD, HALT, ...)
    PUSH_S15 k          signed 15-bit immediate (-16384..16383)
    JMP t / JZ t        branch target: label, or absolute integer
    WORD 0xF8000        raw 20-bit word (undefined opcode)
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from typing import Dict, List, Optional, Tuple

from opcodes import BY_MNEMONIC, INSTRUCTIONS, encode

_TOKEN = re.compile(r"^\s*(?:(\w+):)?\s*(?:(\w+)(?:\s+([^\s;]+))?)?\s*(?:;.*)?$")


def _parse_line(line: str, path: str, lineno: int
                ) -> Tuple[Optional[str], Optional[str], Optional[str]]:
    """Return (label, mnemonic, operand) — any may be None."""
    if not line.strip() or line.lstrip().startswith(';'):
        return None, None, None
    m = _TOKEN.fullmatch(line)
    if not m:
        raise ValueError(f"{path}: invalid syntax at line {lineno}: {line!r}")
    return m.group(1), m.group(2).upper() if m.group(2) else None, m.group(3)


def _parse_int(text: str) -> int:
    t = text.strip()
    if t.lower().startswith("0x"):
        return int(t, 16)
    if t.startswith("-") or t.isdigit():
        return int(t, 10)
    raise ValueError(f"bad integer literal: {text!r}")


def assemble(source: str, rom_depth: int = 1024
             ) -> Tuple[List[int], Dict[str, int], List[str]]:
    """Assemble source text -> (words, symbols, listing lines)."""
    if not 2 <= rom_depth <= 32768:
        raise ValueError("ROM depth must be between 2 and 32768")
    # ---------------- pass 1: sizes and labels ----------------
    words: List[Optional[int]] = []
    symbols: Dict[str, int] = {}
    listing: List[str] = []
    pending: List[Tuple[int, int, str, str, str, int]] = []

    for lineno, raw in enumerate(source.splitlines(), 1):
        label, mnem, operand = _parse_line(raw, "<src>", lineno)
        if label is None and mnem is None:
            listing.append(f"{len(words):04d}                ; {raw}")
            continue
        if label:
            if label in symbols:
                raise ValueError(f"duplicate label {label!r} (line {lineno})")
            symbols[label] = len(words)

        if mnem is None:
            listing.append(f"{len(words):04d}                ; {raw}")
            continue

        # Every emitting path, including WORD and unresolved immediates.
        if len(words) >= rom_depth:
            raise ValueError(f"program exceeds ROM depth {rom_depth} (line {lineno})")

        if mnem == "WORD":
            if operand is None:
                raise ValueError(f"WORD needs an operand (line {lineno})")
            w = _parse_int(operand)
            if not (0 <= w <= 0xFFFFF):
                raise ValueError(f"WORD out of 20-bit range (line {lineno})")
            words.append(w)
            listing.append(f"{len(words)-1:04d} {w:05x}            WORD {operand}")
            continue

        ins = BY_MNEMONIC.get(mnem)
        if ins is None:
            raise ValueError(f"unknown mnemonic {mnem!r} (line {lineno})")

        if ins.imm_kind == "none":
            if operand is not None:
                raise ValueError(f"{mnem} takes no operand (line {lineno})")
            w = encode(ins.opcode, 0)
            words.append(w)
            listing.append(f"{len(words)-1:04d} {w:05x}            {mnem}")
        else:
            if operand is None:
                raise ValueError(f"{mnem} needs an operand (line {lineno})")
            # Record both the word index and the position of the placeholder
            # line in `listing` -- comment/label-only lines make the two
            # diverge, so pass 2 must patch the recorded line, not [index].
            pending.append((len(words), len(listing), mnem, operand,
                            "<src>", lineno))
            words.append(None)  # placeholder
            listing.append(f"{len(words)-1:04d} ?????           {mnem} {operand}")

    # ---------------- pass 2: resolve immediates ----------------
    for index, listing_pos, mnem, operand, path, lineno in pending:
        ins = BY_MNEMONIC[mnem]
        if operand in symbols:
            value = symbols[operand]
        else:
            value = _parse_int(operand)
        if ins.imm_kind == "s15":
            if not (-(2**14) <= value <= 2**14 - 1):
                raise ValueError(
                    f"immediate {value} out of signed 15-bit range "
                    f"(line {lineno})")
        else:  # u15 branch target
            if not (0 <= value <= 2**15 - 1):
                raise ValueError(
                    f"branch target {value} out of 15-bit range "
                    f"(line {lineno})")
            if value >= rom_depth:
                raise ValueError(
                    f"branch target {value} outside ROM depth {rom_depth} "
                    f"(line {lineno})")
        w = encode(ins.opcode, value)
        words[index] = w
        listing[listing_pos] = (f"{index:04d} {w:05x}            "
                               f"{mnem} {operand} (= {value})")

    # Only the CLI hex writer pads the returned instruction words.
    assert all(w is not None for w in words)
    return [int(w) for w in words], symbols, listing


def main(argv: Optional[List[str]] = None) -> int:
    ap = argparse.ArgumentParser(description="Laboratory 1 assembler")
    ap.add_argument("source")
    ap.add_argument("-o", "--out-dir", default="build")
    ap.add_argument("-n", "--name", default=None)
    ap.add_argument("-s", "--rom-depth", type=int, default=1024)
    args = ap.parse_args(argv)

    name = args.name or args.source.rsplit("/", 1)[-1].removesuffix(".asm")
    with open(args.source) as f:
        source = f.read()

    words, symbols, listing = assemble(source, args.rom_depth)

    import os
    os.makedirs(args.out_dir, exist_ok=True)
    hex_path = os.path.join(args.out_dir, f"{name}.hex")
    with open(hex_path, "w") as f:
        for w in words:
            f.write(f"{w:05x}\n")
        for _ in range(args.rom_depth - len(words)):
            f.write("00000\n")

    with open(os.path.join(args.out_dir, f"{name}.lst"), "w") as f:
        f.write("\n".join(listing) + "\n")
    with open(os.path.join(args.out_dir, f"{name}.sym.json"), "w") as f:
        json.dump({"symbols": symbols, "rom_depth": args.rom_depth,
                  "words": len(words)}, f, indent=2)

    print(f"{args.source}: {len(words)} words -> {hex_path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
