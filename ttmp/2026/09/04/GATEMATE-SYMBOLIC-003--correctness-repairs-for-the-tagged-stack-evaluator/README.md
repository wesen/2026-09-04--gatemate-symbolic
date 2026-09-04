# Correctness repairs for the tagged stack evaluator

This is the completed implementation workspace for ticket GATEMATE-SYMBOLIC-003.
Start with [the ticket index](index.md), then read the repair design and detailed
diary. Final validation records 198 passing tests and four matching board UART
captures. The original failing synthesis evidence is preserved separately.

## Structure

- **design-doc/**: Repair design, implementation references, and synthesis diagnosis
- **reference/**: Reference documentation and API contracts
- **playbooks/**: Operational playbooks and procedures
- **scripts/**: Utility scripts and automation
- **sources/**: External sources and imported documents
- **various/**: Scratch or meeting notes, working notes
- **archive/**: Optional space for deprecated or reference-only artifacts

## Getting Started

Use docmgr commands to manage this workspace:

- Add documents: `docmgr doc add --ticket GATEMATE-SYMBOLIC-003 --doc-type design-doc --title "My Design"`
- Import sources: `docmgr import file --ticket GATEMATE-SYMBOLIC-003 --file /path/to/doc.md`
- Update metadata: `docmgr meta update --ticket GATEMATE-SYMBOLIC-003 --field Status --value review`
