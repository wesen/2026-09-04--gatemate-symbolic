`default_nettype none
// Stop-and-wait UART control. Compute is enabled only by a bounded T command;
// all other requests inspect or modify a stopped engine. No automatic retries.
module dataflow_link #(parameter integer CLK_HZ=10_000_000,BAUD=115200,COMMAND_TIMEOUT=CLK_HZ/5)(
 input wire clk,rst_n,rx,output wire tx,output wire idle
);
 wire rx_valid,rx_error;wire [7:0] rx_data;
 uart_rx #(.CLK_HZ(CLK_HZ),.BAUD(BAUD)) receiver(.clk(clk),.rst_n(rst_n),.rx(rx),.valid(rx_valid),.framing_error(rx_error),.data(rx_data));
 reg reset_hold,in_valid,out_ready,cancel_valid;
 reg graph_write,graph_commit;
 reg [2:0] graph_index;
 reg [23:0] graph_descriptor;
 reg [3:0] graph_size;
 wire graph_writable,graph_acceptable;
 reg [79:0] in_token;
 reg [1:0] cancel_context;
 reg [7:0] debug_addr;
 wire in_ready,out_valid,cancel_ready;wire [79:0] out_token,debug_data;
 reg [31:0] ticks_left;
 wire engine_enable=ticks_left!=0;
 dataflow_core core(.clk(clk),.rst_n(rst_n&&!reset_hold),.enable(engine_enable),.in_valid(in_valid),.in_ready(in_ready),.in_token(in_token),
   .out_valid(out_valid),.out_ready(out_ready),.out_token(out_token),.cancel_valid(cancel_valid),.cancel_context(cancel_context),.cancel_ready(cancel_ready),
   .graph_write(graph_write),.graph_commit(graph_commit),.graph_index(graph_index),.graph_descriptor(graph_descriptor),.graph_size(graph_size),.graph_writable(graph_writable),.graph_acceptable(graph_acceptable),.debug_addr(debug_addr),.debug_data(debug_data),.trace_valid(),.trace_token(),.quiescent(idle));
 reg [7:0] command;reg [5:0] digits;reg bad;
 reg [87:0] request;
 reg [31:0] idle_count;
 reg [2:0] query_wait;
 wire hex_valid=(rx_data>="0"&&rx_data<="9")||(rx_data>="a"&&rx_data<="f")||(rx_data>="A"&&rx_data<="F");
 wire [3:0] nibble=rx_data<="9"?4'(rx_data-"0"):rx_data<="F"?4'(rx_data-"A"+10):4'(rx_data-"a"+10);
 reg [7:0] request_xor;
 always @*begin request_xor=0;for(integer n=0;n<11;n=n+1)request_xor=request_xor^request[n*8+:8];end
 function automatic [7:0] checksum(input [79:0] data);
   begin checksum=0;for(integer n=0;n<10;n=n+1)checksum=checksum^data[n*8+:8];end
 endfunction
 function automatic [7:0] hex_ascii(input [3:0] v);
   hex_ascii=v<10?8'h30+{4'b0,v}:8'h41+{4'b0,v}-8'd10;
 endfunction
 reg tx_active,tx_long,tx_last;
 reg [4:0] tx_index,tx_length;
 reg [31:0] short_data;
 reg [87:0] response;
 reg [7:0] response_kind;
 wire uart_ready;
 wire uart_start=tx_active&&!tx_last&&uart_ready;
 reg [7:0] uart_data;
 always @*begin
   if(!tx_long)uart_data=8'(short_data>>(8*(3-tx_index)));
   else if(tx_index==0)uart_data=response_kind;
   else if(tx_index==23)uart_data=8'h0a;
   else uart_data=hex_ascii(4'(response>>(4*(22-tx_index))));
 end
 uart_tx #(.CLK_HZ(CLK_HZ),.BAUD(BAUD)) transmitter(.clk(clk),.rst_n(rst_n),.start(uart_start),.data(uart_data),.ready(uart_ready),.tx(tx));
 task automatic short_reply(input [31:0] data,input [4:0] length);
   begin short_data<=data;tx_length<=length;tx_long<=0;tx_active<=1;tx_last<=0;tx_index<=0;end
 endtask
 task automatic record_reply(input [7:0] kind,input [79:0] data);
   begin response_kind<=kind;response<={data,checksum(data)};tx_length<=24;tx_long<=1;tx_active<=1;tx_last<=0;tx_index<=0;end
 endtask
 always @(posedge clk or negedge rst_n)begin
   if(!rst_n)begin
     reset_hold<=0;in_valid<=0;out_ready<=0;cancel_valid<=0;in_token<=0;cancel_context<=0;debug_addr<=0;ticks_left<=0;
     graph_write<=0;graph_commit<=0;graph_index<=0;graph_descriptor<=0;graph_size<=0;
     command<=0;digits<=0;bad<=0;request<=0;idle_count<=0;query_wait<=0;
     tx_active<=0;tx_long<=0;tx_last<=0;tx_index<=0;tx_length<=0;short_data<=0;response<=0;response_kind<=0;
   end else begin
     reset_hold<=0;in_valid<=0;out_ready<=0;cancel_valid<=0;graph_write<=0;graph_commit<=0;
     if(uart_start)begin if(tx_index==tx_length-1'b1)tx_last<=1;else tx_index<=tx_index+1'b1;end
     if(tx_active&&tx_last&&uart_ready)begin tx_active<=0;tx_last<=0;end
     if(ticks_left!=0)begin ticks_left<=ticks_left-1'b1;if(ticks_left==1)short_reply({"A",8'h0a,16'b0},2);end
     if(query_wait!=0)begin query_wait<=query_wait-1'b1;if(query_wait==1)record_reply("S",debug_data);end
     if(!tx_active&&ticks_left==0&&query_wait==0&&!in_valid&&!out_ready&&!cancel_valid&&!reset_hold&&!graph_write&&!graph_commit)begin
       if(rx_error)begin command<=0;digits<=0;bad<=0;request<=0;idle_count<=0;short_reply({"!01",8'h0a},4);end
       else if(rx_valid)begin
         idle_count<=0;
         if(rx_data==8'h0a)begin
           if(bad)short_reply({"!01",8'h0a},4);
           else if(command=="R"&&digits==0)begin reset_hold<=1;short_reply({"A",8'h0a,16'b0},2);end
           else if(command=="P"&&digits==0)begin
             if(out_valid)begin record_reply("O",out_token);out_ready<=1;end
             else short_reply({"N",8'h0a,16'b0},2);
           end else if((command=="I"&&digits==22)||((command=="T"||command=="W")&&digits==10)||((command=="Q"||command=="C"||command=="G")&&digits==4))begin
             if(request_xor!=0)short_reply({"!03",8'h0a},4);
             else case(command)
               "I":if(in_ready)begin in_token<=request[87:8];in_valid<=1;short_reply({"A",8'h0a,16'b0},2);end else short_reply({"!04",8'h0a},4);
               "T":if(request[39:8]<=1_000_000)begin
                 ticks_left<=request[39:8];if(request[39:8]==0)short_reply({"A",8'h0a,16'b0},2);
               end else short_reply({"!02",8'h0a},4);
               "W":if(request[39:32]<7&&graph_writable)begin graph_index<=request[34:32];graph_descriptor<=request[31:8];graph_write<=1;short_reply({"A",8'h0a,16'b0},2);end else short_reply({"!02",8'h0a},4);
               "G":if(request[15:8]>=1&&request[15:8]<=7)begin graph_size<=request[11:8];command<=8'hfe;end else short_reply({"!02",8'h0a},4);
               "Q":begin debug_addr<=request[15:8];query_wait<=3;end
               // Cancellation readiness depends on the selected context, so
               // select it first and decide in a dedicated following cycle.
               "C":if(request[15:8]<4)begin cancel_context<=request[9:8];query_wait<=0;command<=8'hff;end else short_reply({"!02",8'h0a},4);
             endcase
           end else short_reply({"!01",8'h0a},4);
           if(!((command=="C"&&digits==4&&!bad&&request_xor==0&&request[15:8]<4)||(command=="G"&&digits==4&&!bad&&request_xor==0&&request[15:8]>=1&&request[15:8]<=7)))command<=0;
           digits<=0;bad<=0;request<=0;
         end else if(command==0)begin command<=rx_data;digits<=0;bad<=0;request<=0;end
         else if(digits<22&&hex_valid)begin request<={request[83:0],nibble};digits<=digits+1'b1;end
         else bad<=1;
       end else if(command==8'hfe)begin
         command<=0;
         if(graph_acceptable)begin graph_commit<=1;short_reply({"A",8'h0a,16'b0},2);end
         else short_reply({"!02",8'h0a},4);
       end else if(command==8'hff)begin
         command<=0;
         if(cancel_ready)begin cancel_valid<=1;short_reply({"A",8'h0a,16'b0},2);end
         else short_reply({"!05",8'h0a},4);
       end else if(command!=0)begin
         if(idle_count>=COMMAND_TIMEOUT)begin command<=0;digits<=0;bad<=0;request<=0;idle_count<=0;short_reply({"!01",8'h0a},4);end
         else idle_count<=idle_count+1'b1;
       end
     end
   end
 end
endmodule
`default_nettype wire
