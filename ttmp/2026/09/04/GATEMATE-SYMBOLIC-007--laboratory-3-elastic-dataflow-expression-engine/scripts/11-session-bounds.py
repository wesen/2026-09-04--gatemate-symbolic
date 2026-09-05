#!/usr/bin/env python3
from pathlib import Path
p=Path('internal/dataflowide/session.go');s=p.read_text().replace('state.Current = s.frames[len(s.frames)-1]','state.Current = s.frames[len(s.frames)-1]\n        state.Current.Snapshot = state.Current.Snapshot.Clone()').replace('return f, true','f.Snapshot = f.Snapshot.Clone()\n            return f, true').replace(' && !errors.Is(err, context.Canceled)','');p.write_text(s)
p=Path('pkg/dataflow/transaction.go');s=p.read_text().replace('m.Trace = append(m.Trace, Activation{m.Metrics.Cycles, i.Context, i.Epoch, i.Node, v})','m.Trace = append(m.Trace, Activation{m.Metrics.Cycles, i.Context, i.Epoch, i.Node, v})\n                if len(m.Trace) > 4096 { m.Trace = m.Trace[len(m.Trace)-4096:] }');p.write_text(s)
