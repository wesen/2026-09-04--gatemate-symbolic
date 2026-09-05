from pathlib import Path
p=Path('elastic_dataflow/rtl/dataflow_core.sv');s=p.read_text()
s=s.replace('integer gn,gi,gfinals;','integer gn,gi,gfinals,previous_node,previous_port;')
s=s.replace(' reg [13:0] graph_writers;',' reg [7:0] previous_destination;')
s=s.replace('graph_writers=0;gfinals=0;gd=0;target_op=0;','previous_destination=0;gfinals=0;gd=0;target_op=0;')
s=s.replace('if(graph_writers[gd[3:0]])graph_acceptable=0;\n           graph_writers[gd[3:0]]=1;', '''// Compare fixed descriptor pairs in parallel. A sequential dynamic
           // writer bitmap synthesized into a long shift/mux dependency chain.
           for(previous_node=0;previous_node<7;previous_node=previous_node+1)
             for(previous_port=0;previous_port<2;previous_port=previous_port+1)
               if((previous_node<gn || (previous_node==gn&&previous_port<gi)) && previous_port<staged[previous_node][17:16])begin
                 previous_destination=previous_port==0?staged[previous_node][15:8]:staged[previous_node][7:0];
                 if(gd==previous_destination)graph_acceptable=0;
               end''')
p.write_text(s)
