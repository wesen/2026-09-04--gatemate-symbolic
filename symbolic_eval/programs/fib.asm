; Recursive fibonacci with CALL/RET (book Lab 1 extension).
; fib(n) consumes its argument and leaves exactly one result:
;   fib(n) = n              if n < 2
;          = fib(n-2)+fib(n-1)  otherwise
; The recursive case keeps n alive across both calls:
;   [n] -> [n, n-2] -> CALL -> [n, fib(n-2)] -> SWAP -> [fib(n-2), n]
;      -> [fib(n-2), n-1] -> CALL -> [fib(n-2), fib(n-1)] -> ADD
; fib(10) = 55 -> EMIT -> T0:00000037
        PUSH_S15 10
        CALL fib
        EMIT
        HALT

fib:
        DUP
        PUSH_S15 2
        LT                  ; BOOL(n < 2)
        JZ fib_rec          ; n >= 2 -> recursive case
        RET                 ; base case: n is the result

fib_rec:
        DUP
        PUSH_S15 2
        SUB                 ; [n, n-2]
        CALL fib            ; [n, fib(n-2)]
        SWAP                ; [fib(n-2), n]
        PUSH_S15 1
        SUB                 ; [fib(n-2), n-1]
        CALL fib            ; [fib(n-2), fib(n-1)]
        ADD                 ; [fib(n-1) + fib(n-2)]
        RET
