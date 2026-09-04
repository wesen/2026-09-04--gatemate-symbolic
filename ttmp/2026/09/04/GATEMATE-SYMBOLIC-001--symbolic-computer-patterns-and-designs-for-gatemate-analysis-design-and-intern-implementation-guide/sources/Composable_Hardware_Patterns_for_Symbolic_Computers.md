---
title: "Composable Hardware Patterns for Symbolic Computers"
subtitle: "From abstract-machine semantics to GateMate FPGA experiments"
author: ""
date: "September 2026"
documentclass: book
classoption:
  - oneside
  - openany
papersize: letter
fontsize: 10pt
geometry:
  - inner=0.9in
  - outer=0.75in
  - top=0.8in
  - bottom=0.85in
mainfont: "Noto Serif"
sansfont: "Inter"
monofont: "DejaVu Sans Mono"
titlegraphic: "assets/cover_architecture.png"
toc: true
toc-depth: 1
numbersections: true
colorlinks: true
linkcolor: blue
urlcolor: blue
header-includes:
  - \usepackage{microtype}
  - \usepackage{fvextra}
  - \fvset{breaklines=true,breakanywhere=true,fontsize=\small}
  - \usepackage{booktabs}
  - \usepackage{longtable}
  - \usepackage{tocloft}
  - \setlength{\cftchapnumwidth}{2.4em}
  - \setlength{\cftsecnumwidth}{3.6em}
  - \setlength{\cftsubsecnumwidth}{4.5em}
  - \makeatletter
  - \renewcommand{\@pnumwidth}{2.2em}
  - \renewcommand{\@tocrmarg}{3.2em}
  - \makeatother
  - |
    \renewcommand{\maketitle}{%
      \begin{titlepage}
      \centering
      \vspace*{0.045\textheight}
      {\fontsize{27}{33}\selectfont\bfseries Composable Hardware Patterns for Symbolic Computers\par}
      \vspace{0.8em}
      {\Large From abstract-machine semantics to GateMate FPGA experiments\par}
      \vspace{2.2em}
      \includegraphics[width=0.86\textwidth]{assets/cover_architecture.png}
      \vfill
      {\large September 2026\par}
      \end{titlepage}%
    }
  - \usepackage{array}
  - \usepackage{caption}
  - \captionsetup{font=small,labelfont=bf}
  - \usepackage{needspace}
  - \usepackage{enumitem}
  - \setlist{nosep}
  - \setlength{\emergencystretch}{3em}
  - \raggedbottom
---

# Preface {-}

This book develops a pattern language for processors and accelerators whose work is dominated by symbolic values, dynamic structure, irregular control, reversible state, graph-shaped data, or managed memory. Its immediate practical target is the Cologne Chip GateMate FPGA family. Its larger aim is to make hardware design more compositional.

The phrase *hardware pattern* is used here in a strict sense. A pattern is not merely a familiar block such as a FIFO, ALU, cache, or pipeline stage. It is a reusable state-transition contract that identifies:

- the state it owns;
- the data and metadata it accepts and produces;
- the ordering and visibility it guarantees;
- the finite resources it consumes;
- the way it stalls, faults, replays, commits, and rolls back;
- the physical assumptions it makes about latency, fanout, ports, and locality; and
- the invariant that permits one implementation to replace another.

The catalog is derived from several architectural traditions: Kogge's synthesis of symbolic-computer architecture; Landin's explicit abstract-machine state; Warren's WAM and its treatment of environments, choice points, and trails; dataflow activation; Tomasulo scheduling; precise interrupt machinery; lazy graph reduction; garbage collection; capability systems; decoupled access/execute; systolic computation; and latency-insensitive design. The pattern names and organization are original to this textbook. Historical mechanisms are generalized into contracts that can be recombined.

The book has four intended uses. It can be read as an architecture text, used as a catalog during design reviews, followed as a sequence of FPGA laboratories, or used to structure experimental research. The five main laboratories progress from a typed stack evaluator to a small relational query engine. Each laboratory produces a concrete result, carries explicit assertions, and reuses verified infrastructure from the previous laboratories.

The resource numbers in the laboratories are **stop-build budgets**, not measured utilization claims. Actual area, timing, RAM packing, and power must be obtained from the chosen toolchain version, board constraints, RTL, and place-and-route seed. Likewise, the 20-, 40-, and 80-bit formats used throughout are convenient GateMate-oriented physical conventions, not permanent architectural interfaces.

## How to read the book {-}

Readers new to symbolic architecture should read Parts I and II in order. Readers already familiar with abstract machines can begin with Chapter 3, then use the pattern catalog selectively. FPGA implementation teams can begin with Chapters 6 and 14 through 18 before undertaking the laboratories. Architecture researchers should pay particular attention to the commitment model in Chapter 4 and the experimental method in Chapter 27.

Every catalog entry follows the same shape: intent, problem, forces, contract, operation, composition, FPGA realization, verification, failure signs, and lineage. This repetition is deliberate. It makes omissions visible. If a proposed mechanism has no overflow behavior, no rollback owner, or no precise slow path, the empty field is itself a design finding.

## Notation {-}

A pattern contract is written

\[
P = \langle S,D,F,O,V,R,B,T,I \rangle
\]

where the components mean state, representation, flow, ordering, visibility, recovery, bounds, timing, and invariant. The symbol \(P \otimes Q\) denotes composition, subject to closure conditions defined in Chapter 5.

Four commitment levels recur throughout:

- \(\mu\): microarchitecturally provisional;
- \(A\): architecturally retired;
- \(S\): semantically committed; and
- \(E\): externally visible.

Signal and field names appear in `monospace`. Pseudocode specifies an abstract transition unless a listing is explicitly labeled SystemVerilog. Assertions are properties to adapt to the local reset and sampling convention.

## Safety and scope {-}

These designs are educational and experimental. They are not presented as production-quality safety, security, cryptographic, medical, aerospace, or life-support components. A successful simulation and FPGA demonstration do not substitute for timing closure, CDC analysis, electrical review, formal verification, fault analysis, or application-specific certification.

# Part I - Foundations

# Symbolic computers as architecture laboratories

## Learning objectives

By the end of this chapter, you should be able to:

1. explain why symbolic workloads expose architectural mechanisms that ordinary integer benchmarks can hide;
2. distinguish source-language constructs from abstract-machine transitions;
3. identify common state-machine shapes across functional, logic, dataflow, and out-of-order machines; and
4. state why a pattern language should be organized around contracts rather than named blocks.

## What makes computation symbolic?

A workload is symbolic when values carry structure and meaning that affect control. A term can be an integer, atom, variable, pair, closure, continuation, thunk, capability, error, or pointer to another term. Operations therefore do more than apply arithmetic. They inspect tags, follow indirections, allocate nodes, bind variables, capture continuations, suspend on missing inputs, and sometimes reverse earlier mutations.

This does not imply that symbolic computing is a separate universe from conventional CPU design. A modern out-of-order core also manipulates symbolic identities: register-renaming tags name future values; reorder-buffer entries represent provisional effects; load/store queue entries encode unresolved ordering; branch masks name cancellation domains. The difference is mainly where semantics are exposed. Symbolic machines make continuations, alternatives, graph updates, and language-level commitment explicit. Conventional cores usually make instruction retirement explicit while leaving higher-level semantics to software.

Symbolic architecture is therefore a useful laboratory. It contains unusually clear examples of four general problems:

- **representation:** how a value carries enough metadata to be interpreted safely;
- **activation:** how work becomes eligible when operands, authority, or demand arrive;
- **lifetime:** where dynamic state lives and when it may be reclaimed;
- **reversibility:** how provisional effects become committed or are undone.

## Kogge's architectural perspective

Peter M. Kogge's *The Architecture of Symbolic Computers* surveys machines for functional and logic-language execution and treats interpreters, abstract machines, microcode, hardware organizations, graph reduction, unification, memory management, and parallel evaluation as connected design levels [KOG91]. The important lesson for this book is methodological: begin with the semantic mechanism, then decide how much to place in software, firmware, microarchitecture, or dedicated hardware.

A machine intended to evaluate a functional program may describe state in terms of an operand stack, environment, control sequence, continuation, heap, and update stack. A logic machine may describe arguments, environments, choice points, heap allocation, trail state, and clause alternatives. Neither description commits the designer to a single datapath. One transition might be a bytecode, a macro-instruction, a multi-cycle controller sequence, a decoupled engine request, or software running on an ordinary CPU.

That separation is the opening for patterns. The contract is stable; the implementation is replaceable.

## Four lineages, one set of state-machine shapes

Landin's SECD machine made evaluation state explicit as stack, environment, control, and dump [LAN64]. Warren's WAM made logic-program execution explicit through registers, environments, choice points, heap, trail, and a compact instruction system [WAR83]. Dataflow machines associated work with operand presence rather than one global program counter [DEN75]. Tomasulo's algorithm represented unavailable operands by producer tags and woke dependent operations when results appeared [TOM67].

The mechanisms are not interchangeable, but their shapes rhyme:

| Historical mechanism | Generalized shape |
|---|---|
| SECD dump | explicit continuation record |
| WAM choice point | minimal restart checkpoint |
| WAM trail | selective undo log |
| Prolog cut | semantic commitment fence |
| graph-reduction update | memoizing future with stable identity |
| dataflow matching | readiness-driven activation |
| Tomasulo tag | unresolved value represented by producer identity |
| reorder buffer | ordered visibility boundary |
| nonblocking cache miss entry | latency-separated outstanding transaction |
| systolic cell | repeated actor with local communication |

The pattern language in this book extracts these state-machine shapes while preserving their different commitment levels.

## Why named blocks are insufficient

Consider the instruction "add a FIFO between the scanner and unifier." A FIFO description alone omits the questions that decide correctness:

- Does the FIFO preserve order per query or globally?
- What identifies a cancelled query?
- What happens when it fills while the upstream engine holds a partially applied binding?
- Are faults entries in the same stream?
- Can reset discard accepted but uncommitted work?
- Is the downstream `ready` allowed to form a combinational path back to the scanner?
- Does the FIFO store the tag, epoch, authority, and poison bits with the payload?

The useful reusable unit is therefore not "FIFO." It is an **Elastic Channel** or **Credit-Based Backpressure** contract, plus the representation and cancellation contracts carried through it.

## A first composition

A reversible logic-variable binding can be written as:

```text
bind(variable, value):
    old = binding[variable]
    append trail(variable, old)
    binding[variable] = value
```

This small transition already composes four patterns:

1. Tagged Value Word identifies a variable reference.
2. Self-Reference Sentinel identifies an unbound variable.
3. Mutation Trail records the old value.
4. Choice-Point Snapshot supplies the trail boundary to which failure will unwind.

Add a Commit / Cut Fence and the same state can eventually become irreversible. Add Context / Epoch-Carrying Tokens and several searches can share a unifier. Add a Delayed Irreversible Store and results can be safely sent to a host. Architecture emerges through composition.

## Chapter summary

Symbolic computers make representation, control, lifetime, and reversibility explicit. Their mechanisms can be generalized into reusable state-transition shapes. A useful hardware pattern must specify more than structure: it must state ordering, visibility, recovery, capacity, timing, and invariants.

## Exercises

1. For a conventional five-stage RISC pipeline, list at least five states that are microarchitecturally provisional but not architectural.
2. Compare a software future and a Tomasulo producer tag. Which contract fields are similar, and which differ?
3. Choose one source-language feature such as exceptions, generators, pattern matching, or transactions. Write three possible implementation levels for it: software sequence, abstract instruction, and dedicated engine.
4. Identify an example in which a FIFO preserves data but still violates semantics because it loses metadata.

![The book moves from explicit semantics through pattern contracts to reusable RTL and measured FPGA experiments.](assets/cover_architecture.png){ width=92% }

# Abstract machines and refinement

## Learning objectives

You should be able to define a semantic state vector, write atomic transitions, distinguish architectural state from implementation state, and construct a refinement relation between a reference model and RTL.

## The abstract machine is the executable specification

An abstract machine is a state-transition system:

\[
M = \langle X, \rightarrow, O \rangle
\]

where \(X\) is the set of legal states, \(\rightarrow\) is the transition relation, and \(O\) is the observation function visible to software or the environment. A hardware realization may take many clock cycles to implement one abstract transition. It may also perform internal steps that leave \(O\) unchanged. Those steps are *stuttering* with respect to the abstract machine.

The abstract machine should be small enough to reason about but complete enough to determine every visible behavior. For a tagged stack evaluator, a suitable state is:

```text
M_stack = <pc, stack, depth, output_stream, fault, halted>
```

The implementation may contain instruction registers, a top-of-stack cache, RAM read pipelines, an ALU result register, and FIFO occupancy. These are refinement state, not necessarily architecture.

For a search machine:

```text
M_search = <domains, choices, trail, trail_top, output_stream, status>
```

For a graph reducer:

```text
M_graph = <control, heap, continuation_stack, result, fault>
```

The key design move is to name semantic state before drawing the datapath.

## Atomic transitions

An abstract transition groups all changes that must be observed together. Consider typed addition:

```text
ADD:
    require depth >= 2
    if stack[-1].tag == INT and stack[-2].tag == INT:
        stack[-2] = INT(stack[-2].payload + stack[-1].payload)
        pop stack[-1]
        pc = pc + 1
    else:
        fault = TYPE_FAULT(pc, stack[-2].tag, stack[-1].tag)
        leave pc and stack unchanged
```

A multi-cycle RTL controller may read one operand, read another, compute, and then write. The abstract transition does not complete until the final write and `pc` update commit together. A fault detected in the middle must either restore all provisional changes or prevent them from becoming visible.

Atomicity is not the same as single-cycle execution. It is a statement about observation.

## Refinement maps

A refinement map \(R\) converts implementation state \(H\) into abstract state \(M\):

\[
R(H) = M
\]

For the stack evaluator, `R` may concatenate two top-of-stack registers with the live range of a deeper stack RAM. It may ignore prefetched instructions and invalid pipeline registers. A correct implementation must satisfy two obligations:

1. **Safety:** every committed implementation action corresponds to a legal abstract transition.
2. **Progress:** under stated environmental assumptions, an enabled abstract transition is not postponed forever.

A practical co-simulation test does not need a theorem prover to exploit this structure. Record one *commit packet* whenever RTL completes an abstract transition:

```text
commit_packet = {
    sequence,
    opcode,
    old_pc,
    new_pc,
    state_digest,
    visible_event,
    fault
}
```

Apply the same operation to a software model and compare the projected states at each packet.

## Transition families versus source constructs

A source construct is often too large or unstable to be an architectural unit. For example, a high-level pattern match may compile into tag tests, field reads, equality checks, bindings, and branches. A stable transition family is more useful:

```text
DEREF      follow variable references to a terminal term
BIND       make one reversible variable binding
MATCH      compare a value with a constructor or atom
CONSTRUCT  allocate and initialize one node shape
FORCE      demand a thunk and update it once
COMMIT     advance a semantic visibility frontier
```

These transitions can be encoded as bytecodes, internal operations, or engine requests. The instruction format can change while the semantic model remains stable.

## Nondeterminism belongs in the model

Hardware descriptions often hide scheduling nondeterminism by assuming one convenient timing. A compositional model makes permitted nondeterminism explicit. An elastic dataflow engine may emit independent contexts in either order. Its observation function therefore compares results by `(context, epoch, destination)` rather than by wall-clock cycle. Within one context, however, the graph may require deterministic single-assignment behavior.

Likewise, a parallel relational engine may be allowed to return solutions in any order or may promise database order. This choice changes the scheduling and buffering contract. It cannot be postponed until verification.

## Exceptions and resource exhaustion

A complete abstract machine includes finite-resource outcomes. `TRAIL_FULL`, `CHOICE_FULL`, `HEAP_FULL`, and `RESULT_BLOCKED` are not implementation embarrassments; they are part of the pressure contract. The design must choose among:

- stall before changing state;
- spill into another region;
- trap to software;
- restart from a named checkpoint;
- reject an input before accepting it; or
- provide an explicitly lossy interface.

An abstract machine with infinite stacks can still be useful for mathematical semantics, but the FPGA implementation needs an additional bounded refinement layer that defines overflow precisely.

## Worked micro-example: precise `EMIT`

Suppose `EMIT` pops the top value and sends it to a ready/valid output. The correct transition is:

```text
if out_ready:
    out_valid = 1 for the accepted transfer
    out_data  = stack[-1]
    pop stack
    pc = pc + 1
else:
    hold pc, stack, and pending output stable
```

A common incorrect implementation pops when it asserts `out_valid`, then waits for `out_ready`. If the channel stalls, the producer no longer owns the original value and may duplicate or lose it. The abstract transition reveals the bug immediately: acceptance, not assertion, is the commitment event.

## Chapter summary

The abstract machine is a compact executable specification. Atomic transitions define what must become visible together. Refinement maps allow multi-cycle and speculative implementations to be compared at commit points. Finite-resource outcomes and permitted nondeterminism belong in the model.

## Exercises

1. Write an abstract transition for a stack `DIV` instruction that reports divide-by-zero precisely.
2. Define a refinement map for a four-entry top-of-stack cache backed by RAM.
3. State whether the order of solutions is observable in a search accelerator of your choosing.
4. Add `HEAP_FULL` to a bump allocator without allowing a partially initialized object to become reachable.

# The hardware pattern contract

## Learning objectives

You should be able to fill out the nine fields of a pattern contract, distinguish intent from implementation, and use the invariant as the central design and verification object.

![A reusable pattern is defined by nine linked obligations, not by a block diagram alone.](assets/pattern_contract.png){ width=92% }

## The nine fields

A pattern is represented as

\[
P = \langle S,D,F,O,V,R,B,T,I \rangle.
\]

The fields are:

| Field | Question |
|---|---|
| \(S\), state | Which registers, memories, pointers, counters, and ownership bits does the pattern own? |
| \(D\), data | What payload and metadata encodings cross its boundary? |
| \(F\), flow | When is an item accepted, held, rejected, retried, or cancelled? |
| \(O\), ordering | Which events must preserve program, context, address, or age order? |
| \(V\), visibility | At which commitment level can each effect be observed? |
| \(R\), recovery | How are faults, replay, rollback, reset, and partial progress handled? |
| \(B\), bounds | What finite structures exist, and what happens at capacity? |
| \(T\), timing | What latency, throughput, port, clock, locality, and fanout assumptions are made? |
| \(I\), invariant | What must always hold for the pattern to be correct and replaceable? |

A module interface can be derived from these fields, but the fields cannot reliably be reconstructed from a port list. `valid`, `ready`, and `data` say nothing by themselves about ordering, epochs, or reset semantics.

## Example contract: one-entry elastic register

**State \(S\).** One occupancy bit `full` and one payload register `data_q` containing the complete payload-plus-metadata record.

**Data \(D\).** An uninterpreted `WIDTH`-bit item. If tags, epochs, errors, or capabilities exist, they are inside that item.

**Flow \(F\).** Input is accepted on `in_valid && in_ready`. Output is transferred on `out_valid && out_ready`. A full register may simultaneously transfer its old item and accept a new one.

**Ordering \(O\).** FIFO order for all accepted items.

**Visibility \(V\).** Acceptance into the register does not by itself commit the semantic operation represented by the item.

**Recovery \(R\).** Reset discards occupancy. Therefore reset is permitted only where discarding buffered work is architecturally legal or coordinated with a higher-level restart.

**Bounds \(B\).** Capacity is exactly one item. Full propagates backpressure.

**Timing \(T\).** `in_ready` may depend combinationally on `out_ready`. Composition must avoid a zero-delay ready loop.

**Invariant \(I\).** Each accepted item is emitted exactly once and in order, unless a specified reset or cancellation operation invalidates it.

## Intent, forces, and consequences

The contract states correctness. A pattern description also explains why the contract is useful.

- **Intent** names the recurring design objective.
- **Problem** describes the failure of a naive design.
- **Forces** identify competing objectives such as latency versus state size, or locality versus flexibility.
- **Consequences** state what the pattern makes easier and what it makes more expensive.

For example, a Mutation Trail replaces whole-state copying with proportional logging. It is attractive when few old locations change after a checkpoint. It is less attractive when nearly every location changes or when random rollback points are required. Its correctness invariant is independent of that performance tradeoff.

## Pattern, primitive, and idiom

This book distinguishes three levels:

- A **primitive** is a technology-specific element such as a GateMate block RAM or CPE configuration.
- An **idiom** is a local coding practice such as registering a RAM address or holding `data` stable while stalled.
- A **pattern** is an architectural contract that can admit several primitives and idioms.

The same Elastic Channel can be implemented by registers, a small FIFO, a block-RAM FIFO, or an asynchronous queue. The same Choice-Point Snapshot can be a register stack, BRAM, host memory, or compressed circular buffer. Conversely, one block RAM may simultaneously participate in several patterns by storing tags, checkpoints, and queue records in different partitions.

## Pattern cards as review artifacts

A completed pattern card is useful in design reviews because it exposes unstated assumptions. Reviewers should challenge the card with questions such as:

- Which clock edge transfers ownership?
- Can a fault occur after an externally visible effect?
- Which generation prevents an old response from matching a new request?
- What prevents two writers from updating the same logical cell?
- How is a full queue reflected upstream?
- Can this broadcast meet timing after the design is clustered?
- Can debug access bypass the metadata checks?

The card should evolve with the RTL. It is not documentation written after implementation.

## Chapter summary

The nine-field contract turns familiar mechanisms into composable architectural units. The invariant is the anchor; flow, visibility, recovery, capacity, and timing prevent local correctness from becoming global failure.

## Exercises

1. Fill out all nine fields for a synchronous single-port RAM wrapper.
2. Fill out the contract for a pulse-based interface. Identify why it composes poorly with backpressure.
3. Write a contract for a two-way arbiter and state whether fairness is safety or liveness.
4. Review an existing RTL module and list three assumptions present in the code but absent from its interface.

# Commitment, time, and reversibility

## Learning objectives

You should be able to distinguish microarchitectural, architectural, semantic, and external commitment; assign a rollback owner; and prevent irreversible effects from crossing the wrong boundary.

![A symbolic processor may have several commitment boundaries beyond ordinary instruction retirement.](assets/commitment_ladder.png){ width=92% }

## The four levels

The following levels form a useful design vocabulary:

1. **Microarchitecturally provisional, \(\mu\).** Work may still be squashed because of a branch, replay condition, transient conflict, or exception.
2. **Architecturally retired, \(A\).** The instruction or abstract operation belongs to the sequential machine state.
3. **Semantically committed, \(S\).** A language-level alternative, transaction, search branch, thunk update, or workflow can no longer reverse the effect.
4. **Externally visible, \(E\).** Another protection domain, processor, device, file, network endpoint, or user can observe the effect.

Traditional precise-exception mechanisms concentrate on \(\mu \rightarrow A\) [SMP88]. Logic programming demonstrates that an operation can be architecturally executed yet semantically reversible: a variable binding can exist while the search still retains alternatives [WAR83]. Device and host interfaces add \(S \rightarrow E\).

The levels are conceptual, not necessarily separate pipeline stages. A simple in-order evaluator may collapse \(\mu\) and \(A\). A deterministic computation with no backtracking may collapse \(A\) and \(S\). An internal result written to private RAM may never reach \(E\) until a later host command.

## Every mutation declares a frontier

For every mutable location, record:

```text
owner              which component may write it
current level      μ, A, S, or E
rollback owner     checkpoint, ROB, trail, transaction, or none
commit event       the condition that advances irreversibility
external path      the queue or gate through which it becomes visible
```

Consider a logic-variable binding in an out-of-order symbolic core. Before retirement, the ROB or checkpoint owns rollback. After retirement but before the active search branch commits, the semantic trail owns rollback. After cut, the binding is semantically permanent. If it is serialized to a host, a result queue controls external visibility.

A design bug arises if both the ROB and trail independently believe they own the same physical write. One may restore an older value after the other has already reassigned or reclaimed the entry. The solution may be staged logging, versioned entries, or buffering the binding until retirement. What matters is explicit ownership.

![Rollback mechanisms may be nested, but their authority over a mutation must not overlap ambiguously.](assets/rollback_ownership.png){ width=92% }

## Three recovery organizations

Precise processors illustrate three general arrangements [SMP88]:

- **Commit queue:** keep results provisional and expose them in order.
- **History buffer:** update a working state early while logging overwritten values.
- **Future or shadow file:** maintain separate provisional and committed images.

The same choices appear at higher semantic levels. A search engine can buffer bindings until a branch succeeds, update bindings and trail old values, or maintain persistent versions. A graph reducer can stage a thunk result, overwrite the node with an undo record, or hold old and new node versions. The best choice depends on read latency, write frequency, checkpoint density, and state size.

## Irreversible effects

An effect is irreversible at a given level when the local mechanism cannot withdraw or compensate it. Examples include:

- an accepted ready/valid transfer to a consumer that does not support cancellation;
- a device command;
- a message observed by another context;
- a persistent-memory write;
- a result printed or returned to a host;
- a speculative cache or predictor effect observable through a side channel.

The Delayed Irreversible Store pattern holds these effects behind a queue or shadow structure until the required frontier is crossed. Speculation Shadow Structures extend the principle to microarchitectural state [KHA19].

## Compensation is not rollback

Some distributed systems cannot prevent an effect from escaping. They issue a compensating action instead. Compensation belongs in the recovery contract, but it is weaker than rollback: an observer may see both the original and compensating actions. Hardware experiments in this book avoid this ambiguity by delaying all host-visible results until acceptance by a commit FIFO.

## Worked ownership table

For the capstone relational engine:

| State | First visible level | Rollback owner | Commit event |
|---|---|---|---|
| scanner index | \(\mu\) | context checkpoint | match accepted |
| variable binding | \(A\) | mutation trail | cut or search completion |
| newly allocated term | \(A\) | heap-top checkpoint | branch commit |
| solution tuple in local register | \(S\) | result formatter | FIFO acceptance |
| tuple received by host | \(E\) | none locally | external handshake |

The table guides both RTL structure and waveform interpretation.

## Chapter summary

Instruction retirement is only one commitment boundary. Symbolic and managed machines often need semantic and external frontiers as well. Every mutable effect must have one active rollback owner and one named commit event at each relevant level.

## Exercises

1. Classify branch-predictor updates at the four commitment levels under a security-sensitive threat model.
2. Draw an ownership timeline for a store issued by an out-of-order core inside a software transaction.
3. Explain why an output FIFO can be a semantic boundary even when it is entirely on-chip.
4. Propose a versioned logging scheme that safely hands rollback ownership from a ROB to a trail.

# Composition calculus

## Learning objectives

You should be able to test two pattern contracts for closure, recognize common composition cycles, and derive system-level invariants from local ones.

![Composition succeeds only when representation, flow, order, visibility, recovery, capacity, physical assumptions, and faults remain closed.](assets/composition_closure.png){ width=92% }

## Representation closure

Every output state from one pattern must be meaningful to the next. This includes more than ordinary values. A consumer must understand invalid, empty, poison, error, capability, epoch, forwarding, and cancellation states. If the producer emits a tagged value and the FIFO stores only the payload, composition fails even though the FIFO is bit-correct for its declared width.

A useful rule is:

```text
payload and semantic metadata form one indivisible record
```

This rule extends through register bypasses, block RAMs, serialization, DMA, debug access, and context switching.

## Flow closure

A composition must be able to stop without losing work. Ready/valid stages must hold data stable. Credit counts must be conserved. There must be no combinational loop such as:

```text
A.ready depends on B.ready
B.ready depends on C.ready
C.ready depends on A.ready
```

and no resource dependency cycle such as a full completion FIFO preventing the only unit that can drain a request FIFO from running.

For each channel, write:

```text
accept condition
hold condition
cancel condition
drain condition
reset condition
```

Then examine cycles in the wait-for graph.

## Ordering closure

Ordering guarantees may be global, per context, per address, per destination, or absent. Composition is legal only if the combined path preserves every required order. An arbitration network that reorders contexts may be valid for a dataflow engine but invalid for a database-order relational query. A nonblocking cache can reorder responses if tags restore the intended association, but a tagless consumer may require in-order response.

## Visibility closure

Provisional effects must not bypass their commitment gate. Typical bypass paths include debug writes, performance counters shared across domains, cache fills, direct memory access, exception reporting, and a "fast" result path that skips the commit FIFO. Visibility closure requires auditing every path, not just the nominal datapath.

## Recovery closure

Rollback domains must be disjoint, nested, or coordinated. A younger checkpoint may unwind to an older trail mark, but an older mechanism must not reclaim storage still referenced by a younger one. Epoch cancellation must cover every delayed response, including error replies. Replay must not repeat an externally visible operation.

A system-level recovery invariant is:

> After any legal rollback, the refined abstract state equals a state that existed at the named checkpoint, and no invalidated epoch can later modify live state.

## Capacity closure

For every finite structure, define:

- capacity;
- reservation point;
- full detection;
- upstream pressure;
- maximum number of entries an accepted operation may still require; and
- recovery from overflow.

The reservation point matters. A unification operation that accepts a term pair and can generate two trail writes must not begin when only one trail slot remains unless it can stop at an internal precise point.

Capacity closure often requires **credits for compound operations**. Reserve the worst-case resources before acceptance or split the operation into transitions, each of which reserves only what it can consume atomically.

## Physical closure

A logically composable broadcast may not be physically composable. Tag wakeup across eight entries can become a timing and power problem across hundreds. A single global heap port can serialize a multi-engine machine. A ready chain can become the critical path. Physical closure asks whether placement, fanout, wire distance, RAM ports, and clock assumptions remain realizable after composition.

Hierarchies are common refinements:

- cluster reservation stations and wakeup domains;
- bank memory by access function or context;
- replace broadcast with explicit destination lists;
- insert elastic stages on long routes;
- convert combinational backpressure to credits.

## Fault closure

Every partially executed transition must be able to complete, roll back, or report a precise progress point. A fused assist that allocates three nodes and faults on the fourth needs either an allocation checkpoint, a log, or an internal index that software can resume safely. Fault closure also includes metadata faults: a malformed tag must not be interpreted as a valid pointer merely because the slow path is unavailable.

## Eight composition laws

The practical laws used throughout the book are:

1. Metadata is part of the value.
2. Every finite structure has a pressure contract.
3. Every effect has a visibility level.
4. Every rollback domain has one active owner.
5. Every indirection has a termination policy.
6. Every broadcast has a physical scope.
7. Compiler promises are checked or dispensable.
8. Every slow path begins from a named precise state.

## Worked composition review

Suppose a pipelined multiplier is inserted into the dataflow engine. The local interface is ready/valid. A complete review asks:

