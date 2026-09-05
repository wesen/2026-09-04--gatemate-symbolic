# Laboratory 3: an elastic dataflow expression engine

## Result

Evaluate several independent expression DAGs concurrently:

\[
y = (a \times b) + (c \times d) + \operatorname{boolToInt}(e < f).
\]

Two directed contexts are:

```text
context 0:
    a=7, b=6, c=3, d=5, e=2, f=9
    y = 42 + 15 + 1 = 58

context 1:
    a=10, b=-2, c=4, d=8, e=7, f=1
    y = -20 + 32 + 0 = 12
```

The contexts may complete in either order. Every result must carry the correct context and epoch. Cancelling a context while its multiplication is in flight must prevent any old result or fault from modifying the reused context.

## Patterns composed

| Pattern | Role |
|---|---|
| Tagged Value Word | carries typed operands and faults |
| Ready-Operand Firing | activates a node only when all required operands are present |
| Producer Tag as Future | names unresolved node results |
| Reservation Station | retains ready work while units are occupied |
| Tagged Result Wakeup | routes completions to explicit consumers |
| Elastic Channel | separates functional behavior from pipeline latency |
| Credit-Based Backpressure | bounds long or registered queue paths |
| Context / Epoch-Carrying Token | makes cancellation and context reuse safe |
| Fork / Join Collector | provides the conceptual shape of multi-input nodes |
| Atomic Metadata Propagation | keeps context, epoch, destination, type, and value aligned |

![The dataflow engine turns operand arrivals into bounded activations and routes completions through elastic stages.](assets/dataflow_engine.png){ width=92% }

## Graph structure

Number the graph nodes:

```text
n0 = MUL(a, b)
n1 = MUL(c, d)
n2 = ADD(n0, n1)
n3 = LT(e, f)
n4 = BOOL_TO_INT(n3)
n5 = ADD(n2, n4)   // final result
```

Each descriptor contains:

```text
opcode
required operand mask
up to two explicit destinations
destination input-port selectors
final-result flag
```

For example:

```text
node  opcode       destination 0    destination 1
n0    MUL          n2.portA        none
n1    MUL          n2.portB        none
n2    ADD          n5.portA        none
n3    LT           n4.portA        none
n4    BOOL_TO_INT  n5.portB        none
n5    ADD          output          none
```

Explicit destinations make communication physical and bounded. A node requiring higher fanout is followed by one or more `COPY` nodes. This avoids a global associative broadcast in the first implementation.

## Event record

Use the common 80-bit envelope:

```text
[79:72] context
[71:64] epoch
[63:58] destination node
[57]    destination port
[56]    final-result flag
[55:48] source / producer identity
[47:40] control / fault subtype
[39:0]  value40
```

The event is accepted only when both the control and `value40` are stored together. A bad-tag fault is represented as an error value in the same envelope so that one context can fail without stopping others.

## Operand-store address

For `C` contexts and `N` nodes:

```text
slot_index = context * N + node
```

Store A and B operands in separate 40-bit memories. Keep compact arrays for:

```text
valid_A[slot]
valid_B[slot]
issued[slot]
slot_epoch[slot]
```

For small `C*N`, these bits fit registers and allow direct last-arrival checks. Values remain in BRAM. Larger machines bank by context or node range.

## Operand arrival transition

On input event `T`:

```text
if T.epoch != live_epoch[T.context]:
    stale_drop_count++
    consume and discard T

else if destination node or port is invalid:
    emit CONTEXT_ERROR(BAD_DESTINATION)

else if target slot epoch does not match T.epoch:
    initialize slot for T.epoch
    clear both valid bits and issued

if selected operand valid is already 1:
    emit CONTEXT_ERROR(DUPLICATE_OPERAND)

else if this arrival would complete the node
        and ready FIFO has no reserved capacity:
    do not accept T yet

else:
    write T.value into selected operand store
    set selected valid bit

    if all required operands are now present and issued == 0:
        enqueue activation(context, epoch, node)
        issued = 1
```

