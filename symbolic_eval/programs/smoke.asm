; Basic sanity: arithmetic, compare, emit several values, halt.
        PUSH_S15 2
        PUSH_S15 3
        ADD             ; 5
        EMIT
        PUSH_S15 -1
        EMIT
        PUSH_TRUE
        EMIT
        PUSH_FALSE
        EMIT
        HALT
