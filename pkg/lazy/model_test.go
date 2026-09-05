package lazy

import (
	"context"
	"math/rand"
	"reflect"
	"testing"
)

func run(t *testing.T, image Image, limit int) (Word, Snapshot, *Model) {
	t.Helper()
	m := NewModelWithStack(limit)
	ctx := context.Background()
	if _, e := m.Execute(ctx, Operation{Kind: "load", Image: &image}); e != nil {
		t.Fatal(e)
	}
	if _, e := m.Execute(ctx, Operation{Kind: "force", Root: image.Root}); e != nil {
		t.Fatal(e)
	}
	for n := 0; n < 10000; n++ {
		_, e := m.Execute(ctx, Operation{Kind: "tick", Ticks: 1})
		if e != nil {
			t.Fatal(e)
		}
		s, _ := m.Snapshot(ctx)
		if s.Valid {
			return s.Result, s, m
		}
	}
	t.Fatal("reducer did not terminate")
	return 0, Snapshot{}, nil
}
func TestExamplesAndHeldOutput(t *testing.T) {
	cases := map[string]Word{"shared": Integer(168), "cycle": Fault(CyclicThunk), "indirection": Integer(-42), "indirection-cycle": Fault(IndirectionCycle), "overflow": Fault(ArithmeticOverflow)}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			i, _ := Example(name)
			got, s, m := run(t, i, StackCapacity)
			if got != want {
				t.Fatalf("got %x want %x", got, want)
			}
			if len(s.Stack) != 0 {
				t.Fatal("output with live continuations")
			}
			for _, w := range s.Heap {
				if w.Tag() == Blackhole {
					t.Fatal("permanent blackhole")
				}
			}
			ctx := context.Background()
			_, _ = m.Execute(ctx, Operation{Kind: "tick", Ticks: 11})
			held, _ := m.Snapshot(ctx)
			if !held.Valid || held.Result != got || !reflect.DeepEqual(held.Heap, s.Heap) || held.Counters.Stalls != s.Counters.Stalls+11 {
				t.Fatal("held output changed")
			}
			if _, e := m.Execute(ctx, Operation{Kind: "force", Root: i.Root}); e == nil {
				t.Fatal("force while output held")
			}
			v, _ := m.Execute(ctx, Operation{Kind: "poll"})
			if v == nil || *v != got {
				t.Fatal("poll")
			}
			_, _ = m.Execute(ctx, Operation{Kind: "force", Root: i.Root})
			_, _ = m.Execute(ctx, Operation{Kind: "tick", Ticks: 2000})
			again, _ := m.Snapshot(ctx)
			if again.Result != want || !again.Valid {
				t.Fatal("repeat force")
			}
			if name == "shared" && (s.Counters.Muls != 1 || s.Counters.Claims != 1 || s.Counters.Updates != 1 || again.Counters.Muls != 1) {
				t.Fatalf("memoization %+v %+v", s.Counters, again.Counters)
			}
		})
	}
}
func TestReserveBeforeClaimAndUnwind(t *testing.T) {
	i := Image{[]Word{Integer(1), Node(Thunk, 0, 0), Node(Add, 1, 0), Node(Thunk, 2, 0)}, 3}
	got, s, _ := run(t, i, 1)
	if got != Fault(StackOverflow) || s.Heap[3] != got || s.Heap[1].Tag() != Thunk || s.Counters.Claims != 1 || s.Counters.Updates != 1 {
		t.Fatalf("reserve/unwind %x %+v", got, s)
	}
}
func TestAddressFailFastAndIndirection(t *testing.T) {
	for _, i := range []Image{{[]Word{Node(Thunk, 42, 0)}, 0}, {[]Word{Node(Add, 42, 1), Node(Thunk, 2, 0), Integer(9)}, 0}, {[]Word{Node(Thunk, 1, 0), Node(Ind, 1, 0)}, 0}} {
		got, s, _ := run(t, i, StackCapacity)
		ref, heap := Reference(i, StackCapacity)
		if got != ref || !reflect.DeepEqual(heap, s.Heap) {
			t.Fatal("fault mismatch")
		}
	}
}
func TestGeneratedGraphsMatchRecursiveReference(t *testing.T) {
	r := rand.New(rand.NewSource(9009))
	for trial := 0; trial < 400; trial++ {
		nodes := []Word{Integer(int32(r.Intn(101) - 50)), Integer(int32(r.Intn(101) - 50))}
		for n := 2; n < 30; n++ {
			a, b := uint16(r.Intn(n)), uint16(r.Intn(n))
			switch r.Intn(5) {
			case 0:
				nodes = append(nodes, Node(Thunk, a, 0))
			case 1:
				nodes = append(nodes, Node(Ind, a, 0))
			case 2:
				nodes = append(nodes, Node(Mul, a, b))
			default:
				nodes = append(nodes, Node(Add, a, b))
			}
		}
		i := Image{nodes, 29}
		limit := r.Intn(20) + 1
		want, heap := Reference(i, limit)
		got, s, _ := run(t, i, limit)
		if got != want || !reflect.DeepEqual(s.Heap, heap) {
			t.Fatalf("trial %d result %x != %x", trial, got, want)
		}
	}
}
func TestSnapshotIsolationValidationCancellation(t *testing.T) {
	i, _ := Example("shared")
	m := NewModel()
	ctx := context.Background()
	_, _ = m.Execute(ctx, Operation{Kind: "load", Image: &i})
	i.Nodes[0] = 0
	s, _ := m.Snapshot(ctx)
	s.Heap[0] = 0
	s2, _ := m.Snapshot(ctx)
	if s2.Heap[0] != Integer(21) {
		t.Fatal("heap alias")
	}
	bad := Image{[]Word{Node(Blackhole, 0, 0)}, 0}
	if _, e := m.Execute(ctx, Operation{Kind: "load", Image: &bad}); e == nil {
		t.Fatal("accepted external blackhole")
	}
	after, _ := m.Snapshot(ctx)
	if after.Heap[0] != Integer(21) {
		t.Fatal("invalid load mutated heap")
	}
	c, cancel := context.WithCancel(ctx)
	cancel()
	if _, e := m.Execute(c, Operation{Kind: "reset"}); e == nil {
		t.Fatal("ignored cancellation")
	}
	if e := m.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e := m.Snapshot(ctx); e == nil {
		t.Fatal("closed snapshot")
	}
}
func TestTracePrefixLoss(t *testing.T) {
	nodes := []Word{Integer(1)}
	for n := 1; n < 80; n++ {
		nodes = append(nodes, Node(Thunk, uint16(n-1), 0))
	}
	_, s, _ := run(t, Image{nodes, 79}, StackCapacity)
	if len(s.Trace) != 64 || s.Dropped != 158-64 || s.Counters.Claims != 79 || s.Counters.Updates != 79 {
		t.Fatalf("trace %+v", s.Counters)
	}
}
