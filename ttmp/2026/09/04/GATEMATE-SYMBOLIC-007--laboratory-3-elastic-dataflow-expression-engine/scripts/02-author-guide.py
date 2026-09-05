#!/usr/bin/env python3
"""Materialize the initial intern guide before implementation begins."""
from pathlib import Path
T=Path(__file__).resolve().parents[1]
p=T/'design-doc/01-elastic-dataflow-engine-intern-analysis-design-and-implementation-guide.md'
front=p.read_text().split('---',2)[1]
body=r'''
# Laboratory 3: An Elastic Dataflow Expression Engine

## Executive summary and scope

This laboratory evaluates expression graphs with several independent executions in flight. It builds the third laboratory from *Composable Hardware Patterns for Symbolic Computers*, using tagged values, operand readiness, bounded storage, elastic arithmetic pipelines, explicit result destinations, and context epochs. The principal example is `y = a*b + c*d + boolToInt(e<f)`. Context 0 with inputs `[7,6,3,5,2,9]` must return 58; context 1 with `[10,-2,4,8,7,1]` must return 12. Results may arrive in either order and are identified by context and epoch.

The implementation will have four context slots, eight-bit epochs, the book's six-node descriptor graph, and a seventh COPY node for directed fanout tests. A Go semantic model and a bounded transaction model establish behavior before RTL. The FPGA will expose an input-event interface, cancellation, an elastic result output, counters, and a UART laboratory control link. A Go host API and command will run the directed examples and physical cancellation checks. The earlier tagged CPU, queens solver, and graph-coloring browser remain separate experiments.

The first version uses a fixed descriptor graph. Dynamic graph loading, external-memory nodes, speculative branches, and integration into the graph-coloring React interface are future work. This scope implements the concurrent machine itself, including physical verification; it does not present a software simulation as FPGA execution. The circuit must tolerate multiplier latency changes and output stalls without changing numerical results or losing metadata.

## 1. What changes from the preceding laboratories

The tagged stack CPU sequences instructions and carries types with values. The rollback engines maintain one search state and reconstruct earlier versions of that state. A dataflow engine has a different control problem: many operations may be waiting for different operands, several operations may be executing, and a completed operation must find its consumers without assuming a global instruction order.

The current repository provides reusable electrical and storage components. `symbolic_eval/rtl/symbolic_types_pkg.sv` defines the 40-bit value representation and canonical integer/boolean constructors. `symbolic_eval/rtl/rv_reg.sv` documents ready/valid ownership and one-entry elasticity. `symbolic_eval/rtl/sync_sdp_ram.sv` supplies synchronous operand storage. `graph_microscope/rtl/uart_rx.sv` and `symbolic_eval/rtl/uart_tx.sv` provide the physical serial byte interfaces. The GateMate constraint files already establish a 10 MHz clock and Olimex board pins.

The existing graph core does not implement concurrent contexts, operand matching, or cancellation of arithmetic already in flight. Its trace-ready gate freezes a sequential solver at semantic boundaries. That mechanism cannot replace a distributed ready/valid network. The new RTL will therefore live in `elastic_dataflow/rtl/`, with reusable Go types and models in `pkg/dataflow/`. Existing low-level modules are reused directly; no compatibility adapter is introduced.

| Existing file | Observed contract | Use in this laboratory |
|---|---|---|
| `symbolic_eval/rtl/symbolic_types_pkg.sv` | Tag, flags, and 32-bit payload form a 40-bit word. | Preserve integer/boolean/error representation. |
| `symbolic_eval/rtl/rv_reg.sv` | A blocked valid item remains stable; replacement is allowed on acceptance. | Reference for elastic stages and assertions. |
| `symbolic_eval/rtl/sync_sdp_ram.sv` | Reads become available on a subsequent clock edge. | Separate A/B operand memories and explicit issue latency. |
| `graph_microscope/rtl/uart_rx.sv` | Synchronized start/data/stop sampling at a configurable divider. | Receive laboratory commands. |
| `symbolic_eval/rtl/uart_tx.sv` | Ready/start handshake for one 8N1 byte. | Transmit command acknowledgements and results. |
| `pkg/microscope/serial.go` | Bounded stop-and-wait host exchanges. | Design reference; new protocol is independent. |

## 2. Expressions, descriptors, and contexts

A directed acyclic graph represents dependencies between operations. A node can execute only after all required operands have arrived. Each node executes at most once within one context epoch. The graph topology is shared, while operand values and readiness are indexed by context, so different input sets execute independently.

```text
n0 = MUL(a, b)
n1 = MUL(c, d)
n2 = ADD(n0, n1)
n3 = LT(e, f)
n4 = BOOL_TO_INT(n3)
n5 = ADD(n2, n4)       final result
n6 = COPY(x)          optional source, sends x to n0.A and n1.A
```

The COPY node is unused when all six ordinary source inputs are provided. Its separate test graph usage replaces the `a` and `c` source arrivals with one copied source. This exercises two explicit destinations and delivery tracking without adding dynamic graph configuration. Ordinary and COPY source modes must not both supply those same operand ports in one epoch; doing so is a duplicate-operand fault.

| Node | Operation | Required mask | Destination 0 | Destination 1 |
|---|---|---|---|---|
| 0 | MUL | `11` | node 2 port A | none |
| 1 | MUL | `11` | node 2 port B | none |
| 2 | ADD | `11` | node 5 port A | none |
| 3 | LT | `11` | node 4 port A | none |
| 4 | BOOL_TO_INT | `01` | node 5 port B | none |
| 5 | ADD | `11` | final output | none |
| 6 | COPY | `01` | node 0 port A | node 1 port A |

Unary operations require only A. An invalid B slot must not prevent their activation, and sending B to a unary node is malformed. The descriptor contains operation, required mask, explicit destinations and port selectors, and final-result status. Producer identity travels with the result and supplies the descriptor lookup for routing; no associative global tag broadcast is needed.

## 3. Typed values and the 80-bit envelope

Every data token carries both value and identity. A value uses `tag[3:0]`, `flags[3:0]`, and `payload[31:0]`. Integer tag 0 represents a signed 32-bit payload; boolean tag 1 permits only payload 0 or 1. Flags are zero in this lab. Error tag D carries a defined fault code. Arithmetic rejects wrong tags and noncanonical values rather than reinterpreting their bits.

The event envelope follows the book:

```text
79       72 71       64 63      58 57 56 55       48 47       40 39         0
+----------+-----------+----------+--+--+-----------+-----------+-------------+
| context  | epoch     | dest node|pt|F | producer  | subtype   | value40     |
+----------+-----------+----------+--+--+-----------+-----------+-------------+
```

The destination node and port describe where an operand is to be stored. A unit completion retains producer identity and its value so the router can construct destination events. Final outputs retain context and epoch. All 80 bits move atomically through a FIFO or pipeline stage; keeping a value in one register and its epoch in unrelated control is a correctness defect under stalls.

ADD and SUB operate on signed 32-bit integers and check overflow. LT compares signed integers and produces a canonical boolean. BOOL_TO_INT converts the canonical boolean to integer 0 or 1. MUL initially requires both operands to fit signed 16 bits and produces a sign-extended signed 32-bit product. COPY preserves a supported canonical value. Width violations, bad tags, invalid destinations, duplicate operands, and invalid descriptors become context-local tagged errors.

## 4. Operand storage and exactly-once activation

For four contexts and seven nodes, `slot = context*7 + node`. Two 40-bit memories store A and B. Compact register maps store A-valid, B-valid, issued, and ready/pending state. Cancellation clears the maps for that context; memory contents need not be erased because values are meaningful only when the current maps authorize them.

The first implementation uses an indexed activation reservation for every context/node pair. A slot can transition from incomplete to ready exactly once per epoch. That same slot provides persistent representation for its activation, so accepting the final operand cannot lose a ready operation due to a full unrelated ready FIFO. The scheduler scans ready entries and selects one whose unit class can accept work, avoiding head-of-line blocking behind a busy multiplier.

This deliberately differs from the book's suggested shared ready FIFO. A bounded ready bitmap is appropriate for the fixed 28-node instance: it removes an avoidable queue dependency while keeping storage strictly bounded. Metrics report the number and high-water mark of pending activations. Input, completion, and final-output paths still use finite FIFOs. A later larger graph may justify banked activation queues; that change would need an explicit reservation and deadlock argument.

```text
accept_operand(token):
    if context or destination is invalid: report bounded error
    else if token.epoch != live_epoch[context]: count stale; discard
    else if context is closed or has a pending error: discard without mutation
    else if operand port is not required: set first context error
    else if selected operand is already valid: set first context error
    else:
        write value into selected A/B memory
        set its valid bit
        if required operands are now all present and not issued:
            set pending_activation[slot]
            set issued[slot]
```

Only one operand-store write is committed per cycle. The arbiter alternates external input and routed completion events when both can proceed, so a sustained external stream cannot indefinitely starve internal dependencies. A duplicate never overwrites the first value. Issued remains set until cancellation or context reuse; consuming a pending activation does not make it legal to fire that node again.

## 5. Issue, arithmetic pipelines, and completion routing

```mermaid
flowchart TD
    Input[Input token FIFO] --> Arrival[Epoch check and operand-write arbiter]
    Arrival --> Operands[A and B synchronous memories]
    Arrival --> Ready[Indexed ready activation reservations]
    Ready --> Issue[Descriptor and operand read pipeline]
    Operands --> Issue
    Issue --> ALU[Elastic ADD SUB LT BOOL COPY unit]
    Issue --> MUL[Configurable elastic multiplier pipeline]
    ALU --> Merge[Completion arbiter and FIFO]
    MUL --> Merge
    Merge --> Router[Explicit destinations and delivered mask]
    Router --> Arrival
    Router --> Output[Final output FIFO]
    Cancel[Context epoch and cancellation] --> Arrival
    Cancel --> Issue
    Cancel --> Router
```

Issue requests descriptor and operand addresses, waits for registered memory output, captures both operands and metadata, and holds the operation until its selected unit accepts it. The unit takes ownership only on valid-and-ready. There are separate arithmetic classes for the short ALU operations and multiplication, so a long multiplication does not prevent an independent comparison from executing.

The multiplier's depth is a parameter tested at 1, 2, 4, and 8 stages. Arithmetic may be computed at admission and carried through elastic stages for this small laboratory; the architectural requirement is that changing pipeline depth changes latency and storage, not operand matching or result identity. Synthesis determines the actual implementation of the signed 16×16 operation. We will report mapping results rather than assume a particular multiplier primitive.

When two units finish together, arbitration accepts one complete metadata record into the bounded completion queue and holds the other. The router retains one completion while delivering its explicit destinations. A two-bit delivered mask records accepted destinations. If destination 0 has accepted and destination 1 stalls, the router must retain the result and mask without resending destination 0. Cancellation of the producer epoch makes remaining delivery stale and releases the entry.

```text
route(completion):
    if epoch is stale or context is closed: consume; count drop
    else if value is an error: latch first error for context; consume
    else if descriptor is final:
        retain until output storage accepts exactly one result
        close this context epoch
    else:
        for each not-yet-delivered explicit destination:
            construct an operand event with the same context and epoch
            set its delivered bit only after acceptance
        release completion when all destinations are accepted
```

Every queue has finite capacity and a conservation invariant: accepted minus removed equals occupancy, within bounds. A blocked producer holds its entire item stable. The activation reservation pool eliminates a shared-ready-queue cycle, but fairness and downstream output draining still matter to progress. An output stall may intentionally stop a completed context; it must never cause data loss, duplicate execution, or metadata changes.

## 6. Cancellation, errors, and context reuse

Cancellation increments a context's live epoch and invalidates its operand/activation maps. Old tokens may remain in input storage, issue staging, arithmetic stages, the completion FIFO, or the router. Their storage is released through normal dequeue/retirement, while epoch checks prevent them from changing current operands or producing a new live result. Stale drops are counted.

Epoch checks belong at state-changing exits, not only at ingress. A multiply can accept a current token, remain in flight through cancellation, and finish after the same context has received new operands. Its old epoch must still accompany the product. The new execution must receive only results carrying the new epoch.

An already offered final output obeys the same ready/valid stability rule as other channels. Cancellation of that specific context is temporarily refused while its output is externally held, rather than withdrawing a blocked valid result. Queued outputs that have not been offered can be discarded as stale at their dequeue boundary. Tests must exercise this boundary and document any resulting cancellation backpressure. The host always identifies outputs by both context and epoch.

Eight-bit epochs can wrap. The lab permits a wrap only after the entire engine's queues, unit stages, router, and issue pipeline have drained. Pending incomplete operands are cleared by the cancellation itself; they are not in-flight messages. Small-epoch simulation configurations force wrap frequently and must prove the rule. Simply using eight bits without a drain rule would leave a stale-token alias hazard.

The first context error wins. It closes normal mutation for that epoch and retains one error result until output storage accepts it. Other contexts keep operating. The error keeps the failing epoch, avoiding a policy where incrementing the epoch before emission makes the error appear stale. Explicit cancellation can supersede pending internal work, subject to the stable offered-output rule. An invalid context number has no context storage to close and is consumed/reported as a protocol/global diagnostic rather than indexing outside an array.

## 7. Go models, APIs, and evidence boundaries

`pkg/dataflow` will define Value, Token, Descriptor, Result, and error constants. A semantic evaluator consumes operands in arbitrary order and computes ready nodes without cycle timing. Its result set is compared by `(context, epoch)`. A separate transaction model adds bounded queues, issue delay, configurable multiplier latency, stalls, and cancellation. Unit arithmetic also has direct mathematical tests so shared implementation mistakes do not become the only oracle.

```go
type Token struct {
    Context uint8
    Epoch uint8
    Node uint8
    Port uint8
    Final bool
    Producer uint8
    Subtype uint8
    Value Value
}

// Planned host boundary; all transport operations have a context deadline.
type Device interface {
    Inject(context.Context, Token) error
    Cancel(context.Context, uint8) error
    Tick(context.Context, uint32) error
    Poll(context.Context) (*Token, error)
    Stats(context.Context) (Stats, error)
    Reset(context.Context) error
    Close() error
}
```

The UART link is an instrument control plane. It remains clocked while computation is paused. Injection queues a bounded input token; Tick advances the engine a requested bounded number of enabled cycles; Poll accepts a final output; Cancel changes an epoch; Stats reports epochs, occupancy, and counters. This separation allows a physical test to stop after a multiplication enters a pipeline, cancel the context, reuse it, and then release the old completion. UART speed otherwise makes cancellation of a few-cycle multiply difficult to reproduce physically.

The planned line protocol uses ASCII hex with XOR checksums for payload commands and replies. `I` carries the ten-byte token plus checksum. `C` carries a context byte plus checksum. `T` carries a 32-bit enabled-cycle count plus checksum. `P`, `Q`, and `R` have no payload and mean poll, query, and global reset. Successful commands return `A` newline; no available result returns `N` newline; output returns `O` plus the ten-byte token and checksum. Query returns a fixed documented status record. Malformed commands, full input storage, and blocked cancellation return explicit errors without claiming acceptance. Exact status offsets will be finalized with the codec in P4 and added to this guide.

The Go command under `cmd/dataflow-lab/` will use Glazed flags, zerolog, and the root module. It will expose model and physical demonstration modes, device path, and log level. Tests can call the device API directly. The command must identify its execution source and must not reopen the board behind the graph service's serial ownership.

## 8. Directed examples and invariants

The book's adversarial input order is `ctx0.a, ctx1.f, ctx0.d, ctx1.a, ctx0.f, ctx1.b, ctx0.b, ctx1.e, ctx0.c, ctx1.c, ctx0.e, ctx1.d`. Both results must emerge with the correct identities despite comparisons and multiplications becoming ready at different times. No test should require a fixed global output order.

The cancellation test injects enough operands for context 2's first multiplication, advances until unit acceptance, cancels, supplies new-epoch inputs, and releases the old result after new operands exist. The expected result contains only new inputs; the stale counter increases, and no old fault or operand modifies the new slot. A separate test cancels queued input and pending activations. Context-local type, duplicate, and overflow faults must not stop an independent valid context.

```text
channel valid && !ready -> next valid and full record unchanged
issued[slot] -> required operands were accepted for this epoch
each context/epoch/node activates at most once
stale event -> no operand write and no current output
accepted unit operation -> exactly one completion, including errors
each explicit destination -> delivered at most once
each context epoch -> at most one accepted final result or error
epoch wrap -> all old message storage globally drained
```

Test FIFO depths down to one where supported, use random input/output stalls, vary multiplier latency, and record deterministic seeds. A model test proves semantic values; an RTL trace proves implementation transactions; a physical capture proves programming, pins, clock, UART, and the real device path. These are complementary claims. An untested ablation is not a measured optimization.

## 9. Decision records

### Fixed descriptors and an optional COPY source

- **Context:** The book makes dynamic loading an extension; the core problem is elastic scheduling.
- **Options:** Runtime descriptors, fixed six-node graph, or fixed graph plus a bounded fanout exercise.
- **Decision:** Fixed six-node expression plus a seventh COPY descriptor.
- **Rationale:** Covers arithmetic dependencies, unary readiness, and partial fanout without introducing a graph-loader protocol.
- **Consequences:** Host input events must follow one source mode; duplicate ports are errors. Arbitrary graphs are not advertised.
- **Status:** accepted.

### Indexed activation reservations

- **Context:** A shared ready FIFO can couple completion routing to a blocked issue head.
- **Options:** Credit-reserved ready FIFO with per-unit stations, or one bounded activation reservation per single-assignment slot.
- **Decision:** Use indexed ready/issued maps with unit-aware selection.
- **Rationale:** The fixed 28-slot machine has a natural bounded place for every activation, making last-arrival capacity explicit.
- **Consequences:** More readiness bits and a bounded selection network; no ready-FIFO-depth ablation is claimed for this implementation. Queue-depth experiments apply to the actual input/completion/output FIFOs.
- **Status:** accepted.

### Epoch invalidation with explicit output ownership

- **Context:** Cancellation races with both arithmetic and blocked final output.
- **Options:** Search every queue, invalidate by epoch, or withdraw visible output on cancel.
- **Decision:** Epoch checks discard stale internal records; cancellation waits when it would withdraw an already offered output. Wrap requires global drain.
- **Rationale:** Maintains ready/valid stability and prevents old arithmetic from modifying reused contexts.
- **Consequences:** Cancellation can be backpressured at the final-output boundary. Tests must distinguish acceptance from request.
- **Status:** accepted.

### Instrument clock-enable control

- **Context:** A physical multiply finishes much faster than host UART commands arrive.
- **Options:** Enormous arithmetic delays, simulation-only cancellation, or a separately clocked control plane with bounded enabled-cycle commands.
- **Decision:** Keep UART live while advancing computation through explicit Tick commands.
- **Rationale:** Reproduces in-flight cancellation on physical hardware without changing arithmetic latency parameters.
- **Consequences:** Wall-clock command time is not engine cycle time. All computation state transitions and handshakes must obey enable consistently.
- **Status:** accepted.

## 10. Task-by-task implementation plan

### P1: Guide and delivery

Archive the lab chapter and requested printing skill when available, create this guide and diary, print the overall plan and phase start, relate source files, run docmgr doctor, and dry-run then upload the guide to reMarkable. Commit the concrete contract before implementation.

### P2: Go executable contracts

Implement value/envelope codecs, descriptors, independent arithmetic cases, semantic evaluation, and a transaction model with bounded storage and cancellation. Test 58/12, signed multiplication, unary nodes, duplicate operands, wrong tags, overflow, COPY fanout, output stalls, and reduced epoch wrap. Commit tested software contracts and diary evidence.

### P3: Elastic RTL

Add the dataflow package, parameterized FIFO and arithmetic pipeline, operand store/ready maps, issue sequencer, completion router, cancellation, and error policy. Use the existing synchronous RAM and board-independent simulation. Compare transactions/results against models, assert stall stability, and run latency/depth configurations. Commit passing RTL and tests.

### P4: UART and physical execution

Add the byte-command link and board top, synthesis script/Makefile, Go serial driver, and Glazed demonstration command. Test UART framing in simulation, synthesize/route at 10 MHz, stop the previous serial service before programming, and verify physical 58/12 and cancellation under enabled-cycle control. Preserve raw frames, command history, image hash, and mapping logs. Commit the integrated laboratory.

### P5: Measurements and handoff

Run directed/randomized stress, reduced queues/epochs, pipeline latency variants, and relevant prior-lab regressions. Record accepted source events, activations by unit, blocked cycles, occupancy high-water marks, stale drops, duplicate faults, and enabled cycles. Compare one context with interleaved contexts. Optional architecture experiments such as a second ALU or broadcast routing are explicitly deferred unless implemented and measured. Reconcile guide with final code, upload a completed guide under a new name, finish diary/tasks, print completion, commit and push.

Each phase has a printed start and done slip with a commit reference on completion. The overall plan plus five phase pairs totals eleven slips. Scripts belong in scripts directories, downloaded references in sources, and generated build products remain ignored.

## 11. Risks, review focus, and acceptance

The critical risks are accepting a final operand without an activation record, reading synchronous RAM too early, losing metadata under stalls, double-delivering a COPY destination, admitting stale completions after cancellation, wrapping epochs with old tokens present, and withdrawing a held final output. These are explicit tests, not reasons to silently simplify the machine into sequential evaluation.

The physical mapping target follows the book's caution around 8,000 CPEs and eight physical RAM blocks. If synthesis exceeds the target, inspect the actual cause—register arrays, muxing, multiplier structure, or FIFO mapping—and record a measured simplification. Do not quote pre-route frequency as final timing. A working two-context physical example plus a cancellation trace is required for claiming board completion.

Acceptance requires correct 58 and 12 results keyed by context/epoch, exactly-once activation, stable complete records under backpressure, context-local errors, stale suppression during actual in-flight reuse, enforced epoch wrap policy, passing simulation and host checks, a routed/programmed device, documented measurements, and the completed ticket/diary/delivery receipts. No user decision blocks this scoped design.

## References and reading order

Read the archived `sources/laboratory-3.md` first for the book's experiment and exercises. Then read the existing tagged-value package, elastic register, synchronous RAM wrapper, and UART receiver/transmitter named in section 1. During implementation, proceed through `pkg/dataflow` types and model, `elastic_dataflow/rtl` arithmetic and controller, the UART link, then `cmd/dataflow-lab`. The tests and archived physical records are the authoritative evidence for implemented behavior and measured limits.
'''
p.write_text('---'+front+'---\n'+body.strip()+'\n')
tasks=['P1: Publish intern guide and design contract.', 'P2: Implement and test typed semantic and transaction models.',
       'P3: Implement elastic RTL scheduling, routing, and epoch cancellation.', 'P4: Implement UART host control and demonstrate physical execution.',
       'P5: Finish stress tests, measurements, documentation, and handoff.']
(T/'tasks.md').write_text('# Tasks\n\n'+'\n'.join('- [ ] '+task for task in tasks)+'\n')
print(f'Wrote {len(body.split())} guide words')
