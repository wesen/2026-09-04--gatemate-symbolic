`timescale 1ns/1ps
module tb_graph_link;
 localparam integer BIT_NS=80;
 reg clk=0,rst_n=0,rx=1;
 always #5 clk=~clk;
 wire tx,loaded,done,fault;
 graph_link #(.CLK_HZ(800000),.BAUD(100000),.COMMAND_TIMEOUT(500)) dut(.clk(clk),.rst_n(rst_n),.rx(rx),.tx(tx),.loaded(loaded),.done(done),.fault_valid(fault));
 reg [7:0] load_bytes[0:25];
 reg [7:0] response[0:99];
 integer response_length,negative,restart,step_count;
 reg [1023:0] load_path;
 task automatic send_byte(input [7:0] value);
  begin
   rx=0;#(BIT_NS);
   for(integer b=0;b<8;b=b+1)begin rx=value[b];#(BIT_NS);end
   rx=1;#(BIT_NS);
  end
 endtask
 task automatic receive_line;
  reg [7:0] value;
  begin
   response_length=0;value=0;
   while(value!=10)begin
    @(negedge tx);#(BIT_NS+BIT_NS/2);
    for(integer b=0;b<8;b=b+1)begin value[b]=tx;#(BIT_NS);end
    if(tx!==1)$fatal(1,"UART stop bit");
    if(response_length>=100)$fatal(1,"response overflow");
    response[response_length]=value;response_length=response_length+1;
   end
  end
 endtask
 task automatic send_command(input [7:0] code);
  begin send_byte(code);send_byte(10);end
 endtask
 task automatic print_response;
  begin for(integer j=0;j<response_length;j=j+1)$write("%c",response[j]);end
 endtask
 task automatic load_graph;
  begin fork
   begin for(integer j=0;j<26;j=j+1)send_byte(load_bytes[j]);end
   receive_line();
  join end
 endtask
 reg held;
 reg [63:0] saved_domains;
 reg [7:0] saved_prop;
 reg [3:0] saved_ctop;
 reg [6:0] saved_ttop,saved_base;
 always @(posedge clk)begin
  held=dut.trace_valid&&!dut.trace_ready&&rst_n&&loaded&&!dut.reset_hold;
  saved_domains=dut.domains;saved_prop=dut.propagated;saved_ctop=dut.choice_top;saved_ttop=dut.trail_top;saved_base=dut.trail_base;
  #1;
  if(held&&!dut.reset_hold&&rst_n&&(dut.domains!==saved_domains||dut.propagated!==saved_prop||dut.choice_top!==saved_ctop||dut.trail_top!==saved_ttop||dut.trail_base!==saved_base))$fatal(1,"state changed under event stall");
 end
 initial begin
  if(!$value$plusargs("load=%s",load_path))$fatal(1,"missing load file");
  negative=0;restart=0;
  if($value$plusargs("negative=%d",negative))begin end
  if($value$plusargs("restart=%d",restart))begin end
  $readmemh(load_path,load_bytes);
  repeat(8)@(negedge clk);rst_n=1;repeat(8)@(negedge clk);
  if(negative)begin
   fork send_byte("L");receive_line();join
   if(response_length!=4||response[0]!="!"||response[2]!="1")$fatal(1,"partial command recovery");
   repeat(20)@(negedge clk);
   load_bytes[24]=load_bytes[24]=="0" ? "1" : "0";
   load_graph();
   if(response_length!=4||response[0]!="!"||response[2]!="3")$fatal(1,"checksum rejection");
   $readmemh(load_path,load_bytes);
  end
  load_graph();if(response_length!=2||response[0]!="A")$fatal(1,"load rejected");
  for(step_count=0;step_count<10000;step_count=step_count+1)begin
   repeat(40)@(negedge clk);
   fork send_command("S");receive_line();join
   print_response();
   if(response_length!=70||response[0]!="E")$fatal(1,"expected event");
   if(restart&&step_count==4)begin
    fork send_command("R");receive_line();join
    if(response_length!=2||response[0]!="A")$fatal(1,"reset rejected");
    $display("RESET");restart=0;
   end else if(response[9]=="0"&&(response[10]=="A"||response[10]=="B"))begin
    repeat(100)@(negedge clk);
    fork send_command("S");receive_line();join
    if(response_length!=2||response[0]!="Z")$fatal(1,"terminal stepping");
    $display("DONE");$finish;
   end
  end
  $fatal(1,"event watchdog");
 end
 initial begin #1000000000;$fatal(1,"simulation watchdog");end
endmodule
