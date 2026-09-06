`default_nettype none
module lfl_link #(parameter integer CLK_HZ=10_000_000,BAUD=115200,COMMAND_TIMEOUT=CLK_HZ/5)(input wire clk,rst_n,rx,output wire tx,output wire idle);
 wire rx_valid,rx_error;wire [7:0] rx_data;
 uart_rx #(.CLK_HZ(CLK_HZ),.BAUD(BAUD)) receiver(.clk(clk),.rst_n(rst_n),.rx(rx),.valid(rx_valid),.framing_error(rx_error),.data(rx_data));
 reg reset_hold,load_valid,force_valid,result_ready;
 reg [2:0]load_kind;
 reg [15:0]load_address,force_ref,debug_address;
 reg [127:0]load_data;
 reg [31:0]ticks_left;
 reg [2:0]query_wait;
 wire [15:0]result_ref;wire result_valid,load_ready,force_ready;wire [127:0]debug_data;
 lfl_core core(.clk(clk),.rst_n(rst_n&&!reset_hold),.enable(ticks_left!=0),.load_valid(load_valid),.load_kind(load_kind),.load_address(load_address),.load_data(load_data),.load_ready(load_ready),.force_valid(force_valid),.force_ref(force_ref),.force_ready(force_ready),.result_valid(result_valid),.result_ref(result_ref),.result_ready(result_ready),.debug_select(query_wait!=0),.debug_address(debug_address),.debug_data(debug_data),.idle(idle));
 reg [7:0]command;reg [5:0]digits;reg bad;
 reg [151:0]request;
 reg [31:0]idle_count;
 reg loading;
 reg [15:0]code_count,heap_count,code_next,heap_next,prov_next;
 wire hex_valid=(rx_data>="0"&&rx_data<="9")||(rx_data>="a"&&rx_data<="f")||(rx_data>="A"&&rx_data<="F");
 wire [3:0]nibble=rx_data<="9"?4'(rx_data-"0"):rx_data<="F"?4'(rx_data-"A"+10):4'(rx_data-"a"+10);
 reg [7:0]request_xor;
 always @*begin request_xor=0;for(integer n=0;n<19;n=n+1)request_xor=request_xor^request[n*8+:8];end
 function automatic [7:0]checksum(input[127:0]data);begin checksum=0;for(integer n=0;n<16;n=n+1)checksum=checksum^data[n*8+:8];end endfunction
 function automatic [7:0]hex_ascii(input[3:0]v);hex_ascii=v<10?8'h30+{4'b0,v}:8'h41+{4'b0,v}-8'd10;endfunction
 reg tx_active,tx_long,tx_last;
 reg [5:0]tx_index,tx_length;
 reg [31:0]short_data;
 reg [135:0]response;
 reg [7:0]response_kind;
 wire uart_ready;
 wire uart_start=tx_active&&!tx_last&&uart_ready;
 reg [7:0]uart_data;
 always @*begin
  if(!tx_long)uart_data=8'(short_data>>(8*(3-tx_index)));
  else if(tx_index==0)uart_data=response_kind;
  else if(tx_index==35)uart_data=8'h0a;
  else uart_data=hex_ascii(4'(response>>(4*(34-tx_index))));
 end
 uart_tx #(.CLK_HZ(CLK_HZ),.BAUD(BAUD)) transmitter(.clk(clk),.rst_n(rst_n),.start(uart_start),.data(uart_data),.ready(uart_ready),.tx(tx));
 task automatic short_reply(input[31:0]data,input[5:0]length);begin short_data<=data;tx_length<=length;tx_long<=0;tx_active<=1;tx_last<=0;tx_index<=0;end endtask
 task automatic ack;begin short_reply({"A",8'h0a,16'b0},2);end endtask
 task automatic reject;begin short_reply({"!02",8'h0a},4);end endtask
 task automatic record_reply(input[7:0]kind,input[127:0]data);begin response_kind<=kind;response<={data,checksum(data)};tx_length<=36;tx_long<=1;tx_active<=1;tx_last<=0;tx_index<=0;end endtask
 always @(posedge clk or negedge rst_n)begin
  if(!rst_n)begin
   reset_hold<=0;load_valid<=0;load_kind<=0;load_address<=0;load_data<=0;force_valid<=0;force_ref<=0;result_ready<=0;debug_address<=0;ticks_left<=0;query_wait<=0;
   command<=0;digits<=0;bad<=0;request<=0;idle_count<=0;loading<=0;code_count<=0;heap_count<=0;code_next<=0;heap_next<=0;prov_next<=0;
   tx_active<=0;tx_long<=0;tx_last<=0;tx_index<=0;tx_length<=0;short_data<=0;response<=0;response_kind<=0;
  end else begin
   reset_hold<=0;load_valid<=0;force_valid<=0;result_ready<=0;
   if(uart_start)begin if(tx_index==tx_length-1'b1)tx_last<=1;else tx_index<=tx_index+1'b1;end
   if(tx_active&&tx_last&&uart_ready)begin tx_active<=0;tx_last<=0;end
   if(ticks_left!=0)begin ticks_left<=ticks_left-1'b1;if(ticks_left==1)ack();end
   if(query_wait!=0)begin query_wait<=query_wait-1'b1;if(query_wait==1)record_reply("S",debug_data);end
   if(!tx_active&&ticks_left==0&&query_wait==0&&!load_valid&&!force_valid&&!result_ready&&!reset_hold)begin
    if(rx_error)begin command<=0;digits<=0;bad<=0;request<=0;idle_count<=0;short_reply({"!01",8'h0a},4);end
    else if(rx_valid)begin
     idle_count<=0;
     if(rx_data==8'h0a)begin
      if(bad)short_reply({"!01",8'h0a},4);
      else if(command=="R"&&digits==0)begin reset_hold<=1;loading<=0;code_next<=0;heap_next<=0;prov_next<=0;ack();end
      else if(command=="P"&&digits==0)begin if(result_valid)begin record_reply("O",{112'b0,result_ref});result_ready<=1;end else short_reply({"N",8'h0a,16'b0},2);end
      else if(command=="K"&&digits==0)begin
       if(loading&&load_ready&&code_next==code_count&&heap_next==heap_count&&prov_next==heap_count)begin load_kind<=4;load_valid<=1;loading<=0;ack();end else reject();
      end
      else if((command=="B"&&digits==18)||(command=="C"&&digits==38)||(command=="H"&&digits==26)||((command=="V"||command=="T")&&digits==10)||((command=="Q"||command=="F")&&digits==6))begin
       if(request_xor!=0)short_reply({"!03",8'h0a},4);
       else case(command)
        "B":begin
         if(load_ready&&request[71:56]>0&&request[71:56]<=2048&&request[55:40]>=13&&request[55:40]<=2048&&request[39:24]<request[55:40]&&request[23:8]>=13&&request[23:8]<=request[55:40])begin
          load_kind<=0;load_data<={64'b0,request[71:8]};load_valid<=1;loading<=1;code_count<=request[71:56];heap_count<=request[55:40];code_next<=0;heap_next<=0;prov_next<=0;ack();
         end else reject();
        end
        "C":if(loading&&load_ready&&request[151:136]==code_next&&code_next<code_count&&request[135:128]<=9&&request[127:120]==0&&request[23:8]==0)begin load_kind<=1;load_address<=code_next;load_data<=request[135:8];load_valid<=1;code_next<=code_next+1'b1;ack();end else reject();
        "H":if(loading&&load_ready&&request[103:88]==heap_next&&heap_next<heap_count&&request[79:72]==0&&((request[87:80]<=2)||(request[87:80]==5)||(request[87:80]==8)||(request[87:80]==13)))begin load_kind<=2;load_address<=heap_next;load_data<={48'b0,request[87:8]};load_valid<=1;heap_next<=heap_next+1'b1;ack();end else reject();
        "V":if(loading&&load_ready&&request[39:24]==prov_next&&prov_next<heap_count)begin load_kind<=3;load_address<=prov_next;load_data<={112'b0,request[23:8]};load_valid<=1;prov_next<=prov_next+1'b1;ack();end else reject();
        "F":if(force_ready)begin force_ref<=request[23:8];force_valid<=1;ack();end else reject();
        "T":if(request[39:8]<=1_000_000)begin ticks_left<=request[39:8];if(request[39:8]==0)ack();end else reject();
        "Q":begin debug_address<=request[23:8];query_wait<=3;end
       endcase
      end else short_reply({"!01",8'h0a},4);
      command<=0;digits<=0;bad<=0;request<=0;
     end else if(command==0)begin command<=rx_data;digits<=0;bad<=0;request<=0;end
     else if(digits<38&&hex_valid)begin request<={request[147:0],nibble};digits<=digits+1'b1;end
     else bad<=1;
    end else if(command!=0)begin if(idle_count>=COMMAND_TIMEOUT)begin command<=0;digits<=0;bad<=0;request<=0;idle_count<=0;short_reply({"!01",8'h0a},4);end else idle_count<=idle_count+1'b1;end
   end
  end
 end
endmodule
`default_nettype wire
