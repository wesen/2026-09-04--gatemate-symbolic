# Source provenance

Collected 2026-09-04 using `scripts/03-collect-sources.sh` and `defuddle parse URL --md -o FILE`. These are reference copies, not substitutes for the implementation's actual tool versions. The initial unversioned Yosys URL returned HTTP 404; the versioned documentation below succeeded.

| Local file | Primary source | Role |
|---|---|---|
| yosys-synth-gatemate.md | https://yosyshq.readthedocs.io/projects/yosys/en/v0.50/cmd/synth_gatemate.html | GateMate-specific synthesis stages and mapping commands |
| olimex-gatematea1-evb.md | https://www.olimex.com/Products/FPGA/GateMate/GateMateA1-EVB/open-source-hardware | Board identity and official hardware documentation links |

The original book is already stored in the adjacent implementation ticket: `../../GATEMATE-SYMBOLIC-001--symbolic-computer-patterns-and-designs-for-gatemate-analysis-design-and-intern-implementation-guide/sources/Composable_Hardware_Patterns_for_Symbolic_Computers.md`. Its Laboratory 1 section is the local design source; it was not downloaded again.

Fresh local test and synthesis outputs are under `../reference/validation/`. Historical hardware observations are in the original ticket diary, Steps 10 and 12. No fresh board capture or routed timing measurement was performed in this review.
