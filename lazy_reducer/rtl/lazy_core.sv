`default_nettype none
module lazy_core #(parameter integer HEAP_DEPTH=1024,STACK_DEPTH=512,IND_LIMIT=1024)(
 input wire clk,rst_n,enable,
 input wire load_valid,input wire [15:0] load_address,input wire [39:0] load_word,output wire load_ready,
 input wire force_valid,input wire [15:0] force_root,output wire force_ready,
 output wire result_valid,output reg [39:0] result,input wire result_ready,
 input wire debug_select,input wire [15:0] debug_address,output reg [79:0] debug_data,
 output wire idle
);
 localparam IDLE=0,FETCH=1,FWAIT=2,EVAL=3,RET=4,RWAIT=5,RDISPATCH=6,UWAIT=7,UWRITE=8,MULTIPLY=9,OUTPUT=10;
 localparam INT=0,ADD=1,MUL=2,THUNK=3,IND=4,BLACKHOLE=5,ERROR=13;
 localparam ADDRESS_FAULT=1,STACK_OVERFLOW=2,CYCLIC_THUNK=3,IND_CYCLE=4,TYPE_FAULT=5,ARITH_OVERFLOW=6,OWNERSHIP_FAULT=7;
 reg [3:0] state;
 reg [15:0] current,heap_size,sp,hops,heap_address,stack_address,update_address;
 reg [31:0] counters[0:11];
 reg [63:0] mul_acc,mul_shift;
 reg [31:0] mul_bits;
 reg mul_negative;
 reg [5:0] mul_count;
 wire [63:0] mul_next=mul_acc+(mul_bits[0]?mul_shift:64'b0);
 wire [63:0] mul_signed=mul_negative?(~mul_next+64'd1):mul_next;
 reg heap_write,stack_write;
 reg [15:0] heap_wa,stack_wa;
 reg [39:0] heap_wd;
 reg [79:0] stack_wd;
 wire [39:0] heap_rd;
 wire [79:0] stack_rd;
 wire [3:0] tag=heap_rd[39:36];
 wire canonical=heap_rd[35:32]==0&&((tag==INT||tag==ADD||tag==MUL||tag==ERROR)||((tag==THUNK||tag==IND)&&heap_rd[15:0]==0)||(tag==BLACKHOLE&&heap_rd[31:0]==0));
 wire [32:0] add_result={stack_rd[31],stack_rd[31:0]}+{result[31],result[31:0]};
 wire [15:0] heap_ra=debug_select&&debug_address[15:12]==1?{4'b0,debug_address[11:0]}:heap_address;
 wire [15:0] stack_ra=debug_select&&debug_address[15:12]==2?{4'b0,debug_address[11:0]}:stack_address;
 assign idle=state==IDLE;
 assign force_ready=idle;
 assign load_ready=idle&&load_address<HEAP_DEPTH&&load_address<=heap_size;
 assign result_valid=state==OUTPUT;
 sync_sdp_ram #(.DEPTH(HEAP_DEPTH),.WIDTH(40)) heap(.clk(clk),.wr_en(heap_write),.wr_addr(heap_wa[$clog2(HEAP_DEPTH)-1:0]),.wr_data(heap_wd),.rd_addr(heap_ra[$clog2(HEAP_DEPTH)-1:0]),.rd_data(heap_rd));
 sync_sdp_ram #(.DEPTH(STACK_DEPTH),.WIDTH(80)) stack(.clk(clk),.wr_en(stack_write),.wr_addr(stack_wa[$clog2(STACK_DEPTH)-1:0]),.wr_data(stack_wd),.rd_addr(stack_ra[$clog2(STACK_DEPTH)-1:0]),.rd_data(stack_rd));
 // This mux is the only heap mutation owner, including idle construction.
 always @* begin
  heap_write=0;heap_wa=0;heap_wd=0;stack_write=0;stack_wa=0;stack_wd=0;
  if(load_valid&&load_ready)begin heap_write=1;heap_wa=load_address;heap_wd=load_word;end
  else if(enable)case(state)
   EVAL:if(canonical&&current<heap_size&&sp<STACK_DEPTH)begin
    if(tag==ADD||tag==MUL)begin stack_write=1;stack_wa=sp;stack_wd={4'd1,tag,heap_rd[15:0],16'b0,40'b0};end
    if(tag==THUNK)begin stack_write=1;stack_wa=sp;stack_wd={4'd3,4'b0,current,16'b0,40'b0};heap_write=1;heap_wa=current;heap_wd={4'd5,36'b0};end
   end
   RDISPATCH:if(stack_rd[79:76]==1&&result[39:36]!=ERROR)begin stack_write=1;stack_wa=sp-1'b1;stack_wd={4'd2,stack_rd[75:72],32'b0,result};end
   UWRITE:if(update_address<heap_size&&heap_rd=={4'd5,36'b0})begin heap_write=1;heap_wa=update_address;heap_wd=result;end
   default:begin end
  endcase
 end
 task automatic fault(input [31:0] code);
 begin result<={4'd13,4'b0,code};state<=RET;hops<=0;counters[11]<=counters[11]+1'b1;end
 endtask
 integer n;
 always @(posedge clk or negedge rst_n)begin
  if(!rst_n)begin
   state<=IDLE;current<=0;heap_size<=0;sp<=0;hops<=0;heap_address<=0;stack_address<=0;update_address<=0;result<=0;
   mul_acc<=0;mul_shift<=0;mul_bits<=0;mul_negative<=0;mul_count<=0;
   for(n=0;n<12;n=n+1)counters[n]<=0;
  end else begin
   if(load_valid&&load_ready&&load_address==heap_size)heap_size<=heap_size+1'b1;
   if(force_valid&&force_ready)begin current<=force_root;state<=FETCH;hops<=0;result<=0;end
   if(result_valid&&result_ready)state<=IDLE;
   if(enable)begin
    counters[0]<=counters[0]+1'b1;
    case(state)
     FETCH:if(current>=heap_size)fault(ADDRESS_FAULT);else begin heap_address<=current;state<=FWAIT;end
     FWAIT:state<=EVAL;
     EVAL:begin
      counters[1]<=counters[1]+1'b1;
      if(tag!=IND)hops<=0;
      if(!canonical)fault(TYPE_FAULT);
      else case(tag)
       INT,ERROR:begin result<=heap_rd;state<=RET;end
       IND:if(hops==IND_LIMIT)fault(IND_CYCLE);else begin hops<=hops+1'b1;counters[7]<=counters[7]+1'b1;current<=heap_rd[31:16];state<=FETCH;end
       BLACKHOLE:begin counters[8]<=counters[8]+1'b1;fault(CYCLIC_THUNK);end
       ADD,MUL,THUNK:if(sp==STACK_DEPTH)fault(STACK_OVERFLOW);else begin
        sp<=sp+1'b1;if(sp+1>counters[9])counters[9]<=sp+1'b1;current<=heap_rd[31:16];state<=FETCH;
        if(tag==THUNK)begin counters[3]<=counters[3]+1'b1;counters[2]<=counters[2]+1'b1;end
       end
       default:fault(TYPE_FAULT);
      endcase
     end
     RET:if(sp==0)state<=OUTPUT;else begin stack_address<=sp-1'b1;state<=RWAIT;end
     RWAIT:state<=RDISPATCH;
     RDISPATCH:case(stack_rd[79:76])
      1:if(result[39:36]==ERROR)begin sp<=sp-1'b1;state<=RET;end else begin current<=stack_rd[71:56];hops<=0;state<=FETCH;end
      2:begin
       sp<=sp-1'b1;state<=RET;
       if(result[39:36]!=ERROR)begin
        if(stack_rd[75:72]==MUL)counters[5]<=counters[5]+1'b1;else counters[6]<=counters[6]+1'b1;
        if(result[39:32]!=0||stack_rd[39:32]!=0)fault(TYPE_FAULT);
        else if(stack_rd[75:72]==ADD)begin if(add_result[32]!=add_result[31])fault(ARITH_OVERFLOW);else result<={8'b0,add_result[31:0]};end
        else if(stack_rd[75:72]==MUL)begin
         mul_acc<=0;mul_shift<={32'b0,stack_rd[31]?(~stack_rd[31:0]+32'd1):stack_rd[31:0]};
         mul_bits<=result[31]?(~result[31:0]+32'd1):result[31:0];mul_negative<=stack_rd[31]^result[31];mul_count<=0;state<=MULTIPLY;
        end else fault(TYPE_FAULT);
       end
      end
      3:begin update_address<=stack_rd[71:56];heap_address<=stack_rd[71:56];state<=UWAIT;end
      default:begin sp<=sp-1'b1;fault(OWNERSHIP_FAULT);end
     endcase
     UWAIT:state<=UWRITE;
     UWRITE:begin sp<=sp-1'b1;state<=RET;if(update_address>=heap_size||heap_rd!={4'd5,36'b0})fault(OWNERSHIP_FAULT);else begin counters[2]<=counters[2]+1'b1;counters[4]<=counters[4]+1'b1;end end
     MULTIPLY:begin mul_acc<=mul_next;mul_shift<=mul_shift<<1;mul_bits<=mul_bits>>1;mul_count<=mul_count+1'b1;if(mul_count==31)begin state<=RET;if(mul_signed[63:32]!={32{mul_signed[31]}})fault(ARITH_OVERFLOW);else result<={8'b0,mul_signed[31:0]};end end
     OUTPUT:counters[10]<=counters[10]+1'b1;
     default:begin end
    endcase
   end
  end
 end
 reg [6:0] trace_count;
 reg [31:0] trace_dropped;
 wire mutation=heap_write&&enable&&!idle;
 wire [159:0] trace_rd;
 wire [159:0] trace_wd={counters[0]+32'd1,heap_wa,32'b0,heap_rd,heap_wd};
 sync_sdp_ram #(.DEPTH(64),.WIDTH(160)) trace_ram(.clk(clk),.wr_en(mutation&&trace_count<64),.wr_addr(trace_count[5:0]),.wr_data(trace_wd),.rd_addr(debug_address[6:1]),.rd_data(trace_rd));
 always @(posedge clk or negedge rst_n)begin
  if(!rst_n)begin trace_count<=0;trace_dropped<=0;end
  else if(mutation)begin if(trace_count<64)trace_count<=trace_count+1'b1;else trace_dropped<=trace_dropped+1'b1;end
 end
 always @* begin
  debug_data=0;
  case(debug_address)
   0:debug_data={32'h4c415a59,8'd1,16'(HEAP_DEPTH),16'(STACK_DEPTH),8'd0};
   1:debug_data={4'b0,state,7'b0,result_valid,current,heap_size,sp,hops};
   2:debug_data={40'b0,result};
   3:debug_data={9'b0,trace_count,trace_dropped,32'b0};
   default:begin
    if(debug_address>=16&&debug_address<28)debug_data={48'b0,counters[debug_address-16]};
    if(debug_address[15:12]==1&&{4'b0,debug_address[11:0]}<heap_size)debug_data={40'b0,heap_rd};
    if(debug_address[15:12]==2&&{4'b0,debug_address[11:0]}<sp)debug_data=stack_rd;
    if(debug_address>=16'h3000&&debug_address<16'h3080&&{9'b0,debug_address[6:1]}<trace_count)debug_data=debug_address[0]?trace_rd[79:0]:trace_rd[159:80];
   end
  endcase
 end
endmodule
`default_nettype wire
