package dataflow

import "github.com/pkg/errors"

// Graph is a bounded descriptor image. Loaded graphs use topological node IDs.
// ResetGraph is the original laboratory program, initialized directly at reset.
type Graph struct {
	Count       byte              `json:"count"`
	Descriptors [Nodes]Descriptor `json:"descriptors"`
}

func ResetGraph() Graph { return Graph{Nodes, Descriptors} }
func required(op Opcode) byte {
	if op == Copy || op == BoolToInt {
		return 1
	}
	return 3
}
func (g Graph) Validate() error {
	if g.Count < 1 || g.Count > Nodes {
		return errors.New("graph requires 1..7 nodes")
	}
	var writers [Nodes][2]bool
	finals := 0
	for n := 0; n < int(g.Count); n++ {
		d := g.Descriptors[n]
		if d.Op > Sub || d.Required != required(d.Op) || d.Count > 2 {
			return errors.Errorf("node %d: invalid opcode, required ports or fanout", n)
		}
		if d.Final {
			finals++
			if d.Count != 0 {
				return errors.Errorf("node %d: final node has destinations", n)
			}
		} else if d.Count == 0 {
			return errors.Errorf("node %d: nonfinal node has no consumer", n)
		}
		for i := 0; i < 2; i++ {
			dest := d.Destinations[i]
			if i >= int(d.Count) {
				if dest != (Destination{}) {
					return errors.Errorf("node %d: unused destination must be zero", n)
				}
				continue
			}
			if int(dest.Node) <= n || dest.Node >= g.Count || dest.Port > 1 {
				return errors.Errorf("node %d: destination must point forward within graph", n)
			}
			if g.Descriptors[dest.Node].Required&(1<<dest.Port) == 0 {
				return errors.Errorf("node %d: destination port is unused", n)
			}
			if writers[dest.Node][dest.Port] {
				return errors.Errorf("node %d port %d has multiple writers", dest.Node, dest.Port)
			}
			writers[dest.Node][dest.Port] = true
		}
	}
	if finals != 1 {
		return errors.New("graph must have exactly one final node")
	}
	// Forward edges, nonempty nonfinal fanout and one sink imply all nodes reach it.
	for n := int(g.Count); n < Nodes; n++ {
		if g.Descriptors[n] != (Descriptor{}) {
			return errors.New("inactive descriptors must be zero")
		}
	}
	return nil
}
func (g Graph) Bytes(n byte) [4]byte {
	d := g.Descriptors[n]
	flags := byte(d.Op)<<4 | d.Count
	if d.Final {
		flags |= 8
	}
	return [4]byte{n, flags, d.Destinations[0].Node<<1 | d.Destinations[0].Port, d.Destinations[1].Node<<1 | d.Destinations[1].Port}
}
func (s *state) graph() Graph {
	if s.Graph == nil {
		return ResetGraph()
	}
	return *s.Graph
}
func (s *state) descriptor(n byte) Descriptor { return s.graph().Descriptors[n] }
func (s *state) completion(c, e, n byte, v Value) Token {
	return Token{Context: c, Epoch: e, Node: n, Producer: n, Final: s.descriptor(n).Final, Value: v}
}
func (m *Transaction) LoadGraph(g Graph) error {
	if err := g.Validate(); err != nil {
		return err
	}
	if m.Metrics.Source != 0 || m.Metrics.Cycles != 0 || m.Epoch != ([Contexts]byte{}) || m.Closed != ([Contexts]bool{}) || !m.Quiescent() {
		return errors.New("graph load requires fresh reset before execution")
	}
	m.Graph = &g
	return nil
}
func (m *Semantic) LoadGraph(g Graph) error {
	if err := g.Validate(); err != nil {
		return err
	}
	if m.Slots != ([Contexts][Nodes]slot{}) || m.Epoch != ([Contexts]byte{}) || m.Closed != ([Contexts]bool{}) {
		return errors.New("graph load requires fresh semantic state")
	}
	m.Graph = &g
	return nil
}
