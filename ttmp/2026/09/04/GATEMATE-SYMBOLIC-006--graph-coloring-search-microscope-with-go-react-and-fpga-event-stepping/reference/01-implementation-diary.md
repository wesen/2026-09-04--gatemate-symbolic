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
