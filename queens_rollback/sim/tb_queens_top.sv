`timescale 1ns/1ps
module tb_queens_top;
  parameter integer USE_TRAIL=1,FIRST_ONLY=0,TRAIL_CAPACITY=64;
  reg clk=0;always #5 clk=~clk;
  reg button=1;
  wire tx,led;
  queens_top #(.USE_TRAIL(USE_TRAIL),.FIRST_ONLY(FIRST_ONLY),.TRAIL_CAPACITY(TRAIL_CAPACITY),
               .CLK_HZ(1_000_000),.BAUD(250_000)) dut(
    .clk_10m(clk),.fpga_but(button),.uart_rx_pin(1'b1),.uart_tx_pin(tx),.user_led(led));
  integer restart=0;
  reg restarted=0;
  reg [7:0] byte_value;
  initial begin
    if($value$plusargs("restart=%d",restart)) begin end
    if(restart!=0) begin
      wait(dut.rst_n===1);
      @(negedge tx);
      repeat(3) @(negedge clk);
      button=0;
      wait(dut.rst_n===0);$display("RESTART");
      repeat(8) @(negedge clk);
      button=1;restarted=1;
    end
  end
  initial begin
    forever begin
      wait(dut.rst_n===1);
      fork: frame_or_reset
        begin
          @(negedge tx);
          #60;
          for(integer j=0;j<8;j=j+1) begin byte_value[j]=tx;#40;end
          if(dut.rst_n) begin
            if(tx!==1) $fatal(1,"bad UART stop bit");
            $display("BYTE %02x",byte_value);
          end
        end
        begin @(negedge dut.rst_n);end
      join_any
      disable frame_or_reset;
    end
  end
  initial begin
    if(restart!=0) wait(restarted);
    wait(dut.terminal_sent && dut.u_printer.state_q==dut.u_printer.IDLE && dut.uart_ready);
    repeat(20) @(posedge clk);
    if(dut.done && led!==0) $fatal(1,"done LED");
    $display("TOPDONE %0d %0d %0d",dut.done,dut.fault_code,dut.solution_count);
    $finish;
  end
  initial begin repeat(400000) @(posedge clk);$fatal(1,"board watchdog");end
endmodule
