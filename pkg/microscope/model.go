package microscope

import (
	"context"

	"github.com/pkg/errors"
)

type Model struct {
	graph                         Graph
	adjacent                      [8]byte
	domains                       [8]byte
	propagated                    byte
	choices                       []Choice
	trail                         []TrailEntry
	base                          int
	count, sequence               uint32
	result                        uint32
	fault                         byte
	loaded, terminal              bool
	phase                         string
	source, target, mutVertex     int
	mutMask                       byte
	resume                        string
	cp                            Choice
	lastTrail                     TrailEntry
	ChoiceCapacity, TrailCapacity int
}

var _ Engine = &Model{}

func NewModel() *Model { return &Model{ChoiceCapacity: 8, TrailCapacity: 64} }
func (m *Model) Load(ctx context.Context, g Graph) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a, err := g.Adjacency()
	if err != nil {
		return err
	}
	if m.ChoiceCapacity < 0 || m.ChoiceCapacity > 8 || m.TrailCapacity < 0 || m.TrailCapacity > 64 {
		return errors.New("invalid model capacities")
	}
	m.graph = g
	m.graph.Edges = append([][2]int(nil), g.Edges...)
	m.adjacent = a
	m.loaded = true
	return m.Reset(ctx)
}
func (m *Model) Reset(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !m.loaded {
		return errors.New("load a graph first")
	}
	m.domains = m.graph.Domains()
	m.propagated = ^byte((1 << m.graph.Vertices) - 1)
	m.choices = nil
	m.trail = nil
	m.base = 0
	m.count = 0
	m.sequence = 0
	m.result = 0
	m.fault = 0
	m.terminal = false
	m.phase = "scan"
	m.cp = Choice{}
	m.lastTrail = TrailEntry{}
	return nil
}
func (m *Model) Close() error { return nil }
func (m *Model) event(kind byte) Event {
	m.sequence++
	if kind == Complete || kind == Fault {
		m.terminal = true
	}
	return Event{m.sequence, kind, m.domains, m.propagated, len(m.choices), len(m.trail), m.base, m.count, m.result, m.fault, m.cp.Word(), m.lastTrail.Word()}
}
func (m *Model) fail(code byte) (Event, error) { m.fault = code; return m.event(Fault), nil }
func (m *Model) mutation(v int, mask byte, resume string) {
	m.mutVertex = v
	m.mutMask = mask
	m.resume = resume
	m.phase = "mutate"
}

func (m *Model) Step(ctx context.Context) (Event, error) {
	if !m.loaded {
		return Event{}, errors.New("load a graph first")
	}
	if m.terminal {
		return Event{}, ErrTerminal
	}
	for {
		if err := ctx.Err(); err != nil {
			return Event{}, err
		}
		switch m.phase {
		case "scan":
			zero, source, unresolved := false, -1, -1
			for v, d := range m.domains {
				if d == 0 {
					zero = true
				}
				if Singleton(d) && m.propagated&(1<<v) == 0 && source < 0 {
					source = v
				}
				if !Singleton(d) && unresolved < 0 {
					unresolved = v
				}
			}
			if zero {
				m.phase = "back"
				return m.event(Contradiction), nil
			}
			if source >= 0 {
				m.source = source
				m.target = 0
				m.phase = "prop"
				continue
			}
			if unresolved < 0 {
				m.result = packedResult(m.domains)
				m.count++
				if m.graph.FirstOnly {
					m.choices = nil
					m.base = len(m.trail)
					m.phase = "complete"
				} else {
					m.phase = "back"
				}
				return m.event(Output), nil
			}
			if len(m.choices) >= m.ChoiceCapacity {
				return m.fail(2)
			}
			if len(m.trail) >= m.TrailCapacity {
				return m.fail(1)
			}
			b := first(m.domains[unresolved])
			m.cp = Choice{unresolved, m.domains[unresolved] &^ b, len(m.trail), m.propagated}
			m.choices = append(m.choices, m.cp)
			m.mutation(unresolved, b, "scan")
			return m.event(Create), nil
		case "mutate":
			v, old := m.mutVertex, m.domains[m.mutVertex]
			if m.mutMask&^old != 0 {
				return m.fail(5)
			}
			if old == m.mutMask {
				m.phase = m.resume
				continue
			}
			if len(m.trail) >= m.TrailCapacity {
				return m.fail(1)
			}
			m.lastTrail = TrailEntry{v, old, len(m.choices)}
			m.trail = append(m.trail, m.lastTrail)
			m.domains[v] = m.mutMask
			m.propagated &^= 1 << v
			m.phase = m.resume
			return m.event(Write), nil
		case "prop":
			if !Singleton(m.domains[m.source]) {
				return m.fail(4)
			}
			if m.target == m.source {
				m.phase = "advance"
				continue
			}
			mask := m.domains[m.target]
			if m.adjacent[m.source]&(1<<m.target) != 0 {
				mask &^= m.domains[m.source]
			}
			m.mutation(m.target, mask, "advance")
		case "advance":
			if m.domains[m.target] == 0 {
				m.phase = "scan"
				continue
			}
			if m.target == 7 {
				m.propagated |= 1 << m.source
				m.phase = "scan"
				return m.event(Propagated), nil
			}
			m.target++
			m.phase = "prop"
		case "back":
			if len(m.choices) == 0 {
				m.phase = "complete"
				continue
			}
			m.cp = m.choices[len(m.choices)-1]
			m.phase = "undo"
		case "undo":
			if m.cp.Mark < m.base || m.cp.Mark > len(m.trail) {
				return m.fail(5)
			}
			if len(m.trail) > m.cp.Mark {
				m.lastTrail = m.trail[len(m.trail)-1]
				m.trail = m.trail[:len(m.trail)-1]
				m.domains[m.lastTrail.Vertex] = m.lastTrail.Old
				return m.event(Restore), nil
			}
			m.propagated = m.cp.Propagated
			m.phase = "retry"
			return m.event(Restored), nil
		case "retry":
			if m.cp.Remaining == 0 {
				m.choices = m.choices[:len(m.choices)-1]
				m.phase = "back"
				return m.event(Pop), nil
			}
			if len(m.trail) >= m.TrailCapacity {
				return m.fail(1)
			}
			b := first(m.cp.Remaining)
			m.cp.Remaining &^= b
			m.choices[len(m.choices)-1] = m.cp
			m.mutation(m.cp.Vertex, b, "scan")
			return m.event(Update), nil
		case "complete":
			return m.event(Complete), nil
		default:
			return m.fail(5)
		}
	}
}
