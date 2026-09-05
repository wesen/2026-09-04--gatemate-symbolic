#!/usr/bin/env python3
from pathlib import Path
import json
p=Path('web/package.json');v=json.loads(p.read_text());v['scripts']['build:dataflow']='tsc --noEmit && vite build --config vite.dataflow.config.ts';v['scripts']['dev:dataflow']='vite --config vite.dataflow.config.ts --host 127.0.0.1';p.write_text(json.dumps(v,indent=2)+'\n')
p=Path('Makefile');s=p.read_text();s+='\n.PHONY: dataflow-frontend dataflow-dev\ndataflow-frontend:\n\tgo generate ./internal/dataflowide\ndataflow-dev:\n\tgo run ./cmd/dataflow-ide\n';s=s.replace('go generate ./internal/microscope','go generate ./internal/microscope ./internal/dataflowide');p.write_text(s)
p=Path('.gitignore');p.write_text(p.read_text()+'\n/web/dist-dataflow/\n/internal/dataflowide/assets/\n/elastic_dataflow/projects/\n')
