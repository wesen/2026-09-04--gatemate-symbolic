---
Title: Implementation Diary
Ticket: GATEMATE-SYMBOLIC-006
Status: active
Topics:
    - fpga
    - gatemate
    - symbolic-computers
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-09-04T18:26:11.799692756-04:00
WhatFor: ""
WhenToUse: ""
---

# Implementation Diary

## Goal

Record design, implementation, validation, failures, and physical delivery of the graph-coloring search microscope.

## Step 1: P1 architecture and executable boundaries

Created the graph-coloring microscope ticket and an intern guide before implementation. The design adds runtime graph input and lossless semantic event stepping to the proven rollback discipline, then connects it to a Go session owner and React inspector.

### Prompt Context

**User prompt (verbatim, JSON encoded):**

```json
"Ok, create anew docmgr ticket for that, Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.\n\nThen implement it, using golang + typescript/react for the frontend.\n\ncommit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill).\n\nPrint out a brutalist work slip with the plan / different phases for the ticket. then before stsarting a phase, plrint a split about the phase, and print one when the phase is done."
```

**Assistant interpretation:** Design, publish, and implement the graph-coloring FPGA microscope with Go and React, committed milestones, a detailed diary, and physical phase slips.

**Inferred user intent:** Deliver a working next laboratory with enough explanation and evidence for an intern to maintain it.

### What I did

Mapped queens_core and shared RAM/UART sources; inspected Glazed APIs and primary Go/RTK references; defined graph bounds, event records, protocol checksums, HTTP endpoints, projection ownership, and six implementation phases. Printed overall plan and P1 start receipts.

### Why

The project requires actual FPGA graph loading and internal state visibility. A fixed solver output viewer would not satisfy that scope. Explicit contracts make the model, RTL, transport, and frontend independently testable.

### What worked

Ticket GATEMATE-SYMBOLIC-006 and design/diary documents created. Both printer receipts report printed:true. The design defines at most eight vertices/colors and a 256-event retained history.

### What didn't work

A skill lookup initially used /home/manuel/.codex/skills/defuddle/SKILL.md and reported No such file or directory; the catalog path /home/manuel/.agents/skills/defuddle/SKILL.md was then read successfully. No implementation tests have run yet.

### What I learned

The old trace_valid pulse cannot support lossless serial observation without backpressure and RAM write-enable gating. The UART receive pin is currently unused. The repository needs its first root Go module.

### What was tricky to build

A pending event must hold its supporting state while the host is slow. Controller and memory writes share the same advance condition. Debug staging words are only semantically meaningful on selected events, which the wire contract states explicitly.

### What warrants a second pair of eyes

Review serial framing widths, partial-command recovery, output cut semantics, and historical-snapshot ownership. The current Go web skill contains an incorrect registration-order claim; the guide follows ServeMux pattern specificity from the primary API.

### What should be done in the future

Upload the guide, commit the design, and implement the Go model/protocol before graph RTL.

### Code review instructions

Run docmgr doctor --ticket GATEMATE-SYMBOLIC-006. Review the current-source table and wire-payload widths in the design doc.

### Technical details

Protocol: L with 12 encoded bytes including checksum; S and R commands; E event with 33 payload bytes plus XOR checksum. Events are 70 ASCII bytes. Full design and references are local to the ticket.

## Step 2: P2 first exhaustive test exposed root propagation

Implemented the Go graph model, fixed-width wire codec, and checked host projection. The first exhaustive test immediately exposed a graph-specific difference from queens: a one-color input begins with singleton domains, so protected writes can occur before any choice exists.

### Prompt Context

**User prompt (verbatim, JSON encoded):**

(see Step 1)

**Assistant interpretation:** Design, publish, and implement the graph-coloring FPGA microscope with Go and React, committed milestones, a detailed diary, and physical phase slips.

**Inferred user intent:** Deliver a working next laboratory with enough explanation and evidence for an intern to maintain it.

### What I did