- **Representation:** does the result retain context, epoch, destination, tags, and fault state?
- **Flow:** can the output stall without corrupting pipeline registers?
- **Ordering:** are results ordered, and does the consumer rely on that order?
- **Visibility:** is a result provisional until routed into the destination slot?
- **Recovery:** are old-epoch results discarded at every exit?
- **Capacity:** how many outputs can be in flight when the completion FIFO becomes full?
- **Physical:** does the pipeline meet the target clock and avoid a ready chain?
- **Fault:** how are multiplication overflow and bad tags represented?

The multiplier is composable only after all eight answers are explicit.

## Chapter summary

Composition is the preservation of contracts across boundaries. Local correctness is necessary but insufficient. System design must prove closure for representation, flow, ordering, visibility, recovery, capacity, physical assumptions, and faults.

## Exercises

1. Analyze a trail stack connected to a backtrack engine for all eight closure dimensions.
2. Find a deadlock cycle in a system with a request FIFO, worker, response FIFO, and host that stops consuming while waiting for a completion interrupt generated by the worker.
3. Explain how explicit destination lists replace broadcast while preserving result-wakeup semantics.
4. Define compound-operation credits for an instruction that may allocate two heap nodes and append one trail record.

# GateMate as an experimental substrate

## Learning objectives

You should be able to map the textbook's records and storage structures to GateMate resources, establish conservative resource budgets, and define a repeatable synthesis-and-measurement loop.

## Conservative target

The laboratories use the GateMate CCGM1A1 as a conservative minimum target. The August 2025 datasheet lists 20,480 configurable processing elements, 40,960 flip-flops or latches, and 32 physical 40-Kbit block RAMs; each physical block can instead operate as two independent 20-Kbit memories [COL25-D]. The CPE fabric supports LUT-tree functions, fast arithmetic structures, and composable small multipliers. Larger or later devices may provide more resources, but the experiments should remain parameterized.

The book deliberately avoids claiming a clock frequency. Frequency depends on the exact RTL, memory configuration, board constraints, toolchain version, placement, routing seed, and inserted debug logic. The correct target is obtained through measured iteration.

![The 40-bit value and 80-bit event conventions align with useful GateMate block-RAM organizations without becoming architectural requirements.](assets/gatemate_mapping.png){ width=92% }

## Common physical records

A convenient tagged value is:

```text
value40
+-----------+-----------+--------------------------------+
| tag[3:0]  | flags[3:0]| payload[31:0]                  |
+-----------+-----------+--------------------------------+
```

An event record can be:

```text
event80
+--------------------------------------------------------+
| routing, context, epoch, destination, control [39:0]  |
+--------------------------------------------------------+
| value40                                                |
+--------------------------------------------------------+
```

These widths permit compact memories and queues in native 20-, 40-, or 80-bit organizations. They are not universally optimal. A design needing 48-bit pointers or richer capabilities should change the record and measure the resulting RAM packing.

## Block RAM discipline

The important architectural fact is synchronous memory behavior, not a particular primitive name. Every memory wrapper should expose a request and a response-valid point. Design logic must not depend on a behavioral array producing combinational read data.

Prefer inferred memories first, then inspect synthesis statistics and the technology netlist. Instantiate a GateMate-specific primitive only when inference cannot express required width, mode, initialization, byte mask, or packing. Keep the technology-specific wrapper below the pattern contract.

Assign one logical mutation owner per structure even when the physical RAM is dual-port. Independent scanners may read through separate ports, but heap update, binding write, trail push, and checkpoint allocation should each have arbitration that defines simultaneous access.

## Stop-build budgets

A stop-build budget is a threshold at which the experiment is simplified before implementation becomes opaque. It is not a forecast.

| Experiment | Principal result | RAM budget | CPE budget |
|---|---|---:|---:|
| tagged evaluator | typed result and precise fault | 2 physical blocks | 2,000 |
| rollback solver | enumerate 8-queens solutions | 2 | 3,000 |
| elastic dataflow | interleaved expression DAGs | 8 | 8,000 |
| lazy reducer | shared thunk evaluated once | 6 | 8,000 |
| relational engine | small Prolog-like query | 16 | 12,000 |

If a design exceeds its threshold, first inspect accidental register arrays, width expansion, replicated comparators, broad resets, inferred asynchronous reads, and debug capture. Then revisit the architecture.

## Toolchain loop

The May 2026 GateMate toolchain guide describes a flow built around Yosys, nextpnr-himbaechel, `gmpack`, and programming tools such as `openFPGALoader` [COL26-T]. Tool names and options evolve; reconcile the commands in Appendix C with the locally installed guide and `--help` output.

A repeatable experiment records:

```text
Git commit or source archive
Yosys and nextpnr versions
synthesis command
place-and-route command and seed
board constraint file
resource report
critical path report
simulation seed
reference-model digest
hardware trace digest
```

Without this record, a faster or smaller result may be caused by tool drift rather than an architectural change.

## Instrumentation

The official GateMate integrated logic analyzer can capture internal waveforms at runtime and uses FPGA resources, including block RAM [COL25-I]. Leave headroom for it. The book's designs also expose a compact trace interface so that debug can be implemented by the official ILA, a dedicated trace RAM, or a host stream.

A standard trace record should include:

```text
cycle or sequence
state
context and epoch
operation
accept / commit / rollback
fault
queue occupancies
key addresses or tags
```

Capture events rather than every wide datapath signal whenever possible. Event traces are easier to compare with the abstract model and consume less memory.

## Experimental ladder

![Each laboratory adds one difficult contract while reusing the previously verified substrate.](assets/experiment_ladder.png){ width=92% }

The recommended order is strict:

1. precise tagged transitions and committed output;
2. checkpoints, trails, and rollback;
3. elastic communication, readiness, and epochs;
4. graph mutation, indirection, and ownership;
5. full relational composition.

Do not proceed merely because the demonstration works once. Proceed when randomized stalls, capacity edges, reset rules, faults, and model comparison pass.

## Chapter summary

GateMate provides a useful small-FPGA target with RAM widths that suit compact symbolic records. The design method remains technology-independent: use synchronous interfaces, preserve one mutation owner, budget finite resources, synthesize early, and record every experimental condition.

## Exercises

1. Pack a 40-bit value stack and a 20-bit bytecode ROM into the available RAM modes. Identify which ports each operation requires.
2. Propose a 60-bit event format and compare its packing with the 80-bit convention.
3. List five ways that adding an ILA can change timing or memory availability.
4. Define a machine-readable experiment manifest containing all reproducibility data listed above.

# Part II - The Pattern Catalog

![The catalog is divided by architectural obligation rather than by traditional CPU block.](assets/pattern_families.png){ width=92% }

# Semantic and refinement patterns

## Learning objectives

After this chapter, you should be able to recognize the family’s recurring problem, fill out its pattern contracts, and select compatible patterns for a measured bottleneck.

Patterns 1-5 define the semantic reference model and safe specialization rules.

## Pattern 1: Abstract-Machine Contract

**Intent.** Stabilize observable semantics before selecting instructions, pipelines, or physical storage.

### Context and problem

RTL designed directly from a source-language feature tends to mix semantic obligations with one implementation. Later pipelining, microcoding, or software fallback becomes difficult because no implementation-independent transition exists.

**Forces.** The model must be small enough to execute and verify, yet complete for faults, finite resources, nondeterministic order, and visible effects.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | A named semantic state vector, legal transition relation, observation function, and optional resource-status state. |
| Data and flow \(D,F\) | Requests identify transitions and operands; commit records identify completed transitions and visible events. |
| Ordering \(O\) | The model declares which transitions are sequential and which independent transitions may commute. |
| Visibility \(V\) | Visibility is defined by the observation function rather than by internal clock edges. |
| Recovery \(R\) | A failed implementation transition either leaves the abstract state unchanged or maps to an explicit fault transition. |
| Bounds \(B\) | The mathematical model may be unbounded, but the hardware refinement must add explicit stack, queue, heap, and log limits. |
| Timing and locality \(T\) | No physical latency is required. A realization may stutter for any finite number of cycles permitted by its progress assumptions. |
| Invariant \(I\) | **Every committed hardware step refines zero or one legal abstract transition, and every accepted operation eventually commits, faults, or is explicitly cancelled.** |

### Operation

Write the state tuple; define preconditions and atomic post-state for each operation; build a software interpreter; make RTL emit commit packets that the interpreter checks.

### Composition

All other patterns. It is especially important with Semantic Transition Family, Precise Commit Queue, and Common Fast Path / Precise Slow Path.

### GateMate realization

Keep the model outside vendor-specific code. Use compact state digests and event records so the FPGA can expose refinement checkpoints through a trace RAM or ILA.

### Verification target

Run lockstep comparison at commits, randomize stalls, and prove that no state-changing RTL path exists without a corresponding commit or rollback event.

### Characteristic failure sign

The testbench compares final answers only; partial writes and duplicate outputs remain invisible until hardware deployment.

### Miniature example

The stack evaluator models `ADD` as check-and-replace, allowing a multi-cycle BRAM implementation to fault without popping its operands.

**Lineage.** [LAN64], [WAR83], [KOG91]

## Pattern 2: Explicit Semantic State Vector

**Intent.** Turn control conventions, continuation state, and lifetime boundaries into inspectable named state.

### Context and problem

Hidden state in call stacks, global modes, implicit program-counter conventions, and ad hoc controller phases makes checkpointing and refinement ambiguous.

**Forces.** Explicit state improves verification and migration, but exposing every field architecturally can constrain later implementations.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Named components such as `pc`, environment, operand stack, continuation, heap top, trail top, choice pointer, mode, epoch, and fault. |
| Data and flow \(D,F\) | Transitions declare read and write sets over the components. Implementation-only pipeline state remains outside the vector. |
| Ordering \(O\) | Dependencies follow declared read/write sets; independent components may be banked or updated in parallel. |
| Visibility \(V\) | Only selected components belong to the external observation. Others can remain semantic but not software-addressable. |
| Recovery \(R\) | A checkpoint records sufficient components to reconstruct the named state; omitted components must be recomputable or younger than the checkpoint. |
| Bounds \(B\) | Each component has a legal range and an exhaustion status. Pointer wraparound is never silently accepted. |
| Timing and locality \(T\) | The vector is conceptual. Components may reside in registers, RAM, queues, or distributed tokens. |
| Invariant \(I\) | **At every abstract boundary, each semantic component has exactly one well-defined value independent of transient implementation state.** |

### Operation

List components, ownership, lifetime, visibility level, and update transitions. Partition only after the complete vector is coherent.

### Composition

Choice-Point Snapshot, Explicit Continuation Record, Checkpoint plus Epoch Squash, and Semantic Region Partition.

### GateMate realization

Expose selected components to the ILA and compute a state hash at commit. Avoid resetting large inferred RAMs merely to make the vector appear fully initialized.

### Verification target

Assert legal ranges, pointer monotonicity within epochs, and equality between reconstructed state and the software model.

### Characteristic failure sign

A mode bit lives in a controller latch but is not saved in a checkpoint; replay resumes in the wrong interpretation phase.

### Miniature example

The relational engine makes scan stage, fact index, binding base, choice top, trail top, context, and epoch explicit.

**Lineage.** [LAN64], [WAR83], [KOG91]

## Pattern 3: Structure / Representation Firewall

**Intent.** Separate an object’s abstract structure and operations from its current bit layout.

### Context and problem

A design that lets every consumer decode object headers, pointer compression, forwarding state, and storage-bank details becomes impossible to retarget or collect safely.

**Forces.** Direct field access is locally fast; a firewall adds interfaces and may require a fast path plus assist. The long-term benefit is replaceable representation.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Representation descriptors, access functions, optional caches, and a small set of constructor, selector, equality, and update operations. |
| Data and flow \(D,F\) | Consumers request semantic operations such as `term_functor`, `child(i)`, `environment_parent`, or `resolve_reference` rather than indexing raw words. |
| Ordering \(O\) | The firewall preserves the ordering required by object mutation and may allow independent reads to reorder. |
| Visibility \(V\) | Internal forwarding and compressed forms are hidden. The consumer observes the resolved abstract object or a precise error. |
| Recovery \(R\) | If a representation is uncommon, malformed, moved, or protected, the firewall takes a precise assist without exposing partial decode state. |
| Bounds \(B\) | Caches and outstanding access records are finite and exert backpressure; maximum indirection depth is explicit. |
| Timing and locality \(T\) | Fast structural cases may be combinational or pipelined. Long indirections and descriptor reads use elastic responses. |
| Invariant \(I\) | **Constructor, selector, equality, and update behavior is invariant under any legal representation change.** |

### Operation

Define the abstract object API first; centralize tag/header decode; preserve a software or microcoded path for uncommon shapes.

### Composition

Tagged Value Word, Indirection / Forwarding Cell, Shape-Specialized Encoding, Capability / Descriptor Gate, and Incremental Collection Barrier.

### GateMate realization

Place common tag checks near the RAM output. Use one resolver pipeline rather than duplicating deep decode across every engine.

### Verification target

Generate several physical encodings for the same object graph and compare all semantic access results and faults.

### Characteristic failure sign

A collector installs a forwarding cell that one execution unit understands while another treats it as an ordinary payload.

### Miniature example

A pair can be headerless in the heap while the unifier still uses the same `field(0)` and `field(1)` requests.

**Lineage.** [KOG91], [JON92], [WAT19]

## Pattern 4: Common Fast Path / Precise Slow Path

**Intent.** Make dominant cases shallow and local while retaining complete semantics through a restartable assist.

### Context and problem

A uniform general datapath wastes cycles on frequent simple tags and shapes, but an unchecked fast path often corrupts state before discovering an uncommon case.

**Forces.** The fast path wants early mutation and narrow checks; precision requires delaying commitment or recording exact progress.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Fast-path pipeline state, slow-path request record, progress marker, and fault or assist return state. |
| Data and flow \(D,F\) | The input is classified before irreversible mutation. The slow-path request carries original operands and the exact continuation point. |
| Ordering \(O\) | Fast and slow paths must appear in the same abstract operation order even if their latency differs. |
| Visibility \(V\) | Neither path exposes partial results. Both cross the same architectural or semantic commit gate. |
| Recovery \(R\) | A miss, unusual tag, overflow, protection failure, or structural depth transfers to an assist from a named precise state. |
| Bounds \(B\) | Assist queues are finite. When full, the classifier stalls before consuming the operation. |
| Timing and locality \(T\) | Fast latency is optimized; slow latency is elastic. The classifier fanout and bypass to commit are part of timing closure. |
| Invariant \(I\) | **For every input, fast and slow execution either produce the same abstract result or the same defined fault.** |

### Operation

Profile cases; define a side-effect-free predicate for the common case; implement a single commit point shared by direct and assisted results.

### Composition

Type / Shape Dispatch, Fused Semantic Assist with Decomposition Escape, Capability / Descriptor Gate, and Semantic Transition Family.

### GateMate realization

Use CPE logic for common tag predicates and a small request FIFO for assists. Avoid duplicating wide operands if a stable operand store can be referenced.

### Verification target

Force every boundary condition into the slow path, compare with unfused software execution, and stall the assist queue randomly.

### Characteristic failure sign

The fast path pops the stack before detecting a boxed integer, leaving the slow path no precise input state.

### Miniature example

Small tagged integer addition executes directly; overflow or boxed operands invoke an assist that sees the untouched original operands.

**Lineage.** [KOG91], [PAT81]

## Pattern 5: Checked Compiler Hint

**Intent.** Exploit compiler knowledge without making correctness depend on an infallible compiler.

### Context and problem

Compilers know liveness, shape, non-aliasing, determinacy, expected tags, and escape behavior, but silently trusting those claims expands the trusted computing base.

**Forces.** Runtime checking costs area and latency; ignoring useful facts wastes specialization opportunities. Some hints can be probabilistic, others require proof.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Hint fields, optional validation state, fallback selector, and diagnostic counters. |
| Data and flow \(D,F\) | The operation carries a hint or references metadata. Hardware verifies it, treats it as advisory, or enters a safe fallback. |
| Ordering \(O\) | Hints may affect scheduling but cannot weaken required semantic order or authority checks. |
| Visibility \(V\) | A wrong hint may change performance and diagnostics, never the visible result or protection decision. |
| Recovery \(R\) | Validation failure redirects to a general path from the original input state; repeated failure may disable the hint source. |
| Bounds \(B\) | Hint tables and validation queues are bounded. Missing entries default to the general path. |
| Timing and locality \(T\) | Checks should be placed where required data already exists; wide global hint tables can destroy locality. |
| Invariant \(I\) | **Removing all hints leaves a correct, perhaps slower, implementation.** |

### Operation

Classify each compiler claim as checked, advisory, proof-carrying, or trusted; provide a no-hint mode used continuously in verification.

### Composition

Liveness-Driven Frame Trimming, Lifetime Promotion on Escape, Shape-Specialized Encoding, and Common Fast Path / Precise Slow Path.

### GateMate realization

Encode short hints in instructions or descriptor RAM. Count successful and failed predictions to justify their storage cost.

### Verification target

Corrupt hints in simulation and require identical commit traces; prove authority and bounds checks do not depend on advisory bits.

### Characteristic failure sign

A `no_alias` hint suppresses a required ordering check and changes the program result when the compiler is wrong.

### Miniature example

A bytecode marks a local value as dead; hardware verifies that no active continuation bitmap references the slot before trimming it.

**Lineage.** [WAR83], [PAT81]

## Chapter review

The semantic and refinement patterns should not be selected by name alone. For each candidate, write its nine-field contract, identify its commitment level, and measure whether its added state and routing remove the dominant semantic cost.

## Exercises

1. Choose two patterns from this chapter and test their composition across all eight closure dimensions from Chapter 5.
2. Rewrite one pattern for an implementation with no block RAM, then for an implementation with abundant dual-port RAM.
3. Identify a pattern whose invariant is satisfied locally but can still fail globally because of missing metadata or rollback ownership.
4. Create a negative test that produces the characteristic failure sign of one pattern without violating unrelated interfaces.

# Representation and metadata patterns

## Learning objectives

After this chapter, you should be able to recognize the family’s recurring problem, fill out its pattern contracts, and select compatible patterns for a measured bottleneck.

Patterns 6-13 define what a value means and how that meaning survives movement.

## Pattern 6: Tagged Value Word

**Intent.** Carry enough metadata with every value for safe interpretation and direct dispatch.

### Context and problem

Without tags, symbolic values require out-of-band type maps or speculative interpretation. Either approach can separate meaning from the payload during movement or recovery.

**Forces.** Tags consume bits and widen storage paths, but remove loads and make polymorphic operations explicit.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Payload, tag, auxiliary flags, and optional error, generation, authority, or ownership metadata. |
| Data and flow \(D,F\) | Every transport moves the complete record. Consumers handle each legal tag or produce a precise fault. |
| Ordering \(O\) | Tags follow the same ordering as payloads; metadata updates are atomic with payload updates. |
| Visibility \(V\) | The tag determines which observations and operations are legal. Invalid tags never alias valid values. |
| Recovery \(R\) | Fault and poison tags propagate according to explicit rules; stale generations are rejected before update. |
| Bounds \(B\) | Tag space is finite. Reserved encodings and extension or boxed forms are specified before exhaustion. |
| Timing and locality \(T\) | Tag decode fanout can dominate a pipeline stage. Decode locally, cache common predicates, or predecode in operand stores. |
| Invariant \(I\) | **No observer can see a payload without the metadata required to interpret it, or new metadata paired with an old payload.** |

### Operation

Choose the minimum hot metadata needed on frequent paths; define a cold descriptor path for larger metadata.

### Composition

Immediate-or-Boxed Split, Type / Shape Dispatch, Atomic Metadata Propagation, Capability / Descriptor Gate, and Producer Tag as Future.

### GateMate realization

The textbook `value40` uses 4 tag bits, 4 flags, and 32 payload bits, matching useful RAM widths. Change it if address or capability requirements demand.

### Verification target

Inject every tag at every consumer, test transport through all queues and RAMs, and assert legal tag sets on commit.

### Characteristic failure sign

DMA or debug logic copies only payload bits; a pointer later appears as an integer or loses authority metadata.

### Miniature example

The stack ALU accepts `INT + INT`, propagates `ERROR`, and reports a precise type fault for `BOOL + INT`.

**Lineage.** [KOG91], [WAR83], [WAT19]

## Pattern 7: Immediate-or-Boxed Split

**Intent.** Represent common small values directly and exceptional-size values indirectly.

### Context and problem

Uniform heap allocation for integers, booleans, characters, and small enumerations creates avoidable traffic, allocation pressure, and pointer chasing.

**Forces.** Immediate range competes with tag and pointer bits. Multiple representations can complicate equality and arithmetic overflow.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Immediate encodings, boxed-object format, allocation path, and canonicalization policy. |
| Data and flow \(D,F\) | Consumers decode immediate values locally; overflow or uncommon forms request boxing, unboxing, or a general arithmetic path. |
| Ordering \(O\) | Box allocation and publication preserve program order where identity is observable. Pure numeric values may canonicalize independently. |
| Visibility \(V\) | The mathematical value is visible, not the incidental immediate versus boxed representation unless identity semantics explicitly expose it. |
| Recovery \(R\) | Allocation failure and overflow leave source values unchanged and take a precise slow path. |
| Bounds \(B\) | Immediate range is finite; boxed heap and allocation queues have explicit capacity. |
| Timing and locality \(T\) | Immediate arithmetic is short. Boxed access introduces RAM latency and should use elastic continuation state. |
| Invariant \(I\) | **Every represented mathematical value has one defined equality and conversion behavior across immediate and boxed forms.** |

### Operation

Select a signed immediate range; define overflow detection; preserve original operands until the box path reserves space.

### Composition

Tagged Value Word, Bump-Allocated Generation, Common Fast Path / Precise Slow Path, and Shape-Specialized Encoding.

### GateMate realization

Use fast carry logic for immediate arithmetic and reserve one tag for boxed numeric references. Measure whether wider payloads improve total RAM use.

### Verification target

Test boundary values, overflow in both directions, mixed immediate/boxed equality, and allocation pressure.

### Characteristic failure sign

The immediate adder wraps silently and produces an in-range tag with an incorrect payload.

### Miniature example

A 32-bit immediate multiply that exceeds the payload range allocates a big-integer object or raises a defined numeric-overflow fault.

**Lineage.** [KOG91]

## Pattern 8: Capability / Descriptor Gate

**Intent.** Require explicit authority, bounds, and interpretation metadata before dereference or control transfer.

### Context and problem

A raw pointer says where but not whether the holder may read, write, execute, traverse, or expose the referenced object.

**Forces.** Rich capabilities widen state and memory paths; compressed or split metadata needs extra lookups. The gate must remain on every access path.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Reference payload, validity, permissions, bounds or descriptor reference, sealing state, and provenance or generation. |
| Data and flow \(D,F\) | Loads, stores, jumps, node updates, and message sends present an operation and capability. The gate returns authorization plus translated access or a fault. |
| Ordering \(O\) | Permission checks occur before the protected effect. Revocation and descriptor updates obey the chosen memory-order model. |
| Visibility \(V\) | Only authorized data or effects cross the gate. Speculative results remain shadowed until permission is resolved. |
| Recovery \(R\) | Failure produces a precise capability fault with no protected mutation. Stale or forged metadata is rejected. |
| Bounds \(B\) | Descriptor caches and outstanding checks are finite and stall safely. Capability derivation cannot overflow into greater authority. |
| Timing and locality \(T\) | Place hot permission summaries with the value and cold bounds in local descriptor RAM when practical. Gate latency may be pipelined. |
| Invariant \(I\) | **Derived authority can remain equal or become more restrictive, but cannot increase without an explicit privileged operation.** |

### Operation

Define authority algebra; centralize all access paths; make debug, DMA, and assists use the same gate or a separately audited privileged interface.

### Composition

Tagged Value Word, Structure / Representation Firewall, Atomic Metadata Propagation, and Delayed Irreversible Store.

### GateMate realization

A compact experiment can use object-base, limit, and permission fields in descriptor RAM indexed by a tagged reference.

### Verification target

Attempt underflow, overflow, permission escalation, stale generation, sealed-object misuse, and bypass through every alternate path.

### Characteristic failure sign

The main load pipeline checks bounds, but the graph-update assist writes the same heap through an unchecked maintenance port.

### Miniature example

A graph node reference permits read and traversal but not update; only the reducer’s update capability can replace a thunk.

**Lineage.** [WAT19]

## Pattern 9: Self-Reference Sentinel

**Intent.** Encode a special state through a unique structural relation rather than a separate status field.

### Context and problem

A separate `bound` bit or empty marker costs storage and can become inconsistent with the referenced payload.

**Forces.** Structural sentinels are compact and often initialize cheaply, but they must not be confused with ordinary cycles or malformed graphs.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | A reference cell whose sentinel predicate is commonly `cell == address_of_cell`, plus optional generation metadata. |
| Data and flow \(D,F\) | Initialization writes self references. Dereference tests the sentinel before following the reference. |
| Ordering \(O\) | Sentinel initialization and use obey the same ownership and ordering as ordinary reference writes. |
| Visibility \(V\) | The sentinel denotes one abstract state such as unbound, empty, or unclaimed; consumers do not observe it as an ordinary edge. |
| Recovery \(R\) | Cycle detection distinguishes the permitted self sentinel from longer illegal cycles. Rollback restores the exact prior reference. |
| Bounds \(B\) | Reference address width bounds the represented cells. Reuse requires generation control if stale references can survive. |
| Timing and locality \(T\) | The equality test is local and cheap. Long dereference chains are separately bounded or compressed. |
| Invariant \(I\) | **The sentinel predicate is unique for every live cell and cannot be produced accidentally by an ordinary legal update.** |

### Operation

Define the predicate and cycle policy first; centralize dereference; never duplicate a weaker test in individual engines.

### Composition

Mutation Trail, Tagged Value Word, Indirection / Forwarding Cell, and Capability / Descriptor Gate.

### GateMate realization

Initialize binding RAM by construction or a reset walk rather than forcing a large synchronous clear network.

### Verification target

Test self-unbound, one-hop and long bindings, two-node cycles, rollback to unbound, and address reuse.

### Characteristic failure sign

A generic cycle detector reports the legal unbound self-reference as corruption, or a malformed two-node cycle spins forever.

### Miniature example

Logic variable `V3` is unbound when binding slot 3 contains `REF(3)`.

**Lineage.** [WAR83]

## Pattern 10: Shape-Specialized Encoding

**Intent.** Give a frequent object shape a shorter representation and access path than the fully general format.

### Context and problem

Uniform object headers and descriptor walks add bandwidth and latency to pairs, short tuples, closures, or messages that dominate the workload.

**Forces.** Special shapes consume tag space and multiply boundary cases. Too many shapes turn decode into the bottleneck.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Dedicated tag or compact header, fixed field positions, and conversion or fallback path to general objects. |
| Data and flow \(D,F\) | Constructors allocate the specialized form when predicates hold; selectors decode fields directly; uncommon sizes use the general representation. |
| Ordering \(O\) | Construction publication is atomic. Conversion preserves identity rules and field order. |
| Visibility \(V\) | Consumers observe the same abstract constructor and fields independent of specialized or general storage. |
| Recovery \(R\) | Malformed specialized objects fault precisely; migration can install an indirection or rewrite references under a collector protocol. |
| Bounds \(B\) | Tag encodings and specialized-size range are finite. The general form is the extension path. |
| Timing and locality \(T\) | Specialized decode should fit locally near RAM output. Excessive variants increase mux depth and routing. |
| Invariant \(I\) | **Constructor and selector semantics are identical for specialized and general representations.** |

### Operation

Profile shape frequency and bytes moved; specialize only when header avoidance or fixed addressing materially changes the hot path.

### Composition

Tagged Value Word, Structure / Representation Firewall, Immediate-or-Boxed Split, and Indirection / Forwarding Cell.

### GateMate realization

Pairs or binary relation facts fit naturally in one 40-bit word when field widths permit; larger terms use referenced records.

### Verification target

Round-trip construct/select for both forms, compare equality and hashing, and exercise migration between forms.

### Characteristic failure sign

A specialized list tag is understood by traversal but not by garbage-collector root scanning.

### Miniature example

A binary fact record stores relation ID and two compact arguments without a separate heap header.

**Lineage.** [WAR83], [KOG91]

## Pattern 11: Indirection / Forwarding Cell

**Intent.** Preserve stable logical identity while moving, evaluating, replacing, or promoting an object.

### Context and problem

Consumers may hold references when an object must move or when a placeholder must be replaced by its computed result.

**Forces.** Indirection avoids eagerly rewriting all references, but chains add latency and cycles require detection.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | A distinguishable forwarding or indirection tag, target reference, optional generation, and chain-compression state. |
| Data and flow \(D,F\) | Reference resolution follows targets until a terminal object, sentinel, fault, or bounded-depth condition occurs. |
| Ordering \(O\) | Publishing a forwarding cell occurs after the target is initialized and before the old storage is reclaimed. |
| Visibility \(V\) | The logical object remains the same across movement or update; internal hops are not normally visible. |
| Recovery \(R\) | Chains terminate, compress, or fault. Rollback either restores the old cell or treats forwarding as outside the rollback domain. |
| Bounds \(B\) | Maximum chain depth and resolver queues are bounded. Reclamation waits until no legal reference needs the old cell. |
| Timing and locality \(T\) | Single-hop resolution can be local; multi-hop resolution is an elastic loop with a progress counter. |
| Invariant \(I\) | **Resolution reaches the unique current representative or reports a defined cycle or stale-reference fault.** |

### Operation

Define terminal tags and cycle policy; initialize targets before forwarding; optionally perform path compression under one mutation owner.

### Composition

Demand Cell with Update and Blackhole, Lifetime Promotion on Escape, Structure / Representation Firewall, and Incremental Collection Barrier.

### GateMate realization

Use a centralized resolver sharing one heap read pipeline. Cache one resolved target per active context only if generations prevent staleness.

### Verification target

Generate chains, self sentinels, illegal cycles, moved objects, concurrent readers, and reclamation boundaries.

### Characteristic failure sign

An object is reclaimed before a reader following the old forwarding cell has completed.

### Miniature example

A thunk becomes an `IND` node pointing to a separately allocated value, so all existing references observe the memoized result.

**Lineage.** [KOG91], [JON92], [LIE83]

## Pattern 12: Semantic Region Partition

**Intent.** Store state with different lifetime, access, protection, and rollback behavior in different regions.

### Context and problem

One universal memory or stack forces incompatible port, allocation, collection, and recovery rules onto all state.

