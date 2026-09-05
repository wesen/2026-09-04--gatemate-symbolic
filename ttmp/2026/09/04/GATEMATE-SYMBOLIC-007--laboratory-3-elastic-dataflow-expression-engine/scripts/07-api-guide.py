#!/usr/bin/env python3
from pathlib import Path
root=Path(__file__).resolve().parents[1]
p=root/'reference/02-dataflow-api-and-debug-register-reference.md'
front='---'+p.read_text().split('---',2)[1]+'---\n\n'
p.write_text(front+'''# Dataflow API and debug register reference

## Ownership and execution

The engine accepts tokens while its computational state is stopped. An explicit tick operation enables a bounded number of computation cycles. Polling consumes one available final result. Cancellation changes one context's epoch and clears its operand metadata, subject to the offered-output and epoch-wrap guards. These operations are separate so an experiment can stop with a multiply genuinely in flight, inspect it, cancel its context, and resume the remaining work.

The Go interface in `pkg/dataflow/engine.go` is the implemented application boundary:

```go
type Engine interface {
    Execute(context.Context, Operation) (*Token, error)
    Snapshot(context.Context) (Snapshot, error)
    Close() error
}
```

`Operation.Kind` is reset, inject, tick, cancel, or poll. Inject requires a token; tick takes a count bounded by 1,000,000; cancel takes a context in 0..3. Reset optionally accepts a model configuration. A physical device rejects configuration changes because those parameters belong to its bitstream. Execute returns a token only for a successful nonempty poll. A nil token on poll means no currently offered output. Other successful operations return nil.

This command-based boundary replaces the preliminary per-method interface sketch in the IDE design. It gives the HTTP control route and the scenario executor one validated operation type. Both the transaction model and physical serial engine implement it directly. The lower-level model methods remain useful for cycle-level reference tests; there is no prior public application API to preserve.

One application owner must serialize operations and snapshots. `Serial` additionally locks the complete multi-page read so a tick cannot be inserted between two pages by another caller of the same instance. A snapshot is returned only when all required pages have been received and decoded. A partial serial read returns an error and no publishable state.

## Typed envelope

The ten binary bytes are ordered most significant byte first. ASCII transport encodes each binary byte with two hexadecimal characters.

| Byte | Meaning |
| --- | --- |
| 0 | context |
| 1 | epoch |
| 2 | node in bits 7..2; operand port in bit 1; final flag in bit 0 |
| 3 | producer node, or 255 for an external source |
| 4 | subtype; zero for ordinary data, error code for terminal faults |
| 5 | value tag in bits 7..4 and value flags in bits 3..0 |
| 6..9 | 32-bit payload, big endian |

INT is tag 0. BOOL is tag 1 with payload zero or one. Both require zero flags. ERROR is tag 13; it is an engine-produced terminal fault value, not a legal arithmetic operand. Go's Value is an unsigned integer holding the complete 40-bit record. JSON therefore carries it exactly as a number: its maximum is below JavaScript's exact-integer limit. Display an INT payload by interpreting its low 32 bits as signed; do not display the entire 40-bit representation as a signed arithmetic result.

Fault codes are 1 bad destination, 2 duplicate operand, 3 bad tag, 4 integer overflow, and 5 bad descriptor. A fault closes only its context. Other contexts retain their operand state and may complete. Incoming tokens with old epochs are discarded; they do not become new faults.

## UART request and response grammar

The board uses 115200 baud, eight data bits, no parity, and one stop bit. Each request ends with LF. The host waits for the complete response before sending another request. There is no pipelining of requests over this link.

| Request | Payload before checksum | Successful response |
| --- | --- | --- |
| `R` + LF | none | `A` + LF |
| `I` + hex + LF | ten-byte token and one-byte XOR checksum | `A` + LF |
| `T` + hex + LF | four-byte unsigned tick count and checksum | `A` + LF after all enabled cycles |
| `C` + hex + LF | one-byte context and checksum | `A` + LF |
| `Q` + hex + LF | one-byte debug page and checksum | `S` + 20 data hex characters + 2 checksum characters + LF |
| `P` + LF | none | `O` + token/checksum + LF, or `N` + LF |

The XOR of all payload bytes including the checksum must be zero. The leading ASCII command character is not included. A zero tick count is legal and acknowledges without advancing the engine. For example, `T0000000101` followed by LF advances one enabled cycle. `Q0000` followed by LF reads the capability page. `C0202` followed by LF cancels context two when its cancellation guard permits it.

Rejected requests return `!01` for malformed syntax or an incomplete-command timeout, `!02` for an out-of-range control argument, `!03` for checksum failure, `!04` for full input storage, and `!05` for blocked cancellation, each followed by LF. The partial-command timeout is 200 ms at the board clock. The host gives an exchange a two-second deadline and checks its caller's cancellation context during reads and writes.

A full-queue rejection is a definite rejection. The caller can tick the engine, then explicitly submit the same token again. An absent, malformed, or checksum-invalid response is different: the operation may have executed. The serial engine marks itself unsynchronized and refuses further ordinary operations until an explicit reset. It never retries an uncertain mutation automatically. Reset waits for the board's partial-command timeout and drains input before issuing the new reset command.

## Debug pages

Every debug response contains ten binary data bytes. Values in this table are described from most significant to least significant byte unless a bit number is given. Invalid queue entries and pipeline stages must be ignored according to their validity metadata. Operand RAM contents without a corresponding valid flag are not architectural values.

| Page | Data |
| --- | --- |
| 0 | version, contexts, nodes, multiply latency, input depth, completion depth, output depth, epoch width, zero, zero |
| 1 | epochs 3/2/1/0; closed mask in upper nibble and error-pending mask in lower nibble; input count; ready count; completion count; output count; quiescent in bit 7 of final byte |
| 2..9 | two 32-bit counters followed by 16 zero bits |
| 10 | issue token metadata, phase in subtype byte, issue-valid in bit 0 of final byte |
| 11 | captured issue operand A then B, each 40 bits |
| 12 | router token |
| 13 | router-valid in bit 0; delivered-destination mask in bits 2..1 |
| 16..23 | multiply pipeline stage tokens, stage zero first |
| 24 | multiply valid mask in bits 7..0; ALU valid in bit 8 |
| 25 | ALU stage token |
| 32..59 | operand RAM pairs for slot `7 * context + node`; A then B |
| 64..91 | slot flags: valid A bit 0, valid B bit 1, issued bit 2, pending bit 3 |
| 96..103 | input queue entries relative to current head |
| 112..119 | completion queue entries relative to current head |
| 128..135 | output queue entries relative to current head |
| 144..147 | pending terminal error token for contexts 0..3 |

The sixteen counters, in page order, are cycles, accepted external sources, activations, multiply activations, ALU activations, aggregate stale discards, duplicates, invalid contexts, unit-busy cycles, blocked-unit cycles, blocked-router cycles, consumed outputs, input high-water mark, ready high-water mark, completion high-water mark, and output high-water mark. They are unsigned 32-bit counters and wrap naturally.

An issue phase of one waits for the operand read, two captures a response belonging to the selected issue slot, and three waits to dispatch the captured operands into the chosen unit. The RAM read port is synchronous. While paused, the debug address selects a RAM slot and the UART control waits three clocks before capturing its response. On resume, the issue logic checks the previous read address before using RAM data. Thus a debugger cannot replace an activation's operands by reading a different slot during a pause.

The model and RTL share token semantics, bounded storage principles, result values, and cancellation contracts. Their scheduler microsteps are not promised to be cycle-identical. Model issue phases and performance counters describe the model; hardware pages describe the physical engine. The IDE must preserve the source label and must not use model state to fill missing hardware fields.

## Running checked examples

```sh
go run ./cmd/dataflow-lab --engine model --example book --format json
go run ./cmd/dataflow-lab --engine serial --device /dev/ttyACM0 \\
  --example cancel --format json --wire-log /tmp/dataflow-cancel-wire.log
```

Each invocation explicitly resets its selected engine. Book checks results 58 and 12 from interleaved contexts. Copy checks fanout and result 78. Fault checks a duplicate-operand terminal error alongside an unaffected result 12. Cancel stops only after a snapshot shows a multiply stage occupied, changes context two's epoch, injects both stale and new work, and verifies the new result 12. Each example has bounded injection retries and a bounded drain budget. Wire logs contain the actual requests and responses used to obtain the final snapshot.

## Implementation map

- `elastic_dataflow/rtl/dataflow_link.sv`: parser, checksum, bounded stepping, cancellation handshake, debug response capture, UART transmission.
- `elastic_dataflow/rtl/dataflow_core.sv`: architectural state, storage handshakes, issue phases, arithmetic routing, epoch guards, and debug address map.
- `pkg/dataflow/protocol.go`: wire encoding, record decoding, complete snapshot decoding.
- `pkg/dataflow/serial.go`: serialized device ownership, deadlines, explicit resynchronization, and uncertainty handling.
- `pkg/dataflow/engine.go`: validated operations, application interface, and detached model snapshots.
- `pkg/dataflow/examples.go`: device-independent checked examples.
- `cmd/dataflow-lab/main.go`: Glazed CLI fields, structured results, logging, and optional wire capture.
- `elastic_dataflow/sim/tb_dataflow_link.sv`: UART-level framing, pause, tick, debug, cancellation, polling, and bounds tests.
''')
