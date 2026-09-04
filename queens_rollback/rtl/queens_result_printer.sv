// Holds an accepted board until all UART frames finish; sends terminal status once.
`default_nettype none
module queens_result_printer(
  input logic clk,rst_n,
  input logic result_valid,
  output logic result_ready,
  input logic [23:0] result_data,
  input logic done,fault_valid,
  input logic [3:0] fault_code,
  input logic [31:0] solution_count,
  input logic uart_ready,
  output logic uart_start,
  output logic [7:0] uart_data,
  output logic terminal_sent
);
  typedef enum logic [1:0] {IDLE,SEND,LAST} state_t;
  state_t state_q;
  logic [31:0] value_q;
  logic [7:0] prefix_q;
  logic [3:0] digits_q,index_q;
  logic [3:0] nibble;
  assign result_ready=(state_q==IDLE)&&uart_ready&&!terminal_sent;
  assign uart_start=(state_q==SEND)&&uart_ready;
  always_comb begin
    nibble=0;uart_data=0;
    if(index_q==0) uart_data=prefix_q;
    else if(index_q==1) uart_data=":";
    else if(index_q==digits_q+2) uart_data=8'h0d;
    else if(index_q==digits_q+3) uart_data=8'h0a;
    else begin
      nibble=4'(value_q >> (4*(digits_q+1-index_q)));
      uart_data=(nibble<10) ? (8'h30+{4'd0,nibble}) : (8'h41+{4'd0,nibble}-8'd10);
    end
  end
  always_ff @(posedge clk or negedge rst_n) begin
    if(!rst_n) begin
      state_q<=IDLE;value_q<=0;prefix_q<=0;digits_q<=0;index_q<=0;terminal_sent<=0;
    end else case(state_q)
      IDLE: if(uart_ready && !terminal_sent) begin
        index_q<=0;
        if(result_valid) begin
          value_q<={8'd0,result_data};prefix_q<="Q";digits_q<=6;state_q<=SEND;
        end else if(done || fault_valid) begin
          terminal_sent<=1;state_q<=SEND;
          if(fault_valid) begin value_q<={28'd0,fault_code};prefix_q<="F";digits_q<=2;end
          else begin value_q<=solution_count;prefix_q<="D";digits_q<=8;end
        end
      end
      SEND: if(uart_ready) begin
        if(index_q==digits_q+3) state_q<=LAST;
        else index_q<=index_q+1'b1;
      end
      LAST: if(uart_ready) state_q<=IDLE;
      default: state_q<=IDLE;
    endcase
  end
endmodule
`default_nettype wire
