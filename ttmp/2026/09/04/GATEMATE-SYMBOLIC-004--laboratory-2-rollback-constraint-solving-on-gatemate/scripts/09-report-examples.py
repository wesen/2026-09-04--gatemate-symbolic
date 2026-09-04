#!/usr/bin/env python3
"""Check the report's execution examples against the model and archived RTL trace."""
from pathlib import Path
import json
import sys

ticket = Path(__file__).resolve().parents[1]
repo = next(p for p in ticket.parents if (p/'queens_rollback').is_dir())
sys.path.insert(0,str(repo/'queens_rollback/tools'))
from queens_model import Machine, Event
from oracle import solve_rows, unpack

model = Machine(trail=True).run()
assert [unpack(word) for word in model.output] == solve_rows()
assert model.output[0] == 0x672be0
frames = [frame.splitlines()[0] for frame in model.events]
contradiction = next(i for i,line in enumerate(frames) if line.startswith('E 5 '))
restored = next(i for i in range(contradiction,len(frames)) if frames[i].startswith('E 7 '))
archive = (ticket/'reference/validation/trail-first-solution-trace.log').read_text()
selected = frames[contradiction-1:contradiction+3] + frames[restored-1:restored+3]
assert all(line in archive for line in selected)
cut = Machine(trail=True,first_only=True).run()
result = {'first_rows':unpack(model.output[0]),'first_word':f'{model.output[0]:06X}',
          'initial_propagated_domains':[f'{(int(frames[9].split()[2],16)>>(8*c))&255:02X}' for c in range(8)],
          'first_contradiction_and_retry':selected,'cut_output_frame':cut.events[-2].splitlines()[0],
          'cut_complete_frame':cut.events[-1].splitlines()[0],
          'snapshot_ms_at_10mhz':53951/10000,'trail_ms_at_10mhz':74523/10000,
          'cycle_increase_percent':100*(74523/53951-1),
          'traffic_increase_percent':100*(106480/69888-1)}
(ticket/'reference/validation/report-examples.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
