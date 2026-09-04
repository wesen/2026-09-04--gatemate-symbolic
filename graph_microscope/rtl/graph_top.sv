`default_nettype none
module graph_top(input wire clk_10m,fpga_but,uart_rx_pin,output wire user_led,uart_tx_pin);
 wire cfg_rst_n,rst_n;
 reg button1,button2;
 CC_USR_RSTN u_cfg(.USR_RSTN(cfg_rst_n));
 always @(posedge clk_10m)begin button1<=fpga_but;button2<=button1;end
 reset_sync u_reset(.clk(clk_10m),.arst_n(cfg_rst_n&&button2),.rst_n(rst_n));
 wire loaded,done,fault_valid;
 graph_link u_link(.clk(clk_10m),.rst_n(rst_n),.rx(uart_rx_pin),.tx(uart_tx_pin),.loaded(loaded),.done(done),.fault_valid(fault_valid));
 reg [23:0] counter;
 always @(posedge clk_10m or negedge rst_n)if(!rst_n)counter<=0;else counter<=counter+1'b1;
 assign user_led=done ? 1'b0 : !loaded ? 1'b1 : fault_valid ? ~counter[21] : ~counter[23];
endmodule
`default_nettype wire
