#!/usr/bin/env python3
"""Build and copy the fixed three-file frontend contract into Go embed inputs."""
from pathlib import Path
import shutil
import subprocess

root=Path(__file__).resolve().parents[1]
subprocess.run(['pnpm','--dir',str(root/'web'),'build'],check=True)
target=root/'internal/microscope/assets'
target.mkdir(parents=True,exist_ok=True)
for name in ['index.html','app.js','app.css']:
    source=root/'web/dist'/name
    assert source.is_file(),f'Missing frontend build artifact: {source}'
    shutil.copyfile(source,target/name)
