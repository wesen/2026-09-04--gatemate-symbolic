---
Title: Investigation diary
Ticket: GATEMATE-SYMBOLIC-002
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
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-002--tagged-stack-evaluator-implementation-review-and-intern-guide/design-doc/01-tagged-stack-evaluator-analysis-design-and-implementation-review.md
      Note: Primary review deliverable
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-002--tagged-stack-evaluator-implementation-review-and-intern-guide/scripts/01-review-probes.py
      Note: Reproducible model and RTL counterexamples
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-002--tagged-stack-evaluator-implementation-review-and-intern-guide/scripts/02-baseline.sh
      Note: Baseline validation workflow
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-002--tagged-stack-evaluator-implementation-review-and-intern-guide/scripts/03-collect-sources.sh
      Note: Defuddle source archive workflow
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-002--tagged-stack-evaluator-implementation-review-and-intern-guide/scripts/04-upload-review.sh
      Note: ReMarkable delivery workflow
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-002--tagged-stack-evaluator-implementation-review-and-intern-guide/sources/README.md
      Note: Source provenance
ExternalSources: []
Summary: Chronological evidence gathering, review authoring, validation, and delivery.
LastUpdated: 2026-09-04T15:40:00-04:00
WhatFor: Reproduce the investigation and distinguish current evidence from earlier claims.
WhenToUse: Reviewing or continuing the implementation review.
---


# Diary

## Goal

Review the implemented evaluator at baseline commit `c7f9dc9`, explain the complete system to a new intern, preserve reproducible evidence, and deliver the review through docmgr and reMarkable.

## Step 1: Establish scope and preserve executable review evidence

I created a separate review ticket instead of editing the earlier implementation history. The repository is a Python and SystemVerilog hardware project, despite the general Go-oriented workspace instructions. I mapped the assembler, executable model, two cores, board wrapper, tests, build flow, and original implementation commits before drawing conclusions.

The existing suite passes, but small programs exercising adjacent instructions and physical representation boundaries expose several discrepancies. I preserved probes and their raw output in the new ticket so the report can distinguish demonstrated defects from proposed improvements. Production implementation files remain at the reviewed baseline.

### Prompt Context

**User prompt (verbatim):**

> Analyze the work done here, and create a new docmgr ticket for review. Store all scripts you write in the scripts/ folder, store all downloaded resources you might want to consult in the sources folder (use defuddle for markdown). Create  a detailed analysis / design / implementation review report that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.
>
> commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill)

**Assistant interpretation:** Perform an evidence-backed review, author an intern guide and remediation design, commit the ticket artifacts at milestones, and upload the finished document.

**Inferred user intent:** Make the implemented system understandable and reviewable, including limitations that a successful demo does not reveal.

### What I did

- Read the ticket-research-docmgr-remarkable, docmgr, diary, remarkable-upload, and defuddle skills.
- Ran `git status --short`, `git log --reverse --format='%h %s'`, `rg --files symbolic_eval`, and read all RTL, Python tools, test harnesses, programs, build configuration, and prior diary evidence relevant to the review.
- Created `GATEMATE-SYMBOLIC-002` with `docmgr ticket create-ticket`, then added the design doc and diary using `docmgr doc add`.
- Added `scripts/01-review-probes.py`, `scripts/02-baseline.sh`, and `scripts/03-collect-sources.sh`; generated instrumented testbenches also live in `scripts/`.
- Ran `bash <ticket>/scripts/02-baseline.sh`. It sources the documented OSS CAD Suite environment, records versions, runs the baseline suite, collects test names, and executes isolated review probes.

### Why

- The previous guide describes the intended architecture. This review must compare that intent against actual observable behavior.
- Directed counterexamples make review findings actionable without making unrequested implementation changes.

### What worked

- Baseline: `123 passed in 6.11s`.
- Reproduced BRAM empty DUP, register-core post-EMIT depth corruption, both cores' PC wrapping, BRAM stale fault tags, and assembler acceptance of malformed/oversized input.
- Cycle instrumentation demonstrated that return depth changes one cycle before the CALL/RET commit trace, even though those traces match the model.
- Every core probe still printed `TB_PASS`; the additional model comparison exposes what the local monitors miss.

