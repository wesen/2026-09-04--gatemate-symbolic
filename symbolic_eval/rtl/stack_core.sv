// stack_core.sv — Laboratory 1 tagged stack evaluator, register-stack
// version (book Lab 1, "Register-stack version").
//
// Contract (design doc §6):
//   * No architectural mutation before all checks pass: EXECUTE validates
//     everything and stages complete next-state data; COMMIT is the single
//     mutation owner and the only place registers and the stack change.
//   * A commit pulse fires only in COMMIT, on output acceptance in
//     OUTPUT_WAIT, or on entry to a precise fault state.
//   * EMIT pops only when the output channel accepts (Delayed Irreversible
//     Store): out_valid is held stable with out_data while blocked.
//   * Faults are precise: pc and depth stay at their pre-instruction values.
//
// Check order per instruction (must match the model exactly):
//   capacity/underflow -> operand tags -> canonicality -> branch target.
`default_nettype none

module stack_core #(
  parameter int ROM_DEPTH    = 1024,   // 1K x 20 instruction ROM
  parameter int STACK_DEPTH = 32,     // register stack (P3 version)
  parameter int RSTACK_DEPTH = 16    // return-address stack (registers)
) (
  input  logic clk,
  input  logic rst_n,        // async-assert, sync-release (reset_sync output)

  // Program ROM port: synchronous read, data valid one cycle after addr.
  output logic [$clog2(ROM_DEPTH)-1:0] rom_addr,
  input  logic [19:0] rom_data,

  // Commit trace (one-cycle pulse per retired instruction / fault / output).
  output logic        trace_valid,
  output logic        trace_fetch,  // opcode is not meaningful on a fetch fault
  output logic        fault_fetch,
  output logic [31:0] trace_seq,
  output logic [$clog2(ROM_DEPTH+1)-1:0] trace_pc_old,
  output logic [$clog2(ROM_DEPTH+1)-1:0] trace_pc_new,
  output logic [4:0]  trace_op,
  output logic [$clog2(STACK_DEPTH+1)-1:0] trace_depth,
  output logic [1:0]  trace_event,          // symbolic_types_pkg::EV_COMMIT / symbolic_types_pkg::EV_OUTPUT / symbolic_types_pkg::EV_FAULT
  output logic [3:0]  trace_fault,           // valid when symbolic_types_pkg::EV_FAULT
  output logic [3:0]  trace_tag1, trace_tag0,
  output logic [3:0]  trace_out_tag,         // valid when symbolic_types_pkg::EV_OUTPUT
  output logic [31:0] trace_out_payload,

  // Output channel (ready/valid).
  output logic        out_valid,
  output logic [39:0] out_data,
  input  logic        out_ready,

  // Precise fault record (first fault wins; machine stops).
  output logic        fault_valid,
  output logic [3:0]  fault_code,
  output logic [$clog2(ROM_DEPTH+1)-1:0] fault_pc,
  output logic [4:0]  fault_op,
  output logic [$clog2(STACK_DEPTH+1)-1:0] fault_depth,
  output logic [3:0]  fault_tag1,
  output logic [3:0]  fault_tag0,

  output logic        halted,
  output logic [$clog2(ROM_DEPTH+1)-1:0] pc_o,
  output logic [$clog2(STACK_DEPTH+1)-1:0] depth_o,
  output logic [$clog2(RSTACK_DEPTH+1)-1:0] rdepth_o
);

  // ---------------------------------------------------------------- states
  typedef enum logic [3:0] {
    S_RESET, S_FETCH, S_FETCH_WAIT, S_DECODE, S_EXECUTE, S_COMMIT,
    S_OUTPUT_WAIT, S_FAULT, S_HALTED
  } state_t;

  state_t state_q, state_d;

  // ------------------------------------------------- architectural state
  logic [$clog2(ROM_DEPTH+1)-1:0]        pc_q, pc_d;
  logic [$clog2(STACK_DEPTH+1)-1:0]    depth_q, depth_d;
  symbolic_types_pkg::value40_t                            stack_q [0:STACK_DEPTH-1];
  logic [19:0]                         ir_q, ir_d;
  logic [31:0]                         seq_q, seq_d;
  logic                                halted_q, halted_d;

  // staged next-state (EXECUTE -> COMMIT): complete before any mutation
  logic [$clog2(ROM_DEPTH+1)-1:0]        npc_q, npc_d;
  logic [$clog2(STACK_DEPTH+1)-1:0]    ndepth_q, ndepth_d;
  logic                                wrA_en_q, wrA_en_d;
  logic                                wrB_en_q, wrB_en_d;
  logic [$clog2(STACK_DEPTH)-1:0]      wrA_addr_q, wrA_addr_d;
  logic [$clog2(STACK_DEPTH)-1:0]      wrB_addr_q, wrB_addr_d;
  symbolic_types_pkg::value40_t wrA_data_q;
  symbolic_types_pkg::value40_t wrA_data_d;
  symbolic_types_pkg::value40_t wrB_data_q;
  symbolic_types_pkg::value40_t wrB_data_d;
  symbolic_types_pkg::value40_t pending_q;
  symbolic_types_pkg::value40_t pending_d;
  logic                                halt_stage_q, halt_stage_d;

  // return-address stack (continuation state; book Lab 1 extension)
  logic [$clog2(ROM_DEPTH+1)-1:0]        rstack_q [0:RSTACK_DEPTH-1];
  logic [$clog2(RSTACK_DEPTH+1)-1:0]  rdepth_q, rdepth_d;
  logic                                wrR_en_q, wrR_en_d;
  logic [$clog2(RSTACK_DEPTH)-1:0]    wrR_addr_q, wrR_addr_d;
  logic [$clog2(ROM_DEPTH+1)-1:0]       wrR_data_q, wrR_data_d;

  logic [$clog2(RSTACK_DEPTH+1)-1:0] nrdepth_q, nrdepth_d;

  logic trace_fetch_q, trace_fetch_d, fault_fetch_q, fault_fetch_d;

  // fault record
  logic                                fv_q, fv_d;
  logic [3:0]                          fc_q, fc_d;
  logic [$clog2(ROM_DEPTH+1)-1:0]        fpc_q, fpc_d;
  logic [4:0]                          fop_q, fop_d;
  logic [$clog2(STACK_DEPTH+1)-1:0]    fdepth_q, fdepth_d;
  logic [3:0]                          ft1_q, ft1_d;
  logic [3:0]                          ft0_q, ft0_d;

  // trace snapshot registers (valid for one cycle after trace_valid pulses)
  logic        trace_valid_q, trace_valid_d;
  logic [31:0] trace_seq_q, trace_seq_d;
  logic [$clog2(ROM_DEPTH+1)-1:0] trace_pc_old_q, trace_pc_old_d;
  logic [$clog2(ROM_DEPTH+1)-1:0] trace_pc_new_q, trace_pc_new_d;
  logic [4:0]  trace_op_q, trace_op_d;
  logic [$clog2(STACK_DEPTH+1)-1:0] trace_depth_q, trace_depth_d;
  logic [1:0]  trace_event_q, trace_event_d;
  logic [3:0]  trace_fault_q, trace_fault_d;
  logic [3:0]  trace_tag1_q, trace_tag1_d;
  logic [3:0]  trace_tag0_q, trace_tag0_d;
  logic [3:0]  trace_out_tag_q, trace_out_tag_d;
  logic [31:0] trace_out_payload_q, trace_out_payload_d;

  // ------------------------------------------------------- decode helpers
  logic [4:0]  op;
  logic [14:0] imm;
  assign op  = ir_q[19:15];
  assign imm = ir_q[14:0];

  symbolic_types_pkg::value40_t top0_w;
  symbolic_types_pkg::value40_t top1_w;
  always_comb begin
    top0_w = '0;
    top1_w = '0;
    if (depth_q >= 1) top0_w = stack_q[depth_q-1];  // newest
    if (depth_q >= 2) top1_w = stack_q[depth_q-2];  // next
  end

  // 64-bit arithmetic with precise signed 32-bit overflow detection (DR-2).
  logic signed [63:0] add_w, sub_w, mul_w;
  logic add_ovf, sub_ovf, mul_ovf;
  always_comb begin
    add_w = 64'($signed(top1_w.payload)) + 64'($signed(top0_w.payload));
    sub_w = 64'($signed(top1_w.payload)) - 64'($signed(top0_w.payload));
    mul_w = 64'($signed(top1_w.payload)) * 64'($signed(top0_w.payload));
    add_ovf = (add_w[63:32] != {32{add_w[31]}});
    sub_ovf = (sub_w[63:32] != {32{sub_w[31]}});
    mul_ovf = (mul_w[63:32] != {32{mul_w[31]}});
  end

  // ------------------------------------------------------------- the FSM
  always_comb begin
    // defaults: hold everything
    state_d     = state_q;
    pc_d       = pc_q;
    depth_d    = depth_q;
    ir_d       = ir_q;
    seq_d      = seq_q;
    halted_d   = halted_q;
    npc_d      = npc_q;
    ndepth_d   = ndepth_q;
    wrA_en_d   = 1'b0;
    wrB_en_d   = 1'b0;
    wrA_addr_d = wrA_addr_q;
    wrB_addr_d = wrB_addr_q;
    wrA_data_d = wrA_data_q;
    wrB_data_d = wrB_data_q;
    pending_d  = pending_q;
    halt_stage_d = halt_stage_q;
    rdepth_d   = rdepth_q;
    nrdepth_d  = nrdepth_q;
    wrR_en_d   = 1'b0;
    wrR_addr_d = wrR_addr_q;
    wrR_data_d = wrR_data_q;
    fv_d       = fv_q;
    fc_d       = fc_q;
    fpc_d      = fpc_q;
    fop_d      = fop_q;
    fdepth_d   = fdepth_q;
    ft1_d      = ft1_q;
    ft0_d      = ft0_q;
    trace_valid_d       = 1'b0;
    trace_seq_d         = trace_seq_q;
    trace_pc_old_d      = trace_pc_old_q;
    trace_pc_new_d      = trace_pc_new_q;
    trace_op_d          = trace_op_q;
    trace_depth_d       = trace_depth_q;
    trace_event_d       = trace_event_q;
    trace_fault_d       = trace_fault_q;
    trace_tag1_d        = trace_tag1_q;
    trace_tag0_d        = trace_tag0_q;
    trace_out_tag_d     = trace_out_tag_q;
    trace_out_payload_d = trace_out_payload_q;
    rom_addr = (pc_q < ROM_DEPTH) ? pc_q[$clog2(ROM_DEPTH)-1:0] : '0;
    out_valid  = 1'b0;

    trace_fetch_d = 1'b0;
    fault_fetch_d = fault_fetch_q;

    case (state_q)
      S_RESET: begin
        if (rst_n) begin
          state_d = S_FETCH;
        end
      end

      S_FETCH: begin
        if (pc_q >= ROM_DEPTH) begin
          do_fault(symbolic_types_pkg::F_BAD_BRANCH);
          trace_fetch_d = 1'b1;
          fault_fetch_d = 1'b1;
        end else begin
          state_d = S_FETCH_WAIT;
        end
      end

      S_FETCH_WAIT: begin
        ir_d    = rom_data;
        state_d = S_DECODE;
      end

      S_DECODE: begin
        state_d = S_EXECUTE;
      end

      // All precondition checks happen here. Nothing architectural is
      // mutated; EXECUTE only stages complete next-state data.
      S_EXECUTE: begin
        nrdepth_d = rdepth_q;
        // Build a complete candidate from current architectural state.
        // EMIT retires outside COMMIT, so previous staged depth is stale.
        npc_d = pc_q + 1'b1;
        ndepth_d = depth_q;
        halt_stage_d = 1'b0;
        case (op)
          symbolic_types_pkg::OP_PUSH_S15, symbolic_types_pkg::OP_PUSH_TRUE, symbolic_types_pkg::OP_PUSH_FALSE: begin
            if (depth_q == STACK_DEPTH[$clog2(STACK_DEPTH+1)-1:0]) begin
              do_fault(symbolic_types_pkg::F_STACK_OVERFLOW);
            end else begin
              wrA_en_d   = 1'b1;
              wrA_addr_d = depth_q[$clog2(STACK_DEPTH)-1:0];
              wrA_data_d = (op == symbolic_types_pkg::OP_PUSH_S15)  ? symbolic_types_pkg::mk_int(symbolic_types_pkg::sx15(imm)) :
                           (op == symbolic_types_pkg::OP_PUSH_TRUE) ? symbolic_types_pkg::mk_bool(1'b1)    :
                                                  symbolic_types_pkg::mk_bool(1'b0);
              ndepth_d = depth_q + 1'b1;
              npc_d    = pc_q + 1'b1;
              state_d  = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_ADD, symbolic_types_pkg::OP_SUB, symbolic_types_pkg::OP_MUL, symbolic_types_pkg::OP_EQ, symbolic_types_pkg::OP_LT: begin
            if (depth_q < 2) begin
              do_fault(symbolic_types_pkg::F_STACK_UNDERFLOW);
            end else if (op != symbolic_types_pkg::OP_EQ &&
                         (top1_w.tag != symbolic_types_pkg::TAG_INT || top0_w.tag != symbolic_types_pkg::TAG_INT)) begin
              do_fault(symbolic_types_pkg::F_TYPE_FAULT);
            end else if (op == symbolic_types_pkg::OP_ADD && add_ovf) begin
              do_fault(symbolic_types_pkg::F_ARITH_OVERFLOW);
            end else if (op == symbolic_types_pkg::OP_SUB && sub_ovf) begin
              do_fault(symbolic_types_pkg::F_ARITH_OVERFLOW);
            end else if (op == symbolic_types_pkg::OP_MUL && mul_ovf) begin
              do_fault(symbolic_types_pkg::F_ARITH_OVERFLOW);
            end else begin
              wrA_en_d   = 1'b1;
              wrA_addr_d = depth_q - 2;
              unique case (op)
                symbolic_types_pkg::OP_ADD: wrA_data_d = symbolic_types_pkg::mk_int(add_w[31:0]);
                symbolic_types_pkg::OP_SUB: wrA_data_d = symbolic_types_pkg::mk_int(sub_w[31:0]);
                symbolic_types_pkg::OP_MUL: wrA_data_d = symbolic_types_pkg::mk_int(mul_w[31:0]);
                symbolic_types_pkg::OP_EQ:  wrA_data_d = symbolic_types_pkg::mk_bool(top1_w == top0_w);
                symbolic_types_pkg::OP_LT:  wrA_data_d = symbolic_types_pkg::mk_bool($signed(top1_w.payload) <
                                             $signed(top0_w.payload));
                default: ;
              endcase
              ndepth_d = depth_q - 1;
              npc_d    = pc_q + 1'b1;
              state_d  = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_DUP: begin
            if (depth_q == 0) begin
              do_fault(symbolic_types_pkg::F_STACK_UNDERFLOW);
            end else if (depth_q == STACK_DEPTH[$clog2(STACK_DEPTH+1)-1:0]) begin
              do_fault(symbolic_types_pkg::F_STACK_OVERFLOW);
            end else begin
              wrA_en_d   = 1'b1;
              wrA_addr_d = depth_q[$clog2(STACK_DEPTH)-1:0];
              wrA_data_d = top0_w;
              ndepth_d   = depth_q + 1'b1;
              npc_d      = pc_q + 1'b1;
              state_d    = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_DROP, symbolic_types_pkg::OP_JZ: begin
            if (depth_q == 0) begin
              do_fault(symbolic_types_pkg::F_STACK_UNDERFLOW);
            end else if (op == symbolic_types_pkg::OP_JZ && top0_w.tag != symbolic_types_pkg::TAG_BOOL) begin
              do_fault(symbolic_types_pkg::F_TYPE_FAULT);
            end else if (op == symbolic_types_pkg::OP_JZ && top0_w.payload > 32'd1) begin
              do_fault(symbolic_types_pkg::F_NONCANON_BOOL);
            end else if (op == symbolic_types_pkg::OP_JZ && imm >= ROM_DEPTH) begin
              do_fault(symbolic_types_pkg::F_BAD_BRANCH);
            end else begin
              if (op == symbolic_types_pkg::OP_JZ && top0_w.payload == 32'd0)
                npc_d = imm;
              else
                npc_d = pc_q + 1'b1;
              ndepth_d = depth_q - 1;
              state_d  = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_SWAP: begin
            if (depth_q < 2) begin
              do_fault(symbolic_types_pkg::F_STACK_UNDERFLOW);
            end else begin
              wrA_en_d   = 1'b1;   // position NOS <- old TOS
              wrA_addr_d = depth_q - 2;
              wrA_data_d = top0_w;
              wrB_en_d   = 1'b1;   // position TOS <- old NOS
              wrB_addr_d = depth_q - 1;
              wrB_data_d = top1_w;
              npc_d      = pc_q + 1'b1;
              state_d    = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_JMP: begin
            if (imm >= ROM_DEPTH) begin
              do_fault(symbolic_types_pkg::F_BAD_BRANCH);
            end else begin
              npc_d   = imm;
              state_d = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_EMIT: begin
            if (depth_q == 0) begin
              do_fault(symbolic_types_pkg::F_STACK_UNDERFLOW);
            end else begin
              pending_d = top0_w;
              state_d   = S_OUTPUT_WAIT;
            end
          end

          symbolic_types_pkg::OP_HALT: begin
            halt_stage_d = 1'b1;
            npc_d        = pc_q;      // pc stays at HALT
            ndepth_d     = depth_q;
            state_d      = S_COMMIT;
          end

          symbolic_types_pkg::OP_CALL: begin
            if (rdepth_q == RSTACK_DEPTH[$clog2(RSTACK_DEPTH+1)-1:0]) begin
              do_fault(symbolic_types_pkg::F_RSTACK_OVERFLOW);
            end else if (imm >= ROM_DEPTH) begin
              do_fault(symbolic_types_pkg::F_BAD_BRANCH);
            end else begin
              wrR_en_d   = 1'b1;
              wrR_addr_d = rdepth_q[$clog2(RSTACK_DEPTH)-1:0];
              wrR_data_d = pc_q + 1'b1;
              nrdepth_d  = rdepth_q + 1'b1;
              npc_d      = imm;
              ndepth_d   = depth_q;
              state_d    = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_RET: begin
            if (rdepth_q == 0) begin
              do_fault(symbolic_types_pkg::F_RSTACK_UNDERFLOW);
            end else begin
              npc_d    = rstack_q[rdepth_q-1];
              nrdepth_d = rdepth_q - 1'b1;
              ndepth_d = depth_q;
              state_d  = S_COMMIT;
            end
          end

          default: begin
            do_fault(symbolic_types_pkg::F_BAD_OPCODE);
          end
        endcase
      end

      // The single mutation owner: apply staged state, pulse the trace.
      S_COMMIT: begin
        rdepth_d = nrdepth_q;
        pc_d    = npc_q;
        depth_d = ndepth_q;
        seq_d   = seq_q + 1'b1;
        trace_valid_d  = 1'b1;
        trace_seq_d    = seq_q;
        trace_pc_old_d = pc_q;
        trace_pc_new_d = npc_q;
        trace_op_d     = op;
        trace_depth_d  = ndepth_q;
        trace_event_d  = symbolic_types_pkg::EV_COMMIT;
        if (halt_stage_q) begin
          halted_d = 1'b1;
          state_d  = S_HALTED;
        end else begin
          state_d = S_FETCH;
        end
      end

      // EMIT: hold the value stable until the channel accepts it; the pop
      // and the pc advance happen only on the acceptance edge.
      S_OUTPUT_WAIT: begin
        out_valid = 1'b1;
        if (out_ready) begin
          depth_d        = depth_q - 1'b1;
          pc_d           = pc_q + 1'b1;
          seq_d          = seq_q + 1'b1;
          trace_valid_d  = 1'b1;
          trace_seq_d    = seq_q;
          trace_pc_old_d = pc_q;
          trace_pc_new_d = pc_q + 1'b1;
          trace_op_d     = op;
          trace_depth_d  = depth_q - 1'b1;
          trace_event_d  = symbolic_types_pkg::EV_OUTPUT;
          trace_out_tag_d     = pending_q.tag;
          trace_out_payload_d = pending_q.payload;
          state_d        = S_FETCH;
        end
      end

      S_FAULT: begin
        // Precise fault latched in EXECUTE (pulse fired there); stay here.
        state_d = S_FAULT;
      end

      S_HALTED: begin
        state_d = S_HALTED;
      end

      default: state_d = S_RESET;
    endcase
  end

  // Stage a precise fault and fire the FAULT trace pulse in the same cycle
  // (called from the EXECUTE comb block only).
  task automatic do_fault(input logic [3:0] code);
    fv_d     = 1'b1;
    fc_d     = code;
    fpc_d    = pc_q;
    fop_d    = op;
    fdepth_d = depth_q;
    ft1_d    = top1_w.tag;
    ft0_d    = top0_w.tag;
    state_d  = S_FAULT;
    seq_d         = seq_q + 1'b1;
    trace_valid_d = 1'b1;
    trace_seq_d   = seq_q;
    trace_pc_old_d = pc_q;
    trace_pc_new_d = pc_q;      // precise: pc unchanged
    trace_op_d     = op;
    trace_depth_d  = depth_q;   // precise: depth unchanged
    trace_event_d  = symbolic_types_pkg::EV_FAULT;
    trace_fault_d  = code;
    trace_tag1_d   = top1_w.tag;
    trace_tag0_d   = top0_w.tag;
  endtask

  // ------------------------------------------------------ one mutation owner
  always_ff @(posedge clk or negedge rst_n) begin
    if (!rst_n) begin
      trace_fetch_q <= 1'b0;
      fault_fetch_q <= 1'b0;
      state_q       <= S_RESET;
      pc_q          <= '0;
      depth_q       <= '0;
      ir_q          <= '0;
      seq_q         <= '0;
      halted_q      <= 1'b0;
      rdepth_q      <= '0;
      nrdepth_q <= '0;
      wrR_en_q      <= 1'b0;
      wrR_addr_q    <= '0;
      wrR_data_q    <= '0;
      npc_q         <= '0;
      ndepth_q      <= '0;
      wrA_en_q      <= 1'b0;
      wrB_en_q      <= 1'b0;
      wrA_addr_q    <= '0;
      wrB_addr_q    <= '0;
      wrA_data_q    <= '0;
      wrB_data_q    <= '0;
      pending_q     <= '0;
      halt_stage_q  <= 1'b0;
      fv_q          <= 1'b0;
      fc_q          <= symbolic_types_pkg::F_NONE;
      fpc_q         <= '0;
      fop_q         <= '0;
      fdepth_q      <= '0;
      ft1_q         <= '0;
      ft0_q         <= '0;
      trace_valid_q <= 1'b0;
      // stack_q contents are NOT reset: depth_q = 0 invalidates them
      // (book ch. 15: "Reset invalidates occupancy; it does not need to
      // clear data_q because invalid data is never observed").
    end else begin
      state_q       <= state_d;
      trace_fetch_q <= trace_fetch_d;
      fault_fetch_q <= fault_fetch_d;
      pc_q          <= pc_d;
      depth_q       <= depth_d;
      ir_q          <= ir_d;
      seq_q         <= seq_d;
      halted_q      <= halted_d;
      rdepth_q      <= rdepth_d;
      nrdepth_q <= nrdepth_d;
      wrR_en_q      <= wrR_en_d;
      wrR_addr_q    <= wrR_addr_d;
      wrR_data_q    <= wrR_data_d;
      npc_q         <= npc_d;
      ndepth_q      <= ndepth_d;
      wrA_en_q      <= wrA_en_d;
      wrB_en_q      <= wrB_en_d;
      wrA_addr_q    <= wrA_addr_d;
      wrB_addr_q    <= wrB_addr_d;
      wrA_data_q    <= wrA_data_d;
      wrB_data_q    <= wrB_data_d;
      pending_q     <= pending_d;
      halt_stage_q  <= halt_stage_d;
      fv_q          <= fv_d;
      fc_q          <= fc_d;
      fpc_q         <= fpc_d;
      fop_q         <= fop_d;
      fdepth_q      <= fdepth_d;
      ft1_q         <= ft1_d;
      ft0_q         <= ft0_d;
      trace_valid_q <= trace_valid_d;
      trace_seq_q         <= trace_seq_d;
      trace_pc_old_q      <= trace_pc_old_d;
      trace_pc_new_q      <= trace_pc_new_d;
      trace_op_q          <= trace_op_d;
      trace_depth_q       <= trace_depth_d;
      trace_event_q       <= trace_event_d;
      trace_fault_q       <= trace_fault_d;
      trace_tag1_q        <= trace_tag1_d;
      trace_tag0_q        <= trace_tag0_d;
      trace_out_tag_q     <= trace_out_tag_d;
      trace_out_payload_q <= trace_out_payload_d;

      // Stack writes happen only in COMMIT (the single mutation owner).
      if (state_q == S_COMMIT) begin
        if (wrA_en_q) stack_q[wrA_addr_q] <= wrA_data_q;
        if (wrB_en_q) stack_q[wrB_addr_q] <= wrB_data_q;
        if (wrR_en_q) rstack_q[wrR_addr_q] <= wrR_data_q;
      end
    end
  end

  assign trace_fetch = trace_fetch_q;
  assign fault_fetch = fault_fetch_q;

  initial begin
    if (ROM_DEPTH < 2 || ROM_DEPTH > 32768 || STACK_DEPTH < 2 || RSTACK_DEPTH < 2)
      $fatal(1, "unsupported core memory parameters");
  end

  // ------------------------------------------------------------- outputs
  assign trace_seq          = trace_seq_q;
  assign trace_pc_old       = trace_pc_old_q;
  assign trace_pc_new       = trace_pc_new_q;
  assign trace_op           = trace_op_q;
  assign trace_depth        = trace_depth_q;
  assign trace_event        = trace_event_q;
  assign trace_fault        = trace_fault_q;
  assign trace_tag1         = trace_tag1_q;
  assign trace_tag0         = trace_tag0_q;
  assign trace_out_tag      = trace_out_tag_q;
  assign trace_out_payload  = trace_out_payload_q;
  assign trace_valid        = trace_valid_q;

  always_comb begin
    out_data = pending_q;
  end

  assign fault_valid = fv_q;
  assign fault_code  = fc_q;
  assign fault_pc    = fpc_q;
  assign fault_op    = fop_q;
  assign fault_depth = fdepth_q;
  assign fault_tag1  = ft1_q;
  assign fault_tag0  = ft0_q;
  assign halted      = halted_q;
  assign pc_o        = pc_q;
  assign depth_o     = depth_q;
  assign rdepth_o    = rdepth_q;

endmodule

`default_nettype wire
