#!/usr/bin/env python3
from pathlib import Path
root=Path(__file__).resolve().parents[1]
p=root/'reference/03-implemented-parser-api-and-validation-handoff.md';s=p.read_text()
body=Path('pkg/lazylang/syntax/README.md').read_text()
body+='\n## Recorded qualification\n\nThe focused suite passed with 94.5% statement coverage. A 15-second fuzz run completed 145,639 executions with no failure. Repository-wide Go tests, parser race tests, go vet ./... and go build ./... passed. Logs are retained under reference/validation/parser-*.log. The examples are syntax-qualified; no evaluation result is claimed.\n\nCode checkpoints: bde6d48 adds the lexer and AST; 460ba86 adds recursive-descent/Pratt parsing and recovery. The final phase adds examples, fuzz properties and this handoff.\n'
p.write_text('---'+s.split('---',2)[1]+'---\n\n'+body)
p=root/'index.md';s=p.read_text()
s=s.replace('The compiler and runtime are not implemented.','The syntax package is now implemented: a bounded lexer and recursive-descent/Pratt parser, with source spans, diagnostics, recovery and six parsed examples. Binding resolution, type checking, compilation and runtime execution remain future work.')
s=s.replace('No Lab 4 runtime code is changed by this ticket.','The qualified Lab 4 runtime remains a separate experiment.')
if '## Implemented parser' not in s:s+='\n## Implemented parser\n\n[Parser API and validation handoff](reference/03-implemented-parser-api-and-validation-handoff.md) documents pkg/lazylang/syntax. Parser substeps S1–S3 are complete; I1 remains open for the binding/type checker and independent semantic evaluator. The focused suite, repository tests, race check, vet, build and 145,639 fuzz executions passed.\n'
p.write_text(s)
(root/'sources/index-body.md').write_text(s.split('---',2)[2].lstrip())
