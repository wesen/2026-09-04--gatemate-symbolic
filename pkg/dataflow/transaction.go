package dataflow

import (
	"context"

	"github.com/pkg/errors"
)

type Config struct{ InputDepth, CompletionDepth, OutputDepth, MulLatency, EpochBits int }

func DefaultConfig() Config { return Config{8, 8, 8, 4, 8} }

type Metrics struct {
	Cycles, Source, Activations, Mul, ALU, StaleInput, StaleIssue, StaleCompletion, StaleRouter, StaleOutput, Invalid, Duplicates, UnitBusy, UnitBlocked, RouterBlocked uint32
	InputHigh, ReadyHigh, CompletionHigh, OutputHigh                                                                                                                    int
}
type Activation struct {
	Cycle                uint32
	Context, Epoch, Node byte
	Value                Value
}
type issue struct {
	Context, Epoch, Node byte
	Wait                 int
	Values               [2]Value
}
type routing struct {
	Token     Token
	Delivered byte
}

// Transaction is a bounded clocked reference, separate from immediate semantics.
type Transaction struct {
	state
	Config                   Config
	Metrics                  Metrics
	Trace                    []Activation
	input, completed, output []Token
	alu, mul                 []*Token
	issue                    *issue
	router                   *routing
	pendingError             [Contexts]*Token
	nextSlot, nextUnit       int
	preferRouter             bool
}

