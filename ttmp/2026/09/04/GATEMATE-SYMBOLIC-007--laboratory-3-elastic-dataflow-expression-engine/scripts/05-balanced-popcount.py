#!/usr/bin/env python3
"""Replace serial 28-addend debug population count with a bounded sum tree."""
from pathlib import Path
p=Path('elastic_dataflow/rtl/dataflow_core.sv');s=p.read_text()
s=s.replace(' reg [7:0] ready_count;', ''' wire [1:0] ready_pairs[0:13];
 wire [2:0] ready_quads[0:6];
 wire [3:0] ready_octets[0:3];
 wire [4:0] ready_halves[0:1];
 for(genvar g=0;g<14;g=g+1)begin:count_pairs
   assign ready_pairs[g]={1'b0,pending[g*2]}+{1'b0,pending[g*2+1]};
 end
 for(genvar g=0;g<7;g=g+1)begin:count_quads
   assign ready_quads[g]={1'b0,ready_pairs[g*2]}+{1'b0,ready_pairs[g*2+1]};
 end
 for(genvar g=0;g<3;g=g+1)begin:count_octets
   assign ready_octets[g]={1'b0,ready_quads[g*2]}+{1'b0,ready_quads[g*2+1]};
 end
 assign ready_octets[3]={1'b0,ready_quads[6]};
 assign ready_halves[0]={1'b0,ready_octets[0]}+{1'b0,ready_octets[1]};
 assign ready_halves[1]={1'b0,ready_octets[2]}+{1'b0,ready_octets[3]};
 wire [7:0] ready_count={3'b0,ready_halves[0]}+{3'b0,ready_halves[1]};''')
s=s.replace('selected_slot=0;ready_count=0;eligible=0;','selected_slot=0;eligible=0;').replace('     if(pending[scan])ready_count=ready_count+1\'b1;\n','')
p.write_text(s)
