// uart_tx.sv — 8-N-1 UART transmitter (textbook §3.11). 10-bit shift register
// with a baud divider. Accepts a byte via a one-cycle `start` pulse when
// `ready`. The io_block derives `start` from a single acceptance edge, so a
// held io_req never re-triggers a byte.
`default_nettype none

module uart_tx #(
    parameter int CLK_HZ = 10_000_000,
    parameter int BAUD   = 115_200
) (
    input  logic        clk,
    input  logic        rst_n,
    input  logic        start,    // one-cycle pulse
    input  logic [7:0]  data,
    output logic        ready,
    output logic        tx
);
    localparam int DIVISOR = (CLK_HZ + BAUD/2) / BAUD;
    localparam int BAUD_W  = (DIVISOR <= 1) ? 1 : $clog2(DIVISOR);

    logic [9:0]             frame;
    logic [3:0]             bits_left;
    logic [BAUD_W-1:0]      baud_count;
    logic                   busy;

    assign ready = !busy;
    assign tx    = busy ? frame[0] : 1'b1;   // idle high

    always_ff @(posedge clk) begin
        if (!rst_n) begin
            frame      <= 10'h3FF;  // idle: all ones (stop+idle high)
            bits_left  <= 4'd0;
            baud_count <= '0;
            busy       <= 1'b0;
        end else if (!busy) begin
            if (start) begin
                frame      <= {1'b1, data, 1'b0};  // stop, d7..d0, start
                bits_left  <= 4'd10;
                baud_count <= DIVISOR - 1;
                busy       <= 1'b1;
            end
        end else begin
            if (baud_count == 0) begin
                baud_count <= DIVISOR - 1;
                frame      <= {1'b1, frame[9:1]};  // shift right, fill with idle 1
                bits_left  <= bits_left - 1'b1;
                if (bits_left == 4'd1) begin
                    busy <= 1'b0;   // frame complete after this bit
                end
            end else begin
                baud_count <= baud_count - 1'b1;
            end
        end
    end

endmodule

`default_nettype wire
