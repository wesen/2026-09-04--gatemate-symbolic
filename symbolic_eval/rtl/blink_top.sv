// blink_top.sv — first hardware: prove clock pin, LED pin, toolchain, config
// reset, and the evidence workflow (textbook §1.6.3).
//
// LED_BIT selects which counter bit drives the LED. On hardware the default
// 23 gives a ~0.6 Hz visible blink at 10 MHz; in simulation the testbench
// overrides it to a low bit so a toggle is observable in microseconds.
`default_nettype none

module blink_top #(
    parameter int LED_BIT = 23
) (
    input  logic clk_10m,
    output logic user_led
);
    logic cfg_rst_n;
    logic rst_n;
    logic [23:0] counter;

    CC_USR_RSTN u_cfg_reset (
        .USR_RSTN(cfg_rst_n)
    );

    reset_sync u_reset_sync (
        .clk    (clk_10m),
        .arst_n (cfg_rst_n),
        .rst_n  (rst_n)
    );

    always_ff @(posedge clk_10m) begin
        if (!rst_n)
            counter <= '0;
        else
            counter <= counter + 24'd1;
    end

    assign user_led = counter[LED_BIT];

endmodule

`default_nettype wire
