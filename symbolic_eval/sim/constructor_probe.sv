module constructor_probe (
  input wire [31:0] payload,
  input wire b,
  output wire [39:0] truth, falsity, zero, integer_value, boolean_value
);
  assign truth = symbolic_types_pkg::mk_bool(1'b1);
  assign falsity = symbolic_types_pkg::mk_bool(1'b0);
  assign zero = symbolic_types_pkg::mk_int(32'd0);
  assign integer_value = symbolic_types_pkg::mk_int(payload);
  assign boolean_value = symbolic_types_pkg::mk_bool(b);
endmodule