**Forces.** Partitioning permits specialization but fragments capacity and creates cross-region references and spill paths.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Named regions such as code, constants, construction heap, call environment, choice stack, trail, store queue, and message buffers. |
| Data and flow \(D,F\) | Each region has allocation, access, ownership, and reclamation interfaces. Cross-region references pass through explicit rules. |
| Ordering \(O\) | Region-specific order is preserved; transfers between regions define publication order. |
| Visibility \(V\) | Regions may correspond to different commitment levels. Movement across a boundary can advance visibility. |
| Recovery \(R\) | Checkpoints often save region tops. Spill, promotion, and collection recover pressure without corrupting other regions. |
| Bounds \(B\) | Each region has an independent finite capacity and a defined full policy; global admission accounts for compound use. |
| Timing and locality \(T\) | Bank and place by access function. Keep frequently interacting regions physically close or connect them elastically. |
| Invariant \(I\) | **Every live object resides in exactly one region whose lifetime and mutation policy is valid for that object.** |

### Operation

Classify state by lifetime, read/write ratio, rollback, and locality before assigning RAM blocks.

### Composition

Bump-Allocated Generation, Bank by Access Function, Choice-Point Snapshot, Mutation Trail, and Delayed Irreversible Store.

### GateMate realization

Map program ROM, stack, trail, choice points, heap, and trace to separately inferred memories so synthesis can choose appropriate modes.

### Verification target

Fill each region independently, test cross-region references, and verify that rollback and reset affect only intended regions.

### Characteristic failure sign

A trail entry and the binding it protects share a circular RAM whose unrelated wraparound overwrites the binding before unwind.

### Miniature example

The queens solver keeps eight current domains in registers, old domains in trail RAM, alternatives in a choice stack, and committed boards in a result FIFO.

**Lineage.** [WAR83], [KOG91], [LIE83]

## Pattern 13: Atomic Metadata Propagation

**Intent.** Move and update payload-defining metadata atomically with the payload across every transport and storage path.

### Context and problem

Tags, poison, epochs, capability validity, ownership, and fault state can be dropped or torn when a datapath was designed around payload bits alone.

**Forces.** Wider records consume routing and RAM, but separate shadow metadata demands synchronization and often more ports.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | A composite record or coupled payload/metadata memories with one ownership and commit protocol. |
| Data and flow \(D,F\) | Acceptance, forwarding, write enable, ECC, serialization, and replay apply to the complete record. |
| Ordering \(O\) | Metadata and payload share order. Any split implementation uses matching sequence or address identity. |
| Visibility \(V\) | A consumer never sees a new payload with old metadata or vice versa. |
| Recovery \(R\) | Rollback, reset, and fault propagation restore or invalidate the complete record. |
| Bounds \(B\) | All queues budget total record width. Metadata stores cannot overflow independently without stalling payload movement. |
| Timing and locality \(T\) | Wide fanout is minimized by carrying compact hot metadata and resolving cold descriptors locally. |
| Invariant \(I\) | **For every live item, payload and semantic metadata refer to the same operation, context, epoch, and authority state.** |

### Operation

Audit every path: bypass, RAM, FIFO, DMA, debug, reset, context switch, assist, and error response.

### Composition

Tagged Value Word, Context / Epoch-Carrying Token, Capability / Descriptor Gate, and Speculation Shadow Structures.

### GateMate realization

Pack hot fields into `value40` or `event80`. When using parallel RAMs, drive address and write-enable from one registered command.

### Verification target

Attach sequence numbers in simulation, perturb stalls independently, and assert sequence equality at every payload/metadata reunion.

### Characteristic failure sign

A data FIFO stalls while a parallel tag FIFO advances, pairing the next item’s tag with the current payload.

### Miniature example

A multiplier pipeline carries context, epoch, destination, overflow, and value through identical valid stages.

**Lineage.** [KOG91], [WAT19], [KHA19]

## Chapter review

The representation and metadata patterns should not be selected by name alone. For each candidate, write its nine-field contract, identify its commitment level, and measure whether its added state and routing remove the dominant semantic cost.

## Exercises

1. Choose two patterns from this chapter and test their composition across all eight closure dimensions from Chapter 5.
2. Rewrite one pattern for an implementation with no block RAM, then for an implementation with abundant dual-port RAM.
3. Identify a pattern whose invariant is satisfied locally but can still fail globally because of missing metadata or rollback ownership.
4. Create a negative test that produces the characteristic failure sign of one pattern without violating unrelated interfaces.

# Control, frames, lifetime, and nondeterminism

## Learning objectives

After this chapter, you should be able to recognize the family’s recurring problem, fill out its pattern contracts, and select compatible patterns for a measured bottleneck.

Patterns 14-22 make continuations, escape, alternatives, rollback, and demand explicit.

## Pattern 14: Explicit Continuation Record

**Intent.** Represent what happens next as data that can be saved, moved, inspected, and resumed.

### Context and problem

Return addresses alone are insufficient when resumption also requires environment, destination, exception, update, or result-routing state.

**Forces.** Richer records make suspension precise but consume memory and can expose stale references if lifetimes are not coordinated.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Continuation kind, resume address or state, environment reference, destinations, saved operands, exception target, and optional context/epoch. |
| Data and flow \(D,F\) | Operations push, replace, inspect, and pop records; resumption reconstructs all semantic control state from the record. |
| Ordering \(O\) | Continuation order is normally LIFO, but first-class or task continuations may be queued or copied under explicit semantics. |
| Visibility \(V\) | A continuation is semantically private until invoked or exported; invocation effects cross normal commit boundaries. |
| Recovery \(R\) | Checkpointing saves the continuation pointer and any younger records; stale context or epoch records are rejected. |
| Bounds \(B\) | Stack or pool capacity is finite. Push reserves a slot before changing current control. |
| Timing and locality \(T\) | Record width, RAM read latency, and push/pop throughput shape the controller. Cache the top record when beneficial. |
| Invariant \(I\) | **Resuming a live continuation requires no unrecorded transient state and recreates the intended abstract control point exactly.** |

### Operation

Enumerate continuation kinds; define fields per kind; make every suspend point construct a complete record before relinquishing control.

### Composition

Split-Lifetime Frame, Tail-State Reuse, Demand Cell with Update and Blackhole, Fork / Join Collector, and Choice-Point Snapshot.

### GateMate realization

Store 40-bit compact frames in BRAM and keep one decoded top frame in registers. Use a tagged union rather than a maximal record if widths differ greatly.

### Verification target

Suspend and resume at every controller state, inject stalls between record write and control transfer, and verify reference-model equivalence.

### Characteristic failure sign

A reducer pushes only the return address, then loses the left operand required after the right child is evaluated.

### Miniature example

An `UPDATE` continuation records the thunk address so the computed value can replace the correct blackhole.

**Lineage.** [LAN64], [KOG91], [JON92]

## Pattern 15: Split-Lifetime Frame

**Intent.** Keep short-lived operands in fast local state and store only values that survive calls, suspension, migration, or rollback in longer-lived frames.

### Context and problem

A full activation record for every temporary causes excessive memory traffic and enlarges checkpoints and collector root sets.

**Forces.** Aggressive splitting improves locality but complicates liveness analysis, debugging, exceptions, and deoptimization.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Fast operand registers or stack cache, long-lived environment slots, lifetime metadata, and transfer operations between them. |
| Data and flow \(D,F\) | Local operations consume fast state; call, suspend, or escape points materialize only values needed later. |
| Ordering \(O\) | Materialization occurs before control can leave the current fast context. Resume restores required operands in declared order. |
| Visibility \(V\) | Unmaterialized temporaries are not visible beyond the local transition; environment state is visible to continuations and collectors. |
| Recovery \(R\) | Faults before materialization retain fast state; faults after materialization can reconstruct from the environment and progress marker. |
| Bounds \(B\) | Fast slots and environment slots are finite. Spill or trap occurs before overwriting a live value. |
| Timing and locality \(T\) | The split is driven by access frequency and RAM latency. Avoid broad muxes that erase the benefit of a small fast set. |
| Invariant \(I\) | **Every value live across a boundary is present in a recoverable long-lived location before that boundary is crossed.** |

### Operation

Compute boundary liveness; define materialize and restore transitions; keep a conservative all-in-frame mode for validation.

### Composition

Explicit Continuation Record, Liveness-Driven Frame Trimming, Tail-State Reuse, and Checked Compiler Hint.

### GateMate realization

Use two or four top-of-stack registers backed by synchronous RAM. Maintain a precise logical depth independent of cache occupancy.

### Verification target

Randomly force suspension at legal boundaries and compare resumed execution; disable splitting and require identical commit traces.

### Characteristic failure sign

A value needed by an exception handler remains only in a register overwritten by the fault sequence.

### Miniature example

The tagged evaluator caches two operands while deeper values remain in BRAM; only live stack depth defines architectural state.

**Lineage.** [LAN64], [WAR83], [PAT81]

## Pattern 16: Liveness-Driven Frame Trimming

**Intent.** Reclaim frame slots and root references immediately after their last semantic use.

### Context and problem

Dead slots consume frame space, extend object lifetimes in garbage-collected systems, and inflate checkpoint bandwidth.

**Forces.** Precise liveness saves space but requires compiler metadata and must include exceptions, backtracking, debug, and deoptimization paths.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Live-slot bitmap or trim count, frame pointer, optional reconstruction metadata, and compiler hint validation. |
| Data and flow \(D,F\) | At a trim point, slots proven dead become inaccessible and cease to count as roots; physical memory may be reused later. |
| Ordering \(O\) | Trimming occurs after all ordinary and exceptional reads of a slot. Backtracking cannot restore a value that was semantically live at the checkpoint. |
| Visibility \(V\) | Dead values are no longer observable through the abstract machine, though debug modes may retain separate history. |
| Recovery \(R\) | A failed liveness check falls back to an untrimmed frame. Checkpoint restore uses the live shape associated with its control point. |
| Bounds \(B\) | Bitmap and metadata tables are finite. Unknown control points conservatively retain all slots. |
| Timing and locality \(T\) | Use compact masks and update them at coarse control points rather than on every cycle. |
| Invariant \(I\) | **No continuation, exception, rollback, collector, or debug operation can legally read a slot after it is trimmed.** |

### Operation

Derive liveness from the transition graph; classify hints; validate against continuation descriptors in debug builds.

### Composition

Checked Compiler Hint, Split-Lifetime Frame, Semantic Region Partition, and Tail-State Reuse.

### GateMate realization

For small frames, hold a live bitmap in registers. For large bytecode programs, store frame-shape descriptors in ROM.

### Verification target

Instrument every slot read and assert its live bit; compare root sets and execution with trimming disabled.

### Characteristic failure sign

A slot is dead on the normal path but needed after a type fault; trimming makes the slow path unrestartable.

### Miniature example

After the first relation goal binds `Y`, temporary scanner fields not needed by the second goal are removed from the active context frame.

**Lineage.** [WAR83], [KOG91]

## Pattern 17: Tail-State Reuse

**Intent.** Execute a tail transfer without accumulating an additional continuation or frame.

### Context and problem

A call whose caller has no remaining work needlessly consumes stack and checkpoint state if implemented as call plus return.

**Forces.** Reuse improves space and latency but is illegal when debuggers, exceptions, dynamic-wind behavior, or deferred actions still require the caller.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Current frame, continuation pointer, callee arguments, and optional debug or exception metadata. |
| Data and flow \(D,F\) | The transition validates tail position, replaces the current activation with the callee state, and jumps without pushing a return record. |
| Ordering \(O\) | Argument movement must handle overlap and preserve required evaluation order. |
| Visibility \(V\) | The discarded caller is no longer semantically observable except through explicitly retained diagnostic history. |
| Recovery \(R\) | If validation fails or metadata cannot be preserved, use the ordinary call path. Faults during argument setup leave the caller recoverable. |
| Bounds \(B\) | No additional continuation capacity is required, but callee frame requirements still need reservation. |
| Timing and locality \(T\) | Parallel move logic or temporary registers may be needed. Keep the tail path regular rather than building a large permutation network. |
| Invariant \(I\) | **At the reuse point, no future legal transition can return to or inspect the discarded caller state.** |

### Operation

Mark tail transitions in the abstract machine; verify exception and debug policy; implement a safe ordinary-call fallback.

### Composition

Explicit Continuation Record, Split-Lifetime Frame, Liveness-Driven Frame Trimming, and Checked Compiler Hint.

### GateMate realization

A compact evaluator can overwrite the current environment descriptor and `pc` after moving arguments through a small temporary buffer.

### Verification target

Compare bounded-recursion stack depth with and without reuse; inject faults during argument transfer.

### Characteristic failure sign

A pending update continuation is discarded as though it were an ordinary caller, leaving a thunk permanently blackholed.

### Miniature example

A tail-recursive list walker reuses one environment record for every element and maintains constant stack depth.

**Lineage.** [LAN64], [KOG91]

## Pattern 18: Lifetime Promotion on Escape

**Intent.** Keep objects in cheap local storage until a reference is about to outlive that storage, then promote them safely.

### Context and problem

Allocating every temporary in a long-lived heap wastes bandwidth, but returning a pointer into a reclaimable local frame creates dangling references.

**Forces.** Promotion adds copying and reference rewriting at escape points; conservative promotion loses some locality benefit.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Local object, destination-region allocation, forwarding or rewritten reference, and escape metadata. |
| Data and flow \(D,F\) | Before an escaping reference is published, allocate destination space, copy the object graph required by the policy, and update or forward the source reference. |
| Ordering \(O\) | Destination initialization precedes publication. Recursive graph promotion follows a defined traversal order and preserves sharing. |
| Visibility \(V\) | External consumers see only the promoted stable identity or a forwarding reference that resolves to it. |
| Recovery \(R\) | Allocation failure leaves the local object and source reference unchanged. Partial copies are discarded via destination-region checkpoint. |
| Bounds \(B\) | Promotion worklists and destination space are finite; failure stalls or traps before publication. |
| Timing and locality \(T\) | Common non-escaping paths remain local. Escape checks should be placed at known boundary operations. |
| Invariant \(I\) | **No live reference survives to storage that can be reclaimed, and promotion preserves object structure and sharing according to the language policy.** |

### Operation

Classify escape boundaries; reserve destination space; copy under one owner; publish last; optionally install a forwarding cell.

### Composition

Checked Compiler Hint, Indirection / Forwarding Cell, Bump-Allocated Generation, and Structure / Representation Firewall.

### GateMate realization

Use a local register or small RAM region and a heap-top checkpoint. Count promoted bytes to validate whether the split pays off.

### Verification target

Create closures and messages with nested references, force destination exhaustion, and reclaim local storage immediately after publication in simulation.

### Characteristic failure sign

A closure captures a stack variable by raw address; later invocation reads a reused stack slot.

### Miniature example

A locally constructed pair is copied to the heap only when inserted into a long-lived graph node.

**Lineage.** [KOG91], [LIE83]

## Pattern 19: Choice-Point Snapshot

**Intent.** Capture only the state required to resume a remaining alternative.

### Context and problem

Copying the complete machine state at every nondeterministic branch is expensive, while saving too little makes exact resumption impossible.

**Forces.** Small snapshots rely on complementary undo and region-top mechanisms. Larger snapshots simplify recovery but increase bandwidth and capacity pressure.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Alternative target or remaining options, selected live arguments, continuation/environment state, previous choice pointer, allocation marks, and trail mark. |
| Data and flow \(D,F\) | A branch reserves and writes the snapshot before mutating branch-specific state. Failure restores the snapshot and selects the next alternative. |
| Ordering \(O\) | Choice points form an age order, usually LIFO for depth-first search. Other search orders require a different container and state-sharing policy. |
| Visibility \(V\) | The snapshot is semantically provisional metadata. It becomes dead at cut, trust, or search exhaustion. |
| Recovery \(R\) | Restoration plus unwinding younger effects must recreate the exact state at branch entry. |
| Bounds \(B\) | Choice capacity is finite. Branch creation stalls or faults before choosing when no entry is available. |
| Timing and locality \(T\) | Compact records fit RAM; caching the top entry reduces recovery latency. Multiword records need atomic allocation and valid-last publication. |
| Invariant \(I\) | **Restoring a choice point and all associated younger recovery state yields the same abstract state that existed immediately before the corresponding choice.** |

### Operation

Identify non-recomputable state; combine a snapshot with trail and region marks; write valid only after the full entry is stored.

### Composition

Mutation Trail, Commit / Cut Fence, Semantic Region Partition, and Explicit Continuation Record.

### GateMate realization

A 20- or 40-bit record can hold a small solver’s variable, remaining options, trail mark, and controller state. Larger contexts use multiple words plus a valid bit.

### Verification target

Hash full abstract state at branch creation and after restoration; exhaust capacity and interrupt multiword pushes.

### Characteristic failure sign

The snapshot saves `pc` and arguments but omits read/write mode, so unification resumes with the wrong transition semantics.

### Miniature example

The queens solver stores selected column, remaining row bits, trail mark, propagation bitmap, and prior choice pointer.

**Lineage.** [WAR83], [KOG91]

## Pattern 20: Mutation Trail

**Intent.** Undo selected mutations by logging prior values rather than copying all mutable state.

### Context and problem

Whole-state snapshots scale with state size even when a branch changes only a few old locations.

**Forces.** Trails are efficient for sparse changes and stack-like rollback. They add write traffic and are less suitable for arbitrary version access.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Trail memory, top pointer, entry format containing location and old value or inverse operation, and active checkpoint marks. |
| Data and flow \(D,F\) | Before modifying a location that predates the active rollback boundary, append enough information to restore it. Unwind entries in reverse order. |
| Ordering \(O\) | Log-before-write order is mandatory. Reverse unwind preserves dependencies between repeated writes. |
| Visibility \(V\) | Trailed writes may be architecturally present but semantically reversible until their branch commits. |
| Recovery \(R\) | Rollback decrements the top, reads the entry, restores the location, and repeats to the checkpoint mark. Commit truncates unreachable history. |
| Bounds \(B\) | Capacity is finite. The writer must reserve required slots before changing protected state. Newly allocated younger state may avoid logging if region rollback discards it. |
| Timing and locality \(T\) | One push per cycle is simple; multi-write operations require reservation, serialization, or a wider trail port. |
| Invariant \(I\) | **Unwinding exactly the entries younger than a checkpoint restores every protected location to its checkpoint value.** |

### Operation

Define which locations require trailing; centralize mutation; log old value before write; suppress redundant logging only under a proved rule.

### Composition

Choice-Point Snapshot, Commit / Cut Fence, History / Undo Buffer, and Bump-Allocated Generation.

### GateMate realization

Use BRAM for entries and keep `trail_top` in a register. A dedicated unwind FSM arbitrates restoration writes to the protected store.

### Verification target

Compare state hashes before choice and after unwind; generate repeated writes to one location; fill the trail to the last entry.

### Characteristic failure sign

A binding is written in the same cycle that a full trail is discovered, leaving no record for rollback.

### Miniature example

Each queens domain narrowing logs `{column, old_mask}`; backtracking restores masks until the choice point’s trail mark.

**Lineage.** [WAR83], [KOG91]

## Pattern 21: Commit / Cut Fence

**Intent.** Declare that older alternatives and their recovery state are no longer reachable.

### Context and problem

Checkpoint and trail structures grow until the machine knows that no legal future operation can request old states.

**Forces.** Early commitment frees resources and simplifies output, but a wrong fence changes semantics irreversibly.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Semantic commit frontier, choice pointer, trail reclaim point, region marks, and optional pending-output state. |
| Data and flow \(D,F\) | After all prior required results are safely staged, the fence discards alternatives, advances reclaim marks, and changes future failure behavior. |
| Ordering \(O\) | The fence is ordered after effects whose success justifies commitment and before operations that reuse reclaimed entries. |
| Visibility \(V\) | It advances effects from architecturally present to semantically irreversible. External visibility may still be delayed. |
| Recovery \(R\) | A fence itself is not rolled back by older alternatives. Faults before the fence retain recovery state; faults after it cannot request discarded choices. |
| Bounds \(B\) | Commit can release trail, choice, and heap capacity. Reclamation may be incremental to avoid a long combinational operation. |
| Timing and locality \(T\) | Frontier updates should be small and atomic. Physical clearing of old RAM entries is unnecessary; pointer movement is sufficient. |
| Invariant \(I\) | **After the fence commits, no legal continuation, failure, or cancellation can name a recovery state older than the new frontier.** |

### Operation

Name the semantic event that justifies commitment; ensure required outputs are buffered; update pointers atomically; invalidate stale epochs if needed.

### Composition

Choice-Point Snapshot, Mutation Trail, Delayed Irreversible Store, and Context / Epoch-Carrying Token.

### GateMate realization

Implement cut as pointer updates and optional epoch increment. Do not reset large stacks or trails to reclaim them.

### Verification target

Run the same search with and without cut; require the cut result to be a prefix or policy-defined subset and prove no older restore occurs.

### Characteristic failure sign

The machine discards alternatives before the first solution enters the output FIFO; downstream backpressure then loses the only result.

### Miniature example

In `FIRST_ONLY` queens mode, accept the first board into the result FIFO, then discard all choice points and trail entries.

**Lineage.** [WAR83], [KOG91]

## Pattern 22: Demand Cell with Update and Blackhole

**Intent.** Evaluate deferred work at most once, detect recursive demand, and memoize the result for all later consumers.

### Context and problem

Naive lazy evaluation can duplicate expensive computation; concurrent or recursive forcing can race, deadlock, or create inconsistent updates.

**Forces.** The cell must represent unevaluated, claimed, completed, and exceptional states. Ownership and waiter management add mutation to an otherwise functional graph.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Thunk body and environment, atomic claim or blackhole state, owner or version, result or indirection, waiter list, and exception state. |
| Data and flow \(D,F\) | The first forcer atomically claims the cell and evaluates its body. Same-owner re-entry detects a cycle; other consumers wait. Completion replaces the cell and wakes waiters. |
| Ordering \(O\) | Claim precedes body evaluation; completed value publication precedes wakeup. Competing claims serialize. |
| Visibility \(V\) | Consumers observe one authoritative result or exception. Intermediate blackhole ownership is internal unless diagnostics expose it. |
| Recovery \(R\) | Failure installs a defined exceptional result or restores the thunk under explicit retry semantics. Cancelled owners cannot leave permanent blackholes. |
| Bounds \(B\) | Waiter storage, evaluator contexts, and update queues are finite. Full waiter capacity stalls a new force before it loses its continuation. |
| Timing and locality \(T\) | Heap arbitration and RAM read-modify-write latency dominate. A local claim cache must be generation-safe. |
| Invariant \(I\) | **At most one successful evaluation becomes authoritative, every completed force observes that result, and recursive self-demand is detected.** |

### Operation

Define state tags; serialize claim/update; push a complete update continuation; publish value before wakeup; specify owner cancellation.

### Composition

Indirection / Forwarding Cell, Explicit Continuation Record, Producer Tag as Future, Tagged Result Wakeup, and Bump-Allocated Generation.

### GateMate realization

Begin with one evaluator, where any blackhole is a cycle. Add an owner field and waiter FIFO only after single-context correctness.

### Verification target

Count body executions, force shared thunks repeatedly, create direct and indirect cycles, cancel an owner, and saturate waiters.

### Characteristic failure sign

A thunk is overwritten with blackhole before an `UPDATE` continuation is safely pushed; stack overflow leaves the graph permanently poisoned.

### Miniature example

Four references to `delay(21*2)` all return 42 while the multiplier execution counter increments once.

**Lineage.** [KOG91], [JON92]

## Chapter review

The control, frames, lifetime, and nondeterminism should not be selected by name alone. For each candidate, write its nine-field contract, identify its commitment level, and measure whether its added state and routing remove the dominant semantic cost.

## Exercises

1. Choose two patterns from this chapter and test their composition across all eight closure dimensions from Chapter 5.
2. Rewrite one pattern for an implementation with no block RAM, then for an implementation with abundant dual-port RAM.
3. Identify a pattern whose invariant is satisfied locally but can still fail globally because of missing metadata or rollback ownership.
4. Create a negative test that produces the characteristic failure sign of one pattern without violating unrelated interfaces.

# Scheduling and communication

## Learning objectives

After this chapter, you should be able to recognize the family’s recurring problem, fill out its pattern contracts, and select compatible patterns for a measured bottleneck.

Patterns 23-32 determine when work becomes ready and how bounded work moves.

## Pattern 23: Ready-Operand Firing

**Intent.** Make an operation eligible when all operands required by its selected behavior are present.

### Context and problem

A single global program counter or fixed issue order stalls independent work behind missing inputs and obscures natural graph parallelism.

**Forces.** Readiness scheduling tolerates variable latency but needs presence state, duplicate protection, and bounded work queues.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Operand slots, per-slot valid state, operation descriptor, issued or fired bit, context and epoch, and ready queue. |
| Data and flow \(D,F\) | Operand arrival fills one empty slot. The transition that makes the last required operand present enqueues exactly one activation. |
| Ordering \(O\) | Operands for one activation are single-assignment within an epoch. Independent activations may fire in any permitted order. |
| Visibility \(V\) | Firing is internal; the operation’s result follows its own commit and routing contract. |
| Recovery \(R\) | Epoch cancellation invalidates slots and queued activations. Replay clears issued state only under a defined new generation. |
| Bounds \(B\) | Operand stores and ready queues are finite. A last-operand arrival is accepted only when space exists to represent the newly ready activation. |
| Timing and locality \(T\) | Last-arrival detection and queue insertion can be a critical path. Use registered presence bits, banking, or a two-step ready-mark protocol. |
| Invariant \(I\) | **Each live activation fires at most once after all required operands are present and never fires with a missing or stale operand.** |

### Operation

Define required operand masks per opcode; reject duplicates; atomically mark `issued` when enqueuing; clear state on completion or new epoch.

### Composition

Producer Tag as Future, Reservation Station, Tagged Result Wakeup, Context / Epoch-Carrying Token, and Fork / Join Collector.

### GateMate realization

Keep valid and issued bits in registers for small engines while values reside in 40-bit RAM banks. Explicit destination lists avoid associative broadcast.

### Verification target

Permute operand arrival order, duplicate tokens, cancel at every stage, fill the ready FIFO, and assert one fire per activation.

### Characteristic failure sign

Two operands arrive on adjacent cycles and each observes the other as present before `issued` is set, causing duplicate enqueue.

### Miniature example

A dataflow `ADD` node enters the ready FIFO exactly when its second input token arrives.

**Lineage.** [DEN75], [KOG91]

## Pattern 24: Producer Tag as Future

**Intent.** Represent an unavailable value by the identity of the computation that will eventually produce it.

### Context and problem

Consumers otherwise need to stall at the producer, poll a location, or reserve a physical value slot before production.

**Forces.** Tags decouple production from consumption but require unique lifetimes, matching, and generation-safe reuse.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Producer identifier, generation or epoch, ready/value alternative, and consumer subscriptions or destination list. |
| Data and flow \(D,F\) | A consumer operand contains either a value or a producer tag. Completion resolves matching subscriptions or routes to explicit destinations. |
| Ordering \(O\) | The tag denotes one production event. Result ordering may vary, but every consumer associates the correct event with its operand. |
| Visibility \(V\) | The tag itself is provisional dependency state. The resolved value advances according to the consumer operation’s commitment. |
| Recovery \(R\) | Cancelled or replayed producers invalidate their generation; stale completions are discarded and never satisfy new consumers. |
| Bounds \(B\) | Tag namespace, subscriptions, and in-flight generations are finite. Reuse waits until all possible stale references are excluded. |
| Timing and locality \(T\) | Associative matching scales poorly. Cluster tags, use indexed destination tables, or encode direct consumers. |
| Invariant \(I\) | **Within its live generation, each producer tag names at most one authoritative value or one defined fault.** |

### Operation

Allocate tag with generation; install consumer references; publish completion; retire tag only after consumers and delayed paths cannot reference it.

### Composition

Ready-Operand Firing, Reservation Station, Tagged Result Wakeup, Context / Epoch-Carrying Token, and Demand Cell with Update and Blackhole.

### GateMate realization

For small FPGA graphs, use node/context/epoch as a producer identity and route to explicit destinations stored in descriptor ROM.

### Verification target

Force tag wraparound in a reduced namespace, delay old results, cancel producers, and assert stale results never wake a new consumer.

### Characteristic failure sign

A tag is reused immediately after issue-queue removal while a multiplier result carrying the old tag is still in flight.

### Miniature example

An unresolved operand slot stores `{context, epoch, producer_node}` until that node’s completion event arrives.

**Lineage.** [TOM67], [DEN75]

## Pattern 25: Reservation Station

**Intent.** Hold waiting operations near scheduling logic so functional units accept only work that can make progress.

### Context and problem

If an operation occupies a functional unit while waiting for operands, it blocks independent ready work and couples unit latency to issue order.

**Forces.** Distributed stations improve utilization but require matching, selection, age or fairness policy, and finite-entry management.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Operation, destination, ready values, unresolved producer tags, context/epoch, age, fault metadata, and valid state. |
| Data and flow \(D,F\) | Dispatch reserves an entry. Wakeup fills operands. Eligible selection sends one entry to a ready unit and releases or marks the entry according to replay policy. |
| Ordering \(O\) | Selection may prioritize age, context fairness, or unit locality. Architectural order is restored by a later commit mechanism if required. |
| Visibility \(V\) | Station state is microarchitecturally provisional and not externally visible. |
| Recovery \(R\) | Squash invalidates entries by epoch or branch mask. Replay preserves the operation and clears completion state without duplicating effects. |
| Bounds \(B\) | Entry count and operand storage are finite. Dispatch backpressures before accepting an operation without a station. |
| Timing and locality \(T\) | Wakeup compare and selection can dominate timing and power. Small distributed stations or indexed wakeup fit FPGA fabric better than one wide associative window. |
| Invariant \(I\) | **Every valid entry is either eventually issued, explicitly cancelled, or faulted, and no entry issues before required operands and unit capacity are available.** |

### Operation

Choose allocation, wakeup, selection, and release events; separate operand presence from functional-unit readiness; make fairness policy explicit.

### Composition

Producer Tag as Future, Tagged Result Wakeup, Precise Commit Queue, Replayable Operation, and Context / Epoch-Carrying Token.

### GateMate realization

Start with an eight-entry register array and one comparator per operand. Measure before increasing width; partition by functional-unit class when needed.

### Verification target

Randomize unit latency and stalls, check no double issue, ensure every accepted entry drains under fairness assumptions, and squash delayed entries.

### Characteristic failure sign

An entry is freed at issue even though the operation can replay, so the original operands and destination are lost after a memory conflict.

### Miniature example

A multiply reservation station waits on one producer tag while independent additions issue through a separate station.

**Lineage.** [TOM67]

## Pattern 26: Tagged Result Wakeup

**Intent.** Deliver a completion only to consumers that subscribed to the producer identity or explicit destination.

### Context and problem

Polling wastes bandwidth, while unqualified broadcast can overwrite unrelated operands or wake stale work.

