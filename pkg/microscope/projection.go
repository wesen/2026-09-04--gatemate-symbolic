package microscope

import "github.com/pkg/errors"

type Snapshot struct {
	Event
	Name    string       `json:"name"`
	Choices []Choice     `json:"choices"`
	Trail   []TrailEntry `json:"trail"`
}

func Initial(g Graph) Snapshot {
	return Snapshot{Event: Event{Domains: g.Domains(), Propagated: ^byte((1 << g.Vertices) - 1)}, Name: "READY", Choices: []Choice{}, Trail: []TrailEntry{}}
}
func (s Snapshot) Clone() Snapshot {
	s.Choices = append([]Choice{}, s.Choices...)
	s.Trail = append([]TrailEntry{}, s.Trail...)
	return s
}

// Apply validates a semantic transition before publishing an immutable snapshot.
func Apply(g Graph, previous Snapshot, e Event) (Snapshot, error) {
	bad := func(message string) (Snapshot, error) {
		return Snapshot{}, errors.Errorf("event %d %s: %s", e.Sequence, EventName(e.Kind), message)
	}
	if previous.Terminal() {
		return bad("event after terminal state")
	}
	if e.Sequence != previous.Sequence+1 {
		return bad("noncontiguous sequence")
	}
	if e.Kind < Create || e.Kind > Fault || e.ChoiceTop > 8 || e.ChoiceTop < 0 || e.TrailTop > 64 || e.TrailTop < 0 || e.Base < 0 || e.Base > e.TrailTop {
		return bad("invalid fields")
	}
	next := previous.Clone()
	expectedDomains := previous.Domains
	expectedP := previous.Propagated
	expectedCount, expectedBase := previous.Count, previous.Base
	switch e.Kind {
	case Create:
		c := DecodeChoice(e.ChoiceWord)
		if e.ChoiceWord&16383 != 0 || c.Vertex >= g.Vertices || c.Mark != len(next.Trail) || c.Propagated != previous.Propagated || Singleton(previous.Domains[c.Vertex]) || previous.Domains[c.Vertex] == 0 || c.Remaining != previous.Domains[c.Vertex]&^first(previous.Domains[c.Vertex]) {
			return bad("invalid new checkpoint")
		}
		next.Choices = append(next.Choices, c)
	case Update:
		if len(next.Choices) == 0 {
			return bad("update without checkpoint")
		}
		old := next.Choices[len(next.Choices)-1]
		c := DecodeChoice(e.ChoiceWord)
		expected := old
		expected.Remaining &^= first(expected.Remaining)
		if old.Remaining == 0 || c != expected || e.ChoiceWord&16383 != 0 {
			return bad("invalid alternative update")
		}
		next.Choices[len(next.Choices)-1] = c
	case Write:
		entry := DecodeTrail(e.TrailWord)
		if e.TrailWord>>20 != 0 || e.TrailWord&15 != 0 || entry.Vertex >= g.Vertices || entry.Level != len(next.Choices) || entry.Old != previous.Domains[entry.Vertex] || entry.Old == 0 {
			return bad("invalid mutation record")
		}
		mask := e.Domains[entry.Vertex]
		if mask == entry.Old || mask&^entry.Old != 0 {
			return bad("mutation is not narrowing")
		}
		expectedDomains[entry.Vertex] = mask
		expectedP &^= 1 << entry.Vertex
		next.Trail = append(next.Trail, entry)
	case Propagated:
		added := e.Propagated &^ previous.Propagated
		if !Singleton(added) || e.Propagated|previous.Propagated != e.Propagated {
			return bad("invalid propagation publication")
		}
		for v := 0; v < 8; v++ {
			if added&(1<<v) != 0 && !Singleton(previous.Domains[v]) {
				return bad("propagation source is not singleton")
			}
		}
		expectedP = e.Propagated
	case Contradiction:
		found := false
		for _, d := range e.Domains {
			found = found || d == 0
		}
		if !found {
			return bad("contradiction without empty domain")
		}
	case Restore:
		if len(next.Trail) <= previous.Base {
			return bad("restore below base")
		}
		entry := next.Trail[len(next.Trail)-1]
		if e.TrailWord != entry.Word() {
			return bad("restore record mismatch")
		}
		if previous.Domains[entry.Vertex]&^entry.Old != 0 {
			return bad("restore old mask is not a superset")
		}
		expectedDomains[entry.Vertex] = entry.Old
		next.Trail = next.Trail[:len(next.Trail)-1]
	case Restored:
		if len(next.Choices) == 0 {
			return bad("restoration without checkpoint")
		}
		c := next.Choices[len(next.Choices)-1]
		if len(next.Trail) != c.Mark {
			return bad("restoration did not reach mark")
		}
		expectedP = c.Propagated
	case Pop:
		if len(next.Choices) == 0 || next.Choices[len(next.Choices)-1].Remaining != 0 {
			return bad("pop of nonexhausted checkpoint")
		}
		next.Choices = next.Choices[:len(next.Choices)-1]
	case Output:
		for _, d := range e.Domains {
			if !Singleton(d) {
				return bad("output without singleton domains")
			}
		}
		if packedResult(e.Domains) != e.Result {
			return bad("output packing mismatch")
		}
		rows := e.Rows(g.Vertices)
		for _, edge := range g.Edges {
			if rows[edge[0]] == rows[edge[1]] {
				return bad("output violates edge")
			}
		}
		for _, r := range rows {
			if r >= g.Colors {
				return bad("output color outside graph")
			}
		}
		expectedCount++
		if g.FirstOnly {
			next.Choices = []Choice{}
			expectedBase = len(next.Trail)
		}
	case Complete:
		if len(next.Choices) != 0 {
			return bad("completion with alternatives")
		}
	case Fault:
		if e.Fault == 0 || e.Fault > 5 {
			return bad("invalid fault code")
		}
	}
	if e.Kind != Fault && e.Fault != 0 {
		return bad("unexpected fault")
	}
	if e.Domains != expectedDomains || e.Propagated != expectedP {
		return bad("unannounced domain or propagation change")
	}
	if e.Count != expectedCount || e.Base != expectedBase || e.ChoiceTop != len(next.Choices) || e.TrailTop != len(next.Trail) {
		return bad("pointer or accepted-count mismatch")
	}
	if e.Kind != Output && e.Result != previous.Result {
		return bad("unannounced result change")
	}
	next.Event = e
	next.Name = EventName(e.Kind)
	return next, nil
}
