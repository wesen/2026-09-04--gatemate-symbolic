// tb_top.sv — board-level testbench for top.sv. Decodes the UART pin at
// 115200 8N1 (87 cycles/bit at 10 MHz) and prints one UARTBYTE line per
// received byte, then a TOPDONE line with the final machine state. The
// pytest suite (sim/test_top.py) compares the byte stream against the
// model's EMIT values and the exit criteria.
//
// Plusargs: +rom=path.hex (+max_cycles=N watchdog, default 300000 cycles)
`timescale 1ns/1ps

module tb_top;

  logic clk_10m = 1'b0;
  always #50 clk_10m = ~clk_10m;   // 10 MHz

  logic fpga_but = 1'b1;           // not pressed (active low)
  logic uart_rx_pin = 1'b1;
  wire  user_led;
  wire  uart_tx_pin;

  top #(
    .ROM_DEPTH (1024),
    .DEEP_DEPTH(512)
  ) dut (
    .clk_10m    (clk_10m),
    .fpga_but   (fpga_but),
    .user_led   (user_led),
    .uart_tx_pin(uart_tx_pin),
    .uart_rx_pin(uart_rx_pin)
  );

  localparam int BIT_NS = 8700;   // 10 MHz / 115200 ~ 87 cycles

  logic [7:0] b;
  integer max_cycles = 300000;
  integer restart_cycle = 0;
  logic restarted = 0;

  string hex_file;
  integer i;

  initial begin
    if (!$value$plusargs("rom=%s", hex_file)) begin
      $display("FAIL: missing +rom=<file.hex>");
      $finish;
    end
    void'($value$plusargs("max_cycles=%d", max_cycles));

    for (i = 0; i < 1024; i = i + 1)
      dut.u_rom.mem[i] = 20'h00000;
    $readmemh(hex_file, dut.u_rom.mem);

    void'($value$plusargs("restart_cycle=%d", restart_cycle));

    fork
      begin : uart_monitor
        forever begin
          wait (dut.rst_n === 1'b1);
          fork : frame_or_reset
            begin
              @(negedge uart_tx_pin);
              #(BIT_NS + BIT_NS/2);
              for (int j = 0; j < 8; j = j + 1) begin
                b[j] = uart_tx_pin;
                #BIT_NS;
              end
              if (dut.rst_n) $display("UARTBYTE %02x", b);
            end
            begin
              @(negedge dut.rst_n);
            end
          join_any
          disable frame_or_reset;
        end
      end
      begin : done
        if (restart_cycle > 0) wait (restarted);
        wait (dut.halted || dut.fault_valid);
      end
      begin : watchdog
        repeat (max_cycles) @(posedge clk_10m);
        $display("FAIL: watchdog timeout");
      end
    join_any

    // Drain: after halt/fault, rv_reg and the printer may still be framing
    // up to two accepted values (2 x 13 bytes); wait generously (60 byte
    // times) for the wire to go quiet.
    #(60 * 10 * BIT_NS);

    if (dut.halted && user_led !== 1'b0) $display("FAIL: halted LED is not lit");
    $display("TOPDONE %0d %0d %0d %0d %0d", dut.halted, dut.fault_valid,
             dut.pc_dbg, dut.depth_dbg, dut.rdepth_dbg);
    $finish;
  end

  initial begin
    #1;
    if (restart_cycle > 0) begin
      repeat (restart_cycle) @(posedge clk_10m);
      fpga_but <= 1'b0;
      repeat (8) @(posedge clk_10m);
      $display("RESTART");
      fpga_but <= 1'b1;
      restarted <= 1'b1;
    end
  end
endmodule