**Forces.** Broadcast is simple at small scale; explicit routes and clustering reduce fanout but add descriptor storage and multicast handling.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Completion record containing producer tag or destination list, value and metadata, context, epoch, and optional sequence. |
| Data and flow \(D,F\) | A result is accepted by matching live operands or routed destinations. Each destination records the value once and may trigger ready firing. |
| Ordering \(O\) | All consumers of one result observe the same value. Delivery order across independent producers is unspecified unless a higher contract requires it. |
| Visibility \(V\) | Wakeup is internal dependency resolution, not architectural commit. |
| Recovery \(R\) | Epoch and generation checks discard stale completions. Duplicate delivery is detected or made idempotent per operand slot. |
| Bounds \(B\) | Completion queues, multicast lists, and destination ports are finite. A producer stalls before losing a result. |
| Timing and locality \(T\) | Global compare fanout limits frequency. Use hierarchical multicast, banked operand stores, or copy nodes to bound scope. |
| Invariant \(I\) | **A completion satisfies exactly the live consumers of its named production event and no stale or unrelated consumer.** |

### Operation

Choose broadcast scope; carry complete metadata; validate generation before writing slots; make duplicate behavior explicit.

### Composition

Producer Tag as Future, Ready-Operand Firing, Reservation Station, Elastic Channel, and Atomic Metadata Propagation.

### GateMate realization

The dataflow experiment stores up to two destinations in node ROM. Higher fanout is represented by explicit `COPY` nodes.

### Verification target

Delay completions through random queues, recycle tags aggressively, duplicate events, and check operand-slot sequence numbers.

### Characteristic failure sign

The value is broadcast with producer tag but without context epoch; a cancelled computation wakes a reused node slot.

### Miniature example

A multiplier completion routes to the A port of one add node and the B port of another, then becomes backpressured until both writes are accepted.

**Lineage.** [TOM67], [DEN75]

## Pattern 27: Access / Execute Decoupling

**Intent.** Allow address generation and memory movement to progress independently from arithmetic or symbolic execution.

### Context and problem

A single sequential controller alternates between memory latency and computation, leaving one side idle and exposing every miss directly.

**Forces.** Decoupling hides latency and enables specialization, but queues can run ahead incorrectly when aliasing, control, or ordering is unresolved.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Access stream state, execute stream state, data or address queues, dependence metadata, and synchronization tokens. |
| Data and flow \(D,F\) | The access side produces values or memory tokens; the execute side consumes them. Control tokens coordinate loops, branches, and exceptions. |
| Ordering \(O\) | Memory and semantic order are explicit per stream. Alias checks or synchronization prevent unsafe reordering. |
| Visibility \(V\) | Access results are provisional until paired with the correct execute operation and commitment context. |
| Recovery \(R\) | Replay or cancellation invalidates queued tokens by sequence or epoch. Faults travel with the corresponding access token. |
| Bounds \(B\) | Queues are finite; access runahead is bounded by credits and by the maximum speculative state that can be cancelled. |
| Timing and locality \(T\) | Long routes benefit from elastic channels. Queue depth should cover measured latency variance rather than arbitrary large margins. |
| Invariant \(I\) | **Every execute consumption corresponds to the correct access production under the declared memory-order and context rules.** |

### Operation

Separate state machines; define token identities and synchronization; add alias or dependence checks before permitting runahead.

### Composition

Elastic Channel, Credit-Based Backpressure, Context / Epoch-Carrying Token, Nonblocking Miss Record, and Replayable Operation.

### GateMate realization

Use a fact scanner or heap walker as the access engine and an unifier or reducer as execute. Connect with BRAM FIFOs if route latency is significant.

### Verification target

Vary memory latency, create alias conflicts, cancel runahead work, fill queues, and compare with a coupled reference execution.

### Characteristic failure sign

The access engine reads beyond a cut and sends facts into a reused query context without an epoch, producing extra solutions.

### Miniature example

A relation scanner streams candidate fact records while the unifier processes the previous candidate.

**Lineage.** [SMI82]

## Pattern 28: Elastic Channel

**Intent.** Make module latency, stalls, and physical retiming independent from functional behavior.

### Context and problem

Pulse interfaces and fixed-latency assumptions lose data when a consumer stalls or when an implementation gains a pipeline stage.

**Forces.** Ready/valid adds state and reverse control. A combinational ready chain may become a timing loop or long critical path.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Payload register or FIFO storage, occupancy, complete metadata record, and optional skid capacity. |
| Data and flow \(D,F\) | Transfer occurs only on `valid && ready`. While blocked, the producer holds `valid` and data stable. Relay stages may be inserted without changing sequence. |
| Ordering \(O\) | FIFO order for the channel. Parallel channels require an explicit rule for alignment or independent ordering. |
| Visibility \(V\) | Transfer changes ownership but does not necessarily cross an architectural or semantic commit boundary. |
| Recovery \(R\) | Reset and cancellation semantics specify whether buffered items are discarded, drained, or marked invalid by epoch. |
| Bounds \(B\) | Capacity is exact and backpressure is lossless. The producer does not assume eventual readiness unless the environment contract states fairness. |
| Timing and locality \(T\) | Register long forward and reverse paths; use credits when combinational backpressure cannot span the physical distance. |
| Invariant \(I\) | **Every accepted item is delivered exactly once and in order, unless a specified reset or cancellation rule invalidates it.** |

### Operation

Define acceptance and hold rules; include all metadata; insert one-entry stages at boundaries; audit ready loops.

### Composition

Credit-Based Backpressure, Context / Epoch-Carrying Token, Tagged Result Wakeup, and Delayed Irreversible Store.

### GateMate realization

Use register stages for small records and BRAM FIFOs for depth. Avoid resetting the RAM contents; reset pointers and valid state.

### Verification target

Assert stability under stall, conservation of item sequence numbers, occupancy bounds, and eventual drain under fair readiness.

### Characteristic failure sign

A producer changes `data` every cycle while keeping `valid` high and the consumer deasserts `ready`.

### Miniature example

The multiplier’s internal latency can change from two to five stages without changing the dataflow scheduler.

**Lineage.** [CAR01]

## Pattern 29: Credit-Based Backpressure

**Intent.** Control a long-distance or deeply buffered flow using explicit capacity tokens rather than a combinational reverse path.

### Context and problem

Ready propagation across many stages hurts timing, and a sender otherwise cannot know whether downstream storage exists.

**Forces.** Credits improve physical timing but introduce accounting state, return latency, reset coordination, and the possibility of leaks.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Credit counter at sender, downstream slots, credit-return events, generation or reset epoch, and optional reserved credits for compound operations. |
| Data and flow \(D,F\) | Sending consumes one or more credits. Releasing downstream capacity returns the same number. The sender stalls at zero. |
| Ordering \(O\) | Credit events may return out of order if they are fungible; data ordering follows the forward channel contract. |
| Visibility \(V\) | Credit ownership is flow-control state, not semantic commitment. |
| Recovery \(R\) | Reset establishes one agreed initial count. Cancellation must still release every occupied slot exactly once. |
| Bounds \(B\) | Initial credits equal usable capacity. Counter width covers the full range without wraparound. |
| Timing and locality \(T\) | The reverse path can be pipelined. Return latency determines how much capacity is needed for full throughput. |
| Invariant \(I\) | **Credits plus occupied downstream slots remain equal to the configured usable capacity, modulo explicitly reserved entries.** |

### Operation

Define the slot-release event; centralize accounting; carry reset epoch if endpoints can reset independently; assert conservation.

### Composition

Elastic Channel, Access / Execute Decoupling, Nonblocking Miss Record, and Fork / Join Collector.

### GateMate realization

For on-chip GateMate experiments, use credits when crossing several pipeline stages or connecting a producer to a BRAM FIFO whose `ready` path is registered.

### Verification target

Drop, duplicate, and delay credit returns in negative tests; assert no send at zero and no counter above capacity.

### Characteristic failure sign

A cancelled context invalidates data entries but does not return their credits, eventually deadlocking the machine.

### Miniature example

The scanner may have eight candidate facts in flight because the unifier response path owns eight credits.

**Lineage.** [CAR01]

## Pattern 30: Context / Epoch-Carrying Token

**Intent.** Keep overlapping threads, graph activations, queries, search branches, and speculative generations separate.

### Context and problem

Physical queues and pipelines outlive the logical context state that launched them. Reused identifiers can accept stale responses.

**Forces.** Context and epoch bits widen every event, but allow cancellation by invalidation instead of searching all structures.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Context identifier, current epoch table, token epoch, optional branch or sequence tag, and stale-drop counters. |
| Data and flow \(D,F\) | Every request and response carries context and epoch. Before updating live state, a consumer compares the token with the current epoch. |
| Ordering \(O\) | Order is normally per context and operation. Independent contexts may interleave or complete out of order. |
| Visibility \(V\) | Old-epoch effects are never visible. A valid token still follows the normal commitment path. |
| Recovery \(R\) | Cancellation increments or replaces the epoch and clears local state; delayed old tokens are discarded when they emerge. |
| Bounds \(B\) | Epoch width is finite. Wraparound requires quiescence, a larger generation, or proof that no old token survives. |
| Timing and locality \(T\) | Comparisons occur at every state-update boundary. Centralized epoch tables should be banked or replicated read-only when fanout grows. |
| Invariant \(I\) | **No token whose `(context, epoch)` differs from the live context generation can modify live state or produce visible output.** |

### Operation

Attach identity at ingress; propagate atomically; check at every delayed exit; define wraparound and context reuse.

### Composition

Atomic Metadata Propagation, Checkpoint plus Epoch Squash, Producer Tag as Future, Access / Execute Decoupling, and Speculation Shadow Structures.

### GateMate realization

Eight context bits and eight epoch bits fit comfortably in the textbook `event80`. Start with small epoch widths in verification to force wrap tests.

### Verification target

Cancel at every pipeline stage, delay old multiplier and RAM responses, reuse contexts, and assert stale-drop behavior.

### Characteristic failure sign

The request carries epoch but the exception response does not, so an old fault terminates a new query.

### Miniature example

Cancelling context 3 increments its epoch; outstanding fact-scan and multiplier events for the prior epoch drain harmlessly.

**Lineage.** [TOM67], [KHA19]

## Pattern 31: Fork / Join Collector

**Intent.** Represent parallel subcomputations and an explicit policy for combining their completions.

### Context and problem

Ad hoc counters fail when children can cancel, fault, duplicate, or return out of order, and when not all children are required.

**Forces.** A general join supports rich predicates but consumes state. Fixed fanout and one completion policy are cheaper.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Parent identity, child labels, required or quorum mask, received mask, partial results, exception policy, context/epoch, and continuation. |
| Data and flow \(D,F\) | Fork allocates parent state and emits children. Join accepts one completion per live child label and fires when its predicate is satisfied. |
| Ordering \(O\) | Child arrival order is arbitrary. Result assembly follows label order or an explicitly commutative reduction. |
| Visibility \(V\) | The joined result becomes visible only after the parent operation’s normal commit. Early children remain provisional. |
| Recovery \(R\) | Cancellation invalidates outstanding children by epoch. Exception policy may fail-fast, collect-all, or choose a successful alternative. |
| Bounds \(B\) | Parent table and child slots are finite. Fork reserves complete join state and enough downstream capacity before emission. |
| Timing and locality \(T\) | Mask updates and reductions can be local. Large fanout should be hierarchical or streaming. |
| Invariant \(I\) | **Each required live child contributes at most once, and the parent fires exactly when the declared completion predicate first becomes true.** |

### Operation

Assign unique child labels; define duplicate, late, cancelled, and exceptional arrivals; publish parent readiness atomically.

### Composition

Ready-Operand Firing, Context / Epoch-Carrying Token, Credit-Based Backpressure, and Local Systolic Cell.

### GateMate realization

Use a small register-based join table for four or eight children. Wider joins can store masks in BRAM and process one completion per cycle.

### Verification target

Permute children, duplicate responses, cancel parents, inject mixed faults, and fill the parent table before accepting another fork.

### Characteristic failure sign

A simple decrementing counter reaches zero after the same child replies twice, firing without another required child.

### Miniature example

A parallel term comparison forks two field comparisons and joins only after both report equality.

**Lineage.** [DEN75], [KOG91]

## Pattern 32: Local Systolic Cell

**Intent.** Replace global movement with repeated small actors that communicate mainly with neighbors.

### Context and problem

Centralized memories and broadcasts dominate wire cost when a regular computation repeatedly reuses local data.

**Forces.** Systolic locality offers predictable placement and throughput but handles irregular control, sparse graphs, and variable-length objects poorly.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Per-cell operator, small local state, neighbor channels, stationary operand or partial result, and bypass or boundary state. |
| Data and flow \(D,F\) | Tokens enter an edge, transform or combine locally, and advance through elastic neighbor links. One chosen data class remains stationary. |
| Ordering \(O\) | Spatial order and wavefront timing are part of the contract. Elastic links can tolerate local stalls while preserving token sequence. |
| Visibility \(V\) | Intermediate cell state is internal; edge outputs pass through normal commitment and context checks. |
| Recovery \(R\) | Faults and cancellation propagate as tagged tokens or flush a bounded spatial epoch. Cells must not retain stale stationary state. |
| Bounds \(B\) | Each link and cell has finite buffering. Boundary injection uses credits to avoid filling a closed wavefront. |
| Timing and locality \(T\) | The pattern is explicitly physical: neighbor distance, regular replication, and stationary-data reuse are the reason to use it. |
| Invariant \(I\) | **Every output equals the specified composition of cell transitions over the input wavefront, and no token is duplicated or overtaken contrary to the spatial schedule.** |

### Operation

Choose stationary data, wavefront direction, boundary protocol, bypass behavior, and context tagging before laying out cells.

### Composition

Elastic Channel, Credit-Based Backpressure, Fork / Join Collector, Bank by Access Function, and Access / Execute Decoupling.

### GateMate realization

GateMate CPE regularity can support small arrays, but route and RAM placement must be measured. Keep control outside the cells where possible.

### Verification target

Compare against a sequential kernel across stalls, inject bubbles and cancellation, and inspect placement to confirm local routing.

### Characteristic failure sign

A global tag broadcast is added to every cell, eliminating the locality and timing benefit that justified the array.

### Miniature example

A row of cells performs fixed-shape term comparisons while a control engine handles exceptional tags and variable-length structures.

**Lineage.** [KUN82]

## Chapter review

The scheduling and communication should not be selected by name alone. For each candidate, write its nine-field contract, identify its commitment level, and measure whether its added state and routing remove the dominant semantic cost.

## Exercises

1. Choose two patterns from this chapter and test their composition across all eight closure dimensions from Chapter 5.
2. Rewrite one pattern for an implementation with no block RAM, then for an implementation with abundant dual-port RAM.
3. Identify a pattern whose invariant is satisfied locally but can still fail globally because of missing metadata or rollback ownership.
4. Create a negative test that produces the characteristic failure sign of one pattern without violating unrelated interfaces.

# Recovery, speculation, and precise visibility

## Learning objectives

After this chapter, you should be able to recognize the family’s recurring problem, fill out its pattern contracts, and select compatible patterns for a measured bottleneck.

Patterns 33-39 separate provisional execution from committed observation.

## Pattern 33: Precise Commit Queue

**Intent.** Permit overlapped or out-of-order execution while exposing effects in a required architectural order.

### Context and problem

Completion order can differ from program order, and a younger result must not become architectural if an older operation faults.

**Forces.** A commit queue gives precise state but consumes result storage and can suffer head-of-line blocking.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Ordered entries containing operation age, destination, result or completion state, fault, branch or epoch metadata, and side-effect references. |
| Data and flow \(D,F\) | Dispatch allocates in order. Execution completes entries in any legal order. The oldest complete entry retires atomically or reports its fault. |
| Ordering \(O\) | Retirement order is the required architectural order, normally a sequential prefix. Internal completion order is unrestricted. |
| Visibility \(V\) | Execution results remain at level \(\mu\) until retirement advances them to \(A\). |
| Recovery \(R\) | Branch or fault recovery invalidates younger entries and restores associated rename, queue, or checkpoint state. |
| Bounds \(B\) | Queue capacity is finite. Dispatch reserves an entry before accepting the operation; full prevents further allocation. |
| Timing and locality \(T\) | Head lookup is local, but result write ports and broadcast can be expensive. Store references to result buffers when full payload storage is costly. |
| Invariant \(I\) | **At every interrupt or fault boundary, visible architectural state equals a prefix of the sequential abstract transition stream.** |

### Operation

Separate allocation, completion, and retirement; attach faults to entries; delay stores and other irreversible effects until permitted.

### Composition

Reservation Station, Delayed Irreversible Store, Checkpoint plus Epoch Squash, and Future / Shadow File.

### GateMate realization

A small eight- or sixteen-entry queue can use registers; larger entries use RAM plus valid/completion bitmaps. Multiword entries publish valid last.

### Verification target

Complete entries in every permutation, inject an old fault, stall retirement, and compare architectural state at every retirement packet.

### Characteristic failure sign

A functional unit writes the architectural register file directly, so a younger value survives when an older instruction faults.

### Miniature example

A symbolic operation computes a reversible binding out of order, but it cannot enter the architectural binding state until its commit entry reaches the head.

**Lineage.** [SMP88]

## Pattern 34: History / Undo Buffer

**Intent.** Update the working state early while retaining old values sufficient for rollback.

### Context and problem

Keeping all new values in a reorder structure adds read indirection; early update improves access but overwrites the recovery image.

**Forces.** History logging favors fast reads but every write consumes log bandwidth, and repeated writes can create redundant entries.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Working state, ordered history entries containing destination and old value, checkpoint marks, and retirement or reclaim frontier. |
| Data and flow \(D,F\) | Before an early update, log the old value. Recovery restores entries in reverse order. Commitment discards history no longer needed. |
| Ordering \(O\) | Log-before-write and reverse restore are mandatory. Retirement order determines when history is reclaimable. |
| Visibility \(V\) | The working state is provisional even if physically updated; the committed observation uses the recovery frontier. |
| Recovery \(R\) | On squash, unwind to the checkpoint mark. Faults during unwind are treated as internal fatal conditions or use protected memory. |
| Bounds \(B\) | History capacity and write bandwidth are finite. Admission accounts for worst-case writes or splits operations into separately logged steps. |
| Timing and locality \(T\) | Working-state reads are fast. History RAM and unwind path can be slower if recovery latency is acceptable. |
| Invariant \(I\) | **Restoring all entries younger than a checkpoint reproduces the exact working state represented at that checkpoint.** |

### Operation

Centralize early writes, log old values, tag entries by checkpoint or age, and define when history can be reclaimed.

### Composition

Precise Commit Queue, Mutation Trail, Checkpoint plus Epoch Squash, and Replayable Operation.

### GateMate realization

Reuse the trail-stack substrate but keep commitment levels distinct. A microarchitectural history buffer and semantic trail must not both own one update.

### Verification target

Generate repeated writes, nested checkpoints, commit interleaving, and full-buffer conditions; compare restored hashes.

### Characteristic failure sign

The old value is read after the new value is written because RAM latency was overlooked, so the log records the replacement rather than history.

### Miniature example

A speculative rename-map update logs the old physical-register mapping for branch recovery.

**Lineage.** [SMP88], [WAR83]

## Pattern 35: Future / Shadow File

**Intent.** Maintain separate provisional and committed state images so recovery has a clean base.

### Context and problem

Undo logging can become bandwidth-intensive, while direct updates to the committed file destroy precise state.

**Forces.** Duplicated storage simplifies recovery but adds synchronization, ports, and copy or reconstruction work.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Committed file, speculative working file, validity or mapping state, retirement update path, and recovery copy or reset mechanism. |
| Data and flow \(D,F\) | Execution reads and writes the working image. Retirement advances selected results into the committed image. Recovery reconstructs or resets the working image from committed state. |
| Ordering \(O\) | Committed updates follow architectural order. Working updates follow dependency rules. |
| Visibility \(V\) | The committed image represents level \(A\); the working image remains \(\mu\)-provisional. |
| Recovery \(R\) | Recovery discards provisional validity or copies committed values. Long recovery may block dispatch but preserves precision. |
| Bounds \(B\) | Both images and mapping metadata are finite. Copy bandwidth determines worst-case recovery latency. |
| Timing and locality \(T\) | Dual storage can map to separate RAMs. Read bypass between new working values and consumers must be defined. |
| Invariant \(I\) | **The committed image is always a legal architectural state, and the working image can be reconstructed from it plus live committed-to-working mappings.** |

### Operation

Define which state is duplicated, how retirement synchronizes it, and how recovery handles values newer than the committed image.

### Composition

Precise Commit Queue, Speculation Shadow Structures, and Checkpoint plus Epoch Squash.

### GateMate realization

A small architectural register file and a separate speculative register file are straightforward FPGA experiments; avoid large synchronous clear networks.

### Verification target

Fault after arbitrary working writes, rebuild, and compare committed state; test retirement and recovery in adjacent cycles.

### Characteristic failure sign

A recovery copy starts while retirement continues modifying the committed file, producing a mixed image.

### Miniature example

A symbolic core keeps committed control registers separate from provisional scheduler state and copies only a compact map on recovery.

**Lineage.** [SMP88]

## Pattern 36: Delayed Irreversible Store

**Intent.** Hold effects that cannot be locally rolled back until all required commitment conditions are satisfied.

### Context and problem

Speculative stores, messages, host results, and device commands can escape before a later fault or semantic failure is known.

**Forces.** Buffering preserves precision but consumes capacity, adds latency, and may require ordering or merging.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Effect record, destination, payload and metadata, age, context/epoch, required commitment level, and queue status. |
| Data and flow \(D,F\) | The producer stages the effect. A release controller sends it only when all older faults are resolved and semantic commitment permits external visibility. |
| Ordering \(O\) | Release follows the required program, address, context, or transaction order. Private stores may merge under defined rules. |
| Visibility \(V\) | Effects remain below \(E\) while buffered and cross to external visibility only on downstream acceptance. |
| Recovery \(R\) | Squash or backtrack invalidates unreleased entries. Once accepted externally, local rollback is not claimed. |
| Bounds \(B\) | Queue capacity is finite. The originating operation reserves an entry before modifying state that assumes the effect will be emitted. |
| Timing and locality \(T\) | Use registered ready/valid or credits at the external boundary. Long-latency devices must not create combinational backpressure into execution. |
| Invariant \(I\) | **No effect becomes externally observable before its required architectural and semantic commitment frontiers, and each released effect occurs exactly once.** |

### Operation

Classify irreversible effects; stage complete records; gate release by commitment; retain entries until downstream acceptance.

### Composition

Precise Commit Queue, Commit / Cut Fence, Elastic Channel, Context / Epoch-Carrying Token, and Speculation Shadow Structures.

### GateMate realization

The laboratories use a result commit FIFO. Stores or host messages can use the same substrate with different ordering fields.

### Verification target

Stall the receiver, inject faults and cuts before release, cancel contexts, and count exactly-once accepted transfers.

### Characteristic failure sign

The relational engine unbinds variables immediately after asserting a one-cycle result pulse; a stalled host loses the solution.

### Miniature example

A completed solution tuple remains in the commit FIFO while search state backtracks and discovers later solutions.

**Lineage.** [SMP88], [KHA19]

## Pattern 37: Checkpoint plus Epoch Squash

**Intent.** Recover quickly by restoring compact restart state and invalidating younger work by generation.

### Context and problem

Searching every queue and pipeline to remove wrong-path work is slow and physically expensive.

**Forces.** Epoch invalidation makes squash cheap but leaves dead work consuming resources until it drains and requires wraparound discipline.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Compact checkpoint, live epoch table, epoch in all delayed tokens, and resource-reclamation state. |
| Data and flow \(D,F\) | Create a checkpoint before speculation. Younger work carries the new epoch. On failure, restore checkpoint state and change the live epoch so old work becomes inert. |
| Ordering \(O\) | The restored control and allocation pointers define the new order. Old-epoch completions may arrive in any order but cannot update state. |
| Visibility \(V\) | Squashed work never advances visibility. Restored state resumes at the checkpoint level. |
| Recovery \(R\) | Every delayed exit validates epoch. Resources held by stale work are released as it drains or through an explicit reclamation walk. |
| Bounds \(B\) | Checkpoint count, epoch width, and dead-work occupancy are finite. Admission prevents dead work from permanently exhausting a non-draining resource. |
| Timing and locality \(T\) | Epoch comparison is local and cheap; checkpoint creation and restore must meet the control-path timing target. |
| Invariant \(I\) | **After squash, the visible and live internal state equals the checkpoint, and no token from an invalidated epoch can modify it.** |

### Operation

Choose checkpoint fields; propagate epoch atomically; check every update boundary; define wraparound and dead-resource reclamation.

### Composition

Context / Epoch-Carrying Token, Precise Commit Queue, History / Undo Buffer, and Replayable Operation.

### GateMate realization

Use small epoch counters deliberately in simulation to test wrap handling, then widen for hardware. Keep checkpoint state in registers or compact RAM records.

### Verification target

Delay every response type beyond squash, including faults and credits; assert all stale effects are inert and all capacity returns.

### Characteristic failure sign

Result values are epoch-checked but credit returns are not, so a stale completion corrupts the new generation’s capacity count.

### Miniature example

A dataflow context increments its epoch on cancellation; multiplier results already in the pipeline are dropped at completion.

**Lineage.** [TOM67], [SMP88]

## Pattern 38: Replayable Operation

**Intent.** Speculate on a condition that can be checked later and re-execute locally when the assumption fails.

### Context and problem

Waiting for every alias, bank, cache, or resource condition before issue sacrifices parallelism.

**Forces.** Replay improves utilization only when failures are uncommon and operations are idempotent below their visibility frontier.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Original operation record, operands or producer identities, destination, completion state, replay reason, retry count, and epoch. |
| Data and flow \(D,F\) | Issue provisionally. If a later check fails, clear completion and requeue the same operation without exposing its prior attempt. |
| Ordering \(O\) | Replays preserve the operation’s architectural age. Other independent operations may execute around it subject to commit order. |
| Visibility \(V\) | All attempts remain provisional; only one successful attempt reaches commit. |
| Recovery \(R\) | Fault or excessive retries take a precise slow path. Any temporary resources from a failed attempt are released exactly once. |
| Bounds \(B\) | Replay queues and retry counts are finite. Admission retains sufficient station or history state until success or terminal fault. |
| Timing and locality \(T\) | Detection latency and reissue bandwidth determine benefit. Avoid global pipeline flush when a local replay is sufficient. |
| Invariant \(I\) | **Re-execution at the current visibility level is idempotent, and exactly one successful attempt may commit.** |

### Operation

Identify the late predicate; keep original state; isolate side effects; bound retries; define a terminal path.

### Composition

Reservation Station, Checkpoint plus Epoch Squash, Nonblocking Miss Record, and Delayed Irreversible Store.

### GateMate realization

A small replay FIFO can hold operation indices rather than full records when reservation-station entries remain allocated.

### Verification target

Force repeated conflicts, stall reissue, cancel during replay, and check that visible counters or outputs do not count failed attempts.

### Characteristic failure sign

A replayed operation increments a performance-visible or semantic counter on every failed attempt.

### Miniature example

A bank conflict causes a heap read to retry locally while unrelated arithmetic continues.

**Lineage.** [TOM67], [SMP88]

## Pattern 39: Speculation Shadow Structures

**Intent.** Keep provisional microarchitectural side effects separate from state observable by committed work or another domain.

### Context and problem

Even when registers and stores retire precisely, speculative cache fills, predictor updates, translations, or metadata can leak information or perturb committed execution.

**Forces.** Shadowing reduces leakage but duplicates storage, adds merge paths, and may still leave unmodeled channels.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Per-context or per-epoch shadow entries, ownership metadata, merge or discard controller, and protected shared structure. |
| Data and flow \(D,F\) | Speculative accesses update shadow state. Commitment merges authorized entries; squash discards them without touching shared committed state. |
| Ordering \(O\) | Merges obey architectural and coherence order. Speculative access order remains private to the shadow domain. |
| Visibility \(V\) | Shadow effects remain below the chosen visibility boundary until merge. |
| Recovery \(R\) | Squash invalidates shadow entries by epoch. Faults during merge stop at a precise entry or use an atomic merge protocol. |
| Bounds \(B\) | Shadow capacity is finite. Full may stall speculation or selectively bypass only operations proven non-observable. |
| Timing and locality \(T\) | Placement should keep shadows near clients and bound merge bandwidth. Tag comparison and partitioning must be included in timing and power analysis. |
| Invariant \(I\) | **No uncommitted context can alter protected shared state through a path covered by the shadow contract.** |

### Operation

Enumerate every persistent microarchitectural effect; choose shadow scope; gate merge by commitment; audit bypasses.

### Composition

Context / Epoch-Carrying Token, Future / Shadow File, Delayed Irreversible Store, and Capability / Descriptor Gate.

### GateMate realization

A teaching design can shadow a small result cache or descriptor cache per context and compare traces with speculation disabled.

### Verification target

Probe shared-state differences after squashed operations, fill shadows, interleave domains, and test merge/squash races.

### Characteristic failure sign

Data values are shadowed, but replacement-policy state is updated in the shared cache and remains observable.

### Miniature example

A speculative term-descriptor lookup fills a private descriptor cache that merges only when the operation commits.

**Lineage.** [KHA19]

## Chapter review

The recovery, speculation, and precise visibility should not be selected by name alone. For each candidate, write its nine-field contract, identify its commitment level, and measure whether its added state and routing remove the dominant semantic cost.

## Exercises

1. Choose two patterns from this chapter and test their composition across all eight closure dimensions from Chapter 5.
2. Rewrite one pattern for an implementation with no block RAM, then for an implementation with abundant dual-port RAM.
3. Identify a pattern whose invariant is satisfied locally but can still fail globally because of missing metadata or rollback ownership.
4. Create a negative test that produces the characteristic failure sign of one pattern without violating unrelated interfaces.

# Allocation, collection, and memory latency

## Learning objectives

After this chapter, you should be able to recognize the family’s recurring problem, fill out its pattern contracts, and select compatible patterns for a measured bottleneck.

Patterns 40-44 organize dynamic storage and tolerate long memory operations.

## Pattern 40: Bump-Allocated Generation

**Intent.** Make common allocation a bounds check followed by a pointer increment.

### Context and problem

Free-list search and per-object metadata are unnecessary for short-lived objects allocated and reclaimed in groups.

**Forces.** Sequential allocation is fast and checkpoint-friendly, but long-lived objects need copying, promotion, or another region.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Region base, allocation top, limit, object layout, initialization state, and optional generation or ownership tag. |
| Data and flow \(D,F\) | Reserve contiguous words, initialize them, then publish the reference. Reclaim by resetting a region top or by collector processing. |
| Ordering \(O\) | Initialization precedes publication. Multiple allocators require partitioning or atomic reservation. |
| Visibility \(V\) | Uninitialized reserved space is not visible. Published objects follow normal semantic commitment. |
| Recovery \(R\) | Allocation failure leaves `top` unchanged. A saved top supports wholesale rollback of younger allocations. |
| Bounds \(B\) | Region size is finite. Reservation checks the complete object size and any alignment before changing top. |
| Timing and locality \(T\) | The common path is a small adder and comparator. Initialization bandwidth and RAM ports often dominate. |
| Invariant \(I\) | **Every published reference denotes a fully initialized object within the allocated region, and regions do not overlap.** |

