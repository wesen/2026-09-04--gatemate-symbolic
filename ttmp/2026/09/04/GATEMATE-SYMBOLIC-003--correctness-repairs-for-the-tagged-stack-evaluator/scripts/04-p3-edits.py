#!/usr/bin/env python3
"""Widen architectural addresses and wire explicit fetch-fault context."""
from pathlib import Path
import re
ROOT=next(p for p in Path(__file__).resolve().parents if (p/'symbolic_eval').is_dir())/'symbolic_eval'
def one(s,old,new):
    assert s.count(old)==1,(old,s.count(old))
    return s.replace(old,new)
for core in ['stack_core','stack_core_bram']:
    p=ROOT/'rtl'/f'{core}.sv';s=p.read_text()
    s=s.replace('$clog2(ROM_DEPTH)', '$clog2(ROM_DEPTH+1)')
    s=one(s,'$clog2(ROM_DEPTH+1)-1:0] rom_addr','$clog2(ROM_DEPTH)-1:0] rom_addr')
    s=one(s,'  output logic        trace_valid,','  output logic        trace_valid,\n  output logic        trace_fetch,  // opcode is not meaningful on a fetch fault\n  output logic        fault_fetch,')
    s=one(s,'  // fault record','  logic trace_fetch_q, trace_fetch_d, fault_fetch_q, fault_fetch_d;\n\n  // fault record')
    s=one(s,'    case (state_q)',"    trace_fetch_d = 1'b0;\n    fault_fetch_d = fault_fetch_q;\n\n    case (state_q)")
    s,n=re.subn(r'^    rom_addr\s*= pc_q;',"rom_addr = (pc_q < ROM_DEPTH) ? pc_q[$clog2(ROM_DEPTH)-1:0] : '0;",s,flags=re.M);assert n==1
    s=s.replace('imm >= ROM_DEPTH[14:0]','imm >= ROM_DEPTH')
    s=one(s,'    if (!rst_n) begin',"    if (!rst_n) begin\n      trace_fetch_q <= 1'b0;\n      fault_fetch_q <= 1'b0;")
    s,n=re.subn(r'(state_q\s*<= state_d;)',r'\1\n      trace_fetch_q <= trace_fetch_d;\n      fault_fetch_q <= fault_fetch_d;',s,flags=re.M);assert n==1
    s=one(s,'  // ------------------------------------------------------------- outputs','  assign trace_fetch = trace_fetch_q;\n  assign fault_fetch = fault_fetch_q;\n\n  // ------------------------------------------------------------- outputs')
    if core=='stack_core':
        s=one(s,'''      S_FETCH: begin
        state_d = S_FETCH_WAIT;   // rom_addr = pc_q; data next cycle
      end''','''      S_FETCH: begin
        if (pc_q >= ROM_DEPTH) begin
          do_fault(symbolic_types_pkg::F_BAD_BRANCH);
          trace_fetch_d = 1'b1;
          fault_fetch_d = 1'b1;
        end else begin
          state_d = S_FETCH_WAIT;
        end
      end''')
        guard='STACK_DEPTH'
    else:
        s=one(s,'S_OUTPUT_WAIT, S_FAULT, S_HALTED','S_OUTPUT_WAIT, S_FAULT, S_HALTED, S_FETCH_CONTEXT, S_FETCH_FAULT')
        s=one(s,'      S_FETCH: state_d = S_FETCH_WAIT;', '''      S_FETCH: begin
        if (pc_q >= ROM_DEPTH) begin
          if (tc_q == 1 && dc_q != 0) begin
            ram_rd_addr = dc_q - 1'b1;
            state_d = S_FETCH_CONTEXT;
          end else begin
            state_d = S_FETCH_FAULT;
          end
        end else begin
          state_d = S_FETCH_WAIT;
        end
      end

      S_FETCH_CONTEXT: begin
        opnd_d = ram_rd_data;
        state_d = S_FETCH_FAULT;
      end

      S_FETCH_FAULT: begin
        do_fault(symbolic_types_pkg::F_BAD_BRANCH);
        trace_fetch_d = 1'b1;
        fault_fetch_d = 1'b1;
      end''')
        guard='DEEP_DEPTH'
    s=one(s,'  // ------------------------------------------------------------- outputs',f'''  initial begin
    if (ROM_DEPTH < 2 || ROM_DEPTH > 32768 || {guard} < 2 || RSTACK_DEPTH < 2)
      $fatal(1, "unsupported core memory parameters");
  end

  // ------------------------------------------------------------- outputs''')
    p.write_text(s)
    p=ROOT/'sim'/f'tb_{core}.sv';s=p.read_text()
    s=s.replace('$clog2(ROM_DEPTH)', '$clog2(ROM_DEPTH+1)')
    s=one(s,'$clog2(ROM_DEPTH+1)-1:0] rom_addr','$clog2(ROM_DEPTH)-1:0] rom_addr')
    s=one(s,'  logic        trace_valid;','  logic        trace_valid;\n  logic trace_fetch, fault_fetch;')
    s=one(s,'    .trace_valid(trace_valid),','    .trace_fetch(trace_fetch), .fault_fetch(fault_fetch),\n    .trace_valid(trace_valid),')
    s=one(s,'    case (o)','    if (trace_fetch) return "FETCH";\n    case (o)')
    p.write_text(s)
p=ROOT/'rtl/top.sv';s=p.read_text()
s=one(s,'$clog2(ROM_DEPTH)-1:0] pc_dbg','$clog2(ROM_DEPTH+1)-1:0] pc_dbg')
s=one(s,'    .trace_valid(),','    .trace_fetch(),\n    .fault_fetch(),\n    .trace_valid(),');p.write_text(s)
