#!/usr/bin/env python3
"""Archive only consulted primary API documentation as Defuddle Markdown."""
from pathlib import Path
import subprocess
T=Path(__file__).resolve().parents[1]
for name,url in [('rtk-query','https://redux-toolkit.js.org/rtk-query/overview'),('go-serial','https://pkg.go.dev/go.bug.st/serial'),('go-http','https://pkg.go.dev/net/http')]:
    subprocess.run(['defuddle','parse',url,'--md','-o',str(T/'sources'/f'{name}.md')],check=True,timeout=60)
