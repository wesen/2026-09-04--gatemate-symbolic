# Laboratory 4: lazy graph reducer

The reducer evaluates a heap graph through explicit continuations. A shared
thunk is claimed as BLACKHOLE, evaluated once, and replaced by its value or
memoized error. The shared example returns 168 with one multiplication; the
self-demand example returns CYCLIC_THUNK and leaves no permanent claim.

Implemented: a 1024x40 heap, 512x80 continuation stack, signed checked int32
ADD/MUL, bounded indirection traversal, stable held output, 64 mutation records
with loss count, Go reference/model/serial engines, and a React heap inspector.
This is one evaluator. Reset aborts and discards the image; there is no live
cancellation, allocator, garbage collector or multi-evaluator extension.

```sh
go run ./cmd/lazy-lab --example shared --format json
make lazy-frontend
# Start servers in tmux; use an available loopback port.
tmux new-session -d -s lazy-model \
  'go run -tags embed ./cmd/lazy-ide --engine model --listen 127.0.0.1:18089'
```

Open http://127.0.0.1:18089/, load the shared graph, Force root, and Step three
model cycles to inspect the live claim. Advance 1000 cycles, Poll, then Force
again to verify that the multiplication count remains one. The model and RTL
have different cycle schedules; snapshots always identify their source.

```sh
source /home/manuel/fpga/oss-cad-suite/environment
bash lazy_reducer/scripts/test.sh
make -C lazy_reducer bit
make -C lazy_reducer load
go run ./cmd/lazy-lab --engine serial --device /dev/ttyACM0 --example shared --format json
tmux new-session -d -s lazy-fpga \
  'go run -tags embed ./cmd/lazy-ide --engine serial --device /dev/ttyACM0 --listen 127.0.0.1:18090'
```

Only one process may own UART. Before replacing a web server use
`lsof-who -p PORT -k`. The lazy image has a distinct LAZY capability signature;
the dataflow image cannot serve this protocol. Board SRAM programming is volatile.

Validation: 400 Go generated graph comparisons, 120 RTL graph comparisons,
directed UART tests, Go race/build/vet/Glazed/vulnerability checks, seventeen
frontend tests and model browser checks. Final routed timing passed at
24.65 MHz for the 10 MHz constraint. Physical SRAM programming and qualification passed after reconnection: 60
randomized graphs, five directed examples, live claims, held/repeated results,
nested updates, trace overflow and 512-frame overflow unwinding. Physical
browser checks passed with five screenshots and no console errors or warnings.

The [ticket](../ttmp/2026/09/05/GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector/index.md)
contains the intern guide, implemented API reference, detailed diary, scripts,
test evidence and screenshots. The prepared physical test is:

```sh
go test ./pkg/lazy -run TestPhysicalLazyQualification -count=1 -v \
  -lazy-physical-device /dev/ttyACM0
```
