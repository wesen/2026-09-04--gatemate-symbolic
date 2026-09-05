#!/usr/bin/env python3
"""Build the separate lazy entry point into the Go static-file contract."""
from pathlib import Path
import shutil,subprocess
root=Path(__file__).resolve().parents[1]
subprocess.run(['pnpm','--dir',str(root/'web'),'build:lazy'],check=True)
target=root/'internal/lazyide/assets';target.mkdir(parents=True,exist_ok=True)
for name in ['index.html','app.js','app.css']:
 source=root/'web/dist-lazy'/('lazy/index.html' if name=='index.html' else name)
 if not source.is_file():raise RuntimeError(f'Missing lazy asset: {source}')
 shutil.copyfile(source,target/name)
