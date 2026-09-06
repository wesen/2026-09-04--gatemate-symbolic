from pathlib import Path
T=Path(__file__).resolve().parents[1]
p=T/'reference/01-implementation-diary.md'
s='''

## Step 14: Publish the LFL1 technical article to the Obsidian vault

Wrote a new 5,526-word textbook-style article explaining the implemented language from demand semantics through parsing, lexical scope, packed compilation, graph execution, synchronous FPGA memory, UART control, and source-aware inspection. Preserved the older Lab 4 article and copied seven existing screenshots into the dated vault assets folder so the new report is self-contained.

The report grounds its execution examples in the qualified source at `64ebbe2` and the physical test log. It distinguishes semantic agreement from cycle agreement, retained trace entries from aggregate mutations, and a requested list prefix from a confirmed terminating nil. Published only the report and its seven images in vault commit `1a56e1108225f8b4a61670d5e8221e9e45d5302e`, then pushed `origin/main` and verified both references match.

### Prompt Context

**User prompt (verbatim):**

```text
write a detailed project report for the obsidian vault as a deep dive technical analysis blog post using a textbook writing style (no analogies, see skill).      
 Commit and push the bsidian vault when done (go-go-parc vault).   
```

**Assistant interpretation:** Write a new explanatory article about LFL1's operation, reuse UI captures, and publish it in go-go-parc.

**Inferred user intent:** Preserve a durable technical account that teaches the system independently of its implementation chronology.

### What I did

- Applied the vault-writing, textbook-authoring, Obsidian Markdown, and diary instructions.
- Rechecked the physical qualification log, API declarations, source file paths, prior article style, and vault state.
- Retained scripts 29–33 for article generation, validation/publication, Mermaid rendering, scoped vault commit, and this diary entry.
- Wrote `sources/lfl1-project-report.md`, including two Mermaid diagrams, seven screenshots, detailed heap-address examples, pseudocode, representation tables, API references, and pinned source links.
- Validated local targets for every pinned source link and related-note wikilink, balanced code fences, and uniquely matched each screenshot to its source capture.
- Rendered both Mermaid diagrams using the local browser bundle in a temporary page, then closed that page without disturbing the running IDE.
- Copied the article and images, verified SHA-256 hashes, staged exactly eight intended files, checked the staged diff, committed, and pushed the vault.

### Why

- The existing implementation handoff explains maintenance and qualification; the new article develops the execution principles in teaching order.
- Keeping physical and model evidence explicitly labeled prevents an accurate screenshot from supporting an inaccurate timing claim.
- A self-contained vault asset directory preserves the article when the source checkout is unavailable.

### What worked

- Physical shared-expression counts explain one multiplication, three additions, four claims, and four updates against concrete heap addresses.
- The squares counts reconcile 438 mutation events with 64 retained entries and 374 drops.
- Both Mermaid diagrams rendered successfully in Playwright.
- The vault push succeeded: `9581261..1a56e11 main -> main`; local HEAD and origin/main both equal the full publication commit above.
- Existing unrelated notes were excluded from staging. The final vault status retained the unrelated untracked Research note.

### What didn't work

The first `python3 .../scripts/30-publish-project-report.py` validation attempt calculated the repository root one directory too shallow and stopped before publication:

```text
AssertionError: (PosixPath('/home/manuel/code/wesen/2026-09-04--gatemate-symbolic/ttmp'), 'pkg/lazylang/syntax/parser.go')
```

Changed `root=T.parents[3]` to `root=T.parents[4]`. The next validation passed. This was a report-validation script error, not a project implementation failure. No application code changed or application test reruns were needed.

### What I learned

- The named global `double` program has a different claim count from an inline lambda because the global function is itself represented by a thunk.
- Prefix completion and stream termination need separate explanations; eight observed values do not establish the next constructor.
- The routed implementation exceeded its planned CPE target while satisfying physical fit and timing, so the article records both facts.

### What was tricky to build

The main difficulty was maintaining identity across several forms of evidence: source artifact, runtime heap addresses, backend, and observation boundary. Grounded the address walkthrough in the named shared example, labeled model-only transient-state screenshots, and used the physical test log for physical counters. The validation script's initial root calculation was corrected before any vault copy.

### What warrants a second pair of eyes

Review the shared-address walkthrough and the distinction between enabled non-stall ticks and wall-clock latency. Check that future changes do not update these historical results without changing the pinned source revision.

### What should be done in the future

N/A for this publication. Future language extensions should receive new dated reports rather than overwriting this qualified snapshot.

### Code review instructions

Read `sources/lfl1-project-report.md` alongside `reference/validation/i4-physical-tests.log`. Inspect `reference/validation/project-report-manifest.json` for all eight published paths and hashes. Script 30 validates content references; script 31 renders the diagrams. Inspect vault commit `1a56e11` to confirm its exact scope.

### Technical details

- Vault article: `Projects/2026/09/05/ARTICLE - GateMate Symbolic - Inside a Lazy Functional Language.md`.
- Article length: 5,526 whitespace-delimited words; 403 lines.
- Assets: seven original PNG captures, copied without modification.
- Source implementation: `64ebbe291a4a1e23c8cb21a6dcb2db2be767b9f2`.
- Vault publication: `1a56e1108225f8b4a61670d5e8221e9e45d5302e`.
- Remote: `ssh://git@github.com/go-go-golems/go-go-parc`.
'''
assert '## Step 14:' not in p.read_text()
p.write_text(p.read_text()+s)
(T/'reference/validation/project-report-publication.txt').write_text('Vault commit: 1a56e1108225f8b4a61670d5e8221e9e45d5302e\nPush: 9581261..1a56e11 main -> main\nHEAD equals origin/main verified.\nMermaid: 2/2 rendered successfully.\nArticle and seven PNG asset hashes verified before commit.\n')
