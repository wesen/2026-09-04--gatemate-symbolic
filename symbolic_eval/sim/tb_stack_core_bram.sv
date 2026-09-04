// tb_stack_core_bram.sv — differential testbench for stack_core_bram
// (BRAM + two-entry top cache). Adds the refinement invariant check:
//   architectural depth == tc + dc at every cycle after reset. Prints TRACE/FINAL lines in the exact format of the Python
// reference model (tools/stack_model.py), so sim/test_directed.py can diff
// them line by line.
//
// Plusargs:
//   +rom=path.hex      program image (one 20-bit word per line, 5 hex digits)
//   +max_cycles=N      watchdog (default 200000)
//   +stall_seed=N      randomize out_ready (backpressure); N<0 = never stall
//
// The ROM lives in this testbench (zero-filled, then $readmemh) and is read
// synchronously, matching the core's one-cycle-latency ROM contract.
`timescale 1ns/1ps

module tb_stack_core_bram;


  parameter int ROM_DEPTH   = 1024;
  parameter int DEEP_DEPTH  = 30;   // total stack = DEEP_DEPTH + 2

  logic clk = 1'b0;
  logic rst_n = 1'b0;
  always #50 clk = ~clk;   // 10 MHz

  logic [$clog2(ROM_DEPTH)-1:0] rom_addr;
  logic [19:0] rom_data;
  logic [19:0] rom_mem [0:ROM_DEPTH-1];

  // output channel with optional random backpressure
  logic        out_ready;
  logic        out_valid;
  logic [39:0] out_data;

  logic        trace_valid;
  logic [31:0] trace_seq;
  logic [$clog2(ROM_DEPTH)-1:0] trace_pc_old, trace_pc_new;
  logic [4:0]  trace_op;
  logic [$clog2(DEEP_DEPTH+3)-1:0] trace_depth;
  logic [1:0]  trace_event;
  logic [3:0]  trace_fault;
  logic [3:0]  trace_tag1, trace_tag0, trace_out_tag;
  logic [31:0] trace_out_payload;
  logic        fault_valid;
  logic [3:0]  fault_code;
  logic [$clog2(ROM_DEPTH)-1:0] fault_pc;
  logic [4:0]  fault_op;
  logic [$clog2(DEEP_DEPTH+3)-1:0] fault_depth;
  logic [3:0]  fault_tag1, fault_tag0;
  logic        halted;
  logic [$clog2(ROM_DEPTH)-1:0] pc_o;
  logic [$clog2(16+1)-1:0] rdepth_o;
  logic [$clog2(DEEP_DEPTH+3)-1:0] depth_o;
  logic [1:0] tc_o;
  logic [$clog2(DEEP_DEPTH+1)-1:0] dc_o;

  stack_core_bram #(
    .ROM_DEPTH(ROM_DEPTH),
    .DEEP_DEPTH(DEEP_DEPTH)
  ) dut (
    .clk(clk), .rst_n(rst_n),
    .rom_addr(rom_addr), .rom_data(rom_data),
    .trace_valid(trace_valid), .trace_seq(trace_seq),
    .trace_pc_old(trace_pc_old), .trace_pc_new(trace_pc_new),
    .trace_op(trace_op), .trace_depth(trace_depth),
    .trace_event(trace_event), .trace_fault(trace_fault),
    .trace_tag1(trace_tag1), .trace_tag0(trace_tag0),
    .trace_out_tag(trace_out_tag), .trace_out_payload(trace_out_payload),
    .out_valid(out_valid), .out_data(out_data), .out_ready(out_ready),
    .fault_valid(fault_valid), .fault_code(fault_code),
    .fault_pc(fault_pc), .fault_op(fault_op), .fault_depth(fault_depth),
    .fault_tag1(fault_tag1), .fault_tag0(fault_tag0),
    .halted(halted), .pc_o(pc_o), .depth_o(depth_o),
    .rdepth_o(rdepth_o), .tc_o(tc_o), .dc_o(dc_o)
  );

  // synchronous ROM read (one-cycle latency contract)
  always_ff @(posedge clk) rom_data <= rom_mem[rom_addr];

  // ------------------------------------------------------------- helpers
  function automatic string op_name(input logic [4:0] o);
    case (o)
      5'h00: return "PUSH_S15";
      5'h01: return "PUSH_TRUE";
      5'h02: return "PUSH_FALSE";
      5'h03: return "ADD";
      5'h04: return "SUB";
      5'h05: return "MUL";
      5'h06: return "EQ";
      5'h07: return "LT";
      5'h08: return "DUP";
      5'h09: return "DROP";
      5'h0A: return "SWAP";
      5'h0B: return "JMP";
      5'h0C: return "JZ";
      5'h0D: return "EMIT";
      5'h0E: return "HALT";
      5'h0F: return "CALL";
      5'h10: return "RET";
      default: return "BAD";
    endcase
  endfunction

  function automatic string fault_name(input logic [3:0] f);
    case (f)
      4'd0: return "NONE";
      4'd1: return "STACK_UNDERFLOW";
      4'd2: return "STACK_OVERFLOW";
      4'd3: return "TYPE_FAULT";
      4'd4: return "ARITH_OVERFLOW";
      4'd5: return "BAD_OPCODE";
      4'd6: return "BAD_BRANCH_TARGET";
      4'd7: return "NONCANONICAL_BOOL";
      4'd8: return "RSTACK_UNDERFLOW";
      4'd9: return "RSTACK_OVERFLOW";
      default: return "?";
    endcase
  endfunction

  function automatic string event_name(input logic [1:0] e);
    case (e)
      2'd0: return "COMMIT";
      2'd1: return "OUTPUT";
      2'd2: return "FAULT";
      default: return "?";
    endcase
  endfunction

  // ---------------------------------------------------------- bookkeeping
  integer emit_count = 0;
  integer errors = 0;
  integer max_cycles = 200000;
  integer cycle = 0;
  integer stall_seed = -1;
  integer rand_state = 1;

  // Count accepted outputs (one acceptance = one pop, checked vs FINAL).
  always @(posedge clk) begin
    if (out_valid && out_ready) emit_count = emit_count + 1;
  end

  // Blocked producer must hold the item stable (book ch. 15 assertion).
  // Monitors sample pre-edge values (blocking reads + NBA capture), so they
  // see exactly what the DUT sampled at the same edge - no #1 races.
  logic        pv_valid = 1'b0;
  logic        pv_ready = 1'b1;
  logic [39:0] pv_data = '0;
  always @(posedge clk) begin
    if (rst_n && pv_valid && !pv_ready) begin
      if (out_valid !== 1'b1 || out_data !== pv_data) begin
        $display("ASSERT_FAIL: blocked output not stable (valid=%0b)",
                 out_valid);
        errors = errors + 1;
      end
    end
    pv_valid <= out_valid;
    pv_ready <= out_ready;
    pv_data  <= out_data;
  end

  // pc may only change on a retired instruction (trace pulse) or reset.
  logic [$clog2(ROM_DEPTH)-1:0] pc_prev = '0;
  always @(posedge clk) begin
    if (rst_n && !trace_valid && cycle > 3 && !halted) begin
      if (pc_o !== pc_prev) begin
        $display("ASSERT_FAIL: pc changed without commit (%0d -> %0d)",
                 pc_prev, pc_o);
        errors = errors + 1;
      end
    end
    pc_prev <= pc_o;
  end

  // Refinement invariant (book Lab 1 exercise 2): depth == tc + dc.
  always @(posedge clk) begin
    if (rst_n && cycle > 4) begin
      if (depth_o !== tc_o + dc_o) begin
        $display("ASSERT_FAIL: depth %0d != tc %0d + dc %0d",
                 depth_o, tc_o, dc_o);
        errors = errors + 1;
      end
    end
  end

  // Trace printing (one cycle after the pulse; no back-to-back pulses by
  // construction: minimum 5 cycles per instruction).
  always @(posedge clk) begin
    #1;
    if (trace_valid) begin
      if (trace_event == 2'd2)  // FAULT
        $display("TRACE %0d %0d %0d %s %0d FAULT %s %h %h",
                 trace_seq, trace_pc_old, trace_pc_new, op_name(trace_op),
                 trace_depth, fault_name(trace_fault), trace_tag1, trace_tag0);
      else if (trace_event == 2'd1)  // OUTPUT
        $display("TRACE %0d %0d %0d %s %0d OUTPUT %h %08h",
                 trace_seq, trace_pc_old, trace_pc_new, op_name(trace_op),
                 trace_depth, trace_out_tag, trace_out_payload);
      else
        $display("TRACE %0d %0d %0d %s %0d COMMIT",
                 trace_seq, trace_pc_old, trace_pc_new, op_name(trace_op),
                 trace_depth);
    end
  end

  // Random backpressure on the output channel: a registered stall bit so
  // out_ready is stable within a cycle and sampled identically by the DUT
  // and the monitors.
  logic stall_q = 1'b0;
  always @(posedge clk) begin
    if (stall_seed >= 0)
      stall_q <= (($random(rand_state) & 3) == 0);
    else
      stall_q <= 1'b0;
  end
  assign out_ready = !stall_q;

  // ------------------------------------------------------------- stimulus
  string hex_file;
  initial begin
    if (!$value$plusargs("rom=%s", hex_file)) begin
      $display("FAIL: missing +rom=<file.hex>");
      $finish;
    end
    void'($value$plusargs("max_cycles=%d", max_cycles));
    if ($value$plusargs("stall_seed=%d", stall_seed)) begin
      rand_state = stall_seed;
      if (stall_seed == 0) rand_state = 1;
    end

    for (int i = 0; i < ROM_DEPTH; i++) rom_mem[i] = 20'h00000;
    $readmemh(hex_file, rom_mem);

    repeat (4) @(posedge clk);
    rst_n <= 1'b1;

    // Wait for halt, fault, or watchdog.
    forever begin
      @(posedge clk);
      cycle = cycle + 1;
      #1;
      if (halted || fault_valid) break;
      if (cycle >= max_cycles) begin
        $display("FAIL: watchdog timeout after %0d cycles", cycle);
        errors = errors + 1;
        break;
      end
    end

    repeat (2) @(posedge clk);
    begin
      string fname;
      fname = "NONE";
      if (fault_valid) fname = fault_name(fault_code);
      $display("FINAL %0d %s %0d %0d %0d %0d", halted, fname,
               pc_o, depth_o, emit_count, rdepth_o);
    end
    if (errors == 0)
      $display("TB_PASS");
    else
      $display("TB_FAIL %0d", errors);
    $finish;
  end

  // Return depth is architectural: publish it on retirement only.
  logic [$clog2(16+1)-1:0] rdepth_prev = '0;
  always @(posedge clk) begin
    #2;
    if (rst_n && cycle > 4 && !trace_valid && rdepth_o !== rdepth_prev) begin
      $display("ASSERT_FAIL: return depth changed without retirement");
      errors = errors + 1;
    end
    rdepth_prev = rdepth_o;
  end

endmodule
