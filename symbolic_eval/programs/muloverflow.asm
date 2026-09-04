; ARITH_OVERFLOW at the second MUL: 16383*16383 = 268402689 fits, then
; *16383 overflows signed 32-bit.
        PUSH_S15 16383
        PUSH_S15 16383
        MUL
        PUSH_S15 16383
        MUL
        HALT
