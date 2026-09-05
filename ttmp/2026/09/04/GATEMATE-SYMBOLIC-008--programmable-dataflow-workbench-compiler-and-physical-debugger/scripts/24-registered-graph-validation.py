from pathlib import Path
p=Path('elastic_dataflow/rtl/dataflow_core.sv');s=p.read_text().replace('output reg graph_acceptable','output wire graph_acceptable')
a=s.index(' integer gn,gi,gfinals');b=s.index(' integer init_node;',a)
s=s[:a]+''' // Independent predicates feed reductions, not procedural priority chains.
 wire [6:0] node_bad,final_nodes;
 wire [13:0] edge_bad,edge_active;
 wire [7:0] edge_destination[0:13];
 wire [195:0] duplicate_edges;
 genvar vn,ve,va,vb;
 generate for(vn=0;vn<7;vn=vn+1)begin:validate_nodes
   assign final_nodes[vn]=vn<graph_size&&staged[vn][19];
   assign node_bad[vn]=vn<graph_size&&(!staged_valid[vn]||staged[vn][23]||staged[vn][18]||staged[vn][22:20]>5||staged[vn][17:16]>2||
     (staged[vn][19]?(staged[vn][17:16]!=0):(staged[vn][17:16]==0)));
   for(ve=0;ve<2;ve=ve+1)begin:validate_edges
     wire [7:0] target=ve==0?staged[vn][15:8]:staged[vn][7:0];
     wire [2:0] target_op=target[7:1]<7?staged[target[3:1]][22:20]:3'd7;
     assign edge_destination[vn*2+ve]=target;
     assign edge_active[vn*2+ve]=vn<graph_size&&ve<staged[vn][17:16];
     assign edge_bad[vn*2+ve]=vn<graph_size&&
       (ve<staged[vn][17:16]?(target[7:1]<=vn||target[7:1]>=graph_size||target[7:1]>=7||(target[0]&&(target_op==3||target_op==4))):(target!=0));
   end
 end
 for(va=0;va<14;va=va+1)begin:duplicate_a
   for(vb=0;vb<14;vb=vb+1)begin:duplicate_b
     if(va<vb)assign duplicate_edges[va*14+vb]=edge_active[va]&&edge_active[vb]&&edge_destination[va]==edge_destination[vb];
     else assign duplicate_edges[va*14+vb]=1'b0;
   end
 end endgenerate
 wire validation_comb=graph_size>=1&&graph_size<=7&&!(|node_bad)&&!(|edge_bad)&&!(|duplicate_edges)&&
   final_nodes!=0&&(final_nodes&(final_nodes-7'd1))==0;
 reg validation_result;
 reg [3:0] validation_size;
 always @(posedge clk or negedge rst_n)begin
   if(!rst_n)begin validation_result<=0;validation_size<=0;end
   else begin
     validation_size<=graph_size;
     if(graph_write||graph_commit)validation_result<=0;
     else validation_result<=validation_comb;
   end
 end
 assign graph_acceptable=pristine&&validation_result&&validation_size==graph_size;
'''+s[b:]
p.write_text(s)
p=Path('elastic_dataflow/rtl/dataflow_link.sv');s=p.read_text().replace('reg [3:0] graph_size;','reg [3:0] graph_size;reg [1:0] graph_wait;')
s=s.replace('graph_size<=0;','graph_size<=0;graph_wait<=0;',1).replace('graph_size<=request[11:8];command<=8\'hfe;',"graph_size<=request[11:8];graph_wait<=3;command<=8'hfe;")
s=s.replace('end else if(command==8\'hfe)begin\n         command<=0;',"end else if(command==8'hfe)begin\n         if(graph_wait!=0)graph_wait<=graph_wait-1'b1;\n         else begin\n         command<=0;")
s=s.replace('else short_reply({"!02",8\'h0a},4);\n       end else if(command==8\'hff)', 'else short_reply({"!02",8\'h0a},4);\n         end\n       end else if(command==8\'hff)')
p.write_text(s)
p=Path('elastic_dataflow/sim/tb_programmable.sv');s=p.read_text().replace('graph_size=count;#1;', 'graph_size=count;repeat(3)@(negedge clk);#1;');p.write_text(s)
