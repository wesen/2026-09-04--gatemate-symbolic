# Laboratory 4: a lazy graph reducer

## Result

Preload a shared graph:

```text
x    = delay(21 * 2)
root = (x + x) + (x + x)
```

The reducer must return:

```text
INT(168)
```

while the multiplier body executes exactly once. All later demands for `x` observe the memoized value.

A second graph is cyclic:

```text
x = delay(1 + x)
```

Forcing it must return `CYCLIC_THUNK` rather than spinning forever or leaving the heap permanently blackholed.

## Patterns composed

| Pattern | Role |
|---|---|
| Tagged Value Word | distinguishes node kinds and result values |
| Structure / Representation Firewall | centralizes node decode and heap access |
| Indirection / Forwarding Cell | preserves stable graph references after update or movement |
| Demand Cell with Update and Blackhole | claims a thunk, memoizes once, and detects cycles |
| Explicit Continuation Record | makes recursive evaluation a bounded, inspectable stack |
| Type / Shape Dispatch | selects node behavior |
| Bump-Allocated Generation | constructs test graphs and future extensions |
| Common Fast Path / Precise Slow Path | returns values and indirections quickly, faults precisely |
| Delayed Irreversible Store | commits final results to the host interface |

![The reducer serializes thunk claim and update through the heap mutation owner while continuations carry all resumption state.](assets/lazy_reducer.png){ width=92% }

## Node representation

Use one 40-bit word per compact node:

```text
node40
[39:36] tag
[35:32] owner, flags, or small opcode
[31:16] field A
[15:0]  field B
```

Suggested tags:

| Tag | Meaning | Fields |
|---|---|---|
| `INT` | immediate integer node | A:B form signed 32-bit value |
| `ADD` | binary addition | A and B are child addresses |
| `MUL` | binary multiplication | A and B are child addresses |
| `THUNK` | deferred computation | A is body address; B is environment or zero |
| `IND` | forwarding cell | A is target address |
| `BLACKHOLE` | claimed thunk | owner bits name evaluator |
| `ERROR` | exceptional graph result | fields contain code and origin |
| `FREE` | unallocated heap word | fields ignored |

This compact format limits the graph to 16-bit addresses and small environments. That is adequate for the laboratory. A general reducer would use multiword closures or descriptor references behind the representation firewall.

## Continuation records

Three frame kinds are sufficient:

```text
EVAL_RIGHT:
    opcode
    right_child_address

APPLY:
    opcode
    left_value

UPDATE:
    thunk_address
    prior thunk metadata if retry policy requires it
```

A 40-bit tagged union can encode them:

```text
[39:36] frame kind
[35:32] opcode / flags
[31:16] address
[15:0]  compact value fragment or reserved
```

A full 32-bit left value does not fit beside all fields. Several options exist:

1. use a two-word `APPLY` frame;
2. restrict arithmetic to 16-bit values in the first implementation;
3. store left values in a parallel continuation-value RAM;
4. widen frames to 80 bits.

For clarity, use 80-bit continuation entries initially:

```text
upper 40 bits: kind, opcode, addresses, context
lower 40 bits: saved value40
```

Measure later whether a compact specialization is worthwhile.

## Abstract state

```text
M = <current_address,
     control_state,
     heap,
     continuation_stack,
     result_value,
     output_stream,
     evaluator_id,
     heap_top,
     counters,
     fault>
```

The single-evaluator baseline uses one `evaluator_id`. It still records the owner in a blackhole so the extension to multiple evaluators has a defined path.

## Heap access firewall

All node reads and writes pass through a heap service:

```text
read request:
    address, context, epoch, continuation tag

read response:
    node, address, context, epoch, fault

mutation request:
    address, expected tag/version, replacement node

mutation response:
    success, observed tag/version, fault
```

The expected tag or version enables an atomic claim abstraction even if the initial implementation serializes all accesses through one controller. No other module writes the heap RAM directly.

## Main evaluation transitions

### Integer node

```text
EVAL INT(v):
    result_value = INT(v)
    state = RETURN
```

### Indirection node

```text
EVAL IND(target):
    current_address = target
    state = FETCH
```