### What didn't work

- CAD tools were absent from the initial command path; the documented environment supplied them after notifying the user.
- `make versions` logged `/home/manuel/fpga/oss-cad-suite/bin/nextpnr-himbaechel: line 26: /home/manuel/.config/yosyshq/fonts.conf: Read-only file system`; the version command continued. This is an environment wrapper warning, not a design failure.
- Initial `bash <ticket>/scripts/03-collect-sources.sh` failed with `Error loading content: getaddrinfo ENOTFOUND yosyshq.readthedocs.io` under the network sandbox.
- The network-enabled retry reached the host but the unversioned documentation URL returned `Error loading content: Resource was not loaded. Status: 404`. Updated the script to the versioned primary-source URL returned by web search. No production software debugging or fixes were attempted.

### What I learned

- Matching retirement traces does not establish full architectural-state equivalence: data values are absent from ordinary commits, and return depth appears only in FINAL.
- Both the register and BRAM representations need instruction-boundary tests, not just individual opcode demonstrations.
- The current board has 514 operand slots (512 deep plus two cached), whereas most differential tests use 32 or fewer.

### What was tricky to build

- The testbench's own PASS is weaker than model agreement. The probe runner records both separately and retains unified diffs.
- Return-stack timing is invisible in normal TRACE/FINAL comparisons. I generated separate copies of the existing benches with a post-edge PC/return-depth/trace monitor, preserving the original harness unchanged.
- Tool discovery and source retrieval required separating missing command-path setup, filesystem wrapper warnings, sandbox DNS restrictions, and an actual missing web page.

### What warrants a second pair of eyes

- PC escape policy must be decided consistently for sequential execution and CALL return addresses.
- BRAM fault metadata needs the true second logical operand even for opcodes that otherwise do not fetch it.
- Early return-depth publication violates the documented commitment contract, but this probe does not demonstrate a failed current recursive program.

### What should be done in the future

- Promote minimal probes into normal regression tests as part of a separate implementation change.
- Add full-state refinement and per-cycle non-retirement assertions, including both stacks.

### Code review instructions

- Start with `scripts/01-review-probes.py`, then read `reference/validation/probe-results.json` and individual `.diff` files.
- Reproduce from the repository root with `bash <ticket>/scripts/02-baseline.sh`.
- Inspect `stack_core*-call-ret-timing.log` for RDEPTH changes while TRACE is zero.

### Technical details

- Toolchain: OSS CAD Suite `20260825`, Yosys `0.68+130`, Icarus Verilog `14.0 (devel)`, Python `3.11.6`.
- Traces use zero-based sequence and PC; data-stack depth is post-retirement, except a fault preserves the prior depth.
- Probe artifacts are evidence of baseline behavior, not passing correctness regressions.

## Step 2: Write the technical review and proposed remediation design

I wrote a roughly 9,500-word guide that explains the architectural machine before its physical representations. It covers instruction encoding, type and fault rules, the assembler/model APIs, clocked staging, both stacks, BRAM reads and refills, output ownership, UART framing, reset, synthesis, tests, and the earlier implementation history. The review separates fresh evidence from historical hardware observations throughout.

Nine finding groups connect concrete source locations to observed discrepancies or verification gaps. The report then proposes four design decisions, phased implementation work, and a validation matrix. It is a review deliverable: proposed changes are explicitly labeled and the production implementation remains unchanged.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Turn the investigation into a complete intern-oriented analysis and implementation review, including actionable remedies.

**Inferred user intent:** Make the current system safe to reason about and extend by understanding both its working design and its limitations.

**Commit (review artifacts):** `9c09a92` — "review: bootstrap GATEMATE-SYMBOLIC-002 and preserve baseline probes"

### What I did

