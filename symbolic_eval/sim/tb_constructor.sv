module tb_constructor;
  reg [31:0] payload = 32'hffffffff;
  reg b = 1'b1;
  wire [39:0] truth, falsity, zero, integer_value, boolean_value;
  constructor_probe dut(.*);
  initial begin
    #1;
    if (truth !== 40'h1000000001 || falsity !== 40'h1000000000 ||
        zero !== 40'h0000000000 || integer_value !== 40'h00ffffffff ||
        boolean_value !== 40'h1000000001)
      $fatal(1, "constructor tags/flags lost: true=%h false=%h zero=%h int=%h bool=%h",
             truth, falsity, zero, integer_value, boolean_value);
    b = 0;
    payload = 32'h80000000;
    #1;
    if (integer_value !== 40'h0080000000 || boolean_value !== 40'h1000000000)
      $fatal(1, "dynamic constructor mismatch");
    $display("CONSTRUCTORS_PASS");
    $finish;
  end
endmodule
