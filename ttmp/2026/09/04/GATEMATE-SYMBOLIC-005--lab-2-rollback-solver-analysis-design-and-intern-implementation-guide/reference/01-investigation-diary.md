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

## Step 2: Resume, recover the complete guide, and validate PDF delivery

The user resumed normal work after retracting an unrelated formatting instruction. I corrected recovery to select only the exact target document patch, recovered the complete 7,261-word draft, and reviewed it against the saved experiment. The guide now exists as a real deliverable, with two rendered diagrams, proposed APIs, timing tables, implementation phases, and explicit evidence limits.

### Prompt Context

**User prompt (verbatim):** "sorry this was actually unrelated (the color stuff), forget abou tit and continue normally"

**Assistant interpretation:** Ignore the unrelated formatting preference and resume the unfinished Lab 2 design and reMarkable delivery.

**Inferred user intent:** Complete the originally requested technical guide and deliver it normally.

### What I did

Required the exact Add File target when selecting the archived authored patch. Added a delivery preparation script that validates metadata, experiment results, fenced blocks, and image paths. Rendered a local PDF preview with pandoc and inspected page four, including the architecture figure and packed record tables. Related five primary sources and ran docmgr doctor successfully.

### Why

The previous recovery selected another tool call containing the same text marker. Matching the target path fixes selection without rewriting the authored analysis or discarding its details.

### What worked

Recovery succeeded on the first resumed repair attempt. The document checks passed, the local preview rendered to 16 pages without warnings, and the inspected diagram/table page was legible. The reMarkable dry-run succeeded. The upload input uses absolute asset paths and a conventional lowercase title field so temporary bundle rendering can resolve the images and title.

### What didn't work

No resumed recovery or PDF-render failure. Historical errors remain in Step 1. The reMarkable upload is recorded separately after its command completes.

### What I learned

Exact artifact identity is a better recovery discriminator than a phrase that can recur in diagnostic source code. Render preparation can preserve relative image links in the source while giving the uploader absolute paths.

### What was tricky to build

The guide distinguishes precise protected-write faults from search contradictions, bitmap restoration from individual trail pops, and semantic events from memory cycles. The software experiment is explicitly not presented as synchronous-memory or hardware cut verification.

### What warrants a second pair of eyes

Review the proposed conservative 64-entry trail bound and 8-entry choice bound, reserved-entry publication sequence, result acceptance/cut timing, and baseline-versus-trail traffic accounting.

### What should be done in the future

Implement the design under project ticket GATEMATE-SYMBOLIC-004 after review. This documentation ticket does not implement production Lab 2 RTL.

### Code review instructions

Read the complete design document and saved experiment JSON. Run scripts/06-prepare-delivery.py and docmgr doctor for this ticket. Inspect the PDF preview or delivered reMarkable document for the diagrams and tables.

### Technical details

The local preview is /tmp/gatemate005-delivery/preview.pdf. Reproducible source files and validation logs remain in the ticket. The intended remote folder is /ai/2026/09/04/GATEMATE-SYMBOLIC-005.

## Step 3: Confirm reMarkable delivery and publish the completed ticket

The completed guide uploaded successfully, including its two rendered diagrams. The documentation ticket is now ready for review with every task complete; the separate Lab 2 project ticket continues to track implementation work.

### Prompt Context

**User prompt (verbatim):** (see Step 2; delivery continues the original guide/upload request).

**Assistant interpretation:** Finish delivery and publish the completed documentation after validation.

**Inferred user intent:** Receive the guide on reMarkable and retain a reviewable repository record.

**Commit (guide):** `da4ea92` — `docs: complete Lab 2 design and intern implementation guide`.

### What I did

Ran the prepared bundle upload, read its successful receipt, checked the final documentation task, updated the ticket to review, and preserved the validation and upload logs for the final commit and push.

### Why

The earlier checkpoint was intentionally incomplete. Delivery status must now reflect the successful guide upload without erasing the historical failure record.

### What worked

The command exited zero and reported `OK: uploaded GATEMATE-SYMBOLIC-005 Lab 2 Design and Intern Guide.pdf -> /ai/2026/09/04/GATEMATE-SYMBOLIC-005`. Doctor passed and both diagrams were included by absolute image paths in the render input.

### What didn't work

No upload failure or authentication retry occurred.

### What I learned

The validated source and prepared rendering input can retain different frontmatter while preserving the same guide body and figures.

### What was tricky to build

The delivery copy uses conventional lowercase title metadata, while the source document retains docmgr frontmatter and related-file metadata. The recorded upload receipt is the delivery confirmation; no redundant cloud listing is needed.

### What warrants a second pair of eyes

Review the proposed design choices and evidence limits before implementing Lab 2. The uploaded document is a design guide, not a claim that production solver RTL exists.

### What should be done in the future

Proceed with the implementation phases in GATEMATE-SYMBOLIC-004 after design review.

### Code review instructions

Read the design guide, document-check.json, doctor.log, and upload.log. The earlier interrupted-checkpoint diary entry remains historical.

### Technical details

The delivery script is scripts/07-upload-guide.sh. Remote destination: /ai/2026/09/04/GATEMATE-SYMBOLIC-005. The successful receipt is stored in reference/validation/upload.log.
