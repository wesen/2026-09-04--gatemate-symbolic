"""Model tests for the Laboratory 1 reference machine (book: directed tests
1, 2, 4, 5, 7, 8 + every fault case).

The Book's Program A expected commit trace is reproduced character-for-
character; Program B must enter TYPE_FAULT with an unchanged architectural
state.
"""

import pytest

from opcodes import (EVENT_COMMIT, EVENT_OUTPUT, Fault, Tag, encode)
from stack_model import (Machine, Value40, load_hex, make_bool, make_int,
                         run_program)

from asm20 import assemble as _assemble

# The book's first program: ((7 + 5) * 3) == 36 -> BOOL(true).
PROGRAM_A = [
    encode(0x00, 7),    # 0: PUSH_S15 7
    encode(0x00, 5),    # 1: PUSH_S15 5
    encode(0x03),       # 2: ADD
    encode(0x00, 3),    # 3: PUSH_S15 3
    encode(0x05),       # 4: MUL
    encode(0x00, 36),   # 5: PUSH_S15 36
    encode(0x06),       # 6: EQ
    encode(0x0D),       # 7: EMIT
    encode(0x0E),       # 8: HALT
]

# The book's second program: BOOL(true) + INT(4) -> TYPE_FAULT at the ADD.
PROGRAM_B = [
    encode(0x01),       # 0: PUSH_TRUE
    encode(0x00, 4),    # 1: PUSH_S15 4
    encode(0x03),       # 2: ADD
]

# Book Laboratory 1, "expected commit trace".
EXPECTED_A = [
    "TRACE 0 0 1 PUSH_S15 1 COMMIT",
    "TRACE 1 1 2 PUSH_S15 2 COMMIT",
    "TRACE 2 2 3 ADD 1 COMMIT",
    "TRACE 3 3 4 PUSH_S15 2 COMMIT",
    "TRACE 4 4 5 MUL 1 COMMIT",
    "TRACE 5 5 6 PUSH_S15 2 COMMIT",
    "TRACE 6 6 7 EQ 1 COMMIT",
    "TRACE 7 7 8 EMIT 0 OUTPUT 1 00000001",
    "TRACE 8 8 8 HALT 0 COMMIT",
]


class TestProgramA:
    def test_trace_matches_book_exactly(self):
        m = run_program(PROGRAM_A)
        lines = [r.line() for r in m.trace]
        assert lines == EXPECTED_A

    def test_output_and_final_state(self):
        m = run_program(PROGRAM_A)
        assert m.output == [make_bool(True)]
        assert m.halted and m.fault is None
        assert m.depth == 0
        assert m.final_line() == "FINAL 1 NONE 8 0 1 0"

    def test_stack_values_along_the_way(self):
        m = Machine(program=PROGRAM_A)
        steps = [m.step() for _ in range(9)]
        assert all(s is not None for s in steps)
        tops = [(r.op, r.depth) for r in m.trace]
        assert tops == [
            ("PUSH_S15", 1), ("PUSH_S15", 2), ("ADD", 1),
            ("PUSH_S15", 2), ("MUL", 1), ("PUSH_S15", 2),
            ("EQ", 1), ("EMIT", 0), ("HALT", 0),
        ]
        # ((7+5)*3) == 36 -> 36 == 36 -> true
        assert m.output[0] == make_bool(True)


class TestProgramB:
    def test_type_fault_is_precise(self):
        m = run_program(PROGRAM_B)
        assert m.fault is Fault.TYPE_FAULT
        assert m.fault_pc == 2
        # Precise: pc and stack unchanged by the failing instruction.
        assert m.pc == 2
        assert m.stack == [make_bool(True), make_int(4)]
        assert m.depth == 2
        assert m.output == []          # no output transfer
        assert not m.halted
        assert m.trace[-1].line() == \
            "TRACE 2 2 2 ADD 2 FAULT TYPE_FAULT 1 0"

    def test_fault_stops_the_machine(self):
        m = run_program(PROGRAM_B)
        assert m.step() is None       # no transition after a fault
        assert m.seq == 3             # PUSH_TRUE, PUSH_S15, fault record


