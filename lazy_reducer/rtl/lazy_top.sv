`default_nettype none
module lazy_top(input wire clk_10m,fpga_but,uart_rx_pin,output wire user_led,uart_tx_pin);
 wire cfg_rst_n,rst_n,idle;
 reg button1,button2;
 CC_USR_RSTN configuration_reset(.USR_RSTN(cfg_rst_n));
 always @(posedge clk_10m)begin button1<=fpga_but;button2<=button1;end
 reset_sync reset_button(.clk(clk_10m),.arst_n(cfg_rst_n&&button2),.rst_n(rst_n));
 lazy_link link(.clk(clk_10m),.rst_n(rst_n),.rx(uart_rx_pin),.tx(uart_tx_pin),.idle(idle));
 assign user_led=!idle;
endmodule
`default_nettype wire
