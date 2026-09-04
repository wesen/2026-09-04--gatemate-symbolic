`timescale 1ns/1ps
module tb_queens;
  parameter integer FIRST_ONLY=0, CHOICE_CAPACITY=8;
  reg clk=0; always #5 clk=~clk;
  reg rst_n=0;
  wire valid,done,fault,trace_valid;
  reg ready=1;
  wire [23:0] data;
  wire [3:0] code,kind,ctop;
  wire [63:0] domains;
  wire [7:0] prop;
  wire [6:0] ttop,base;
  wire [31:0] count,cycles,writes,hwrites,hreads,cwrites,stalls;
  queens_core #(.FIRST_ONLY(FIRST_ONLY),.CHOICE_CAPACITY(CHOICE_CAPACITY)) dut(
    .clk(clk),.rst_n(rst_n),.result_valid(valid),.result_ready(ready),.result_data(data),
    .done(done),.fault_valid(fault),.fault_code(code),.solution_count(count),
    .trace_valid(trace_valid),.trace_kind(kind),.domains_o(domains),.propagated_o(prop),
    .choice_top_o(ctop),.trail_top_o(ttop),.trail_base_o(base),.cycles(cycles),
    .domain_writes(writes),.history_writes(hwrites),.history_reads(hreads),
    .choice_writes(cwrites),.result_stalls(stalls));
  integer i,n=0,seed=1,stall_mode=0;
  reg [31:0] rng=1;
  reg held=0;
  reg [23:0] held_data;
  initial begin
    if($value$plusargs("seed=%d",seed)) rng=seed;
    if($value$plusargs("stall=%d",stall_mode)) begin end
    repeat(3) @(negedge clk);
    rst_n=1;
  end
  always @(negedge clk) if(rst_n) begin
    rng=(rng<<1)^((rng[31])?32'h04c11db7:0);
    ready=(stall_mode==0) || (rng[2:0]==0);
  end
  always @(posedge clk) begin
    if(rst_n) begin
      if(held && (!valid || data!==held_data)) $fatal(1,"blocked result changed");
      held=valid&&!ready; held_data=data;
      if(valid&&ready) $display("RESULT %06x",data);
    end else held=0;
    #1;
    if(rst_n) begin
      n=n+1;
      if(trace_valid) begin
        $display("E %0d %016x %02x %0d %0d %0d %0d %06x %0d",kind,domains,prop,ctop,ttop,base,count,data,code);
        $write("C");for(i=0;i<ctop;i=i+1) $write(" %026x",dut.u_choices.mem[i]);$display("");
        $display("T");
      end
      if(done||fault) begin
        $display("STATS %0d %0d %0d %0d %0d %0d",cycles,writes,hwrites,hreads,cwrites,stalls);
        $finish;
      end
      if(n>500000) $fatal(1,"watchdog");
    end
  end
endmodule
