#!/usr/bin/env python3
"""Seed the separate graph laboratory from the frozen queens controller."""
from pathlib import Path
T=Path(__file__).resolve().parents[1]
repo=next(p for p in T.parents if (p/'queens_rollback').is_dir())
target=repo/'graph_microscope/rtl/graph_core.sv'
assert not target.exists(),'This one-time seed must not overwrite implemented graph RTL'
s=(repo/'queens_rollback/rtl/queens_core.sv').read_text()
s=s.replace('module queens_core','module graph_core')
s=s.replace('parameter bit USE_TRAIL=0,','parameter bit USE_TRAIL=1,')
s=s.replace('parameter int CHOICE_CAPACITY=8,\n  parameter bit FIRST_ONLY=0','parameter int CHOICE_CAPACITY=8')
s=s.replace('input logic clk, rst_n,','input logic clk, rst_n,\n  input logic [63:0] initial_domains, adjacency,\n  input logic [7:0] active_vertices,\n  input logic first_only, trace_ready,\n  output wire [39:0] choice_debug,\n  output wire [19:0] trail_debug,')
s=s.replace('S_LOG_WRITE,S_APPLY,S_TCHECK,S_TWAIT,S_TCAPTURE,S_TAPPLY}', 'S_LOG_WRITE,S_APPLY,S_TCHECK,S_TWAIT,S_TCAPTURE,S_TAPPLY,S_INIT}')
s=s.replace('wire cp_wr_en = state_q==S_CP_WRITE || state_q==S_UPDATE_WRITE;', 'wire advance = !trace_valid || trace_ready;\n  wire cp_wr_en = advance && (state_q==S_CP_WRITE || state_q==S_UPDATE_WRITE);')
s=s.replace('.wr_en(USE_TRAIL && state_q==S_LOG_WRITE)', '.wr_en(advance && USE_TRAIL && state_q==S_LOG_WRITE)')
s=s.replace('assign propagated_o=propagated_q;', 'assign choice_debug=cp_q[39:0];\n  assign trail_debug=trace_kind==queens_types_pkg::E_RESTORE ? trail_entry_q : trail_wr_data;\n  assign propagated_o=propagated_q;')
s=s.replace('state_q<=S_SCAN; write_resume_q<=S_SCAN;', 'state_q<=S_INIT; write_resume_q<=S_SCAN;')
s=s.replace('end else begin\n      trace_valid<=0;', 'end else if(advance) begin\n      trace_valid<=0;')
s=s.replace('case(state_q)\n        S_SCAN:', 'case(state_q)\n        S_INIT: begin\n          for(c=0;c<8;c=c+1) domains_q[c]<=initial_domains[8*c+:8];\n          propagated_q<=~active_vertices;state_q<=S_SCAN;\n        end\n        S_SCAN:')
s=s.replace('domains_q[target_q] & ~queens_types_pkg::attack(source_q,row_q,target_q)', 'domains_q[target_q] & ~(adjacency[8*int\'(source_q)+int\'(target_q)] ? domains_q[source_q] : 8\'d0)')
s=s.replace('trail_entry_q[8:4]==0 || trail_entry_q[8:4]>choice_top_q ||','trail_entry_q[8:4]>choice_top_q ||')
s=s.replace('if(FIRST_ONLY)', 'if(first_only)')
s=s.replace('// One deterministic search controller with snapshot or trail recovery storage.','// Graph coloring with runtime adjacency and lossless semantic event backpressure.\n// Seeded from the queens laboratory; initial singleton propagation permits level zero.')
target.parent.mkdir(parents=True,exist_ok=True);target.write_text(s)
