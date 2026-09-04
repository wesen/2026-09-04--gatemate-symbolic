"""Executable reference model for the Laboratory 1 tagged stack evaluator.

The model is the abstract machine

    M = <pc, stack, output_stream, fault, halted>

from the book (Laboratory 1, "Semantic machine"). It implements every opcode,
every fault case, and the retirement discipline, and it produces the same
commit-trace records the RTL testbenches print, so differential testing is a
line-by-line comparison.

Semantics decisions (design doc decision records):

- DR-2: ADD/SUB/MUL use *precise* signed 32-bit overflow -> ARITH_OVERFLOW.
- DR-3: EQ yields BOOL(true) iff tag, flags, and payload all match; mixed tags
  yield BOOL(false) (not a fault).
- DR-4: reset is a global experiment abort; the model simply starts from the
  initial image.
- HALT: pc stays at the HALT instruction (stack effect "unchanged"); the event
  is a normal COMMIT and the machine is halted afterwards.
- Check order per instruction (all checks precede any mutation):
  underflow/overflow capacity -> operand tags -> canonicality -> branch target.
- JZ pops the condition only when all checks pass; a non-BOOL operand is
  TYPE_FAULT, a BOOL with payload not in {0,1} is NONCANONICAL_BOOL.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import List, Optional, Tuple

from opcodes import (
    BY_MNEMONIC,
    EVENT_COMMIT,
    EVENT_FAULT,
    EVENT_NAMES,
    EVENT_OUTPUT,
    Fault,
    INSTRUCTIONS,
    Tag,
    decode,
    sign_extend_s15,
)


# ---------------------------------------------------------------------------
# value40
# ---------------------------------------------------------------------------

@dataclass(frozen=True)
class Value40:
    """One tagged word: tag[3:0], flags[3:0], payload[31:0]."""

    tag: int
    payload: int
    flags: int = 0

    def __post_init__(self) -> None:
        assert 0 <= self.tag <= 0xF and 0 <= self.flags <= 0xF
        assert 0 <= self.payload <= 0xFFFFFFFF

    def word(self) -> int:
        """40-bit packed representation (matches RTL value40_t bit order)."""
        return (self.tag << 36) | (self.flags << 32) | self.payload

    def text(self) -> str:
        """Canonical trace text, e.g. INT(7), BOOL(1)."""
        return f"{Tag(self.tag).name}({self.payload})"

    @staticmethod
    def from_word(word: int) -> "Value40":
        return Value40(
            tag=(word >> 36) & 0xF,
            flags=(word >> 32) & 0xF,
            payload=word & 0xFFFFFFFF,
        )


def make_int(x: int) -> Value40:
    if not (-(2**31) <= x <= 2**31 - 1):
        raise ValueError(f"INT payload out of signed 32-bit range: {x}")
    # two's-complement payload bits
    return Value40(tag=int(Tag.INT), payload=x & 0xFFFFFFFF)


def make_bool(b: bool) -> Value40:
    return Value40(tag=int(Tag.BOOL), payload=1 if b else 0)


def _s32(v: Value40) -> int:
    """Signed interpretation of an INT payload."""
    p = v.payload
    return p - (1 << 32) if p & (1 << 31) else p


# ---------------------------------------------------------------------------
# Commit trace record
# ---------------------------------------------------------------------------

@dataclass
class TraceRecord:
    seq: int
    pc_old: int
    pc_new: int
    op: str
    depth: int
    event: int

    # Extra fields for FAULT / OUTPUT records (None otherwise).
    fault_code: Optional[Fault] = None
    fault_tag1: int = 0  # tag of NOS operand at the faulting instruction
    fault_tag0: int = 0  # tag of TOS operand at the faulting instruction
    out_value: Optional[Value40] = None

    def line(self) -> str:
        """Canonical trace line, identical format to the RTL testbench."""
        s = (f"TRACE {self.seq} {self.pc_old} {self.pc_new} "
             f"{self.op} {self.depth} {EVENT_NAMES[self.event]}")
        if self.event == EVENT_FAULT:
            s += (f" {self.fault_code.name} {self.fault_tag1:x} "
                  f"{self.fault_tag0:x}")
        if self.event == EVENT_OUTPUT:
            s += f" {self.out_value.tag:x} {self.out_value.payload:08x}"
        return s


# ---------------------------------------------------------------------------
# The machine
# ---------------------------------------------------------------------------

@dataclass
class Machine:
    """Reference machine. Direct construction of arbitrary states is allowed
    (tests use it to reach otherwise unreachable fault cases)."""

    program: List[int]                      # 20-bit words
    rom_depth: int = 1024                  # BAD_BRANCH_TARGET threshold
    total_depth: int = 32                  # STACK_OVERFLOW threshold

    pc: int = 0
    stack: List[Value40] = field(default_factory=list)  # [-1] is top0
    output: List[Value40] = field(default_factory=list)
    fault: Optional[Fault] = None
    fault_pc: int = 0
    fault_op: str = ""
    fault_tag1: int = 0
    fault_tag0: int = 0
    halted: bool = False

    seq: int = 0
    trace: List[TraceRecord] = field(default_factory=list)

    # ------------------------------------------------------------------ utils

    @property
    def depth(self) -> int:
        return len(self.stack)

    def _top0(self) -> Value40:
        return self.stack[-1]

    def _top1(self) -> Value40:
        return self.stack[-2]

    def _fault(self, code: Fault, op: str, tag1: int = None,
               tag0: int = None) -> TraceRecord:
        """Enter the precise fault state: nothing about the architectural
        state changes; pc stays at the faulting instruction.

        Operand tags default to the current top-two stack tags (0 when
        absent) so BAD_OPCODE / underflow records still carry the machine's
        operand context - the RTL does the same (guarded by depth)."""
        if tag1 is None:
            tag1 = self.stack[-2].tag if len(self.stack) >= 2 else 0
        if tag0 is None:
            tag0 = self.stack[-1].tag if len(self.stack) >= 1 else 0
        self.fault = code
        self.fault_pc = self.pc
        self.fault_op = op
        self.fault_tag1 = tag1
        self.fault_tag0 = tag0
        rec = TraceRecord(self.seq, self.pc, self.pc, op, self.depth,
                          EVENT_FAULT, fault_code=code,
                          fault_tag1=tag1, fault_tag0=tag0)
        self.seq += 1
        self.trace.append(rec)
        return rec

    # ------------------------------------------------------------------ step

    def step(self) -> Optional[TraceRecord]:
        """One abstract transition. Returns the trace record, or None when
        the machine is halted or faulted (no further transitions)."""
        if self.halted or self.fault is not None:
            return None

        if self.pc >= self.rom_depth:
            return self._fault(Fault.BAD_BRANCH_TARGET, "FETCH")

        word = self.program[self.pc] if self.pc < len(self.program) else 0
        opcode, imm = decode(word)
        ins = INSTRUCTIONS.get(opcode)
        if ins is None:
            return self._fault(Fault.BAD_OPCODE, "BAD")
        op = ins.mnemonic

        # ---- capacity checks (before any mutation) ------------------------
        if ins.stack_in > self.depth:
            return self._fault(Fault.STACK_UNDERFLOW, op)
        grows = ins.stack_out - ins.stack_in
        if grows > 0 and self.depth + grows > self.total_depth:
            return self._fault(Fault.STACK_OVERFLOW, op)

        # ---- per-opcode validation + transition ---------------------------
        if op == "PUSH_S15":
            return self._commit(op, [make_int(sign_extend_s15(imm))],
                                self.pc + 1)
        if op == "PUSH_TRUE":
            return self._commit(op, [make_bool(True)], self.pc + 1)
        if op == "PUSH_FALSE":
            return self._commit(op, [make_bool(False)], self.pc + 1)

        if op in ("ADD", "SUB", "MUL", "LT"):
            a, b = self._top1(), self._top0()
            if a.tag != int(Tag.INT) or b.tag != int(Tag.INT):
                return self._fault(Fault.TYPE_FAULT, op,
                                   tag1=a.tag, tag0=b.tag)
            sa, sb = _s32(a), _s32(b)
            if op == "ADD":
                r = sa + sb
            elif op == "SUB":
                r = sa - sb
            elif op == "MUL":
                r = sa * sb
            else:  # LT
                return self._commit(op, [make_bool(sa < sb)], self.pc + 1)
            if not (-(2**31) <= r <= 2**31 - 1):
                return self._fault(Fault.ARITH_OVERFLOW, op,
                                   tag1=a.tag, tag0=b.tag)
            return self._commit(op, [make_int(r)], self.pc + 1)

        if op == "EQ":
            a, b = self._top1(), self._top0()
            eq = (a.tag == b.tag and a.flags == b.flags
                  and a.payload == b.payload)
            return self._commit(op, [make_bool(eq)], self.pc + 1)

        if op == "DUP":
            return self._commit(op, [self._top0(), self._top0()], self.pc + 1)
        if op == "DROP":
            return self._commit(op, [], self.pc + 1)
        if op == "SWAP":
            return self._commit(op, [self._top0(), self._top1()], self.pc + 1)

        if op == "JMP":
            if imm >= self.rom_depth:
                return self._fault(Fault.BAD_BRANCH_TARGET, op)
            return self._commit(op, [], imm)

        if op == "JZ":
            c = self._top0()
            if c.tag != int(Tag.BOOL):
                return self._fault(Fault.TYPE_FAULT, op, tag0=c.tag)
            if c.payload not in (0, 1):
                return self._fault(Fault.NONCANONICAL_BOOL, op, tag0=c.tag)
            if imm >= self.rom_depth:
                return self._fault(Fault.BAD_BRANCH_TARGET, op)
            return self._commit(op, [], imm if c.payload == 0 else self.pc + 1)

        if op == "EMIT":
            v = self._top0()
            self.output.append(v)
            rec = TraceRecord(self.seq, self.pc, self.pc + 1, op,
                              self.depth - 1, EVENT_OUTPUT, out_value=v)
            self.seq += 1
            self.trace.append(rec)
            self.stack.pop()
            self.pc += 1
            return rec

        if op == "HALT":
            self.halted = True
            rec = TraceRecord(self.seq, self.pc, self.pc, op, self.depth,
                              EVENT_COMMIT)
            self.seq += 1
            self.trace.append(rec)
            return rec

        raise AssertionError(f"unhandled opcode {op}")

    def _commit(self, op: str, new_tops: List[Value40], next_pc: int) -> TraceRecord:
        """Apply a consuming instruction: drop stack_in, push new_tops.

        pc_old is the faulting/retiring instruction's address, captured before
        the architectural mutation. Depth in the record is the depth after the
        transition (matches the book's expected commit trace)."""
        pc_old = self.pc
        ins = BY_MNEMONIC[op]
        for _ in range(ins.stack_in):
            self.stack.pop()
        self.stack.extend(new_tops)
        self.pc = next_pc
        rec = TraceRecord(self.seq, pc_old, next_pc, op, self.depth,
                          EVENT_COMMIT)
        self.seq += 1
        self.trace.append(rec)
        return rec

    # ------------------------------------------------------------------ run

    def run(self, max_steps: int = 100_000) -> List[TraceRecord]:
        for _ in range(max_steps):
            if self.halted or self.fault is not None:
                break
            self.step()
        return self.trace

    def final_line(self) -> str:
        code = self.fault.name if self.fault else "NONE"
        return (f"FINAL {int(self.halted)} {code} {self.pc} {self.depth} "
                f"{len(self.output)}")


def assemble_words(text_lines: List[str]) -> List[int]:
    """Load a .hex file (one 20-bit word per line, 5 hex digits)."""
    words = []
    for line in text_lines:
        line = line.strip()
        if not line:
            continue
        words.append(int(line, 16))
    return words


def load_hex(path: str) -> List[int]:
    with open(path) as f:
        return assemble_words(f.readlines())


def run_program(words: List[int], rom_depth: int = 1024,
                total_depth: int = 32, max_steps: int = 100_000
                ) -> Machine:
    m = Machine(program=words, rom_depth=rom_depth, total_depth=total_depth)
    m.run(max_steps)
    return m