class TestOpcodesAtMinimumDepth:
    def test_push_variants(self):
        m = run_program([encode(0x01), encode(0x02), encode(0x00, -1),
                         encode(0x0E)])
        assert m.stack == [make_bool(True), make_bool(False), make_int(-1)]

    def test_sign_extension(self):
        m = run_program([encode(0x00, -1), encode(0x00, 16383),
                         encode(0x00, -16384), encode(0x0E)])
        assert m.stack == [make_int(-1), make_int(16383), make_int(-16384)]

    def test_sub_top_is_right_operand(self):
        m = run_program([encode(0x00, 10), encode(0x00, 3), encode(0x04),
                         encode(0x0E)])
        assert m.stack == [make_int(7)]          # 10 - 3

    def test_lt_signed(self):
        m = run_program([encode(0x00, -5), encode(0x00, 3), encode(0x07),
                         encode(0x0E)])
        assert m.stack == [make_bool(True)]      # -5 < 3

    def test_eq_mixed_tags_is_false_not_fault(self):
        m = run_program([encode(0x01), encode(0x00, 1), encode(0x06),
                         encode(0x0E)])
        assert m.stack == [make_bool(False)]

    def test_eq_same_tag_payload(self):
        m = run_program([encode(0x00, 7), encode(0x00, 7), encode(0x06),
                         encode(0x0E)])
        assert m.stack == [make_bool(True)]

    def test_dup_drop_swap(self):
        m = run_program([encode(0x00, 1), encode(0x08), encode(0x0A),
                         encode(0x09), encode(0x0E)])
        assert m.stack == [make_int(1)]

    def test_jmp_forward_and_back(self):
        prog = [encode(0x0B, 2),   # 0: JMP 2
                encode(0x0E),     # 1: HALT (skipped)
                encode(0x0E)]     # 2: HALT
        m = run_program(prog)
        assert m.halted and m.pc == 2

    def test_jz_taken_and_not_taken(self):
        prog = [encode(0x02),      # 0: PUSH_FALSE
                encode(0x0C, 3),    # 1: JZ 3 (taken)
                encode(0x0E),      # 2: HALT (skipped)
                encode(0x01),      # 3: PUSH_TRUE
                encode(0x0C, 6),    # 4: JZ 6 (not taken)
                encode(0x0E),      # 5: HALT
                encode(0x0E)]      # 6: HALT (skipped)
        m = run_program(prog)
        assert m.halted and m.pc == 5
        assert [r.pc_old for r in m.trace] == [0, 1, 3, 4, 5]

    def test_emit_output_stream_order(self):
        prog = [encode(0x00, 1), encode(0x0D),
                encode(0x01), encode(0x0D), encode(0x0E)]
        m = run_program(prog)
        assert m.output == [make_int(1), make_bool(True)]


class TestFaults:
    def _fault(self, prog, code, **kw):
        m = run_program(prog, **kw)
        assert m.fault is code, m.trace
        return m

    def test_underflow_every_consumer(self):
        for prog in ([encode(0x03)],                 # ADD on empty
                     [encode(0x00, 1), encode(0x03)],  # ADD with depth 1
                     [encode(0x09)],                 # DROP
                     [encode(0x0A)],                 # SWAP
                     [encode(0x0D)],                 # EMIT
                     [encode(0x0C, 0)]):             # JZ
            self._fault(prog, Fault.STACK_UNDERFLOW)

    def test_stack_overflow(self):
        prog = [encode(0x00, 1)] * 5
        m = self._fault(prog, Fault.STACK_OVERFLOW, total_depth=4)
        # Precise: the 5th push did not happen.
        assert m.depth == 4 and m.pc == 4

    def test_type_fault_lt_and_jz(self):
        self._fault([encode(0x01), encode(0x00, 1), encode(0x07)],
                    Fault.TYPE_FAULT)                     # LT on BOOL
        self._fault([encode(0x00, 1), encode(0x0C, 0)],
                    Fault.TYPE_FAULT)                     # JZ on INT

    def test_noncanonical_bool_needs_constructed_state(self):
        # Not reachable from legal bytecode in Lab 1 (EQ/LT/PUSH_BOOL are
        # canonical), so build the state directly - the RTL must match.
        m = Machine(program=[encode(0x0C, 1)])
        m.stack = [Value40(tag=int(Tag.BOOL), payload=5)]
        m.step()
        assert m.fault is Fault.NONCANONICAL_BOOL
        assert m.pc == 0 and m.depth == 1

    def test_arith_overflow_add_sub_mul(self):
        big = (1 << 31) - 1
        for op in (0x03, 0x05):  # ADD, MUL
            m = Machine(program=[encode(op)])
            m.stack = [make_int(big), make_int(2)]
            m.step()
            assert m.fault is Fault.ARITH_OVERFLOW
            assert m.pc == 0 and m.depth == 2   # precise
        m = Machine(program=[encode(0x04)])     # SUB: INT_MIN - 1
        m.stack = [make_int(-(2**31)), make_int(1)]
        m.step()
        assert m.fault is Fault.ARITH_OVERFLOW

    def test_bad_opcode(self):
        m = self._fault([encode(0x1F)], Fault.BAD_OPCODE)
        assert m.pc == 0 and m.depth == 0

    def test_bad_branch_target_jmp_and_jz(self):
        self._fault([encode(0x0B, 0x7FFF)], Fault.BAD_BRANCH_TARGET)
        self._fault([encode(0x02), encode(0x0C, 0x7FFF)],
                    Fault.BAD_BRANCH_TARGET)
        # Target exactly rom_depth-1 is legal.
        m = run_program([encode(0x0B, 1), encode(0x0E)], rom_depth=2)
        assert m.halted and m.pc == 1

    def test_pc_past_program_within_rom(self):
        # Falling off the end of the program but inside the ROM: unwritten
        # ROM words read as 0x00000 = PUSH_S15 0 (zero-filled ROM), so the
        # machine keeps pushing until the stack overflows. Defined behavior,
        # matched by the RTL program_rom (explicit zero init).
        m = run_program([encode(0x00, 1)], total_depth=8)
        assert m.fault is Fault.STACK_OVERFLOW

    def test_pc_escape_past_rom_is_bad_branch_target(self):
        m = run_program([encode(0x0B, 1)], rom_depth=1)
        assert m.fault is Fault.BAD_BRANCH_TARGET

    def test_all_faults_leave_state_precise(self):
        cases = [
            ([encode(0x03)], Fault.STACK_UNDERFLOW),
            ([encode(0x01), encode(0x00, 1), encode(0x03)],
             Fault.TYPE_FAULT),
        ]
        for prog, code in cases:
            m = run_program(prog)
            assert m.pc == m.fault_pc