- Wrote the primary design document with 15 sections, complete ISA/fault tables, five fixed-width explanatory diagrams, API/file references, pseudocode, finding severities, four proposed decisions, and five implementation phases.
- Archived the official Yosys synth_gatemate and Olimex board pages using defuddle; added `sources/README.md` with URLs, collection method, and original-book provenance.
- Ran `source /home/manuel/fpga/oss-cad-suite/environment && make synth PROG=fib` from `symbolic_eval/`, retaining the command output and full Yosys log.
- Related seven core implementation files to the primary report with `docmgr doc relate`.
- Verified findings against raw diffs, current test collection, and source declarations; corrected the blink-top parameter reference to LED_BIT during review.

### Why

- A novice needs the meaning of a state transition before the FSM and BRAM optimization can make sense.
- A useful review must say what to change, why, and which observations would demonstrate that the change worked.
- Resource counts need their tool phase attached: synthesis CC_L2T4 cells and routed packed CPEs are different measurements.

### What worked

- Source collection succeeded using the versioned Yosys URL after the errors recorded in Step 1.
- Fresh synthesis completed: two CC_BRAM_20K, one CC_MULT, 495 CC_L2T4 cells.
- The direct call/return trace matches the model while the cycle monitor reveals early return-depth publication, making a useful concrete example of verification limits.
- The first review milestone committed successfully after using the approved Git-metadata escalation.

### What didn't work

- Initial `git add <ticket> && git commit ...` failed with `fatal: Unable to create '/home/manuel/code/wesen/2026-09-04--gatemate-symbolic/.git/index.lock': Read-only file system`. Repeating the authorized commit with filesystem escalation succeeded.
- Synthesis emitted `Warning: Replacing memory \rstack_q with list of registers. See rtl/stack_core_bram.sv:731` and `Warning: Resizing cell port top.u_rom.mem.0.0.A_DO from 10 bits to 20 bits.` The build completed; these warnings are preserved rather than mislabeled as simulation failures.

### What I learned

- The original instruction metadata's generator claim is aspirational: the inspected RTL constants and testbench name tables are manually mirrored.
- A synchronized button input followed by an async-assert reset synchronizer does not make the physical button-to-reset path asynchronous end to end.
- The default BRAM read can collide with a spill into empty RAM; correctness depends on not consuming that incidental read result, rather than the comment's stronger no-collision claim.

### What was tricky to build

- The report must describe the intended precise machine while clearly identifying where implementation diverges; each finding is separated from the explanatory contract and has an observed result.
- Historical hardware claims are easily mistaken for fresh validation. The report ties previous observations to original diary steps and distinguishes them from this review's synthesis and simulation.
- PDF portability favors fixed-width diagrams over unrendered Mermaid source. The report uses plain text diagrams and avoids needing an external diagram-rendering service.

### What warrants a second pair of eyes

- DR-2's suggested wider PC also requires a FETCH fault representation in the RTL trace; merely widening one register is insufficient.
- Full-state snapshots must reconstruct tc=1 and tc=2 correctly and compare only live values.
- The proposed fault-context read must preserve the first-fault contract and all stack state while it completes.

### What should be done in the future

- Execute the report's five phases in a separate correctness implementation ticket.
- Re-establish routed timing and hardware captures after the fixes, preserving commit and program image identity.

### Code review instructions

- Read sections 1 and 10 first for the verdict and findings; sections 4–8 explain the system needed to assess them.
- Compare each reproduced finding against `reference/validation/` and its cited RTL/Python source.
- Check sections 11–12 for explicit decisions, file-level work, and acceptance observations.

### Technical details

- No production files changed. Baseline remains `c7f9dc9` for implementation references.
- Web pages: Yosys v0.50 command documentation and official Olimex product page; actual installed compiler version is recorded separately.
- The full report and diary will be bundled with a depth-two table of contents for reMarkable; delivery receipt follows in Step 3.

## Step 3: Validate the ticket and deliver the finished report

The completed review passed docmgr validation and was committed as a second milestone. I dry-ran reMarkable delivery, then narrowed the upload to the report after automatic approval review rejected including the diary. The report-only upload succeeded in the existing account hierarchy; the detailed diary and all investigation evidence remain available in the ticket.

