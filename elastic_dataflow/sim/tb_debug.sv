`timescale 1ns/1ps
module tb_debug;
 reg clk=0;always #5 clk=~clk;
 reg rst_n=0,enable=0,in_valid=0,out_ready=0,cancel_valid=0;
 reg [79:0] in_token=0;wire in_ready,out_valid,cancel_ready,trace_valid,quiescent;
 wire [79:0] out_token,trace_token;
 reg [1:0] cancel_context=0;reg [7:0] debug_addr=0;wire [79:0] debug_data;
 reg graph_write=0,graph_commit=0;reg [2:0] graph_index=0;reg [23:0] graph_descriptor=0;
 reg [3:0] graph_size=0;wire graph_writable,graph_acceptable;
 reg debug_control_valid=0;reg [7:0] debug_flags=0,debug_node=255,debug_context=255;wire halted;
 dataflow_core #(.INPUT_DEPTH(1),.COMPLETION_DEPTH(1),.OUTPUT_DEPTH(1)) dut(.*);
 task send(input [7:0] ctx);
 begin @(negedge clk);in_token={ctx,8'b0,6'b0,2'b0,8'hff,8'b0,40'd7};in_valid=1;
 @(posedge clk);while(!in_ready)@(posedge clk);@(negedge clk);in_valid=0;end
 endtask
 initial begin
 repeat(3)@(negedge clk);rst_n=1;
 @(negedge clk);graph_write=1;graph_descriptor=24'h480000;@(negedge clk);graph_write=0;graph_size=1;
 @(negedge clk);if(!graph_acceptable)$fatal(1,"final COPY graph");graph_commit=1;@(negedge clk);graph_commit=0;
 // Fill output and router first; only arm full when the remaining completion fills.
 enable=1;send(0);send(1);send(2);
 repeat(50)@(negedge clk);
 debug_flags=8'h82;debug_control_valid=1;@(negedge clk);debug_control_valid=0;
 repeat(3)@(negedge clk);if(!halted||dut.stop_reason!=2)$fatal(1,"completion full breakpoint");
 if(dut.trace_dropped==0)$fatal(1,"simultaneous events not counted");
 $display("PASS completion-full hardware breakpoint and simultaneous-event loss accounting");$finish;
 end
 initial begin #100000;$fatal(1,"timeout");end
endmodule
