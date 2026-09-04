"""Differential RTL tests for the BRAM core (two-entry top cache over
sync_sdp_ram). Same programs as the register core, plus the refinement
invariant (depth == tc + dc, checked in the testbench) and random
legal/illegal program generation (book Lab 1 "Verification plan", random
test generation) compared model-vs-RTL with random output backpressure.
"""

import os
import random
import subprocess
import sys

import pytest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "tools"))

from opcodes import INSTRUCTIONS, encode  # noqa: E402
from stack_model import run_program  # noqa: E402
from test_directed import PROGRAMS, _ensure_hex  # noqa: E402

ROOT = os.path.join(os.path.dirname(__file__), "..")
BUILD = os.path.join(ROOT, "build")

# BRAM core configs: (deep_depth, model total_depth, stall_seeds)
BRAM_PROGRAMS = [(name, deep, seeds)
                 for (name, _d, seeds) in PROGRAMS
                 for deep in ([30] if name != "stackoverflow" else [6])]


def _run_rtl_bram(hex_path, deep_depth, stall_seed=None):
    tag = f"tb_bram_d{deep_depth}"
    vvp = os.path.join(BUILD, f"{tag}.vvp")
    subprocess.run([
        "iverilog", "-g2012", "-s", "tb_stack_core_bram", "-o", vvp,
        os.path.join(ROOT, "rtl", "symbolic_types_pkg.sv"),
        os.path.join(ROOT, "rtl", "sync_sdp_ram.sv"),
        os.path.join(ROOT, "rtl", "stack_core_bram.sv"),
        os.path.join(ROOT, "sim", "tb_stack_core_bram.sv"),
        f"-Ptb_stack_core_bram.DEEP_DEPTH={deep_depth}",
    ], check=True, capture_output=True)
    cmd = ["vvp", vvp, f"+rom={hex_path}"]
    if stall_seed is not None:
        cmd.append(f"+stall_seed={stall_seed}")
    out = subprocess.run(cmd, check=True, capture_output=True, text=True,
                         timeout=120)
    lines = [l for l in out.stdout.splitlines()
             if l.startswith(("TRACE", "FINAL", "TB_", "FAIL",
                             "ASSERT_FAIL"))]
    assert any(l.startswith("TB_PASS") for l in lines), out.stdout
    return [l for l in lines if l.startswith(("TRACE", "FINAL"))]


def _model_lines_for(words, total_depth):
    m = run_program(words, total_depth=total_depth)
    return [r.line() for r in m.trace] + [m.final_line()]


def _words_of(name):
    from asm20 import assemble
    with open(os.path.join(ROOT, "programs", f"{name}.asm")) as f:
        words, _, _ = assemble(f.read())
    return words


@pytest.mark.parametrize("name,deep,seeds", BRAM_PROGRAMS)
def test_directed_bram(name, deep, seeds):
    hex_path = _ensure_hex(name)
    total = deep + 2
    expected = _model_lines_for(_words_of(name), total)
    got = _run_rtl_bram(hex_path, deep)
    assert got == expected, _diff(got, expected)
    for seed in seeds:
        got_s = _run_rtl_bram(hex_path, deep, stall_seed=seed)
        assert got_s == expected, f"stall_seed={seed} changed the trace"


def _diff(got, expected):
    import difflib
    return "\n".join(difflib.unified_diff(
        expected, got, "model", "rtl", lineterm=""))


# ---------------------------------------------------------------------------
# Random program generation (book: maintain a typed software stack, choose
# only legal instructions, occasionally inject one illegal instruction).
# ---------------------------------------------------------------------------

PUSH_OPS = [0x00, 0x01, 0x02]
BIN_OPS = [0x03, 0x04, 0x05, 0x06, 0x07]
INT_TAG, BOOL_TAG = 0, 1


