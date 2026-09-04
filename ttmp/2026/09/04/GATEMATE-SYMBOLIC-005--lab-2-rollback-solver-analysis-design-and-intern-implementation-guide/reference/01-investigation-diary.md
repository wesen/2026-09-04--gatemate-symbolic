---
Title: Investigation diary
Ticket: GATEMATE-SYMBOLIC-005
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
LastUpdated: 2026-09-04T17:02:17.080672905-04:00
WhatFor: ""
WhenToUse: ""
---

# Diary

## Goal

Preserve the Lab 2 design investigation, validation evidence, and explicit delivery status.

## Step 1: Preserve the interrupted design investigation

The investigation established the Lab 2 refactoring scope and validated a small software experiment before authoring the guide. The large document patch was rejected; recovery attempts also failed, so the guide remains unsaved and work stopped. The user then requested committing and pushing the current state and opening it in a browser.

### Prompt Context

**User prompt (verbatim):** "commit and push and open in browser"

**Assistant interpretation:** Publish the current repository checkpoint and open its web page without resuming the interrupted document recovery.

**Inferred user intent:** Make the current work accessible for inspection.

### What I did

Created the ticket, inspected book and RTL sources, wrote the independent oracle and snapshot/trail experiment, and rendered architecture and rollback diagrams. Before committing, updated the ticket to distinguish saved evidence from the unfinished guide.

### Why

The checkpoint must not imply that a generated document template is the requested completed deliverable.

### What worked

The experiment returned exactly 92 distinct boards in the required order. Snapshot and trail semantic events matched. Full enumeration measured 3980 effective domain writes, maximum trail depth 32, and maximum choice depth 6. Abstract restoration and limited capacity checks passed; these are not RTL or board results.

### What didn't work

The guide patch failed with `apply_patch verification failed: invalid patch: multiple operations target ...`. The first recovery run failed with `json.decoder.JSONDecodeError: Expecting value: line 1 column 1 (char 0)`. After one correction, the second run failed with `IndexError: list index out of range` while extracting the target Add File body. Neither run changed the guide. Work stopped under the user's two-attempt debugging rule.

### What I learned

For this small domain state, proposed 20-bit trail records generate more logical history-write traffic than the 72-bit snapshot payloads. Performance benefits require measurement rather than assumption.

### What was tricky to build

Recovery must select the intended document patch rather than another tool call containing the same marker text. The saved recovery script is diagnostic work in progress, not a reliable document-generation command.

### What warrants a second pair of eyes

Review the experiment's scope: it uses recursive software control and abstract stalls, and does not validate synchronous RAM timing or physical cut publication.

### What should be done in the future

Resume guide authoring or recovery when requested, validate the completed document, and upload it to reMarkable. No upload has occurred.

### Code review instructions

Start with `scripts/01-queens-design-experiment.py` and `reference/validation/design-experiment.json`. The intended guide under `design-doc/` is still a template. Diagram sources are scripts 02 and 03; script 04 renders them.

### Technical details

The first board is `[0,4,7,5,2,6,1,3]`, packed as `672be0`. All new scripts and experiment outputs are inside this ticket. The user's publication request authorizes pushing this checkpoint despite the incomplete guide.
