from pathlib import Path
p=Path('elastic_dataflow/rtl/dataflow_core.sv');s=p.read_text()
s=s.replace('input wire [7:0] debug_addr,','input wire graph_write,graph_commit,input wire [2:0] graph_index,input wire [23:0] graph_descriptor,input wire [3:0] graph_size,\n output wire graph_writable,output reg graph_acceptable,\n input wire [7:0] debug_addr,')
anchor=' reg [7:0] epoch[0:3];'
block=''' reg [23:0] descriptors[0:6], staged[0:6];
 reg [6:0] staged_valid;
 reg [3:0] active_count;
 reg pristine;
 assign graph_writable=pristine;
 function automatic [2:0] opcode(input [5:0] node);
   opcode=node<active_count?descriptors[node][22:20]:3'd7;
 endfunction
 function automatic [1:0] required_ports(input [5:0] node);
   required_ports=(opcode(node)==3 || opcode(node)==4)?2'b01:2'b11;
 endfunction
 function automatic [6:0] destination(input [5:0] node,input second);
   destination=second?descriptors[node][6:0]:descriptors[node][14:8];
 endfunction
 function automatic [79:0] make_completion(input [7:0] ctx,ep,input [5:0] node,input [39:0] value);
   make_completion={ctx,ep,node,1'b0,descriptors[node][19],2'b0,node,8'b0,value};
 endfunction
 integer gn,gi,gfinals;
 reg [13:0] graph_writers;
 reg [7:0] gd;
 reg [2:0] target_op;
 always @* begin
   graph_acceptable=pristine&&graph_size>=1&&graph_size<=7;
   graph_writers=0;gfinals=0;gd=0;target_op=0;
   for(gn=0;gn<7;gn=gn+1)if(gn<graph_size)begin
     if(!staged_valid[gn] || staged[gn][23] || staged[gn][18] || staged[gn][22:20]>5 || staged[gn][17:16]>2)graph_acceptable=0;
     if(staged[gn][19])begin gfinals=gfinals+1;if(staged[gn][17:16]!=0)graph_acceptable=0;end
     else if(staged[gn][17:16]==0)graph_acceptable=0;
     for(gi=0;gi<2;gi=gi+1)begin
       gd=gi==0?staged[gn][15:8]:staged[gn][7:0];
       if(gi<staged[gn][17:16])begin
         if(gd[7:1]<=gn || gd[7:1]>=graph_size || gd[7:1]>=7)graph_acceptable=0;
         else begin
           target_op=staged[gd[3:1]][22:20];
           if(gd[0]&&(target_op==3||target_op==4))graph_acceptable=0;
           if(graph_writers[gd[3:0]])graph_acceptable=0;
           graph_writers[gd[3:0]]=1;
         end
       end else if(gd!=0)graph_acceptable=0;
     end
   end
   if(gfinals!=1)graph_acceptable=0;
 end
 integer init_node;
 always @(posedge clk or negedge rst_n)begin
   if(!rst_n)begin
     pristine<=1;active_count<=7;staged_valid<=0;
     for(init_node=0;init_node<7;init_node=init_node+1)begin
       staged[init_node]<=0;
       descriptors[init_node]<={1'b0,dataflow_pkg::opcode(init_node),init_node==5,1'b0,(init_node==5?2'd0:init_node==6?2'd2:2'd1),
          (init_node==5?8'b0:{1'b0,dataflow_pkg::destination(init_node,1'b0)}),
          (init_node==6?{1'b0,dataflow_pkg::destination(init_node,1'b1)}:8'b0)};
     end
   end else begin
     if(enable || (in_valid&&in_ready) || cancel_fire)pristine<=0;
     if(graph_write&&graph_writable&&graph_index<7)begin staged[graph_index]<=graph_descriptor;staged_valid[graph_index]<=1;end
     if(graph_commit&&graph_acceptable)begin
       active_count<=graph_size;staged_valid<=0;
       for(init_node=0;init_node<7;init_node=init_node+1)descriptors[init_node]<=init_node<graph_size?staged[init_node]:24'b0;
     end
   end
 end
'''
s=s.replace(anchor,block+anchor)
# Only core runtime lookup, preserving qualified initialization helpers above.
pos=s.index(anchor);head=s[:pos];body=s[pos:]
body=body.replace('dataflow_pkg::opcode(', 'opcode(').replace('dataflow_pkg::required_ports(', 'required_ports(').replace('dataflow_pkg::destination(', 'destination(').replace('dataflow_pkg::completion(', 'make_completion(')
body=body.replace('action_token[63:58]>=7','action_token[63:58]>=active_count')
body=body.replace('router_token[63:58]!=6||delivered[0]','descriptors[router_token[63:58]][17:16]==1||delivered[0]')
body=body.replace("{8'd1,8'd4,8'd7", "{8'd2,8'd4,8'd7")
body=body.replace('case(debug_addr)','case(debug_addr)\n     155:debug_data={4\'b0,active_count,1\'b0,staged_valid,7\'b0,pristine,56\'b0};')
# existing default is range decode
body=body.replace('default:begin','default:begin\n       if(debug_addr>=148&&debug_addr<=154)debug_data={48\'b0,5\'b0,(3\'(debug_addr-148)),descriptors[debug_addr-148]};',1)
p.write_text(head+body)
# Wire explicit inactive graph controls in existing testbenches.
for p in Path('elastic_dataflow/sim').glob('*.sv'):
 s=p.read_text()
 if 'dataflow_core' in s:s=s.replace('.debug_addr(', ".graph_write(1'b0),.graph_commit(1'b0),.graph_index(3'b0),.graph_descriptor(24'b0),.graph_size(4'b0),.graph_writable(),.graph_acceptable(),.debug_addr(")
 p.write_text(s)