### Operation

Reserve, initialize, publish; save tops at rollback boundaries; choose promotion or collection for survivors.

### Composition

Semantic Region Partition, Lifetime Promotion on Escape, Nursery plus Remembered Set, and Mutation Trail.

### GateMate realization

Keep `heap_top` in a register and use BRAM for nodes. Multiword objects use a valid-last header or remain unreachable until the final write.

### Verification target

Allocate at the boundary, force exact-full and one-too-large cases, interrupt initialization, and roll back to saved tops.

### Characteristic failure sign

The reference is returned before the second word of the object is initialized, allowing another engine to read garbage.

### Miniature example

The lazy reducer allocates graph nodes sequentially; a search checkpoint saves only `heap_top` for younger-node rollback.

**Lineage.** [KOG91], [LIE83]

## Pattern 41: Nursery plus Remembered Set

**Intent.** Collect recent allocations frequently without scanning the entire older heap.

### Context and problem

Most objects may die young, but old objects can point into the young region and must be treated as roots during a young collection.

**Forces.** The nursery makes common allocation and collection cheap; every old-to-young write needs a reliable barrier and remembered-set capacity.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Young region, old region, remembered set or card table, age or region metadata, and collector scan state. |
| Data and flow \(D,F\) | Allocate in the nursery. On an old-object write that creates a young reference, record the source object or card before the edge can be missed by collection. |
| Ordering \(O\) | Barrier record publication precedes or is atomic with the pointer write according to the collector invariant. |
| Visibility \(V\) | Object semantics are unchanged; generation is management metadata unless exposed for diagnostics. |
| Recovery \(R\) | Collector and mutator coordinate through a defined snapshot, coloring, or stop-the-world point. Remembered-set overflow stalls or triggers a full scan. |
| Bounds \(B\) | Nursery and remembered set are finite. Duplicate source records may be permitted if idempotent. |
| Timing and locality \(T\) | Barrier latency belongs on the write path. Card granularity trades storage for rescanning bandwidth. |
| Invariant \(I\) | **Every live old-to-young edge is either represented in the remembered set or discoverable through another guaranteed root path.** |

### Operation

Choose collection invariant and card/object granularity; centralize old-object writes; make overflow safe.

### Composition

Bump-Allocated Generation, Incremental Collection Barrier, Atomic Metadata Propagation, and Semantic Region Partition.

### GateMate realization

A compact experiment can use one bit per old-region card in registers and scan remembered cards through a BRAM reader.

### Verification target

Create and remove cross-generation edges, overflow the set, duplicate records, and compare live-object reachability with a full graph scan.

### Characteristic failure sign

A vector or DMA store bypasses the barrier and creates an unrecorded old-to-young edge that collection reclaims.

### Miniature example

A promoted closure is old; updating one captured field to a newly allocated pair sets the corresponding remembered card.

**Lineage.** [LIE83]

## Pattern 42: Incremental Collection Barrier

**Intent.** Maintain a collector invariant while mutator and collector steps interleave.

### Context and problem

A collector that assumes a static graph can miss references changed during traversal, while stopping the whole machine may violate latency goals.

**Forces.** Read or write barriers add cost to frequent accesses. Different collector algorithms require different barrier semantics.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Collector phase, object color or forwarding state, barrier queue, mutator operation, and optional snapshot metadata. |
| Data and flow \(D,F\) | On a relevant read or write, the barrier shades, records, forwards, or redirects references before the collector invariant could be violated. |
| Ordering \(O\) | Barrier action is ordered with the mutator access as required by the chosen algorithm. Collector work may otherwise interleave. |
| Visibility \(V\) | The mutator observes the same abstract object graph; color and forwarding are internal representation state. |
| Recovery \(R\) | Barrier queue overflow cannot silently skip work. It stalls, performs work inline, or invokes a safe global phase. |
| Bounds \(B\) | Queue, mark stack, and phase metadata are finite. Each access path has a defined pressure response. |
| Timing and locality \(T\) | Barrier checks should use hot metadata close to the object path. Long collector operations are decoupled through queues. |
| Invariant \(I\) | **After every interleaving of mutator and collector transitions, the selected reachability or snapshot invariant remains true.** |

### Operation

Select the collector first; derive the exact barrier; enumerate all mutator paths including assists, DMA, atomics, and debug.

### Composition

Nursery plus Remembered Set, Indirection / Forwarding Cell, Structure / Representation Firewall, and Speculation Shadow Structures.

### GateMate realization

Implement the barrier at one heap-access firewall. Keep phase and hot color bits compact; stream collector work through elastic queues.

### Verification target

Model-check small heaps across all interleavings, stall barrier queues, and compare reclaimed sets with a stop-the-world reference collector.

### Characteristic failure sign

Normal scalar stores use the barrier, but thunk updates through a dedicated reducer port do not.

### Miniature example

During incremental marking, replacing a thunk with a reference to a white result records or shades that result before the edge becomes visible.

**Lineage.** [LIE83]

## Pattern 43: Nonblocking Miss Record

**Intent.** Track outstanding memory misses so independent accesses continue and later requests can merge safely.

### Context and problem

A blocking cache or memory port stalls all clients behind one long-latency line even when their accesses are independent.

**Forces.** Miss records improve concurrency but require response matching, merge lists, fault distribution, and finite entry handling.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Miss address, transaction or generation tag, state, waiting consumers, context/epoch, fault status, and returned data buffer. |
| Data and flow \(D,F\) | On miss, allocate a record and launch one transaction. A later access to the same block merges when legal. Response fills the record and wakes consumers. |
| Ordering \(O\) | Per-address ordering and memory consistency determine which requests may merge or bypass. Responses match by transaction identity. |
| Visibility \(V\) | Returned data remains provisional to each consumer until its operation commits. |
| Recovery \(R\) | Faults are delivered to every affected consumer under defined semantics. Cancelled consumers release their waiter entries without losing the transaction owner. |
| Bounds \(B\) | Record count and waiters per record are finite. Full backpressures new misses while hits may continue if resources permit. |
| Timing and locality \(T\) | Associative address lookup and merge fanout can be expensive. Bank records by index and bound waiter lists. |
| Invariant \(I\) | **Every memory response maps to exactly one live transaction generation, and every accepted waiter receives one matching data or fault result.** |

### Operation

Define allocate, merge, response, wake, cancellation, and release transitions; carry generations through the external memory path.

### Composition

Access / Execute Decoupling, Producer Tag as Future, Tagged Result Wakeup, and Replayable Operation.

### GateMate realization

Even without an external cache, use the pattern for variable-latency host memory or a shared heap service. Small arrays fit registers.

### Verification target

Return responses out of order, merge reads, cancel waiters, recycle transaction tags, and exhaust records.

### Characteristic failure sign

A response uses only the line index, so a delayed old miss fills a newly allocated record for a different address.

### Miniature example

Two graph walkers request the same absent heap block; one transaction is launched and both continuations are woken on return.

**Lineage.** [KRO81]

## Pattern 44: Bank by Access Function

**Intent.** Partition storage according to semantic access behavior rather than only low address bits.

### Context and problem

A uniform multiported memory is expensive when different streams have predictable, distinct read/write patterns.

**Forces.** Functional banking reduces port conflicts but complicates objects whose roles change and operations that cross banks.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Named banks, mapping function from semantic role to bank/address, arbitration for cross-role access, and migration or forwarding state. |
| Data and flow \(D,F\) | Each operation selects the bank appropriate to its access function. Role changes use an explicit move, promotion, or forwarding transition. |
| Ordering \(O\) | Per-bank order is maintained; cross-bank operations define an order or use a transaction record. |
| Visibility \(V\) | Bank choice is normally internal. Moving between regions may change lifetime or commitment policy. |
| Recovery \(R\) | Bank conflict stalls or replays. Migration failure leaves the source valid and unpublished destination state reclaimable. |
| Bounds \(B\) | Each bank has finite capacity and ports. Admission accounts for the target bank, not only total free words. |
| Timing and locality \(T\) | Banking is a physical-locality pattern. Place engines near their dominant banks and connect exceptional cross-bank access elastically. |
| Invariant \(I\) | **Every logical object maps to one authoritative bank and address, and all permitted accesses reach that authority or a valid forwarding representative.** |

### Operation

Classify access streams by read/write ratio, lifetime, rollback, and engine locality; measure conflicts; provide role-change paths.

### Composition

Semantic Region Partition, Local Systolic Cell, Access / Execute Decoupling, and Indirection / Forwarding Cell.

### GateMate realization

Give trail writes, graph reads, continuation accesses, and result commits separate inferred memories before attempting a heavily multiported unified store.

### Verification target

Generate worst-case cross-bank patterns, fill one bank while others are empty, migrate live objects, and count conflicts.

### Characteristic failure sign

Total memory has space but the construction bank is full and no spill policy exists, so allocation deadlocks despite apparent capacity.

### Miniature example

The relational engine separates read-mostly fact RAM from write-heavy binding and trail memories.

**Lineage.** [KOG91], [KUN82]

## Chapter review

The allocation, collection, and memory latency should not be selected by name alone. For each candidate, write its nine-field contract, identify its commitment level, and measure whether its added state and routing remove the dominant semantic cost.

## Exercises

1. Choose two patterns from this chapter and test their composition across all eight closure dimensions from Chapter 5.
2. Rewrite one pattern for an implementation with no block RAM, then for an implementation with abundant dual-port RAM.
3. Identify a pattern whose invariant is satisfied locally but can still fail globally because of missing metadata or rollback ownership.
4. Create a negative test that produces the characteristic failure sign of one pattern without violating unrelated interfaces.

# Instruction and internal-operation patterns

## Learning objectives

After this chapter, you should be able to recognize the family’s recurring problem, fill out its pattern contracts, and select compatible patterns for a measured bottleneck.

Patterns 45-50 connect semantic transitions to compact code and accelerators.

## Pattern 45: Semantic Transition Family

**Intent.** Expose a stable abstract-machine transition family rather than encoding one transient source-language syntax directly.

### Context and problem

General CPUs may spend many operations reconstructing a frequent semantic transition, while language-specific opcodes can become brittle and underused.

**Forces.** A useful family must be frequent, stable, precisely faultable, and lowerable in software. Over-specialization creates permanent ISA cost.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Operation code, explicit semantic operands, mode or descriptor, progress state, result and fault fields, and optional software-fallback address. |
| Data and flow \(D,F\) | Each operation implements one named abstract transition or a bounded family with an exact decomposition into simpler transitions. |
| Ordering \(O\) | The family declares dependencies and commitment like any other operation. Internal implementation may be multi-cycle or decoupled. |
| Visibility \(V\) | Visible behavior is the corresponding abstract transition, not internal micro-steps. |
| Recovery \(R\) | Uncommon inputs, resource pressure, or interruption take a precise assist or decomposed sequence from a named progress point. |
| Bounds \(B\) | Any internal stack, queue, allocation, or log demand is reserved or exposed as a precise capacity fault. |
| Timing and locality \(T\) | Encoding density and front-end simplicity compete with semantic richness. Internal regular operations can isolate backend timing. |
| Invariant \(I\) | **For every operation instance, hardware behavior equals the specified abstract transition or its defined fault, independent of implementation path.** |

### Operation

Profile transition sequences; select stable boundaries; specify software expansion first; then add a direct implementation only when measured.

### Composition

Abstract-Machine Contract, Dense Veneer / Regular Internal Core, Common Fast Path / Precise Slow Path, and Fused Semantic Assist.

### GateMate realization

A teaching core can encode `DEREF`, `BIND`, `FORCE`, or `COMMIT` as bytecodes while implementing them with shared engines.

### Verification target

Compare direct execution against software decomposition for all tags, faults, stalls, and capacity edges.

### Characteristic failure sign

An instruction named after a high-level feature performs hidden allocation and I/O with no precise intermediate state or fallback.

### Miniature example

`BIND` dereferences one variable, reserves a trail slot if required, writes the binding, and reports success as one semantic transition.

**Lineage.** [WAR83], [KOG91]

## Pattern 46: Type / Shape Dispatch

**Intent.** Combine tag or shape classification with selection of the next semantic action.

### Context and problem

Long chains of compare-and-branch operations dominate interpreters, unifiers, reducers, and managed runtimes.

**Forces.** Wide dispatch tables reduce control depth but consume encoding and can amplify tag fanout. Unknown forms need a safe default.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Input tag or descriptor class, dispatch table or logic, selected target, default fault/assist, and optional profile counters. |
| Data and flow \(D,F\) | Read metadata, select a transition or target, and carry the original value unchanged to that target. |
| Ordering \(O\) | Dispatch itself does not reorder operations; selected paths preserve the enclosing machine order. |
| Visibility \(V\) | The classification is internal. Any fault or selected operation crosses the normal commit boundary. |
| Recovery \(R\) | Unknown, protected, malformed, or cold forms select a precise default path rather than aliasing a valid case. |
| Bounds \(B\) | Table entries and target encodings are finite. Dynamic tables require protected updates and a valid default during reconfiguration. |
| Timing and locality \(T\) | Keep hot tag decode local and register high-fanout outputs. Hierarchical dispatch can be faster than one wide table. |
| Invariant \(I\) | **Every legal metadata state selects exactly one permitted action, and every illegal or unsupported state selects the defined default.** |

### Operation

Enumerate the complete tag space; define priority only where overlaps are intentional; carry original operands into the target path.

### Composition

Tagged Value Word, Structure / Representation Firewall, Mode-Dependent Transition, and Common Fast Path / Precise Slow Path.

### GateMate realization

Use CPE LUT trees for compact tag classification and a ROM for larger constructor or opcode tables.

### Verification target

Sweep all tag values, mutate table entries if programmable, and prove unsupported encodings cannot reach privileged paths.

### Characteristic failure sign

Reserved tag `0xE` falls through to the integer case because the `case` statement lacks a default and simulation X behavior differs from synthesis.

### Miniature example

The reducer dispatches `INT`, `IND`, `ADD`, `MUL`, `THUNK`, `BLACKHOLE`, and `ERROR` nodes to distinct controller transitions.

**Lineage.** [WAR83], [KOG91]

## Pattern 47: Mode-Dependent Transition

**Intent.** Reuse one operation family across closely related phases while carrying the phase explicitly with the context.

### Context and problem

Read versus construct, match versus allocate, normal versus replay, or consume versus produce often share encoding and datapath but differ in side effects.

**Forces.** Mode reuse saves decode and code density, but hidden global mode breaks overlap, checkpointing, and reentrancy.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Explicit mode in the operation or context, shared operands, phase-specific destination and side-effect state, and checkpointed mode. |
| Data and flow \(D,F\) | The same operation code selects a mode-specific transition. Every delayed request carries the mode or an immutable context identity that resolves it. |
| Ordering \(O\) | Mode changes are ordered semantic transitions. Concurrent contexts may hold different modes. |
| Visibility \(V\) | Mode-specific effects cross the normal visibility frontier; the mode itself may be semantic state if it changes interpretation. |
| Recovery \(R\) | Checkpoint and replay restore the exact mode. Invalid mode/opcode combinations fault before mutation. |
| Bounds \(B\) | Mode encoding is finite. Mode-specific resource needs are reserved separately. |
| Timing and locality \(T\) | Shared datapaths reduce area, but late mode muxes can lengthen timing. Decode mode early into local control. |
| Invariant \(I\) | **An operation is interpreted using the mode belonging to its own context and checkpoint, never a stale global latch.** |

### Operation

List transition behavior per mode; carry mode through all queues; save it in checkpoints; provide explicit mode-switch operations.

### Composition

Type / Shape Dispatch, Context / Epoch-Carrying Token, Choice-Point Snapshot, and Semantic Transition Family.

### GateMate realization

Store mode in the context record and include it in request tokens only when engines can overlap mode changes.

### Verification target

Interleave contexts in different modes, squash and restore around a mode change, and inject invalid combinations.

### Characteristic failure sign

A global unification mode changes while an older RAM response is in flight, causing a read-phase operation to allocate as though constructing.

### Miniature example

One `UNIFY_FIELD` opcode compares an existing field in read mode and writes a newly constructed field in construct mode.

**Lineage.** [WAR83]

## Pattern 48: Run-Length Semantic Operation

**Intent.** Compress repeated transitions whose intermediate values are not semantically observed.

### Context and problem

Fetch and decode bandwidth can dominate runs of `skip`, `pop`, `clear`, `allocate`, or anonymous-field operations.

**Forces.** One counted operation reduces code size but needs an interruption and fault model for partial progress.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Base operation, count, current index or remaining count, operands, and progress checkpoint. |
| Data and flow \(D,F\) | Execute the repeated elementary transition until count reaches zero, a stall occurs, or a fault is detected. Preserve a precise index. |
| Ordering \(O\) | The operation is equivalent to the ordered sequence of elementary transitions. No hidden reordering changes exceptions. |
| Visibility \(V\) | Intermediate effects are visible only as permitted by the elementary sequence; externally atomic behavior requires buffering or rollback. |
| Recovery \(R\) | On interruption or fault, expose or retain the exact completed prefix and resume, roll back, or decompose according to the contract. |
| Bounds \(B\) | Count width is finite. Resource reservation may cover all iterations or proceed one precise iteration at a time. |
| Timing and locality \(T\) | A small loop controller saves instruction bandwidth. Unrolling is used only when ports and timing justify it. |
| Invariant \(I\) | **The final state and fault point equal execution of exactly the specified number of elementary transitions in order.** |

### Operation

Specify elementary semantics first; add a progress index; decide whether the counted form is restartable, atomic, or decomposed.

### Composition

Semantic Transition Family, Fused Semantic Assist, Precise Commit Queue, and Common Fast Path / Precise Slow Path.

### GateMate realization

Use a counter and one transition per cycle for the first FPGA implementation. Compare code size and cycles against expanded bytecode.

### Verification target

Interrupt at every index, exhaust resources mid-run, and compare the completed prefix with expanded execution.

### Characteristic failure sign

A counted allocation updates `heap_top` for all objects before initialization, exposing holes when a later iteration faults.

### Miniature example

`POP_N 6` trims six dead stack slots while retaining an index that makes interruption precise.

**Lineage.** [WAR83], [KOG91]

## Pattern 49: Dense Veneer / Regular Internal Core

**Intent.** Combine compact external code with explicit uniform internal operations.

### Context and problem

Dense encodings save instruction memory but burden every backend stage with irregular fields and hidden dependencies.

**Forces.** Front-end expansion adds bandwidth and state; the backend gains simpler scheduling, faults, tracing, and composition.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | External instruction stream, decoder/expander, internal operation queue, explicit operands and metadata, and mapping for faults or debug. |
| Data and flow \(D,F\) | Fetch compact instructions, expand each into one or more regular internal operations, then execute and commit using uniform contracts. |
| Ordering \(O\) | Expansion preserves program order and associates every internal operation with its external instruction and progress index. |
| Visibility \(V\) | Only the completed external semantic transition is architectural unless the ISA defines intermediate visibility. |
| Recovery \(R\) | Faults report the external instruction and exact internal progress. Squash discards younger expanded operations. |
| Bounds \(B\) | Expansion queue and maximum expansion length are finite. Fetch stalls before accepting an instruction that cannot be represented safely. |
| Timing and locality \(T\) | Decode can be pipelined; regular internal fields reduce muxing later. Excessive expansion can create front-end throughput limits. |
| Invariant \(I\) | **The internal sequence refines the external instruction exactly, including fault, interruption, and commitment behavior.** |

### Operation

Define regular internal fields; write a software expander; bound expansion; retain source-instruction identity through commit.

### Composition

Semantic Transition Family, Fused Semantic Assist, Reservation Station, and Precise Commit Queue.

### GateMate realization

A 20-bit bytecode ROM can feed 40- or 80-bit internal operation records through a small elastic expansion FIFO.

### Verification target

Compare external-instruction reference execution with internal commits, force maximum expansion, and fault at every internal step.

### Characteristic failure sign

An external instruction is retired when its first internal operation completes even though later internal operations can fault.

### Miniature example

A compact relation bytecode expands into `SCAN`, `DEREF`, `BIND`, `PUSH_CHOICE`, and `EMIT` internal requests.

**Lineage.** [WAR83], [PAT81]

## Pattern 50: Fused Semantic Assist with Decomposition Escape

**Intent.** Accelerate a frequent sequence while defining correctness by an equivalent unfused sequence and retaining a safe escape path.

### Context and problem

Repeated transitions may have redundant decode, routing, and intermediate storage, but a monolithic fused unit becomes hard to fault, interrupt, or generalize.

**Forces.** Fusion improves hot-path latency and bandwidth only when common. Decomposition preserves completeness and reduces permanent complexity.

### Contract

| Field | Obligation |
|---|---|
| State \(S\) | Fused operands, internal progress, buffered or logged effects, result, fault, and escape continuation into the elementary sequence. |
| Data and flow \(D,F\) | Attempt the fused path after checking common preconditions. On an uncommon case, complete or roll back partial work and continue with the equivalent elementary operations. |
| Ordering \(O\) | The fused operation preserves the elementary sequence’s ordering and exception point unless stronger atomic semantics are explicitly specified. |
| Visibility \(V\) | Only effects allowed by the decomposed sequence become visible, through the same commitment gates. |
| Recovery \(R\) | Every internal step has a progress marker or is isolated behind a checkpoint. Escape resumes with original or precisely updated state. |
| Bounds \(B\) | The assist reserves maximum common-case resources or checks capacity at precise substeps. Escape queues are finite. |
| Timing and locality \(T\) | Fusion is justified by measured transition frequency and data movement. Long control paths should be microsequenced rather than combinational. |
| Invariant \(I\) | **For all inputs, the fused path plus any escape produces the same abstract state, visible events, and fault point as the defined elementary sequence.** |

### Operation

Write and verify the decomposition first; profile; identify a side-effect-safe common predicate; buffer effects; retain an always-available escape.

### Composition

Common Fast Path / Precise Slow Path, Run-Length Semantic Operation, Semantic Transition Family, and Dense Veneer / Regular Internal Core.

### GateMate realization

Implement the assist as a small FSM sharing existing RAM ports and engines. Count fused successes, escapes, and saved transitions.

### Verification target

Differentially test fused versus unfused mode for random terms, stalls, capacity edges, and injected faults at every progress state.

### Characteristic failure sign

The assist allocates and publishes two nodes, then discovers an uncommon tag and restarts the entire sequence, duplicating the objects.

### Miniature example

A fused `DEREF_AND_BIND` handles the common unbound-variable case in a few cycles and escapes to general unification for structures or long chains.

**Lineage.** [KOG91], [WAR83]

## Chapter review

The instruction and internal-operation patterns should not be selected by name alone. For each candidate, write its nine-field contract, identify its commitment level, and measure whether its added state and routing remove the dominant semantic cost.

## Exercises

1. Choose two patterns from this chapter and test their composition across all eight closure dimensions from Chapter 5.
2. Rewrite one pattern for an implementation with no block RAM, then for an implementation with abundant dual-port RAM.
3. Identify a pattern whose invariant is satisfied locally but can still fail globally because of missing metadata or rollback ownership.
4. Create a negative test that produces the characteristic failure sign of one pattern without violating unrelated interfaces.

# Part III - Reusable FPGA Substrates

# Tagged records and event envelopes

## Learning objectives

You should be able to define one canonical value record, build event envelopes that preserve context and recovery metadata, and keep architectural representation separate from physical packing.

## A common value package

The laboratories use a 40-bit physical value. The tag and flags are part of the value, not sideband signals.

```systemverilog
package symbolic_types_pkg;

  typedef enum logic [3:0] {
    TAG_INT       = 4'h0,
    TAG_BOOL      = 4'h1,
    TAG_REF       = 4'h2,
    TAG_ATOM      = 4'h3,
    TAG_PAIR      = 4'h4,
    TAG_THUNK     = 4'h5,
    TAG_IND       = 4'h6,
    TAG_ERROR     = 4'hD,
    TAG_POISON    = 4'hE,
    TAG_EMPTY     = 4'hF
  } value_tag_t;

  typedef struct packed {
    value_tag_t   tag;
    logic [3:0]   flags;
    logic [31:0]  payload;
  } value40_t;

  typedef struct packed {
    logic [7:0]   context;
    logic [7:0]   epoch;
    logic [7:0]   destination;
    logic [7:0]   control;
    logic [7:0]   sequence;
    value40_t     value;
  } event80_t;

endpackage
```

The exact `event80_t` above is 80 bits because its five 8-bit control fields occupy the upper 40 bits. A real design may use fewer context bits and more destination bits. Preserve the conceptual fields even when their widths change.

## Tag semantics

A tag table is a semantic contract:

| Tag | Meaning | Legal common operations | Required unusual behavior |
|---|---|---|---|
| `INT` | signed payload integer | arithmetic, compare | overflow policy |
| `BOOL` | Boolean payload | condition, equality | reject noncanonical payloads or normalize |
| `REF` | reference to a cell | dereference, capability check | stale/cycle fault |
| `ATOM` | interned symbol | equality, dispatch | table-miss policy |
| `PAIR` | specialized pair reference or inline pair | field selection | general-object fallback |
| `THUNK` | deferred computation | force | claim and blackhole |
| `IND` | forwarding reference | resolve | depth/cycle policy |
| `ERROR` | defined exceptional value | propagate or catch | preserve origin metadata |
| `POISON` | invalid provisional value | block commit | report source on use |
| `EMPTY` | unused storage | none | never commit as ordinary data |

A tag is not a substitute for bounds, authority, ownership, or generation. A `REF` payload still needs a defined address space and, in a protected design, a descriptor or capability gate.

## Canonical constructors

Define constructors and selectors in one package so that reserved fields and canonical encodings remain consistent.

```systemverilog
function automatic value40_t make_int(input logic signed [31:0] x);
  value40_t v;
  v.tag     = TAG_INT;
  v.flags   = 4'b0000;
  v.payload = x;
  return v;
endfunction

function automatic value40_t make_bool(input logic x);
  value40_t v;
  v.tag     = TAG_BOOL;
  v.flags   = 4'b0000;
  v.payload = {31'b0, x};
  return v;
endfunction

function automatic logic is_canonical_bool(input value40_t v);
  return v.tag == TAG_BOOL && v.payload[31:1] == '0;
endfunction
```

Do not scatter literal tag values through controllers. Centralization is part of the Structure / Representation Firewall.

## Error as data versus fault as control

There are two useful error models:

- **Error value:** an operation produces `TAG_ERROR`; downstream graph nodes propagate or handle it like a value.
- **Precise machine fault:** the current abstract transition does not commit, and the machine enters a fault state with original operands preserved.

The stack evaluator uses precise faults for malformed bytecode and type errors. The dataflow engine can use error values because independent contexts should continue. The relational engine may use both: a query-level error value for a malformed term and a machine-level fault for trail corruption.

Choose the model per boundary. Converting a precise fault into an error value is itself a semantic transition.

## Event identity

Every delayed event should answer five questions:

```text
whose work?        context
which generation?  epoch
where next?         destination
what kind?          control
which occurrence?   sequence or producer identity
```

Not every field needs to be globally unique. The tuple must be unique over the maximum lifetime of any delayed response. A small experiment can deliberately use a four-bit sequence field to force wraparound tests. The deployed design must prevent ambiguity through a wider field or a quiescent-reuse rule.

## Physical packing is a refinement

Suppose the logical event requires 68 bits. Three implementations are possible:

1. store it in an 80-bit RAM word;
2. compress fields into 60 bits and use parity or reserved bits;
3. store a 40-bit value and a 28-bit control record in separate memories.

The third choice is correct only if Atomic Metadata Propagation remains satisfied. The two memories need one command identity, aligned valid state, and coupled flow control. A single 80-bit word is often preferable early in development because it removes a class of alignment bugs.

## Trace records

Trace data should be narrower and more semantic than the full state. A 40-bit record might be:

```text
[39:32] event kind
[31:24] context
[23:16] epoch
[15:8]  operation or state
[7:0]   compact argument / occupancy / fault
```

A second record can follow for a wide value only on selected event kinds. This permits a longer history in the same RAM.

## Chapter summary

A canonical record package prevents metadata loss and representation drift. Tags classify values; epochs and destinations classify events. Error values and precise faults serve different purposes. Physical packing may change as long as the complete semantic record remains atomic.

## Exercises

1. Extend `value40_t` with a compact capability summary without increasing width. State the lost payload range and cold-descriptor path.
2. Design a 40-bit trail entry for 1,024 locations with 20-bit old values and four entry kinds.
3. Decide whether `POISON` should propagate as a value or stop the machine in each of the five laboratories.
4. Write a serialization format that transfers `event80_t` across an 8-bit host link without tearing payload and metadata.

# Elastic channels, queues, and credits

## Learning objectives

You should be able to implement ready/valid stages, state their safety and liveness assumptions, break timing loops, and calculate queue or credit requirements for compound operations.

## The ready/valid ownership rule

A transfer occurs on a clock edge for which both `valid` and `ready` are true. Before that edge, the producer owns the item. After it, the consumer owns the accepted copy. While `valid` is asserted and `ready` is deasserted, the producer must hold the item stable.

A one-entry elastic register is:

```systemverilog
module rv_reg #(
  parameter int WIDTH = 40
) (
  input  logic             clk,
  input  logic             rst,

  input  logic             in_valid,
  output logic             in_ready,
  input  logic [WIDTH-1:0] in_data,

  output logic             out_valid,
  input  logic             out_ready,
  output logic [WIDTH-1:0] out_data
);

  logic             full_q;
  logic [WIDTH-1:0] data_q;

  assign in_ready  = !full_q || out_ready;
  assign out_valid = full_q;
  assign out_data  = data_q;

  always_ff @(posedge clk) begin
    if (rst) begin
      full_q <= 1'b0;
    end else if (in_ready) begin
      full_q <= in_valid;
      if (in_valid) begin
        data_q <= in_data;
      end
    end
  end

endmodule
```

This stage supports simultaneous dequeue and enqueue. Reset invalidates occupancy; it does not need to clear `data_q` because invalid data is never observed.

## Core assertions

Adapt reset polarity and temporal syntax to the local tool, but preserve these properties:

```systemverilog
// A blocked producer holds the complete item stable.
assert property (@(posedge clk) disable iff (rst)
  out_valid && !out_ready |=> out_valid && $stable(out_data));

// An empty stage cannot assert output valid.
assert property (@(posedge clk) disable iff (rst)
  !full_q |-> !out_valid);

// Occupancy is Boolean for a one-entry stage.
assert property (@(posedge clk)
  full_q inside {1'b0, 1'b1});
```

