# LFL1 lazy language

LFL1 compiles annotated functional source to a finite lazy machine implemented in Go and on GateMate. The source-aware React IDE exposes code, heap objects, captured environments, continuation frames, allocation progress, mutation traces and incremental list demands.

```sh
go run ./cmd/lazy-language --source examples/lazylang/shared.lazy --action compile
go run ./cmd/lazy-language --source examples/lazylang/shared.lazy --action reference
go run ./cmd/lazy-language --source examples/lazylang/squares.lazy --action run

go generate ./internal/lazylanguageide
go run ./cmd/lazy-language-ide --listen 127.0.0.1:18091 --engine model
```

Run servers in tmux. To use the programmed LFL1 board, start the IDE with `--engine serial --device /dev/ttyACM0`; only one process may own the UART. The serial IDE opens the device on startup and changes its loaded program only after an explicit load/reset operation. Production assets are embedded with `go build -tags embed ./...` after generation.

The pipeline is `syntax.Parse` → `check.Check` → `compile.Compile` → `machine.New` or `serial.Client.Load`. `semantics.Observe` independently evaluates checked source for semantic comparison. All APIs accept context where evaluation or transport may take time; the in-memory machine itself exposes bounded `Tick`, `Demand`, `Poll` and detached `Snapshot` operations.

The profile contains 2048 heap objects, 512 continuation frames and a first-64 mutation trace. There is no garbage collection. Source types are monomorphic Int, Bool, ListInt and functions. A stream prefix never forces the final tail just to determine whether another element exists. Paused observations remain distinct from runtime faults.

The physical implementation routes at 21.26 MHz against a 10 MHz constraint using 34 RAM halves, 14,359 CPE logic resources and 3,804 flip-flops. Six physical programs match model heaps, provenance, semantic counters and trace mutations. The Go model is not cycle-exact: hardware performs additional synchronous reads.

Validation:

```sh
go test ./pkg/lazylang/... ./internal/lazylanguageide
bash lazy_language/scripts/test.sh
bash lazy_language/scripts/build.sh
# Explicit physical opt-in, with no serial IDE running:
go test ./pkg/lazylang/serial -run TestPhysicalPrograms -count=1 -v -lfl-device /dev/ttyACM0
```

Read the [implemented intern handoff](../../ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide/reference/04-implemented-language-runtime-and-ide-intern-handoff.md) for the architecture, binary formats, APIs, physical results and screenshot index. The [implementation diary](../../ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide/reference/01-implementation-diary.md) records exact failures, fixes, commits and validation receipts.
