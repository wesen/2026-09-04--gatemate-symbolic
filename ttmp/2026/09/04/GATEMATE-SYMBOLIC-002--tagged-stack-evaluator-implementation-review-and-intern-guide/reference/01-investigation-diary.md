---
Title: Investigation diary
Ticket: GATEMATE-SYMBOLIC-002
Status: active
Topics: [fpga, gatemate, symbolic-computers, architecture]
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
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
