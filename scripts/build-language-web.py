#!/usr/bin/env python3
"""Build the LFL1 React entry point and copy the explicit embedded asset set."""
from pathlib import Path
import shutil
import subprocess
root = Path(__file__).resolve().parents[1]
subprocess.run(['pnpm', '--dir', str(root / 'web'), 'build:language'], check=True)
target = root / 'internal/lazylanguageide/assets'
target.mkdir(parents=True, exist_ok=True)
for name in ['index.html', 'app.js', 'app.css']:
    source = root / 'web/dist-language' / ('language/index.html' if name == 'index.html' else name)
    if not source.is_file():
        raise RuntimeError(f'Missing language asset: {source}')
    shutil.copyfile(source, target / name)
