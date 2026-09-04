`default_nettype none
package queens_types_pkg;
  localparam [3:0] E_CREATE=1, E_UPDATE=2, E_WRITE=3, E_PROPAGATED=4,
    E_CONTRADICTION=5, E_RESTORE=6, E_RESTORED=7, E_POP=8,
    E_OUTPUT=9, E_COMPLETE=10, E_FAULT=11;
  localparam [3:0] F_NONE=0, F_TRAIL_FULL=1, F_CHOICE_FULL=2,
    F_BAD_DOMAIN_INDEX=3, F_BAD_ONEHOT=4, F_TRAIL_INTEGRITY=5;

  function automatic logic singleton(input logic [7:0] mask);
    singleton = (mask != 0) && ((mask & (mask-8'd1)) == 0);
  endfunction
  function automatic logic [7:0] first_bit(input logic [7:0] mask);
    first_bit = mask & (~mask+8'd1);
  endfunction
  function automatic logic [2:0] row_of(input logic [7:0] mask);
    integer i;
    begin
      row_of=0;
      for(i=0;i<8;i=i+1) if(mask[i]) row_of=3'(i);
    end
  endfunction
  function automatic logic [7:0] attack(input logic [2:0] source_col,
                                        input logic [2:0] source_row,
                                        input logic [2:0] target_col);
    integer r, delta, row;
    begin
      delta = int'(target_col)-int'(source_col);
      row = int'(source_row);
      attack=0;
      for(r=0;r<8;r=r+1)
        if(r==row || r==row+delta || r==row-delta) attack[r]=1'b1;
    end
  endfunction
endpackage
`default_nettype wire
