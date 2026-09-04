// top.sv — Laboratory 1 evaluator on the Olimex GateMateA1-EVB.
//
// Structure (design doc §6.4):
//   CC_USR_RSTN -> reset_sync -> stack_core_bram (BRAM + top cache)
//   program_rom (1K x 20, zero-padded hex image)
//   EMIT output channel -> value printer -> uart_tx (115200 8N1)
//
// Output format: one EMIT value becomes 13 UART bytes
//   'T' <tag-hex-digit> ':' <8 payload hex digits, MSB first> CR LF
// e.g. BOOL(true) -> "T1:00000001\r\n".
//
// LED evidence:
//   running: slow blink (counter[23], ~0.6 Hz)
//   halted (success): solid ON
//   faulted: fast blink (counter[21], ~2.4 Hz)
//
// The FPGA button (active-low) is a global experiment abort (DR-4): pressing
// it async-asserts reset; releasing it synchronously restarts the machine
// from the initial image.
`default_nettype none

module top #(
  parameter int ROM_DEPTH  = 1024,
  parameter int DEEP_DEPTH = 512
) (
  input  logic clk_10m,
  input  logic fpga_but,      // active-low user button (restart)
  output logic user_led,
  output logic uart_tx_pin,
  input  logic uart_rx_pin    // unused (future host input)
);

  // ----------------------------------------------------------- resets
  logic cfg_rst_n;
  logic but_sync1, but_sync2;
  logic arst_n;
  logic rst_n;

  CC_USR_RSTN u_cfg_reset (
    .USR_RSTN(cfg_rst_n)
  );

  // 2FF button synchronizer; fpga_but is active low.
  always_ff @(posedge clk_10m) begin
    but_sync1 <= fpga_but;
    but_sync2 <= but_sync1;
  end

  assign arst_n = cfg_rst_n & but_sync2;   // async assert, sync release

  reset_sync u_reset_sync (
    .clk    (clk_10m),
    .arst_n (arst_n),
    .rst_n  (rst_n)
  );

  // ------------------------------------------------------------ core
  logic [$clog2(ROM_DEPTH)-1:0] rom_addr;
  logic [19:0]                  rom_data;

  logic        out_valid;
  logic [39:0] out_data;
  logic        out_ready;

  logic        fault_valid;
  logic        halted;
  logic [$clog2(ROM_DEPTH+1)-1:0] pc_dbg;
  logic [$clog2(DEEP_DEPTH+3)-1:0] depth_dbg;
  logic [$clog2(16+1)-1:0] rdepth_dbg;

  stack_core_bram #(
    .ROM_DEPTH (ROM_DEPTH),
    .DEEP_DEPTH(DEEP_DEPTH)
  ) u_core (
    .clk    (clk_10m),
    .rst_n  (rst_n),
    .rom_addr(rom_addr),
    .rom_data(rom_data),
    .trace_fetch(),
    .fault_fetch(),
    .trace_valid(),
    .trace_seq(),
    .trace_pc_old(),
    .trace_pc_new(),
    .trace_op(),
    .trace_depth(),
    .trace_event(),
    .trace_fault(),
    .trace_tag1(),
    .trace_tag0(),
    .trace_out_tag(),
    .trace_out_payload(),
    .out_valid(out_valid),
    .out_data (out_data),
    .out_ready(out_ready),
    .fault_valid(fault_valid),
    .fault_code(),
    .fault_pc(),
    .fault_op(),
    .fault_depth(),
    .fault_tag1(),
    .fault_tag0(),
    .halted(halted),
    .pc_o(pc_dbg),
    .depth_o(depth_dbg),
    .rdepth_o(rdepth_dbg),
    .tc_o(),
    .dc_o()
  );

  program_rom #(
    .DEPTH(ROM_DEPTH)
  ) u_rom (
    .clk (clk_10m),
    .addr(rom_addr),
    .data(rom_data)
  );

  // ------------------------------------------ elastic output stage (rv_reg)
  // The book's one-entry elastic register decouples the core's EMIT
  // commitment from the (slow) UART printer: the core pops when rv_reg
  // accepts, and rv_reg holds the value stable while the printer frames it.
  logic        rv_out_valid;
  logic        rv_out_ready;
  logic [39:0] rv_out_data;
  logic        rv_rst;

  assign rv_rst = ~rst_n;

  rv_reg #(
    .WIDTH(40)
  ) u_rv (
    .clk      (clk_10m),
    .rst      (rv_rst),
    .in_valid (out_valid),
    .in_ready (out_ready),
    .in_data  (out_data),
    .out_valid(rv_out_valid),
    .out_ready(rv_out_ready),
    .out_data (rv_out_data)
  );

  // ------------------------------------------- value printer over UART
  logic       uart_start;
  logic [7:0] uart_data;
  logic       uart_ready;

  uart_tx #(
    .CLK_HZ(10_000_000),
    .BAUD  (115_200)
  ) u_uart (
    .clk   (clk_10m),
    .rst_n (rst_n),
    .start (uart_start),
    .data  (uart_data),
    .ready (uart_ready),
    .tx    (uart_tx_pin)
  );

  // 13 bytes: 'T' tag ':' d7 d6 d5 d4 d3 d2 d1 d0 CR LF
  localparam int BYTES = 13;

  function automatic logic [7:0] byte_of(input logic [39:0] v,
                                        input int unsigned idx);
    logic [7:0] r;
    logic [3:0] nib;
    r = 8'h3F;   // '?'
    if (idx == 0)
      r = "T";
    else if (idx == 2)
      r = ":";
    else if (idx == 11)
      r = 8'h0D;
    else if (idx >= 12)
      r = 8'h0A;
    else begin
      if (idx == 1)
        nib = v[39:36];               // tag
      else
        nib = v[31-4*(idx-3) -: 4];    // payload hex, MSB first
      case (nib)
        4'h0: r = "0";  4'h1: r = "1";  4'h2: r = "2";
        4'h3: r = "3";  4'h4: r = "4";  4'h5: r = "5";
        4'h6: r = "6";  4'h7: r = "7";  4'h8: r = "8";
        4'h9: r = "9";  4'hA: r = "A";  4'hB: r = "B";
        4'hC: r = "C";  4'hD: r = "D";  4'hE: r = "E";
        default: r = "F";
      endcase
    end
    byte_of = r;
  endfunction

  typedef enum logic [1:0] {V_IDLE, V_SEND, V_LAST} vstate_t;
  vstate_t vstate_q;
  logic [3:0]  vidx_q;
  logic [39:0] vval_q;

  // The printer accepts a value only when the UART is idle; rv_reg then
  // holds the item stable while bytes are framed (Delayed Irreversible
  // Store boundary on real hardware).
  assign rv_out_ready = (vstate_q == V_IDLE) && uart_ready;

  // uart_start is a one-cycle pulse: V_SEND issues a byte whenever the
  // transmitter is ready; uart_tx then goes busy, so it cannot retrigger.
  assign uart_start = (vstate_q == V_SEND) && uart_ready;
  assign uart_data  = byte_of(vval_q, vidx_q);

  always_ff @(posedge clk_10m or negedge rst_n) begin
    if (!rst_n) begin
      vstate_q <= V_IDLE;
      vidx_q   <= '0;
      vval_q   <= '0;
    end else begin
      unique case (vstate_q)
        V_IDLE: begin
          if (rv_out_valid && rv_out_ready) begin
            vval_q   <= rv_out_data;
            vidx_q   <= '0;
            vstate_q <= V_SEND;
          end
        end
        V_SEND: begin
          if (uart_ready) begin
            vidx_q <= vidx_q + 1'b1;
            if (vidx_q == BYTES-1)
              vstate_q <= V_LAST;
          end
        end
        V_LAST: begin
          if (uart_ready)
            vstate_q <= V_IDLE;
        end
      endcase
    end
  end

  // ------------------------------------------------------------- LED
  logic [23:0] led_counter;

  always_ff @(posedge clk_10m or negedge rst_n) begin
    if (!rst_n)
      led_counter <= '0;
    else
      led_counter <= led_counter + 24'd1;
  end

  // The EVB user LED is active-LOW (LiteX names the pin user_led_n):
  // driving the pin low lights it. led_logic is the logical LED state
  // (1 = lit); the pin is its inversion.
  logic led_logic;

  always_comb begin
    if (halted)
      led_logic = 1'b1;               // success: solid ON (pin low)
    else if (fault_valid)
      led_logic = led_counter[21];     // fault: fast blink
    else
      led_logic = led_counter[23];     // running: slow blink
  end

  assign user_led = ~led_logic;

endmodule

`default_nettype wire
