---
Title: Implementation diary
Ticket: GATEMATE-SYMBOLIC-010
Status: active
Topics:
    - fpga
    - gatemate
    - symbolic-computers
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://ttmp/2026/09/05/GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector/scripts/22-publish-report.py
      Note: Vault publication
    - Path: repo://ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide/scripts/02-collect-sources.sh
      Note: Primary source collection
    - Path: repo://ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide/scripts/03-docs.py
      Note: Diary and document authoring
    - Path: repo://ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide/scripts/04-contract-check.py
      Note: Layout check and diagnosed helper collision
    - Path: repo://ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide/scripts/05-review-design.py
      Note: Review corrections
ExternalSources: []
Summary: ""
LastUpdated: 2026-09-05T18:34:32.906562561-04:00
WhatFor: ""
WhenToUse: ""
---


# Implementation diary

## Goal

Record the report delivery and design of a lazy functional language, with exact evidence, decisions, failures and future implementation gates.

## Step 1: Publish the reducer report and establish the follow-up contract

I first wrote a 4,661-word technical article explaining the physically qualified Lab 4 machine. The article follows representation, continuations, claim/update invariants, execution traces, arithmetic, synchronous memory, UART, observation history and verification. Five physical screenshots were copied into the dated vault assets directory, and only the new article and those assets were committed and pushed.

I then created GATEMATE-SYMBOLIC-010 and gathered the existing runtime contract and primary lazy-language references. The new ticket separates completed design delivery tasks from future implementation tasks. Its central change is a heap-reference result model with closures, lexical environments and lazy constructor fields.

### Prompt Context

**User prompt (verbatim):**

ok, first  write a detailed project report for the obsidian vault as a deep dive technical analysis blog post using a textbook writing style (no analogies, see skill).      
 Commit and push the bsidian vault when done (go-go-parc vault).   

Then create a new docmgr ticket and Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable. For the cool follow up project. 

Then commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill). Print out a brutalist work slip with the plan / different phases for the ticket. then before stsarting a phase, plrint a split about the phase, and print one when the phase is done.

**Assistant interpretation:** Publish the completed reducer report, then prepare and deliver the follow-up language design with a detailed implementation plan.

**Inferred user intent:** Preserve a technical explanation of the working machine and make the next project concrete enough to implement task by task.

**Commit (documentation):** go-go-parc 23de3f4 — docs(gatemate): explain the lazy graph reducer with physical execution figures

### What I did

- Read pkg/lazy/types.go, model.go, reference.go, serial.go, physical_test.go, lazy_core.sv, lazyide/session.go, http.go and frontend types.ts.
- Created the report in ticket 009 reference/03 and published a self-contained vault article through ticket 009 scripts/22-publish-report.py.
- Verified all five relative screenshot targets, balanced fences and SHA-256 asset identities; committed only six intended vault paths and pushed main.
- Created ticket 010, design and diary documents, a source-evidence reference, and three design plus six future implementation tasks.
- Used Defuddle for the Microsoft Research and arXiv pages; retained the STG PDF and pdftotext extraction in sources/.
- Printed REPORT START successfully; attempted REPORT DONE, plan and D1 START, retaining exact errors and generated YAML.

### Why

- The user explicitly requested report publication before creating the follow-up ticket.
- Direct source evidence separates the implemented arithmetic-only reducer from the proposed functional language.
- The proposed language needs weak head normal form and allocation semantics before a useful compiler or IDE can be implemented.

### What worked

- Vault commit 23de3f4 pushed successfully to ssh://github.com/go-go-golems/go-go-parc.
- The report contains a real seven-node execution, physical mutation cycles 9 and 61, cycle error at 23, all fault codes, three diagrams and five physical screenshots.
- Primary source downloads and PDF text extraction succeeded.
- The existing unrelated untracked AgentForum vault note was preserved.

### What didn't work