The reservation before accepting the last operand is a capacity-closure requirement. Once the operand is stored and `issued` is set, the machine must have a representation for the ready activation.

## Unary operations

Unary nodes such as `BOOL_TO_INT` use:

```text
required operand mask = 2'b01
```

The B slot may remain invalid. Readiness must use the descriptor mask rather than assuming every node has two operands.

A zero-input constant node can either be pre-initialized or use a source event. Avoid special hidden firing rules in the first version.

## Activation queue and station

A minimal activation record is:

```text
context
epoch
node
```

At issue, read the descriptor and operands. Because the operand memories are synchronous, issue is a short pipeline:

```text
cycle 0  dequeue activation; request descriptor and operands
cycle 1  receive descriptor/operands; validate epoch and tags
cycle 2  send to selected functional-unit input when ready
```

The activation itself acts as a small reservation station while waiting for the functional unit. To support multiple waiting operations, add separate station entries per unit class.

Do not clear operand slots at dequeue. Clear them when the operation has been safely accepted by the unit and the design no longer needs them for replay. If replay is supported, retain the station or operands until successful completion.

## Functional units

Start with:

```text
integer ADD/SUB/LT       one-cycle elastic unit
BOOL_TO_INT              one-cycle elastic unit
signed 16x16 MUL         pipelined elastic unit
COPY                     one-cycle elastic unit
```

Use 16-bit multiplication initially and sign-extend to 32 bits. This keeps the arithmetic experiment focused on scheduling rather than big-integer semantics. Invalid integer widths or overflow policy should produce a tagged error.

Every unit accepts and returns the full metadata envelope:

```text
context, epoch, producer node, destinations, result tag, flags, payload
```

Changing multiplier latency must not require changing the scheduler. Add or remove internal `rv_reg` stages and rerun the same tests.

## Completion routing

A completion is first checked against `live_epoch[context]`. If stale, it is consumed and counted but does not access operand stores or output.

For each explicit destination:

```text
construct destination event
send through result-router channel
wait for acceptance
```

If a result has two destinations and only one can accept, retain the completion plus a delivered mask:

```text
delivered[0]
delivered[1]
```

Do not resend an already accepted destination. The completion entry is released only when all required destinations have accepted or the epoch becomes stale.

## Final results

For node `n5`, set `final-result`. Its completion enters an output FIFO as:

```text
{context, epoch, value40, status}
```

Output order across contexts is unspecified. Within a context, the graph is single-assignment and emits at most one final result per epoch.

A host or testbench compares by `(context, epoch)`, not by cycle or global output position.

## Cancellation

Each context has `live_epoch[context]`. Cancellation performs:

```text
live_epoch[context]++
clear or lazily invalidate valid_A, valid_B, and issued for that context
clear context-level pending-output state
```

Old events may remain in:

```text
input FIFO
ready FIFO
issue pipeline
functional units
completion FIFO
result router
```

Every state-changing exit checks epoch. Stale entries still release queue slots and credits. They do not need to be physically searched and deleted.

### Epoch wraparound

With an eight-bit epoch, wraparound is possible. A simple laboratory rule is:

```text
A context may not wrap until all global pipelines and queues have drained.
```

Verification should reduce the epoch width to two or three bits and force the rule. A production design can use wider generations or stronger quiescence tracking.

## Duplicate and malformed events

Define context-level errors:

```text
DUPLICATE_OPERAND
BAD_DESTINATION
BAD_TAG
INTEGER_OVERFLOW
GRAPH_DESCRIPTOR_ERROR
```

An error can cancel the context by incrementing its epoch and enqueueing one final `TAG_ERROR` result under the new or old epoch according to policy. A simpler first policy is:

```text
first error wins
context enters ERROR_PENDING
no further normal state updates
error result remains stable until accepted
then context becomes IDLE with next epoch
```

