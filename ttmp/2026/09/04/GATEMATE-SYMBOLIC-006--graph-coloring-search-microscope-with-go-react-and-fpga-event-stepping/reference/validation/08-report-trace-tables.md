---
Title: Decoded Physical Graph Trace Tables
Ticket: GATEMATE-SYMBOLIC-006
Status: complete
Topics: [fpga, gatemate]
DocType: reference
Intent: long-term
Summary: Generated report evidence from archived physical UART frames.
---

## triangle

| Seq | Event | Domains | P | C | T | Base | Count |
|---:|---|---|---|---:|---:|---:|---:|
| 1 | CREATE | 07 07 07 | F8 | 1 | 0 | 0 | 0 |
| 2 | WRITE | 01 07 07 | F8 | 1 | 1 | 0 | 0 |
| 3 | WRITE | 01 06 07 | F8 | 1 | 2 | 0 | 0 |
| 4 | WRITE | 01 06 06 | F8 | 1 | 3 | 0 | 0 |
| 5 | PROPAGATED | 01 06 06 | F9 | 1 | 3 | 0 | 0 |
| 6 | CREATE | 01 06 06 | F9 | 2 | 3 | 0 | 0 |
| 7 | WRITE | 01 02 06 | F9 | 2 | 4 | 0 | 0 |
| 8 | WRITE | 01 02 04 | F9 | 2 | 5 | 0 | 0 |
| 9 | PROPAGATED | 01 02 04 | FB | 2 | 5 | 0 | 0 |
| 10 | PROPAGATED | 01 02 04 | FF | 2 | 5 | 0 | 0 |
| 11 | OUTPUT | 01 02 04 | FF | 2 | 5 | 0 | 1 |
| 12 | RESTORE | 01 02 06 | FF | 2 | 4 | 0 | 1 |
| 13 | RESTORE | 01 06 06 | FF | 2 | 3 | 0 | 1 |
| 14 | RESTORED | 01 06 06 | F9 | 2 | 3 | 0 | 1 |
| 15 | UPDATE | 01 06 06 | F9 | 2 | 3 | 0 | 1 |
| 16 | WRITE | 01 04 06 | F9 | 2 | 4 | 0 | 1 |
| 17 | WRITE | 01 04 02 | F9 | 2 | 5 | 0 | 1 |
| 18 | PROPAGATED | 01 04 02 | FB | 2 | 5 | 0 | 1 |
| 19 | PROPAGATED | 01 04 02 | FF | 2 | 5 | 0 | 1 |
| 20 | OUTPUT | 01 04 02 | FF | 2 | 5 | 0 | 2 |
| 21 | RESTORE | 01 04 06 | FF | 2 | 4 | 0 | 2 |
| 22 | RESTORE | 01 06 06 | FF | 2 | 3 | 0 | 2 |
| 23 | RESTORED | 01 06 06 | F9 | 2 | 3 | 0 | 2 |
| 24 | POP | 01 06 06 | F9 | 1 | 3 | 0 | 2 |

## unsatisfiable

| Seq | Event | Domains | P | C | T | Base | Count |
|---:|---|---|---|---:|---:|---:|---:|
| 1 | CREATE | 03 03 03 | F8 | 1 | 0 | 0 | 0 |
| 2 | WRITE | 01 03 03 | F8 | 1 | 1 | 0 | 0 |
| 3 | WRITE | 01 02 03 | F8 | 1 | 2 | 0 | 0 |
| 4 | WRITE | 01 02 02 | F8 | 1 | 3 | 0 | 0 |
| 5 | PROPAGATED | 01 02 02 | F9 | 1 | 3 | 0 | 0 |
| 6 | WRITE | 01 02 00 | F9 | 1 | 4 | 0 | 0 |
| 7 | CONTRADICTION | 01 02 00 | F9 | 1 | 4 | 0 | 0 |
| 8 | RESTORE | 01 02 02 | F9 | 1 | 3 | 0 | 0 |
| 9 | RESTORE | 01 02 03 | F9 | 1 | 2 | 0 | 0 |
| 10 | RESTORE | 01 03 03 | F9 | 1 | 1 | 0 | 0 |
| 11 | RESTORE | 03 03 03 | F9 | 1 | 0 | 0 | 0 |
| 12 | RESTORED | 03 03 03 | F8 | 1 | 0 | 0 | 0 |
| 13 | UPDATE | 03 03 03 | F8 | 1 | 0 | 0 | 0 |
| 14 | WRITE | 02 03 03 | F8 | 1 | 1 | 0 | 0 |

## root

| Seq | Event | Domains | P | C | T | Base | Count |
|---:|---|---|---|---:|---:|---:|---:|
| 1 | WRITE | 01 00 | FC | 0 | 1 | 0 | 0 |
| 2 | CONTRADICTION | 01 00 | FC | 0 | 1 | 0 | 0 |
| 3 | COMPLETE | 01 00 | FC | 0 | 1 | 0 | 0 |

## first

| Seq | Event | Domains | P | C | T | Base | Count |
|---:|---|---|---|---:|---:|---:|---:|
| 1 | CREATE | 07 07 07 | F8 | 1 | 0 | 0 | 0 |
| 2 | WRITE | 01 07 07 | F8 | 1 | 1 | 0 | 0 |
| 3 | WRITE | 01 06 07 | F8 | 1 | 2 | 0 | 0 |
| 4 | WRITE | 01 06 06 | F8 | 1 | 3 | 0 | 0 |
| 5 | PROPAGATED | 01 06 06 | F9 | 1 | 3 | 0 | 0 |
| 6 | CREATE | 01 06 06 | F9 | 2 | 3 | 0 | 0 |
| 7 | WRITE | 01 02 06 | F9 | 2 | 4 | 0 | 0 |
| 8 | WRITE | 01 02 04 | F9 | 2 | 5 | 0 | 0 |
| 9 | PROPAGATED | 01 02 04 | FB | 2 | 5 | 0 | 0 |
| 10 | PROPAGATED | 01 02 04 | FF | 2 | 5 | 0 | 0 |
| 11 | OUTPUT | 01 02 04 | FF | 0 | 5 | 5 | 1 |
| 12 | COMPLETE | 01 02 04 | FF | 0 | 5 | 5 | 1 |