def gen_program(rng, length=150, total_depth=8, illegal_p=0.06):
    """Generate a random program (word list) plus a note about legality.

    The generator tracks a typed stack model so chosen instructions are
    legal by construction; with probability illegal_p one deliberately
    illegal instruction is injected (type error or underflow or bad opcode).
    """
    words = []
    stack = []          # list of tags
    rom_depth = 1024

    def depth():
        return len(stack)

    while len(words) < length:
        # occasionally inject one illegal instruction
        if rng.random() < illegal_p:
            kind = rng.randrange(3)
            if kind == 0 and depth() >= 2:
                words.append(encode(0x03))          # ADD on maybe-non-INTs
                stack.append(INT_TAG)
                # model will fault if tags wrong; both sides see the same
                # -> keep generator's view simple: fault ends the program
                return words, stack
            elif kind == 1:
                words.append(encode(0x09))          # DROP (maybe underflow)
                if depth() >= 1:
                    stack.pop()
                return words, stack
            else:
                words.append(encode(0x1F, 0))        # BAD_OPCODE
                return words, stack

        choices = ["push"]
        if depth() >= 2 and stack[-1] == INT_TAG and stack[-2] == INT_TAG:
            choices += ["bin"] * 4
        if depth() >= 1:
            choices += ["dup", "drop", "swap", "jz" if stack[-1] == BOOL_TAG
                        else "nop"] * 2
        choices += ["emit"] if depth() >= 1 else []
        kind = rng.choice(choices)

        if kind == "push":
            if depth() >= total_depth:
                continue
            if rng.random() < 0.6:
                words.append(encode(0x00, rng.randrange(-20, 21)))
                stack.append(INT_TAG)
            else:
                words.append(encode(rng.choice([0x01, 0x02])))
                stack.append(BOOL_TAG)
        elif kind == "bin":
            op = rng.choice(BIN_OPS)
            words.append(encode(op))
            stack.pop(); stack.pop()
            stack.append(INT_TAG if op != 0x06 else BOOL_TAG)
            # note: EQ yields BOOL regardless of operand tags
            if op == 0x06:
                stack[-1] = BOOL_TAG
        elif kind == "dup":
            if depth() >= total_depth:
                continue
            words.append(encode(0x08))
            stack.append(stack[-1])
        elif kind == "drop":
            words.append(encode(0x09))
            stack.pop()
        elif kind == "swap":
            if depth() < 2:
                continue
            words.append(encode(0x0A))
            stack[-1], stack[-2] = stack[-2], stack[-1]
        elif kind == "jz":
            # branch forward past the next instruction (target = here+2)
            here = len(words)
            target = here + 2
            if target >= rom_depth:
                continue
            words.append(encode(0x0C, target))
            stack.pop()
            # slot after JZ: an unreachable HALT placeholder? no - JZ only
            # branches when false; we cannot statically know, so make the
            # next instruction a HALT-free NOP that is safe either way:
            # use PUSH_FALSE then... simpler: emit PUSH_TRUE next so the
            # stack stays balanced whichever way? No: on the taken path the
            # pushed value stays. Instead, make the next word a HALT and
            # stop generating (both paths end).
            words.append(encode(0x0E))
            return words, stack
        elif kind == "emit":
            words.append(encode(0x0D))
            stack.pop()
        elif kind == "nop":
            # top is INT and JZ would fault: emit a compare to make a BOOL
            words.append(encode(0x00, 0))
            if depth() + 1 > total_depth:
                continue
            stack.append(INT_TAG)

    words.append(encode(0x0E))
    return words, stack


@pytest.mark.parametrize("seed", range(24))
def test_random_programs(seed):
    rng = random.Random(seed)
    for total_depth, deep in ((8, 6), (12, 10)):
        words, _ = gen_program(rng, length=120, total_depth=total_depth)
        hex_path = os.path.join(BUILD, f"rand_{seed}_{total_depth}.hex")
        os.makedirs(BUILD, exist_ok=True)
        with open(hex_path, "w") as f:
            for w in words:
                f.write(f"{w:05x}\n")
            for _ in range(1024 - len(words)):
                f.write("00000\n")
        expected = _model_lines_for(words, total_depth)
        got = _run_rtl_bram(hex_path, deep, stall_seed=seed)
        assert got == expected, _diff(got, expected)


@pytest.mark.parametrize("seed", range(12))
def test_random_programs_register_core(seed):
    """The register core must pass the same random programs (STACK_DEPTH
    equals the model total_depth)."""
    from test_directed import _run_rtl as _run_reg
    rng = random.Random(1000 + seed)
    words, _ = gen_program(rng, length=120, total_depth=8)
    hex_path = os.path.join(BUILD, f"randreg_{seed}.hex")
    os.makedirs(BUILD, exist_ok=True)
    with open(hex_path, "w") as f:
        for w in words:
            f.write(f"{w:05x}\n")
        for _ in range(1024 - len(words)):
            f.write("00000\n")
    expected = _model_lines_for(words, 8)
    got = _run_reg(hex_path, 8, stall_seed=seed)
    assert got == expected, _diff(got, expected)
