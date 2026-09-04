# Graph search microscope

This laboratory runs configurable graph coloring on a GateMate FPGA and exposes every semantic search event through a Go service and a React inspector. It supports 1–8 vertices, 1–8 colors, undirected edges, complete enumeration, and first-solution termination. The interface displays candidate domains, propagated vertices, checkpoints, mutation trail, and the last 256 event snapshots.

The Go model is a separate, explicitly labeled simulator. Physical execution requires the graph bitstream and a serial connection to the board.

## Start the application

Run these commands from the repository root. Dependencies are Go 1.26.8 or newer and pnpm 10.15.1. Go's normal toolchain selection can download the required patch release.

```sh
make frontend
go run -tags embed ./cmd/search-microscope --engine model --listen 127.0.0.1:8086
```

Open <http://127.0.0.1:8086>. Load a preset before stepping. `Triangle · 3 colors` produces six labeled colorings; `Triangle · 2 colors` is unsatisfiable; `Path · 2 colors` produces two colorings. Stop the model server before starting a physical session on the same port.

To program the Olimex GateMate evaluation board, use the installed OSS CAD suite and openFPGALoader:

```sh
source /home/manuel/fpga/oss-cad-suite/environment
make -C graph_microscope load
go run -tags embed ./cmd/search-microscope --engine serial --device /dev/ttyACM0 --listen 127.0.0.1:8086 --log-level info
```

Use tmux for persistent sessions. Only one process may own the UART. On this machine `/dev/ttyACM0` is the application UART and `/dev/ttyACM1` is the JTAG adapter; verify device assignment if reconnecting hardware. The service identifies its engine at the top of the page.

`Step event` accepts one held semantic event. `Run` repeatedly performs the same operation at a paced interval. `Pause` waits for any current exchange to finish. `Reset graph` restarts the accepted configuration and clears history. Selecting a historical event disables device controls until `Return to live`; it does not reverse hardware execution. If a serial exchange fails, use explicit reset or reload to resynchronize.

## Development and verification

```sh
go test -race ./... -count=1
pnpm --dir web test
pnpm --dir web check
make lint
make govulncheck
go generate ./internal/microscope
go build -tags embed ./...
```

Source the CAD environment before Go tests to include Icarus UART/RTL simulation; those tests explicitly skip when the tools are unavailable. Physical tests are opt-in and require stopping the service first:

```sh
go test ./pkg/microscope -run TestPhysicalGraphEvents -count=1 -v -args -hardware-device /dev/ttyACM0 -hardware-out /tmp/graph-hardware
```

For frontend development, run the Go service in one tmux session and `pnpm --dir web dev` in another. Vite proxies `/api/` to port 8086. The default Go build reads generated frontend assets from the repository; the `embed` build includes HTML, JS, and CSS in the binary. Generated assets are ignored and recreated by `make frontend`.

## Source map

| Layer | Entry points |
|---|---|
| Graph semantics and software model | `pkg/microscope/graph.go`, `model.go` |
| UART framing and device ownership | `pkg/microscope/protocol.go`, `serial.go` |
| Checked event reconstruction | `pkg/microscope/projection.go` |
| HTTP and bounded session history | `internal/microscope/http.go`, `session.go` |
| Server command | `cmd/search-microscope/main.go` |
| React view and API cache | `web/src/App.tsx`, `store.ts`, `GraphView.tsx` |
| FPGA solver and UART commands | `graph_microscope/rtl/graph_core.sv`, `graph_link.sv`, `uart_rx.sv`, `graph_top.sv` |
| Wire-level simulation | `graph_microscope/sim/tb_graph_link.sv`, `pkg/microscope/rtl_test.go` |

The detailed intern guide, implementation diary, measured hardware results, screenshots, and delivery receipts are in [ticket GATEMATE-SYMBOLIC-006](../ttmp/2026/09/04/GATEMATE-SYMBOLIC-006--graph-coloring-search-microscope-with-go-react-and-fpga-event-stepping/index.md).

## API examples

The server accepts strict JSON bodies up to 4096 bytes. Graph mutations require a paused session. Historical requests require the generation returned by `/api/state`.

```sh
curl -s http://127.0.0.1:8086/api/state
curl -s http://127.0.0.1:8086/api/graph -H 'Content-Type: application/json' \
  -d '{"vertices":3,"colors":3,"edges":[[0,1],[1,2],[0,2]],"firstOnly":false}'
curl -s http://127.0.0.1:8086/api/control -H 'Content-Type: application/json' \
  -d '{"action":"step"}'
# Substitute the actual generation from the response:
curl -s 'http://127.0.0.1:8086/api/events/1?generation=1'
```

Errors contain an `error` string: 400 denotes invalid input, 409 denotes a session/generation conflict, 404 denotes unavailable history or an unknown endpoint, and 502 denotes an engine failure. Read-only requests never advance the solver. The loopback service is intended for one local operator and one device.
