#!/usr/bin/env python3
"""Relate implementation and reproduction files to the chronological diary."""
from pathlib import Path
import subprocess

ticket = Path(__file__).resolve().parents[1]
repo = next(p for p in ticket.parents if (p/'queens_rollback').is_dir())
files = sorted(p for p in (repo/'queens_rollback').rglob('*')
               if p.is_file() and 'build' not in p.parts and '__pycache__' not in p.parts
               and '.pytest_cache' not in p.parts and p.suffix in ('.py','.sv','.md'))
files += [repo/'queens_rollback/Makefile', *sorted((ticket/'scripts').glob('*'))]
cmd = ['docmgr','doc','relate','--doc',str(ticket/'reference/01-diary.md')]
for path in files:
    cmd += ['--file-note',f'{path}:Lab 2 implementation or reproducible verification and diary tooling']
subprocess.run(cmd,cwd=repo,check=True)
