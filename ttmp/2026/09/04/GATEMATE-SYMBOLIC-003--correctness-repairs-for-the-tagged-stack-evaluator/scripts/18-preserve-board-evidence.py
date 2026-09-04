#!/usr/bin/env python3
"""Preserve the failing board run before rebuilding the constructor repair."""
from pathlib import Path
import shutil

out = Path(__file__).resolve().parents[1] / 'reference/validation'
archive = out / 'before-constructor-fix'
archive.mkdir(exist_ok=False)
for pattern in ['P5-hardware*', 'P5-*-uart.bin', 'P5-*-build.log',
                'P5-*-load.log', 'P5-*-yosys.log', 'P5-*-nextpnr.log',
                'P5-build*', 'P5-image*', 'P5-yosys.log', 'P5-nextpnr.log']:
    for source in out.glob(pattern):
        if source.is_file():
            shutil.copy2(source, archive / source.name)
print(f'Preserved {len(list(archive.iterdir()))} evidence files')
