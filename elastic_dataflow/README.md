# Elastic dataflow laboratory

A fixed seven-node expression graph runs across four contexts on a GateMate FPGA. Typed 40-bit values travel in 80-bit envelopes through bounded queues, operand matching, arithmetic pipelines, and a completion router. Context epochs support cancellation without withdrawing an already offered output.

The Go/React IDE edits execution scenarios and inspects actual paused FPGA state or an explicitly labeled software model.

```sh
make frontend
go run -tags embed ./cmd/dataflow-ide --engine model
```

Run a web server in tmux for interactive work. The IDE defaults to `http://127.0.0.1:8087/`. To use the programmed FPGA, select `--engine serial --device /dev/ttyACM0`. Only one process should own the UART. Stop an existing web server with `lsof-who -p 8087 -k` before replacing it.

```sh
go run ./cmd/dataflow-lab --engine model --example book --format json
```

The checked examples are `book`, `copy`, `fault`, and `cancel`. Each explicitly resets its selected engine.

- [Intern guide](../ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine/design-doc/01-elastic-dataflow-engine-intern-analysis-design-and-implementation-guide.md)
- [Separate IDE design](../ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine/design-doc/02-dataflow-ide-go-react-architecture-and-intern-implementation-guide.md)
- [API and register reference](../ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine/reference/02-dataflow-api-and-debug-register-reference.md)
- [Implemented handoff, measurements, and screenshots](../ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine/reference/03-implemented-laboratory-handoff-and-screenshot-atlas.md)
- [Detailed implementation diary](../ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine/reference/01-implementation-diary.md)

Load the installed CAD environment before running `scripts/test.sh`, `scripts/test-link.sh`, or `scripts/test-stress.sh`. `scripts/build-board.sh` loads that environment and runs the synthesis/place/route/pack flow. The qualified board build uses queue depths eight, multiply latency four, and the 10 MHz clock constraint.
