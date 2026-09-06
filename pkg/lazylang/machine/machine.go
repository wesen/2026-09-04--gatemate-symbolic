// Package machine implements the finite, stepped LFL1 abstract machine. It owns
// its memories; callers serialize access and use detached snapshots for inspection.
package machine

import (
	"context"
	"math"

	"github.com/pkg/errors"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/compile"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/ir"
)

type State uint8

const (
	Idle State = iota
	CodeIssue
	CodeWait
	Eval
	HeapIssue
	HeapWait
	Enter
	EnvIssue
	EnvWait
	EnvDispatch
	Return
	StackWait
	ReturnDispatch
	Allocate
	UpdateIssue
	UpdateWait
	UpdateWrite
	Multiply
	Output
	Loading
)

type Counters struct {
	Cycles                uint32 `json:"cycles"`
	CodeReads             uint32 `json:"codeReads"`
	HeapReads             uint32 `json:"heapReads"`
	Allocations           uint32 `json:"allocations"`
	AllocatedThunks       uint32 `json:"allocatedThunks"`
	AllocatedFunctions    uint32 `json:"allocatedFunctions"`
	AllocatedEnvironments uint32 `json:"allocatedEnvironments"`
	AllocatedCons         uint32 `json:"allocatedCons"`
	AllocatedIntegers     uint32 `json:"allocatedIntegers"`
	Claims                uint32 `json:"claims"`
	Updates               uint32 `json:"updates"`
	Adds                  uint32 `json:"adds"`
	Subtracts             uint32 `json:"subtracts"`
	Multiplies            uint32 `json:"multiplies"`
	Equalities            uint32 `json:"equalities"`
	ComparisonsLE         uint32 `json:"comparisonsLE"`
	Indirections          uint32 `json:"indirections"`
	EnvironmentSteps      uint32 `json:"environmentSteps"`
	MaxStack              uint32 `json:"maxStack"`
	MaxHeap               uint32 `json:"maxHeap"`
	OutputStalls          uint32 `json:"outputStalls"`
	Faults                uint32 `json:"faults"`
	TraceDropped          uint32 `json:"traceDropped"`
}
type Event struct {
	Cycle   uint32 `json:"cycle"`
	Address uint16 `json:"address"`
	Span    uint16 `json:"span"`
	Kind    uint8  `json:"kind"`
	Old     string `json:"old"`
	New     string `json:"new"`
}
type Allocation struct {
	Active      bool   `json:"active"`
	Base        uint16 `json:"base"`
	ReservedEnd uint16 `json:"reservedEnd"`
	ObjectCount uint16 `json:"objectCount"`
	NextWrite   uint16 `json:"nextWrite"`
	Kind        uint8  `json:"kind"`
	Stage       uint8  `json:"stage"`
}
type Snapshot struct {
	ArtifactID   string     `json:"artifactId"`
	State        State      `json:"state"`
	Root         uint16     `json:"root"`
	ConstantEnd  uint16     `json:"constantEnd"`
	CurrentRef   uint16     `json:"currentRef"`
	CodeID       uint16     `json:"codeId"`
	EnvRef       uint16     `json:"envRef"`
	ResultRef    uint16     `json:"resultRef"`
	Valid        bool       `json:"valid"`
	CommittedTop uint16     `json:"committedTop"`
	IndSteps     uint16     `json:"indSteps"`
	EnvSteps     uint16     `json:"envSteps"`
	Heap         []string   `json:"heap"`
	Provenance   []uint16   `json:"provenance"`
	Stack        []string   `json:"stack"`
	Allocation   Allocation `json:"allocation"`
	Trace        []Event    `json:"trace"`
	Counters     Counters   `json:"counters"`
}
type Machine struct {
	artifactID                       string
	root, constantEnd                uint16
	code                             []ir.Code
	heap                             [ir.Capacity]ir.Object
	provenance                       [ir.Capacity]uint16
	stack                            []ir.Frame
	top                              uint16
	state                            State
	ref, codeID, env, result         uint16
	valid                            bool
	instruction                      ir.Code
	object                           ir.Object
	frame                            ir.Frame
	indSteps, envSteps               uint16
	depth                            uint32
	alloc                            Allocation
	allocObjects                     []ir.Object
	allocSpan                        uint16
	resumeState                      State
	resumeCode, resumeEnv, resumeRef uint16
	mulLeft, mulRight                int64
	mulTicks                         uint8
	counters                         Counters
	trace                            []Event
}

