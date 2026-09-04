"""Authoritative metadata for the Laboratory 1 tagged stack evaluator.

This module is the single source of truth for the instruction set, the value
tags, and the fault codes. The assembler (asm20.py), the executable reference
model (stack_model.py), and the tests import this module. Handwritten RTL
constants and testbench name tables are checked by scripts/check_isa.py.

Book: Composable_Hardware_Patterns_for_Symbolic_Computers.md, Laboratory 1
(ticket sources/, lines 4211-4569). Conventions:

- 20-bit instruction word: [19:15] opcode, [14:0] signed immediate or branch
  target.
- Stack effects use Forth notation: ( a b -- r ) means b is TOS (top0), a is
  NOS (top1).
- The architectural machine is M = <pc, stack, rstack, output_stream, fault, halted>.
- No architectural mutation occurs before all checks for an instruction pass.
- Opcodes not listed below are illegal -> BAD_OPCODE.
"""

from __future__ import annotations

from dataclasses import dataclass
from enum import IntEnum
from typing import Dict


# ---------------------------------------------------------------------------
# Value tags (book ch. 14, "Tagged records and event envelopes")
# ---------------------------------------------------------------------------

class Tag(IntEnum):
    """Semantic type of a value40 word. Lab 1 produces/consumes INT, BOOL."""

    INT = 0x0
    BOOL = 0x1
    REF = 0x2
    ATOM = 0x3
    PAIR = 0x4
    THUNK = 0x5
    IND = 0x6
    ERROR = 0xD
    POISON = 0xE
    EMPTY = 0xF


# Canonical text codes used in traces and by the RTL testbenches.
TAG_NAMES: Dict[int, str] = {
    int(t): t.name for t in Tag
}

# value40 layout: tag[3:0], flags[3:0], payload[31:0].
VALUE_BITS = 40
PAYLOAD_BITS = 32
INT32_MIN = -(2**31)
INT32_MAX = 2**31 - 1


# ---------------------------------------------------------------------------
# Fault codes (book Lab 1, "Fault state"). Closed set.
# ---------------------------------------------------------------------------

class Fault(IntEnum):
    NONE = 0x00
    STACK_UNDERFLOW = 0x01
    STACK_OVERFLOW = 0x02
    TYPE_FAULT = 0x03
    ARITH_OVERFLOW = 0x04
    BAD_OPCODE = 0x05
    BAD_BRANCH_TARGET = 0x06
    NONCANONICAL_BOOL = 0x07
    RSTACK_UNDERFLOW = 0x08
    RSTACK_OVERFLOW = 0x09


# ---------------------------------------------------------------------------
# Instruction metadata
# ---------------------------------------------------------------------------

@dataclass(frozen=True)
class Instruction:
    """Metadata for one opcode.

    Attributes:
        opcode:     5-bit opcode value ([19:15] of the instruction word).
        mnemonic:   uppercase assembler mnemonic.
        imm_kind:   "none", "s15" (signed 15-bit immediate), or "u15"
                    (unsigned 15-bit branch target).
        stack_in:   operand-stack items consumed (precondition count).
        stack_out:  items produced after the consume.
        effect:     stack effect in Forth notation (top0 = rightmost).
        desc:       one-line behavior description.
    """

    opcode: int
    mnemonic: str
    imm_kind: str
    stack_in: int
    stack_out: int
    effect: str
    desc: str


