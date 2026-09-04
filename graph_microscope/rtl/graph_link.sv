`default_nettype none
module graph_link #(
 parameter integer CLK_HZ=10_000_000, BAUD=115200,
 parameter integer TRAIL_CAPACITY=64, CHOICE_CAPACITY=8,
 parameter integer COMMAND_TIMEOUT=CLK_HZ/5
)(input wire clk,rst_n,rx, output wire tx, output wire loaded,done,fault_valid);
 wire rx_valid,rx_error;
 wire [7:0] rx_data;
 uart_rx #(.CLK_HZ(CLK_HZ),.BAUD(BAUD)) u_rx(.clk(clk),.rst_n(rst_n),.rx(rx),.valid(rx_valid),.framing_error(rx_error),.data(rx_data));
 reg loaded_q,reset_hold;
 reg [3:0] vertices_q,colors_q;
 reg first_q;
 reg [63:0] adjacency_q;
 reg [63:0] initial_domains;
 wire [7:0] active_vertices=8'((9'd1<<vertices_q)-1'b1);
 assign loaded=loaded_q;
 always @* begin
  initial_domains=0;
  for(integer i=0;i<8;i=i+1) initial_domains[8*i+:8]=i<vertices_q ? 8'((9'd1<<colors_q)-1'b1) : 8'd1;
 end
 wire trace_valid,trace_ready;
 wire [3:0] trace_kind,fault_code,choice_top;
 wire [63:0] domains;
 wire [7:0] propagated;
 wire [6:0] trail_top,trail_base;
 wire [31:0] solution_count;
 wire [23:0] result_data;
 wire [39:0] choice_debug;
 wire [19:0] trail_debug;
 graph_core #(.TRAIL_CAPACITY(TRAIL_CAPACITY),.CHOICE_CAPACITY(CHOICE_CAPACITY)) u_core(
  .clk(clk),.rst_n(rst_n&&loaded_q&&!reset_hold),.initial_domains(initial_domains),.adjacency(adjacency_q),.active_vertices(active_vertices),
  .first_only(first_q),.trace_ready(trace_ready),.trace_valid(trace_valid),.trace_kind(trace_kind),.choice_debug(choice_debug),.trail_debug(trail_debug),
  .result_ready(1'b1),.result_data(result_data),.done(done),.fault_valid(fault_valid),.fault_code(fault_code),.solution_count(solution_count),
  .domains_o(domains),.propagated_o(propagated),.choice_top_o(choice_top),.trail_top_o(trail_top),.trail_base_o(trail_base));

 reg [7:0] command;
 reg [5:0] digits;
 reg bad;
 reg [95:0] load_shift;
 reg [31:0] idle_count,sequence_q;
 wire hex_valid=(rx_data>="0"&&rx_data<="9")||(rx_data>="A"&&rx_data<="F")||(rx_data>="a"&&rx_data<="f");
 wire [3:0] hex_value=rx_data<="9" ? 4'(rx_data-"0") : rx_data<="F" ? 4'(rx_data-"A"+10) : 4'(rx_data-"a"+10);
 reg graph_valid;
 reg [7:0] load_xor;
 always @* begin
  graph_valid=load_shift[95:88]>=1 && load_shift[95:88]<=8 && load_shift[87:80]>=1 && load_shift[87:80]<=8 && load_shift[79:72]<=1;
  load_xor=0;
  for(integer i=0;i<12;i=i+1) load_xor=load_xor^load_shift[8*i+:8];
  for(integer i=0;i<8;i=i+1) for(integer j=0;j<8;j=j+1) begin
   if(load_shift[64-8*i+j] && (i==j || i>=load_shift[95:88] || j>=load_shift[95:88])) graph_valid=0;
   if(load_shift[64-8*i+j]!=load_shift[64-8*j+i]) graph_valid=0;
  end
 end

 reg step_pending,tx_active,tx_event,tx_last;
 reg [6:0] tx_index,tx_length;
 reg [31:0] short_data;
 reg [271:0] event_frame;
 wire [263:0] event_payload={sequence_q+32'd1,4'd0,trace_kind,domains,propagated,4'd0,choice_top,1'b0,trail_top,1'b0,trail_base,solution_count,result_data,4'd0,fault_code,choice_debug,4'd0,trail_debug};
 function automatic [7:0] event_xor(input [263:0] payload);
  begin event_xor=0;for(integer i=0;i<33;i=i+1)event_xor=event_xor^payload[8*i+:8];end
 endfunction
 function automatic [7:0] ascii_hex(input [3:0] nibble);
  begin ascii_hex=nibble<10 ? 8'h30+{4'd0,nibble} : 8'h41+{4'd0,nibble}-8'd10;end
 endfunction
 wire uart_ready;
 wire uart_start=tx_active&&!tx_last&&uart_ready;
 reg [7:0] uart_data;
 always @* begin
  if(!tx_event)uart_data=8'(short_data >> (8*(3-tx_index)));
  else if(tx_index==0)uart_data="E";
  else if(tx_index==69)uart_data=8'h0a;
  else uart_data=ascii_hex(4'(event_frame >> (4*(68-tx_index))));
 end
 uart_tx #(.CLK_HZ(CLK_HZ),.BAUD(BAUD)) u_tx(.clk(clk),.rst_n(rst_n),.start(uart_start),.data(uart_data),.ready(uart_ready),.tx(tx));
 assign trace_ready=step_pending&&!tx_active&&trace_valid&&!reset_hold;
 task automatic respond(input [31:0] bytes,input [6:0] length);
  begin short_data<=bytes;tx_length<=length;tx_event<=0;tx_active<=1;tx_last<=0;tx_index<=0;end
 endtask
 always @(posedge clk or negedge rst_n) begin
  if(!rst_n)begin
   loaded_q<=0;reset_hold<=0;vertices_q<=1;colors_q<=1;first_q<=0;adjacency_q<=0;
   command<=0;digits<=0;bad<=0;load_shift<=0;idle_count<=0;sequence_q<=0;step_pending<=0;
   tx_active<=0;tx_event<=0;tx_last<=0;tx_index<=0;tx_length<=0;short_data<=0;event_frame<=0;
  end else begin
   reset_hold<=0;
   if(uart_start)begin if(tx_index==tx_length-1'b1)tx_last<=1;else tx_index<=tx_index+1'b1;end
   if(tx_active&&tx_last&&uart_ready)begin tx_active<=0;tx_last<=0;end
   if(step_pending&&!tx_active)begin
    if(trace_valid)begin event_frame<={event_payload,event_xor(event_payload)};sequence_q<=sequence_q+1'b1;step_pending<=0;tx_event<=1;tx_active<=1;tx_last<=0;tx_index<=0;tx_length<=70;end
    else if(done||fault_valid)begin step_pending<=0;respond({"Z",8'h0a,16'd0},2);end
   end
   if(!tx_active&&!step_pending)begin
    if(rx_error)begin command<=0;digits<=0;bad<=0;idle_count<=0;respond({"!01",8'h0a},4);end
    else if(rx_valid)begin
     idle_count<=0;
     if(rx_data==8'h0a)begin
      if(command=="L"&&digits==24&&!bad)begin
       if(load_xor!=0)respond({"!03",8'h0a},4);
       else if(!graph_valid)respond({"!02",8'h0a},4);
       else begin
        vertices_q<=load_shift[91:88];colors_q<=load_shift[83:80];first_q<=load_shift[72];
        for(integer i=0;i<8;i=i+1)adjacency_q[8*i+:8]<=load_shift[64-8*i+:8];
        loaded_q<=1;reset_hold<=1;sequence_q<=0;respond({"A",8'h0a,16'd0},2);
       end
      end else if(command=="S"&&digits==0&&!bad&&loaded_q)step_pending<=1;
      else if(command=="R"&&digits==0&&!bad&&loaded_q)begin reset_hold<=1;sequence_q<=0;respond({"A",8'h0a,16'd0},2);end
      else respond({"!01",8'h0a},4);
      command<=0;digits<=0;bad<=0;load_shift<=0;
     end else if(command==0)begin command<=rx_data;digits<=0;bad<=0;load_shift<=0;end
     else if(command=="L"&&digits<24&&hex_valid)begin load_shift<={load_shift[91:0],hex_value};digits<=digits+1'b1;end
     else bad<=1;
    end else if(command!=0)begin
     if(idle_count>=COMMAND_TIMEOUT)begin command<=0;digits<=0;bad<=0;idle_count<=0;respond({"!01",8'h0a},4);end
     else idle_count<=idle_count+1'b1;
    end
   end
  end
 end
endmodule
`default_nettype wire
