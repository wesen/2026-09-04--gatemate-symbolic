`default_nettype none
module uart_rx #(parameter integer CLK_HZ=10_000_000, BAUD=115200)(
 input wire clk,rst_n,rx,
 output reg valid, framing_error,
 output reg [7:0] data
);
 localparam integer DIV=(CLK_HZ+BAUD/2)/BAUD;
 localparam integer CW=$clog2(DIV+1);
 reg sync1,sync2;
 reg [1:0] state;
 reg [CW-1:0] countdown;
 reg [2:0] bit_index;
 reg [7:0] shift;
 always @(posedge clk or negedge rst_n) begin
  if(!rst_n) begin sync1<=1;sync2<=1;end
  else begin sync1<=rx;sync2<=sync1;end
 end
 always @(posedge clk or negedge rst_n) begin
  if(!rst_n) begin state<=0;countdown<=0;bit_index<=0;shift<=0;data<=0;valid<=0;framing_error<=0;end
  else begin
   valid<=0;framing_error<=0;
   case(state)
    0: if(!sync2) begin state<=1;countdown<=CW'(DIV/2-1);end
    1: if(countdown!=0) countdown<=countdown-1'b1;
       else if(sync2) state<=0;
       else begin state<=2;countdown<=CW'(DIV-1);bit_index<=0;end
    2: if(countdown!=0) countdown<=countdown-1'b1;
       else begin shift[bit_index]<=sync2;countdown<=CW'(DIV-1);if(bit_index==7)state<=3;else bit_index<=bit_index+1'b1;end
    3: if(countdown!=0) countdown<=countdown-1'b1;
       else begin state<=0;if(sync2)begin data<=shift;valid<=1;end else framing_error<=1;end
   endcase
  end
 end
 initial if(DIV<4) $fatal(1,"UART receiver requires at least four clocks per bit");
endmodule
`default_nettype wire
