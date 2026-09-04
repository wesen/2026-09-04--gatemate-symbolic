; Stack boundary exercise: spill past a two-entry top cache, then drain.
; Exercises push/spill, DROP refill, binary-op refill, SWAP.
        PUSH_S15 1
        PUSH_S15 2
        PUSH_S15 3     ; spill begins here in the BRAM version
        PUSH_S15 4
        PUSH_S15 5
        PUSH_S15 6
        ADD            ; 5+6 = 11 (operands: top0=6, top1=5? no: see listing)
        DROP
        SWAP
        DROP
        DUP
        EMIT
        DROP
        DROP
        PUSH_S15 10
        PUSH_S15 20
        SUB
        EMIT            ; 10-20 = -10
        HALT
