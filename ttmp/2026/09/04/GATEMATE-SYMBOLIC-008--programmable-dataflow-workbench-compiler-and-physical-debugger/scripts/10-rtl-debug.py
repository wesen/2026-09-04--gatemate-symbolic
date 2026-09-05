from pathlib import Path
p=Path('elastic_dataflow/rtl/dataflow_core.sv');s=p.read_text()
s=s.replace('input wire graph_write,graph_commit,','input wire debug_control_valid,input wire [7:0] debug_flags,debug_node,debug_context,output reg halted,\n input wire graph_write,graph_commit,')
s=s.replace('wire ce=enable&&!cancel_fire;', 'wire ce=enable&&!cancel_fire&&!halted;')
anchor=' always @* begin\n   debug_data=0;'
block=''' reg [2:0] break_mask,stop_reason;
 reg [7:0] break_node,break_context;
 reg [31:0] stop_cycle,trace_dropped;
 reg [5:0] trace_count;
 reg [6:0] event_bits;
 reg [79:0] event_token;
 reg [7:0] event_kind;
 reg [3:0] event_count,stale_count;
 wire [31:0] event_cycle=metrics[0]+{31'b0,ce};
 wire [79:0] trace_read_token;
 wire [39:0] trace_read_meta;
 wire trace_write=event_count!=0&&trace_count<32&&!debug_control_valid;
 wire [4:0] trace_read_index=debug_addr[5:1];
 sync_sdp_ram #(.DEPTH(32),.WIDTH(80)) debug_tokens(.clk(clk),.wr_en(trace_write),.wr_addr(trace_count[4:0]),.wr_data(event_token),.rd_addr(trace_read_index),.rd_data(trace_read_token));
 sync_sdp_ram #(.DEPTH(32),.WIDTH(40)) debug_metadata(.clk(clk),.wr_en(trace_write),.wr_addr(trace_count[4:0]),.wr_data({event_cycle,event_kind}),.rd_addr(trace_read_index),.rd_data(trace_read_meta));
 always @*begin
   stale_count={3'b0,output_stale}+{3'b0,(ce&&mul_stale)}+{3'b0,(ce&&alu_stale)}+
       {3'b0,(ce&&stale_action)}+{3'b0,(ce&&issue_phase!=0&&!issue_fresh)};
   event_bits={stale_count!=0,cancel_fire,(out_valid&&out_ready),router_mark,completion_push,dispatch,commit_operand};
   event_count={3'b0,event_bits[0]}+{3'b0,event_bits[1]}+{3'b0,event_bits[2]}+{3'b0,event_bits[3]}+{3'b0,event_bits[4]}+{3'b0,event_bits[5]}+stale_count;
   event_token=0;event_kind=0;
   // Later clauses have priority. Every unrecorded candidate is counted.
   if(event_bits[0])begin event_kind=1;event_token=action_token;end
   if(event_bits[3])begin event_kind=4;event_token=action_token;end
   if(event_bits[2])begin event_kind=3;event_token=completion_input;end
   if(event_bits[1])begin event_kind=2;event_token={issue_token[79:40],dataflow_pkg::evaluate(opcode(issue_token[63:58]),issue_a,issue_b)};end
   if(event_bits[6])begin
     event_kind=7;
     if(output_stale)event_token=output_head;
     else if(ce&&mul_stale)event_token=mul_token;
     else if(ce&&alu_stale)event_token=alu_token;
     else if(ce&&issue_phase!=0&&!issue_fresh)event_token=issue_token;
     else event_token=selected_router?router_token:action_token;
   end
   if(event_bits[4])begin event_kind=5;event_token=out_token;end
   if(event_bits[5])begin event_kind=6;event_token={{6'b0,cancel_context},(epoch[cancel_context]+8'd1),64'b0};end
 end
 wire issue_break=dispatch&&break_mask[0]&&(break_node==255||break_node=={2'b0,issue_token[63:58]})&&
   (break_context==255||break_context==issue_token[79:72]);
 wire full_break=ce&&break_mask[1]&&completion_count==COMPLETION_DEPTH;
 wire stale_break=break_mask[2]&&stale_count!=0;
 always @(posedge clk or negedge rst_n)begin
   if(!rst_n)begin halted<=0;break_mask<=0;break_node<=255;break_context<=255;stop_reason<=0;stop_cycle<=0;trace_count<=0;trace_dropped<=0;end
   else if(debug_control_valid)begin
     break_mask<=debug_flags[2:0];break_node<=debug_node;break_context<=debug_context;
     if(debug_flags[7])begin halted<=0;stop_reason<=0;end
     if(debug_flags[6])begin trace_count<=0;trace_dropped<=0;end
   end else begin
     if(event_count!=0)begin
       if(trace_count<32)begin trace_count<=trace_count+1'b1;trace_dropped<=trace_dropped+{28'b0,event_count}-1'b1;end
       else trace_dropped<=trace_dropped+{28'b0,event_count};
     end
     if(issue_break||full_break||stale_break)begin halted<=1;stop_reason<={stale_break,full_break,issue_break};stop_cycle<=event_cycle;end
   end
 end
'''
s=s.replace(anchor,block+anchor)
s=s.replace("155:debug_data=", "156:debug_data={2'b0,trace_count,trace_dropped,7'b0,halted,5'b0,stop_reason,5'b0,break_mask,break_node,break_context};\n     157:debug_data={stop_cycle,48'b0};\n     155:debug_data=")
s=s.replace('default:begin\n       if(debug_addr>=148', "default:begin\n       if(debug_addr>=160&&debug_addr<=223)debug_data=debug_addr[0]?{40'b0,trace_read_meta}:trace_read_token;\n       if(debug_addr>=148")
p.write_text(s)
# Connect explicit inactive controls in all direct core benches.
for p in Path('elastic_dataflow/sim').glob('*.sv'):
 s=p.read_text()
 if 'dataflow_core' in s:
  s=s.replace('dut(',"dut(.debug_control_valid(1'b0),.debug_flags(8'b0),.debug_node(8'hff),.debug_context(8'hff),.halted(),",1)
 p.write_text(s)
