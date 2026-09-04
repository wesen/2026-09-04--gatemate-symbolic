// rv_reg.sv — one-entry elastic register (book ch. 15, verbatim contract).
//
// Ready/valid ownership rule: a transfer occurs on a clock edge for which
// both valid and ready are true. While blocked (valid asserted, ready
// deasserted) the producer holds the item stable. Reset invalidates
// occupancy; data_q need not be cleared because invalid data is never
// observed.
`default_nettype none

module rv_reg #(
  parameter int WIDTH = 40
) (
  input  logic             clk,
  input  logic             rst,

  input  logic             in_valid,
  output logic             in_ready,
  input  logic [WIDTH-1:0] in_data,

  output logic             out_valid,
  input  logic             out_ready,
  output logic [WIDTH-1:0] out_data
);

  logic             full_q;
  logic [WIDTH-1:0] data_q;

  assign in_ready  = !full_q || out_ready;
  assign out_valid = full_q;
  assign out_data  = data_q;

  always_ff @(posedge clk) begin
    if (rst) begin
      full_q <= 1'b0;
    end else if (in_ready) begin
      full_q <= in_valid;
      if (in_valid) begin
        data_q <= in_data;
      end
    end
  end

endmodule

`default_nettype wire