_INSTRUCTIONS = [
    Instruction(0x00, "PUSH_S15",  "s15", 0, 1, "( -- INT(k) )",
                "sign-extend 15-bit immediate and push"),
    Instruction(0x01, "PUSH_TRUE",  "none", 0, 1, "( -- BOOL(1) )",
                "push canonical true"),
    Instruction(0x02, "PUSH_FALSE", "none", 0, 1, "( -- BOOL(0) )",
                "push canonical false"),
    Instruction(0x03, "ADD", "none", 2, 1, "( INT a, INT b -- INT(a+b) )",
                "typed add; precise signed 32-bit overflow"),
    Instruction(0x04, "SUB", "none", 2, 1, "( INT a, INT b -- INT(a-b) )",
                "typed subtract (top0 is right operand)"),
    Instruction(0x05, "MUL", "none", 2, 1, "( INT a, INT b -- INT(a*b) )",
                "typed multiply; precise signed 32-bit overflow"),
    Instruction(0x06, "EQ", "none", 2, 1, "( x, y -- BOOL(x==y) )",
                "tag-policy equality; mixed tags yield BOOL(false)"),
    Instruction(0x07, "LT", "none", 2, 1, "( INT a, INT b -- BOOL(a<b) )",
                "typed signed compare"),
    Instruction(0x08, "DUP", "none", 1, 2, "( x -- x, x )",
                "duplicate top0; capacity checked first"),
    Instruction(0x09, "DROP", "none", 1, 0, "( x -- )",
                "drop top0; underflow checked first"),
    Instruction(0x0A, "SWAP", "none", 2, 2, "( x, y -- y, x )",
                "exchange top0 and top1"),
    Instruction(0x0B, "JMP", "u15", 0, 0, "( -- )",
                "absolute jump; target checked"),
    Instruction(0x0C, "JZ", "u15", 1, 0, "( BOOL c -- )",
                "branch if false; rejects non-Boolean and noncanonical"),
    Instruction(0x0D, "EMIT", "none", 1, 0, "( x -- )",
                "send top0 to the output channel; pop only on acceptance"),
    Instruction(0x0E, "HALT", "none", 0, 0, "( -- )",
                "enter halted state"),
    Instruction(0x0F, "CALL", "u15", 0, 0, "( -- )",
                "push return address (pc+1) on the return stack, jump"),
    Instruction(0x10, "RET", "none", 0, 0, "( -- )",
                "pop the return address and jump to it"),
]

INSTRUCTIONS: Dict[int, Instruction] = {i.opcode: i for i in _INSTRUCTIONS}
BY_MNEMONIC: Dict[str, Instruction] = {i.mnemonic: i for i in _INSTRUCTIONS}

OPCODE_BITS = 5
IMM_BITS = 15
OPCODE_ILLEGAL_MIN = 0x11  # first undefined opcode -> BAD_OPCODE

# Event codes on the commit trace (book Lab 1: commit / output / fault).
EVENT_COMMIT = 0
EVENT_OUTPUT = 1
EVENT_FAULT = 2
EVENT_NAMES = {EVENT_COMMIT: "COMMIT", EVENT_OUTPUT: "OUTPUT", EVENT_FAULT: "FAULT"}


def encode(opcode: int, imm: int = 0) -> int:
    """Pack opcode + signed 15-bit immediate into one 20-bit word."""
    assert 0 <= opcode <= 0x1F
    # Full 15-bit representable range: signed immediates (s15, -16384..16383)
    # and unsigned branch targets (u15, 0..32767). Kind-specific validation is
    # the assembler's and model's job.
    assert -(2**14) <= imm <= 2**15 - 1, f"immediate out of 15-bit range: {imm}"
    word = (opcode << IMM_BITS) | (imm & 0x7FFF)
    assert 0 <= word <= 0xFFFFF
    return word


def decode(word: int) -> "tuple[int, int]":
    """Split a 20-bit word into (opcode, signed_or_unsigned_imm)."""
    return (word >> IMM_BITS) & 0x1F, word & 0x7FFF


def sign_extend_s15(imm15: int) -> int:
    """Sign-extend a 15-bit immediate."""
    assert 0 <= imm15 <= 0x7FFF
    return imm15 - 0x8000 if imm15 & 0x4000 else imm15


def mnemonic(word: int) -> str:
    """Mnemonic for a word, or 'BAD' for undefined opcodes."""
    op, _ = decode(word)
    ins = INSTRUCTIONS.get(op)
    return ins.mnemonic if ins else "BAD"
