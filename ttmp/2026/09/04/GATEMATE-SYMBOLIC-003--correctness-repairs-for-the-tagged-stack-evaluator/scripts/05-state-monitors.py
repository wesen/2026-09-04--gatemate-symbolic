#!/usr/bin/env python3
"""Install simulation-only complete live-state observation and stability checks."""
from pathlib import Path
ROOT=next(p for p in Path(__file__).resolve().parents if (p/'symbolic_eval').is_dir())/'symbolic_eval'
for core in ['stack_core','stack_core_bram']:
    p=ROOT/'sim'/f'tb_{core}.sv';s=p.read_text()
    depth='STACK_DEPTH' if core=='stack_core' else '(DEEP_DEPTH+2)'
    data='''    live_data = dut.stack_q[index];''' if core=='stack_core' else '''    if (index < dut.dc_q) live_data = dut.u_deep.mem[index];
    else if (dut.tc_q == 2 && index == dut.dc_q) live_data = dut.top1_q;
    else live_data = dut.top0_q;'''
    monitor=f'''  // Full abstraction, oldest first. Invalid physical storage is ignored.
  function automatic logic [39:0] live_data(input integer index);
{data}
  endfunction

  logic [39:0] previous_data [0:{depth}-1];
  logic [$clog2(ROM_DEPTH+1)-1:0] previous_returns [0:15];
  integer previous_depth = 0;
  integer previous_rdepth = 0;
  logic [$clog2(ROM_DEPTH+1)-1:0] previous_pc = 0;
  always @(posedge clk) begin
    #2;
    if (rst_n && cycle > 4) begin
      if (!trace_valid || trace_event == 2'd2) begin
        if (pc_o !== previous_pc || depth_o != previous_depth || rdepth_o != previous_rdepth) begin
          $display("ASSERT_FAIL: architectural counts or PC changed without retirement");
          errors = errors + 1;
        end
        for (integer j=0; j<depth_o; j=j+1)
          if (live_data(j) !== previous_data[j]) begin
            $display("ASSERT_FAIL: live data changed without retirement at %0d", j);
            errors = errors + 1;
          end
        for (integer j=0; j<rdepth_o; j=j+1)
          if (dut.rstack_q[j] !== previous_returns[j]) begin
            $display("ASSERT_FAIL: live return address changed without retirement");
            errors = errors + 1;
          end
      end
      if (trace_valid) begin
        $write("STATE %0d PC %0d D %0d", trace_seq, pc_o, depth_o);
        for (integer j=0; j<depth_o; j=j+1) $write(" %010h", live_data(j));
        $write(" R %0d", rdepth_o);
        for (integer j=0; j<rdepth_o; j=j+1) $write(" %0d", dut.rstack_q[j]);
        $display("");
        if (trace_event == 2'd2 && fault_fetch !== trace_fetch) begin
          $display("ASSERT_FAIL: fetch context flags disagree");
          errors = errors + 1;
        end
      end
    end
    previous_pc = pc_o;
    previous_depth = depth_o;
    previous_rdepth = rdepth_o;
    for (integer j=0; j<depth_o; j=j+1) previous_data[j] = live_data(j);
    for (integer j=0; j<rdepth_o; j=j+1) previous_returns[j] = dut.rstack_q[j];
  end

'''
    assert s.count('endmodule')==1
    s=s.replace('endmodule',monitor+'endmodule')
    s=s.replace('if (out_valid && out_ready) emit_count = emit_count + 1;', '''if (rst_n && out_valid && out_ready) begin
      emit_count = emit_count + 1;
      $display("XFER %010h", out_data);
    end''')
    p.write_text(s)
