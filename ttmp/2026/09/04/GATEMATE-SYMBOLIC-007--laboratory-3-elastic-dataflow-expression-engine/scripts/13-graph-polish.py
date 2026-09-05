#!/usr/bin/env python3
from pathlib import Path
p=Path('web/src/dataflow/Graph.tsx');s=p.read_text().replace('0 0 850 520','0 0 850 540').replace('const [u,v]=positions[b];return','const [u,v]=positions[b];const destY=v+(port===\'A\'?34:70);return').replace('${u-40},${v+48} ${u},${v+48}','${u-40},${destY} ${u},${destY}').replace('y={v+37}','y={destY-6}').replace('y="492"','y="516"');p.write_text(s)
p=Path('web/dataflow/index.html');s=p.read_text().replace('<title>','<link rel="icon" href="data:,"/><title>');p.write_text(s)
