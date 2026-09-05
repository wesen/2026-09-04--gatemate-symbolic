---
Title: Implementation diary
Ticket: GATEMATE-SYMBOLIC-008
Status: active
Topics:
    - fpga
    - gatemate
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-09-04T22:09:27.769234375-04:00
WhatFor: ""
WhenToUse: ""
---

# Implementation diary

## Goal

Record the programmable workbench design, implementation, validation, physical evidence, and delivery.

## Step 1: Intern guide, execution contracts, and design delivery

Created the new workbench ticket and an intern guide that explains the existing engine before specifying programmable descriptors, a typed compiler, hardware stopping, bounded trace, and React integration. The design preserves seven physical nodes per context and makes capacity and trace loss explicit.

Printed the complete six-phase plan and phase-one start slip. Uploaded the completed guide to reMarkable before implementation; the receipt identifies the ticket-specific remote directory.

### Prompt Context

**User prompt (verbatim):** Exact request as JSON:

```json
"yes do it. Create a new docmgr ticket, then Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.\n\nThen implement, commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill). slPrint out a brutalist work slip with the plan / different phases for the ticket. then before stsarting a phase, plrint a split about the phase, and print one when the phase is done."
```

**Assistant interpretation:** Build the programmable graph/compiler/debugger workbench with an intern guide, physical evidence, and phased delivery.

**Inferred user intent:** Make the existing FPGA execution engine programmable and inspectable while preserving a reviewable account of its correctness.

### What I did

- Created GATEMATE-SYMBOLIC-008, six tracked phase tasks, the design document, and this diary.
- Read existing model, RTL, UART, session, and graph-renderer code; related five implementation paths to the design.
- Retained authoring, printing, upload, and diary helpers in scripts/ and the exact user request in sources/.
- Uploaded GATEMATE 008 Programmable Dataflow Workbench Design.pdf to /ai/2026/09/04/GATEMATE-SYMBOLIC-008.

### Why

- Programmability affects finality, routing, readiness, loading, and historical graph identity; changing only the editor would be incomplete.
- Staged loading after reset prevents old tokens from acquiring a new descriptor interpretation.

### What worked

- Both initial printer receipts report printed: true.
- The upload reports OK: uploaded; no redundant cloud listing was needed.

### What didn't work

- No failures in this phase.

### What I learned

- The current router special-cases COPY node six and final node five; both assumptions must be removed from runtime execution.
- Seven physical nodes are sufficient for the proposed example, but compiler-inserted fanout can consume the remaining capacity.

### What was tricky to build

- Trace candidates can coincide on one clock. The design requires explicit dropped-event accounting rather than pretending a one-record-per-cycle trace is complete.
- Source maps must remain associated with loaded graphs and historical generations.

### What warrants a second pair of eyes

- Review graph activation preconditions, byte encoding, breakpoint post-edge semantics, and trace loss before implementing hardware.

### What should be done in the future

- Implement P2 graph validation and compiler tests, followed by programmable RTL and debug contracts.

### Code review instructions

- Read design-doc/01-programmable-dataflow-workbench-intern-analysis-design-and-implementation-guide.md and reference/validation/p1-upload.log.
- Inspect plan-print.log and p1-start-print.log for physical print receipts.

### Technical details

- Source baseline aaad4ff; active count 1..7; four contexts; two outgoing destinations; 32 proposed trace records.
- Phase task IDs: 13x4, ycgx, g08b, 2bb4, cho9, 54tm.
