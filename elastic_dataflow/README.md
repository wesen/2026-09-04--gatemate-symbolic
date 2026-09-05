# Programmable elastic dataflow workbench

Compile typed expressions into a graph of up to seven nodes and execute it across four contexts on a GateMate FPGA or the Go transaction model. Typed 40-bit values travel in 80-bit tokens through operand RAM, bounded queues, elastic arithmetic units, and an explicit completion router. Context epochs cancel obsolete work without withdrawing an already offered result.

The React workbench provides a typed editor, compiler diagnostics, named inputs, a graph derived from actual engine descriptors, operand/unit inspectors, hardware breakpoints, and a bounded event trace with visible loss accounting. Historical frames are observation records, not reverse execution.

```sh
make dataflow-frontend
tmux new-session -d -s dataflow-model \
  'go run -tags embed ./cmd/dataflow-ide --engine model --listen 127.0.0.1:8087'
```

Open `http://127.0.0.1:8087/`. Compile and load the supplied program, supply named values, then advance cycles and poll the result. With `a=3`, `b=4`, `c=5`, it returns 30:

```text
input a, b, c: int16
let square = a * a
let offset = b * c
let selected = int(a < b)
output square + offset + selected
```

Select node zero, apply the issue breakpoint, and advance 100 cycles. Execution stops after the square issues; the trace contains value 9. Resume clears the stop but does not advance the engine. A subsequent tick request continues computation. The trace stores at most 32 records, with explicit dropped-event counts for full storage or simultaneous physical events.

Save/import/export use `.df` source files. Low-level scenario experiments remain available in a separate expandable editor. The reset program is the original laboratory expression. Loading a program resets the engine, stages descriptors, activates a validated image, and verifies descriptor readback before associating source names.

For the physical board, use the version-two bitstream and `--engine serial --device /dev/ttyACM0`. One process must own the UART. Stop an existing server with `lsof-who -p 8087 -k` before replacing it. Startup resets the selected engine. Model and FPGA agree on checked computation, without identical cycle timing or trace capture bandwidth.

The standalone checked laboratory examples still use the built-in reset graph:

```sh
go run ./cmd/dataflow-lab --engine model --example book --format json
```

Examples are `book`, `copy`, `fault`, and `cancel`. Each resets the selected engine.

- [Programmable workbench intern guide](../ttmp/2026/09/04/GATEMATE-SYMBOLIC-008--programmable-dataflow-workbench-compiler-and-physical-debugger/design-doc/01-programmable-dataflow-workbench-intern-analysis-design-and-implementation-guide.md)
- [Implemented APIs, trace semantics, and qualification](../ttmp/2026/09/04/GATEMATE-SYMBOLIC-008--programmable-dataflow-workbench-compiler-and-physical-debugger/reference/02-implemented-programmable-workbench-api-and-physical-qualification-reference.md)
- [Workbench implementation diary](../ttmp/2026/09/04/GATEMATE-SYMBOLIC-008--programmable-dataflow-workbench-compiler-and-physical-debugger/reference/01-implementation-diary.md)
- [Original engine intern guide](../ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine/design-doc/01-elastic-dataflow-engine-intern-analysis-design-and-implementation-guide.md)

Load the installed OSS CAD Suite environment before running `scripts/test.sh`, `scripts/test-link.sh`, `scripts/test-programmable.sh`, `scripts/test-debug.sh`, or `scripts/test-stress.sh`. `scripts/build-board.sh` loads that environment and performs synthesis, placement, routing and packing. The physical target uses depth-eight queues, multiplier latency four, eight-bit epochs and a 10 MHz clock constraint. Consult the ticket for the actual routed timing result.
