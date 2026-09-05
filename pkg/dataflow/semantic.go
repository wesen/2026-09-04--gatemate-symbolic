package dataflow

import "github.com/pkg/errors"

// Semantic evaluates ready nodes immediately. It has no pipeline timing model.
type Semantic struct {
	state
	Stale, Invalid uint32
}

func (m *Semantic) Cancel(c byte) error {
	if c >= Contexts {
		return errors.New("invalid context")
	}
	m.Epoch[c]++
	m.Closed[c] = false
	m.invalidate(c)
	return nil
}
func (m *Semantic) Deliver(input Token) []Token {
	queue := []Token{input}
	var output []Token
	for len(queue) > 0 {
		t := queue[0]
		queue = queue[1:]
		if t.Context >= Contexts {
			m.Invalid++
			continue
		}
		if t.Epoch != m.Epoch[t.Context] {
			m.Stale++
			continue
		}
		if m.Closed[t.Context] {
			continue
		}
		if code := m.accept(t); code != 0 {
			m.Closed[t.Context] = true
			output = append(output, fault(t.Context, t.Epoch, t.Node, code))
			continue
		}
		s := &m.Slots[t.Context][t.Node]
		if !s.Pending {
			continue
		}
		s.Pending = false
		d := Descriptors[t.Node]
		v := Evaluate(d.Op, s.Values[0], s.Values[1])
		if v.Tag() == 13 {
			m.Closed[t.Context] = true
			output = append(output, fault(t.Context, t.Epoch, t.Node, byte(v)))
			continue
		}
		if d.Final {
			m.Closed[t.Context] = true
			output = append(output, completion(t.Context, t.Epoch, t.Node, v))
			continue
		}
		for i := byte(0); i < d.Count; i++ {
			dest := d.Destinations[i]
			queue = append(queue, Token{Context: t.Context, Epoch: t.Epoch, Node: dest.Node, Port: dest.Port, Producer: t.Node, Value: v})
		}
	}
	return output
}
