`timescale 1ns/1ps
module lfl_core_tb;
 reg clk=0;always #5 clk=~clk;
 reg rst_n=0,enable=0,load_valid=0,force_valid=0,result_ready=0,debug_select=0;
 reg [2:0]load_kind=0;
 reg [15:0]load_address=0,force_ref=0,debug_address=0;
 reg [127:0]load_data=0;
 wire load_ready,force_ready,result_valid,idle;
 wire [15:0]result_ref;
 wire [127:0]debug_data;
 lfl_core dut(.*);
 task tick;begin @(negedge clk);enable=1;@(negedge clk);enable=0;end endtask
 task reset;begin @(negedge clk);rst_n=0;repeat(3)@(negedge clk);rst_n=1;end endtask
 task load(input [2:0]kind,input[15:0]address,input[127:0]data);begin @(negedge clk);if(!load_ready)$fatal(1,"not load ready");load_kind=kind;load_address=address;load_data=data;load_valid=1;@(negedge clk);load_valid=0;end endtask
 task force_node(input[15:0]address);begin @(negedge clk);if(!force_ready)$fatal(1,"not force ready");force_ref=address;force_valid=1;@(negedge clk);force_valid=0;end endtask
 task finish(input[15:0]want_ref,input[79:0]want);integer n;begin
  n=0;while(!result_valid&&n<100000)begin tick();n=n+1;end
  if(!result_valid||result_ref!==want_ref||dut.heap_ram.mem[result_ref]!==want)$fatal(1,"result ref %d want %d object %h want %h state %d",result_ref,want_ref,dut.heap_ram.mem[result_ref],want,dut.state);
  if(dut.sp!=0)$fatal(1,"nonempty output stack");repeat(3)tick();if(result_ref!==want_ref)$fatal(1,"output changed");
 end endtask
 task poll;begin @(negedge clk);result_ready=1;@(negedge clk);result_ready=0;end endtask
 initial begin
  `include "generated_cases.svh"
  $finish;
 end
 initial begin #100000000;$fatal(1,"timeout");end
endmodule
