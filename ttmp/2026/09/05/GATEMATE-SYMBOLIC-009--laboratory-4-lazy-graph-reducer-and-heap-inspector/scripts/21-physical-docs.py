from pathlib import Path
root=Path(__file__).resolve().parents[1]
p=root/'reference/02-implemented-reducer-api-and-qualification-handoff.md'
s=p.read_text()
start=s.index('Physical qualification is pending.')
end=s.index('\n',start)
s=s[:start]+'Physical qualification passed on 2026-09-05 after the board was reconnected. SRAM programming succeeded, all directed and randomized physical tests passed, and the browser was exercised against the serial engine with no console errors or warnings. The earlier disconnected-device failure remains in the diary and p5-program-board-disconnected.log.'+s[end:]
s=s.replace('## Reproduction and remaining physical phase','## Reproduction and physical qualification')
s=s.replace('Those steps remain unexecuted until device access is restored.','These steps passed after reconnection; p5-physical-tests.log records the 15.108-second successful suite and p5-browser-fpga.json records physical browser observations.')
s += """
## Observations from the programmed FPGA

The physical inspector connects to the programmed GateMate through the Go serial engine at http://127.0.0.1:18090/. Its source badge reads PHYSICAL FPGA. These observations come from UART query pages, with execution paused during each snapshot. The model and hardware implement the same semantics but use different cycle schedules.

![Physical live claim and continuation stack](screenshots/fpga-claimed.png)

At enabled cycle 9, address 3 has changed from THUNK(body=2) to BLACKHOLE. The current address is 2. The stack has three frames: two EVAL_RIGHT continuations beneath UPDATE(3). Counters report three dispatched heap reads, one write and one claim. This observation verifies that the update obligation exists while the body is being evaluated.

![Physical shared result and mutation trace](screenshots/fpga-result.png)

The shared expression returns INT(168). Address 3 now holds INT(42), and counters report exactly one multiplication, one claim, one update and three additions. The trace records the claim at cycle 9 and replacement at cycle 61. The screenshot was captured at cycle 1009, including 917 output-stall cycles from the requested tick batch; the displayed total is not an execution-latency measurement. Polling and forcing root 6 again preserves the multiplication count of one, demonstrating reuse of the stored value.

![Physical cycle detection and memoized error](screenshots/fpga-cycle.png)

The recursive example returns ERROR(CYCLIC_THUNK), numeric fault code 3. Address 1 changes from THUNK to BLACKHOLE at cycle 3 and from BLACKHOLE to ERROR at cycle 23. The final counters show one claim, one update, one blackhole observation and one generated fault. The held result is the same ERROR word stored in the heap, and the continuation stack has been unwound.

![Historical physical observation](screenshots/fpga-history.png)

Selecting the earlier physical frame restores that observation in the browser and disables mutation controls. It does not rewind the FPGA. The browser also rejects malformed source JSON without creating a new frame or mutating the device.

![Physical inspector on a narrow viewport](screenshots/fpga-mobile.png)

At a 390-pixel viewport, document scroll width and client width are both 390 pixels. Wide graph and table content remains within its own scrolling region. Five physical screenshots complement the five model screenshots retained above.

The physical suite compared 60 randomized graph results and final heaps with the independent recursive reference, then exercised five directed examples, a live claim and stack, stable held results, repeated forcing, 79 nested thunk updates, mutation-trace overflow, and recovery from a 512-frame stack overflow. All passed against the routed image. These tests establish the tested bounds and examples; they are not a formal proof for every possible heap.
"""
p.write_text(s)
(root/'sources/handoff-body.md').write_text(s.split('---',2)[2].lstrip())
p=root/'design-doc/01-lazy-graph-reducer-intern-analysis-design-and-implementation-guide.md'
s=p.read_text().replace('Physical qualification remains pending because the board was disconnected at programming time.','Physical qualification subsequently passed after reconnection: 60 randomized graphs, five directed examples, bounds and memoization checks, and the physical browser workflow. The handoff includes physical screenshots and measured mutation cycles.')
p.write_text(s)
p=root/'index.md'
s=p.read_text()
a=s.index('P1–P4 are complete.')
b=s.index('\n',a)
s=s[:a]+'P1–P5 are complete. Model, RTL, host and browser validation passed. The GateMate was programmed successfully after reconnection; 60 randomized physical graphs, five directed examples, bounds tests and physical browser checks passed. Final routed timing is 24.65 MHz at the required 10 MHz. The physical inspector is available at http://127.0.0.1:18090/ while its tmux server and board remain connected.'+s[b:]
p.write_text(s)
p=Path('lazy_reducer/README.md')
s=p.read_text().replace('**Physical programming and qualification\nare pending: the board was disconnected at the first programming attempt.**','Physical SRAM programming and qualification passed after reconnection: 60\nrandomized graphs, five directed examples, live claims, held/repeated results,\nnested updates, trace overflow and 512-frame overflow unwinding. Physical\nbrowser checks passed with five screenshots and no console errors or warnings.')
p.write_text(s)
p=root/'scripts/19-audit.py'
s=p.read_text().replace("'physical_qualification':'pending device connection'","'physical_qualification':'passed'")
s=s.replace("images=list(", """assert 'OK: uploaded' in (root/'reference/validation/p5-physical-handoff-upload.log').read_text()
assert '--- PASS: TestPhysicalLazyQualification' in (root/'reference/validation/p5-physical-tests.log').read_text()
assert 'Errors: 0, Warnings: 0' in (root/'reference/validation/p5-fpga-console.log').read_text()
assert (root/'reference/validation/p5-browser-fpga.json').is_file()
images=list(""")
p.write_text(s)