func NewTransaction(c Config) (*Transaction, error) {
	if c.InputDepth < 1 || c.CompletionDepth < 1 || c.OutputDepth < 1 || c.MulLatency < 1 || c.MulLatency > 8 || c.EpochBits < 1 || c.EpochBits > 8 {
		return nil, errors.New("invalid transaction configuration")
	}
	return &Transaction{Config: c, alu: make([]*Token, 1), mul: make([]*Token, c.MulLatency)}, nil
}
func (m *Transaction) Inject(t Token) bool {
	if len(m.input) == m.Config.InputDepth {
		return false
	}
	m.input = append(m.input, t)
	m.Metrics.Source++
	m.highwater()
	return true
}
func (m *Transaction) EpochFor(c byte) byte { return m.Epoch[c] }
func (m *Transaction) Quiescent() bool {
	if len(m.input)+len(m.completed)+len(m.output) != 0 || m.issue != nil || m.router != nil {
		return false
	}
	for _, p := range [][]*Token{m.alu, m.mul} {
		for _, t := range p {
			if t != nil {
				return false
			}
		}
	}
	for c := 0; c < Contexts; c++ {
		if m.pendingError[c] != nil {
			return false
		}
		for _, s := range m.Slots[c] {
			if s.Pending {
				return false
			}
		}
	}
	return true
}
func (m *Transaction) Cancel(c byte) bool {
	if c >= Contexts {
		return false
	}
	if len(m.output) > 0 && m.output[0].Context == c && m.output[0].Epoch == m.Epoch[c] {
		return false
	}
	if int(m.Epoch[c]) == (1<<m.Config.EpochBits)-1 && !m.Quiescent() {
		return false
	}
	m.Epoch[c] = (m.Epoch[c] + 1) & byte((1<<m.Config.EpochBits)-1)
	m.Closed[c] = false
	m.invalidate(c)
	m.pendingError[c] = nil
	return true
}
func (m *Transaction) Poll() *Token {
	for len(m.output) > 0 {
		t := m.output[0]
		m.output = m.output[1:]
		if t.Epoch != m.Epoch[t.Context] {
			m.Metrics.StaleOutput++
			continue
		}
		return &t
	}
	return nil
}
func (m *Transaction) fail(t Token, code byte) {
	if m.Closed[t.Context] {
		return
	}
	m.Closed[t.Context] = true
	for n := 0; n < Nodes; n++ {
		m.Slots[t.Context][n].Pending = false
	}
	f := fault(t.Context, t.Epoch, t.Node, code)
	m.pendingError[t.Context] = &f
	if code == DuplicateOperand {
		m.Metrics.Duplicates++
	}
}
func (m *Transaction) commit(t Token) {
	if t.Context >= Contexts {
		m.Metrics.Invalid++
		return
	}
	if t.Epoch != m.Epoch[t.Context] {
		m.Metrics.StaleInput++
		return
	}
	if m.Closed[t.Context] {
		return
	}
	if code := m.accept(t); code != 0 {
		m.fail(t, code)
	}
}
func (m *Transaction) route() bool {
	if m.router == nil {
		return false
	}
	r := m.router
	t := r.Token
	if t.Epoch != m.Epoch[t.Context] {
		m.Metrics.StaleRouter++
		m.router = nil
		return false
	}
	if m.Closed[t.Context] {
		m.router = nil
		return false
	}
	if t.Value.Tag() == 13 {
		m.fail(t, byte(t.Value))
		m.router = nil
		return false
	}
	d := m.descriptor(byte(t.Producer))
	if d.Final {
		if len(m.output) == m.Config.OutputDepth {
			m.Metrics.RouterBlocked++
			return false
		}
		m.output = append(m.output, t)
		m.Closed[t.Context] = true
		m.router = nil
		return false
	}
	for i := byte(0); i < d.Count; i++ {
		if r.Delivered&(1<<i) != 0 {
			continue
		}
		dest := d.Destinations[i]
		m.commit(Token{Context: t.Context, Epoch: t.Epoch, Node: dest.Node, Port: dest.Port, Producer: t.Producer, Value: t.Value})
		r.Delivered |= 1 << i
		if r.Delivered == (1<<d.Count)-1 {
			m.router = nil
		}
		return true
	}
	m.router = nil
	return false
}
func (m *Transaction) highwater() {
	ready := 0
	for _, slots := range m.Slots {
		for _, s := range slots {
			if s.Pending {
				ready++
			}
		}
	}
	m.Metrics.InputHigh = max(m.Metrics.InputHigh, len(m.input))
	m.Metrics.ReadyHigh = max(m.Metrics.ReadyHigh, ready)
	m.Metrics.CompletionHigh = max(m.Metrics.CompletionHigh, len(m.completed))
	m.Metrics.OutputHigh = max(m.Metrics.OutputHigh, len(m.output))
}
func (m *Transaction) Tick(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.Metrics.Cycles++
	// Pending errors have bounded per-context storage and first priority at output.
	for c, t := range m.pendingError {
		if t != nil && len(m.output) < m.Config.OutputDepth {
			m.output = append(m.output, *t)
			m.pendingError[c] = nil
			break
		}
	}
	// Route one operand or consume one external operand, with alternating priority.
	used := false
	if m.preferRouter || len(m.input) == 0 {
		used = m.route()
	}
	if !used && len(m.input) > 0 {
		t := m.input[0]
		m.input = m.input[1:]
		m.commit(t)
		used = true
		m.preferRouter = true
	} else if used {
		m.preferRouter = false
	}
	if m.router == nil && len(m.completed) > 0 {
		m.router = &routing{Token: m.completed[0]}
		m.completed = m.completed[1:]
	}
	// At most one live completion is transferred per cycle; stale exits need no credit.
	pipes := [][]*Token{m.alu, m.mul}
	taken := false
	for j := 0; j < 2; j++ {
		u := (m.nextUnit + j) % 2
		p := pipes[u]
		t := p[len(p)-1]
		if t == nil {
			continue
		}
		if t.Epoch != m.Epoch[t.Context] {
			p[len(p)-1] = nil
			m.Metrics.StaleCompletion++
			continue
		}
		if !taken && len(m.completed) < m.Config.CompletionDepth {
			m.completed = append(m.completed, *t)
			p[len(p)-1] = nil
			taken = true
			m.nextUnit = 1 - u
		} else {
			m.Metrics.UnitBlocked++
		}
	}
	for _, p := range pipes {
		for i := len(p) - 1; i > 0; i-- {
			if p[i] == nil {
				p[i] = p[i-1]
				p[i-1] = nil
			}
		}
		for _, t := range p {
			if t != nil {
				m.Metrics.UnitBusy++
				break
			}
		}
	}
	if i := m.issue; i != nil {
		if i.Epoch != m.Epoch[i.Context] || m.Closed[i.Context] {
			if i.Epoch != m.Epoch[i.Context] {
				m.Metrics.StaleIssue++
			}
			m.issue = nil
		} else if i.Wait > 0 {
			i.Wait--
		} else {
			d := m.descriptor(byte(i.Node))
			p := m.alu
			if d.Op == Mul {
				p = m.mul
			}
			if p[0] == nil {
				v := Evaluate(d.Op, i.Values[0], i.Values[1])
				t := m.completion(i.Context, i.Epoch, i.Node, v)
				p[0] = &t
				m.Metrics.Activations++
				if d.Op == Mul {
					m.Metrics.Mul++
				} else {
					m.Metrics.ALU++
				}
				m.Trace = append(m.Trace, Activation{m.Metrics.Cycles, i.Context, i.Epoch, i.Node, v})
				if len(m.Trace) > 4096 {
					m.Trace = m.Trace[len(m.Trace)-4096:]
				}
				m.issue = nil
			}
		}
	}
	if m.issue == nil {
		for offset := 0; offset < Contexts*Nodes; offset++ {
			index := (m.nextSlot + offset) % (Contexts * Nodes)
			c, n := index/Nodes, index%Nodes
			s := &m.Slots[c][n]
			if !s.Pending || m.Closed[c] {
				continue
			}
			p := m.alu
			if m.descriptor(byte(n)).Op == Mul {
				p = m.mul
			}
			if p[0] != nil {
				continue
			}
			s.Pending = false
			m.issue = &issue{byte(c), m.Epoch[c], byte(n), 2, s.Values}
			m.nextSlot = (index + 1) % (Contexts * Nodes)
			break
		}
	}
	m.highwater()
	return nil
}