- Initial discovery reads used nonexistent pkg/lazy/examples.go and web/src/lazy/source.ts; rg located the definitions in types.go and types.ts. No software change was needed.
- One functions.exec construction failed before execution with SyntaxError: Invalid or unexpected token because nested Markdown fence characters ended a JavaScript string; rewritten using ordinary quoted command text.
- REPORT DONE returned: remote almanach returned 500: chrome render stage viewport failed ... fork/exec /headless-shell/headless-shell: resource temporarily unavailable. One retry returned the same failure; the first receipt is retained separately.
- The ticket plan request timed out with Client.Timeout exceeded while awaiting headers. D1 START returned the same remote resource error. No remote service changes or debugging fixes were attempted.

### What I learned

- The existing runtime memoizes THUNK cells, not every arithmetic node; repeated root force can repeat additions while preserving the multiplication count.
- The STG PDF title page is version 2.5 dated July 9, 1992; a date initially inferred from the URL directory was corrected against the title page before committing.
- Current renderer availability prevents confirming physical phase slips; generated layouts must not be reported as printed.

### What was tricky to build

- The vault contains unrelated work, so the commit used git commit --only with the article and five exact asset paths.
- Source expression identity and heap object identity must remain separate because recursion allocates multiple objects from one code location.
- Constructor return requires preserving lazy field references; copying only an integer payload from the old update path cannot implement that contract.

### What warrants a second pair of eyes

- Check the article against revision 79f28e6 and the physical validation JSON; cycle totals include stalls.
- Review linked environments versus flat free-variable vectors: simplicity costs lookup work and retained heap space.
- Keep paper claims separate from project-specific capacity targets and proposed APIs.

### What should be done in the future

- D2 will specify the full language, compiler, object and frame layouts, transactional allocation, UART and source-aware IDE.
- Print the retained slips when the external rendering service is available; no failed receipt establishes that paper was produced.

### Code review instructions

- Read the report in go-go-parc Projects/2026/09/05 and check _assets/gatemate-lazy-fpga-*.png.
- Inspect sources/evidence.md and the downloaded title page and sections 3.1/4.2.
- Review tasks.md: I1–I6 are future implementation work and remain unchecked.

### Technical details

- Report source revision: 79f28e6; published vault revision: 23de3f4.
- The report's upload target was the Obsidian vault; the new design document will go to /ai/2026/09/05/GATEMATE-SYMBOLIC-010 on reMarkable.
- Scripts remain under ticket scripts/, and downloaded resources remain under sources/. No runtime source changed in this phase.

## Step 2: Specify the compiler, allocated machine and source-aware IDE

I wrote the full intern guide around a reference-returning lazy runtime. The design defines a monomorphic curried language, linked lexical environments, code descriptors, allocated values, thunk update to indirection, and a Go/RTL execution contract. It includes worked source programs and a manually lowered sharing example with predicted allocations and counters.

I then reviewed the design's cross-component boundaries. The guide now specifies the 80/128/256-bit formats, all low observation pages and counters, a source/run identity model, and a resumable stream operation. A small checker validates the proposed layout arithmetic and serial golden examples without claiming to execute an unimplemented language.

### Prompt Context

**User prompt (verbatim):**

(see Step 1)

**Assistant interpretation:** Publish the completed reducer report, then prepare and deliver the follow-up language design with a detailed implementation plan.

**Inferred user intent:** Preserve a technical explanation of the working machine and make the next project concrete enough to implement task by task.

**Commit (documentation):** 474faaa — baseline, published report source and design-ticket checkpoint

### What I did

