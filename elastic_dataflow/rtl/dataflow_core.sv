`default_nettype none
module dataflow_core #(
 parameter integer INPUT_DEPTH=8, COMPLETION_DEPTH=8, OUTPUT_DEPTH=8,
 parameter integer MUL_LATENCY=4, EPOCH_BITS=8
)(
 input wire clk,rst_n,enable,
 input wire in_valid,output wire in_ready,input wire [79:0] in_token,
 output wire out_valid,input wire out_ready,output wire [79:0] out_token,
 input wire cancel_valid,input wire [1:0] cancel_context,output wire cancel_ready,
 input wire [7:0] debug_addr,output reg [79:0] debug_data,
 output reg trace_valid,output reg [79:0] trace_token,output wire quiescent
);
 reg [7:0] epoch[0:3];
 reg [3:0] closed,error_pending;
 reg [79:0] errors[0:3];
 reg [27:0] valid_a,valid_b,issued,pending;
 reg [31:0] metrics[0:15];
 reg [4:0] next_slot;
 reg prefer_router,prefer_mul;
 wire cancel_fire=cancel_valid&&cancel_ready;
 wire ce=enable&&!cancel_fire;
 wire [79:0] input_head,completion_head,output_head;
 wire input_valid,completion_valid,output_valid;
 wire completion_ready,output_ready;
 reg input_pop,completion_pop,output_push,completion_push;
 reg [79:0] output_input,completion_input;
 wire [7:0] input_count,completion_count,output_count;
 wire [79:0] input_debug,completion_debug,output_debug;
 wire output_stale=output_valid&&(output_head[79:72]>=4 || output_head[71:64]!=epoch[output_head[73:72]]);
 assign out_valid=output_valid&&!output_stale;
 assign out_token=output_head;
 df_fifo #(.DEPTH(INPUT_DEPTH)) inputs(.clk(clk),.rst_n(rst_n),.push(in_valid),.din(in_token),.ready(in_ready),
   .pop(input_pop),.valid(input_valid),.dout(input_head),.count(input_count),.debug_offset(debug_addr[2:0]),.debug_data(input_debug));
 df_fifo #(.DEPTH(COMPLETION_DEPTH)) completions(.clk(clk),.rst_n(rst_n),.push(completion_push),.din(completion_input),.ready(completion_ready),
   .pop(completion_pop),.valid(completion_valid),.dout(completion_head),.count(completion_count),.debug_offset(debug_addr[2:0]),.debug_data(completion_debug));
 df_fifo #(.DEPTH(OUTPUT_DEPTH)) outputs(.clk(clk),.rst_n(rst_n),.push(output_push),.din(output_input),.ready(output_ready),
   .pop(output_stale||(out_valid&&out_ready)),.valid(output_valid),.dout(output_head),.count(output_count),.debug_offset(debug_addr[2:0]),.debug_data(output_debug));

 reg [1:0] issue_phase;
 reg [4:0] issue_slot;
 reg [79:0] issue_token;
 reg [39:0] issue_a,issue_b;
 wire [39:0] ram_a,ram_b;
 reg ram_write;
 reg [4:0] write_slot;
 reg write_port;
 reg [39:0] write_value;
 // Paused debug reads share the synchronous operand ports. An issue that has
 // captured its operands owns registers; earlier issue phases repeat the read.
 wire debug_operand=!enable&&debug_addr>=32&&debug_addr<60;
 wire [4:0] read_slot=debug_operand?(debug_addr-32):issue_slot;
 sync_sdp_ram #(.DEPTH(32),.WIDTH(40)) operands_a(.clk(clk),.wr_en(ram_write&&!write_port),.wr_addr(write_slot),.wr_data(write_value),.rd_addr(read_slot),.rd_data(ram_a));
 sync_sdp_ram #(.DEPTH(32),.WIDTH(40)) operands_b(.clk(clk),.wr_en(ram_write&&write_port),.wr_addr(write_slot),.wr_data(write_value),.rd_addr(read_slot),.rd_data(ram_b));
 // The read-valid guard also handles resuming after debug selected another slot.
 reg [4:0] previous_read_slot;
 always @(posedge clk)previous_read_slot<=read_slot;
 wire issue_fresh=issue_token[71:64]==epoch[issue_token[73:72]]&&!closed[issue_token[73:72]];
 wire issue_mul=dataflow_pkg::opcode(issue_token[63:58])==dataflow_pkg::MUL;
 wire mul_ready,alu_ready,mul_valid,alu_valid;
 reg mul_take,alu_take;
 wire mul_admit=ce&&issue_phase==3&&issue_fresh&&issue_mul;
 wire alu_admit=ce&&issue_phase==3&&issue_fresh&&!issue_mul;
 wire [79:0] mul_token,alu_token,mul_debug,alu_debug;
 wire [7:0] mul_stages,alu_stages;
 df_unit #(.LATENCY(MUL_LATENCY)) multiplier(.clk(clk),.rst_n(rst_n),.ce(ce),.in_valid(mul_admit),.in_ready(mul_ready),
   .in_token(issue_token),.a(issue_a),.b(issue_b),.op(dataflow_pkg::opcode(issue_token[63:58])),.out_valid(mul_valid),.out_ready(mul_take),.out_token(mul_token),
   .debug_valid(mul_stages),.debug_stage(debug_addr[2:0]),.debug_token(mul_debug));
 df_unit #(.LATENCY(1)) alu(.clk(clk),.rst_n(rst_n),.ce(ce),.in_valid(alu_admit),.in_ready(alu_ready),
   .in_token(issue_token),.a(issue_a),.b(issue_b),.op(dataflow_pkg::opcode(issue_token[63:58])),.out_valid(alu_valid),.out_ready(alu_take),.out_token(alu_token),
   .debug_valid(alu_stages),.debug_stage(3'b0),.debug_token(alu_debug));
 wire mul_stale=mul_valid&&(mul_token[71:64]!=epoch[mul_token[73:72]]||closed[mul_token[73:72]]);
 wire alu_stale=alu_valid&&(alu_token[71:64]!=epoch[alu_token[73:72]]||closed[alu_token[73:72]]);
 wire dispatch=(mul_admit&&mul_ready)||(alu_admit&&alu_ready);
 reg router_valid;
 reg [79:0] router_token;
 reg [1:0] delivered;
 reg router_clear,router_mark;
 reg selected_router,action_valid,commit_operand;
 reg [79:0] action_token;
 reg fault_event,close_event,stale_action,invalid_action,duplicate_event;
 reg [1:0] fault_context,close_context;
 reg [79:0] fault_token;
 reg [6:0] dest;
 reg error_emit;
 reg [1:0] error_context;
 integer e,j,k;
 reg found_error;
 wire router_stale=router_valid&&(router_token[79:72]>=4 || router_token[71:64]!=epoch[router_token[73:72]]||closed[router_token[73:72]]);
 assign quiescent=!input_valid&&!completion_valid&&!output_valid&&mul_stages==0&&alu_stages==0&&issue_phase==0&&!router_valid&&pending==0&&error_pending==0;
 assign cancel_ready=!(out_valid&&out_token[79:72]=={6'b0,cancel_context})&&
   ((epoch[cancel_context] != ((1<<EPOCH_BITS)-1)) || quiescent);

 always @* begin
   completion_push=0;completion_input=0;mul_take=0;alu_take=0;
   if(ce)begin
     mul_take=mul_stale;alu_take=alu_stale;
     if(completion_ready)begin
       if(mul_valid&&!mul_stale&&(!alu_valid||alu_stale||prefer_mul))begin
         completion_push=1;completion_input=mul_token;mul_take=1;
       end else if(alu_valid&&!alu_stale)begin
         completion_push=1;completion_input=alu_token;alu_take=1;
       end
     end
   end
 end
 always @* begin
   input_pop=0;completion_pop=ce&&!router_valid&&completion_valid;
   output_push=0;output_input=0;error_emit=0;error_context=0;found_error=0;e=0;
   router_clear=0;router_mark=0;selected_router=0;action_valid=0;action_token=0;dest=0;
   commit_operand=0;fault_event=0;fault_context=0;fault_token=0;
   close_event=0;close_context=0;stale_action=0;invalid_action=0;duplicate_event=0;
   ram_write=0;write_slot=0;write_port=0;write_value=0;
   if(ce)begin
     for(e=0;e<4;e=e+1)if(error_pending[e]&&!found_error)begin
       found_error=1;
       if(output_ready)begin output_push=1;output_input=errors[e];error_emit=1;error_context=e;end
     end
     // One router action or external operand commit per enabled cycle.
     selected_router=router_valid&&(!input_valid||prefer_router);
     if(selected_router)begin
       if(router_stale)begin router_clear=1;stale_action=1;end
       else if(router_token[39:36]==4'hd)begin
         fault_event=1;fault_context=router_token[73:72];
         fault_token=dataflow_pkg::fault(router_token[79:72],router_token[71:64],router_token[63:58],router_token[7:0]);router_clear=1;
       end else if(router_token[56])begin
         if(output_ready&&!output_push)begin output_push=1;output_input=router_token;router_clear=1;close_event=1;close_context=router_token[73:72];end
       end else begin
         dest=dataflow_pkg::destination(router_token[63:58],delivered[0]);
         action_valid=1;action_token={router_token[79:64],dest,1'b0,router_token[55:0]};router_mark=1;
       end
     end else if(input_valid)begin action_valid=1;action_token=input_head;input_pop=1;end
     if(action_valid)begin
       if(action_token[79:72]>=4)invalid_action=1;
       else if(action_token[71:64]!=epoch[action_token[73:72]]||closed[action_token[73:72]])stale_action=1;
       else if(action_token[63:58]>=7||action_token[56]||action_token[47:40]!=0||
           (action_token[57]&&(action_token[63:58]==4||action_token[63:58]==6)))begin
         fault_event=1;fault_context=action_token[73:72];fault_token=dataflow_pkg::fault(action_token[79:72],action_token[71:64],action_token[63:58],1);
       end else begin
         write_slot=action_token[73:72]*7+action_token[63:58];write_port=action_token[57];write_value=action_token[39:0];
         if((write_port&&valid_b[write_slot])||(!write_port&&valid_a[write_slot]))begin
           fault_event=1;fault_context=action_token[73:72];fault_token=dataflow_pkg::fault(action_token[79:72],action_token[71:64],action_token[63:58],2);duplicate_event=1;
         end else begin ram_write=1;commit_operand=1;end
       end
     end
   end
 end

 reg select_valid;
 reg [4:0] selected_slot;
 integer scan;
 wire [1:0] ready_pairs[0:13];
 wire [2:0] ready_quads[0:6];
 wire [3:0] ready_octets[0:3];
 wire [4:0] ready_halves[0:1];
 for(genvar g=0;g<14;g=g+1)begin:count_pairs
   assign ready_pairs[g]={1'b0,pending[g*2]}+{1'b0,pending[g*2+1]};
 end
 for(genvar g=0;g<7;g=g+1)begin:count_quads
   assign ready_quads[g]={1'b0,ready_pairs[g*2]}+{1'b0,ready_pairs[g*2+1]};
 end
 for(genvar g=0;g<3;g=g+1)begin:count_octets
   assign ready_octets[g]={1'b0,ready_quads[g*2]}+{1'b0,ready_quads[g*2+1]};
 end
 assign ready_octets[3]={1'b0,ready_quads[6]};
 assign ready_halves[0]={1'b0,ready_octets[0]}+{1'b0,ready_octets[1]};
 assign ready_halves[1]={1'b0,ready_octets[2]}+{1'b0,ready_octets[3]};
 wire [7:0] ready_count={3'b0,ready_halves[0]}+{3'b0,ready_halves[1]};
 wire [1:0] selected_context=selected_slot>=21?2'd3:selected_slot>=14?2'd2:selected_slot>=7?2'd1:2'd0;
 wire [5:0] selected_node={1'b0,selected_slot}-{4'b0,selected_context}*6'd7;
 reg [27:0] eligible;
 always @* begin
   select_valid=0;selected_slot=0;eligible=0;
   for(scan=0;scan<28;scan=scan+1)begin
     eligible[scan]=pending[scan]&&!closed[scan/7]&&
       ((dataflow_pkg::opcode(scan%7)==dataflow_pkg::MUL&&mul_ready)||
        (dataflow_pkg::opcode(scan%7)!=dataflow_pkg::MUL&&alu_ready));
   end
   // Low indices win each descending pass. The second pass restricts the
   // choice to the unwrapped region, implementing a circular first-set scan.
   for(scan=27;scan>=0;scan=scan-1)if(eligible[scan])begin select_valid=1;selected_slot=scan;end
   for(scan=27;scan>=0;scan=scan-1)if(eligible[scan]&&scan>=next_slot)begin select_valid=1;selected_slot=scan;end
 end
 always @(posedge clk or negedge rst_n)begin
   if(!rst_n)begin
     for(j=0;j<4;j=j+1)begin epoch[j]<=0;errors[j]<=0;end
     for(j=0;j<16;j=j+1)metrics[j]<=0;
     closed<=0;error_pending<=0;valid_a<=0;valid_b<=0;issued<=0;pending<=0;
     issue_phase<=0;issue_slot<=0;issue_token<=0;issue_a<=0;issue_b<=0;
     router_valid<=0;router_token<=0;delivered<=0;next_slot<=0;prefer_router<=0;prefer_mul<=0;trace_valid<=0;trace_token<=0;
   end else begin
     trace_valid<=0;
     if(in_valid&&in_ready)metrics[1]<=metrics[1]+1'b1;
     if(out_valid&&out_ready)metrics[11]<=metrics[11]+1'b1;
     metrics[5]<=metrics[5]+(output_stale?32'd1:32'd0)+(ce&&mul_stale?32'd1:32'd0)+(ce&&alu_stale?32'd1:32'd0)+
       (ce&&stale_action?32'd1:32'd0)+(ce&&issue_phase!=0&&!issue_fresh?32'd1:32'd0);
     if(input_count>metrics[12])metrics[12]<=input_count;
     if(ready_count>metrics[13])metrics[13]<=ready_count;
     if(completion_count>metrics[14])metrics[14]<=completion_count;
     if(output_count>metrics[15])metrics[15]<=output_count;
     if(cancel_fire)begin
       epoch[cancel_context]<=(epoch[cancel_context]+1'b1)&((1<<EPOCH_BITS)-1);
       closed[cancel_context]<=0;error_pending[cancel_context]<=0;
       for(j=0;j<7;j=j+1)begin valid_a[cancel_context*7+j]<=0;valid_b[cancel_context*7+j]<=0;issued[cancel_context*7+j]<=0;pending[cancel_context*7+j]<=0;end
     end else if(ce)begin
       metrics[0]<=metrics[0]+1'b1;
       if(mul_stages!=0||alu_stages!=0)metrics[8]<=metrics[8]+1'b1;
       if((mul_valid&&!mul_take)||(alu_valid&&!alu_take))metrics[9]<=metrics[9]+1'b1;
       if(selected_router&&!router_clear&&!router_mark)metrics[10]<=metrics[10]+1'b1;
       if(duplicate_event)metrics[6]<=metrics[6]+1'b1;
       if(invalid_action)metrics[7]<=metrics[7]+1'b1;
       if(completion_push)prefer_mul<=!mul_take;
       if(action_valid||selected_router)prefer_router<=!selected_router;
       if(completion_pop)begin router_valid<=1;router_token<=completion_head;delivered<=0;end
       if(router_clear)router_valid<=0;
       if(router_mark)begin
         if(router_token[63:58]!=6||delivered[0])begin delivered<=2'b11;router_valid<=0;end
         else delivered<=2'b01;
       end
       if(error_emit)error_pending[error_context]<=0;
       if(commit_operand)begin
         if(write_port)valid_b[write_slot]<=1;else valid_a[write_slot]<=1;
         if(!issued[write_slot]&&((write_port&&valid_a[write_slot])||(!write_port&&(valid_b[write_slot]||dataflow_pkg::required_ports(action_token[63:58])==1))))begin
           issued[write_slot]<=1;pending[write_slot]<=1;
         end
       end
       if(issue_phase!=0&&!issue_fresh)issue_phase<=0;
       else case(issue_phase)
         0:if(select_valid)begin
           issue_slot<=selected_slot;issue_token<=dataflow_pkg::completion({6'b0,selected_context},epoch[selected_context],selected_node,40'b0);
           pending[selected_slot]<=0;next_slot<=selected_slot==27?0:selected_slot+1'b1;issue_phase<=1;
         end
         1:issue_phase<=2;
         2:if(previous_read_slot==issue_slot&&!debug_operand)begin issue_a<=ram_a;issue_b<=ram_b;issue_phase<=3;end
         3:if(dispatch)begin
           issue_phase<=0;trace_valid<=1;trace_token<={issue_token[79:40],dataflow_pkg::evaluate(dataflow_pkg::opcode(issue_token[63:58]),issue_a,issue_b)};
           metrics[2]<=metrics[2]+1'b1;
           if(issue_mul)metrics[3]<=metrics[3]+1'b1;else metrics[4]<=metrics[4]+1'b1;
         end
       endcase
       // First fault closes just this context; clearing pending wins over a
       // same-cycle ready reservation or selection for the failed context.
       if(fault_event)begin
         closed[fault_context]<=1;error_pending[fault_context]<=1;errors[fault_context]<=fault_token;
         for(j=0;j<7;j=j+1)pending[fault_context*7+j]<=0;
       end
       if(close_event)closed[close_context]<=1;
     end
   end
 end
 always @* begin
   debug_data=0;
   case(debug_addr)
     0:debug_data={8'd1,8'd4,8'd7,8'(MUL_LATENCY),8'(INPUT_DEPTH),8'(COMPLETION_DEPTH),8'(OUTPUT_DEPTH),8'(EPOCH_BITS),16'b0};
     1:debug_data={epoch[3],epoch[2],epoch[1],epoch[0],closed,error_pending,input_count,ready_count,completion_count,output_count,quiescent,7'b0};
     10:debug_data={issue_token[79:48],6'b0,issue_phase,39'b0,(issue_phase!=0)};
     11:debug_data={issue_a,issue_b};
     12:debug_data=router_token;
     13:debug_data={77'b0,delivered,router_valid};
     24:debug_data={71'b0,alu_stages[0],mul_stages};
     25:debug_data=alu_debug;
     default:begin
       if(debug_addr>=2&&debug_addr<=9)debug_data={metrics[(debug_addr-2)*2],metrics[(debug_addr-2)*2+1],16'b0};
       else if(debug_addr>=16&&debug_addr<24)debug_data=mul_debug;
       else if(debug_addr>=32&&debug_addr<60)debug_data={ram_a,ram_b};
       else if(debug_addr>=64&&debug_addr<92)debug_data={76'b0,pending[debug_addr-64],issued[debug_addr-64],valid_b[debug_addr-64],valid_a[debug_addr-64]};
       else if(debug_addr>=96&&debug_addr<104)debug_data=input_debug;
       else if(debug_addr>=112&&debug_addr<120)debug_data=completion_debug;
       else if(debug_addr>=128&&debug_addr<136)debug_data=output_debug;
       else if(debug_addr>=144&&debug_addr<148)debug_data=errors[debug_addr-144];
     end
   endcase
 end
endmodule
`default_nettype wire