This policy should be explicit because automatic epoch increment before error emission can make the error appear stale to the output path.

## Queue sizing

A conservative initial design uses:

```text
input event FIFO       8 entries
ready FIFO             16 entries
one station per unit   2 to 4 entries
completion FIFO        8 entries
output FIFO            8 entries
```

These are starting points, not guarantees. Instrument high-water marks and stalls. Reduce capacities in verification to one or two entries to expose hidden assumptions.

The last-operand arrival reserves one ready entry. A functional-unit input reserves space through its internal elasticity. A unit must not accept an operation if its maximum in-flight outputs cannot eventually be retained under the declared backpressure design.

## Ready/valid assertions

For every channel:

```text
valid && !ready  =>  valid and full record stable next cycle
```

For operand stores:

```text
issued implies all required valid bits
one epoch has at most one activation per node
stale event cannot set a valid bit
duplicate arrival cannot overwrite the first value
```

For completion:

```text
each live unit acceptance produces exactly one completion or one terminal fault
completion destination delivered at most once
final result emitted at most once per context epoch
```

## Reference model

The semantic model is a DAG evaluator keyed by `(context, epoch, node)`. It can accept source inputs in any order. When all inputs for a node are available, it computes the node and recursively makes results available to destinations.

The model deliberately ignores cycle timing and queue capacity. The transaction model adds finite queues, random unit latency, and cancellation. The RTL commit trace is compared with the transaction model for internal activations and with the semantic model for final values.

## Directed schedule

Inject source events in an adversarial interleaving:

```text
ctx0.a, ctx1.f, ctx0.d, ctx1.a,
ctx0.f, ctx1.b, ctx0.b, ctx1.e,
ctx0.c, ctx1.c, ctx0.e, ctx1.d
```

Then randomly stall the multiplier and output. The final outputs remain:

```text
(context 0, value INT(58))
(context 1, value INT(12))
```

in either order.

Cancellation test:

1. Supply `ctx2.a` and `ctx2.b` so `n0` enters the multiplier.
2. Cancel context 2.
3. Reuse context 2 with the next epoch and a new input set.
4. Delay the old multiplier result until after new operands arrive.
5. Prove the old result is counted stale and never fills a new slot.

## Measurements

Record:

```text
source events accepted
activations fired
operation count by unit
unit busy and blocked cycles
ready FIFO and completion FIFO high-water marks
router destination stalls
stale events dropped by location
duplicate-event faults
cycles per completed context
```

Run ablations:

- multiplier latency 1, 2, 4, and 8;
- ready FIFO depth 1 through 16;
- one versus two add units;
- explicit destinations versus small broadcast for the same graph;
- one context versus several interleaved contexts.

The purpose is to see which pattern removes the actual bottleneck.

## GateMate mapping

A conservative design may use:

- two 40-bit operand RAMs;
- descriptor ROM;
- ready and completion FIFOs;
- output FIFO;
- optional trace RAM;
- CPE logic for valid maps, routing, add/compare, and control;
- composed small multipliers or inferred multiplication according to synthesis results.

Stop and simplify around eight physical RAM blocks and 8,000 CPEs. If associative matching dominates area, return to explicit destinations and indexed slots. If ready logic dominates timing, insert elastic registers or credits.

## Extensions

- Add a dynamic graph loader.
- Add higher-fanout copy trees.
- Add a nonblocking external-memory node.
- Add replay for bank conflicts.
- Add a per-context speculation checkpoint and branch/select nodes.
- Cluster nodes and compare local versus global routing.

## Exercises

1. Prove exactly-once activation when both operands can arrive in the same cycle through two ingress ports.
2. Design a two-destination router that permits either destination to stall independently.
3. Calculate the minimum epoch width for a known maximum event lifetime and cancellation rate, or state a quiescence rule.
4. Add a three-input node and derive its last-arrival and capacity reservation logic.
5. Compare producer-tag broadcast with explicit destinations for a graph of 32 nodes and average fanout 1.4.

