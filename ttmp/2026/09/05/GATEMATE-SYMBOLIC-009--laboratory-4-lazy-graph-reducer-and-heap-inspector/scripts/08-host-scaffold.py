#!/usr/bin/env python3
from pathlib import Path
import json
repo=Path(__file__).resolve().parents[6]
for name in ['assets_embed.go','assets_dev.go']:
 s=(repo/'internal/dataflowide'/name).read_text().replace('dataflowide','lazyide');(repo/'internal/lazyide'/name).write_text(s)
for target,source in [('lazy-ide','dataflow-ide'),('lazy-lab','dataflow-lab')]:
 s=(repo/'cmd'/source/'main.go').read_text().replace('dataflow-ide','lazy-ide').replace('dataflow-lab','lazy-lab').replace('dataflowide','lazyide').replace('/pkg/dataflow','/pkg/lazy').replace('dataflow.','lazy.').replace('var engine dataflow.Engine','var engine lazy.Engine').replace('dataflow IDE','lazy IDE').replace('dataflow scenario IDE','lazy heap inspector').replace('elastic dataflow example','lazy graph example').replace('127.0.0.1:8087','127.0.0.1:8089')
 s=s.replace('"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazy"','lazy "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazy"') if target=='lazy-lab' else s
 s=s.replace('engine, err = lazy.NewTransaction(lazy.DefaultConfig())','engine = lazy.NewModel()').replace('engine, err = df.NewTransaction(df.DefaultConfig())','engine = df.NewModel()')
 s=s.replace('fields.WithChoices("book", "copy", "fault", "cancel"), fields.WithDefault("book")','fields.WithChoices("shared", "cycle", "indirection", "indirection-cycle", "overflow"), fields.WithDefault("shared")')
 if target=='lazy-ide':
  s='\n'.join(line for line in s.split('\n') if 'Projects string' not in line and 'fields.New("projects"' not in line)
  a=s.index('\tprojects, err :=');b=s.index('\tvar engine',a);s=s[:a]+s[b:];s=s.replace('ide.NewHandler(session, projects)','ide.NewHandler(session)')
 (repo/'cmd'/target).mkdir(exist_ok=True);(repo/'cmd'/target/'main.go').write_text(s)
s=(repo/'scripts/build-dataflow-web.py').read_text().replace('dataflow','lazy');(repo/'scripts/build-lazy-web.py').write_text(s)
(repo/'internal/lazyide/generate.go').write_text('package lazyide\n\n//go:generate python3 ../../scripts/build-lazy-web.py\n')
s=(repo/'web/vite.dataflow.config.ts').read_text().replace('dataflow','lazy').replace('8087','8089');(repo/'web/vite.lazy.config.ts').write_text(s)
p=repo/'web/package.json';d=json.loads(p.read_text());d['scripts']['build:lazy']='tsc --noEmit && vite build --config vite.lazy.config.ts';d['scripts']['dev:lazy']='vite --config vite.lazy.config.ts --host 127.0.0.1';p.write_text(json.dumps(d,indent=2)+'\n')
