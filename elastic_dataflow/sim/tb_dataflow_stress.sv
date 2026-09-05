`timescale 1ns/1ps
module tb_dataflow_stress;
 parameter DEPTH=1,LATENCY=4;
 reg clk=0;always #5 clk=~clk;
 reg rst_n=0,enable=1,in_valid=0,out_ready=0;
 wire in_ready,out_valid,trace_valid,quiescent,cancel_ready;
 reg [79:0] in_token=0;wire [79:0] out_token,trace_token,debug_data;
 dataflow_core #(.INPUT_DEPTH(DEPTH),.COMPLETION_DEPTH(DEPTH),.OUTPUT_DEPTH(DEPTH),.MUL_LATENCY(LATENCY)) dut(
 .clk(clk),.rst_n(rst_n),.enable(enable),.in_valid(in_valid),.in_ready(in_ready),.in_token(in_token),.out_ready(out_ready),.out_valid(out_valid),.out_token(out_token),
 .cancel_valid(1'b0),.cancel_context(2'b0),.cancel_ready(cancel_ready),.graph_write(1'b0),.graph_commit(1'b0),.graph_index(3'b0),.graph_descriptor(24'b0),.graph_size(4'b0),.graph_writable(),.graph_acceptable(),.debug_addr(8'b0),.debug_data(debug_data),.trace_valid(trace_valid),.trace_token(trace_token),.quiescent(quiescent));
 reg [31:0] random_state=32'h86564731;
 function automatic [31:0] next_random(input [31:0] x);next_random={x[30:0],x[31]^x[21]^x[1]^x[0]};endfunction
 reg [79:0] tokens[0:23],tmp;
 reg [79:0] held_output,held_mul,held_alu;
 reg hold_output=0,hold_mul=0,hold_alu=0;
 integer value,which,sent,results,acts,cycle;
 always @(posedge clk)begin
   if(!rst_n)begin hold_output=0;hold_mul=0;hold_alu=0;end
   else begin
     if(hold_output&&(!out_valid||out_token!==held_output))$fatal(1,"output stability");
     if(hold_mul&&(!dut.mul_valid||dut.mul_token!==held_mul))$fatal(1,"multiply exit stability");
     if(hold_alu&&(!dut.alu_valid||dut.alu_token!==held_alu))$fatal(1,"ALU exit stability");
     hold_output=out_valid&&!out_ready;held_output=out_token;
     hold_mul=dut.mul_valid&&!(dut.ce&&dut.mul_take);held_mul=dut.mul_token;
     hold_alu=dut.alu_valid&&!(dut.ce&&dut.alu_take);held_alu=dut.alu_token;
     if(in_valid&&in_ready)begin $display("SRC %020h",in_token);sent=sent+1;end
     if(out_valid&&out_ready)begin $display("OUT %020h",out_token);results=results+1;end
     if(trace_valid)begin $display("ACT %020h",trace_token);acts=acts+1;end
   end
 end
 initial begin
   for(integer test=0;test<50;test=test+1)begin
     @(negedge clk);rst_n=0;in_valid=0;out_ready=0;enable=1;
     repeat(3)@(negedge clk);$display("CASE %0d",test);rst_n=1;sent=0;results=0;acts=0;
     for(integer c=0;c<4;c=c+1)for(integer p=0;p<6;p=p+1)begin
       random_state=next_random(random_state);value=integer'(random_state%201)-100;
       tokens[c*6+p]={8'(c),8'b0,6'(p<2?0:p<4?1:3),1'(p%2),1'b0,8'hff,8'b0,8'b0,32'(value)};
     end
     for(integer n=23;n>0;n=n-1)begin random_state=next_random(random_state);which=random_state%(n+1);tmp=tokens[n];tokens[n]=tokens[which];tokens[which]=tmp;end
     cycle=0;
     while(cycle<10000&&(sent<24||!quiescent))begin
       @(negedge clk);random_state=next_random(random_state);
       // Keep a presented token stable until accepted, even across pauses.
       in_valid=sent<24;in_token=sent<24?tokens[sent]:80'b0;
       enable=random_state[2:0]!=0;out_ready=random_state[5:3]==0;cycle=cycle+1;
     end
     @(negedge clk);in_valid=0;out_ready=0;
     if(!quiescent||results!=4||acts!=24)$fatal(1,"stress did not drain: sent=%0d outputs=%0d acts=%0d",sent,results,acts);
     $display("END %0d %0d",test,cycle);
   end
   $display("PASS stress DEPTH=%0d LATENCY=%0d",DEPTH,LATENCY);$finish;
 end
endmodule
