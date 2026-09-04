; BAD_BRANCH_TARGET: JMP 0x7FFF is outside the ROM. The assembler rejects
; out-of-ROM targets, so encode the word raw: JMP 0x7FFF = 0x5FFFF.
        WORD 0x5FFFF
        HALT
