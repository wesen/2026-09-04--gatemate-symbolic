; RSTACK_OVERFLOW: infinite self-call fills the return stack (16 entries)
; and faults precisely on the 17th CALL.
self:
        CALL self
        HALT
