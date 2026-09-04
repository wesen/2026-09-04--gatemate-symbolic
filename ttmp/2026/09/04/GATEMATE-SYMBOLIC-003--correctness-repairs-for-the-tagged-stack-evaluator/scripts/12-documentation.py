#!/usr/bin/env python3
"""Align comments and operational guides with the repaired contract."""
from pathlib import Path
ROOT=next(p for p in Path(__file__).resolve().parents if (p/'symbolic_eval').is_dir())/'symbolic_eval'
def edit(path,old,new):
    p=ROOT/path;s=p.read_text();assert old in s,(path,old);p.write_text(s.replace(old,new))
edit('tools/opcodes.py','model (stack_model.py), the RTL package header generator, and the tests all\nimport this module. Keeping one table prevents the failure mode where four\nhand-copied tables gradually diverge.', 'model (stack_model.py), and the tests import this module. Handwritten RTL\nconstants and testbench name tables are checked by scripts/check_isa.py.')
for path in ['tools/opcodes.py','tools/stack_model.py']:
    edit(path,'<pc, stack, output_stream, fault, halted>','<pc, stack, rstack, output_stream, fault, halted>')
edit('rtl/stack_core.sv','everything and stages complete next-state data; COMMIT is the single\n//     mutation owner and the only place registers and the stack change.', 'everything and stages complete next-state data. Architectural PC/stacks\n//     change only at COMMIT or EMIT acceptance; internal registers may change\n//     during preparation. Fault publication preserves both stacks and PC.')
edit('rtl/top.sv','// it async-asserts reset; releasing it synchronously restarts the machine\n// from the initial image.', '// it passes through a two-flop synchronizer before asserting reset; release\n// synchronously restarts the machine from the initial image.')
edit('rtl/top.sv','// holds the item stable while bytes are framed (Delayed Irreversible\n  // Store boundary on real hardware).','// transfers ownership to vval_q while bytes are framed. The elastic stage\n  // can then accept another value.')
edit('rtl/uart_tx.sv','// `ready`. The io_block derives `start` from a single acceptance edge, so a\n// held io_req never re-triggers a byte.', '// `ready`. The value printer asserts start only while the transmitter is\n// ready; busy suppresses additional starts until the frame completes.')
edit('rtl/sync_sdp_ram.sv','// design they never target the same address in the same cycle)', '// design incidental same-address read results during a spill are unused)')
edit('constraints/olimex_gatematea1_evb.ccf','# -> IO_SB_B7 (active-LOW, the _n suffix). top.sv inverts so but_sync=1 means\n# pressed.', '# -> IO_SB_B7 (active-LOW, the _n suffix). top.sv preserves active-low\n# polarity through the button synchronizer.')
edit('Makefile','build/$(PROG).{hex,bin,lst,sym.json}','build/$(PROG).{hex,lst,sym.json}')
edit('playbooks/PB-01-adding-an-opcode.md','stage complete next-state data (`npc_d`, `ndepth_d`, write descriptors).','stage complete next-state data (`npc_d`, `ndepth_d`, `nrdepth_d`, write descriptors).')
edit('playbooks/PB-01-adding-an-opcode.md','registers only change in `COMMIT`','architectural PC and stacks only change in `COMMIT`')
edit('playbooks/PB-01-adding-an-opcode.md','sync with the table; there is no generator).','sync with the table; `scripts/check_isa.py` verifies these handwritten mirrors).')
edit('playbooks/PB-03-adding-a-fault-code.md','adding codes 10+ means\nwidening again','adding code 16 or higher requires\nwidening')
edit('playbooks/PB-04-board-workflow.md','make test                  # 123 tests: run these first, always','make test                  # 197 tests, including complete state comparison')
edit('playbooks/PB-04-board-workflow.md','make load                  # openFPGALoader over the RP2040 DirtyJTAG bridge','make load PROG=fib         # keep PROG: load rebuilds its selected image')
edit('playbooks/PB-04-board-workflow.md','async reset, release = synchronous restart from the initial image.', 'synchronized assertion, release = synchronous restart from the initial image.')
edit('playbooks/PB-02-debugging-a-trace-mismatch.md','| Output values differ, trace identical | `EMIT` pending register or printer path, not the core |','| UART bytes differ but complete core state/transfer checks pass | inspect printer and UART framing |')
p=ROOT/'playbooks/PB-02-debugging-a-trace-mismatch.md';p.write_text(p.read_text()+'''

## Complete-state checks

Both core runners compare `STATE` records (all live 40-bit data words and return
addresses), `XFER` records (accepted output including flags), and TRACE/FINAL.
The benches assert live state is unchanged between retirements and on a fault.
An empty DUP can preserve counts while being semantically wrong; use the full
snapshot diff to find such errors. `trace_fetch` formats a fetch failure as FETCH;
its opcode field is not an instruction in that case. Fault addresses can equal
ROM_DEPTH, while physical memory addresses remain in range.

`build/verification-coverage.json` records exercised opcodes, faults, branches,
and cache-count transitions for the most recent pytest run. The counters measure
executed tests; they are not a formal coverage proof.
''')
p=ROOT/'playbooks/PB-03-adding-a-fault-code.md';p.write_text(p.read_text()+'''

Fetch faults are published from FETCH (or the BRAM fetch-context states), with
`trace_fetch`/`fault_fetch` asserted. Keep fault PC wide enough for ROM_DEPTH.
For tc=1 and dc>0 the second operand tag comes from a synchronous deep-RAM read;
never use stale opnd_q without acquiring context. Constructed-state RTL cases in
`test_verification.py` cover noncanonical Booleans and arithmetic extremes.
''')
