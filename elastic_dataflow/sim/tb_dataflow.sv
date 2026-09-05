`timescale 1ns/1ps
module tb_dataflow;
 import dataflow_pkg::*;
 parameter DEPTH=1,LATENCY=4;
 reg clk=0;always #5 clk=~clk;
 reg rst_n=0,enable=1,in_valid=0,out_ready=0,cancel_valid=0;
 reg [79:0] in_token=0;
 wire in_ready,out_valid,cancel_ready,trace_valid,quiescent;
 wire [79:0] out_token,trace_token;
 reg [1:0] cancel_context=0;
 reg [7:0] debug_addr=0;wire [79:0] debug_data;
 dataflow_core #(.INPUT_DEPTH(DEPTH),.COMPLETION_DEPTH(DEPTH),.OUTPUT_DEPTH(DEPTH),.MUL_LATENCY(LATENCY),.EPOCH_BITS(2)) dut(.*);
 integer outputs=0,activations=0,ticks=0,seen[0:3][0:3][0:6];
 reg check_stability=0;reg [79:0] held;
 reg [31:0] random_state=32'h2547321;
 reg auto_poll=0;
 integer expected[0:3];
 task reset;
   begin
     @(negedge clk);rst_n=0;enable=1;in_valid=0;out_ready=0;cancel_valid=0;auto_poll=0;
     repeat(3)@(negedge clk);rst_n=1;outputs=0;activations=0;
     for(integer c=0;c<4;c=c+1)for(integer ep=0;ep<4;ep=ep+1)for(integer n=0;n<7;n=n+1)seen[c][ep][n]=0;
   end
 endtask
 task send(input [7:0] ctx,ep,input [5:0] n,input p,input [39:0] value);
   begin
     @(negedge clk);in_token={ctx,ep,n,p,1'b0,8'hff,8'b0,value};in_valid=1;
     @(posedge clk);while(!in_ready)@(posedge clk);
     @(negedge clk);in_valid=0;
   end
 endtask
 task inputs(input [7:0] ctx,ep,input integer a,b,c,d,e,f);
   begin send(ctx,ep,0,0,{8'b0,32'(a)});send(ctx,ep,3,1,{8'b0,32'(f)});send(ctx,ep,1,1,{8'b0,32'(d)});
     send(ctx,ep,0,1,{8'b0,32'(b)});send(ctx,ep,3,0,{8'b0,32'(e)});send(ctx,ep,1,0,{8'b0,32'(c)});end
 endtask
 task cancel(input [1:0] ctx);
   begin @(negedge clk);cancel_context=ctx;cancel_valid=1;#1;if(!cancel_ready)$fatal(1,"unexpected cancel refusal");
     @(negedge clk);cancel_valid=0;end
 endtask
 task drain(input integer count);
   begin
     auto_poll=1;
     for(integer t=0;t<5000&&!quiescent;t=t+1)@(negedge clk);
     if(!quiescent||outputs!=count)$fatal(1,"drain failed count=%0d expected=%0d pending=%h issue=%d",outputs,count,dut.pending,dut.issue_phase);
     auto_poll=0;out_ready=0;
   end
 endtask
 always @(negedge clk)begin
   random_state={random_state[30:0],random_state[31]^random_state[21]^random_state[1]^random_state[0]};
   if(auto_poll)out_ready=random_state[2:0]==0;
 end
 always @(posedge clk)begin
   ticks=ticks+1;if(ticks>100000)$fatal(1,"global timeout");
   if(!rst_n)check_stability=0;
   else begin
     if(check_stability&&(!out_valid||out_token!==held))$fatal(1,"blocked output changed");
     check_stability=out_valid&&!out_ready;held=out_token;
     if(out_valid&&out_ready)begin
       if(out_token[71:64]!=dut.epoch[out_token[73:72]])$fatal(1,"stale output escaped");
       if($signed(out_token[31:0])!=expected[out_token[73:72]])$fatal(1,"bad output %h expected %0d",out_token,expected[out_token[73:72]]);
       outputs=outputs+1;
     end
     if(trace_valid)begin
       if(seen[trace_token[73:72]][trace_token[65:64]][trace_token[60:58]])$fatal(1,"duplicate activation %h",trace_token);
       seen[trace_token[73:72]][trace_token[65:64]][trace_token[60:58]]=1;activations=activations+1;
       $display("ACT %020h",trace_token);
     end
   end
 end
 initial begin
   reset();expected[0]=58;expected[1]=12;
   send(0,0,0,0,7);send(1,0,3,1,1);send(0,0,1,1,5);send(1,0,0,0,10);
   send(0,0,3,1,9);send(1,0,0,1,40'h00fffffffe);send(0,0,0,1,6);send(1,0,3,0,7);
   send(0,0,1,0,3);send(1,0,1,0,4);send(0,0,3,0,2);send(1,0,1,1,8);
   wait(out_valid);repeat(100)@(negedge clk);
   cancel_context=out_token[73:72];#1;if(cancel_ready)$fatal(1,"cancel accepted offered output");
   drain(2);if(activations!=12)$fatal(1,"expected twelve activations");
   $display("PASS interleave + held output");
   reset();expected[2]=12;
   send(2,0,0,0,7);send(2,0,0,1,6);
   wait(dut.mul_stages!=0);@(negedge clk);enable=0;
   cancel(2);send(2,0,1,0,3);enable=1;
   inputs(2,1,10,-2,4,8,7,1);drain(1);
   if(dut.metrics[5]<2)$fatal(1,"stale work not counted");
   $display("PASS inflight cancellation");
   reset();expected[0]=78;
   send(0,0,6,0,7);send(0,0,0,1,6);send(0,0,1,1,5);send(0,0,3,0,2);send(0,0,3,1,9);
   drain(1);if(activations!=7)$fatal(1,"COPY activation count");
   $display("PASS COPY fanout");
   reset();expected[0]=2;expected[1]=12;
   send(0,0,0,0,7);send(0,0,0,0,8);inputs(1,0,10,-2,4,8,7,1);drain(2);
   $display("PASS isolated duplicate fault");
   reset();expected[0]=58;
   // Inspect a different RAM address between every enabled cycle. Operand
   // response ownership must survive an arbitrary pause in any issue phase.
   inputs(0,0,7,6,3,5,2,9);
   for(integer t=0;t<200&&!out_valid;t=t+1)begin
     @(negedge clk);enable=0;debug_addr=59;repeat(4)@(negedge clk);
     enable=1;@(negedge clk);
   end
   enable=1;drain(1);$display("PASS paused operand inspection");
   reset();cancel(0);cancel(0);cancel(0);enable=0;send(1,0,0,0,5);
   cancel_context=0;#1;if(cancel_ready)$fatal(1,"wrap accepted before drain");
   enable=1;repeat(4)@(negedge clk);cancel(0);if(dut.epoch[0]!=0)$fatal(1,"epoch wrap failed");
   $display("PASS epoch drain guard");
   $display("PASS dataflow DEPTH=%0d LATENCY=%0d",DEPTH,LATENCY);$finish;
 end
endmodule
