#!/usr/bin/env python3
from pathlib import Path
p=Path('web/src/dataflow/App.tsx');s=p.read_text().replace('const fileInput=useRef<HTMLInputElement>(null);','const fileInput=useRef<HTMLInputElement>(null);const initialized=useRef(false);')
s=s.replace("if(examples&&!workspace.source)dispatch(actions.setSource(examples.book))","if(examples&&!initialized.current){initialized.current=true;if(!workspace.source)dispatch(actions.setSource(examples.book))}")
s=s.replace("snapshot?.source==='serial'?'PHYSICAL FPGA':'TRANSACTION MODEL'","!snapshot?'CONNECTING':snapshot.source==='serial'?'PHYSICAL FPGA':'TRANSACTION MODEL'")
s=s.replace("snapshot?.source==='serial'?'GateMate · 10 MHz':'Software reference · explicit cycles'","!snapshot?'Awaiting engine snapshot':snapshot.source==='serial'?'GateMate · 10 MHz':'Software reference · explicit cycles'")
p.write_text(s)
p=Path('web/src/dataflow/App.test.tsx');s=p.read_text().replace("expect(screen.getByText('Model configuration')).toBeInTheDocument()","expect(await screen.findByText('Model configuration')).toBeInTheDocument()");p.write_text(s)
