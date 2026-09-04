; Program B (book exit criterion): BOOL(true) + INT(4) -> precise TYPE_FAULT
; at the ADD: pc and stack unchanged, no output.
        PUSH_TRUE
        PUSH_S15 4
        ADD
        HALT