- Wrote sources/design.md and populated the design-doc: 9,388 words with three Mermaid diagrams, pseudocode, worked programs, API tables, file maps and I1–I6 implementation substeps.
- Selected Int, Bool, ListInt and curried function types with explicit annotations, pure expressions, lazy arguments and constructor fields.
- Specified 2048x80 heap, 2048x128 code, 512x128 stack, provenance and bounded trace; logical storage totals 540672 bits before physical packing.
- Defined transactional allocation and nonallocating ERROR/UPDATE unwinding; specified productive recursive lists separately from cyclic scalar demand.
- Specified LFL1 loading, query pages, HTTP compile/load/control APIs, immutable artifacts, source spans, run IDs and resumable list Next.
- Added scripts/04-contract-check.py, format-schema.json and a hand-derived shared-design-fixture.json.
- Corrected signed-literal parsing and compiler-pass ordering during review; added precise observation pages, state numbers and counter definitions.
- Generated D1 DONE and D2 START layouts without claiming physical printing while the remote renderer was unavailable.

### Why

- Compound lazy values require stable heap references and cannot use the integer-only direct update of Lab 4.
- A linked environment makes lexical scope and recursion explicit for an intern, while its space and lookup costs remain visible.
- Multiobject allocation must publish only complete objects, and out-of-memory unwinding must not depend on further allocation.
- Source identity, allocation identity and demand origin differ and must not be collapsed in the UI.

### What worked

- The format checker passed after one helper-name fix, confirming 80-bit objects, 128-bit code/frames/capabilities and 256-bit traces.
- The hand-derived fixture contains 14 code records, 17 initial objects, root 15 and a predicted result reference 25; it is explicitly labeled as a design fixture.
- The golden force frame is F000f0f followed by LF; the begin and heap-write frames are retained in contract-check.json.
- The design makes the AST reference independent of emitted code and separates semantic comparison from machine/RTL layout comparison.

### What didn't work

- The first format-check run raised TypeError: pack() got multiple values for argument 'kind', because the helper selector collided with the frame field. Renaming the selector to layout_name fixed it; the next run passed. Exact error retained in contract-check-first-error.log.
- Physical printing remains unavailable after the D1 request failures; later layouts are retained as unprinted, not reported as successful receipts.

### What I learned

- Even the tiny shared function example has three thunk claims: main, x and the lazy argument. One multiplication does not mean one total runtime claim.
- A finite prefix of an unbounded list still consumes finite but unreclaimed heap in this design. The exact eight-square program must pass the capacity gate before hardware work.
- Packed 80-bit and 128-bit values require hex strings or decoded DTOs in JSON; the old Number-based 40-bit codec cannot be reused.

### What was tricky to build

- Application needs two distinct environment pointers: the argument thunk captures the caller, while the function body extends the closure's captured environment.
- A list Next spans several UART operations; substeps must be persisted so a budget pause or lost HTTP response cannot duplicate consumption.
- Allocation provenance survives mutation and trace loss, while event origin identifies the currently executing source span.
- The initial diagram grouped typing before resolution ambiguously; the final guide gives a single concrete pass order.

### What warrants a second pair of eyes

- Review the hand-derived fixture's predicted allocation sequence against the future compiler and machine; the checker only validates layout and framing.
- Check the private-allocation publication and trace sequence, especially pauses between object and provenance writes.
- Validate actual BRAM packing and routing in I4; the design's resource figures are targets, not measurements.
- Confirm all externally visible protocol fields against Appendix B before implementing consumers.

### What should be done in the future

- D3: validate docs and diagrams, upload the final guide and source reference to reMarkable, record receipts and push.
- I1–I6 remain future runtime/compiler/UI implementation work; this request produces the design package.

### Code review instructions

- Run python3 scripts/04-contract-check.py from the checkout and inspect reference/validation/contract-check.json.
- Read sections 8–10 for allocation and update obligations, sections 13–15 for transport and host workflow, and Appendix A for the concrete fixture.
- Use docmgr doctor and the delivery audit before upload.

### Technical details

- All scripts reside in ticket scripts/; source resources and design fixtures reside in sources/.
- No compiler, runtime, RTL, CLI or frontend implementation changed in this phase.
- The proposed target is separate from the qualified Lab 4 experiment and has no compatibility adapter.
