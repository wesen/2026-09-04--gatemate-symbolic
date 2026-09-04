package microscope

import (
	"context"
	"reflect"
	"testing"
)

func oracle(g Graph) [][]int {
	var results [][]int
	r := make([]int, g.Vertices)
	var visit func(int)
	visit = func(v int) {
		if v == g.Vertices {
			results = append(results, append([]int(nil), r...))
			return
		}
		for color := 0; color < g.Colors; color++ {
			ok := true
			for _, e := range g.Edges {
				u := -1
				if e[0] == v && e[1] < v {
					u = e[1]
				}
				if e[1] == v && e[0] < v {
					u = e[0]
				}
				if u >= 0 && r[u] == color {
					ok = false
				}
			}
			if ok {
				r[v] = color
				visit(v + 1)
			}
		}
	}
	visit(0)
	return results
}
func enumerate(t *testing.T, g Graph, model *Model) ([][]int, []Event) {
	t.Helper()
	ctx := context.Background()
	if err := model.Load(ctx, g); err != nil {
		t.Fatal(err)
	}
	snapshot := Initial(g)
	var outputs [][]int
	var events []Event
	for i := 0; i < 100000; i++ {
		e, err := model.Step(ctx)
		if err != nil {
			t.Fatal(err)
		}
		wire := EncodeEvent(e)
		decoded, err := DecodeEvent(wire)
		if err != nil || decoded != e {
			t.Fatalf("round trip: %v %v", decoded, err)
		}
		next, err := Apply(g, snapshot, e)
		if err != nil {
			t.Fatal(err)
		}
		snapshot = next
		events = append(events, e)
		if e.Kind == Output {
			outputs = append(outputs, e.Rows(g.Vertices))
		}
		if e.Terminal() {
			return outputs, events
		}
	}
	t.Fatal("watchdog")
	return nil, nil
}

func TestAllSmallGraphs(t *testing.T) {
	// Every simple graph on four vertices, with one through three colors.
	pairs := [][2]int{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}}
	for mask := 0; mask < 64; mask++ {
		for colors := 1; colors <= 3; colors++ {
			g := Graph{Vertices: 4, Colors: colors}
			for bit, e := range pairs {
				if mask&(1<<bit) != 0 {
					g.Edges = append(g.Edges, e)
				}
			}
			got, events := enumerate(t, g, NewModel())
			want := oracle(g)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("graph %v: got %v want %v", g, got, want)
			}
			if events[len(events)-1].Kind != Complete {
				t.Fatal("unexpected fault")
			}
		}
	}
}
func TestGraphExamples(t *testing.T) {
	for _, g := range []Graph{{3, 3, [][2]int{{0, 1}, {1, 2}, {0, 2}}, false}, {3, 2, [][2]int{{0, 1}, {1, 2}, {0, 2}}, false}, {4, 2, [][2]int{{0, 1}, {1, 2}, {2, 3}}, false}, {8, 1, nil, false}, {1, 8, nil, false}} {
		got, _ := enumerate(t, g, NewModel())
		if !reflect.DeepEqual(got, oracle(g)) {
			t.Fatal(g)
		}
		g.FirstOnly = true
		got, events := enumerate(t, g, NewModel())
		want := oracle(g)
		if len(want) > 0 {
			want = want[:1]
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("cut", g)
		}
		if len(got) > 0 {
			last := events[len(events)-1]
			if last.ChoiceTop != 0 || last.Base != last.TrailTop {
				t.Fatal("cut pointers")
			}
		}
	}
}
func TestCapacityFaultsAndReset(t *testing.T) {
	g := Graph{3, 3, [][2]int{{0, 1}, {1, 2}, {0, 2}}, false}
	for _, caps := range [][2]int{{0, 64}, {8, 0}, {1, 64}, {8, 1}} {
		m := NewModel()
		m.ChoiceCapacity = caps[0]
		m.TrailCapacity = caps[1]
		_, events := enumerate(t, g, m)
		if events[len(events)-1].Kind != Fault {
			t.Fatal("expected fault", caps)
		}
		if _, err := m.Step(context.Background()); err != ErrTerminal {
			t.Fatal(err)
		}
		if err := m.Reset(context.Background()); err != nil {
			t.Fatal(err)
		}
		e, _ := m.Step(context.Background())
		if e.Sequence != 1 {
			t.Fatal(e)
		}
	}
}
func TestInvalidGraphsAndCancelledStep(t *testing.T) {
	for _, g := range []Graph{{0, 2, nil, false}, {9, 2, nil, false}, {2, 0, nil, false}, {2, 9, nil, false}, {2, 2, [][2]int{{0, 0}}, false}, {2, 2, [][2]int{{0, 2}}, false}, {2, 2, [][2]int{{0, 1}, {1, 0}}, false}} {
		if _, err := g.Adjacency(); err == nil {
			t.Fatal(g)
		}
	}
	m := NewModel()
	if err := m.Load(context.Background(), Graph{2, 2, nil, false}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := m.Step(ctx); err != context.Canceled {
		t.Fatal(err)
	}
	e, _ := m.Step(context.Background())
	if e.Sequence != 1 {
		t.Fatal(e)
	}
}
func TestProjectionRejectsCorruptionAndCopies(t *testing.T) {
	g := Graph{2, 2, [][2]int{{0, 1}}, false}
	m := NewModel()
	_ = m.Load(context.Background(), g)
	initial := Initial(g)
	e, _ := m.Step(context.Background())
	broken := e
	broken.Sequence++
	if _, err := Apply(g, initial, broken); err == nil {
		t.Fatal("sequence gap")
	}
	s, err := Apply(g, initial, e)
	if err != nil {
		t.Fatal(err)
	}
	next, _ := m.Step(context.Background())
	broken = next
	broken.TrailWord ^= 512
	if _, err := Apply(g, s, broken); err == nil {
		t.Fatal("wrong old mask")
	}
	copy := s.Clone()
	copy.Choices[0].Remaining = 0
	if s.Choices[0].Remaining == 0 {
		t.Fatal("snapshot alias")
	}
}
