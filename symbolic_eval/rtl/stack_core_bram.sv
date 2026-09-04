// stack_core_bram.sv — Laboratory 1 tagged stack evaluator, BRAM version
// with a two-entry top cache (book Lab 1, "BRAM stack with two-entry top
// cache"; Split-Lifetime Frame pattern).
//
// Representation:
//   top0_q       newest value            (tc_q >= 1)
//   top1_q       next value              (tc_q == 2)
//   deep RAM     older values, newest at RAM[dc_q-1]
//   tc_q         0, 1, or 2 cached values
//   dc_q         live values in RAM
//   dt_q         first free RAM address (== dc_q; kept per the book)
//
// Invariant:  architectural depth == tc_q + dc_q, and the logical stack
// from newest to oldest is top0, top1, RAM[dc-1], RAM[dc-2], ...
// The invariant is updated atomically at each commit point (COMMIT state or
// EMIT acceptance); RAM reads that repair the representation are internal
// stuttering and never let the invariant dangle across a trace pulse.
//
// Same cycle-level contract and check order as stack_core.sv (register
// version); only the stack representation differs. The differential tests
// run both cores against the same model traces.
`default_nettype none

module stack_core_bram #(
  parameter int ROM_DEPTH  = 1024,    // 1K x 20 instruction ROM
  parameter int DEEP_DEPTH  = 512,   // deep stack slots in RAM
  parameter int RSTACK_DEPTH = 16    // return-address stack (registers)
) (
  input  logic clk,
  input  logic rst_n,

  output logic [$clog2(ROM_DEPTH)-1:0] rom_addr,
  input  logic [19:0] rom_data,

  output logic        trace_valid,
  output logic [31:0] trace_seq,
  output logic [$clog2(ROM_DEPTH)-1:0] trace_pc_old,
  output logic [$clog2(ROM_DEPTH)-1:0] trace_pc_new,
  output logic [4:0]  trace_op,
  output logic [$clog2(DEEP_DEPTH+3)-1:0] trace_depth,
  output logic [1:0]  trace_event,
  output logic [3:0]  trace_fault,
  output logic [3:0]  trace_tag1,
  output logic [3:0]  trace_tag0,
  output logic [3:0]  trace_out_tag,
  output logic [31:0] trace_out_payload,

  output logic        out_valid,
  output logic [39:0] out_data,
  input  logic        out_ready,

  output logic        fault_valid,
  output logic [3:0]  fault_code,
  output logic [$clog2(ROM_DEPTH)-1:0] fault_pc,
  output logic [4:0]  fault_op,
  output logic [$clog2(DEEP_DEPTH+3)-1:0] fault_depth,
  output logic [3:0]  fault_tag1,
  output logic [3:0]  fault_tag0,

  output logic        halted,
  output logic [$clog2(ROM_DEPTH)-1:0] pc_o,
  output logic [$clog2(DEEP_DEPTH+3)-1:0] depth_o,
  output logic [$clog2(RSTACK_DEPTH+1)-1:0] rdepth_o,

  // debug observation (tb asserts depth_o == tc_o + dc_o)
  output logic [1:0]  tc_o,
  output logic [$clog2(DEEP_DEPTH+1)-1:0] dc_o
);

  localparam int TOTAL_DEPTH = DEEP_DEPTH + 2;

  typedef enum logic [3:0] {
    S_RESET, S_FETCH, S_FETCH_WAIT, S_DECODE, S_RDWAIT,
    S_EXECUTE, S_COMMIT, S_OUTPUT_WAIT, S_FAULT, S_HALTED
  } state_t;

  // read purposes (why RAM[dc-1] or RAM[dc-2] was requested)
  typedef enum logic [2:0] {
    R_NONE, R_OPND, R_FILL, R_POP, R_FILL2
  } rpurpose_t;

  state_t state_q, state_d;
  rpurpose_t rp_q, rp_d;

  // ------------------------------------------------- architectural state
  logic [$clog2(ROM_DEPTH)-1:0]     pc_q, pc_d;
  logic [$clog2(TOTAL_DEPTH+1)-1:0]  depth_q, depth_d;
  logic [19:0]                        ir_q, ir_d;
  logic [31:0]                        seq_q, seq_d;
  logic                               halted_q, halted_d;

  // top cache + deep region
  symbolic_types_pkg::value40_t top0_q;
  symbolic_types_pkg::value40_t top0_d;
  symbolic_types_pkg::value40_t top1_q;
  symbolic_types_pkg::value40_t top1_d;
  logic [1:0] tc_q, tc_d;                       // 0, 1, 2
  logic [$clog2(DEEP_DEPTH+1)-1:0] dc_q, dc_d;  // live deep values
  logic [$clog2(DEEP_DEPTH+1)-1:0] dt_q, dt_d; // first free address (== dc)

  // captured RAM reads
  symbolic_types_pkg::value40_t opnd_q;
  symbolic_types_pkg::value40_t opnd_d;   // operand a fetched from RAM
  symbolic_types_pkg::value40_t rf_q;
  symbolic_types_pkg::value40_t rf_d;       // refill value for top1 after a binary op
  symbolic_types_pkg::value40_t pf_q;
  symbolic_types_pkg::value40_t pf_d;       // refill value for top0 after a pop

  // staged next-state (EXECUTE -> COMMIT)
  logic [$clog2(ROM_DEPTH)-1:0]     npc_q, npc_d;
  logic [$clog2(TOTAL_DEPTH+1)-1:0] ndepth_q, ndepth_d;
  logic [1:0]  stc_q, stc_d;         // staged tc
  logic [$clog2(DEEP_DEPTH+1)-1:0] sdc_q, sdc_d;  // staged dc
  symbolic_types_pkg::value40_t sval_q;
  symbolic_types_pkg::value40_t sval_d;       // pushed value / binary result
  logic        srefill_q, srefill_d; // top1 <= rf_q at COMMIT
  logic [1:0]  spop_q, spop_d;       // pop action: 0 shift, 1 refill, 2 zero
  logic        sswapread_q, sswapread_d;
  logic        halt_stage_q, halt_stage_d;
  symbolic_types_pkg::value40_t pending_q;
  symbolic_types_pkg::value40_t pending_d;

  // return-address stack (continuation state; book Lab 1 extension)
  logic [$clog2(ROM_DEPTH)-1:0]       rstack_q [0:RSTACK_DEPTH-1];
  logic [$clog2(RSTACK_DEPTH+1)-1:0]  rdepth_q, rdepth_d;
  logic                               wrR_en_q, wrR_en_d;
  logic [$clog2(RSTACK_DEPTH)-1:0]   wrR_addr_q, wrR_addr_d;
  logic [$clog2(ROM_DEPTH)-1:0]      wrR_data_q, wrR_data_d;

  // fault record
  logic fv_q, fv_d;
  logic [3:0] fc_q, fc_d;
  logic [$clog2(ROM_DEPTH)-1:0] fpc_q, fpc_d;
  logic [4:0] fop_q, fop_d;
  logic [$clog2(TOTAL_DEPTH+1)-1:0] fdepth_q, fdepth_d;
  logic [3:0] ft1_q, ft1_d, ft0_q, ft0_d;

  // trace snapshot registers
  logic        trace_valid_q, trace_valid_d;
  logic [31:0] trace_seq_q, trace_seq_d;
  logic [$clog2(ROM_DEPTH)-1:0] trace_pc_old_q, trace_pc_old_d;
  logic [$clog2(ROM_DEPTH)-1:0] trace_pc_new_q, trace_pc_new_d;
  logic [4:0]  trace_op_q, trace_op_d;
  logic [$clog2(TOTAL_DEPTH+1)-1:0] trace_depth_q, trace_depth_d;
  logic [1:0]  trace_event_q, trace_event_d;
  logic [3:0]  trace_fault_q, trace_fault_d;
  logic [3:0]  trace_tag1_q, trace_tag1_d, trace_tag0_q, trace_tag0_d;
  logic [3:0]  trace_out_tag_q, trace_out_tag_d;
  logic [31:0] trace_out_payload_q, trace_out_payload_d;

  // ------------------------------------------------------------- the RAM
  logic                                   ram_wr_en;
  logic [$clog2(DEEP_DEPTH)-1:0]          ram_wr_addr;
  logic [39:0]                            ram_wr_data;
  logic [$clog2(DEEP_DEPTH)-1:0]           ram_rd_addr;
  logic [39:0]                            ram_rd_data;

  sync_sdp_ram #(
    .DEPTH(DEEP_DEPTH),
    .WIDTH(40)
  ) u_deep (
    .clk     (clk),
    .wr_en   (ram_wr_en),
    .wr_addr (ram_wr_addr),
    .wr_data (ram_wr_data),
    .rd_addr (ram_rd_addr),
    .rd_data (ram_rd_data)
  );

  // ------------------------------------------------------- decode helpers
  logic [4:0]  op;
  logic [14:0] imm;
  assign op  = ir_q[19:15];
  assign imm = ir_q[14:0];

  symbolic_types_pkg::value40_t a_w;
  symbolic_types_pkg::value40_t b_w;   // operand a (NOS), operand b (TOS)
  always_comb begin
    b_w = top0_q;
    if (tc_q == 2'd2)
      a_w = top1_q;
    else
      a_w = opnd_q;     // fetched from RAM[dc-1] when tc == 1
  end

  logic signed [63:0] add_w, sub_w, mul_w;
  logic add_ovf, sub_ovf, mul_ovf;
  always_comb begin
    add_w = 64'($signed(a_w.payload)) + 64'($signed(b_w.payload));
    sub_w = 64'($signed(a_w.payload)) - 64'($signed(b_w.payload));
    mul_w = 64'($signed(a_w.payload)) * 64'($signed(b_w.payload));
    add_ovf = (add_w[63:32] != {32{add_w[31]}});
    sub_ovf = (sub_w[63:32] != {32{sub_w[31]}});
    mul_ovf = (mul_w[63:32] != {32{mul_w[31]}});
  end

  // ------------------------------------------------------------ the FSM
  always_comb begin
    // defaults: hold everything
    state_d = state_q;
    rp_d    = rp_q;
    pc_d = pc_q; depth_d = depth_q; ir_d = ir_q; seq_d = seq_q;
    halted_d = halted_q;
    top0_d = top0_q; top1_d = top1_q;
    tc_d = tc_q; dc_d = dc_q; dt_d = dt_q;
    opnd_d = opnd_q; rf_d = rf_q; pf_d = pf_q;
    npc_d = npc_q; ndepth_d = ndepth_q;
    stc_d = stc_q; sdc_d = sdc_q;
    sval_d = sval_q;
    srefill_d = srefill_q;
    spop_d = spop_q;
    sswapread_d = sswapread_q;
    halt_stage_d = halt_stage_q;
    pending_d = pending_q;
    rdepth_d   = rdepth_q;
    wrR_en_d   = 1'b0;
    wrR_addr_d = wrR_addr_q;
    wrR_data_d = wrR_data_q;
    fv_d = fv_q; fc_d = fc_q; fpc_d = fpc_q; fop_d = fop_q;
    fdepth_d = fdepth_q; ft1_d = ft1_q; ft0_d = ft0_q;
    trace_valid_d = 1'b0;
    trace_seq_d = trace_seq_q;
    trace_pc_old_d = trace_pc_old_q;
    trace_pc_new_d = trace_pc_new_q;
    trace_op_d = trace_op_q;
    trace_depth_d = trace_depth_q;
    trace_event_d = trace_event_q;
    trace_fault_d = trace_fault_q;
    trace_tag1_d = trace_tag1_q;
    trace_tag0_d = trace_tag0_q;
    trace_out_tag_d = trace_out_tag_q;
    trace_out_payload_d = trace_out_payload_q;
    rom_addr = pc_q;
    out_valid = 1'b0;
    ram_wr_en = 1'b0;
    ram_wr_addr = '0;
    ram_wr_data = '0;
    ram_rd_addr = (dc_q != 0) ? dc_q - 1'b1 : '0;  // guarded default

    case (state_q)
      S_RESET: begin
        if (rst_n) state_d = S_FETCH;
      end

      S_FETCH: state_d = S_FETCH_WAIT;

      S_FETCH_WAIT: begin
        ir_d = rom_data;
        state_d = S_DECODE;
      end

      // Decide whether the representation needs a RAM read before EXECUTE.
      // Reads are side-effect-free, so prefetching before the instruction's
      // checks can never change architectural behavior or fault records.
      S_DECODE: begin
        rp_d = R_NONE;
        if (((op == symbolic_types_pkg::OP_ADD) || (op == symbolic_types_pkg::OP_SUB) || (op == symbolic_types_pkg::OP_MUL) ||
             (op == symbolic_types_pkg::OP_EQ) || (op == symbolic_types_pkg::OP_LT) || (op == symbolic_types_pkg::OP_SWAP)) &&
            (tc_q == 2'd1) && (dc_q != 0)) begin
          rp_d = R_OPND;             // operand a lives at RAM[dc-1]
          ram_rd_addr = dc_q - 1'b1;
          state_d = S_RDWAIT;
        end else if (((op == symbolic_types_pkg::OP_ADD) || (op == symbolic_types_pkg::OP_SUB) || (op == symbolic_types_pkg::OP_MUL) ||
                      (op == symbolic_types_pkg::OP_EQ) || (op == symbolic_types_pkg::OP_LT)) &&
                     (tc_q == 2'd2) && (dc_q != 0)) begin
          rp_d = R_FILL;             // post-op top1 lives at RAM[dc-1]
          ram_rd_addr = dc_q - 1'b1;
          state_d = S_RDWAIT;
        end else if (((op == symbolic_types_pkg::OP_DROP) || (op == symbolic_types_pkg::OP_JZ) || (op == symbolic_types_pkg::OP_EMIT)) &&
                     (tc_q == 2'd1) && (dc_q != 0)) begin
          rp_d = R_POP;              // post-pop top0 lives at RAM[dc-1]
          ram_rd_addr = dc_q - 1'b1;
          state_d = S_RDWAIT;
        end else begin
          state_d = S_EXECUTE;
        end
      end

      // Capture the read data (valid in this cycle) and continue.
      S_RDWAIT: begin
        unique case (rp_q)
          R_OPND:  opnd_d = ram_rd_data;
          R_FILL:  rf_d = ram_rd_data;
          R_POP:   pf_d = ram_rd_data;
          R_FILL2: rf_d = ram_rd_data;
          default: ;
        endcase
        rp_d = R_NONE;
        if (rp_q == R_FILL2)
          state_d = S_COMMIT;
        else
          state_d = S_EXECUTE;
      end

      // All precondition checks; stage complete next-state data. No
      // architectural mutation happens here.
      S_EXECUTE: begin
        case (op)
          symbolic_types_pkg::OP_PUSH_S15, symbolic_types_pkg::OP_PUSH_TRUE, symbolic_types_pkg::OP_PUSH_FALSE, symbolic_types_pkg::OP_DUP: begin
            if (op == symbolic_types_pkg::OP_DUP && depth_q == 0) begin
              do_fault(symbolic_types_pkg::F_STACK_UNDERFLOW);
            end else if (depth_q == TOTAL_DEPTH[$clog2(TOTAL_DEPTH+1)-1:0]) begin
              do_fault(symbolic_types_pkg::F_STACK_OVERFLOW);
            end else begin
              sval_d = (op == symbolic_types_pkg::OP_PUSH_S15)  ? symbolic_types_pkg::mk_int(symbolic_types_pkg::sx15(imm)) :
                       (op == symbolic_types_pkg::OP_PUSH_TRUE) ? symbolic_types_pkg::mk_bool(1'b1)    :
                       (op == symbolic_types_pkg::OP_PUSH_FALSE)? symbolic_types_pkg::mk_bool(1'b0)    :
                                               top0_q;          // DUP
              npc_d = pc_q + 1'b1;
              ndepth_d = depth_q + 1'b1;
              if (tc_q == 2'd2) begin          // spill: old top1 goes to RAM
                stc_d = 2'd2;
                sdc_d = dc_q + 1'b1;
              end else begin
                stc_d = tc_q + 2'd1;
                sdc_d = dc_q;
              end
              state_d = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_ADD, symbolic_types_pkg::OP_SUB, symbolic_types_pkg::OP_MUL, symbolic_types_pkg::OP_EQ, symbolic_types_pkg::OP_LT: begin
            if (depth_q < 2) begin
              do_fault(symbolic_types_pkg::F_STACK_UNDERFLOW);
            end else if (op != symbolic_types_pkg::OP_EQ &&
                         (a_w.tag != symbolic_types_pkg::TAG_INT || b_w.tag != symbolic_types_pkg::TAG_INT)) begin
              do_fault(symbolic_types_pkg::F_TYPE_FAULT);
            end else if (op == symbolic_types_pkg::OP_ADD && add_ovf) begin
              do_fault(symbolic_types_pkg::F_ARITH_OVERFLOW);
            end else if (op == symbolic_types_pkg::OP_SUB && sub_ovf) begin
              do_fault(symbolic_types_pkg::F_ARITH_OVERFLOW);
            end else if (op == symbolic_types_pkg::OP_MUL && mul_ovf) begin
              do_fault(symbolic_types_pkg::F_ARITH_OVERFLOW);
            end else begin
              unique case (op)
                symbolic_types_pkg::OP_ADD: sval_d = symbolic_types_pkg::mk_int(add_w[31:0]);
                symbolic_types_pkg::OP_SUB: sval_d = symbolic_types_pkg::mk_int(sub_w[31:0]);
                symbolic_types_pkg::OP_MUL: sval_d = symbolic_types_pkg::mk_int(mul_w[31:0]);
                symbolic_types_pkg::OP_EQ:  sval_d = symbolic_types_pkg::mk_bool(a_w == b_w);
                symbolic_types_pkg::OP_LT:  sval_d = symbolic_types_pkg::mk_bool($signed(a_w.payload) <
                                         $signed(b_w.payload));
                default: ;
              endcase
              npc_d = pc_q + 1'b1;
              ndepth_d = depth_q - 1'b1;
              if (tc_q == 2'd2) begin
                // operands cached; refill from RAM if the deep region is
                // non-empty (rf_q already captured at DECODE)
                srefill_d = (dc_q != 0);
                stc_d = (dc_q != 0) ? 2'd2 : 2'd1;
                sdc_d = (dc_q != 0) ? dc_q - 1'b1 : dc_q;
                state_d = S_COMMIT;
              end else begin
                // operand a came from RAM; the new top1 lives at RAM[dc-2]
                // when dc >= 2 (read issued now, applied at COMMIT)
                if (dc_q >= 2) begin
                  rp_d = R_FILL2;
                  ram_rd_addr = dc_q - 2'd2;
                  srefill_d = 1'b1;
                  stc_d = 2'd2;
                  sdc_d = dc_q - 2'd2;
                  state_d = S_RDWAIT;
                end else begin
                  srefill_d = 1'b0;
                  stc_d = 2'd1;
                  sdc_d = 2'd0;
                  state_d = S_COMMIT;
                end
              end
            end
          end

          symbolic_types_pkg::OP_DROP, symbolic_types_pkg::OP_JZ: begin
            if (depth_q == 0) begin
              do_fault(symbolic_types_pkg::F_STACK_UNDERFLOW);
            end else if (op == symbolic_types_pkg::OP_JZ && top0_q.tag != symbolic_types_pkg::TAG_BOOL) begin
              do_fault(symbolic_types_pkg::F_TYPE_FAULT);
            end else if (op == symbolic_types_pkg::OP_JZ && top0_q.payload > 32'd1) begin
              do_fault(symbolic_types_pkg::F_NONCANON_BOOL);
            end else if (op == symbolic_types_pkg::OP_JZ && imm >= ROM_DEPTH[14:0]) begin
              do_fault(symbolic_types_pkg::F_BAD_BRANCH);
            end else begin
              // stage the pop action from the representation
              if (tc_q == 2'd2) begin
                spop_d = 2'd0;               // shift top1 into top0
                stc_d = 2'd1;
                sdc_d = dc_q;
              end else if (dc_q != 0) begin
                spop_d = 2'd1;                // top0 <= pf_q, shrink deep
                stc_d = 2'd1;
                sdc_d = dc_q - 1'b1;
              end else begin
                spop_d = 2'd2;                // cache empties
                stc_d = 2'd0;
                sdc_d = dc_q;
              end
              if (op == symbolic_types_pkg::OP_JZ && top0_q.payload == 32'd0)
                npc_d = imm;
              else
                npc_d = pc_q + 1'b1;
              ndepth_d = depth_q - 1'b1;
              state_d = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_SWAP: begin
            if (depth_q < 2) begin
              do_fault(symbolic_types_pkg::F_STACK_UNDERFLOW);
            end else begin
              sswapread_d = (tc_q == 2'd1);
              if (tc_q == 2'd1) begin
                stc_d = 2'd2;      // the RAM operand becomes top0
                sdc_d = dc_q - 1'b1;
              end else begin
                stc_d = 2'd2;
                sdc_d = dc_q;
              end
              npc_d = pc_q + 1'b1;
              ndepth_d = depth_q;
              state_d = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_JMP: begin
            if (imm >= ROM_DEPTH[14:0]) begin
              do_fault(symbolic_types_pkg::F_BAD_BRANCH);
            end else begin
              npc_d = imm;
              ndepth_d = depth_q;
              stc_d = tc_q;
              sdc_d = dc_q;
              state_d = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_EMIT: begin
            if (depth_q == 0) begin
              do_fault(symbolic_types_pkg::F_STACK_UNDERFLOW);
            end else begin
              pending_d = top0_q;
              state_d = S_OUTPUT_WAIT;
            end
          end

          symbolic_types_pkg::OP_HALT: begin
            halt_stage_d = 1'b1;
            npc_d = pc_q;
            ndepth_d = depth_q;
            stc_d = tc_q;
            sdc_d = dc_q;
            state_d = S_COMMIT;
          end

          symbolic_types_pkg::OP_CALL: begin
            if (rdepth_q == RSTACK_DEPTH[$clog2(RSTACK_DEPTH+1)-1:0]) begin
              do_fault(symbolic_types_pkg::F_RSTACK_OVERFLOW);
            end else if (imm >= ROM_DEPTH[14:0]) begin
              do_fault(symbolic_types_pkg::F_BAD_BRANCH);
            end else begin
              wrR_en_d   = 1'b1;
              wrR_addr_d = rdepth_q[$clog2(RSTACK_DEPTH)-1:0];
              wrR_data_d = pc_q + 1'b1;
              rdepth_d   = rdepth_q + 1'b1;
              npc_d      = imm;
              ndepth_d   = depth_q;
              stc_d      = tc_q;
              sdc_d      = dc_q;
              state_d    = S_COMMIT;
            end
          end

          symbolic_types_pkg::OP_RET: begin
            if (rdepth_q == 0) begin
              do_fault(symbolic_types_pkg::F_RSTACK_UNDERFLOW);
            end else begin
              npc_d    = rstack_q[rdepth_q-1];
              rdepth_d = rdepth_q - 1'b1;
              ndepth_d = depth_q;
              stc_d    = tc_q;
              sdc_d    = dc_q;
              state_d  = S_COMMIT;
            end
          end

          default: begin
            do_fault(symbolic_types_pkg::F_BAD_OPCODE);
          end
        endcase
      end

      // The single mutation owner (plus EMIT acceptance below): apply the
      // staged representation update and pulse the trace.
      S_COMMIT: begin
        pc_d = npc_q;
        depth_d = ndepth_q;
        seq_d = seq_q + 1'b1;
        trace_valid_d = 1'b1;
        trace_seq_d = seq_q;
        trace_pc_old_d = pc_q;
        trace_pc_new_d = npc_q;
        trace_op_d = op;
        trace_depth_d = ndepth_q;
        trace_event_d = symbolic_types_pkg::EV_COMMIT;
        tc_d = stc_q;
        dc_d = sdc_q;
        dt_d = sdc_q;   // dt mirrors dc (no holes in this design)

        case (op)
          symbolic_types_pkg::OP_PUSH_S15, symbolic_types_pkg::OP_PUSH_TRUE, symbolic_types_pkg::OP_PUSH_FALSE, symbolic_types_pkg::OP_DUP: begin
            unique case (tc_q)
              2'd0: begin top0_d = sval_q; end
              2'd1: begin top1_d = top0_q; top0_d = sval_q; end
              default: begin  // spill old top1 into RAM
                ram_wr_en = 1'b1;
                ram_wr_addr = dc_q[$clog2(DEEP_DEPTH)-1:0];
                ram_wr_data = top1_q;
                top1_d = top0_q;
                top0_d = sval_q;
              end
            endcase
            // staged counts were computed from tc_q == 2 spill case:
            if (tc_q == 2'd2) begin
              dc_d = dc_q + 1'b1;
              dt_d = dc_q + 1'b1;
            end
          end

          symbolic_types_pkg::OP_ADD, symbolic_types_pkg::OP_SUB, symbolic_types_pkg::OP_MUL, symbolic_types_pkg::OP_EQ, symbolic_types_pkg::OP_LT: begin
            top0_d = sval_q;
            if (srefill_q) top1_d = rf_q;
          end

          symbolic_types_pkg::OP_SWAP: begin
            if (sswapread_q) begin
              top1_d = top0_q;
              top0_d = opnd_q;
            end else begin
              top1_d = top0_q;
              top0_d = top1_q;
            end
          end

          symbolic_types_pkg::OP_DROP, symbolic_types_pkg::OP_JZ: begin
            unique case (spop_q)
              2'd0: top0_d = top1_q;    // shift (tc 2 -> 1)
              2'd1: top0_d = pf_q;      // refill from RAM (tc stays 1)
              default: ;                // cache empties
            endcase
          end

          default: ;  // JMP / HALT: no stack change
        endcase

        if (halt_stage_q) begin
          halted_d = 1'b1;
          state_d = S_HALTED;
        end else begin
          state_d = S_FETCH;
        end
      end

      // EMIT: the value is offered (out_valid) and the pop, pc advance, and
      // trace pulse happen only on the acceptance edge (Delayed Irreversible
      // Store). The representation repair (pop refill) uses the prefetched
      // pf_q, so the invariant depth == tc + dc holds at the pulse.
      S_OUTPUT_WAIT: begin
        out_valid = 1'b1;
        if (out_ready) begin
          pc_d = pc_q + 1'b1;
          depth_d = depth_q - 1'b1;
          seq_d = seq_q + 1'b1;
          trace_valid_d = 1'b1;
          trace_seq_d = seq_q;
          trace_pc_old_d = pc_q;
          trace_pc_new_d = pc_q + 1'b1;
          trace_op_d = op;
          trace_depth_d = depth_q - 1'b1;
          trace_event_d = symbolic_types_pkg::EV_OUTPUT;
          trace_out_tag_d = pending_q.tag;
          trace_out_payload_d = pending_q.payload;
          // apply the staged pop action
          if (tc_q == 2'd2) begin
            top0_d = top1_q;
            tc_d = 2'd1;
            dc_d = dc_q;
          end else if (dc_q != 0) begin
            top0_d = pf_q;
            tc_d = 2'd1;
            dc_d = dc_q - 1'b1;
          end else begin
            tc_d = 2'd0;
            dc_d = dc_q;
          end
          dt_d = dc_d;
          state_d = S_FETCH;
        end
      end

      S_FAULT: state_d = S_FAULT;
      S_HALTED: state_d = S_HALTED;
      default: state_d = S_RESET;
    endcase
  end

  // Stage a precise fault and fire the FAULT trace pulse in the same cycle.
  // Operand tags are guarded by depth: the cache registers hold stale
  // values when the stack is shallower than two.
  task automatic do_fault(input logic [3:0] code);
    fv_d = 1'b1;
    fc_d = code;
    fpc_d = pc_q;
    fop_d = op;
    fdepth_d = depth_q;
    ft1_d = (depth_q >= 2) ? a_w.tag : 4'h0;
    ft0_d = (depth_q >= 1) ? b_w.tag : 4'h0;
    state_d = S_FAULT;
    seq_d = seq_q + 1'b1;
    trace_valid_d = 1'b1;
    trace_seq_d = seq_q;
    trace_pc_old_d = pc_q;
    trace_pc_new_d = pc_q;
    trace_op_d = op;
    trace_depth_d = depth_q;
    trace_event_d = symbolic_types_pkg::EV_FAULT;
    trace_fault_d = code;
    trace_tag1_d = (depth_q >= 2) ? a_w.tag : 4'h0;
    trace_tag0_d = (depth_q >= 1) ? b_w.tag : 4'h0;
  endtask

  // ------------------------------------------------------ one mutation owner
  always_ff @(posedge clk or negedge rst_n) begin
    if (!rst_n) begin
      state_q <= S_RESET;
      rp_q <= R_NONE;
      pc_q <= '0;
      depth_q <= '0;
      ir_q <= '0;
      seq_q <= '0;
      halted_q <= 1'b0;
      rdepth_q <= '0;
      wrR_en_q <= 1'b0;
      wrR_addr_q <= '0;
      wrR_data_q <= '0;
      top0_q <= '0;
      top1_q <= '0;
      tc_q <= 2'd0;
      dc_q <= '0;
      dt_q <= '0;
      opnd_q <= '0;
      rf_q <= '0;
      pf_q <= '0;
      npc_q <= '0;
      ndepth_q <= '0;
      stc_q <= 2'd0;
      sdc_q <= '0;
      sval_q <= '0;
      srefill_q <= 1'b0;
      spop_q <= 2'd0;
      sswapread_q <= 1'b0;
      halt_stage_q <= 1'b0;
      pending_q <= '0;
      fv_q <= 1'b0;
      fc_q <= symbolic_types_pkg::F_NONE;
      fpc_q <= '0;
      fop_q <= '0;
      fdepth_q <= '0;
      ft1_q <= '0;
      ft0_q <= '0;
      trace_valid_q <= 1'b0;
      trace_seq_q <= '0;
      trace_pc_old_q <= '0;
      trace_pc_new_q <= '0;
      trace_op_q <= '0;
      trace_depth_q <= '0;
      trace_event_q <= 2'd0;
      trace_fault_q <= symbolic_types_pkg::F_NONE;
      trace_tag1_q <= '0;
      trace_tag0_q <= '0;
      trace_out_tag_q <= '0;
      trace_out_payload_q <= '0;
    end else begin
      state_q <= state_d;
      rp_q <= rp_d;
      pc_q <= pc_d;
      depth_q <= depth_d;
      ir_q <= ir_d;
      seq_q <= seq_d;
      halted_q <= halted_d;
      rdepth_q <= rdepth_d;
      wrR_en_q <= wrR_en_d;
      wrR_addr_q <= wrR_addr_d;
      wrR_data_q <= wrR_data_d;
      top0_q <= top0_d;
      top1_q <= top1_d;
      tc_q <= tc_d;
      dc_q <= dc_d;
      dt_q <= dt_d;
      opnd_q <= opnd_d;
      rf_q <= rf_d;
      pf_q <= pf_d;
      npc_q <= npc_d;
      ndepth_q <= ndepth_d;
      stc_q <= stc_d;
      sdc_q <= sdc_d;
      sval_q <= sval_d;
      srefill_q <= srefill_d;
      spop_q <= spop_d;
      sswapread_q <= sswapread_d;
      halt_stage_q <= halt_stage_d;
      pending_q <= pending_d;
      fv_q <= fv_d;
      fc_q <= fc_d;
      fpc_q <= fpc_d;
      fop_q <= fop_d;
      fdepth_q <= fdepth_d;
      ft1_q <= ft1_d;
      ft0_q <= ft0_d;
      trace_valid_q <= trace_valid_d;
      trace_seq_q <= trace_seq_d;
      trace_pc_old_q <= trace_pc_old_d;
      trace_pc_new_q <= trace_pc_new_d;
      trace_op_q <= trace_op_d;
      trace_depth_q <= trace_depth_d;
      trace_event_q <= trace_event_d;
      trace_fault_q <= trace_fault_d;
      trace_tag1_q <= trace_tag1_d;
      trace_tag0_q <= trace_tag0_d;
      trace_out_tag_q <= trace_out_tag_d;
      trace_out_payload_q <= trace_out_payload_d;

      // Return-stack writes happen only in COMMIT (single mutation owner).
      if (state_q == S_COMMIT && wrR_en_q)
        rstack_q[wrR_addr_q] <= wrR_data_q;
    end
  end

  // ------------------------------------------------------------- outputs
  assign trace_valid = trace_valid_q;
  assign trace_seq = trace_seq_q;
  assign trace_pc_old = trace_pc_old_q;
  assign trace_pc_new = trace_pc_new_q;
  assign trace_op = trace_op_q;
  assign trace_depth = trace_depth_q;
  assign trace_event = trace_event_q;
  assign trace_fault = trace_fault_q;
  assign trace_tag1 = trace_tag1_q;
  assign trace_tag0 = trace_tag0_q;
  assign trace_out_tag = trace_out_tag_q;
  assign trace_out_payload = trace_out_payload_q;

  always_comb begin
    out_data = pending_q;
  end

  assign fault_valid = fv_q;
  assign fault_code = fc_q;
  assign fault_pc = fpc_q;
  assign fault_op = fop_q;
  assign fault_depth = fdepth_q;
  assign fault_tag1 = ft1_q;
  assign fault_tag0 = ft0_q;
  assign halted = halted_q;
  assign pc_o = pc_q;
  assign depth_o = depth_q;
  assign rdepth_o = rdepth_q;
  assign tc_o = tc_q;
  assign dc_o = dc_q;

endmodule

`default_nettype wire
