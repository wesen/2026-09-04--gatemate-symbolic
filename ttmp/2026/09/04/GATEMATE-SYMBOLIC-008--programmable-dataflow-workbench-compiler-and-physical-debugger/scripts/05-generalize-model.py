from pathlib import Path
p=Path('pkg/dataflow/types.go');s=p.read_text().replace('type state struct {','type state struct {\n Graph *Graph')
s=s.replace('t.Node >= Nodes || t.Port > 1 || Descriptors[t.Node].Required','t.Node >= s.graph().Count || t.Port > 1 || s.descriptor(t.Node).Required')
s=s.replace('v.Valid&Descriptors[t.Node].Required == Descriptors[t.Node].Required','v.Valid&s.descriptor(t.Node).Required == s.descriptor(t.Node).Required')
p.write_text(s)
for name in ['semantic.go','transaction.go','engine.go']:
 p=Path('pkg/dataflow')/name;s=p.read_text()
 import re
 s=re.sub(r'Descriptors\[([^]]+)\]',r'm.descriptor(byte(\1))',s)
 s=re.sub(r'(?<![.\w])completion\(',r'm.completion(',s)
 p.write_text(s)
p=Path('pkg/dataflow/engine.go');s=p.read_text().replace('Kind    string  `json:"kind"`','Kind    string  `json:"kind"`\n Graph *Graph `json:"graph,omitempty"`')
s=s.replace('switch o.Kind {\n\tcase "reset":','switch o.Kind {\n case "load":\n  if o.Graph==nil { return errors.New("load requires graph") }; return o.Graph.Validate()\n\tcase "reset":',1)
s=s.replace('type Snapshot struct {','type Snapshot struct {\n Graph Graph `json:"graph"`')
idx=s.index('func (m *Transaction) Execute');s=s[:idx]+s[idx:].replace('switch o.Kind {','switch o.Kind {\n case "load":\n  return nil, m.LoadGraph(*o.Graph)',1)
s=s.replace('s := Snapshot{Source: "model",','s := Snapshot{Graph: m.graph(), Source: "model",')
p.write_text(s)
