from pathlib import Path
p=Path('elastic_dataflow/rtl/dataflow_core.sv');s=p.read_text().replace('(action_token[57]&&(action_token[63:58]==4||action_token[63:58]==6))','(action_token[57]&&required_ports(action_token[63:58])==1)');p.write_text(s)
for name in ['tb_dataflow.sv','tb_dataflow_stress.sv']:
 p=Path('elastic_dataflow/sim')/name;s=p.read_text();s=s.replace('dut(.*);',"dut(.graph_write(1'b0),.graph_commit(1'b0),.graph_index(3'b0),.graph_descriptor(24'b0),.graph_size(4'b0),.graph_writable(),.graph_acceptable(),.*);");p.write_text(s)
p=Path('elastic_dataflow/rtl/dataflow_link.sv');s=p.read_text()
s=s.replace('reg reset_hold,in_valid,out_ready,cancel_valid;', '''reg reset_hold,in_valid,out_ready,cancel_valid;
 reg graph_write,graph_commit;
 reg [2:0] graph_index;
 reg [23:0] graph_descriptor;
 reg [3:0] graph_size;
 wire graph_writable,graph_acceptable;''')
s=s.replace('.debug_addr(debug_addr)', '.graph_write(graph_write),.graph_commit(graph_commit),.graph_index(graph_index),.graph_descriptor(graph_descriptor),.graph_size(graph_size),.graph_writable(graph_writable),.graph_acceptable(graph_acceptable),.debug_addr(debug_addr)')
s=s.replace('command<=0;digits<=0;bad<=0;request<=0;idle_count<=0;query_wait<=0;', 'graph_write<=0;graph_commit<=0;graph_index<=0;graph_descriptor<=0;graph_size<=0;\n     command<=0;digits<=0;bad<=0;request<=0;idle_count<=0;query_wait<=0;',1)
s=s.replace('reset_hold<=0;in_valid<=0;out_ready<=0;cancel_valid<=0;\n', 'reset_hold<=0;in_valid<=0;out_ready<=0;cancel_valid<=0;graph_write<=0;graph_commit<=0;\n')
s=s.replace('!cancel_valid&&!reset_hold)', '!cancel_valid&&!reset_hold&&!graph_write&&!graph_commit)')
s=s.replace('(command=="T"&&digits==10)', '((command=="T"||command=="W")&&digits==10)')
s=s.replace('(command=="Q"||command=="C")', '(command=="Q"||command=="C"||command=="G")')
s=s.replace('"Q":begin debug_addr', '''"W":if(request[39:32]<7&&graph_writable)begin graph_index<=request[34:32];graph_descriptor<=request[31:8];graph_write<=1;short_reply({"A",8'h0a,16'b0},2);end else short_reply({"!02",8'h0a},4);
               "G":if(request[15:8]>=1&&request[15:8]<=7)begin graph_size<=request[11:8];command<=8'hfe;end else short_reply({"!02",8'h0a},4);
               "Q":begin debug_addr''')
s=s.replace('if(!(command=="C"&&digits==4&&!bad&&request_xor==0&&request[15:8]<4))command<=0;', 'if(!((command=="C"&&digits==4&&!bad&&request_xor==0&&request[15:8]<4)||(command=="G"&&digits==4&&!bad&&request_xor==0&&request[15:8]>=1&&request[15:8]<=7)))command<=0;')
s=s.replace('end else if(command==8\'hff)begin', '''end else if(command==8'hfe)begin
         command<=0;
         if(graph_acceptable)begin graph_commit<=1;short_reply({"A",8'h0a,16'b0},2);end
         else short_reply({"!02",8'h0a},4);
       end else if(command==8'hff)begin''')
p.write_text(s)
p=Path('pkg/dataflow/serial.go');s=p.read_text();pos=s.index('\tvar command string')
s=s[:pos]+''' if o.Kind=="load" {
  for n:=byte(0);n<o.Graph.Count;n++ {
   bytes:=o.Graph.Bytes(n)
   line,err:=s.exchange(ctx,EncodeRequest('W',bytes[:]))
   if err!=nil{return nil,err};if line!="A\\n"{s.synchronized=false;return nil,errors.New("expected descriptor acknowledgement")}
  }
  line,err:=s.exchange(ctx,EncodeRequest('G',[]byte{o.Graph.Count}))
  if err!=nil{return nil,err};if line!="A\\n"{s.synchronized=false;return nil,errors.New("expected graph activation acknowledgement")}
  return nil,nil
 }
'''+s[pos:]
s=s.replace('{144, 147}', '{144, 155}');p.write_text(s)
p=Path('pkg/dataflow/protocol.go');s=p.read_text().replace('cap[0] != 1','cap[0] != 2');needle='s.Config = Config{'
pos=s.index(needle);s=s[:pos]+'''graphStatus, err := page(155)
 if err!=nil{return Snapshot{},err}
 if graphStatus[0]<1||graphStatus[0]>Nodes{return Snapshot{},errors.New("invalid active graph count")}
 s.Graph.Count=graphStatus[0]
 for n:=byte(0);n<Nodes;n++ {
  raw,e:=page(148+n);if e!=nil{return Snapshot{},e}
  if raw[6]!=n||raw[7]&0x84!=0||raw[7]>>4>5||raw[7]&3>2||raw[8]>13||raw[9]>13{return Snapshot{},errors.New("invalid descriptor page")}
  if n>=s.Graph.Count {if raw[7]!=0||raw[8]!=0||raw[9]!=0{return Snapshot{},errors.New("nonzero inactive descriptor")};continue}
  op:=Opcode(raw[7]>>4)
  s.Graph.Descriptors[n]=Descriptor{Op:op,Required:required(op),Count:raw[7]&3,Final:raw[7]&8!=0,Destinations:[2]Destination{{raw[8]>>1,raw[8]&1},{raw[9]>>1,raw[9]&1}}}
 }
 // The reset program is initialized directly and has a non-topological COPY.
 if s.Graph!=ResetGraph(){if err:=s.Graph.Validate();err!=nil{return Snapshot{},errors.Wrap(err,"device graph")}}
 '''+s[pos:];p.write_text(s)
p=Path('pkg/dataflow/engine_test.go');s=p.read_text().replace('{144, 147}', '{144, 155}').replace('[10]byte{1, 4, 7','[10]byte{2, 4, 7')
s=s.replace('pages[1] =', '''pages[155]=[10]byte{7}
 for n:=byte(0);n<Nodes;n++{v:=ResetGraph().Bytes(n);var page [10]byte;copy(page[6:],v[:]);pages[148+n]=page}
 pages[1] =''',1);p.write_text(s)
# Existing UART test expects the capability version in its literal.
p=Path('elastic_dataflow/sim/tb_dataflow_link.sv');s=p.read_text().replace('0104070408080808','0204070408080808');p.write_text(s)