func New(a *compile.Artifact) (*Machine, error) {
	if a == nil {
		return nil, errors.New("missing artifact")
	}
	if err := a.Validate(); err != nil {
		return nil, errors.Wrap(err, "load artifact")
	}
	m := &Machine{artifactID: a.ID, root: a.Root, constantEnd: a.ConstantEnd, top: uint16(len(a.Heap)), state: Idle, ref: ir.Absent, env: ir.Absent, result: ir.Absent, trace: []Event{}, stack: []ir.Frame{}}
	m.code = make([]ir.Code, len(a.Code))
	for i, x := range a.Code {
		m.code[i], _ = ir.ParseCode(x)
	}
	for i := range m.heap {
		m.heap[i] = ir.Object{Tag: ir.Free}
	}
	for i, x := range a.Heap {
		m.heap[i], _ = ir.ParseObject(x)
		m.provenance[i] = a.Provenance[i]
	}
	m.counters.MaxHeap = uint32(m.top)
	return m, nil
}
func (m *Machine) Snapshot() Snapshot {
	end := m.top
	if m.alloc.Active {
		end = m.alloc.ReservedEnd
	}
	out := Snapshot{ArtifactID: m.artifactID, State: m.state, Root: m.root, ConstantEnd: m.constantEnd, CurrentRef: m.ref, CodeID: m.codeID, EnvRef: m.env, ResultRef: m.result, Valid: m.valid, CommittedTop: m.top, IndSteps: m.indSteps, EnvSteps: m.envSteps, Heap: []string{}, Provenance: append([]uint16{}, m.provenance[:end]...), Stack: []string{}, Allocation: m.alloc, Trace: append([]Event{}, m.trace...), Counters: m.counters}
	for _, o := range m.heap[:end] {
		out.Heap = append(out.Heap, o.Hex())
	}
	for _, f := range m.stack {
		out.Stack = append(out.Stack, f.Hex())
	}
	return out
}
func (m *Machine) Demand(ref uint16) error {
	if m.state != Idle || m.valid {
		return errors.New("demand requires idle machine")
	}
	m.stack = m.stack[:0]
	m.indSteps = 0
	m.envSteps = 0
	m.enter(ref)
	return nil
}
func (m *Machine) Poll() (uint16, bool) {
	if !m.valid {
		return ir.Absent, false
	}
	ref := m.result
	m.valid = false
	m.state = Idle
	return ref, true
}
func (m *Machine) Tick(ctx context.Context, n uint32) error {
	if n > 1000000 {
		return errors.New("tick budget exceeds 1000000")
	}
	for i := uint32(0); i < n; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		m.step()
	}
	return nil
}
func (m *Machine) fault(code uint32)     { m.counters.Faults++; m.ret(uint16(code - 1)) }
func (m *Machine) ret(ref uint16)        { m.ref = ref; m.state = Return }
func (m *Machine) enter(ref uint16)      { m.ref = ref; m.state = HeapIssue }
func (m *Machine) eval(code, env uint16) { m.codeID = code; m.env = env; m.state = CodeIssue }
func (m *Machine) push(f ir.Frame) bool {
	if len(m.stack) >= ir.StackCapacity {
		m.fault(2)
		return false
	}
	m.stack = append(m.stack, f)
	if uint32(len(m.stack)) > m.counters.MaxStack {
		m.counters.MaxStack = uint32(len(m.stack))
	}
	return true
}
func (m *Machine) pop() { m.stack = m.stack[:len(m.stack)-1] }
func (m *Machine) event(ref uint16, kind uint8, old, new ir.Object) {
	if len(m.trace) < 64 {
		m.trace = append(m.trace, Event{Cycle: m.counters.Cycles, Address: ref, Span: m.provenance[ref], Kind: kind, Old: old.Hex(), New: new.Hex()})
	} else {
		m.counters.TraceDropped++
	}
}

