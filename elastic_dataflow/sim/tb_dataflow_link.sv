`timescale 1ns/1ps
module tb_dataflow_link;
 localparam BIT_NS=80;
 reg clk=0,rst_n=0,rx=1;always #5 clk=~clk;
 wire tx,idle;
 dataflow_link #(.CLK_HZ(800000),.BAUD(100000),.COMMAND_TIMEOUT(500)) dut(.*);
 reg [7:0] response[0:31];integer length;
 reg [79:0] record;
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
 task inject(input [5:0] node,input port,input [39:0] value);
   begin request("I",10,{16'b0,node,port,1'b0,8'hff,8'b0,value},0);ack();end
 endtask
 initial begin
   repeat(8)@(negedge clk);rst_n=1;repeat(8)@(negedge clk);
   fork send_byte("I");receive();join
   if(length!=4||response[2]!="1")$fatal(1,"partial timeout");
   request("Q",1,0,1);if(length!=4||response[2]!="3")$fatal(1,"bad checksum accepted");
   request("R",0,0,0);ack();request("Q",1,0,0);
   if(record!==80'h02040704080808080000)$fatal(1,"capabilities %h",record);
   inject(0,0,7);inject(0,1,6);inject(1,0,3);inject(1,1,5);inject(3,0,2);inject(3,1,9);
   request("Q",1,1,0);if(record[39:32]!=6)$fatal(1,"input did not remain paused %h",record);
   request("T",4,200,0);ack();request("Q",1,2,0);if(record[79:48]!=200)$fatal(1,"enabled cycle count %h",record);
   request("Q",1,32,0);if(record!={40'd7,40'd6})$fatal(1,"operand debug page %h",record);
   request("C",1,0,0);if(length!=4||response[2]!="5")$fatal(1,"cancel held output accepted");
   request("P",0,0,0);if(length!=24||response[0]!="O"||record[39:0]!=58)$fatal(1,"wrong result %h",record);
   request("P",0,0,0);if(length!=2||response[0]!="N")$fatal(1,"result polled twice");
   request("C",1,0,0);ack();request("Q",1,1,0);if(record[55:48]!=1)$fatal(1,"cancel epoch %h",record);
   request("C",1,4,0);if(length!=4||response[2]!="2")$fatal(1,"bad context accepted");
   request("R",0,0,0);ack();
   for(integer n=0;n<8;n=n+1)inject(0,0,n);
   request("I",10,0,0);if(length!=4||response[2]!="4")$fatal(1,"full input not rejected");
   request("R",0,0,0);ack();
   request("W",4,32'h00020203,0);ack();
   request("G",1,2,0);if(length!=4||response[2]!="2")$fatal(1,"partial graph accepted");
   request("W",4,32'h01180000,0);ack();request("G",1,2,0);ack();
   request("Q",1,148,0);if(record[31:0]!=32'h00020203)$fatal(1,"descriptor readback");
   inject(0,0,7);inject(0,1,6);request("T",4,100,0);ack();
   request("P",0,0,0);if(record[39:0]!=84||record[63:58]!=1||!record[56])$fatal(1,"programmable UART result %h",record);
   request("W",4,32'h00020203,0);if(length!=4||response[2]!="2")$fatal(1,"late graph write accepted");
   $display("PASS UART pause, ticks, operands, cancellation, polling, reset, bounds, checksum, timeout");$finish;
 end
 initial begin #100000000;$fatal(1,"UART watchdog");end
endmodule
