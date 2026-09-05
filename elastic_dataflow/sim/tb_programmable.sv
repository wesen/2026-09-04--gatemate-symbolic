`timescale 1ns/1ps
module tb_programmable;
 reg clk=0;always #5 clk=~clk;
 reg rst_n=0,enable=0,in_valid=0,out_ready=0,cancel_valid=0;
 reg [79:0] in_token=0;
 wire in_ready,out_valid,cancel_ready,trace_valid,quiescent;
 wire [79:0] out_token,trace_token;
 reg [1:0] cancel_context=0;
 reg [7:0] debug_addr=0;wire [79:0] debug_data;
 reg graph_write=0,graph_commit=0;
 reg [2:0] graph_index=0;reg [23:0] graph_descriptor=0;
 reg [3:0] graph_size=0;wire graph_writable,graph_acceptable;
 dataflow_core #(.INPUT_DEPTH(1),.COMPLETION_DEPTH(1),.OUTPUT_DEPTH(1)) dut(.debug_control_valid(1'b0),.debug_flags(8'b0),.debug_node(8'hff),.debug_context(8'hff),.halted(),.*);
 task reset;
 begin @(negedge clk);rst_n=0;enable=0;in_valid=0;out_ready=0;graph_write=0;graph_commit=0;
 repeat(3)@(negedge clk);rst_n=1;end
 endtask
 task stage(input [2:0] n,input [23:0] d);
 begin @(negedge clk);graph_index=n;graph_descriptor=d;graph_write=1;@(negedge clk);graph_write=0;end
 endtask
 task activate(input [3:0] count,input expected);
 begin @(negedge clk);graph_size=count;repeat(3)@(negedge clk);#1;if(graph_acceptable!==expected)$fatal(1,"activation guard count=%d got=%b",count,graph_acceptable);
 graph_commit=1;@(negedge clk);graph_commit=0;end
 endtask
 task send(input [5:0] node,input port,input [31:0] value);
 begin @(negedge clk);in_token={8'b0,8'b0,node,port,1'b0,8'hff,8'b0,8'b0,value};in_valid=1;
 @(posedge clk);while(!in_ready)@(posedge clk);@(negedge clk);in_valid=0;end
 endtask
 task expect_result(input [31:0] value,input [5:0] node);
 begin
 for(integer n=0;n<300&&!out_valid;n=n+1)@(negedge clk);
 if(!out_valid||out_token[31:0]!=value||out_token[63:58]!=node||!out_token[56])$fatal(1,"bad programmable output %h",out_token);
 @(negedge clk);out_ready=1;@(negedge clk);out_ready=0;end
 endtask
 initial begin
 reset();stage(0,24'h010600);activate(4,0);
 if(dut.active_count!=7)$fatal(1,"partial staging changed graph");
 stage(1,24'h210400);stage(2,24'h310700);stage(3,24'h180000);activate(4,1);
 debug_addr=148;#1;if(debug_data[23:0]!=24'h010600)$fatal(1,"graph readback");
 enable=1;send(0,0,7);send(0,1,6);send(1,0,2);send(1,1,9);expect_result(43,3);
 if(graph_writable)$fatal(1,"graph writable after execution");activate(4,0);
 reset();stage(0,24'h020203);stage(1,24'h180000);activate(2,1);
 enable=1;send(0,0,7);send(0,1,6);expect_result(84,1);
 reset();stage(0,24'h020202);stage(1,24'h180000);activate(2,0);
 reset();stage(0,24'h010000);stage(1,24'h180000);activate(2,0);
 reset();stage(0,24'h410400);stage(1,24'h410400);stage(2,24'h180000);activate(3,0);
 reset();stage(0,24'h410300);stage(1,24'h480000);activate(2,0);
 $display("PASS programmable graph: staged activation, dynamic op/finality, generic fanout, readback and guards");$finish;
 end
 initial begin #100000;$fatal(1,"timeout");end
endmodule