Added root go.mod and pkg/microscope graph, model, protocol, projection, and tests covering every simple four-vertex graph with one through three colors.

### Why

Exhaustive small cases exercise initial propagation as well as ordinary choices and restoration.

### What worked

Go formatting and module resolution succeeded. Codec and directed model tests reached execution.

### What didn't work

Command: go test ./pkg/microscope -count=1. Error: --- FAIL: TestAllSmallGraphs (0.00s); model_test.go:88: event 1 WRITE: invalid mutation record. This is the first failed implementation test and no repair has yet been attempted.

### What I learned

A valid root propagation write has choice level zero. The queens integrity rule requiring nonzero levels cannot be copied unchanged into a configurable graph solver.

### What was tricky to build

For one color, an edge removes the only color from its neighbor before a CREATE event. The old-domain record still must be logged, but its level is zero and no checkpoint will restore it on root exhaustion.

### What warrants a second pair of eyes

Allow level zero only according to the live choice depth; retain old-mask and narrowing checks. Update the graph RTL integrity rule and guide consistently.

### What should be done in the future

Apply this focused root-level correction and rerun the model suite.

### Code review instructions

Reproduce with go test ./pkg/microscope -run TestAllSmallGraphs -count=1.

### Technical details

The current test enumerates all 64 simple graphs on four vertices at colors 1, 2, and 3. Separately, the P1 staged whitespace check reported a trailing blank line in generated changelog.md; the commit command still ran because the shell was not fail-fast. Subsequent commit sequences use set -e and the formatting will be corrected.

## Step 3: P2 executable graph model and checked event contract complete

Completed the graph model, serial codec, and immutable checked state projection. The focused root-level trail correction passed on its first repair run; the exhaustive graph sweep now agrees with an independent assignment oracle.

### Prompt Context

**User prompt (verbatim, JSON encoded):**

(see Step 1)

**Assistant interpretation:** Design, publish, and implement the graph-coloring FPGA microscope with Go and React, committed milestones, a detailed diary, and physical phase slips.

**Inferred user intent:** Deliver a working next laboratory with enough explanation and evidence for an intern to maintain it.

### What I did

Implemented Graph validation and symmetric adjacency encoding, explicit semantic Model states, 26-byte load and 70-byte event framing, checksums, and Apply validation for every event. Added tests for all 64 four-vertex graphs at one through three colors, directed eight-vertex/first-only cases, capacities, cancellation, malformed frames, sequence gaps, record corruption, and snapshot aliasing.

### Why

These contracts let the FPGA and browser be checked independently against complete semantic observations.

### What worked

go test ./pkg/microscope -count=1 passed after permitting root-level writes. The suite covers all 192 small graph/color combinations plus directed codec and fault cases. P1 guide upload is confirmed by OK: uploaded in reference/validation/design-upload.log.

### What didn't work

The initial root-propagation failure is recorded in Step 2. Its first focused correction succeeded; no additional failure occurred. The guide now records level-zero graph trail entries as an implementation finding.

### What I learned

Complete domain masks and typed choice/trail deltas allow stronger host validation than displaying raw trace words. Snapshot cloning is required so later writes cannot mutate history.

### What was tricky to build

Packed numeric fields use big-endian wire bytes while vertex zero occupies the least significant domain byte. The decoder reverses those eight bytes explicitly. A checksum failure is separate from a semantically invalid transition.

### What warrants a second pair of eyes

Review event-specific projection rules and graph validation before introducing device I/O. Hardware must permit level-zero entries for root propagation and gate all memory writes during trace stalls.

### What should be done in the future

Implement graph-specific RTL and UART command/event transport, then compare physical and simulated events against this model.

### Code review instructions

Run go test -race ./pkg/microscope -count=1; read graph.go, model.go, protocol.go, and projection.go. No board execution is claimed for P2.

### Technical details

New root module github.com/wesen/2026-09-04--gatemate-symbolic uses Go 1.26.1 and github.com/pkg/errors. Model and serial devices will implement the same explicit Engine contract; source identity remains visible to the API.
