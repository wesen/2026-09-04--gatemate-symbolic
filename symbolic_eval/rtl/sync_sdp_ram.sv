// sync_sdp_ram.sv — simple dual-port synchronous RAM wrapper (book ch. 6
// "Block RAM discipline"): every memory exposes a request point (wr_en /
// rd_addr presented during a cycle) and a response-valid point (rd_data is
// valid one cycle later). Design logic must never depend on combinational
// read data. Inferred first; inspect synthesis stats before considering a
// GateMate primitive.
`default_nettype none

module sync_sdp_ram #(
  parameter int DEPTH = 512,     // deep stack slots (2 more live in the cache)
  parameter int WIDTH = 40
) (
  input  logic                  clk,
  // write port
  input  logic                  wr_en,
  input  logic [$clog2(DEPTH)-1:0] wr_addr,
  input  logic [WIDTH-1:0]      wr_data,
  // read port (simple dual port: one write + one read per cycle; in this
  // design they never target the same address in the same cycle)
  input  logic [$clog2(DEPTH)-1:0] rd_addr,
  output logic [WIDTH-1:0]      rd_data
);

  logic [WIDTH-1:0] mem [0:DEPTH-1];

  always_ff @(posedge clk) begin
    if (wr_en)
      mem[wr_addr] <= wr_data;
    rd_data <= mem[rd_addr];   // 1-cycle read latency
  end

endmodule

`default_nettype wire