class TestCallRet:
    """CALL/RET extension: return-address stack, precise rstack faults."""

    def _words(self, name):
        w, _, _ = _assemble(open(f"programs/{name}.asm").read())
        return w

    def test_call_ret_return_address(self):
        prog = [encode(0x0F, 3),   # 0: CALL 3
                encode(0x0E),      # 1: HALT (return lands here)
                encode(0x0E),     # 2: HALT (unreachable)
                encode(0x01),      # 3: sub: PUSH_TRUE
                encode(0x0D),      # 4: EMIT
                encode(0x10)]      # 5: RET
        m = run_program(prog)
        assert m.output == [make_bool(True)]
        assert m.halted and m.pc == 1 and m.rstack == []

    def test_nested_calls(self):
        prog = [encode(0x0F, 3),   # 0: CALL a
                encode(0x0E),      # 1: HALT
                encode(0x0E),      # 2: (unreachable)
                encode(0x0F, 6),  # 3: a: CALL b
                encode(0x10),      # 4: RET
                encode(0x0E),      # 5: (unreachable)
                encode(0x00, 42), # 6: b: PUSH 42
                encode(0x10)]      # 7: RET
        m = run_program(prog)
        assert m.stack == [make_int(42)]
        assert m.halted and m.pc == 1 and m.rstack == []

    def test_ret_underflow_precise(self):
        m = run_program([encode(0x10)])
        assert m.fault is Fault.RSTACK_UNDERFLOW
        assert m.pc == 0 and m.rstack == []
        assert m.trace[-1].line() == "TRACE 0 0 0 RET 0 FAULT RSTACK_UNDERFLOW 0 0"

    def test_call_overflow_precise(self):
        prog = [encode(0x0F, 0), encode(0x0E)]   # self: CALL self
        m = run_program(prog, rstack_depth=4)
        assert m.fault is Fault.RSTACK_OVERFLOW
        assert m.pc == 0 and len(m.rstack) == 4  # precise: 4 live, 5th refused

    def test_fib_recursive(self):
        m = run_program(self._words("fib"))
        assert m.output == [make_int(55)]
        assert m.halted and m.rstack == [] and m.depth == 0

    def test_countdown_and_sq(self):
        m = run_program(self._words("countdown"))
        assert [v.text() for v in m.output] == \
            ["INT(5)", "INT(4)", "INT(3)", "INT(2)", "INT(1)", "BOOL(1)"]
        assert m.halted and m.rstack == []
        m = run_program(self._words("sq"))
        assert m.output == [make_int(25)]


class TestValue40:
    def test_word_packing(self):
        v = Value40(tag=0x1, flags=0x2, payload=0xDEADBEEF)
        assert v.word() == 0x12DEADBEEF
        assert Value40.from_word(v.word()) == v

    def test_make_int_two_complement(self):
        assert make_int(-1).payload == 0xFFFFFFFF
        assert make_int(0).word() == 0


class TestHaltingAndTrace:
    def test_halt_keeps_pc(self):
        m = run_program([encode(0x0E)])
        assert m.halted and m.pc == 0
        assert m.trace[0].line() == "TRACE 0 0 0 HALT 0 COMMIT"

    def test_seq_increments_and_events(self):
        m = run_program(PROGRAM_A)
        assert [r.seq for r in m.trace] == list(range(9))
        assert [r.event for r in m.trace] == [
            EVENT_COMMIT] * 7 + [EVENT_OUTPUT, EVENT_COMMIT]

    def test_run_program_from_hex_loader(self, tmp_path):
        p = tmp_path / "a.hex"
        p.write_text("\n".join(f"{w:05x}" for w in PROGRAM_A) + "\n")
        m = run_program(load_hex(str(p)))
        assert [r.line() for r in m.trace] == EXPECTED_A
