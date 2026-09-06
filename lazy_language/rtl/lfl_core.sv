`default_nettype none
module lfl_core(
 input wire clk,rst_n,enable,
 input wire load_valid,input wire [2:0] load_kind,input wire [15:0] load_address,input wire [127:0] load_data,
 output wire load_ready,
 input wire force_valid,input wire [15:0] force_ref,output wire force_ready,
 output reg result_valid,output reg [15:0] result_ref,input wire result_ready,
 input wire debug_select,input wire [15:0] debug_address,output reg [127:0] debug_data,
 output wire idle
);
 localparam [7:0] INT=0,BOOL=1,NIL=2,CONS=3,FUN=4,THUNK=5,IND=6,BLACKHOLE=7,ENV=8,ERROR=13,FREE=15;
 localparam [7:0] CONST=0,VAR=1,LAMBDA=2,APP=3,LET=4,LETREC=5,PRIM=6,IF=7,MKCONS=8,CASE=9;
 localparam [7:0] ARG=1,PRIGHT=2,PAPPLY=3,BRANCH=4,MATCH=5,UPDATE=6;
 localparam [7:0] IDLE=0,CODE_ISSUE=1,CODE_WAIT=2,EVAL=3,HEAP_ISSUE=4,HEAP_WAIT=5,ENTER=6,ENV_ISSUE=7,ENV_WAIT=8,ENV_DISPATCH=9,RETURN=10,STACK_WAIT=11,RETURN_DISPATCH=12,ALLOCATE=13,UPDATE_ISSUE=14,UPDATE_WAIT=15,UPDATE_WRITE=16,MULTIPLY=17,OUTPUT=18,LOADING=19;
 reg [7:0] state;
 reg loading,committed;
 reg [15:0] code_count,initial_count,root,constant_end,code_loaded,heap_loaded,prov_loaded;
 reg [15:0] current_ref,code_id,env_ref,top,sp,ind_steps,env_steps;
 reg [31:0] depth;
 reg [127:0] instruction,frame;
 reg [79:0] object,returned;
 reg primitive_read;
 reg [31:0] counters[0:22];
 reg [15:0] trace_count;
 reg alloc_active;
 reg [15:0] alloc_base,alloc_end,alloc_count,alloc_next,alloc_span;
 reg [7:0] alloc_kind,alloc_stage,resume_state;
 reg [15:0] resume_code,resume_env,resume_ref;
 reg [79:0] alloc0,alloc1,alloc2;
 reg [63:0] mul_acc,mul_a;
 reg [31:0] mul_b;
 reg mul_negative;
 reg [5:0] mul_step;
 wire [63:0] mul_sum=mul_acc+(mul_b[0]?mul_a:64'b0);
 wire [63:0] mul_signed=mul_negative?(~mul_sum+64'd1):mul_sum;
 wire [7:0] op=instruction[127:120],fkind=frame[127:120],fop=frame[119:112];
 wire [15:0] ca=instruction[111:96],cb=instruction[95:80],cc=instruction[79:64],cspan=instruction[31:16];
 wire [31:0] imm=instruction[63:32];
 wire [15:0] fa=frame[111:96],fb=frame[95:80],fc=frame[79:64],fsaved=frame[47:32],fspan=frame[31:16];
 wire [7:0] otag=object[79:72],rtag=returned[79:72];
 wire [15:0] oa=object[63:48],ob=object[47:32],ra=returned[63:48],rb=returned[47:32];
 wire signed [32:0] left_ext={object[31],object[31:0]},right_ext={returned[31],returned[31:0]};
 wire signed [32:0] add_value=left_ext+right_ext,sub_value=left_ext-right_ext;
 wire [79:0] alloc_word=alloc_next==0?alloc0:alloc_next==1?alloc1:alloc2;
 wire running=enable&&!debug_select;
 assign idle=state==IDLE;
 assign force_ready=committed&&state==IDLE&&!result_valid;
 assign load_ready=state==IDLE||state==LOADING;
 function automatic [79:0] obj(input [7:0] tag,input [15:0] a,b,input [31:0] payload);obj={tag,8'b0,a,b,payload};endfunction
 function automatic [127:0] frm(input [7:0] kind,oper,input [15:0] a,b,c,saved,span);frm={kind,oper,a,b,c,16'b0,saved,span,16'b0};endfunction
 function automatic terminal(input [7:0] tag);terminal=tag<=4||tag==13;endfunction
 reg heap_we,code_we,prov_we,stack_we,trace_we;
 reg [10:0] heap_wa,heap_ra,code_wa,code_ra,prov_wa,prov_ra;
 reg [8:0] stack_wa,stack_ra;
 reg [5:0] trace_wa,trace_ra;
 reg [79:0] heap_wd;
 reg [127:0] code_wd,stack_wd;
 reg [15:0] prov_wd;
 reg [255:0] trace_wd;
 wire [79:0] heap_rd;
 wire [127:0] code_rd,stack_rd;
 wire [15:0] prov_rd;
 wire [255:0] trace_rd;
 sync_sdp_ram #(.DEPTH(2048),.WIDTH(80)) heap_ram(.clk(clk),.wr_en(heap_we),.wr_addr(heap_wa),.wr_data(heap_wd),.rd_addr(heap_ra),.rd_data(heap_rd));
 sync_sdp_ram #(.DEPTH(2048),.WIDTH(128)) code_ram(.clk(clk),.wr_en(code_we),.wr_addr(code_wa),.wr_data(code_wd),.rd_addr(code_ra),.rd_data(code_rd));
 sync_sdp_ram #(.DEPTH(2048),.WIDTH(16)) provenance_ram(.clk(clk),.wr_en(prov_we),.wr_addr(prov_wa),.wr_data(prov_wd),.rd_addr(prov_ra),.rd_data(prov_rd));
 sync_sdp_ram #(.DEPTH(512),.WIDTH(128)) stack_ram(.clk(clk),.wr_en(stack_we),.wr_addr(stack_wa),.wr_data(stack_wd),.rd_addr(stack_ra),.rd_data(stack_rd));
 sync_sdp_ram #(.DEPTH(64),.WIDTH(256)) trace_ram(.clk(clk),.wr_en(trace_we),.wr_addr(trace_wa),.wr_data(trace_wd),.rd_addr(trace_ra),.rd_data(trace_rd));
 reg mutation;
 reg [7:0] mutation_kind;
 reg [15:0] mutation_ref,mutation_span;
 reg [79:0] mutation_old,mutation_new;
 always @*begin
  heap_we=0;code_we=0;prov_we=0;stack_we=0;trace_we=0;
  heap_wa=0;code_wa=load_address[10:0];prov_wa=0;stack_wa=sp[8:0];trace_wa=trace_count[5:0];
  heap_wd=0;code_wd=load_data;prov_wd=0;stack_wd=0;trace_wd=0;
  heap_ra=current_ref[10:0];code_ra=code_id[10:0];prov_ra=current_ref[10:0];stack_ra=9'(sp-1'b1);trace_ra=debug_address[6:1];
  if(state==ENV_ISSUE||state==ENV_WAIT||state==ENV_DISPATCH)heap_ra=env_ref[10:0];
  if(state==UPDATE_ISSUE||state==UPDATE_WAIT||state==UPDATE_WRITE)begin heap_ra=fa[10:0];prov_ra=fa[10:0];end
  mutation=0;mutation_kind=0;mutation_ref=0;mutation_span=0;mutation_old=0;mutation_new=0;
  if(load_valid&&load_ready)begin
   if(load_kind==1&&loading&&load_address==code_loaded&&code_loaded<code_count)begin code_we=1;end
   if(load_kind==2&&loading&&load_address==heap_loaded&&heap_loaded<initial_count)begin heap_we=1;heap_wa=load_address[10:0];heap_wd=load_data[79:0];end
   if(load_kind==3&&loading&&load_address==prov_loaded&&prov_loaded<initial_count)begin prov_we=1;prov_wa=load_address[10:0];prov_wd=load_data[15:0];end
  end
  if(running)begin
   if(state==ENTER&&!primitive_read&&otag==THUNK&&sp<512)begin
    heap_we=1;heap_wa=current_ref[10:0];heap_wd=obj(BLACKHOLE,0,0,0);
    stack_we=1;stack_wd=frm(UPDATE,0,current_ref,0,0,0,prov_rd);
    mutation=1;mutation_kind=2;mutation_ref=current_ref;mutation_span=prov_rd;mutation_old=object;mutation_new=heap_wd;
   end
   if(state==EVAL&&sp<512)begin
    case(op)
     APP:begin stack_we=1;stack_wd=frm(ARG,0,cb,env_ref,0,0,cspan);end
     PRIM:begin stack_we=1;stack_wd=frm(PRIGHT,imm[7:0],cb,env_ref,0,0,cspan);end
     IF:begin stack_we=1;stack_wd=frm(BRANCH,0,cb,env_ref,cc,0,cspan);end
     CASE:begin stack_we=1;stack_wd=frm(MATCH,0,cb,env_ref,cc,0,cspan);end
    endcase
   end
   if(state==RETURN_DISPATCH&&sp!=0&&fkind==PRIGHT&&rtag==INT)begin stack_we=1;stack_wa=9'(sp-1'b1);stack_wd=frm(PAPPLY,fop,0,0,0,current_ref,fspan);end
   if(state==UPDATE_WRITE&&otag==BLACKHOLE&&fa>=constant_end&&fa<top)begin
    heap_we=1;heap_wa=fa[10:0];heap_wd=obj(IND,current_ref,0,0);
    mutation=1;mutation_kind=3;mutation_ref=fa;mutation_span=prov_rd;mutation_old=object;mutation_new=heap_wd;
   end
   if(state==ALLOCATE)begin
    if(alloc_stage==0)begin heap_we=1;heap_wa=11'(alloc_base+alloc_next);heap_wd=alloc_word;end
    if(alloc_stage==1)begin prov_we=1;prov_wa=11'(alloc_base+alloc_next);prov_wd=alloc_span;end
    if(alloc_stage==2)begin mutation=1;mutation_kind=1;mutation_ref=alloc_base+alloc_next;mutation_span=alloc_span;mutation_old=obj(FREE,0,0,0);mutation_new=alloc_word;end
   end
   if(mutation&&trace_count<64)begin trace_we=1;trace_wd={counters[0]+32'd1,mutation_ref,mutation_span,mutation_kind,8'b0,16'b0,mutation_old,mutation_new};end
  end
  if(debug_select)begin heap_ra=debug_address[10:0];code_ra=debug_address[10:0];prov_ra=debug_address[10:0];stack_ra=debug_address[8:0];end
 end
 always @*begin
  debug_data=0;
  case(debug_address)
   0:debug_data={32'h4c464c31,8'd1,16'd2048,16'd512,16'd2048,16'd64,24'b0};
   1:debug_data={state,4'b0,alloc_active,committed,loading,result_valid,current_ref,code_id,env_ref,sp,top,result_ref,16'b0};
   2:debug_data={root,code_count,initial_count,constant_end,ind_steps,env_steps,32'b0};
   3:debug_data={trace_count,counters[22],80'b0};
   4:debug_data={alloc_base,alloc_end,alloc_count,alloc_next,alloc_kind,alloc_stage,48'b0};
   default:begin
    if(debug_address>=16&&debug_address<=38)debug_data={96'b0,counters[debug_address-16]};
    else case(debug_address[15:12])
     1:if(debug_address[11:0]<2048)debug_data={48'b0,heap_rd};
     2:if(debug_address[11:0]<code_count)debug_data=code_rd;
     3:if(debug_address[11:0]<sp)debug_data=stack_rd;
     4:if(debug_address[11:0]<2048)debug_data={112'b0,prov_rd};
     5:if(debug_address[11:1]<trace_count)debug_data=debug_address[0]?trace_rd[127:0]:trace_rd[255:128];
    endcase
   end
  endcase
 end
 task automatic fault(input [31:0] code);begin current_ref<=16'(code-1);state<=RETURN;counters[21]<=counters[21]+1'b1;primitive_read<=0;end endtask
 task automatic eval_at(input [15:0] code,env);begin code_id<=code;env_ref<=env;state<=CODE_ISSUE;end endtask
 task automatic reserve(input [7:0] kind,input [15:0] span,count,input [79:0] a,b,c,input [7:0] next_state,input [15:0] next_code,next_env,next_ref);
  begin
   if({1'b0,top}+{1'b0,count}>2048)fault(8);
   else begin alloc_active<=1;alloc_base<=top;alloc_end<=top+count;alloc_count<=count;alloc_next<=0;alloc_span<=span;alloc_kind<=kind;alloc_stage<=0;alloc0<=a;alloc1<=b;alloc2<=c;resume_state<=next_state;resume_code<=next_code;resume_env<=next_env;resume_ref<=next_ref;state<=ALLOCATE;end
  end
 endtask
 task automatic integer_result(input [31:0] value,input [15:0] span);begin reserve(7,span,1,obj(INT,0,0,value),0,0,RETURN,0,env_ref,top);end endtask
 always @(posedge clk or negedge rst_n)begin
  if(!rst_n)begin
   state<=IDLE;loading<=0;committed<=0;result_valid<=0;result_ref<=16'hffff;
   code_count<=0;initial_count<=0;root<=0;constant_end<=0;code_loaded<=0;heap_loaded<=0;prov_loaded<=0;
   current_ref<=16'hffff;code_id<=0;env_ref<=16'hffff;top<=0;sp<=0;ind_steps<=0;env_steps<=0;depth<=0;instruction<=0;frame<=0;object<=0;returned<=0;primitive_read<=0;
   trace_count<=0;alloc_active<=0;alloc_base<=0;alloc_end<=0;alloc_count<=0;alloc_next<=0;alloc_span<=0;alloc_kind<=0;alloc_stage<=0;resume_state<=0;resume_code<=0;resume_env<=0;resume_ref<=0;alloc0<=0;alloc1<=0;alloc2<=0;
   mul_acc<=0;mul_a<=0;mul_b<=0;mul_negative<=0;mul_step<=0;
   for(integer i=0;i<23;i=i+1)counters[i]<=0;
  end else begin
   if(result_ready&&result_valid)begin result_valid<=0;state<=IDLE;end
   if(load_valid&&load_ready)begin
    case(load_kind)
     0:if(load_data[63:48]>0&&load_data[63:48]<=2048&&load_data[47:32]>=13&&load_data[47:32]<=2048&&load_data[31:16]<load_data[47:32]&&load_data[15:0]>=13&&load_data[15:0]<=load_data[47:32])begin
      loading<=1;committed<=0;state<=LOADING;code_count<=load_data[63:48];initial_count<=load_data[47:32];root<=load_data[31:16];constant_end<=load_data[15:0];code_loaded<=0;heap_loaded<=0;prov_loaded<=0;
     end
     1:if(code_we)code_loaded<=code_loaded+1'b1;
     2:if(heap_we)heap_loaded<=heap_loaded+1'b1;
     3:if(prov_we)prov_loaded<=prov_loaded+1'b1;
     4:if(loading&&code_loaded==code_count&&heap_loaded==initial_count&&prov_loaded==initial_count)begin loading<=0;committed<=1;state<=IDLE;top<=initial_count;counters[19]<=initial_count;end
    endcase
   end
   if(force_valid&&force_ready)begin current_ref<=force_ref;state<=HEAP_ISSUE;sp<=0;ind_steps<=0;env_steps<=0;primitive_read<=0;end
   if(running&&state!=IDLE&&state!=LOADING)begin
    counters[0]<=counters[0]+1'b1;
    if(mutation)begin if(trace_count<64)trace_count<=trace_count+1'b1;else counters[22]<=counters[22]+1'b1;end
    case(state)
     CODE_ISSUE:if(code_id>=code_count)fault(9);else state<=CODE_WAIT;
     CODE_WAIT:begin instruction<=code_rd;state<=EVAL;end
     EVAL:begin
      counters[1]<=counters[1]+1'b1;
      case(op)
       CONST:begin current_ref<=ca;state<=RETURN;end
       VAR:begin depth<=imm;env_steps<=0;state<=ENV_ISSUE;end
       LAMBDA:reserve(1,cspan,1,obj(FUN,ca,env_ref,0),0,0,RETURN,0,env_ref,top);
       APP,PRIM,IF,CASE:if(sp>=512)fault(2);else begin sp<=sp+1'b1;if(sp+1>counters[18])counters[18]<=sp+1'b1;eval_at(ca,env_ref);end
       LET,LETREC:reserve(op==LET?3:4,cspan,2,obj(THUNK,ca,op==LET?env_ref:top+16'd1,0),obj(ENV,top,env_ref,0),0,CODE_ISSUE,cb,top+16'd1,16'hffff);
       MKCONS:reserve(5,cspan,3,obj(THUNK,ca,env_ref,0),obj(THUNK,cb,env_ref,0),obj(CONS,top,top+16'd1,0),RETURN,0,env_ref,top+16'd2);
       default:fault(9);
      endcase
     end
     HEAP_ISSUE:if(current_ref>=top)fault(1);else begin counters[2]<=counters[2]+1'b1;state<=HEAP_WAIT;end
     HEAP_WAIT:begin object<=heap_rd;state<=ENTER;end
     ENTER:begin
      if(primitive_read)begin
       primitive_read<=0;
       if(otag!=INT)fault(5);
       else case(fop)
        0:begin counters[11]<=counters[11]+1'b1;if(add_value[32]!=add_value[31])fault(6);else integer_result(add_value[31:0],fspan);end
        1:begin counters[12]<=counters[12]+1'b1;if(sub_value[32]!=sub_value[31])fault(6);else integer_result(sub_value[31:0],fspan);end
        2:begin counters[13]<=counters[13]+1'b1;mul_acc<=0;mul_a<={32'b0,object[31]?(~object[31:0]+32'd1):object[31:0]};mul_b<=returned[31]?(~returned[31:0]+32'd1):returned[31:0];mul_negative<=object[31]^returned[31];mul_step<=0;state<=MULTIPLY;end
        3:begin counters[14]<=counters[14]+1'b1;current_ref<=object[31:0]==returned[31:0]?11:10;state<=RETURN;end
        4:begin counters[15]<=counters[15]+1'b1;current_ref<=left_ext<=right_ext?11:10;state<=RETURN;end
        default:fault(9);
       endcase
      end else case(otag)
       INT,BOOL,NIL,CONS,FUN,ERROR:begin ind_steps<=0;state<=RETURN;end
       IND:if(ind_steps>=2048)fault(4);else begin ind_steps<=ind_steps+1'b1;counters[16]<=counters[16]+1'b1;current_ref<=oa;state<=HEAP_ISSUE;end
       BLACKHOLE:fault(3);
       THUNK:if(sp>=512)fault(2);else begin sp<=sp+1'b1;if(sp+1>counters[18])counters[18]<=sp+1'b1;counters[9]<=counters[9]+1'b1;ind_steps<=0;eval_at(oa,ob);end
       default:fault(5);
      endcase
     end
     ENV_ISSUE:if(env_ref==16'hffff||env_ref>=top||env_steps>=2048)fault(10);else begin counters[2]<=counters[2]+1'b1;state<=ENV_WAIT;end
     ENV_WAIT:begin object<=heap_rd;state<=ENV_DISPATCH;end
     ENV_DISPATCH:if(otag!=ENV)fault(10);else begin env_steps<=env_steps+1'b1;counters[17]<=counters[17]+1'b1;if(depth==0)begin current_ref<=oa;state<=HEAP_ISSUE;end else begin depth<=depth-1'b1;env_ref<=ob;state<=ENV_ISSUE;end end
     RETURN:if(current_ref>=top)fault(7);else begin counters[2]<=counters[2]+1'b1;state<=STACK_WAIT;end
     STACK_WAIT:begin returned<=heap_rd;frame<=stack_rd;state<=RETURN_DISPATCH;end
     RETURN_DISPATCH:begin
      if(!terminal(rtag))fault(7);
      else if(sp==0)begin result_ref<=current_ref;result_valid<=1;state<=OUTPUT;end
      else if(fkind==UPDATE)state<=UPDATE_ISSUE;
      else if(rtag==ERROR)begin sp<=sp-1'b1;state<=RETURN;end
      else case(fkind)
       ARG:begin sp<=sp-1'b1;if(rtag!=FUN)fault(5);else reserve(2,fspan,2,obj(THUNK,fa,fb,0),obj(ENV,top,rb,0),0,CODE_ISSUE,ra,top+16'd1,16'hffff);end
       PRIGHT:if(rtag!=INT)begin sp<=sp-1'b1;fault(5);end else eval_at(fa,fb);
       PAPPLY:begin sp<=sp-1'b1;if(rtag!=INT)fault(5);else begin primitive_read<=1;current_ref<=fsaved;state<=HEAP_ISSUE;end end
       BRANCH:begin sp<=sp-1'b1;if(rtag!=BOOL)fault(5);else eval_at(returned[31:0]!=0?fa:fc,fb);end
       MATCH:begin sp<=sp-1'b1;if(rtag==NIL)eval_at(fa,fb);else if(rtag!=CONS)fault(5);else reserve(6,fspan,2,obj(ENV,ra,fb,0),obj(ENV,rb,top,0),0,CODE_ISSUE,fc,top+16'd1,16'hffff);end
       default:begin sp<=sp-1'b1;fault(7);end
      endcase
     end
     UPDATE_ISSUE:if(fa<constant_end||fa>=top)begin sp<=sp-1'b1;fault(7);end else begin counters[2]<=counters[2]+1'b1;state<=UPDATE_WAIT;end
     UPDATE_WAIT:begin object<=heap_rd;state<=UPDATE_WRITE;end
     UPDATE_WRITE:begin sp<=sp-1'b1;if(otag!=BLACKHOLE)fault(7);else begin counters[10]<=counters[10]+1'b1;state<=RETURN;end end
     ALLOCATE:case(alloc_stage)
      0:if(alloc_next+1==alloc_count)begin alloc_next<=0;alloc_stage<=1;end else alloc_next<=alloc_next+1'b1;
      1:if(alloc_next+1==alloc_count)begin alloc_next<=0;alloc_stage<=2;top<=alloc_end;if(alloc_end>counters[19])counters[19]<=alloc_end;end else alloc_next<=alloc_next+1'b1;
      2:begin
       counters[3]<=counters[3]+1'b1;
       case(alloc_word[79:72])THUNK:counters[4]<=counters[4]+1'b1;FUN:counters[5]<=counters[5]+1'b1;ENV:counters[6]<=counters[6]+1'b1;CONS:counters[7]<=counters[7]+1'b1;INT:counters[8]<=counters[8]+1'b1;endcase
       if(alloc_next+1==alloc_count)alloc_stage<=3;else alloc_next<=alloc_next+1'b1;
      end
      3:begin alloc_active<=0;state<=resume_state;code_id<=resume_code;env_ref<=resume_env;current_ref<=resume_ref;end
     endcase
     MULTIPLY:begin mul_acc<=mul_sum;mul_a<=mul_a<<1;mul_b<=mul_b>>1;mul_step<=mul_step+1'b1;if(mul_step==31)begin if(mul_signed[63:32]!={32{mul_signed[31]}})fault(6);else integer_result(mul_signed[31:0],fspan);end end
     OUTPUT:counters[20]<=counters[20]+1'b1;
    endcase
   end
  end
 end
endmodule
`default_nettype wire
