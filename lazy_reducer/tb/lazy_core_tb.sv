`timescale 1ns/1ps
module lazy_core_tb;
 reg clk=0;always #5 clk=~clk;
 reg rst_n=0,enable=0,load_valid=0,force_valid=0,result_ready=0,debug_select=0;
 reg [15:0] load_address=0,force_root=0,debug_address=0;
 reg [39:0] load_word=0;
 wire load_ready,force_ready,result_valid,idle;wire[39:0]result;wire[79:0]debug_data;
 lazy_core #(.HEAP_DEPTH(64),.STACK_DEPTH(8),.IND_LIMIT(16)) dut(.*);
 task tick;begin @(negedge clk);enable=1;@(negedge clk);enable=0;end endtask
 task reset;begin @(negedge clk);rst_n=0;repeat(3)@(negedge clk);rst_n=1;end endtask
 task load(input[15:0]a,input[39:0]w);begin @(negedge clk);load_address=a;load_word=w;load_valid=1;@(negedge clk);if(!load_ready&&a!=dut.heap_size-1)$fatal(1,"load rejected");load_valid=0;end endtask
 task force_node(input[15:0]a);begin @(negedge clk);force_root=a;force_valid=1;@(negedge clk);force_valid=0;end endtask
 task finish(input[39:0]want);integer n;begin n=0;while(!result_valid&&n<20000)begin tick();n=n+1;end if(!result_valid||result!==want)$fatal(1,"result %h want %h state %d",result,want,dut.state);if(dut.sp!=0)$fatal(1,"output with stack");repeat(8)tick();if(result!==want)$fatal(1,"held result");end endtask
 task poll;begin @(negedge clk);result_ready=1;@(negedge clk);result_ready=0;end endtask
 initial begin
  reset();load(0,40'd21);load(1,40'd2);load(2,{4'd2,4'b0,16'd0,16'd1});load(3,{4'd3,4'b0,16'd2,16'd0});load(4,{4'd1,4'b0,16'd3,16'd3});load(5,{4'd1,4'b0,16'd3,16'd3});load(6,{4'd1,4'b0,16'd4,16'd5});
  force_node(6);finish(40'd168);if(dut.counters[3]!=1||dut.counters[4]!=1||dut.counters[5]!=1)$fatal(1,"sharing counters");poll();force_node(6);finish(40'd168);if(dut.counters[5]!=1)$fatal(1,"recomputed body");
  reset();load(0,40'd1);load(1,{4'd3,4'b0,16'd2,16'd0});load(2,{4'd1,4'b0,16'd0,16'd1});force_node(1);finish({4'd13,4'b0,32'd3});if(dut.counters[3]!=1||dut.counters[4]!=1)$fatal(1,"cycle unwind");poll();force_node(1);finish({4'd13,4'b0,32'd3});if(dut.counters[3]!=1)$fatal(1,"error not memoized");
  reset();load(0,{4'd4,4'b0,16'd0,16'd0});force_node(0);finish({4'd13,4'b0,32'd4});
  reset();load(0,{8'b0,32'h80000000});load(1,{8'b0,32'hffffffff});load(2,{4'd2,4'b0,16'd0,16'd1});load(3,{4'd3,4'b0,16'd2,16'd0});force_node(3);finish({4'd13,4'b0,32'd6});
  reset();load(0,{8'b0,32'hfffffffd});load(1,40'd7);load(2,{4'd2,4'b0,16'd0,16'd1});force_node(2);finish({8'b0,32'hffffffeb});
  reset();load(0,{4'd3,4'b0,16'd42,16'd0});force_node(0);finish({4'd13,4'b0,32'd1});if(dut.counters[4]!=1)$fatal(1,"address unwind");
  reset();load(0,{4'd1,4'b0,16'd0,16'd0});force_node(0);finish({4'd13,4'b0,32'd2});
  `include "generated_cases.svh"
  $display("PASS lazy core sharing, repeated force, cycles, signed arithmetic, faults, bounds and held output");$finish;
 end
 initial begin #100000000;$fatal(1,"timeout");end
endmodule
