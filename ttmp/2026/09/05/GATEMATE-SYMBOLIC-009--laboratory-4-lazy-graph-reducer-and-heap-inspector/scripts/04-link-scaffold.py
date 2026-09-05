#!/usr/bin/env python3
"""Reuse the proven UART response serializer, with an independent lazy protocol."""
from pathlib import Path
repo=Path(__file__).resolve().parents[5]
# ticket/scripts -> ticket -> day -> month -> year -> ttmp -> repository
repo=Path(__file__).resolve().parents[6]
old=(repo/'elastic_dataflow/rtl/dataflow_link.sv').read_text()
tx=old[old.index(' function automatic [7:0] checksum'):old.index(' always @(posedge clk or negedge rst_n)begin')]
head='''`default_nettype none
module lazy_link #(parameter integer CLK_HZ=10_000_000,BAUD=115200,COMMAND_TIMEOUT=CLK_HZ/5)(input wire clk,rst_n,rx,output wire tx,output wire idle);
 wire rx_valid,rx_error;wire [7:0] rx_data;
 uart_rx #(.CLK_HZ(CLK_HZ),.BAUD(BAUD)) receiver(.clk(clk),.rst_n(rst_n),.rx(rx),.valid(rx_valid),.framing_error(rx_error),.data(rx_data));
 reg reset_hold,load_valid,force_valid,result_ready;
 reg [15:0] load_address,force_root,debug_address;
 reg [39:0] load_word;
 reg [31:0] ticks_left;
 reg [2:0] query_wait;
 wire [39:0] result;wire result_valid,load_ready,force_ready;wire [79:0] debug_data;
 lazy_core core(.clk(clk),.rst_n(rst_n&&!reset_hold),.enable(ticks_left!=0),.load_valid(load_valid),.load_address(load_address),.load_word(load_word),.load_ready(load_ready),.force_valid(force_valid),.force_root(force_root),.force_ready(force_ready),.result_valid(result_valid),.result(result),.result_ready(result_ready),.debug_select(query_wait!=0),.debug_address(debug_address),.debug_data(debug_data),.idle(idle));
 reg [7:0] command;reg [5:0] digits;reg bad;
 reg [63:0] request;
 reg [31:0] idle_count;
 wire hex_valid=(rx_data>="0"&&rx_data<="9")||(rx_data>="a"&&rx_data<="f")||(rx_data>="A"&&rx_data<="F");
 wire [3:0] nibble=rx_data<="9"?4'(rx_data-"0"):rx_data<="F"?4'(rx_data-"A"+10):4'(rx_data-"a"+10);
 reg [7:0] request_xor;
 always @*begin request_xor=0;for(integer n=0;n<8;n=n+1)request_xor=request_xor^request[n*8+:8];end
 wire [3:0] load_tag=load_word[39:36];
 wire load_canonical=load_word[35:32]==0&&((load_tag<=2||load_tag==13)||((load_tag==3||load_tag==4)&&load_word[15:0]==0)||(load_tag==15&&load_word[31:0]==0));
'''
body=''' always @(posedge clk or negedge rst_n)begin
  if(!rst_n)begin
   reset_hold<=0;load_valid<=0;force_valid<=0;result_ready<=0;load_address<=0;load_word<=0;force_root<=0;debug_address<=0;ticks_left<=0;query_wait<=0;
   command<=0;digits<=0;bad<=0;request<=0;idle_count<=0;
   tx_active<=0;tx_long<=0;tx_last<=0;tx_index<=0;tx_length<=0;short_data<=0;response<=0;response_kind<=0;
  end else begin
   reset_hold<=0;load_valid<=0;force_valid<=0;result_ready<=0;
   if(uart_start)begin if(tx_index==tx_length-1'b1)tx_last<=1;else tx_index<=tx_index+1'b1;end
   if(tx_active&&tx_last&&uart_ready)begin tx_active<=0;tx_last<=0;end
   if(ticks_left!=0)begin ticks_left<=ticks_left-1'b1;if(ticks_left==1)short_reply({"A",8'h0a,16'b0},2);end
   if(query_wait!=0)begin query_wait<=query_wait-1'b1;if(query_wait==1)record_reply("S",debug_data);end
   if(!tx_active&&ticks_left==0&&query_wait==0&&!load_valid&&!force_valid&&!result_ready&&!reset_hold)begin
    if(command==8'hfe)begin command<=0;if(load_ready&&load_canonical)begin load_valid<=1;short_reply({"A",8'h0a,16'b0},2);end else short_reply({"!02",8'h0a},4);end
    else if(rx_error)begin command<=0;digits<=0;bad<=0;request<=0;idle_count<=0;short_reply({"!01",8'h0a},4);end
    else if(rx_valid)begin
     idle_count<=0;
     if(rx_data==8'h0a)begin
      if(bad)short_reply({"!01",8'h0a},4);
      else if(command=="R"&&digits==0)begin reset_hold<=1;short_reply({"A",8'h0a,16'b0},2);end
      else if(command=="P"&&digits==0)begin if(result_valid)begin record_reply("O",{40'b0,result});result_ready<=1;end else short_reply({"N",8'h0a,16'b0},2);end
      else if((command=="W"&&digits==16)||(command=="T"&&digits==10)||((command=="Q"||command=="F")&&digits==6))begin
       if(request_xor!=0)short_reply({"!03",8'h0a},4);
       else case(command)
        "W":begin load_address<=request[63:48];load_word<=request[47:8];command<=8'hfe;end
        "F":if(force_ready)begin force_root<=request[23:8];force_valid<=1;short_reply({"A",8'h0a,16'b0},2);end else short_reply({"!02",8'h0a},4);
        "T":if(request[39:8]<=1_000_000)begin ticks_left<=request[39:8];if(request[39:8]==0)short_reply({"A",8'h0a,16'b0},2);end else short_reply({"!02",8'h0a},4);
        "Q":begin debug_address<=request[23:8];query_wait<=3;end
       endcase
      end else short_reply({"!01",8'h0a},4);
      if(!(command=="W"&&digits==16&&!bad&&request_xor==0))command<=0;
      digits<=0;bad<=0;request<=0;
     end else if(command==0)begin command<=rx_data;digits<=0;bad<=0;request<=0;end
     else if(digits<16&&hex_valid)begin request<={request[59:0],nibble};digits<=digits+1'b1;end
     else bad<=1;
    end else if(command!=0)begin if(idle_count>=COMMAND_TIMEOUT)begin command<=0;digits<=0;bad<=0;request<=0;idle_count<=0;short_reply({"!01",8'h0a},4);end else idle_count<=idle_count+1'b1;end
   end
  end
 end
endmodule
`default_nettype wire
'''
(repo/'lazy_reducer/rtl/lazy_link.sv').write_text(head+tx+body)
top=(repo/'elastic_dataflow/rtl/dataflow_top.sv').read_text().replace('dataflow_top','lazy_top').replace('dataflow_link','lazy_link')
(repo/'lazy_reducer/rtl/lazy_top.sv').write_text(top)