Conservation is easiest with sequence numbers in simulation. Increment an ingress model counter on input acceptance and an egress model counter on output transfer. Store expected data in a testbench queue and compare in FIFO order.

## When `in_ready` should be registered

The simple stage has a combinational path from `out_ready` to `in_ready`. A chain of many stages can create a long reverse path. Three remedies are common:

- insert a skid buffer whose upstream readiness is registered;
- use a FIFO that computes full from registered pointers;
- switch to credit-based flow control for a long or deeply pipelined connection.

Do not optimize away elastic boundaries until post-route timing identifies them as a problem. A robust extra cycle is usually cheaper than an untestable fixed-latency assumption.

## FIFO contracts

A synchronous FIFO owns:

```text
storage
read pointer
write pointer
occupancy or wrap state
input and output interfaces
reset policy
```

Important choices include:

- fall-through versus registered output;
- simultaneous read/write behavior at empty and full boundaries;
- whether cancellation removes arbitrary entries or merely makes them stale by epoch;
- whether faults share the same FIFO;
- whether one operation needs several free entries before acceptance.

For symbolic engines, arbitrary deletion is usually unnecessary. Context/epoch tags let stale entries drain and be discarded, greatly simplifying BRAM FIFOs.

## Compound reservations

Suppose one accepted unification request may produce:

```text
up to 2 trail entries
1 completion entry
1 optional error entry
```

There are three safe designs:

1. Reserve all worst-case entries at input acceptance.
2. Split unification into precise sub-transitions, each reserving only its own effect.
3. Hold the original request in a station until downstream capacity becomes available.

A design that accepts first and discovers missing trail capacity after writing one binding violates capacity closure.

## Credits

For a downstream buffer of capacity \(C\):

\[
credits + occupied + reserved = C.
\]

A send consumes a credit; slot release returns one. The release event may be dequeue, commit, cancellation, or completion depending on ownership. Returning a credit when an item is merely inspected can permit overwrite while still in use.

For a forward path with latency \(L_f\), reverse-credit latency \(L_r\), and target initiation interval one, enough buffering is needed to absorb work launched before the sender sees returned capacity. Measurement should determine the practical depth.

## Wait-for graph review

Draw one node for each finite resource and an edge `A -> B` if A cannot release while waiting for B. A cycle is a potential deadlock. For example:

```text
request FIFO full
  waits for worker
worker holds completion
  waits for response FIFO
response FIFO full
  waits for host
host waits for interrupt
interrupt generated only after request FIFO drains
```

The fix can be reserved completion capacity, an out-of-band interrupt, host protocol change, or a drain guarantee. Increasing FIFO depth does not remove a structural cycle.

## Fairness

Safety does not imply progress. A perfectly lossless arbiter can starve one context forever. State environmental and scheduler assumptions:

```text
If out_valid remains asserted, out_ready is asserted infinitely often.
If an entry remains eligible, arbitration selects it within N grants.
Memory eventually returns every accepted transaction or a fault.
```

Bounded fairness is easier to test and reason about than vague eventuality.

## Chapter summary

Elastic channels separate function from latency. Their central safety rule is stable data under stall; their central system risk is a capacity or wait-for cycle. Credits trade a reverse combinational path for explicit accounting. Compound operations reserve all resources needed to remain precise.

## Exercises

1. Modify `rv_reg` to register `in_ready` using a two-entry skid buffer.
2. Derive occupancy transitions for a depth-four FIFO with simultaneous enqueue and dequeue.
3. Write a credit conservation assertion for capacity eight and one reserved emergency slot.
4. Draw the wait-for graph of the dataflow engine and identify which queues must drain independently.

# Checkpoints, trails, and commit queues

## Learning objectives

You should be able to build a reusable rollback substrate, distinguish checkpoint metadata from undo data, and prove that unwind recreates a named state.

## Three different structures

A reversible machine often needs three separate structures:

- **Checkpoint stack:** enough state to resume an alternative.
- **Mutation trail:** old values for selected locations changed after the checkpoint.
- **Commit queue:** effects that must not become externally visible until accepted.

They solve different problems. Combining them into one polymorphic stack usually creates port conflicts and confused lifetime rules.

## Central mutation interface

All reversible writes should pass through one interface:

```text
mutate request:
    context
    epoch
    location
    new value
    active checkpoint mark
    location age / region class

mutate response:
    accepted
    trail entry allocated or suppressed
    completion / fault
```

The mutator determines whether the old value needs logging. A common suppression rule is: state allocated after the checkpoint does not need individual trail entries when rollback discards the entire younger region. This optimization must be proven from allocation age, not guessed from address alone after wraparound.

## Log-before-write pipeline

A synchronous RAM implementation may need:

```text
cycle 0  accept request and reserve trail slot
cycle 1  read old protected value
cycle 2  write trail entry
cycle 3  write new protected value and acknowledge
```

Faster overlap is possible, but the visibility order remains:

```text
reserve -> obtain old value -> publish trail entry -> publish new value
```

If the protected store offers read-during-write behavior, do not rely on vendor ambiguity. Separate the read and write or instantiate a mode with documented semantics.

## Trail record

A generic 40-bit trail record might be:

```text
[39:36] entry kind
[35:24] context and region
[23:12] location
[11:0]  compressed old value or index
```

For the queens solver, a 20-bit entry is enough:

```text
[19:17] column
[16:9]  old domain mask
[8:4]   decision level
[3:0]   flags
```

The record format should reflect the protected store. If old values are 40 bits, a multiword trail entry or a separate old-value RAM may be required.

## Checkpoint publication

A multiword checkpoint must not become visible before all words are written. Two common implementations are:

- write payload words, then set a valid bit in a compact valid array;
- reserve the entry, write all words, then advance the public stack pointer.

The second approach is attractive when only the top pointer defines reachability.

## Unwind engine

A generic unwind sequence is:

```text
while trail_top != target_mark:
    trail_top = trail_top - 1
    entry = trail[trail_top]
    restore protected_store[entry.location] = entry.old_value

restore checkpointed control and region tops
select next alternative or report exhaustion
```

The decrement-before-read convention is important if `trail_top` denotes the first free slot. Document pointer meaning once and use it across simulation, RTL, and trace tools.

Only the unwind engine may write the protected store during restoration. Normal mutation is quiesced or arbitrated after restoration writes.

## Commit and reclaim

Commit usually moves pointers rather than clearing memory. Old contents can remain physically present because they are unreachable. Clearing a large trail or choice RAM wastes cycles and may infer reset logic that prevents BRAM mapping.

A semantic cut can:

```text
choice_top      = entry_choice_mark
trail_base      = current trail_top or reclaim mark
allocation_base = current heap_top
context_epoch   = context_epoch + 1, if stale alternatives exist
```

The exact fields depend on which old references can still arrive.

## Formal properties

Let `snapshot_hash(k)` be the abstract-state hash stored in the testbench when checkpoint `k` is created. After unwind to `k`:

```text
assert abstract_state_hash == snapshot_hash(k)
assert trail_top == checkpoint[k].trail_mark
assert heap_top  == checkpoint[k].heap_mark
assert no valid choice younger than k
```

A stronger proof tracks each protected location and shows that the reverse sequence of trail entries is the inverse of forward mutations.

## Nested commitment levels

If a CPU pipeline sits beneath semantic rollback, document the handoff:

```text
before instruction retirement:
    microarchitectural checkpoint owns provisional writes

after retirement, before search commit:
    semantic trail owns reversible bindings

after cut:
    no local rollback owner
```

One implementation is to delay the actual binding write until retirement, then append or activate the semantic trail entry. Another is to version both records so microarchitectural squash can cancel the not-yet-semantic entry. Do not rely on two independent undo stacks restoring in a coincidentally correct order.

## Chapter summary

Checkpoint, trail, and commit queue are a complete recovery triad: remember where to resume, remember how to reverse selected effects, and delay effects that cannot be reversed. Synchronous RAM timing makes log-before-write order explicit. Reclamation is pointer movement, not clearing.

## Exercises

1. Design a multiword checkpoint entry with valid-last publication.
2. Add trail-entry suppression for locations allocated after the active heap mark and state the proof obligation.
3. Show how repeated writes to one location unwind correctly.
4. Define a protocol that hands a binding from a CPU history buffer to a semantic trail at retirement.

# Synchronous memory and ownership

## Learning objectives

You should be able to wrap synchronous RAM without leaking technology assumptions, define read/write ownership, and choose banking and top-of-structure caches from measured access patterns.

## A technology-neutral read interface

A simple synchronous dual-port wrapper can expose one write command and one read request:

```systemverilog
module sync_sdp_ram #(
  parameter int WIDTH = 40,
  parameter int DEPTH = 512,
  localparam int AW = $clog2(DEPTH)
) (
  input  logic             clk,

  input  logic             wr_en,
  input  logic [AW-1:0]    wr_addr,
  input  logic [WIDTH-1:0] wr_data,

  input  logic             rd_en,
  input  logic [AW-1:0]    rd_addr,
  output logic             rd_valid,
  output logic [WIDTH-1:0] rd_data
);

  logic [WIDTH-1:0] mem [0:DEPTH-1];

  always_ff @(posedge clk) begin
    rd_valid <= rd_en;
    if (rd_en) begin
      rd_data <= mem[rd_addr];
    end
    if (wr_en) begin
      mem[wr_addr] <= wr_data;
    end
  end

endmodule
```

Read-during-write behavior to the same address is intentionally unspecified at the semantic boundary. A client requiring old-data, new-data, or no-change behavior must avoid the collision or use a wrapper whose contract states and implements the selected mode.

## Request identity

One-cycle RAMs still need response identity when several contexts share a port. Register the request metadata alongside the address:

```text
rd_addr_q
rd_context_q
rd_epoch_q
rd_destination_q
rd_kind_q
```

On `rd_valid`, validate epoch and route data using the registered metadata. Treat RAM as a tiny fixed-latency transaction system rather than a wire.

## Mutation ownership

A dual-port RAM permits two physical accesses, not two independent semantic owners. For each logical store, state:

```text
normal read owner
normal write owner
recovery write owner
collector or maintenance owner
debug owner
```

Then define arbitration. A safe priority for a binding store might be:

```text
reset initialization walk
> unwind restore
> committed mutation
> debug read
> normal dereference read
```

The exact priority depends on liveness. A permanent high-priority initialization or debug stream must not starve execution.

## Stack caches

Synchronous RAM adds a cycle to push/pop-heavy structures. A small top cache removes most accesses while preserving a precise logical view.

For a two-entry stack cache:

```text
logical depth 0: no valid top entries
logical depth 1: top0 valid
logical depth >=2: top0 and top1 valid; deeper values in RAM
```

A push shifts `top0` to `top1`; if both are already valid, the old `top1` spills to RAM. A binary operation consumes two tops and places one result in `top0`, potentially refilling `top1` from RAM. The refinement map concatenates cached entries and the live RAM range.

Avoid clever simultaneous spill/refill until the simple FSM passes randomized model comparison.

## Functional banking

The laboratories naturally separate:

- read-mostly program or fact memory;
- push/pop trail memory;
- choice-point memory;
- graph heap with controlled mutation;
- continuation stack;
- operand stores;
- result and trace FIFOs.

This is Bank by Access Function. It reduces port conflicts and clarifies recovery. If one bank becomes the bottleneck, first split its semantic role or replicate read-only data. A heavily multiported unified memory built from registers is usually a poor small-FPGA starting point.

## Initialization

Large arrays should not be synchronously reset unless their primitive supports the desired initialization efficiently. Common alternatives are:

- initialize from a memory file at configuration time;
- reset only pointers and valid bitmaps;
- walk addresses with an initialization FSM;
- use self-reference sentinels established by a short boot sequence;
- treat uninitialized entries as unreachable.

The abstract machine must not observe stale contents merely because they remain physically present.

## Memory arbitration as a protocol

An arbiter is not just a priority mux. It must define:

- request acceptance;
- whether a requester must hold its request;
- response routing;
- fairness;
- cancellation;
- collision behavior;
- write acknowledgment; and
- relationship to commit.

Use ready/valid on both request and response unless fixed latency and exclusive ownership are formally guaranteed.

## Measuring memory pressure

Instrument each store with:

```text
read requests
write requests
conflicts
stall cycles
maximum queue depth
same-address hazards
recovery accesses
maintenance accesses
```

A bank redesign should be justified by these counters. A low utilization percentage can still hide a burst conflict that determines total latency.

## Chapter summary

Synchronous RAM is a transactional resource. Register request identity, define collision behavior, and assign one semantic mutation owner. Top caches and functional banking are refinements selected from access measurements, not assumptions.

## Exercises

1. Add a response metadata pipeline to `sync_sdp_ram` for four contexts.
2. Define a fair arbiter between normal heap reads, thunk claims, and collector accesses.
3. Write the refinement map for a two-entry choice-stack cache.
4. Compare reset-by-clear with valid-bitmap reset for a 512-entry operand store.

# Verification, traces, and experiment control

## Learning objectives

You should be able to combine reference models, assertions, randomized transaction tests, and FPGA traces into one verification strategy.

## Three models

Each laboratory should have three representations:

1. **Semantic model:** a compact interpreter or transition system with no cycle timing.
2. **Transaction model:** queues, finite capacities, and nondeterministic latency, but no vendor primitives.
3. **RTL implementation:** cycle-accurate and synthesizable.

The semantic model detects architectural errors. The transaction model detects flow, cancellation, and capacity errors earlier than RTL. The RTL model detects implementation and timing-interface errors.

## Commit-based comparison

Cycle-by-cycle comparison is too strict when elastic or pipelined implementations differ. Compare at semantic commits.

A commit record can be:

```text
sequence
context and epoch
transition kind
architectural destination
result or state digest
fault
visible output, if any
```

The testbench advances the semantic model until it produces the corresponding transition, then compares. Independent contexts may require per-context commit streams or partial-order comparison.

## State digests

A digest is useful but not sufficient. A simple XOR can hide swaps and paired errors. Prefer a seeded rolling hash over named state components in the testbench. For small states such as eight queens domains, compare the complete state. For hardware trace, a compact digest can identify the first divergent commit, after which the simulation reruns with full visibility.

## Safety properties

Reusable safety assertions include:

```text
no queue underflow or overflow
no write without ownership
no accepted item lost or duplicated
blocked data stable
no stale epoch updates live state
no commit of poison or invalid tag
no rollback below active base
no indirection beyond bound without fault
no external release before commitment
```

Safety finds a finite bad prefix. It should be checked continuously in simulation and, where tool capacity permits, formal property checking.

## Liveness properties

Liveness depends on assumptions:

```text
accepted memory requests eventually return data or fault
an asserted output eventually sees ready
an eligible context is eventually scheduled
unwind eventually reaches its target mark
initialization eventually completes
```

Turn these into bounded properties for the small experiments. For example, an eight-entry round-robin scheduler can promise service within eight grants, excluding cycles in which the selected engine is blocked.

## Randomized pressure

Randomization should target contracts, not merely operands:

- deassert every `ready` independently;
- vary RAM or multiplier response latency in transaction models;
- fill structures to exactly capacity and one request beyond;
- cancel every context at every pipeline stage;
- inject faults before and after reservation points;
- recycle identifiers quickly;
- interrupt multi-cycle assists at every progress index;
- vary simultaneous events such as enqueue plus dequeue or commit plus squash.

Use reproducible seeds and store the seed in the experiment manifest.

## Metamorphic tests

A metamorphic relation predicts how results should change under a transformation:

- inserting elastic stages must not change commit traces;
- increasing FIFO depth must not change results;
- disabling a checked hint must not change results;
- fused and decomposed execution must match;
- reordering independent dataflow input events must not change final values;
- adding unreachable graph nodes must not change reduction;
- renaming logic variables consistently must not change solution structure;
- cut results should match the defined prefix or subset of full search.

These tests are powerful because they do not require a separate expected answer for every random case.

## Hardware trace workflow

A practical FPGA workflow is:

1. Run the same program and seed in simulation.
2. Produce a golden commit trace.
3. Load the program and initial memories into the FPGA.
4. Capture commit events and selected internal transitions with the GateMate ILA or trace RAM.
5. Export the trace and compare sequence, context, transition, digest, and visible events.
6. Locate the first divergence rather than inspecting the final wrong output.

Keep trace overflow explicit. A trace buffer may stop capture, wrap with a generation, or backpressure the machine in a debug build. Silent overwrite invalidates comparison.

## Performance counters

Every experiment reports:

```text
cycles and commits
accepted inputs and emitted outputs
stall cycles by cause
maximum queue, trail, choice, heap, and continuation occupancy
functional-unit utilization
rollback count and entries restored
slow-path and assist count
stale events dropped
fault count by code
```

Counter updates are normally diagnostic and not part of architectural semantics. If read by another protection domain, they can become observable side effects and may need shadowing or access control.

## Exit criteria

A laboratory is complete when:

- directed functional cases pass;
- randomized backpressure passes for many seeds;
- every finite structure reaches its boundary in a test;
- every documented fault is observed and precise;
- model and commit traces agree;
- synthesis maps expected memories and stays within stop-build budgets;
- place and route closes on the selected board constraint;
- hardware trace agrees with simulation; and
- the experiment manifest is archived.

## Chapter summary

Verification follows the same contract structure as design. Semantic models check meaning, transaction models check bounded flow, and RTL checks implementation. Commit traces allow latency-independent comparison. Random pressure and metamorphic tests expose composition errors that ordinary directed tests miss.

## Exercises

1. Write a commit packet for each of the five laboratories.
2. Define a partial-order comparison for two independent dataflow contexts.
3. Propose a metamorphic test for the mutation trail.
4. Decide whether a cycle counter is externally visible in your intended deployment and state the security consequence.

# Part IV - Worked GateMate Laboratories

# Laboratory 1: a precise tagged stack evaluator

## Result

The first machine executes typed stack bytecode. It must produce:

```text
((7 + 5) * 3) == 36  ->  BOOL(true)
```

A second program attempts:

```text
BOOL(true) + INT(4)
```

and must enter `TYPE_FAULT` without changing the stack or advancing `pc` past the failing instruction.

## Patterns composed

| Pattern | Role in the laboratory |
|---|---|
| Abstract-Machine Contract | defines bytecode transitions independently of cycles |
| Explicit Semantic State Vector | names `pc`, stack, output, fault, and halt state |
| Tagged Value Word | distinguishes integers, Booleans, references, and errors |
| Immediate-or-Boxed Split | keeps small integers directly in the 40-bit word |
| Type / Shape Dispatch | selects legal ALU behavior from tags |
| Common Fast Path / Precise Slow Path | handles normal arithmetic directly and faults precisely |
| Split-Lifetime Frame | supports a top-of-stack cache over synchronous RAM |
| Delayed Irreversible Store | commits `EMIT` through a ready/valid FIFO |

![The evaluator checks tags before state mutation and sends output through an explicit commitment boundary.](assets/stack_evaluator.png){ width=92% }

## Semantic machine

Define:

```text
M = <pc, stack, output_stream, fault, halted>
```

`stack` is an ordered sequence of `value40` records. The architectural model does not expose whether the top entries are in registers or RAM.

A minimal 20-bit instruction format is:

```text
[19:15] opcode
[14:0]  signed immediate, branch target, or subfield
```

Use the following initial operation set:

| Operation | Stack effect | Notes |
|---|---|---|
| `PUSH_S15 k` | `-- INT(k)` | sign-extend 15-bit immediate |
| `PUSH_TRUE` | `-- BOOL(1)` | canonical Boolean |
| `PUSH_FALSE` | `-- BOOL(0)` | canonical Boolean |
| `ADD` | `INT a, INT b -- INT(a+b)` | defined overflow policy |
| `SUB` | `INT a, INT b -- INT(a-b)` | top is right operand |
| `MUL` | `INT a, INT b -- INT(a*b)` | low 32 bits or precise overflow |
| `EQ` | `x, y -- BOOL(x==y)` | equality defined by tag policy |
| `LT` | `INT a, INT b -- BOOL(a<b)` | typed compare |
| `DUP` | `x -- x, x` | capacity checked first |
| `DROP` | `x --` | underflow checked first |
| `SWAP` | `x, y -- y, x` | no RAM change if both cached |
| `JMP t` | unchanged | `pc=t` |
| `JZ t` | `BOOL c --` | branch if false; reject non-Boolean |
| `EMIT` | `x --` | pop only on output acceptance |
| `HALT` | unchanged | enter halted state |

The first program is:

```text
address  instruction
0        PUSH_S15 7
1        PUSH_S15 5
2        ADD
3        PUSH_S15 3
4        MUL
5        PUSH_S15 36
6        EQ
7        EMIT
8        HALT
```

The expected commit trace is:

```text
seq  pc  operation      depth  top
0    0   PUSH_S15 7     1      INT(7)
1    1   PUSH_S15 5     2      INT(5)
2    2   ADD            1      INT(12)
3    3   PUSH_S15 3     2      INT(3)
4    4   MUL            1      INT(36)
5    5   PUSH_S15 36    2      INT(36)
6    6   EQ             1      BOOL(true)
7    7   EMIT           0      accepted BOOL(true)
8    8   HALT           0      halted
```

## Precise arithmetic transition

`ADD` is one abstract transition:

```text
precondition:
    depth >= 2

if top1.tag != INT or top0.tag != INT:
    fault = TYPE_FAULT(pc, top1.tag, top0.tag)
    leave pc and stack unchanged

else if addition_overflows(top1.payload, top0.payload):
    fault = ARITH_OVERFLOW(pc)
    leave pc and stack unchanged

else:
    replace top1 and top0 with INT(top1 + top0)
    depth = depth - 1
    pc = pc + 1
```

Here `top0` is the newest stack element. The implementation may detect underflow, fetch a deep operand, calculate, and write over several cycles. No architectural mutation occurs before all checks pass.

## Controller states

A simple BRAM-capable controller uses:

```text
RESET
FETCH_REQ
FETCH_WAIT
DECODE
STACK_READ_REQ
STACK_READ_WAIT
EXECUTE
COMMIT
OUTPUT_WAIT
FAULT
HALTED
```

Not every instruction uses every state. The first implementation should accept a variable number of cycles per bytecode rather than forcing a uniform fixed pipeline.

A commit pulse is generated only in `COMMIT`, on output acceptance in `OUTPUT_WAIT`, or on entry to a precise fault state. The trace packet captures the old and new `pc`, operation, depth, and visible event.

## Register-stack version

Begin with a 16- or 32-entry register stack:

```systemverilog
value40_t stack_q [0:STACK_DEPTH-1];
logic [$clog2(STACK_DEPTH+1)-1:0] depth_q;
```

This version isolates semantic and interface errors from RAM latency. Synthesis area is not the final metric. Its purpose is to validate the model, opcode encoding, fault behavior, result FIFO, and trace comparison.

The state-update style should calculate a next state, then assign registers once:

```systemverilog
always_comb begin
  pc_d    = pc_q;
  depth_d = depth_q;
  fault_d = fault_q;
  commit  = 1'b0;

  case (state_q)
    EXECUTE: begin
      // Validate all preconditions here.
      // Move to COMMIT only with complete next-state data.
    end
    COMMIT: begin
      pc_d   = next_pc_q;
      commit = 1'b1;
    end
  endcase
end
```

Do not write the stack in several unrelated `always_ff` blocks. One semantic mutation owner makes precise fault behavior much easier to prove.

## BRAM stack with two-entry top cache

After the register version passes, introduce:

```text
top0_q       newest value
top1_q       next value
top_count_q  0, 1, or 2
deep_count_q number of live values in RAM
deep_top_q   first free RAM address
```

A useful invariant is:

\[
architectural\ depth = top\_count + deep\_count.
\]

The logical stack from newest to oldest is:

```text
top0, top1, RAM[deep_top-1], RAM[deep_top-2], ...
```

### Push

```text
if top_count == 0:
    top0 = value
    top_count = 1

else if top_count == 1:
    top1 = top0
    top0 = value
    top_count = 2

else:
    reserve one deep RAM slot
    write old top1 to RAM[deep_top]
    deep_top++
    deep_count++
    top1 = top0
    top0 = value
```

### Binary-result commit

When both operands are cached:

```text
top0 = result

if deep_count == 0:
    top_count = 1
else:
    request RAM[deep_top-1]
    on response:
        top1 = response
        deep_top--
        deep_count--
        top_count = 2
```

The abstract `ADD` commit can occur when the new logical stack is completely represented. The implementation may hold `result` while waiting for refill. This is internal stuttering.

## Output commitment

`EMIT` must not pop merely because `out_valid` is asserted. Use a pending-output state:

```text
DECODE EMIT:
    validate depth >= 1
    pending_value = top0
    state = OUTPUT_WAIT

OUTPUT_WAIT:
    out_valid = 1
    out_data  = pending_value

    if out_ready:
        perform stack pop
        pc++
        emit commit packet
        state = FETCH_REQ
```

The pending record must include all metadata. If reset can occur here, define whether the output is discarded and whether software restarts the instruction. For the laboratory, hold reset as a global experiment abort and restart from the initial image.

## Fault state

A fault record might contain:

```text
code
faulting pc
opcode
depth
operand tags
optional payload fragment
```

Capture the first fault and stop. A `CLEAR_FAULT` debug command may restore the initial machine image, but it is outside normal bytecode semantics.

Required fault cases are:

```text
STACK_UNDERFLOW
STACK_OVERFLOW
TYPE_FAULT
ARITH_OVERFLOW
BAD_OPCODE
BAD_BRANCH_TARGET
NONCANONICAL_BOOL
```

Each leaves the stack and `pc` at a precise documented state.

## Verification plan

Directed tests:

1. the expected arithmetic result;
2. every opcode at minimum stack depth;
3. all stack boundary transitions around top-cache spill and refill;
4. all arithmetic boundaries;
5. wrong tags for every typed operation;
6. output stalls of length zero through at least twenty cycles;
7. branch taken and not taken;
8. invalid instruction and branch target.

Random test generation can maintain a typed software stack and choose only legal instructions, then occasionally inject one illegal instruction to test precise fault behavior. Compare after every commit.

Core assertions:

```text
architectural depth = top_count + deep_count
no stack write while faulted
no pc change without commit or defined reset
blocked output remains stable
one output acceptance produces one pop
committed tag is legal and canonical
```

## GateMate mapping and exit criteria

A compact implementation can use one 20-Kbit half for a 1Kx20 instruction ROM and another 20-Kbit half or separate block for a 40-bit stack organization, depending on the inferred mode and desired depth [COL25-D]. Inspect synthesis rather than assuming that source arrays share one physical block.

Stop and simplify if the evaluator exceeds two physical RAM blocks or roughly 2,000 CPEs before debug instrumentation. Likely causes are a register-based deep stack left enabled, wide reset logic, duplicated tag decoders, or an unexpectedly general multiplier.

Hardware evidence of completion is:

```text
one BOOL(true) output transfer
then HALTED

for the bad program:
TYPE_FAULT at the ADD pc
unchanged depth and operand values
no output transfer
```

## Extensions

- Add `CALL`, `RET`, and explicit continuation records.
- Add boxed integers and a heap allocation assist.
- Add a dense external bytecode that expands into regular internal operations.
- Add a replayable multi-cycle multiplier while preserving the same abstract `MUL` transition.
- Add a capability descriptor to `REF` values.

## Exercises

1. Specify precise semantics for signed multiplication overflow.
2. Prove the two-entry top-cache refinement invariant across push, pop, binary operation, and output stall.
3. Add `SELECT` that chooses one of two values based on a Boolean without branching.
4. Design a four-entry return-address stack and state its relationship to architectural continuation state.
5. Compare code density and cycle count with and without `PUSH_TRUE` and `PUSH_FALSE` special opcodes.

# Laboratory 2: rollback constraint solving

## Result

Build an 8-queens solver that uses a choice stack and mutation trail rather than copying the whole board state at every branch. The complete run must emit exactly 92 solutions. With columns selected from left to right and rows selected from least significant bit to most significant bit, the first emitted board must be:

```text
column: 0 1 2 3 4 5 6 7
row:    0 4 7 5 2 6 1 3
```

A `FIRST_ONLY` configuration emits that board, crosses a cut fence, discards every remaining alternative, and halts.

## Patterns composed

| Pattern | Role |
|---|---|
| Explicit Semantic State Vector | names current domains, propagation state, choices, trail, and output |
| Semantic Region Partition | separates current state, restart state, undo data, and results |
| Choice-Point Snapshot | stores only the next alternative and restart marks |
| Mutation Trail | logs old domain masks before narrowing |
| Commit / Cut Fence | discards all alternatives after an accepted first result |
| Run-Length Semantic Operation | performs bounded propagation and unwind loops |
| Delayed Irreversible Store | holds a solved board until the result interface accepts it |
| Common Fast Path / Precise Slow Path | handles ordinary domain narrowing and capacity faults precisely |

![Current domains are small and fast; the trail stores only old values, while the choice stack stores restart metadata.](assets/queens_rollback.png){ width=92% }

## Constraint representation

The board has one queen in each column. For each column `c`, an eight-bit mask records currently permitted rows:

```text
domain[c][r] = 1  means row r remains possible in column c
```

Interpretation:

```text
8'b00000000  contradiction
one-hot       assigned queen
other nonzero unresolved domain
```

The initial state is:

```text
domain[0..7] = 8'hFF
propagated_mask = 8'h00
choice_top = 0
trail_top  = 0
solution_count = 0
```

The current eight domain masks fit naturally in registers. They are read frequently and updated one at a time. The trail and choice stack use RAM.

## Attack calculation

If column `c0` is assigned row `r0`, then another column `c` cannot use:

```text
same row:       r0
rising diagonal r0 + (c - c0)
falling diagonal r0 - (c - c0)
```

Only row indices in `[0,7]` contribute bits. Define:

```text
attack_mask(c0, r0, c):
    mask = 1 << r0
    delta = c - c0

    if 0 <= r0 + delta < 8:
        mask |= 1 << (r0 + delta)

    if 0 <= r0 - delta < 8:
        mask |= 1 << (r0 - delta)

    return mask
```

The updated domain is:

```text
new_domain = old_domain & ~attack_mask
```

The assigned source column is skipped. If `new_domain == old_domain`, no write and no trail entry are needed.

## Abstract state

A complete abstract state is:

```text
M = <domains,
     propagated_mask,
     propagation_cursor,
     choice_stack,
     trail,
     trail_top,
     result_stream,
     first_only,
     status,
     fault>
```

`propagated_mask[c]` is true when the current one-hot assignment in column `c` has already been propagated to all other columns. It is part of restart state because rollback can restore an older domain whose singleton must be propagated again.