Maintain an indirection-step counter. If it exceeds a configured bound, return `INDIRECTION_CYCLE` or invoke a slower cycle detector. A legal self-reference sentinel is not used for graph nodes in this laboratory.

### Binary operation

```text
EVAL ADD(left, right):
    reserve one continuation entry
    push EVAL_RIGHT(ADD, right)
    current_address = left
    state = FETCH
```

The same applies to `MUL`.

### Thunk

```text
EVAL THUNK(body):
    reserve UPDATE continuation capacity
    atomically replace THUNK with BLACKHOLE(owner=evaluator_id)

    if claim succeeds:
        push UPDATE(thunk_address)
        current_address = body
        state = FETCH

    else:
        refetch and dispatch the newly observed node
```

The ordering between continuation reservation and heap claim is crucial. Reserve stack capacity first. Claim second. Push the complete `UPDATE` frame as part of the claimed transition or in a state from which cancellation can restore the thunk. The simplest single-evaluator design holds reset as an experiment abort and completes the push immediately after the serialized claim.

### Blackhole

```text
EVAL BLACKHOLE(owner):
    if owner == evaluator_id:
        result_value = ERROR(CYCLIC_THUNK)
        state = RETURN_ERROR
    else:
        suspend as a waiter in the multi-evaluator extension
```

In the baseline, any blackhole belongs to the only evaluator and therefore signals recursive self-demand.

### Error node

```text
EVAL ERROR(e):
    result_value = ERROR(e)
    state = RETURN_ERROR
```

## Return transitions

When a value is ready:

```text
if continuation stack is empty:
    enter OUTPUT_WAIT with result
else:
    inspect top continuation kind
```

### Returning to `EVAL_RIGHT`

```text
frame = pop EVAL_RIGHT(opcode, right)
push APPLY(opcode, left_value=result)
current_address = right
state = FETCH
```

The replacement can be implemented by overwriting the top frame, avoiding a pop followed by push. Ensure the new frame is fully written before control leaves.

### Returning to `APPLY`

```text
frame = pop APPLY(opcode, left_value)

if result or left_value is ERROR:
    combined = propagated error
else if tags are not INT:
    combined = ERROR(TYPE_FAULT)
else:
    combined = arithmetic(opcode, left_value, result)

result = combined
state = RETURN
```

Increment the multiplier-execution counter only when a valid `MUL` application is accepted by the multiplier. Failed or replayed attempts are diagnostic separately.

### Returning to `UPDATE`

```text
frame = pop UPDATE(thunk_address)

if result is a value that fits directly:
    replacement = node encoding of result
else:
    allocate result node
    replacement = IND(result_address)

atomically replace BLACKHOLE(owner=evaluator_id) with replacement
state = RETURN
```

Publish the value before waking any waiters. In the single-evaluator version there are no independent waiters, but the ordering should already match the future extension.

## Directed graph image

Load:

```text
address  node
0        INT 21
1        INT 2
2        MUL 0, 1
3        THUNK body=2          // x
4        ADD 3, 3
5        ADD 3, 3
6        ADD 4, 5              // root
```

Expected high-level trace:

```text
force root 6
  evaluate 4
    force thunk 3
      claim 3 as BLACKHOLE
      evaluate MUL 2
        21 * 2 -> 42
      update node 3 to INT 42
    second reference to 3 -> immediate 42
    ADD -> 84
  evaluate 5
    both references to node 3 -> immediate 42
    ADD -> 84
  ADD -> 168
emit INT 168
```

The required counters are:

```text
thunk claims              1
successful thunk updates  1
MUL applications          1
final result               INT(168)
```

Depending on the exact controller, node 3 may be read several times. Memoization is demonstrated by body execution count, not by a single heap read.

## Cycle graph

Load:

```text
7        INT 1
8        THUNK body=9          // x
9        ADD 7, 8              // body refers to x
```

Forcing node 8 yields:

```text
THUNK 8 -> BLACKHOLE(owner)
push UPDATE(8)
evaluate ADD 9
evaluate INT 7 -> 1
evaluate node 8 -> BLACKHOLE(same owner)
return ERROR(CYCLIC_THUNK)
unwind continuations according to error policy
update node 8 to ERROR(CYCLIC_THUNK) or restore THUNK
emit error
```

