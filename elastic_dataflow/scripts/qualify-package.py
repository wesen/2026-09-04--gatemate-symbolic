#!/usr/bin/env python3
"""Use explicit package names supported by the installed Yosys frontend."""
from pathlib import Path
import re
for name in ['df_unit.sv', 'dataflow_core.sv']:
    p=Path(__file__).resolve().parents[1]/'rtl'/name
    s=p.read_text().replace(' import dataflow_pkg::*;\n','')
    s=re.sub(r'\b(opcode|required_ports|destination|evaluate|completion|fault|MUL)\b',r'dataflow_pkg::\1',s)
    p.write_text(s)
