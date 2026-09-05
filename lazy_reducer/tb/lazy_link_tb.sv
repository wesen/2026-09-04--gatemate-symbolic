`timescale 1ns/1ps
module lazy_link_tb;
 localparam BIT_NS=80;
 reg clk=0,rst_n=0,rx=1;always #5 clk=~clk;
 wire tx,idle;
 lazy_link #(.CLK_HZ(800000),.BAUD(100000),.COMMAND_TIMEOUT(500)) dut(.*);
 reg [7:0] response[0:31];integer length;
 reg [79:0] record;integer stopped_cycle,trace_entries;reg found_issue;
 function automatic [7:0] hex_ascii(input [3:0] v);hex_ascii=v<10?"0"+v:"A"+v-10;endfunction
 function automatic [3:0] hex_value(input [7:0] v);hex_value=v<="9"?4'(v-"0"):4'(v-"A"+10);endfunction
 task send_byte(input [7:0] v);
   begin rx=0;#(BIT_NS);for(integer b=0;b<8;b=b+1)begin rx=v[b];#(BIT_NS);end rx=1;#(BIT_NS);end
 endtask
 task receive;
   reg [7:0] v;
   begin length=0;v=0;
     while(v!=10)begin
       @(negedge tx);#(BIT_NS+BIT_NS/2);for(integer b=0;b<8;b=b+1)begin v[b]=tx;#(BIT_NS);end
       if(tx!==1||length>=32)$fatal(1,"UART framing");response[length]=v;length=length+1;
     end
   end
 endtask
 task request(input [7:0] kind,input integer bytes,input [79:0] payload,input badsum);
   reg [7:0] sum;reg [7:0] v;
   begin
     repeat(20)@(negedge clk);
     fork
       begin
         send_byte(kind);sum=0;
         for(integer n=bytes-1;n>=0;n=n-1)begin v=8'(payload>>(8*n));sum=sum^v;send_byte(hex_ascii(v[7:4]));send_byte(hex_ascii(v[3:0]));end
         if(bytes!=0)begin sum=sum^{7'b0,badsum};send_byte(hex_ascii(sum[7:4]));send_byte(hex_ascii(sum[3:0]));end
         send_byte(10);
       end
       receive();
     join
     if(length==24)begin
       record=0;sum=0;
       for(integer n=1;n<21;n=n+1)record={record[75:0],hex_value(response[n])};
       for(integer n=0;n<10;n=n+1)sum=sum^record[n*8+:8];
       if(sum!={hex_value(response[21]),hex_value(response[22])})$fatal(1,"response checksum");
     end
   end
 endtask
 task ack;
   if(length!=2||response[0]!="A")$fatal(1,"expected ACK got %c%c%c length=%0d",response[0],response[1],response[2],length);
 endtask
 task load(input [15:0] a,input [39:0] w);begin request("W",7,{24'b0,a,w},0);ack();end endtask
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
