#!/usr/bin/env python3
from pathlib import Path
repo=Path(__file__).resolve().parents[6]
p=repo/'web/src/lazy/types.ts'
p.write_text(p.read_text().replace('integer(n.value,-2**31,2**31-1)','integer(n.value,-(2**31),2**31-1)'))
