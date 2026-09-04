#!/usr/bin/env python3
from pathlib import Path
ROOT=next(p for p in Path(__file__).resolve().parents if (p/'symbolic_eval').is_dir())/'symbolic_eval'
for core in ['stack_core','stack_core_bram']:
    p=ROOT/'sim'/f'tb_{core}.sv';s=p.read_text()
    depth='STACK_DEPTH' if core=='stack_core' else '(DEEP_DEPTH+2)'
    s=s.replace('  string hex_file;',f'  string hex_file;\n  string stack_file;\n  integer stack_count = 0;\n  logic [39:0] stack_init [0:{depth}-1];')
    # Injection after reset has settled but before the first instruction.
    setup='''    #1;
    if ($value$plusargs("stack=%s", stack_file)) begin
      void'($value$plusargs("stack_count=%d", stack_count));
      if (stack_count < 0 || stack_count > '''+depth+''') $fatal(1, "bad initial depth");
      $readmemh(stack_file, stack_init);
      dut.depth_q = stack_count;
'''
    if core=='stack_core':
        setup+='      for (integer j=0; j<stack_count; j=j+1) dut.stack_q[j] = stack_init[j];\n'
    else:
        setup+='''      dut.tc_q = (stack_count >= 2) ? 2 : stack_count;
      dut.dc_q = (stack_count >= 2) ? stack_count-2 : 0;
      dut.dt_q = dut.dc_q;
      for (integer j=0; j<dut.dc_q; j=j+1) dut.u_deep.mem[j] = stack_init[j];
      if (stack_count >= 1) dut.top0_q = stack_init[stack_count-1];
      if (stack_count >= 2) dut.top1_q = stack_init[stack_count-2];
'''
    setup+='    end\n'
    assert s.count("    rst_n <= 1'b1;")==1
    s=s.replace("    rst_n <= 1'b1;", "    rst_n <= 1'b1;\n"+setup);p.write_text(s)
