// CC_USR_RSTN.sv — simulation-only model of the GateMate configuration reset
// primitive (sibling §1.7). An open simulator does not know GateMate
// primitives, so we model USR_RSTN as: low during/after configuration, then
// high after 250 ns. NOT compiled during synthesis (synth_gatemate provides
// the real primitive). Reused unchanged from MATE-16.
`timescale 1ns/1ps

module CC_USR_RSTN (output logic USR_RSTN);
    initial begin
        USR_RSTN = 1'b0;
        #250;
        USR_RSTN = 1'b1;
    end
endmodule
