// reset_sync.sv — async-assert, synchronous-release reset (textbook §1.6.2).
// GateMate CC_USR_RSTN indicates end of configuration; we release internal
// reset only on a rising clk edge so all state starts synchronously.
`default_nettype none

module reset_sync (
    input  logic clk,
    input  logic arst_n,
    output logic rst_n
);
    logic [1:0] pipe;

    always_ff @(posedge clk or negedge arst_n) begin
        if (!arst_n)
            pipe <= 2'b00;
        else
            pipe <= {pipe[0], 1'b1};
    end

    assign rst_n = pipe[1];
endmodule

`default_nettype wire
