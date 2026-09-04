#!/usr/bin/env python3
"""Collect cycle and storage-traffic measurements from checked RTL runs."""
from pathlib import Path
import json
import sys
import tempfile

ticket=Path(__file__).resolve().parents[1]
repo=next(p for p in ticket.parents if (p/'queens_rollback').is_dir())
sys.path[:0]=[str(repo/'queens_rollback/tools'),str(repo/'queens_rollback/sim')]
from test_rtl import run_core
out=ticket/'reference/validation'
out.mkdir(exist_ok=True)
results={}
with tempfile.TemporaryDirectory(prefix='queens-profile-') as directory:
    for name,trail in [('snapshot',False),('trail',True)]:
        stdout=run_core(Path(directory),trail=trail)
        stats=next(line.split()[1:] for line in stdout.splitlines() if line.startswith('STATS '))
        result=dict(zip(['cycles','domain_writes','history_writes','history_reads','choice_writes','result_stalls'],map(int,stats)))
        frames=[line.split() for line in stdout.splitlines() if line.startswith('E ')]
        result['max_choices']=max(int(row[4]) for row in frames)
        result['max_trail']=max(int(row[5]) for row in frames)
        result['solutions']=sum(line.startswith('RESULT ') for line in stdout.splitlines())
        result['protected_history_write_bits']=result['history_writes']*(20 if trail else 72)
        result['protected_history_read_bits']=result['history_reads']*(20 if trail else 72)
        result['choice_metadata_write_bits']=result['choice_writes']*40
        result['choice_read_requests']=sum(int(row[1])==7 for row in frames)
        result['memory_write_request_bits']=(result['choice_writes']*(40 if trail else 104)
                                             + (result['history_writes']*20 if trail else 0))
        result['memory_read_request_bits']=(result['choice_read_requests']*(40 if trail else 104)
                                            + (result['history_reads']*20 if trail else 0))
        result['note']='Request bits include complete physical record width; exclude incidental unconsumed synchronous reads. Logical protected history and metadata fields overlap in the snapshot record.'
        results[name]=result
(out/'backend-profile.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results,indent=2))
