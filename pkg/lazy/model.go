package lazy

import (
	"context"
	"github.com/pkg/errors"
	"sync"
)

type Model struct {
	mu     sync.Mutex
	data   Snapshot
	limit  int
	hops   int
	closed bool
}

var _ Engine = &Model{}

func NewModel() *Model { return NewModelWithStack(StackCapacity) }
func NewModelWithStack(n int) *Model {
	if n < 1 || n > StackCapacity {
		panic("invalid model stack capacity")
	}
	m := &Model{limit: n}
	m.reset()
	return m
}
func (m *Model) reset() {
	m.data = Snapshot{Source: "model", State: "idle", Heap: []Word{}, Stack: []Frame{}, Trace: []Mutation{}}
	m.hops = 0
}
func (m *Model) Close() error { m.mu.Lock(); defer m.mu.Unlock(); m.closed = true; return nil }
func (m *Model) Snapshot(ctx context.Context) (Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	if m.closed {
		return Snapshot{}, errors.New("engine closed")
	}
	s := m.data
	s.Heap = append([]Word{}, s.Heap...)
	s.Stack = append([]Frame{}, s.Stack...)
	s.Trace = append([]Mutation{}, s.Trace...)
	return s, nil
}
func (m *Model) Execute(ctx context.Context, o Operation) (*Word, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.closed {
		return nil, errors.New("engine closed")
	}
	switch o.Kind {
	case "reset":
		m.reset()
	case "load":
		m.reset()
		m.data.Heap = append([]Word{}, o.Image.Nodes...)
	case "force":
		if m.data.State != "idle" {
			return nil, errors.New("poll the offered result or finish execution before forcing")
		}
		m.data.Address = o.Root
		m.data.State = "fetch"
		m.data.Result = 0
		m.hops = 0
	case "tick":
		for n := uint32(0); n < o.Ticks; n++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			m.step()
		}
	case "poll":
		if m.data.Valid {
			v := m.data.Result
			m.data.Valid = false
			m.data.State = "idle"
			return &v, nil
		}
	}
	return nil, nil
}
func (m *Model) fault(code uint32) {
	m.data.Counters.Faults++
	m.data.Result = Fault(code)
	m.data.State = "return"
	m.hops = 0
}
func (m *Model) push(f Frame) bool {
	if len(m.data.Stack) == m.limit {
		m.fault(StackOverflow)
		return false
	}
	m.data.Stack = append(m.data.Stack, f)
	if uint32(len(m.data.Stack)) > m.data.Counters.MaxStack {
		m.data.Counters.MaxStack = uint32(len(m.data.Stack))
	}
	return true
}
func (m *Model) write(a uint16, v Word) {
	old := m.data.Heap[a]
	m.data.Heap[a] = v
	m.data.Counters.Writes++
	e := Mutation{m.data.Counters.Cycles, a, old, v}
	if len(m.data.Trace) < TraceCapacity {
		m.data.Trace = append(m.data.Trace, e)
	} else {
		m.data.Dropped++
	}
}
func (m *Model) step() {
	s := &m.data
	c := &s.Counters
	c.Cycles++
	switch s.State {
	case "idle":
		return
	case "output":
		c.Stalls++
		return
	case "fetch":
		if int(s.Address) >= len(s.Heap) {
			m.fault(AddressFault)
			return
		}
		c.Reads++
		w := s.Heap[s.Address]
		if !w.Canonical() {
			m.fault(TypeFault)
			return
		}
		if w.Tag() != Ind {
			m.hops = 0
		}
		switch w.Tag() {
		case Int, Error:
			s.Result = w
			s.State = "return"
		case Ind:
			if m.hops == IndirectionLimit {
				m.fault(IndirectionCycle)
				return
			}
			m.hops++
			c.Indirections++
			s.Address = w.A()
		case Add, Mul:
			if m.push(Frame{Kind: EvalRight, Op: w.Tag(), Address: w.B()}) {
				s.Address = w.A()
			}
		case Thunk:
			if m.push(Frame{Kind: Update, Address: s.Address}) {
				m.write(s.Address, Node(Blackhole, 0, 0))
				c.Claims++
				s.Address = w.A()
			}
		case Blackhole:
			c.Blackholes++
			m.fault(CyclicThunk)
		default:
			m.fault(TypeFault)
		}
	case "return":
		if len(s.Stack) == 0 {
			s.Valid = true
			s.State = "output"
			return
		}
		i := len(s.Stack) - 1
		f := s.Stack[i]
		switch f.Kind {
		case EvalRight:
			if s.Result.Tag() == Error {
				s.Stack = s.Stack[:i]
			} else {
				s.Stack[i] = Frame{Kind: Apply, Op: f.Op, Value: s.Result}
				s.Address = f.Address
				s.State = "fetch"
				m.hops = 0
			}
		case Apply:
			s.Stack = s.Stack[:i]
			if s.Result.Tag() != Error {
				if f.Op == Mul {
					c.Muls++
				} else {
					c.Adds++
				}
				s.Result = Evaluate(f.Op, f.Value, s.Result)
				if s.Result.Tag() == Error {
					c.Faults++
				}
			}
		case Update:
			s.Stack = s.Stack[:i]
			if int(f.Address) >= len(s.Heap) || s.Heap[f.Address] != Node(Blackhole, 0, 0) {
				m.fault(OwnershipFault)
			} else {
				m.write(f.Address, s.Result)
				c.Updates++
			}
		default:
			m.fault(OwnershipFault)
		}
	}
}