Choose one retry policy:

- **memoized error:** replace the blackhole with `ERROR(CYCLIC_THUNK)`; later force returns the same error immediately;
- **retryable thunk:** restore the original thunk and return the error to the current force.

Use memoized error for the laboratory because it guarantees no permanent blackhole and at-most-once evaluation of the failed body. Store the original body in the `UPDATE` continuation only if retry is supported.

## Error unwinding

An error value propagates through `APPLY` frames without executing arithmetic. On `UPDATE`, it becomes the thunk’s memoized error. `EVAL_RIGHT` may skip the right child when the left child is already an error, but that optimization changes which latent errors or effects are encountered. The baseline should follow a clearly specified left-to-right, fail-fast policy:

```text
if left evaluation returns ERROR:
    do not evaluate right
    discard EVAL_RIGHT frame
    continue returning ERROR
```

This behavior belongs in the abstract model.

## Allocation extension

Add a construction port:

```text
ALLOC(tag, A, B):
    if heap_top == heap_limit:
        return HEAP_FULL
    else:
        reserve address = heap_top
        write complete node
        heap_top++
        return REF(address)
```

A branch or transaction can save `heap_top`; rollback discards all younger nodes. If a young node becomes reachable from an older committed node, the publication or promotion policy must prevent rollback from reclaiming it.

## Multiple evaluators

Only after the baseline passes should two evaluators share the heap. Then `BLACKHOLE` has three cases:

```text
same owner        recursive cycle
different owner   subscribe as waiter
stale owner       recover through owner-generation policy
```

A thunk state may include:

```text
owner evaluator
owner epoch
waiter-list head or compact waiter mask
```

The heap service serializes claim and update. A waiter records a complete continuation and context before relinquishing execution. On completion, Tagged Result Wakeup delivers the memoized value to all live waiters. Waiter capacity is finite and must be reserved before suspension.

Cancellation of the owner requires a policy:

- transfer ownership to a waiter;
- restore the thunk for retry;
- complete with cancellation error; or
- forbid owner cancellation while a claim is live.

The first extension should use restore-for-retry under the heap arbiter.

## Verification

Directed tests:

```text
INT root
one ADD
one MUL
IND chain
shared thunk
nested thunks
direct cycle
indirect cycle
type error
continuation overflow
heap address fault
output stall
```

Assertions:

```text
only heap service writes heap
THUNK-to-BLACKHOLE claim has one owner
successful UPDATE replaces matching owned BLACKHOLE
no output while continuation stack nonempty
continuation push reserves capacity before control transfer
indirection limit produces a fault rather than wraparound
multiplier body count for shared test equals one
```

Metamorphic tests:

- duplicate references to a pure thunk do not change its result;
- replacing a memoized thunk with its value in the initial graph preserves result;
- inserting an `IND` node preserves result within depth bound;
- changing multiplier pipeline depth preserves commit trace;
- adding unreachable graph nodes preserves result and counters excluding heap reads.

## Instrumentation

```text
heap reads and writes
node count by tag
indirections followed
thunks observed, claimed, and updated
blackholes observed
body operations by opcode
maximum continuation depth
output stall cycles
fault count by code
```

Capture heap write events with old tag, new tag, address, owner, and context. This makes claim/update races visible without recording the entire heap.

## GateMate mapping

A baseline can use:

- one 1Kx40 graph heap;
- one 512x80 or equivalent continuation memory;
- result FIFO and optional trace RAM;
- CPE logic for tag dispatch and arithmetic;
- an inferred multiplier or iterative unit.

Stop and simplify around six physical RAM blocks and 8,000 CPEs. If continuation width dominates RAM, first restrict arithmetic to 16-bit saved values or split control and saved-value memories. If heap arbitration dominates cycles, measure whether a second read port helps before adding another mutation owner.

## Extensions

- Add pairs and constructor application.
- Add environments and closure nodes.
- Add path compression for indirections.
- Add a copying nursery and forwarding cells.
- Add multiple evaluators with waiter wakeup.
- Add graph rewrite rules as fused semantic assists.

## Exercises

1. Write a complete invariant for thunk states across claim, evaluation, cancellation, and update.
2. Compare direct value replacement with `IND` to a separately allocated result.
