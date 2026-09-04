; Control flow: JZ taken/not-taken, JMP backward loop, EMIT per iteration.
; The loop counter is an INT; JZ branches on FALSE, so "JZ loop" continues
; while (counter-1 != 0) and falls through to the exit when it is zero.
        PUSH_S15 3
loop:
        DUP
        EMIT            ; emit the counter
        PUSH_S15 1
        SUB             ; counter-1
        DUP
        PUSH_S15 0
        EQ              ; BOOL(counter == 0)
        JZ loop         ; false (nonzero) -> keep looping
        DROP            ; true (zero) -> drop the zero counter
        PUSH_TRUE
        EMIT
        HALT
