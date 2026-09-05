#!/usr/bin/env python3
"""One-time scheduler rewrite: constant indices, bounded context decoding."""
from pathlib import Path
p=Path('elastic_dataflow/rtl/dataflow_core.sv')
s=p.read_text(); a=s.index(' integer scan,idx;'); b=s.index(' always @(posedge clk or negedge rst_n)begin',a)
s=s[:a]+''' integer scan;
 reg [7:0] ready_count;
 wire [1:0] selected_context=selected_slot>=21?2'd3:selected_slot>=14?2'd2:selected_slot>=7?2'd1:2'd0;
 wire [5:0] selected_node={1'b0,selected_slot}-{4'b0,selected_context}*6'd7;
 reg [27:0] eligible;
 always @* begin
   select_valid=0;selected_slot=0;ready_count=0;eligible=0;
   for(scan=0;scan<28;scan=scan+1)begin
     if(pending[scan])ready_count=ready_count+1'b1;
     eligible[scan]=pending[scan]&&!closed[scan/7]&&
       ((dataflow_pkg::opcode(scan%7)==dataflow_pkg::MUL&&mul_ready)||
        (dataflow_pkg::opcode(scan%7)!=dataflow_pkg::MUL&&alu_ready));
   end
   // Low indices win each descending pass. The second pass restricts the
   // choice to the unwrapped region, implementing a circular first-set scan.
   for(scan=27;scan>=0;scan=scan-1)if(eligible[scan])begin select_valid=1;selected_slot=scan;end
   for(scan=27;scan>=0;scan=scan-1)if(eligible[scan]&&scan>=next_slot)begin select_valid=1;selected_slot=scan;end
 end
'''+s[b:]
s=s.replace('dataflow_pkg::completion(selected_slot/7,epoch[selected_slot/7],selected_slot%7,40\'b0)',"dataflow_pkg::completion({6'b0,selected_context},epoch[selected_context],selected_node,40'b0)")
s=s.replace('First dataflow_pkg::fault','First fault')
p.write_text(s)
