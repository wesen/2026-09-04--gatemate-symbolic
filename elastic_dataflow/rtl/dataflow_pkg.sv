`default_nettype none
package dataflow_pkg;
  localparam [2:0] MUL=0, ADD=1, LT=2, BTOI=3, COPY=4, SUB=5;
  function automatic [2:0] opcode(input [5:0] node);
    case(node)
      0,1:opcode=MUL; 2,5:opcode=ADD; 3:opcode=LT;
      4:opcode=BTOI; 6:opcode=COPY; default:opcode=7;
    endcase
  endfunction
  function automatic [1:0] required_ports(input [5:0] node);
    required_ports=(node==4 || node==6)?2'b01:2'b11;
  endfunction
  function automatic [6:0] destination(input [5:0] node,input second);
    // Six-bit node and one-bit operand port.
    case(node)
      0:destination={6'd2,1'b0}; 1:destination={6'd2,1'b1};
      2:destination={6'd5,1'b0}; 3:destination={6'd4,1'b0};
      4:destination={6'd5,1'b1}; 6:destination={second?6'd1:6'd0,1'b0};
      default:destination=7'h7f;
    endcase
  endfunction
  function automatic canonical(input [39:0] v);
    canonical=v[35:32]==0 && (v[39:36]==0 || (v[39:36]==1 && v[31:0]<=1));
  endfunction
  function automatic [39:0] error_value(input [7:0] code);
    error_value={4'hd,4'b0,24'b0,code};
  endfunction
  function automatic [39:0] evaluate(input [2:0] op,input [39:0] a,b);
    reg signed [32:0] sum;
    reg signed [31:0] product;
    begin
      sum=0; product=$signed(a[15:0])*$signed(b[15:0]);
      if(!canonical(a) || (op!=BTOI && op!=COPY && !canonical(b))) evaluate=error_value(3);
      else if(op==COPY) evaluate=a;
      else if(op==BTOI) evaluate=a[39:36]==1?{8'b0,a[31:0]}:error_value(3);
      else if(a[39:36]!=0 || b[39:36]!=0) evaluate=error_value(3);
      else case(op)
        MUL:begin
          if(a[31:16]!={16{a[15]}} || b[31:16]!={16{b[15]}}) evaluate=error_value(4);
          else evaluate={8'b0,product};
        end
        ADD,SUB:begin
          if(op==ADD) sum=$signed({a[31],a[31:0]})+$signed({b[31],b[31:0]});
          else sum=$signed({a[31],a[31:0]})-$signed({b[31],b[31:0]});
          evaluate=sum[32]!=sum[31]?error_value(4):{8'b0,sum[31:0]};
        end
        LT:evaluate={4'h1,4'b0,31'b0,($signed(a[31:0])<$signed(b[31:0]))};
        default:evaluate=error_value(5);
      endcase
    end
  endfunction
  function automatic [79:0] completion(input [7:0] ctx,epoch,input [5:0] node,input [39:0] value);
    completion={ctx,epoch,node,1'b0,(node==5),2'b0,node,8'b0,value};
  endfunction
  function automatic [79:0] fault(input [7:0] ctx,epoch,input [5:0] node,input [7:0] code);
    fault={ctx,epoch,node,1'b0,1'b1,2'b0,node,code,error_value(code)};
  endfunction
endpackage
`default_nettype wire
