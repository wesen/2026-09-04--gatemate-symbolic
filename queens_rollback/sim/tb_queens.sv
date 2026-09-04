`timescale 1ns/1ps
module tb_queens;
  parameter integer FIRST_ONLY=0, CHOICE_CAPACITY=8, USE_TRAIL=0, TRAIL_CAPACITY=64;
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
  queens_core #(.FIRST_ONLY(FIRST_ONLY),.CHOICE_CAPACITY(CHOICE_CAPACITY),
               .USE_TRAIL(USE_TRAIL),.TRAIL_CAPACITY(TRAIL_CAPACITY)) dut(
    .clk(clk),.rst_n(rst_n),.result_valid(valid),.result_ready(ready),.result_data(data),
    .done(done),.fault_valid(fault),.fault_code(code),.solution_count(count),
    .trace_valid(trace_valid),.trace_kind(kind),.domains_o(domains),.propagated_o(prop),
    .choice_top_o(ctop),.trail_top_o(ttop),.trail_base_o(base),.cycles(cycles),
    .domain_writes(writes),.history_writes(hwrites),.history_reads(hreads),
    .choice_writes(cwrites),.result_stalls(stalls));
  integer i,n=0,seed=1,stall_mode=0,restart_mode=0,inject_mode=0,terminal_cycles=0;
  integer blocked_cycles=0;
  reg [31:0] rng=1;
  reg held=0;
  reg [23:0] held_data;
  reg prior_valid=0, injected=0;
  reg [63:0] prior_domains;
  reg [7:0] prior_prop;
  reg [3:0] prior_ctop;
  reg [6:0] prior_ttop,prior_base;
  reg [31:0] prior_count;
  reg [103:0] prior_choices[0:7];
  reg [19:0] prior_trail[0:63];
  initial begin
    if($value$plusargs("seed=%d",seed)) rng=seed;
    if($value$plusargs("stall=%d",stall_mode)) begin end
    if($value$plusargs("restart=%d",restart_mode)) begin end
    if($value$plusargs("inject=%d",inject_mode)) begin end
    repeat(3) @(negedge clk);
    rst_n=1;
    if(restart_mode!=0) begin
      case(restart_mode)
        1: wait(dut.state_q==dut.S_LOG_WRITE);
        2: wait(dut.state_q==dut.S_TAPPLY);
        3: wait(valid);
        4: wait(dut.state_q==dut.S_CP_WRITE);
      endcase
      @(negedge clk);rst_n=0;$display("RESTART");
      repeat(3) @(negedge clk);
      rst_n=1;
    end
    if(inject_mode!=0) begin
      if(inject_mode==1) begin
        wait(dut.state_q==dut.S_CP_WAIT);
        @(negedge clk);dut.u_choices.mem[ctop-1][28:22]=127;
      end else if(inject_mode==2) begin
        wait(dut.state_q==dut.S_PROP);
        @(negedge clk);dut.domains_q[dut.source_q]=8'h03;
      end else if(inject_mode==3) begin
        wait(dut.state_q==dut.S_TWAIT);
        @(negedge clk);dut.u_trail.mem[ttop-1][3:0]=4'hf;
      end
      injected=1;$display("INJECT");
    end
  end
  always @(negedge clk) if(rst_n) begin
    rng=(rng<<1)^((rng[31])?32'h04c11db7:0);
    if(stall_mode==2 && valid && blocked_cycles<200) begin
      ready=0;blocked_cycles=blocked_cycles+1;
    end else ready=(stall_mode!=1) || (rng[2:0]==0);
  end
  always @(posedge clk) begin
    if(rst_n) begin
      if(held && (!valid || data!==held_data)) $fatal(1,"blocked result changed");
      held=valid&&!ready; held_data=data;
      if(valid&&ready) $display("RESULT %06x",data);
    end else begin held=0;prior_valid=0;terminal_cycles=0;end
    #1;
    if(rst_n) begin
      n=n+1;
      if(prior_valid && !injected && (!trace_valid || kind==queens_types_pkg::E_FAULT)) begin
        if(domains!==prior_domains || prop!==prior_prop || ctop!==prior_ctop ||
           ttop!==prior_ttop || base!==prior_base || count!==prior_count)
          $fatal(1,"architectural state changed without successful event");
        for(i=0;i<ctop;i=i+1)
          if(dut.u_choices.mem[i]!==prior_choices[i]) $fatal(1,"live choice changed without event");
        for(i=0;i<ttop;i=i+1)
          if(dut.u_trail.mem[i]!==prior_trail[i]) $fatal(1,"live trail changed without event");
      end
      prior_valid=1;injected=0;
      prior_domains=domains;prior_prop=prop;prior_ctop=ctop;prior_ttop=ttop;prior_base=base;prior_count=count;
      for(i=0;i<ctop;i=i+1) prior_choices[i]=dut.u_choices.mem[i];
      for(i=0;i<ttop;i=i+1) prior_trail[i]=dut.u_trail.mem[i];
      if(terminal_cycles>0 && (trace_valid || valid)) $fatal(1,"event after terminal state");
      if(trace_valid) begin
        $display("E %0d %016x %02x %0d %0d %0d %0d %06x %0d",kind,domains,prop,ctop,ttop,base,count,data,code);
        $write("C");for(i=0;i<ctop;i=i+1) $write(" %026x",dut.u_choices.mem[i]);$display("");
        $write("T");for(i=0;i<ttop;i=i+1) $write(" %05x",dut.u_trail.mem[i]);$display("");
      end
      if(done||fault) begin
        if(terminal_cycles==0) $display("STATS %0d %0d %0d %0d %0d %0d",cycles,writes,hwrites,hreads,cwrites,stalls);
        terminal_cycles=terminal_cycles+1;
        if(terminal_cycles==65) $finish;
      end
      if(n>500000) $fatal(1,"watchdog");
    end
  end
endmodule