// reserve prepares private objects. No evaluator-visible top changes until all
// object and provenance writes complete. Resume fields are installed at the end.
func (m *Machine) reserve(kind uint8, span uint16, objects []ir.Object, state State, code, env, ref uint16) {
	if int(m.top)+len(objects) > ir.Capacity {
		m.fault(8)
		return
	}
	m.alloc = Allocation{Active: true, Base: m.top, ReservedEnd: m.top + uint16(len(objects)), ObjectCount: uint16(len(objects)), Kind: kind}
	m.allocObjects = objects
	m.allocSpan = span
	m.resumeState = state
	m.resumeCode = code
	m.resumeEnv = env
	m.resumeRef = ref
	m.state = Allocate
}
func (m *Machine) allocationStep() {
	a := &m.alloc
	switch a.Stage {
	case 0:
		m.heap[a.Base+a.NextWrite] = m.allocObjects[a.NextWrite]
		a.NextWrite++
		if a.NextWrite == a.ObjectCount {
			a.NextWrite = 0
			a.Stage = 1
		}
	case 1:
		m.provenance[a.Base+a.NextWrite] = m.allocSpan
		a.NextWrite++
		if a.NextWrite == a.ObjectCount {
			m.top = a.ReservedEnd
			a.Stage = 2
			a.NextWrite = 0
			if uint32(m.top) > m.counters.MaxHeap {
				m.counters.MaxHeap = uint32(m.top)
			}
		}
	case 2:
		ref := a.Base + a.NextWrite
		o := m.heap[ref]
		m.event(ref, 1, ir.Object{Tag: ir.Free}, o)
		m.counters.Allocations++
		switch o.Tag {
		case ir.Thunk:
			m.counters.AllocatedThunks++
		case ir.Fun:
			m.counters.AllocatedFunctions++
		case ir.Env:
			m.counters.AllocatedEnvironments++
		case ir.Cons:
			m.counters.AllocatedCons++
		case ir.Int:
			m.counters.AllocatedIntegers++
		}
		a.NextWrite++
		if a.NextWrite == a.ObjectCount {
			a.Stage = 3
		}
	case 3:
		a.Active = false
		m.allocObjects = nil
		m.codeID = m.resumeCode
		m.env = m.resumeEnv
		m.ref = m.resumeRef
		m.state = m.resumeState
	}
}
func (m *Machine) step() {
	if m.state == Idle {
		return
	}
	m.counters.Cycles++
	switch m.state {
	case CodeIssue:
		if int(m.codeID) >= len(m.code) {
			m.fault(9)
			return
		}
		m.instruction = m.code[m.codeID]
		m.state = CodeWait
	case CodeWait:
		m.state = Eval
	case Eval:
		m.counters.CodeReads++
		m.evalInstruction()
	case HeapIssue:
		if m.ref >= m.top {
			m.fault(1)
			return
		}
		m.object = m.heap[m.ref]
		m.counters.HeapReads++
		m.state = HeapWait
	case HeapWait:
		m.state = Enter
	case Enter:
		o := m.object
		switch o.Tag {
		case ir.Int, ir.Bool, ir.Nil, ir.Cons, ir.Fun, ir.Error:
			m.indSteps = 0
			m.ret(m.ref)
		case ir.Ind:
			if m.indSteps >= ir.Capacity {
				m.fault(4)
				return
			}
			m.indSteps++
			m.counters.Indirections++
			m.enter(o.A)
		case ir.Blackhole:
			m.fault(3)
		case ir.Thunk:
			if !m.push(ir.Frame{Kind: ir.Update, A: m.ref, Span: m.provenance[m.ref]}) {
				return
			}
			claimed := ir.Object{Tag: ir.Blackhole}
			m.heap[m.ref] = claimed
			m.event(m.ref, 2, o, claimed)
			m.counters.Claims++
			m.indSteps = 0
			m.eval(o.A, o.B)
		default:
			m.fault(5)
		}
	case EnvIssue:
		if m.env == ir.Absent || m.env >= m.top || m.envSteps >= ir.Capacity {
			m.fault(10)
			return
		}
		m.object = m.heap[m.env]
		m.counters.HeapReads++
		m.state = EnvWait
	case EnvWait:
		m.state = EnvDispatch
	case EnvDispatch:
		if m.object.Tag != ir.Env {
			m.fault(10)
			return
		}
		m.envSteps++
		m.counters.EnvironmentSteps++
		if m.depth == 0 {
			m.enter(m.object.A)
		} else {
			m.depth--
			m.env = m.object.B
			m.state = EnvIssue
		}
	case Return:
		if m.ref >= m.top || !m.heap[m.ref].Terminal() {
			m.fault(7)
			return
		}
		if len(m.stack) == 0 {
			m.result = m.ref
			m.valid = true
			m.state = Output
			return
		}
		m.frame = m.stack[len(m.stack)-1]
		m.state = StackWait
	case StackWait:
		m.state = ReturnDispatch
	case ReturnDispatch:
		m.returnFrame()
	case Allocate:
		m.allocationStep()
	case UpdateIssue:
		if m.frame.A < m.constantEnd || m.frame.A >= m.top {
			m.pop()
			m.fault(7)
			return
		}
		m.object = m.heap[m.frame.A]
		m.counters.HeapReads++
		m.state = UpdateWait
	case UpdateWait:
		m.state = UpdateWrite
	case UpdateWrite:
		if m.object.Tag != ir.Blackhole {
			m.pop()
			m.fault(7)
			return
		}
		o := ir.Object{Tag: ir.Ind, A: m.ref}
		m.heap[m.frame.A] = o
		m.event(m.frame.A, 3, m.object, o)
		m.counters.Updates++
		m.pop()
		m.state = Return
	case Multiply:
		m.mulTicks++
		if m.mulTicks == 32 {
			m.integer(m.mulLeft*m.mulRight, m.frame.Span)
		}
	case Output:
		m.counters.OutputStalls++
	}
}
func (m *Machine) evalInstruction() {
	c := m.instruction
	env := m.env
	base := m.top
	switch c.Op {
	case ir.Const:
		m.ret(c.A)
	case ir.Var:
		m.depth = c.Immediate
		m.envSteps = 0
		m.state = EnvIssue
	case ir.Lambda:
		m.reserve(1, c.Span, []ir.Object{{Tag: ir.Fun, A: c.A, B: env}}, Return, 0, env, base)
	case ir.App:
		if m.push(ir.Frame{Kind: ir.Arg, A: c.B, B: env, Span: c.Span}) {
			m.eval(c.A, env)
		}
	case ir.Let, ir.Letrec:
		captured := env
		kind := uint8(3)
		if c.Op == ir.Letrec {
			captured = base + 1
			kind = 4
		}
		m.reserve(kind, c.Span, []ir.Object{{Tag: ir.Thunk, A: c.A, B: captured}, {Tag: ir.Env, A: base, B: env}}, CodeIssue, c.B, base+1, ir.Absent)
	case ir.Prim:
		if m.push(ir.Frame{Kind: ir.PrimRight, Op: uint8(c.Immediate), A: c.B, B: env, Span: c.Span}) {
			m.eval(c.A, env)
		}
	case ir.If:
		if m.push(ir.Frame{Kind: ir.Branch, A: c.B, B: env, C: c.C, Span: c.Span}) {
			m.eval(c.A, env)
		}
	case ir.MakeCons:
		m.reserve(5, c.Span, []ir.Object{{Tag: ir.Thunk, A: c.A, B: env}, {Tag: ir.Thunk, A: c.B, B: env}, {Tag: ir.Cons, A: base, B: base + 1}}, Return, 0, env, base+2)
	case ir.Case:
		if m.push(ir.Frame{Kind: ir.Match, A: c.B, B: env, C: c.C, Span: c.Span}) {
			m.eval(c.A, env)
		}
	default:
		m.fault(9)
	}
}
func (m *Machine) returnFrame() {
	f := m.frame
	o := m.heap[m.ref]
	base := m.top
	if f.Kind == ir.Update {
		m.state = UpdateIssue
		return
	}
	if o.Tag == ir.Error {
		m.pop()
		m.state = Return
		return
	}
	switch f.Kind {
	case ir.Arg:
		m.pop()
		if o.Tag != ir.Fun {
			m.fault(5)
			return
		}
		m.reserve(2, f.Span, []ir.Object{{Tag: ir.Thunk, A: f.A, B: f.B}, {Tag: ir.Env, A: base, B: o.B}}, CodeIssue, o.A, base+1, ir.Absent)
	case ir.PrimRight:
		if o.Tag != ir.Int {
			m.pop()
			m.fault(5)
			return
		}
		m.stack[len(m.stack)-1] = ir.Frame{Kind: ir.PrimApply, Op: f.Op, SavedRef: m.ref, Span: f.Span}
		m.eval(f.A, f.B)
	case ir.PrimApply:
		m.pop()
		if o.Tag != ir.Int || f.SavedRef >= m.top || m.heap[f.SavedRef].Tag != ir.Int {
			m.fault(5)
			return
		}
		left, right := int64(int32(m.heap[f.SavedRef].Payload)), int64(int32(o.Payload))
		switch ir.Primitive(f.Op) {
		case ir.Add:
			m.counters.Adds++
			m.integer(left+right, f.Span)
		case ir.Sub:
			m.counters.Subtracts++
			m.integer(left-right, f.Span)
		case ir.Mul:
			m.counters.Multiplies++
			m.mulLeft = left
			m.mulRight = right
			m.mulTicks = 0
			m.state = Multiply
		case ir.EQ:
			m.counters.Equalities++
			m.boolean(left == right)
		case ir.LE:
			m.counters.ComparisonsLE++
			m.boolean(left <= right)
		default:
			m.fault(9)
		}
	case ir.Branch:
		m.pop()
		if o.Tag != ir.Bool {
			m.fault(5)
			return
		}
		code := f.C
		if o.Payload != 0 {
			code = f.A
		}
		m.eval(code, f.B)
	case ir.Match:
		m.pop()
		if o.Tag == ir.Nil {
			m.eval(f.A, f.B)
			return
		}
		if o.Tag != ir.Cons {
			m.fault(5)
			return
		}
		m.reserve(6, f.Span, []ir.Object{{Tag: ir.Env, A: o.A, B: f.B}, {Tag: ir.Env, A: o.B, B: base}}, CodeIssue, f.C, base+1, ir.Absent)
	default:
		m.pop()
		m.fault(7)
	}
}
func (m *Machine) integer(n int64, span uint16) {
	if n < math.MinInt32 || n > math.MaxInt32 {
		m.fault(6)
		return
	}
	m.reserve(7, span, []ir.Object{{Tag: ir.Int, Payload: uint32(int32(n))}}, Return, 0, m.env, m.top)
}
func (m *Machine) boolean(v bool) {
	if v {
		m.ret(11)
	} else {
		m.ret(10)
	}
}
