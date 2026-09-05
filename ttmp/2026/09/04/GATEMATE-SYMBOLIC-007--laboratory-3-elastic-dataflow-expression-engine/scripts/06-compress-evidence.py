#!/usr/bin/env python3
"""Retain large synthesis evidence losslessly without noisy plain-text diffs."""
from pathlib import Path
import gzip
root=Path(__file__).resolve().parents[1]/'reference/validation'
for name in ['P3-scheduler-cost.log','P3-synthesis-final.log']:
 p=root/name
 if p.exists():
  p.with_suffix(p.suffix+'.gz').write_bytes(gzip.compress(p.read_bytes(),mtime=0))
  p.unlink()
