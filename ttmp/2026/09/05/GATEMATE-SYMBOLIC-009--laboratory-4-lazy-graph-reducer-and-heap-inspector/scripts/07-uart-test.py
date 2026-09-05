#!/usr/bin/env python3
from pathlib import Path
repo=Path(__file__).resolve().parents[6]
s=(repo/'elastic_dataflow/sim/tb_dataflow_link.sv').read_text();s=s[:s.index(' task inject')].replace('tb_dataflow_link','lazy_link_tb').replace('dataflow_link #','lazy_link #')
s+=''' task load(input [15:0] a,input [39:0] w);begin request("W",7,{24'b0,a,w},0);ack();end endtask
 initial begin
  repeat(8)@(negedge clk);rst_n=1;repeat(8)@(negedge clk);
  fork send_byte("W");receive();join
  if(length!=4||response[2]!="1")$fatal(1,"partial timeout");
  request("Q",2,0,1);if(length!=4||response[2]!="3")$fatal(1,"checksum");
  request("R",0,0,0);ack();request("Q",2,0,0);if(record!==80'h4c415a59010400020000)$fatal(1,"capability %h",record);
  request("W",7,{24'b0,16'd1,40'd21},0);if(length!=4||response[2]!="2")$fatal(1,"noncontiguous load");
  load(0,21);load(1,2);load(2,{4'd2,4'b0,16'd0,16'd1});load(3,{4'd3,4'b0,16'd2,16'd0});load(4,{4'd1,4'b0,16'd3,16'd3});load(5,{4'd1,4'b0,16'd3,16'd3});load(6,{4'd1,4'b0,16'd4,16'd5});
  request("Q",2,16'h1000,0);if(record[39:0]!=21)$fatal(1,"heap readback");
  request("F",2,6,0);ack();request("T",4,12,0);ack();request("Q",2,1,0);if(record[31:16]==0)$fatal(1,"missing continuations");
  request("Q",2,16'h2000,0);if(record[79:76]!=1)$fatal(1,"continuation read");
  request("W",7,0,0);if(length!=4||response[2]!="2")$fatal(1,"live heap write");
  request("T",4,500,0);ack();request("Q",2,2,0);if(record[39:0]!=168)$fatal(1,"result");
  request("Q",2,16'h1003,0);if(record[39:0]!=42)$fatal(1,"memoized heap");
  request("Q",2,21,0);if(record[31:0]!=1)$fatal(1,"multiplier count");
  request("Q",2,3,0);if(record[79:64]!=2)$fatal(1,"mutation count");
  request("Q",2,16'h3001,0);if(record[79:76]!=3||record[39:36]!=5)$fatal(1,"claim trace %h",record);
  request("Q",2,16'h3003,0);if(record[79:76]!=5||record[39:0]!=42)$fatal(1,"update trace %h",record);
  request("P",0,0,0);if(response[0]!="O"||record[39:0]!=168)$fatal(1,"poll");
  request("F",2,6,0);ack();request("T",4,500,0);ack();request("P",0,0,0);if(record[39:0]!=168)$fatal(1,"repeat");request("Q",2,21,0);if(record[31:0]!=1)$fatal(1,"repeat multiplication");
  request("R",0,0,0);ack();load(0,1);load(1,{4'd3,4'b0,16'd2,16'd0});load(2,{4'd1,4'b0,16'd0,16'd1});request("F",2,1,0);ack();request("T",4,100,0);ack();request("P",0,0,0);if(record[39:0]!={4'd13,4'b0,32'd3})$fatal(1,"cycle result");request("Q",2,16'h1001,0);if(record[39:0]!={4'd13,4'b0,32'd3})$fatal(1,"cycle memoization");
  $display("PASS lazy UART framing, reset, load, readback, stack, trace, sharing, cycle and polling");$finish;
 end
 initial begin #100000000;$fatal(1,"UART timeout");end
endmodule
'''
(repo/'lazy_reducer/tb/lazy_link_tb.sv').write_text(s)
