#!/usr/bin/env python3
from pathlib import Path
import json
p=Path('internal/dataflowide/session.go');s=p.read_text().replace('s.results = []df.Token{}\n', 's.results = []df.Token{}\n        s.events = []Event{}\n')
s=s.replace('s.step++\n\treturn nil','if s.scenario.Actions[s.step].Kind != "expect" { s.eventLocked(fmt.Sprintf("action %d: %s", s.step+1, s.scenario.Actions[s.step].Kind), "completed", false) }\n    s.step++\n\treturn nil')
s=s.replace('s.step++\n\t\t\ts.mu.Unlock()', 'if s.scenario.Actions[s.step].Kind != "expect" { s.eventLocked(fmt.Sprintf("action %d: %s", s.step+1, s.scenario.Actions[s.step].Kind), "completed", false) }\n            s.step++\n\t\t\ts.mu.Unlock()')
p.write_text(s)
p=Path('web/vite.dataflow.config.ts');s=p.read_text().replace('plugins: [react()]',"plugins: [react(), { name: 'dataflow-entry', configureServer(server) { server.middlewares.use((request, _response, next) => { if (request.url === '/') request.url = '/dataflow/index.html'; next(); }); } }]");p.write_text(s)
p=Path('web/tsconfig.json');v=json.loads(p.read_text());v['include'].append('vite.dataflow.config.ts');p.write_text(json.dumps(v,indent=2)+'\n')
