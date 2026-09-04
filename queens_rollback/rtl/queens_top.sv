`default_nettype none
module queens_top #(
  parameter bit USE_TRAIL=1, FIRST_ONLY=0,
  parameter int TRAIL_CAPACITY=64, CHOICE_CAPACITY=8,
  parameter int CLK_HZ=10_000_000, BAUD=115_200
) (
  input logic clk_10m, fpga_but, uart_rx_pin,
  output logic user_led, uart_tx_pin
);
  wire cfg_rst_n,rst_n;
  logic button1,button2;
  CC_USR_RSTN u_cfg(.USR_RSTN(cfg_rst_n));
  always_ff @(posedge clk_10m) begin button1<=fpga_but;button2<=button1;end
  reset_sync u_reset(.clk(clk_10m),.arst_n(cfg_rst_n & button2),.rst_n(rst_n));
  wire result_valid,result_ready,done,fault_valid;
  wire [23:0] result_data;
  wire [31:0] solution_count;
  wire [3:0] fault_code;
  queens_core #(.USE_TRAIL(USE_TRAIL),.FIRST_ONLY(FIRST_ONLY),
    .TRAIL_CAPACITY(TRAIL_CAPACITY),.CHOICE_CAPACITY(CHOICE_CAPACITY)) u_core(
    .clk(clk_10m),.rst_n(rst_n),.result_valid(result_valid),.result_ready(result_ready),
    .result_data(result_data),.done(done),.fault_valid(fault_valid),.fault_code(fault_code),
    .solution_count(solution_count));
  wire uart_ready,uart_start,terminal_sent;
  wire [7:0] uart_data;
  queens_result_printer u_printer(.clk(clk_10m),.rst_n(rst_n),
    .result_valid(result_valid),.result_ready(result_ready),.result_data(result_data),
    .done(done),.fault_valid(fault_valid),.fault_code(fault_code),.solution_count(solution_count),
    .uart_ready(uart_ready),.uart_start(uart_start),.uart_data(uart_data),.terminal_sent(terminal_sent));
  uart_tx #(.CLK_HZ(CLK_HZ),.BAUD(BAUD)) u_uart(.clk(clk_10m),.rst_n(rst_n),
    .start(uart_start),.data(uart_data),.ready(uart_ready),.tx(uart_tx_pin));
  logic [23:0] led_counter;
  always_ff @(posedge clk_10m or negedge rst_n)
    if(!rst_n) led_counter<=0;else led_counter<=led_counter+1'b1;
  assign user_led=done ? 1'b0 : fault_valid ? ~led_counter[21] : ~led_counter[23];
endmodule
`default_nettype wire
