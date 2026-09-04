; CALL/RET inside a loop: the subroutine consumes and emits the counter copy.
; Emits 5 4 3 2 1 then BOOL(true).
        PUSH_S15 5
loop:
        DUP
        CALL show          ; ( x -- ) emits x
        PUSH_S15 1
        SUB                ; [x-1]
        DUP
        PUSH_S15 0
        EQ                 ; BOOL(x-1 == 0)
        JZ loop            ; nonzero -> keep looping (JZ jumps on false)
        DROP               ; the zero counter
        PUSH_TRUE
        EMIT
        HALT

show:                      ; ( x -- ) emit and drop
        EMIT
        RET