## Central reversible write

All domain changes use:

```text
write_domain(c, new_mask):
    if domain[c] == new_mask:
        return NO_CHANGE

    if trail is full:
        return TRAIL_FULL without changing domain

    trail[trail_top] = {
        column = c,
        old_domain = domain[c],
        decision_level,
        flags
    }
    trail_top++
    domain[c] = new_mask
    propagated_mask[c] = 0
    return CHANGED
```

The final `propagated_mask[c] = 0` matters. A domain can become a new singleton after narrowing and must later propagate. For a non-singleton, the bit remains irrelevant but safely clear.

A lower-cost implementation can omit `decision_level` from the trail because choice points already store `trail_mark`. Retain it initially for trace and consistency checking.

## Choice-point record

A compact record contains:

```text
variable             3 bits
remaining_options    8 bits
trail_mark           enough bits for trail depth
saved_propagated     8 bits
resume_state         controller state code
```

For deeper parameterized boards, store the propagated bitmap in another word or recompute it conservatively after rollback.

A choice is made as follows:

```text
v = first unresolved column
chosen = least significant set bit(domain[v])
remaining = domain[v] & ~chosen

reserve choice entry
write complete choice point
advance choice_top

write_domain(v, chosen)
return to propagation
```

The snapshot is published before the branch mutation. If choice capacity is exhausted, the machine raises `CHOICE_FULL` without selecting a row.

## Propagation controller

A direct controller is:

```text
FIND_SINGLETON:
    scan columns for a one-hot domain with propagated_mask[c] == 0

    if found:
        source_column = c
        source_row = index of the one bit
        target_column = 0
        state = PROPAGATE

    else if any domain is zero:
        state = FAIL

    else if all domains are one-hot:
        state = SOLUTION

    else:
        state = MAKE_CHOICE

PROPAGATE:
    if target_column == source_column:
        skip
    else:
        narrow target domain using attack mask

    if narrowing produces zero:
        state = FAIL after the reversible write commits
    else if target_column == 7:
        propagated_mask[source_column] = 1
        state = FIND_SINGLETON
    else:
        target_column++
```

A subtle ordering point arises when narrowing creates a zero domain. The write still belongs to the speculative state and is trailed, so rollback can restore it. The machine may alternatively detect zero before write and enter failure without storing the impossible mask. Either choice is valid if the abstract model and trail accounting agree. The first version should write the narrowed value because it makes the state transition simple and traceable.

## Selecting the branch variable

The required deterministic first solution uses:

```text
first unresolved column
least significant available row
```

A more efficient solver might choose the smallest non-singleton domain. That changes search order but not the set of solutions. Treat the heuristic as a Checked Compiler Hint or configurable scheduling policy. Keep the deterministic baseline for trace comparison.

## Backtracking

A choice point denotes a trail boundary and remaining rows. The unwind sequence is:

```text
BACKTRACK_SELECT:
    if choice_top == 0:
        state = COMPLETE
    else:
        cp = choice[choice_top - 1]
        target_trail = cp.trail_mark
        state = UNWIND

UNWIND:
    if trail_top == target_trail:
        domains and propagated_mask are at the branch entry state
        state = TRY_REMAINING
    else:
        trail_top--
        entry = trail[trail_top]
        domain[entry.column] = entry.old_domain
        continue

TRY_REMAINING:
    propagated_mask = cp.saved_propagated

    if cp.remaining_options == 0:
        choice_top--
        state = BACKTRACK_SELECT
    else:
        chosen = least significant set bit(cp.remaining_options)
        cp.remaining_options &= ~chosen
        write updated top choice point
        write_domain(cp.variable, chosen)
        state = FIND_SINGLETON
```

Only the unwind engine writes domains during `UNWIND`. Normal propagation requests are blocked.

### Why reverse order matters

Suppose one column changes from `FF` to `F0`, then from `F0` to `10`. The trail holds:

```text
(column, FF)
(column, F0)
```

Reverse unwind restores `F0` and then `FF`. Forward replay would finish at `F0`, not the checkpoint state.

## Solution commitment

When all domains are one-hot, encode each row as three bits:

```text
solution24 = {
    row_of_column_7,
    row_of_column_6,
    ...,
    row_of_column_0
}
```

The controller enters `SOLUTION_WAIT`:

```text
result_valid = 1
result_data  = solution24

if result_ready:
    solution_count++

    if FIRST_ONLY:
        choice_top = 0
        trail_base = trail_top
        state = COMPLETE
    else:
        state = BACKTRACK_SELECT
```

Do not backtrack while the result is blocked. The current domains are the semantic evidence for the tuple being offered. An alternative implementation can copy the tuple into a commit FIFO and backtrack immediately; then the FIFO, not the current domains, owns the committed result.

## Faults and pressure

| Condition | Precise behavior |
|---|---|
| `TRAIL_FULL` | stop before the next protected write |
| `CHOICE_FULL` | stop before branch selection |
| result FIFO full | hold solution or pending tuple stable |
| invalid domain index | internal fault with no write |
| invalid one-hot decode | internal consistency fault |
| trail underflow | fatal integrity fault |
| search exhausted | emit one completion status or enter `COMPLETE` |

For a pure 8-queens implementation, shallow fixed memories are sufficient. Deliberately parameterize them smaller in simulation to force the pressure paths.

## Reference model

A concise software model can use the same domains and trail, or a simpler recursive board model. Use both:

- a conventional recursive solver verifies the solution set independently;
- a transaction-level model with domains and trail verifies exact search order and commit trace.

Expected invariants include:

```text
no domain has bits outside 8'hFF
all assigned queens are mutually nonattacking
trail_top never exceeds capacity
choice trail marks never exceed current trail_top at creation
unwind to a mark recreates the saved state hash
solution_count is 92 at complete enumeration
FIRST_ONLY emits exactly one board
```

## Instrumentation

Counters:

```text
cycles
decisions
propagation visits
domain writes
trail pushes
trail pops
contradictions
solutions
maximum trail depth
maximum choice depth
result stall cycles
```

The most useful architectural comparison is trail traffic versus full snapshots. A full-state baseline copies eight domain bytes and propagation metadata at every branch. The trail implementation writes only changed domains. Measure both bytes written and cycles.

## GateMate mapping

A small implementation uses:

- eight domain registers and an eight-bit propagation bitmap;
- one 20-bit or 40-bit trail RAM;
- one compact choice RAM;
- a register or small FIFO for result tuples;
- CPE logic for one-hot detection, first-set-bit selection, masks, and counters.

Stop and simplify around two physical RAM blocks and 3,000 CPEs. Excessive area may indicate generalized N-queens arithmetic, broad variable shifts, or an over-wide trace path. Because the board size is eight, precomputed attack masks in ROM or simple case logic may be better than generic signed index arithmetic.

## Experimental sequence

1. Implement a non-trailed recursive-style FSM that stores the full board at each choice.
2. Validate 92 solutions and deterministic order.
3. Replace snapshots of domains with choice points plus trail.
4. Compare write traffic and maximum storage.
5. Add randomized result backpressure.
6. Add deliberately shallow trail and choice capacities.
7. Add `FIRST_ONLY` cut and prove no restore below the cut frontier.
8. Capture a complete first-solution trace in simulation and on FPGA.

## Extensions

- Parameterize to N up to 12 with wider domains.
- Add minimum-domain branching as a checked heuristic.
- Generalize the reversible domain store into a graph-coloring engine.
- Add clause-learning-style conflict records, carefully separating them from the LIFO trail.
- Run two search contexts with epochs and a shared propagator.

## Exercises

1. Prove that the saved propagated bitmap plus trail unwind is sufficient to resume correctly.
2. Derive the maximum possible trail entries for the fixed 8-queens algorithm under the stated propagation order.
3. Replace the trail with a persistent copy-on-write domain store and compare state and bandwidth.
4. Implement a cut after the first ten solutions and define the exact visible result set.
5. Add a branch heuristic without changing the complete solution multiset; state which observable order changes.

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
3. Design a waiter record that permits two evaluators and finite suspension capacity.
4. State the observable difference between fail-fast and evaluate-both-sides error policies.
5. Add a `PAIR` node and define lazy versus strict field evaluation.

# Laboratory 5: a relational query engine

## Result

Load the relation:

```text
edge(1,2)
edge(1,3)
edge(2,4)
edge(2,5)
edge(3,5)
edge(5,6)
```

Execute the two-goal query:

```text
edge(1,Y), edge(Y,Z)
```

A sequential database scan must emit:

```text
Y=2, Z=4
Y=2, Z=5
Y=3, Z=5
```

With a cut after the first accepted complete solution, it must emit only:

```text
Y=2, Z=4
```

This laboratory combines tagged terms, bounded dereference, reversible binding, explicit alternatives, access/execute decoupling, context epochs, and committed output.

## Patterns composed

| Pattern | Role |
|---|---|
| Tagged Value Word | represents atoms, variables, references, and errors |
| Self-Reference Sentinel | encodes an unbound variable as a reference to itself |
| Structure / Representation Firewall | centralizes dereference and term access |
| Type / Shape Dispatch | selects atom equality, binding, or failure |
| Semantic Region Partition | separates facts, bindings, trail, choices, contexts, and results |
| Choice-Point Snapshot | stores scan continuation and trail marks |
| Mutation Trail | reverses bindings on backtracking |
| Commit / Cut Fence | discards older alternatives after accepted cut result |
| Access / Execute Decoupling | overlaps fact scanning with unification |
| Elastic Channel | connects scanner, unifier, and result formatter |
| Context / Epoch-Carrying Token | isolates queries and cancellation generations |
| Delayed Irreversible Store | keeps complete tuples stable under host backpressure |

![The capstone composes read-only fact scanning, bounded dereference, reversible binding, explicit search alternatives, and result commitment.](assets/relational_engine.png){ width=92% }

## Scope

The baseline is intentionally smaller than a complete Prolog system. It supports:

- one binary relation identifier, `edge`;
- compact atom arguments;
- a small binding array for logic variables;
- two fixed query goals;
- depth-first, left-to-right execution;
- chronological backtracking;
- optional cut after the first complete solution.

This is sufficient to exercise the WAM-like mechanisms without first building a parser, general term heap, clause compiler, or full bytecode engine.

## Fact record

A binary fact fits one 40-bit word:

```text
fact40
[39:36] relation ID
[35:18] argument 0
[17:0]  argument 1
```

For the directed test, atoms are unsigned compact integers. Reserve relation ID `1` for `edge`.

The fact table is read-only during a query and can be initialized at configuration or through a host loader. A descriptor supplies:

```text
base address
fact count
relation ID
```

A later design can index by first argument; the baseline performs a sequential scan so search order is unambiguous.

## Term representation

Use `value40` terms:

```text
ATOM(k)   tag=ATOM, payload=k
REF(v)    tag=REF,  payload=variable index
ERROR(e)  tag=ERROR
```

The query templates are:

```text
goal 0: edge(ATOM(1), REF(Y))
goal 1: edge(REF(Y), REF(Z))
```

The fact scanner converts compact fact arguments into `ATOM` values before sending a candidate to the unifier.

## Binding store and self-reference

For each query context, allocate binding slots. Variable `v` is unbound when:

```text
binding[v] = REF(v)
```

A bound variable may contain `ATOM(k)` or a reference to another variable. Initialize Y and Z to their own references.

A context record includes a binding-base address so several contexts can share one physical binding RAM.

## Bounded dereference

The dereference engine is:

```text
deref(term, context, epoch):
    steps = 0

    while term.tag == REF:
        v = term.payload

        if v is outside context variable range:
            return ERROR(BAD_VARIABLE)

        target = binding[context.base + v]

        if target == REF(v):
            return REF(v)          // legal unbound sentinel

        term = target
        steps++

        if steps == MAX_VARIABLES:
            return ERROR(REFERENCE_CYCLE)

    return term
```

The same-reference test must compare the logical variable index for the current context, not the physical RAM address unless those encodings are intentionally identical.

A later optimization can perform path compression. Path compression is a mutation and must be trailed or restricted to bindings whose rollback behavior remains correct. Do not add it in the baseline.

## Unification transition

For compact atom and variable terms:

```text
unify(X, Y):
    Xd = deref(X)
    Yd = deref(Y)

    if Xd is ERROR or Yd is ERROR:
        return error

    if Xd.tag == ATOM and Yd.tag == ATOM:
        return success if Xd.payload == Yd.payload else failure

    if Xd is an unbound REF:
        bind(Xd.variable, Yd)
        return success

    if Yd is an unbound REF:
        bind(Yd.variable, Xd)
        return success

    return failure
```

Variable-variable binding uses a canonical direction, for example bind the higher variable number to the lower. If both dereference to the same unbound variable, success requires no write.

## Reversible binding

All binding writes use the trail substrate:

```text
bind(v, value):
    old = binding[v]

    if old is not the expected unbound REF(v):
        return INTERNAL_BIND_RACE

    if trail has no capacity:
        return TRAIL_FULL without modifying binding

    trail[trail_top] = {context, v, old}
    trail_top++
    binding[v] = value
```

In the baseline, every binding is trailed. A more WAM-like age test can omit trailing for variables younger than the active choice point, but it introduces allocation-age state and should be measured rather than assumed useful.

## Query control plan

The fixed plan has stages:

```text
G0_SCAN       scan facts for edge(1,Y)
G0_UNIFY      unify selected fact with goal 0
G1_SCAN       scan facts for edge(Y,Z)
G1_UNIFY      unify selected fact with goal 1
SOLUTION      format and commit Y,Z
BACKTRACK_G1  resume next goal-1 candidate
BACKTRACK_G0  resume next goal-0 candidate
COMPLETE
FAULT
```

Each scanner request carries:

```text
context
epoch
relation descriptor
start fact index
query stage
continuation tag
```

A scanner response carries the same identity plus candidate arguments and the next fact index.

## First-goal choice point

Before accepting a fact that matches goal 0, preserve the next candidate position:

```text
choice point G0:
    resume stage       G0_SCAN
    next fact index    candidate_index + 1
    trail mark         before bindings for this candidate
    query continuation proceed to G1_SCAN on success
    context and epoch
```

The choice point represents the remaining first-goal alternatives. It is published before binding `Y` for the chosen fact.

Alternatively, the scanner can produce one candidate at a time and the controller can recreate the next index without storing a full choice point. The explicit record is retained because it generalizes to multiple clauses and dynamic plans.

## Second-goal alternatives

At entry to goal 1, save a trail mark after `Y` is bound but before `Z` is bound. Each matching second-goal fact binds `Z`, produces a solution, then unwinds to the goal-1 mark and resumes the next fact.

A choice record can contain:

```text
resume stage       G1_SCAN
next fact index
trail mark         preserve Y, undo Z
outer choice link  G0 choice
context and epoch
```

When goal-1 facts are exhausted, pop its alternative state, unwind to the G0 mark, and resume the next first-goal candidate.

## Directed search trace

The baseline scan order is the listed fact order.

### First candidate

```text
fact 0 = edge(1,2)
match first argument 1
bind Y = 2
begin goal 1 scan
```

Goal 1 matches:

```text
fact 2 = edge(2,4)  -> bind Z=4 -> solution (2,4)
fact 3 = edge(2,5)  -> bind Z=5 -> solution (2,5)
```

After each accepted tuple, unwind only `Z` and resume the next second-goal fact. After goal 1 exhausts, unwind `Y` and resume goal 0.

### Second first-goal candidate

```text
fact 1 = edge(1,3)
bind Y = 3
begin goal 1 scan
fact 4 = edge(3,5)  -> bind Z=5 -> solution (3,5)
```

No other first-goal fact starts with 1. Search completes after unwinding all active bindings and alternatives.

## Result commitment

At `SOLUTION`, dereference Y and Z and construct:

```text
result tuple:
    context
    epoch
    Y value
    Z value
    status = SOLUTION
```

An 80-bit result can hold two compact 18-bit atoms plus context and status. If using two `value40` values, use a 120- or 160-bit logical record split across multiple physical words with valid-last publication.

The safest first transition is:

```text
SOLUTION:
    copy dereferenced Y and Z into pending result registers
    state = SOLUTION_WAIT

SOLUTION_WAIT:
    result_valid = 1
    hold tuple stable

    if result_ready:
        solution_count++
        if cut_after_first:
            state = CUT_COMMIT
        else:
            state = BACKTRACK_G1
```

Bindings are not unwound until the tuple is accepted. A deeper result FIFO permits immediate backtracking after FIFO acceptance.

## Cut

The cut follows the first accepted solution:

```text
CUT_COMMIT:
    discard choice points created since query entry
    reclaim associated trail history after preserving required final state
    invalidate scanner requests for discarded alternatives if any
    mark query COMPLETE
```

For a query whose only externally visible result is already buffered, the implementation can unwind all bindings and clear query-local state after cut. The output FIFO retains the tuple independently.

Ordering is:

```text
accept solution into commit FIFO
then discard alternatives
then reclaim or unwind local search state
then report complete
```

Discarding first risks losing the only result under output backpressure.

## Access/execute decoupling

The scanner and unifier can overlap:

```text
fact RAM -> scanner candidate FIFO -> unifier -> control response FIFO
```

The scanner may run ahead by a bounded number of candidates. Every candidate carries context, epoch, stage, and fact index. A cut or backtrack changes the live epoch or scanner-generation for that stage. Old candidates drain as stale.

Runahead must not alter database-order semantics. Candidates remain ordered in the FIFO, and only one active scan stream per stage/context exists in the baseline.

A credit count equal to candidate FIFO capacity bounds runahead. Cancelling a candidate returns its slot credit when discarded.

## Multi-context extension

After the sequential design passes, support four contexts. A context record holds:

```text
live epoch
query stage
fact index
binding base
trail base and top
choice base and top
pending output state
first-error state
```

Shared engines are:

```text
fact scanner
bounded dereferencer
unifier / binding mutation service
unwind engine
result formatter
```

Requests are ready/valid tokens. A round-robin scheduler chooses among eligible contexts. Long operations return a continuation tag. No global mode exists; stage and mode travel with the context or token.

A simple memory partition allocates fixed binding, trail, and choice slices per context. This avoids allocator complexity and gives hard capacity isolation. A later pooled design can improve utilization but requires per-context ownership and reclaim accounting.

## Query cancellation

Cancellation performs:

```text
increment live epoch
mark context CANCELLING
unwind or discard query-local bindings and choices
wait only for local cleanup, not for stale scanner responses
emit CANCELLED status if required
return context to IDLE
```

Every delayed scanner, dereference, unification, error, and credit event carries epoch. An old fault must not terminate the new query.

If fixed per-context memory slices are used, stale events cannot corrupt another context but can still corrupt the reused generation of the same context; epoch checks remain necessary.

## Fault model

Context-level errors:

```text
BAD_RELATION
BAD_VARIABLE
REFERENCE_CYCLE
MALFORMED_TERM
TRAIL_FULL
CHOICE_FULL
RESULT_TOO_WIDE
CANCELLED
```

Machine-integrity faults:

```text
TRAIL_UNDERFLOW
CHOICE_LINK_CORRUPTION
RAM_RESPONSE_MISMATCH
ILLEGAL_CONTROLLER_STATE
```

A context error produces one committed error result and cleans up the query. A machine-integrity fault stops the experiment and preserves trace state.

## Verification plan

Directed queries:

```text
edge(1,Y), edge(Y,Z)       three expected solutions
edge(5,Y)                  one solution Y=6
edge(6,Y)                  no solutions
edge(X,X)                  no solution in the sample relation
edge(1,Y) with cut         first matching Y only
```

Unification cases:

```text
ATOM == same ATOM
ATOM != different ATOM
unbound variable with ATOM
unbound variable with unbound variable
same variable with itself
one-hop and long reference chains
illegal two-variable cycle
```

Pressure and timing:

```text
stall result for many cycles at every solution
make trail capacity one entry too small
make choice capacity one entry too small
cancel during fact RAM wait
cancel during dereference
cancel after binding but before response
reuse context before old scanner response drains
```

Assertions:

```text
each binding is self-unbound or dereferences within the bound
trail unwind restores the saved binding hash
no stale epoch writes a binding or emits a tuple
one accepted complete solution produces one output tuple
cut prevents restore or scan to an older alternative
result tuple remains stable while blocked
scanner order matches the declared database-order policy
```

## Reference models

Use three software models:

1. a direct relational join over the fact list to compute the expected tuple multiset;
2. a depth-first unification model to compute exact ordered results;
3. a transaction model with candidate queues, finite trail and choice slices, epochs, and randomized engine latency.

The direct join prevents the transaction model from validating its own search bug. The depth-first model validates order. The transaction model validates stalls and cancellation.

## Instrumentation

```text
fact records read
candidate facts produced and rejected
unifications attempted, succeeded, and failed
dereference steps and maximum depth
bindings written
trail pushes and pops
choice pushes and pops
solutions committed
result stall cycles
stale events dropped by engine
scheduler grants by context
faults by code
```

A useful trace entry for every binding mutation is:

```text
context, epoch, variable, old term, new term, trail index, choice depth
```

For FPGA capture, record a compact digest plus selected detailed entries around a trigger such as first solution, cut, or error.

## GateMate mapping

A four-context target can allocate:

| Structure | Initial physical budget |
|---|---:|
| fact table | 1 RAM |
| binding store | 1 RAM or registers |
| trail slices | 2 RAMs |
| choice slices | 1 RAM |
| candidate and response queues | 2 RAMs |
| result FIFO | 1 RAM |
| context descriptors | registers or 1 RAM |
| trace | remaining debug headroom |

Stop and simplify around 16 physical RAM blocks and 12,000 CPEs. Common excess-area causes are wide general terms, fully associative context scheduling, duplicated dereference engines, and arbitrary trail deletion. Preserve explicit indexing and LIFO rollback first.

## Experimental sequence

1. Hard-wire the fact list and execute one goal with no variables.
2. Add one variable and self-reference unbound representation.
3. Add reversible binding and one scan choice point.
4. Add the two-goal query and exact ordered results.
5. Add result backpressure and commit FIFO.
6. Add cut after first accepted solution.
7. Decouple scanner and unifier with one candidate FIFO.
8. Add context and epoch fields even with one context.
9. Add four contexts and round-robin scheduling.
10. Cancel and reuse contexts under delayed responses.
11. Capture the same query commit trace in simulation and hardware.

## Extensions

- Generalize from compact atoms to heap terms and functors.
- Add structure unification with a term work stack.
- Add indexed fact lookup by relation and first argument.
- Add WAM-like read and construct modes.
- Add parallel second-goal searches with ordered result merge.
- Add a dense query bytecode and regular internal operations.
- Add a host-loaded relation database and capability-protected regions.

## Exercises

1. Prove that chronological trail unwind restores Y while undoing only Z between goal-1 alternatives.
2. Design a compact 40-bit choice record for the fixed two-goal plan.
3. Add `edge(X,X)` and explain why variable identity changes unification behavior.
4. Permit parallel second-goal scans while preserving the listed output order.
5. Add path compression to dereference and state exactly which writes require trailing.
6. Generalize result commitment to tuples larger than one physical RAM word using valid-last publication.

# Part V - Advanced Compositions

# An out-of-order symbolic processor

## Objective

This chapter combines dynamic scheduling with language-level reversible state. The important result is not a complete implementation but a correct division of commitment and rollback ownership.

![An out-of-order symbolic core needs distinct boundaries for instruction retirement, semantic commitment, and external release.](assets/out_of_order_symbolic.png){ width=92% }

## Why simple union fails

It is tempting to place WAM-like operations inside an ordinary out-of-order core by adding tagged registers and a trail. That is insufficient. A binding may be:

1. calculated speculatively;
2. retired into architectural state;
3. retained as semantically reversible while older search alternatives exist; and
4. eventually serialized as part of a host-visible solution.

The processor therefore needs the commitment path:

```text
execute
  -> retire instruction
  -> commit search alternative
  -> release external effect
```

A reorder buffer, mutation trail, and result/store queue participate at different levels. Their state and reclamation must be coordinated.

## Internal operation record

A regular internal operation can contain:

```text
operation class
source producer tags or ready values
destination producer tag
architectural destination
context and epoch
semantic checkpoint ID
required trail / heap / effect credits
exception state
external instruction identity
```

Dense bytecode expands into these records. Reservation stations wait on producer tags. Functional units compute tags and values. The precise commit queue retires in order.

## Binding alternatives

Consider `BIND V, X`. Three implementation arrangements are possible.

### Arrangement A: delay the physical binding write until retirement

Execution dereferences and validates operands, then stores the proposed binding in the ROB. At retirement:

1. reserve or activate a semantic trail entry;
2. write the architectural binding;
3. mark the operation retired.

This creates a clean handoff from \(\mu\)-level rollback to semantic rollback. Dependent speculative operations need forwarding from the ROB or producer-tag result.

### Arrangement B: write a speculative binding file

Execution updates a shadow binding file. Retirement copies or merges the value into the architectural reversible binding store and appends the trail entry. Squash discards or reconstructs the shadow state.

This resembles the Future / Shadow File pattern. It uses more storage but preserves a clean committed binding image.

### Arrangement C: early write plus two-level versioned history

Execution writes the working binding store and creates a microarchitectural history record. At retirement, a semantic trail record becomes authoritative. The records carry versions so a branch squash can remove the unretired update without disturbing older semantic history.

This can give fast reads but has the most complicated ownership proof. It should not be the first FPGA experiment.

## Ownership invariant

For each binding version:

```text
if unretired:
    exactly one μ-level recovery owner exists

if retired and branch-reversible:
    exactly one semantic trail owner exists

if semantically committed:
    no local rollback owner exists
```

Ownership transition is atomic with retirement. A trail entry that exists physically but is not yet authoritative carries a valid or level field.

## Search branch and CPU branch

A CPU branch checkpoint and a logic choice point are distinct:

| Property | CPU branch checkpoint | Logic choice point |
|---|---|---|
| reason | predicted control | semantic alternative |
| lifetime | usually short | potentially long |
| commit | branch resolved / retired | cut, trust, exhaustion, or accepted policy |
| rollback state | rename and issue state | arguments, continuation, heap/trail marks |
| visible level | \(\mu\) | often \(A\) to \(S\) |

A logic choice can contain CPU branches. A CPU branch can occur while constructing a choice point. Keep separate identifiers and declare nesting relationships.

## Fault example

Suppose a `DEREF_AND_BIND` assist follows three references, reserves a trail entry, and discovers a capability fault on the target. Correct behavior is:

- no architectural binding retires;
- any provisional resolver state is discarded;
- the ROB entry records the fault;
- younger internal operations are squashed;
- older architectural and semantic state remains intact;
- the semantic trail does not acquire ownership for the failed binding.

The assist may leave read-only cache state unless the security model requires Speculation Shadow Structures.

## Semantic cut

A `CUT` operation retires architecturally only after all older operations are precise. Its semantic action advances the choice/trail frontier. Younger operations created under discarded alternatives carry an invalidated semantic epoch.

If a result justifying the cut must be emitted, the result first enters a commit queue. The cut then discards alternatives. External delivery can occur later under backpressure.

## FPGA research path

Do not begin with a wide superscalar core. A staged research implementation is:

1. use the Laboratory 1 bytecode front end;
2. expand into regular internal operations;
3. add producer tags and an eight-entry reservation window;
4. add a small precise commit queue;
5. support only arithmetic and dereference initially;
6. add delayed-at-retirement binding arrangement A;
7. add semantic choice points and trail;
8. add cut and result commitment;
9. compare against the in-order relational engine.

The primary measurements are not peak instructions per cycle. Measure semantic transitions per cycle, trail and ROB pressure, wakeup fanout, result latency, and area per active context.

## Verification

Required cross-level properties:

```text
architectural state at each retirement is precise
semantic unwind never crosses an unretired operation
CPU squash never restores through a semantically committed cut
one binding version has one active rollback owner
external result contains only retired and semantically valid values
stale CPU and semantic epochs cannot update state
```

Differential modes should include:

```text
in-order execution
out-of-order with one station
out-of-order with full stations
fused assists disabled
dynamic scheduling disabled
```

All modes must produce equivalent semantic commit traces.

## Chapter summary

An out-of-order symbolic core is a commitment hierarchy, not a conventional core plus a trail. Microarchitectural and semantic rollback mechanisms must hand off ownership explicitly. The cleanest first implementation delays reversible architectural writes until retirement.

## Exercises

1. Design a ROB entry for `BIND` under arrangement A.
2. Draw a timeline containing one CPU branch misprediction inside one logic choice point.
3. State how a replayable heap read differs from a replayable binding write.
4. Define a semantic-epoch mechanism that invalidates operations under a cut without confusing CPU branch epochs.

# A managed and memory-safe symbolic core

## Objective

Combine tags, explicit authority, cheap allocation, collection barriers, and precise faults into one managed-language substrate.

![A managed core places safety and collection checks on the common object path, with precise assists for cold cases.](assets/managed_core.png){ width=92% }

## Reference anatomy

A managed reference may need:

```text
hot tag             identifies reference versus immediate
object or capability ID
compact permission summary
region or generation bit
validity or provenance generation
cold descriptor     base, limit, full permissions, shape, collector metadata
```

The hot fields travel with the value. The cold descriptor resides in protected local memory. The hot summary may reject an access early but must never authorize an operation forbidden by the cold descriptor.

## Access sequence

A load of an object field can be:

```text
1. classify value as reference
2. validate compact generation and permission summary
3. read descriptor if required
4. check bounds and operation permission
5. resolve forwarding state
6. perform heap read
7. apply collector read barrier if required
8. return tagged field value
```

The fast path combines steps when metadata is cached and the object is stationary. The slow path receives the untouched original reference and field index.

## Allocation and publication

A nursery allocation follows:

```text
reserve object words
initialize all fields with valid tagged values
perform any required remembered-set or color action
publish the reference
```

Publishing means storing the reference into a live root, object field, message, or output. A capability is derived with no greater authority than the allocating context.

An exception during initialization resets the allocation top or records enough state to discard the partial object. Uninitialized words are unreachable.

## Barrier composition with speculation

A speculative old-to-young store can touch four mechanisms:

- capability gate;
- precise instruction commit;
- semantic rollback, if inside a transaction or search;
- remembered-set barrier.

Two safe arrangements are:

1. buffer the pointer store and barrier record until architectural or semantic commit;
2. update shadow object and remembered-set structures, merging both together at commitment.

Logging only the pointer write while eagerly updating a shared remembered set may preserve correctness but leak speculative behavior or consume unbounded stale entries. The selected visibility model decides whether that is acceptable.

## Moving objects

A copying collector installs a forwarding cell only after the destination object is complete. Readers use the structure firewall:

```text
old reference
  -> capability/generation check
  -> forwarding resolution
  -> destination object
```

Authority must be preserved or monotonically restricted across movement. A forwarding target cannot be substituted by unprivileged software.