p=Path('elastic_dataflow/rtl/dataflow_link.sv');s=p.read_text()
s=s.replace('reg graph_write,graph_commit;','reg graph_write,graph_commit;\n reg debug_control_valid;reg [7:0] debug_flags,debug_node,debug_context;wire halted;')
s=s.replace('.graph_write(graph_write)', '.debug_control_valid(debug_control_valid),.debug_flags(debug_flags),.debug_node(debug_node),.debug_context(debug_context),.halted(halted),.graph_write(graph_write)')
s=s.replace('graph_write<=0;graph_commit<=0;graph_index<=0;', 'debug_control_valid<=0;debug_flags<=0;debug_node<=255;debug_context<=255;graph_write<=0;graph_commit<=0;graph_index<=0;')
s=s.replace('cancel_valid<=0;graph_write<=0;graph_commit<=0;\n','cancel_valid<=0;graph_write<=0;graph_commit<=0;debug_control_valid<=0;\n')
s=s.replace("if(ticks_left!=0)begin ticks_left<=ticks_left-1'b1;if(ticks_left==1)short_reply", "if(ticks_left!=0)begin ticks_left<=halted?0:ticks_left-1'b1;if(ticks_left==1||halted)short_reply")
s=s.replace('&&!graph_commit)', '&&!graph_commit&&!debug_control_valid)')
s=s.replace('else if((command=="I"&&digits==22)', 'else if((command=="B"&&digits==8)||(command=="I"&&digits==22)')
s=s.replace('"W":if(request', '''"B":if((request[31:24]&8'h38)==0&&(request[23:16]<7||request[23:16]==255)&&(request[15:8]<4||request[15:8]==255))begin
                 debug_flags<=request[31:24];debug_node<=request[23:16];debug_context<=request[15:8];debug_control_valid<=1;short_reply({"A",8'h0a,16'b0},2);
               end else short_reply({"!02",8'h0a},4);
               "W":if(request''')
p.write_text(s)
