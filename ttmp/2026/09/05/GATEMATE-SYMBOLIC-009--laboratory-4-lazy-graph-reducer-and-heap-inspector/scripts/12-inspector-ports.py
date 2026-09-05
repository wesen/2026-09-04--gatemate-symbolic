#!/usr/bin/env python3
from pathlib import Path
root=Path(__file__).resolve().parents[1];repo=root.parents[4]
p=root/'scripts/11-browser-model.js';p.write_text(p.read_text().replace('127.0.0.1:8090','127.0.0.1:18089'))
for name in ['cmd/lazy-ide/main.go','web/vite.lazy.config.ts']:
 p=repo/name;p.write_text(p.read_text().replace('127.0.0.1:8089','127.0.0.1:18090'))
