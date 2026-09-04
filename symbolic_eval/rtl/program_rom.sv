// program_rom.sv — 20-bit instruction ROM, synchronous read (one-cycle
// latency) so it infers block RAM (book ch. 6 "Block RAM discipline").
//
// Initialization: in simulation the testbench loads top.u_rom.mem directly
// via $readmemh (the .hex file is zero-padded to DEPTH by the assembler, so
// every word is defined). In synthesis, pass -DPROG_HEX='"file.hex"' to the
// tool chain (MATE-16 pattern).
`default_nettype none

module program_rom #(
  parameter int DEPTH = 1024
) (
  input  logic                       clk,
  input  logic [$clog2(DEPTH)-1:0]   addr,
  output logic [19:0]                data
);

  logic [19:0] mem [0:DEPTH-1];

`ifdef PROG_HEX
  initial begin
    $readmemh(`PROG_HEX, mem);
  end
`endif

  always_ff @(posedge clk) begin
    data <= mem[addr];
  end

endmodule

`default_nettype wire
