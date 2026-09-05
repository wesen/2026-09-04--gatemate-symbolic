`default_nettype none
// Small bounded token storage. Debug offsets are relative to the queue head;
// readers must pair debug data with count, because unused entries are unspecified.
module df_fifo #(parameter integer DEPTH=8)(
 input wire clk,rst_n,input wire push,input wire [79:0] din,output wire ready,
 input wire pop,output wire valid,output wire [79:0] dout,
 output reg [7:0] count,input wire [2:0] debug_offset,output wire [79:0] debug_data
);
 localparam integer AW=DEPTH<2?1:$clog2(DEPTH);
 reg [79:0] memory[0:DEPTH-1];
 reg [AW-1:0] head,tail;
 wire put=push&&ready,take=pop&&valid;
 wire [3:0] debug_sum={1'b0,head}+debug_offset;
 wire [3:0] debug_index=debug_sum>=DEPTH?debug_sum-DEPTH:debug_sum;
 assign ready=count<DEPTH;
 assign valid=count!=0;
 assign dout=memory[head];
 assign debug_data=debug_offset<count?memory[debug_index]:80'b0;
 always @(posedge clk or negedge rst_n) begin
   if(!rst_n)begin head<=0;tail<=0;count<=0;end
   else begin
     if(put)begin memory[tail]<=din;tail<=tail==DEPTH-1?0:tail+1'b1;end
     if(take)head<=head==DEPTH-1?0:head+1'b1;
     case({put,take})2'b10:count<=count+1'b1;2'b01:count<=count-1'b1;default:;endcase
   end
 end
endmodule
`default_nettype wire
