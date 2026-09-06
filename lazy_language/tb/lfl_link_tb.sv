`timescale 1ns/1ps
module lfl_link_tb;
 localparam BIT_NS=80;
 reg clk=0,rst_n=0,rx=1;always #5 clk=~clk;
 wire tx,idle;
 lfl_link #(.CLK_HZ(800000),.BAUD(100000),.COMMAND_TIMEOUT(500)) dut(.*);
 reg [7:0]response[0:39];integer length;
 reg [127:0]record;
 function automatic [7:0]hex_ascii(input[3:0]v);hex_ascii=v<10?"0"+v:"A"+v-10;endfunction
 function automatic [3:0]hex_value(input[7:0]v);hex_value=v<="9"?4'(v-"0"):4'(v-"A"+10);endfunction
 task send_byte(input[7:0]v);begin rx=0;#(BIT_NS);for(integer b=0;b<8;b=b+1)begin rx=v[b];#(BIT_NS);end rx=1;#(BIT_NS);end endtask
 task receive;
  reg [7:0]v;begin length=0;v=0;while(v!=10)begin @(negedge tx);#(BIT_NS+BIT_NS/2);for(integer b=0;b<8;b=b+1)begin v[b]=tx;#(BIT_NS);end if(tx!==1||length>=40)$fatal(1,"UART framing");response[length]=v;length=length+1;end end
 endtask
 task request(input[7:0]kind,input integer bytes,input[143:0]payload,input badsum);
  reg [7:0]sum,v;
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
   if(length==36)begin
    record=0;sum=0;
    for(integer n=1;n<33;n=n+1)record={record[123:0],hex_value(response[n])};
    for(integer n=0;n<16;n=n+1)sum=sum^record[n*8+:8];
    if(sum!={hex_value(response[33]),hex_value(response[34])})$fatal(1,"response checksum");
   end
  end
 endtask
 task ack;if(length!=2||response[0]!="A")$fatal(1,"expected ACK got %c%c%c length=%d",response[0],response[1],response[2],length);endtask
 task rejected;if(length!=4||response[0]!="!"||response[2]!="2")$fatal(1,"expected state rejection");endtask
 initial begin
  repeat(8)@(negedge clk);rst_n=1;repeat(8)@(negedge clk);
  fork send_byte("C");receive();join
  if(length!=4||response[2]!="1")$fatal(1,"partial timeout");
  request("Q",2,0,1);if(length!=4||response[2]!="3")$fatal(1,"checksum rejection");
  request("R",0,0,0);ack();request("Q",2,0,0);
  if(record!==128'h4c464c31010800020008000040000000)$fatal(1,"signature %h",record);
  request("F",2,15,0);rejected();
  `include "link_cases.svh"
  request("F",2,15,0);ack();request("T",4,12,0);ack();
  request("Q",2,1,0);if(record[63:48]==0)$fatal(1,"missing stack");
  request("Q",2,16'h3000,0);if(record[127:120]!=6)$fatal(1,"UPDATE frame readback");
  request("H",12,0,0);rejected();request("F",2,15,0);rejected();
  request("T",4,2000,0);ack();request("Q",2,1,0);if(record[23:16]!=25)$fatal(1,"result page %h",record);
  request("Q",2,16'h1019,0);if(record[79:0]!==80'd168)$fatal(1,"result object %h",record);
  request("Q",2,29,0);if(record[31:0]!=1)$fatal(1,"multiplication count");
  request("Q",2,3,0);if(record[127:112]!=15)$fatal(1,"trace count %h",record);
  request("P",0,0,0);if(response[0]!="O"||record!=25)$fatal(1,"poll result");
  request("P",0,0,0);if(length!=2||response[0]!="N")$fatal(1,"empty poll");
  request("F",2,15,0);ack();request("T",4,200,0);ack();request("P",0,0,0);if(record!=25)$fatal(1,"repeat demand");
  request("T",4,1000001,0);rejected();
  $display("PASS LFL1 UART framing, checked load/readback, continuations, trace, sharing and polling");$finish;
 end
 initial begin #100000000;$fatal(1,"UART timeout");end
endmodule