This step completes the requested review and delivery. The ticket stays in review status because it presents findings for human review; the implementation fixes are a proposed follow-up, not unfinished documentation work.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Finish validation, commit the report, upload the requested deliverable, and preserve a truthful delivery receipt.

**Inferred user intent:** Receive a navigable technical report on the tablet with reproducible supporting work in the repository.

**Commit (report and evidence):** `18df652` — "docs: complete tagged evaluator review and remediation design"

### What I did

- Ran `docmgr doctor --ticket GATEMATE-SYMBOLIC-002 --stale-after 30`: all checks passed.
- Ran `git diff --check`, then committed the completed review, source archive, synthesis evidence, and diary.
- Added `scripts/04-upload-review.sh` and ran it with `--dry-run` before real delivery.
- Removed the diary from the external payload after the approval rejection, checked the original ticket's successful delivery record, and reran the dry-run for the report-only payload.
- Uploaded the report with a depth-two table of contents to `/ai/2026/09/04/GATEMATE-SYMBOLIC-002`.
- Recorded the exact successful receipt in `reference/validation/upload.log`; checked all documentation-delivery tasks.

### Why

- The user's direct request authorizes delivery of the finished report. Keeping the diary local resolved the reviewer's payload-scope concern without requiring a new user permission request.
- The remarkable-upload skill treats `OK: uploaded` as sufficient delivery evidence; no redundant cloud listing or account probe was needed.

### What worked

- Report-only dry-run and upload completed successfully.
- Exact receipt: `OK: uploaded GATEMATE-SYMBOLIC-002 Implementation Review.pdf -> /ai/2026/09/04/GATEMATE-SYMBOLIC-002`.
- The existing original-ticket diary documented prior delivery in the same `/ai/2026/09/04` account hierarchy.
- Docmgr validation and whitespace checks passed; all deliverable tasks are complete.

### What didn't work

- Automatic approval review rejected the initial report-and-diary upload before execution: `This uploads repository-derived review and diary contents to an external reMarkable destination whose ownership/trust is not established; the user authorized uploading the finished report but not the diary bundle or this specific sensitive payload to that destination.`
- I did not retry the rejected bundle. I reduced the payload to the explicitly requested report, cited the existing delivery evidence in the new escalation, and the narrowed action was approved.
- An inspection confirmed no upload log existed for the rejected action: `ls: cannot access 'ttmp/2026/09/04/GATEMATE-SYMBOLIC-002--tagged-stack-evaluator-implementation-review-and-intern-guide/reference/validation/upload.log': No such file or directory`. The successful action subsequently created that log.

### What I learned

- Skill-suggested bundling can exceed the payload scope recognized by automatic approval review. The report is the requested remote deliverable; the diary is still fully preserved locally.

### What was tricky to build

- Delivery needed a concrete payload adjustment rather than a repeated permission request or an indirect bypass. The upload script now contains only the primary report and an explanatory comment.
- Earlier diary entries record the initial plan to bundle both documents. This entry supersedes that plan with the actual report-only delivery, preserving the chronology.

### What warrants a second pair of eyes

- Readability on the physical tablet can be assessed by opening the uploaded PDF. Successful upload confirms delivery, not a manual visual inspection of every page.
- Findings F1–F9 remain review items; no claim is made that the underlying implementation defects were fixed.

### What should be done in the future

- Use report sections 11–12 to scope the next correctness ticket.
- Keep the historical board evidence distinct from new measurements when validating later implementation changes.

### Code review instructions

- Open the primary report, then follow its finding-specific `.diff` links/file names into `reference/validation/`.
- Confirm the upload receipt and docmgr doctor log.
- Inspect the three review commits and verify that production code was not changed.

### Technical details

- Remote document: `GATEMATE-SYMBOLIC-002 Implementation Review.pdf`.
- Remote folder: `/ai/2026/09/04/GATEMATE-SYMBOLIC-002`.
- Remote payload: primary report only. Local artifacts: report, full diary, scripts, downloaded sources, and raw validation evidence.