## Managed exceptions

Different failures should remain distinct:

```text
TYPE_FAULT
NULL_OR_INVALID_REFERENCE
BOUNDS_FAULT
PERMISSION_FAULT
STALE_REFERENCE
HEAP_FULL
BARRIER_FULL
COLLECTOR_RETRY
INDIRECTION_CYCLE
```

Each fault states whether the operation can retry after collection, whether the original operands remain live, and which effects have occurred. `HEAP_FULL` often invokes a collector and retries from the original allocation request. `PERMISSION_FAULT` normally does not retry.

## Small GateMate experiment

Extend Laboratory 1 with:

- `ALLOC_PAIR`;
- `LOAD_FIELD 0/1`;
- `STORE_FIELD 0/1`;
- descriptor RAM containing base, limit, and read/write permissions;
- a young-region bit;
- one remembered bit per old object or card;
- a precise `COLLECT_REQUIRED` slow path rather than a full collector initially.

Directed program:

```text
p = alloc_pair(INT(7), INT(9))
load p.field0             -> INT(7)
store p.field1 = INT(11)
load p.field1             -> INT(11)
```

Negative tests attempt out-of-bounds field 2, store through a read-only descriptor, use a stale generation, and fill the nursery.

## Security review

Audit all paths that can access object storage:

```text
normal scalar load/store
fused assists
reducer thunk update
collector move and forwarding
DMA or host loader
debug and ILA-visible control
reset and initialization
atomic operations
vector or bulk copy
```

A privileged maintenance path is acceptable only when its authority and use are explicit. It must not be an accidental bypass.

## Measurements

```text
fast versus descriptor-assisted accesses
capability faults by cause
allocation words and objects
barrier records
old-to-young edges
forwarding hops
collector or retry stalls
metadata RAM bandwidth
tag and permission decode critical paths
```

Ablate the compact permission summary and compare descriptor traffic. Widen payload versus metadata and compare total RAM use.

## Chapter summary

Managed execution composes representation, authority, lifetime, collection, and precise fault patterns. The central rule is that publication occurs only after initialization, authorization, and required barrier state are complete. Every maintenance path participates in the same safety contract.

## Exercises

1. Design a compact descriptor for a pair object and state its capability derivation rule.
2. Place a remembered-set barrier relative to a buffered speculative store.
3. Define retry semantics for `HEAP_FULL` without duplicating allocation.
4. Explain how stale generations interact with forwarding cells.

# A spatial symbolic accelerator

## Objective

Use systolic locality for regular symbolic kernels while keeping irregular control and representation changes in elastic boundary engines.

## Suitable kernels

Symbolic workloads are irregular at the application level, but their hot inner operations can be regular:

- compare fixed-shape tuples;
- hash or filter tagged records;
- perform parallel field equality;
- scan relation facts;
- apply small rewrite rules;
- traverse a frontier with regular edge records;
- compare constructor tags and selected fields;
- process batches of compact unification candidates.

The accelerator should target the measured regular kernel, not force the whole abstract machine into a rigid array.

## Boundary/control split

A useful composition is:

```text
control engine
  -> access/execute queues
  -> shape and capability firewall
  -> local systolic cells
  -> join/reduction collector
  -> precise result queue
```

The control engine handles variable-length structures, faults, choices, and continuations. Cells handle a fixed transition over locally streamed records.

## Example: fact-filter pipeline

Build four cells:

```text
cell 0  relation-ID compare
cell 1  first-argument tag/value compare
cell 2  second-argument quick classification
cell 3  pack matching candidate and destination
```

The query template remains stationary in the cells. Fact records stream through. Each cell adds a match bit or error state and forwards the complete context/epoch record.

This reduces repeated query-template reads and keeps comparisons local. A candidate with an unbound variable leaves the detailed binding operation to the unifier at the boundary.

## Elastic systolic timing

Pure systolic arrays often assume fixed movement each cycle. Symbolic exceptions and shared outputs require elasticity. Place a one-entry channel between cells:

```text
cell_i output accepted only when cell_{i+1} can accept
```

For longer arrays, credits or small local FIFOs break the reverse ready path. Bubbles preserve order. Context and epoch propagate with every record.

## Stationary-data choice

Select one stationary data class:

- query pattern stationary, facts stream;
- descriptor and permissions stationary, references stream;
- rewrite rule stationary, nodes stream;
- graph frontier stationary, edges stream.

The choice determines reuse and memory traffic. Measure bytes moved per accepted semantic transition rather than only operations per cycle.

## Exceptional bypass

Every cell has a simple common predicate. If a record requires a deep descriptor, long dereference, boxed value, or error handling, it sets an assist flag and bypasses arithmetic modification. At the array exit, an assist queue sends the original record plus progress to a general engine.

The bypass must preserve order if the external contract requires it. One solution attaches sequence numbers and uses a merge queue. Another allows out-of-order candidate completion but retains per-context result ordering downstream.

## Fork/join unification

For a fixed two-field constructor, the control engine can fork two field comparisons into parallel lanes. A join record contains:

```text
context, epoch, parent operation
required mask = 2'b11
received mask
field results
fault policy
```

Only if both comparisons succeed does the parent issue a binding or success result. Duplicate child completions are rejected by label.

## GateMate experiment

Use Laboratory 5 as a baseline, then replace sequential fact prefiltering with a four-cell pipeline. Keep the general unifier unchanged.

Measure:

```text
facts accepted per cycle
candidate reduction ratio
cell stall cycles
assist rate
candidate FIFO occupancy
CPE and RAM increase
post-route frequency
energy proxy: toggles or routed fanout if available
end-to-end query cycles
```

A faster fact stream is useless if the unifier or result queue is already saturated. The pattern succeeds when local filtering reduces total data movement or raises useful semantic throughput.

## Scaling limits

Signs that the array is no longer appropriate include:

- assist rate approaches common-path rate;
- query templates change too frequently to amortize loading;
- global control or broadcast signals reach every cell;
- local buffers fill behind one shared output;
- cells need general variable-length traversal;
- routing, not cell logic, determines timing.

At that point, cluster smaller arrays behind a scheduler or return work to a general engine.

## Chapter summary

Systolic design is a locality pattern, not merely a parallelism pattern. Keep a regular transition and stationary data in local cells; keep dynamic graph control, faults, and commitment at elastic boundaries. Measure end-to-end semantic throughput and bytes moved.

## Exercises

1. Design a two-lane field-comparison fork/join unit.
2. Preserve database-order results when common cases use the array and uncommon cases use a variable-latency assist.
3. Choose stationary data for graph rewrite matching and justify it from expected reuse.
4. Draw a wait-for graph for an elastic four-cell array with one shared assist queue.

# Measurement-driven architecture research

## Objective

Turn the pattern language into a repeatable method for proposing, implementing, and evaluating new symbolic-computing hardware.

![Pattern selection begins with semantic transition measurements and ends with another measurement, not with architectural fashion.](assets/pattern_selection.png){ width=92% }

## Profile transitions, not source syntax

Count semantic events:

```text
tag classes observed
object shapes and sizes
dereference chain lengths
choice points and alternatives
trail writes per branch
allocation bytes and survival
ready-queue occupancy
result fanout
continuation depth
slow-path frequency
bytes moved per transition
commit and rollback distance
```

A source construct such as pattern matching may expand into very different costs depending on representation and data. Optimize the dominant transition or movement.

## State classification worksheet

For every semantic state component, record:

| Question | Purpose |
|---|---|
| lifetime | choose register, stack, region, or collected heap |
| read/write frequency | choose ports and caches |
| locality and consumers | choose banking and placement |
| recomputability | choose checkpoint versus reconstruction |
| commitment level | choose rollback and output gates |
| maximum occupancy | choose capacity and overflow policy |
| visibility | choose architectural interface and security treatment |
| mutation owner | avoid conflicting writers |

This worksheet usually reveals the initial pattern candidates.

## Hypothesis format

A useful architectural hypothesis is specific:

> For relation workloads with average candidate rejection above 80 percent, a stationary-query fact-filter pipeline will reduce bytes delivered to the unifier by at least 4x and improve committed solutions per cycle without exceeding 20 percent additional CPEs.

The hypothesis names workload condition, mechanism, predicted causal metric, end metric, and resource bound. "A systolic engine will be faster" is not testable enough.

## Baselines and ablations

Every pattern experiment needs a baseline that preserves semantics:

- full snapshots versus trail;
- coupled scan/execute versus queues;
- global broadcast versus explicit destinations;
- repeated thunk evaluation versus update;
- in-order versus dynamic scheduling;
- general representation versus shape specialization;
- slow path only versus checked fast path;
- fused versus decomposed operations.

An ablation disables one mechanism while holding other parameters fixed. Record tool versions and seeds so area or frequency differences are attributable.

## Metrics

Report at least:

```text
correct semantic outcomes
committed transitions per cycle
end-to-end latency distribution
CPE, flip-flop, and RAM use
post-route critical path and frequency
queue, trail, heap, and station occupancy
bytes or words moved by storage class
rollback work
assist and fault rates
compile and place-route reproducibility across seeds
```

Peak unit utilization without committed useful work can be misleading. Stale and replayed work should be counted separately.

## Workload sets

Use three levels:

1. **microtransition tests:** isolate tag decode, bind, force, wakeup, or trail push;
2. **structured kernels:** queens, expression DAG, shared thunk, relational join;
3. **mixed traces:** interleave contexts, errors, cancellation, and capacity pressure.

A pattern justified only by one hand-selected input is not yet general.

## Physical evidence

Inspect:

- inferred RAM configurations;
- replicated logic and high-fanout nets;
- critical ready or tag-decode paths;
- placement of engines relative to banks;
- route congestion;
- effect of inserting elastic stages;
- timing variation across seeds.

A logical speedup may disappear if the composed design creates a global route. Physical closure is part of the research claim.

## Verification evidence

Archive:

```text
semantic model and version
random seeds
commit traces
assertion results
negative tests
capacity-boundary tests
hardware trace comparison
known unsupported cases
```

A performance result from an implementation with untested rollback or output backpressure is not an architectural result.

## Pattern evolution

A pattern can be refined into variants:

```text
centralized versus distributed reservation stations
broadcast versus explicit wakeup
full-value versus inverse-operation trail
register versus BRAM choice stack
fixed-slice versus pooled per-context memory
memoized-error versus retryable thunk
registered ready versus credits
```

Variants share the invariant but differ in forces and physical assumptions. Record the condition under which one dominates.

## A research roadmap

A disciplined sequence for the GateMate work is:

```text
Phase 1  reproduce all five laboratories
Phase 2  extract reusable RTL and transaction models
Phase 3  measure bottlenecks and create baselines
Phase 4  add one advanced pattern at a time
Phase 5  compare across workload families
Phase 6  combine only mechanisms whose local benefit survives composition
Phase 7  publish source, manifests, traces, and negative tests
```

The pattern language is useful only if it supports removal as well as addition. A mechanism that does not improve committed semantic work under its stated context should be deleted or narrowed.

## Chapter summary

Architecture research begins with an explicit machine, profiles semantic transitions, selects the smallest applicable patterns, checks contract closure, and measures again. Correctness, finite pressure, and physical implementation are parts of the result rather than prerequisites hidden from it.

## Exercises

1. Write a complete hypothesis for replacing explicit destinations with clustered tag broadcast.
2. Define an ablation plan for the queens trail.
3. Choose metrics that distinguish useful work from stale or replayed dataflow operations.
4. Create a result table template that includes semantic, resource, timing, and verification evidence.

![The long-term target is a symbolic processing tile assembled from replaceable, verified protocols.](assets/capstone.png){ width=92% }

# Appendices

# Pattern quick reference

The table is an index, not a substitute for the full contract. The question column is useful during design reviews.

| No. | Pattern | Primary design question |
|---:|---|---|

| 1 | Abstract-Machine Contract | What exact abstract transition is being implemented? |

| 2 | Explicit Semantic State Vector | Which semantic state is explicit and checkpointable? |

| 3 | Structure / Representation Firewall | Can representation change without changing consumers? |

| 4 | Common Fast Path / Precise Slow Path | Where is the precise handoff to the uncommon case? |

| 5 | Checked Compiler Hint | Can a wrong hint change only performance? |

| 6 | Tagged Value Word | Does every value carry the metadata needed to interpret it? |

| 7 | Immediate-or-Boxed Split | Which common values avoid allocation, and how does overflow escape? |

| 8 | Capability / Descriptor Gate | Where is authority checked on every access path? |

| 9 | Self-Reference Sentinel | Is the structural sentinel unique and cycle-safe? |

| 10 | Shape-Specialized Encoding | Which dominant object shape deserves a direct encoding? |

| 11 | Indirection / Forwarding Cell | How does logical identity survive movement or update? |

| 12 | Semantic Region Partition | Are states with different lifetime and rollback rules separated? |

| 13 | Atomic Metadata Propagation | Can payload and metadata ever become misaligned? |

| 14 | Explicit Continuation Record | Can suspended control resume from recorded state alone? |

| 15 | Split-Lifetime Frame | Which values must survive the next boundary? |

| 16 | Liveness-Driven Frame Trimming | When can a frame slot cease to exist or be a root? |

| 17 | Tail-State Reuse | Is there any legal path back to the caller being reused? |

| 18 | Lifetime Promotion on Escape | What happens when a local reference escapes? |

| 19 | Choice-Point Snapshot | What minimum state restarts the next alternative? |

| 20 | Mutation Trail | Which old values must be logged before mutation? |

| 21 | Commit / Cut Fence | What event makes older alternatives unreachable? |

| 22 | Demand Cell with Update and Blackhole | Who owns a deferred computation, and how are cycles handled? |

| 23 | Ready-Operand Firing | What exact event makes an operation ready once? |

| 24 | Producer Tag as Future | How is an unavailable value named without ambiguity? |

| 25 | Reservation Station | Where does waiting work live while units stay available? |

| 26 | Tagged Result Wakeup | How does a result reach only live consumers? |

| 27 | Access / Execute Decoupling | Which memory work can run ahead safely? |

| 28 | Elastic Channel | Can latency change without losing or duplicating items? |

| 29 | Credit-Based Backpressure | Are downstream slots accounted for without a ready path? |

| 30 | Context / Epoch-Carrying Token | Can a stale response modify a reused context? |

| 31 | Fork / Join Collector | What child-completion predicate fires the parent? |

| 32 | Local Systolic Cell | Which data remains local and stationary? |

| 33 | Precise Commit Queue | What preserves a precise architectural prefix? |

| 34 | History / Undo Buffer | Which old values permit early working-state update? |

| 35 | Future / Shadow File | Would separate provisional and committed images simplify recovery? |

| 36 | Delayed Irreversible Store | Which effects must wait before becoming irreversible? |

| 37 | Checkpoint plus Epoch Squash | Can compact restore plus epoch invalidation replace a search? |

| 38 | Replayable Operation | Is re-execution idempotent below the visibility boundary? |

| 39 | Speculation Shadow Structures | Which speculative side effects require private shadow state? |

| 40 | Bump-Allocated Generation | Can allocation be reserve, initialize, then publish? |

| 41 | Nursery plus Remembered Set | How are all old-to-young edges made discoverable? |

| 42 | Incremental Collection Barrier | Which barrier preserves the chosen collector invariant? |

| 43 | Nonblocking Miss Record | How are variable-latency misses identified and merged? |

| 44 | Bank by Access Function | Which access functions deserve separate banks and ports? |

| 45 | Semantic Transition Family | Is this transition stable enough to expose directly? |

| 46 | Type / Shape Dispatch | Can tag and shape testing select the action safely? |

| 47 | Mode-Dependent Transition | Is mode explicit per context and checkpoint? |

| 48 | Run-Length Semantic Operation | Where is the precise progress index inside a counted operation? |

| 49 | Dense Veneer / Regular Internal Core | Can compact code expand into regular internal operations? |

| 50 | Fused Semantic Assist with Decomposition Escape | Can fusion always escape to an equivalent decomposition? |

# Suggested project organization

A repository should keep semantic models, transaction models, reusable RTL, experiments, and result manifests separate.

```text
symbolic-hw/
  README.md
  LICENSE
  Makefile

  docs/
    pattern-contracts/
    experiment-notes/

  models/
    common/
      values.py
      events.py
      commit_trace.py
    stack_model.py
    queens_model.py
    dataflow_model.py
    reducer_model.py
    relations_model.py

  rtl/
    common/
      symbolic_types_pkg.sv
      rv_reg.sv
      rv_fifo.sv
      credit_counter.sv
      sync_sdp_ram.sv
      sync_tdp_ram.sv
      epoch_guard.sv
      fault_latch.sv
      trail_stack.sv
      checkpoint_stack.sv
      commit_fifo.sv
      priority_bit.sv
      onehot_check.sv
      trace_sink.sv

    exp1_stack/
      stack_core.sv
      tagged_alu.sv
      program_rom.sv
      top.sv

    exp2_queens/
      queens_control.sv
      queens_propagator.sv
      top.sv

    exp3_dataflow/
      token_router.sv
      operand_store.sv
      activation_queue.sv
      node_issue.sv
      result_router.sv
      top.sv

    exp4_reducer/
      graph_heap.sv
      reducer_control.sv
      continuation_stack.sv
      top.sv

    exp5_relations/
      fact_scanner.sv
      dereference.sv
      unifier.sv
      query_context.sv
      top.sv

  tb/
    common/
      rv_monitors.sv
      commit_scoreboard.py
      random_stall.py
    exp1_stack/
    exp2_queens/
    exp3_dataflow/
    exp4_reducer/
    exp5_relations/

  programs/
    stack/
    graphs/
    relations/

  boards/
    <board-name>/
      board.ccf
      board.sdc
      top_io.sv

  scripts/
    build.py
    compare_trace.py
    summarize_reports.py

  results/
    manifests/
    synthesis/
    place_route/
    traces/
```

## Build targets

A useful Makefile interface is:

```text
make sim EXP=stack SEED=1
make formal BLOCK=rv_fifo
make synth EXP=queens BOARD=<board>
make impl EXP=dataflow BOARD=<board> SEED=4
make bit EXP=reducer BOARD=<board>
make program EXP=relations BOARD=<board>
make trace-compare EXP=relations RUN=<id>
make report EXP=all
```

The build system should emit a manifest automatically rather than relying on terminal history.

## Parameter discipline

Expose parameters that represent architectural experiments:

```text
number of contexts
queue depths
trail and choice capacities
value and address widths
functional-unit counts and latency
broadcast or explicit routing mode
cut policy
memoized-error or retry policy
```

Do not parameterize every internal wire. Excessive genericity obscures synthesis and verification. A parameter is justified when the experiment will vary it or when board portability requires it.

# Representative GateMate build flow

The following commands reflect the May 2026 official quick-start structure [COL26-T]. Reconcile them with the locally installed OSS CAD Suite and board documentation.

## Synthesis

```bash
mkdir -p build

yosys -ql build/yosys.log -p '
  read_verilog -defer -sv \
    rtl/common/*.sv \
    rtl/exp1_stack/*.sv;

  synth_gatemate \
    -top top \
    -luttree \
    -nomx8;

  stat;
  write_json build/top.json;
  write_verilog build/top_synth.v;
'
```

For broader SystemVerilog support, use the `slang` frontend supplied by the installed toolchain when needed:

```bash
yosys -m slang -ql build/yosys.log -p '
  read_slang rtl/common/*.sv rtl/exp1_stack/*.sv;
  synth_gatemate -top top -luttree -nomx8;
  stat;
  write_json build/top.json;
'
```

Check the synthesis log for inferred `CC_BRAM_20K`, `CC_BRAM_40K`, arithmetic cells, unexpectedly large register arrays, latches, and warnings.

## Place and route

```bash
nextpnr-himbaechel \
  --device=CCGM1A1 \
  --json build/top.json \
  -o ccf=boards/<board>/board.ccf \
  -o out=build/top.txt \
  --sdc boards/<board>/board.sdc \
  --router router2
```

Optional device-specific operating or timing modes should be selected from the current guide and recorded in the manifest. Review utilization and timing reports after every implementation.

## Bitstream

```bash
gmpack build/top.txt build/top.bit
```

Additional SPI, CRC, background-reconfiguration, or boot options are board and deployment specific.

## Programming

JTAG and SPI examples in the official guide use `openFPGALoader` with a board identifier [COL26-T]. Determine the exact board string and connection from the board documentation:

```bash
openFPGALoader -b <board-id> build/top.bit
```

Do not publish a project-wide programming command until it has been verified against the actual board and cabling.

## Timing constraints

A minimal SDC begins with the real board clock:

```tcl
create_clock -name sys_clk \
  -period <period-ns> \
  [get_ports clk]
```

Add generated clocks, false paths, and multicycle paths only when their semantics are proven. Do not use false paths to hide an unexplained critical path.

## Reproducibility manifest

```yaml
experiment: exp3_dataflow
rtl_commit: <git-sha>
board: <board-name>
device: CCGM1A1
clock_period_ns: <value>
yosys_version: <value>
nextpnr_version: <value>
gmpack_version: <value>
placer_seed: <value>
parameters:
  contexts: 4
  ready_depth: 8
  multiplier_latency: 3
simulation_seed: 918273
model_digest: <sha256>
commit_trace_digest: <sha256>
```

# Reusable assertion templates

These templates express intent. Adapt syntax to the formal or simulation tool in use.

## Ready/valid stability

```systemverilog
property p_stable_when_blocked;
  @(posedge clk) disable iff (rst)
    valid && !ready |=> valid && $stable(data);
endproperty
assert property (p_stable_when_blocked);
```

## No stale epoch update

```systemverilog
property p_live_epoch_on_write;
  @(posedge clk) disable iff (rst)
    state_write_en |->
      write_epoch == live_epoch[write_context];
endproperty
assert property (p_live_epoch_on_write);
```

## Trail capacity before protected write

```systemverilog
property p_trail_reserved_before_write;
  @(posedge clk) disable iff (rst)
    protected_write_en && write_requires_trail |->
      trail_entry_valid && trail_entry_committed;
endproperty
assert property (p_trail_reserved_before_write);
```

## Output commitment

```systemverilog
property p_pop_only_on_output_accept;
  @(posedge clk) disable iff (rst)
    output_pop |-> out_valid && out_ready;
endproperty
assert property (p_pop_only_on_output_accept);
```

## Exactly-once activation

```systemverilog
property p_issue_requires_operands;
  @(posedge clk) disable iff (rst)
    activation_enqueue |->
      all_required_valid && !issued_before;
endproperty
assert property (p_issue_requires_operands);
```

## Owned heap update

```systemverilog
property p_heap_write_authorized;
  @(posedge clk) disable iff (rst)
    heap_write_en |->
      heap_owner_grant &&
      request_epoch == live_epoch[request_context];
endproperty
assert property (p_heap_write_authorized);
```

## Commit contains no poison

```systemverilog
property p_no_poison_commit;
  @(posedge clk) disable iff (rst)
    commit_valid |-> commit_value.tag != TAG_POISON;
endproperty
assert property (p_no_poison_commit);
```

## Pointer bounds

```systemverilog
assert property (@(posedge clk)
  stack_top <= STACK_CAPACITY);

assert property (@(posedge clk)
  trail_top >= trail_base && trail_top <= TRAIL_CAPACITY);
```

## Credit conservation

For formal verification, represent sent and returned credit counts in a widened ghost counter:

```text
initial capacity
  = current credits
  + occupied downstream entries
  + reserved but not yet occupied entries
```

Assert the equality every cycle and separately assert each term stays within range.

## Checkpoint restoration

A testbench or formal harness can snapshot a small complete state:

```systemverilog
assert property (@(posedge clk) disable iff (rst)
  unwind_done |->
    abstract_state == checkpoint_state[restored_checkpoint]);
```

For large memories, use per-location history properties or a collision-resistant verification hash plus targeted direct comparisons.

# Glossary

**Abstract machine.** An implementation-independent state-transition system defining observable behavior.

**Activation.** A particular dynamic instance of an operation or graph node.

**Architectural state.** State visible through the processor or abstract-machine contract after retirement.

**Backpressure.** A lossless flow-control response that prevents a producer from transferring when downstream capacity is unavailable.

**Blackhole.** A temporary thunk state indicating that evaluation has been claimed and recursive demand must be detected or suspended.

**Choice point.** A compact record that resumes an untried semantic alternative.

**Commit.** The event that advances an effect across a declared visibility frontier.

**Commit packet.** A trace record describing one completed abstract transition.

**Continuation.** Explicit data sufficient to resume a suspended computation.

**Credit.** A token representing one unit of available downstream capacity.

**Cut.** A semantic commitment operation that discards older alternatives and their recovery state.

**Dataflow firing.** Activation of an operation when required operands are present.

**Dereference.** Repeatedly following reference bindings or indirections to a terminal value or legal sentinel.

**Epoch.** A generation attached to work so that responses from cancelled or replaced state can be rejected.

**External visibility.** Observability by another processor, domain, device, file, network endpoint, or user.

**Forwarding cell.** An object representation redirecting an old logical location to a current representative.

**Future.** A placeholder naming a value that will be produced later.

**Invariant.** A property required to hold for every reachable state of a pattern implementation.

**Metadata.** Information such as type, validity, authority, context, epoch, fault, ownership, or destination that defines how a payload is interpreted.

**Mutation trail.** A reverse-ordered log of prior values or inverse operations used for rollback.

**Precise fault.** A fault reported at a state corresponding to a defined abstract transition boundary.

**Provisional effect.** An internal effect that has not crossed its required commitment boundary.

**Refinement.** A relation showing that implementation transitions preserve the observations of an abstract machine.

**Replay.** Re-execution of an operation whose prior attempt remained below its visibility frontier.

**Reservation station.** Storage for a waiting operation, operands, producer tags, destination, and scheduling metadata.

**Semantic commitment.** The point after which a language-level alternative, transaction, or deferred state change can no longer reverse an effect.

**Sentinel.** A distinguished representation, often structural, denoting a special state such as unbound or empty.

**Shadow structure.** Private provisional state merged into shared state only at commitment.

**Stuttering.** Implementation steps that do not change the abstract observation.

**Thunk.** A deferred computation plus the environment or state required to evaluate it.

**Visibility frontier.** The boundary at which an effect changes from provisional to observable at a specified level.

**Wakeup.** Resolution of a waiting operand or continuation when a producer result arrives.

# Bibliography

[KOG91] Peter M. Kogge. *The Architecture of Symbolic Computers*. McGraw-Hill Series in Supercomputing and Parallel Processing, 1991. ISBN 0-07-035596-7.

[LAN64] Peter J. Landin. “The Mechanical Evaluation of Expressions.” *The Computer Journal*, 6(4):308-320, 1964. DOI: 10.1093/comjnl/6.4.308.

[WAR83] David H. D. Warren. *An Abstract Prolog Instruction Set*. Technical Note 309, SRI International, 1983.

[TOM67] Robert M. Tomasulo. “An Efficient Algorithm for Exploiting Multiple Arithmetic Units.” *IBM Journal of Research and Development*, 11(1):25-33, 1967. DOI: 10.1147/rd.111.0025.

[SMP88] James E. Smith and Andrew R. Pleszkun. “Implementing Precise Interrupts in Pipelined Processors.” *IEEE Transactions on Computers*, 37(5):562-573, 1988.

[DEN75] Jack B. Dennis and David P. Misunas. “A Preliminary Architecture for a Basic Data-Flow Processor.” In *Proceedings of the 2nd Annual Symposium on Computer Architecture*, 1975.

[SMI82] James E. Smith. “Decoupled Access / Execute Computer Architectures.” In *Proceedings of the 9th Annual Symposium on Computer Architecture*, 1982.

[KRO81] David Kroft. “Lockup-Free Instruction Fetch/Prefetch Cache Organization.” In *Proceedings of the 8th Annual Symposium on Computer Architecture*, 1981.

[KUN82] H. T. Kung. “Why Systolic Architectures?” *Computer*, 15(1):37-46, 1982.

[CAR01] Luca P. Carloni, Kenneth L. McMillan, and Alberto L. Sangiovanni-Vincentelli. “Theory of Latency-Insensitive Design.” *IEEE Transactions on Computer-Aided Design of Integrated Circuits and Systems*, 20(9):1059-1076, 2001.

[JON92] Simon L. Peyton Jones. “Implementing Lazy Functional Languages on Stock Hardware: The Spineless Tagless G-machine.” *Journal of Functional Programming*, 2(2):127-202, 1992.

[LIE83] Henry Lieberman and Carl Hewitt. “A Real-Time Garbage Collector Based on the Lifetimes of Objects.” *Communications of the ACM*, 26(6):419-429, 1983.

[WAT19] Robert N. M. Watson, Peter G. Neumann, Jonathan Woodruff, Michael Roe, Hesham Almatary, Jonathan Anderson, John Baldwin, et al. *An Introduction to CHERI*. University of Cambridge Computer Laboratory Technical Report UCAM-CL-TR-941, 2019.

[KHA19] Khaled N. Khasawneh, Esmaeil Mohammadian Koruyeh, Chengyu Song, Dmitry Evtyushkin, Dmitry Ponomarev, and Nael Abu-Ghazaleh. “SafeSpec: Banishing the Spectre of a Meltdown with Leakage-Free Speculation.” In *Proceedings of the 56th ACM/IEEE Design Automation Conference*, 2019. Preprint: arXiv:1806.05179.

[PAT81] David A. Patterson and Carlo H. Sequin. “RISC I: A Reduced Instruction Set VLSI Computer.” In *Proceedings of the 8th Annual Symposium on Computer Architecture*, 1981.

[COL25-D] Cologne Chip AG. *GateMate FPGA CCGM1A1/CCGM1A2 Datasheet*, DS1001, August 2025. <https://www.colognechip.com/docs/ds1001-gatemate1-datasheet-latest.pdf>

[COL26-T] Cologne Chip AG. *GateMate FPGA Toolchain Installation User Guide*, UG1002, May 2026. <https://www.colognechip.com/docs/ug1002-toolchain-install-latest.pdf>

[COL25-I] Cologne Chip AG. *GateMate FPGA Integrated Logic Analyzer User Guide*, UG1005, August 2025. <https://www.colognechip.com/docs/ug1005-gatemate1-ila-latest.pdf>

[NPNR] YosysHQ. *nextpnr: portable FPGA place and route tool*. Repository and documentation. <https://github.com/YosysHQ/nextpnr>

# Colophon

This textbook is an original synthesis and design framework. The named patterns are not claimed as terminology used by the cited authors. Historical mechanisms are credited in each pattern’s lineage and generalized into contracts for modern compositional hardware design.

The Markdown manuscript is the source of truth for the accompanying PDF. Figures are generated from Graphviz descriptions and are included in the source bundle. Resource figures remain design budgets until measured in a specific RTL implementation and toolchain run.
