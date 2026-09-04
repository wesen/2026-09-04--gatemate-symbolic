// symbolic_types_pkg.sv — shared types for the Laboratory 1 tagged stack
// evaluator (book ch. 14 "Tagged records and event envelopes", Lab 1 ISA).
//
// This package mirrors tools/opcodes.py (the single source of truth):
// tags, the 20-bit instruction encoding, fault codes, and trace event codes
// must stay in sync with the Python table.
`default_nettype none

package symbolic_types_pkg;

  // ------------------------------------------------------------------ tags
  typedef enum logic [3:0] {
    TAG_INT   = 4'h0,
    TAG_BOOL  = 4'h1,
    TAG_REF   = 4'h2,
    TAG_ATOM  = 4'h3,
    TAG_PAIR  = 4'h4,
    TAG_THUNK = 4'h5,
    TAG_IND   = 4'h6,
    TAG_ERROR = 4'hD,
    TAG_POISON= 4'hE,
    TAG_EMPTY = 4'hF
  } value_tag_t;

  // value40: tag[3:0], flags[3:0], payload[31:0]
  // (tag is a plain vector: yosys cannot propagate enum-typed struct fields
  // out of package functions; the value_tag_t enum remains the documentation
  // of the tag assignment)
  typedef struct packed {
    logic [3:0]  tag;
    logic [3:0]  flags;
    logic [31:0] payload;
  } value40_t;

  // --------------------------------------------------------------- opcodes
  localparam [4:0] OP_PUSH_S15  = 5'h00;
  localparam [4:0] OP_PUSH_TRUE = 5'h01;
  localparam [4:0] OP_PUSH_FALSE= 5'h02;
  localparam [4:0] OP_ADD       = 5'h03;
  localparam [4:0] OP_SUB       = 5'h04;
  localparam [4:0] OP_MUL       = 5'h05;
  localparam [4:0] OP_EQ        = 5'h06;
  localparam [4:0] OP_LT        = 5'h07;
  localparam [4:0] OP_DUP       = 5'h08;
  localparam [4:0] OP_DROP      = 5'h09;
  localparam [4:0] OP_SWAP      = 5'h0A;
  localparam [4:0] OP_JMP       = 5'h0B;
  localparam [4:0] OP_JZ        = 5'h0C;
  localparam [4:0] OP_EMIT      = 5'h0D;
  localparam [4:0] OP_HALT      = 5'h0E;
  localparam [4:0] OP_CALL      = 5'h0F;
  localparam [4:0] OP_RET       = 5'h10;

  // ---------------------------------------------------------------- faults
  // (4 bits: the Lab 1 set plus the CALL/RET extension's return-stack faults)
  localparam [3:0] F_NONE            = 4'd0;
  localparam [3:0] F_STACK_UNDERFLOW = 4'd1;
  localparam [3:0] F_STACK_OVERFLOW  = 4'd2;
  localparam [3:0] F_TYPE_FAULT      = 4'd3;
  localparam [3:0] F_ARITH_OVERFLOW  = 4'd4;
  localparam [3:0] F_BAD_OPCODE      = 4'd5;
  localparam [3:0] F_BAD_BRANCH      = 4'd6;
  localparam [3:0] F_NONCANON_BOOL   = 4'd7;
  localparam [3:0] F_RSTACK_UNDERFLOW = 4'd8;
  localparam [3:0] F_RSTACK_OVERFLOW = 4'd9;

  // ---------------------------------------------------------- trace events
  localparam [1:0] EV_COMMIT = 2'd0;
  localparam [1:0] EV_OUTPUT = 2'd1;
  localparam [1:0] EV_FAULT  = 2'd2;

  // ------------------------------------------------------------ helpers
  // (classic function-name assignment style: yosys does not accept the
  // `return` statement in functions)
  function automatic value40_t mk_int(input logic signed [31:0] x);
    // Assign all packed bits together: constant-folded field assignments
    // lose tag/flags in the supported Yosys frontend (see test_synthesis.py).
    mk_int = {4'h0, 4'h0, x};
  endfunction

  function automatic value40_t mk_bool(input logic b);
    mk_bool = {4'h1, 4'h0, 31'b0, b};
  endfunction

  // Sign-extend a 15-bit immediate.
  function automatic logic signed [31:0] sx15(input logic [14:0] imm);
    logic signed [31:0] r;
    r = {{17{imm[14]}}, imm};
    sx15 = r;
  endfunction

endpackage

`default_nettype wire
