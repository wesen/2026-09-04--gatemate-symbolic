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
  typedef struct packed {
    value_tag_t  tag;
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

  // ---------------------------------------------------------------- faults
  localparam [2:0] F_NONE            = 3'd0;
  localparam [2:0] F_STACK_UNDERFLOW = 3'd1;
  localparam [2:0] F_STACK_OVERFLOW  = 3'd2;
  localparam [2:0] F_TYPE_FAULT      = 3'd3;
  localparam [2:0] F_ARITH_OVERFLOW  = 3'd4;
  localparam [2:0] F_BAD_OPCODE      = 3'd5;
  localparam [2:0] F_BAD_BRANCH      = 3'd6;
  localparam [2:0] F_NONCANON_BOOL   = 3'd7;

  // ---------------------------------------------------------- trace events
  localparam [1:0] EV_COMMIT = 2'd0;
  localparam [1:0] EV_OUTPUT = 2'd1;
  localparam [1:0] EV_FAULT  = 2'd2;

  // ------------------------------------------------------------ helpers
  function automatic value40_t mk_int(input logic signed [31:0] x);
    value40_t v;
    v.tag     = TAG_INT;
    v.flags   = 4'h0;
    v.payload = x;
    return v;
  endfunction

  function automatic value40_t mk_bool(input logic b);
    value40_t v;
    v.tag     = TAG_BOOL;
    v.flags   = 4'h0;
    v.payload = b ? 32'd1 : 32'd0;
    return v;
  endfunction

  // Sign-extend a 15-bit immediate.
  function automatic logic signed [31:0] sx15(input logic [14:0] imm);
    return {{17{imm[14]}}, imm};
  endfunction

endpackage

`default_nettype wire
