"""Assembler tests: round-trip, label resolution, error cases, and the
exit-criteria program hex fixtures."""

import pytest

from asm20 import assemble
from opcodes import encode


class TestRoundTrip:
    def test_arith_program_a(self):
        words, syms, _ = assemble(open("programs/arith.asm").read())
        assert words[:9] == [
            encode(0x00, 7), encode(0x00, 5), encode(0x03),
            encode(0x00, 3), encode(0x05), encode(0x00, 36),
            encode(0x06), encode(0x0D), encode(0x0E),
        ]

    def test_typefault_program_b(self):
        words, _, _ = assemble(open("programs/typefault.asm").read())
        assert words[:3] == [encode(0x01), encode(0x00, 4), encode(0x03)]

    def test_labels_and_backward_jump(self):
        src = """
        PUSH_S15 3
loop:
        DUP
        PUSH_S15 1
        SUB
        DUP
        JZ done
        JMP loop
done:
        DROP
        HALT
"""
        words, syms, _ = assemble(src, rom_depth=16)
        assert syms == {"loop": 1, "done": 7}
        # JZ done -> target 7; JMP loop -> target 1
        assert words[5] == encode(0x0C, 7)
        assert words[6] == encode(0x0B, 1)

    def test_word_directive(self):
        words, _, _ = assemble("WORD 0x1F\nHALT\n")
        assert words == [0x1F, encode(0x0E)]

    def test_negative_immediate(self):
        words, _, _ = assemble("PUSH_S15 -1\nHALT\n")
        assert words[0] == encode(0x00, -1)

    def test_rom_padding_only_in_hex_writer(self):
        # assemble() returns only the real words; main() pads the .hex file
        # to ROM depth with zero words (defined: PUSH_S15 0).
        words, _, _ = assemble("HALT\n", rom_depth=8)
        assert words == [encode(0x0E)]


class TestErrors:
    def test_unknown_mnemonic(self):
        with pytest.raises(ValueError, match="unknown mnemonic"):
            assemble("FROB 1\n")

    def test_duplicate_label(self):
        with pytest.raises(ValueError, match="duplicate label"):
            assemble("x:\nx:\nHALT\n")

    def test_imm_range(self):
        with pytest.raises(ValueError, match="15-bit range"):
            assemble("PUSH_S15 16384\n")
        with pytest.raises(ValueError, match="15-bit range"):
            assemble("PUSH_S15 -16385\n")

    def test_branch_outside_rom(self):
        with pytest.raises(ValueError, match="outside ROM"):
            assemble("JMP 99\n", rom_depth=16)

    def test_stray_operand(self):
        with pytest.raises(ValueError, match="takes no operand"):
            assemble("ADD 5\n")


class TestAssembledProgramsRun:
    """Every program in programs/ must assemble and be consumable by the
    model (they run later in the differential RTL tests too)."""

    PROGRAMS = ["arith", "typefault", "smoke", "deep", "branch",
                "muloverflow", "underflow", "badop", "badbranch",
                "jztype", "stackoverflow"]

    def test_all_assemble(self):
        for name in self.PROGRAMS:
            words, _, _ = assemble(open(f"programs/{name}.asm").read())
            assert len(words) <= 1024
            assert all(0 <= w <= 0xFFFFF for w in words)
