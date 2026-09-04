; STACK_OVERFLOW (with total_depth small, e.g. 8 in the test harness):
; PUSH_S15 0 is also what unwritten ROM contains.
        PUSH_S15 1
        PUSH_S15 2
        PUSH_S15 3
        PUSH_S15 4
        PUSH_S15 5
        PUSH_S15 6
        PUSH_S15 7
        PUSH_S15 8
        PUSH_S15 9
        HALT
