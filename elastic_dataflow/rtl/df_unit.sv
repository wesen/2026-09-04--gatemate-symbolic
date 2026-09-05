`default_nettype none
module df_unit #(parameter integer LATENCY=1)(
 input wire clk,rst_n,ce,input wire in_valid,output wire in_ready,
 input wire [79:0] in_token,input wire [39:0] a,b,input wire [2:0] op,
 output wire out_valid,input wire out_ready,output wire [79:0] out_token,
 output wire [7:0] debug_valid,input wire [2:0] debug_stage,output wire [79:0] debug_token
);
 reg [LATENCY-1:0] valid;
 reg [79:0] token[0:LATENCY-1];
 wire [LATENCY:0] ready;
 wire [39:0] result=dataflow_pkg::evaluate(op,a,b);
 assign ready[LATENCY]=out_ready;
 for(genvar g=0;g<LATENCY;g=g+1)begin:stages
   assign ready[g]=!valid[g]||ready[g+1];
 end
 assign in_ready=ce&&ready[0];
 assign out_valid=valid[LATENCY-1];
 assign out_token=token[LATENCY-1];
 assign debug_valid={{(8-LATENCY){1'b0}},valid};
 assign debug_token=debug_stage<LATENCY?token[debug_stage]:80'b0;
 integer i;
 always @(posedge clk or negedge rst_n)begin
   if(!rst_n)valid<=0;
   else if(ce)begin
     for(i=LATENCY-1;i>0;i=i-1)if(ready[i])begin
       valid[i]<=valid[i-1]; if(valid[i-1])token[i]<=token[i-1];
     end
     if(ready[0])begin valid[0]<=in_valid;if(in_valid)token[0]<={in_token[79:40],result};end
   end
 end
endmodule
`default_nettype wire
