; Subroutine demo: sq(3) + sq(4) = 9 + 16 = 25 -> EMIT -> T0:00000019
        PUSH_S15 3
        CALL sq
        PUSH_S15 4
        CALL sq
        ADD
        EMIT
        HALT

sq:                       ; ( x -- x*x )
        DUP
        MUL
        RET
