#!/usr/bin/env python3
"""Checked mechanical edits for P2; retain implementation steps for review."""
from pathlib import Path
import sys
ROOT=next(p for p in Path(__file__).resolve().parents if (p/'symbolic_eval').is_dir())/'symbolic_eval'
def replace(text,old,new):
    assert text.count(old)==1, (old,text.count(old))
    return text.replace(old,new)
if sys.argv[1]=='monitors':
    for core in ['stack_core','stack_core_bram']:
        p=ROOT/'sim'/f'tb_{core}.sv'; s=p.read_text()
        monitor='''  // Return depth is architectural: publish it on retirement only.
  logic [$clog2(16+1)-1:0] rdepth_prev = '0;
  always @(posedge clk) begin
    #2;
    if (rst_n && cycle > 4 && !trace_valid && rdepth_o !== rdepth_prev) begin
      $display("ASSERT_FAIL: return depth changed without retirement");
      errors = errors + 1;
    end
    rdepth_prev = rdepth_o;
  end

'''
        s=replace(s,'endmodule',monitor+'endmodule');p.write_text(s)
else:
    import re
    for core in ['stack_core','stack_core_bram']:
        p=ROOT/'rtl'/f'{core}.sv';s=p.read_text()
        s=replace(s,'  // fault record','  logic [$clog2(RSTACK_DEPTH+1)-1:0] nrdepth_q, nrdepth_d;\n\n  // fault record')
        s=replace(s,'    rdepth_d   = rdepth_q;','    rdepth_d   = rdepth_q;\n    nrdepth_d  = nrdepth_q;')
        s=replace(s,'      S_EXECUTE: begin','      S_EXECUTE: begin\n        nrdepth_d = rdepth_q;')
        s=replace(s,"rdepth_d   = rdepth_q + 1'b1;","nrdepth_d  = rdepth_q + 1'b1;")
        s=replace(s,"rdepth_d = rdepth_q - 1'b1;","nrdepth_d = rdepth_q - 1'b1;")
        s=replace(s,'      S_COMMIT: begin','      S_COMMIT: begin\n        rdepth_d = nrdepth_q;')
        s,n=re.subn(r"(rdepth_q\s*<= '0;)",r"\1\n      nrdepth_q <= '0;",s);assert n==1
        s,n=re.subn(r'(rdepth_q\s*<= rdepth_d;)',r'\1\n      nrdepth_q <= nrdepth_d;',s);assert n==1
        if core.endswith('bram'):
            start=s.index('        if (((op ==',s.index('      S_DECODE: begin'))
            end=s.index("          rp_d = R_OPND;",start)
            s=s[:start]+"        if ((tc_q == 2'd1) && (dc_q != 0)) begin\n"+s[end:]
            start=s.index('        end else if (((op == symbolic_types_pkg::OP_DROP)',s.index('      S_DECODE: begin'))
            end=s.index('        end else begin',start)
            s=s[:start]+s[end:]
            s=replace(s,'          R_OPND:  opnd_d = ram_rd_data;', '''          R_OPND: begin
            // All instructions need accurate top-two fault context;
            // the same read supplies the post-pop top for DROP/JZ/EMIT.
            opnd_d = ram_rd_data;
            pf_d = ram_rd_data;
          end''')
            s=s.replace('R_NONE, R_OPND, R_FILL, R_POP, R_FILL2','R_NONE, R_OPND, R_FILL, R_FILL2').replace('          R_POP:   pf_d = ram_rd_data;\n','')
        p.write_text(s)
