#!/usr/bin/env python3
from pathlib import Path
root=Path(__file__).resolve().parents[1]
p=root/'reference/02-implemented-reducer-api-and-qualification-handoff.md';s=p.read_text();p.write_text(s.split('---',2)[0]+'---'+s.split('---',2)[1]+'---\n\n'+(root/'sources/handoff-body.md').read_text())
p=root/'scripts/16-browser-fpga.js';p.write_text((root/'scripts/11-browser-model.js').read_text().replace('127.0.0.1:18089','127.0.0.1:18090'))
