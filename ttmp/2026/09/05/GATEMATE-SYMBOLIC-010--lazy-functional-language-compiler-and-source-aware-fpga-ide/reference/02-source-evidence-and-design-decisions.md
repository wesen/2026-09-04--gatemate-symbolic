---
Title: Source evidence and design decisions
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
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-09-05T18:34:33.023928008-04:00
WhatFor: ""
WhenToUse: ""
---

# Source evidence and design decisions

## Evidence boundary

Lab 4 is implemented and physically qualified at source revision 79f28e6. The new lazy functional language is a design, not an implemented compiler or hardware image. The source article was published first, in go-go-parc commit 23de3f4, with five copied physical screenshots and a 4,661-word technical narrative.

Existing source files were read directly: pkg/lazy/types.go, model.go, reference.go, serial.go and physical_test.go; lazy_reducer/rtl/lazy_core.sv; internal/lazyide/session.go and http.go; and web/src/lazy/types.ts. The lab-book chapter is archived in the preceding ticket at sources/laboratory-4.md. It explicitly identifies constructor, environment, closure and allocation extensions, while the implemented baseline has no runtime allocator.

## Primary research consulted

Peyton Jones's STG paper describes a machine for non-strict higher-order languages, with an intermediate language and explicit operational semantics. Its closure discussion connects function code with captured variables and distinguishes values from unevaluated suspensions. We use it to identify the responsibilities a language runtime must specify; the proposed expression-code machine is not an STG implementation. [Microsoft Research publication](https://www.microsoft.com/en-us/research/publication/implementing-lazy-functional-languages-on-stock-hardware-the-spineless-tagless-g-machine/), [author-hosted PDF](https://www.microsoft.com/en-us/research/wp-content/uploads/1992/04/spineless-tagless-gmachine.pdf), especially sections 3.1 and 4.2.

Nakata and Hasegawa discuss small-step and big-step call-by-need semantics, including cyclic calculi and heap-based suspension and memoization. This supports separating a recursive semantic evaluator from the explicit hardware transition system. We do not claim their equivalence theorem proves this proposed bounded machine. [Research paper abstract](https://arxiv.org/abs/0907.4640).

The web pages were extracted with Defuddle into sources/stg-publication.md and sources/call-by-need-semantics.md. The PDF and a pdftotext extraction are stored beside them. Script 02-collect-sources.sh records the URLs and extraction commands. The PDF title page identifies version 2.5 dated 9 July 1992; the publication page dates the journal article July 1992. Bibliographic dates are separate from this ticket's retrieval date, 2026-09-05.

## Decisions to carry into the guide

| Decision | Rationale and cost |
|---|---|
| Use a small typed, curried language | It supports higher-order map and partial application without a multi-argument calling convention. |
| Return heap references to weak head normal form | Constructors can retain unevaluated fields; completed thunks can forward to compound results. |
| Use fixed 80-bit objects and 128-bit expression records | Explicit widths simplify initial memory scheduling; this increases BRAM use. |
| Capture a linked lexical environment | Binding semantics and recursive environments are inspectable; lookup is linear in lexical depth and captures can retain excess objects. |
| Use a bounded bump allocator with no collection | Allocation and publication can be specified first; long-running streams eventually exhaust memory. |
| Keep the qualified Lab 4 target as a separate experiment | The new image and protocol have their own capabilities; no compatibility layer is required. |
| Keep source identity separate from runtime identity | One source expression can allocate many objects, and one shared object can be demanded from many source locations. |
| Preserve pause and reset semantics | Exhausted tick budget pauses with claims intact; only reset aborts the loaded experiment. |

These are project-specific design choices. Capacity targets, performance expectations, API shapes and proposed filenames are not facts taken from the research papers and are not measurements of a built system.
