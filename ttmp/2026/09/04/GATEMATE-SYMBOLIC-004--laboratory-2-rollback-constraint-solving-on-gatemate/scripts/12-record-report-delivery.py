#!/usr/bin/env python3
"""Record report provenance, validate the delivered body, and finish the diary."""
from pathlib import Path
import hashlib
import json
import re
import subprocess

ticket = Path(__file__).resolve().parents[1]
vault = Path('/home/manuel/code/wesen/go-go-golems/go-go-parc')
note = vault/'Projects/2026/09/04/ARTICLE - GateMate Symbolic - Inside an FPGA Rollback Solver.md'
report = ticket/'reference/02-inside-the-eight-queens-rollback-engine-technical-project-report.md'
body = report.read_text().split('---',2)[2].strip()
assert body in note.read_text()
revision = subprocess.check_output(['git','rev-parse','HEAD'],cwd=vault,text=True).strip()
upstream = subprocess.check_output(['git','rev-parse','origin/main'],cwd=vault,text=True).strip()
assert revision == upstream
receipt = dict(vault_note=str(note),vault_commit=revision,origin_main=upstream,
               body_words=len(body.split()),sha256=hashlib.sha256(note.read_bytes()).hexdigest(),
               source_revision='4345ee4a9eb3460d6e94646d77e08dffa91489a0',body_matches_ticket=True)
(ticket/'reference/validation/report-vault-delivery.json').write_text(json.dumps(receipt,indent=2)+'\n')
diary = ticket/'reference/01-diary.md'
old = diary.read_text()
assert 'Textbook report, detailed backfill, and vault delivery' not in old
# Normalize section order only in newly authored retrospective entries.
prefix,rest = old.split('## Step 8:',1)
future = '\n### What should be done in the future\n\nUse this corrected commit attribution when reviewing the implementation. No new implementation work is implied by this retrospective entry.\n'
rest = rest.replace(future,'')
rest = rest.replace('\n### Code review instructions',future+'\n### Code review instructions')
old = prefix+'## Step 8:'+rest
prompt = 'write a detailed project report for the obsidian vault as a deep dive technical analysis blog post using a textbook writing style (no analogies, see skill).      \n Commit and push the bsidian vault when done (go-go-parc vault).   \n\n^-- for the go-go-parc vault.'
entry = '''
## Step 13: Textbook report, detailed backfill, and vault delivery

Wrote a 5,295-word technical article about the implemented rollback solver and delivered it as a new dated note in go-go-parc. The article teaches domain propagation, complete checkpoint state, synchronous memory publication, reverse restoration, fault precision, output acceptance, and the measured resource/cycle tradeoff. It emphasizes system behavior rather than implementation chronology.

The user also requested detailed diary backfill. Steps 8–12 preserve the original entries while adding retrospective implementation explanations, exact milestone attribution, print chronology, test provenance, and review instructions. They explicitly distinguish reconstructed detail from contemporaneous observations; they do not invent missing diagnostics or new test runs.

### Prompt Context

**User prompt (verbatim, JSON-encoded to preserve whitespace):**

```json
PROMPT_JSON
```

**Additional user prompt (verbatim):** (see Step 8)

**Assistant interpretation:** Write and publish the system-focused textbook report in the explicitly named vault, and make the implementation diary detailed enough to review and reproduce the completed work.

**Inferred user intent:** Preserve both a durable technical explanation and a trustworthy engineering record.

**Commit (analyzed implementation):** 4345ee4a9eb3460d6e94646d77e08dffa91489a0.

**Commit (vault report):** VAULT_REVISION — pushed to go-go-parc origin/main.

### What I did

- Read the vault-writing, textbook-authoring, and diary skills and inspected the existing tagged-stack CPU article to match vault metadata and technical depth.
- Read queens_core, queens_model, the mask helpers, formatter, board top, shared RAM/UART, test harness, profile metrics, first-result traces, and hardware evidence.
- Created the report in reference/02-inside-the-eight-queens-rollback-engine-technical-project-report.md and added scripts/09-report-examples.py to validate every selected contradiction/retry frame against the archived RTL trace and the model.
- Used scripts/10-backfill-diary.py for the retrospective entries and scripts/11-export-vault-report.py to validate and create a new vault note without replacing any historical note.
- Committed only the new vault article and pushed it. The pre-existing untracked AgentForum research note was left untouched.

### Why

The report needs to explain why the machine works, including the exact state that must be restored and the difference between a core handshake and a completed UART record. The diary needs to retain provenance so a later reader can identify the implementing commit rather than mistaking a previously available milestone for the change itself.

### What worked

The report example script passed and reproduced first board 672BE0, the first contradiction, restoration to mark 15, retry from F0 to 20, and the cut with top/base 29. Export checks verified more than 5,000 body words, balanced code fences, no template placeholders, and required example/metric anchors. The delivered vault body matches the ticket body. Vault HEAD and origin/main both identify the report commit. Docmgr doctor passed for the ticket.

### What didn't work

No model-example, export, Git, or push failure occurred. A read command accidentally included head -60 ' /dev/null'; it reported "head: cannot open ' /dev/null' for reading: No such file or directory" after the intended source reads had succeeded. This was a harmless discovery-command typo, not a software failure. No hardware or RTL changes were needed for the report, so the full 58-test suite was not rerun.

### What I learned

The implementation's terminal_sent name describes selection of the terminal record, not completed physical transmission. The UART divider is 87 at 10 MHz, so 932 bytes require at least 81.084 milliseconds of serial bit time; the eight-second capture window is an observation interval, not solve time. The report states those distinctions explicitly. The trail allocates more logical record bits than snapshots despite using fewer mapped RAM_HALF resources.

### What was tricky to build

The original diary's Steps 3–5 named the preceding milestone because entries were written before their code commits. The backfill preserves those historical entries and records the actual mapping: snapshots 0f15376, trail 214a39a, boundary checks 1a5194f. Raw trace words pack column zero into the least significant byte, so example decoding must follow that order. Physical evidence was restricted to what the UART can observe; internal trace and fault/reset claims are labeled as simulation evidence.

### What warrants a second pair of eyes

Review the record-bit calculations, the distinction between RAM_HALF and whole physical blocks, the serial-time lower bound, and the boundary where propagated bits are restored. Check that the article's fixed source revision and source line references remain appropriate if later implementation work is added.

### What should be done in the future

Use a new dated follow-up note for future solver extensions or new measurements. Do not silently rewrite this report's measured snapshot. No additional implementation or publication work is required for this request.

### Code review instructions

Run ticket scripts/09-report-examples.py and inspect reference/validation/report-examples.json. Compare the report body with the vault note and the stored report-vault-delivery.json hash. Inspect the vault commit to confirm it contains only the intended article. Read retrospective Steps 8–12 alongside the implementing commits and preserved print receipts.

### Technical details

Vault path: Projects/2026/09/04/ARTICLE - GateMate Symbolic - Inside an FPGA Rollback Solver.md. The note uses article frontmatter, native Mermaid diagrams, mathematical definitions, pseudocode, trace excerpts, resource/API tables, immutable source links, and two existing vault wikilinks. No external sources were downloaded, and all scripts written during this work are in the ticket scripts directory.
'''.replace('PROMPT_JSON',json.dumps(prompt)).replace('VAULT_REVISION',revision)
diary.write_text(old+entry)
print(json.dumps(receipt,indent=2))
