#!/usr/bin/env python3
"""Install the implemented guide into its docmgr document and index evidence."""
from pathlib import Path
import hashlib
import json
import subprocess
root = Path(__file__).resolve().parents[6]
ticket = Path(__file__).resolve().parents[1]
doc = ticket / 'reference/04-implemented-language-runtime-and-ide-intern-handoff.md'
front = doc.read_text().split('---', 2)[1]
front = front.replace('Summary: ""', 'Summary: "Implemented APIs, execution semantics, physical qualification and screenshot handoff."')
body = (ticket / 'sources/implementation-guide.md').read_text()
doc.write_text('---' + front + '---\n\n' + body)
index = ticket / 'index.md'
text = index.read_text()
old = 'The design package is the current delivery. The syntax package is now implemented: a bounded lexer and recursive-descent/Pratt parser, with source spans, diagnostics, recovery and six parsed examples. Binding resolution, type checking, compilation and runtime execution remain future work. Tasks D1–D3 cover analysis, design and publication; tasks I1–I6 remain the future implementation sequence. The qualified Lab 4 runtime remains a separate experiment.'
new = 'The complete language, compiler, allocated Go machine, synchronous FPGA runtime, checked UART and Go/React IDE are implemented. Six physical source programs match model heap/provenance and semantic behavior. The FPGA routes at 21.26 MHz against a 10 MHz constraint. Eleven UI screenshots are retained for the diary and later report. The qualified Lab 4 runtime remains a separate experiment.'
text = text.replace(old, new)
link = '- [Implemented runtime and IDE handoff](reference/04-implemented-language-runtime-and-ide-intern-handoff.md): current APIs, measured results, execution principles, reproduction commands and screenshot index.\n'
if link not in text:
    text = text.replace('## Read first\n\n', '## Read first\n\n' + link)
text = text.replace('design delivery and future implementation milestones.', 'completed design and implementation milestones.')
index.write_text(text)
guide = ticket / 'design-doc/01-lazy-functional-language-intern-analysis-design-and-implementation-guide.md'
text = guide.read_text()
notice = '\n\n## Implemented qualification update\n\nThe original design below is retained as the design record. The [implemented runtime and IDE handoff](../reference/04-implemented-language-runtime-and-ide-intern-handoff.md) describes the delivered APIs and measured behavior. All implementation phases are complete. Final routing reports 21.26 MHz, 34 RAM halves, 3,804 flip-flops and 14,359 CPE logic resources. Logic exceeds the preliminary 10,000-resource estimate. The Go machine and hardware agree semantically; synchronous hardware rereads cause different cycle, heap-read and trace-timestamp totals. Six physical programs and eleven UI screenshots provide the qualification evidence.\n'
if '## Implemented qualification update' not in text:
    guide.write_text(text + notice)
paths = list((ticket / 'reference/validation').glob('*.log')) + list((ticket / 'reference/screenshots').glob('*.png'))
evidence = [{'path': str(p.relative_to(ticket)), 'bytes': p.stat().st_size, 'sha256': hashlib.sha256(p.read_bytes()).hexdigest()} for p in sorted(paths)]
rtl = [root / 'lazy_language/rtl' / name for name in ['lfl_core.sv','lfl_link.sv','lfl_top.sv']]
bit = root / 'lazy_language/build/top.bit'
audit = {'implementation_commit': subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(), 'routed_mhz':21.26, 'constraint_mhz':10, 'ram_halves':34, 'cpe_logic':14359, 'flip_flops':3804, 'physical_programs':6, 'screenshots':11, 'rtl_sha256':{str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in rtl}, 'bitstream_sha256':hashlib.sha256(bit.read_bytes()).hexdigest(), 'evidence':evidence}
(ticket / 'reference/validation/implementation-audit.json').write_text(json.dumps(audit,indent=2)+'\n')
print(json.dumps({'guide_words':len(body.split()),'screenshots':len(list((ticket/'reference/screenshots').glob('*.png'))),'evidence_files':len(evidence)},indent=2))
