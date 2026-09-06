package machine

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/check"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/compile"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/ir"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/semantics"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/syntax"
)

func build(t *testing.T, source string) *Machine {
	t.Helper()
	a, ds := compile.Compile(context.Background(), source)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	m, err := New(a)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func demand(t *testing.T, m *Machine, ref uint16) uint16 {
	t.Helper()
	if err := m.Demand(ref); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100000; i++ {
		if err := m.Tick(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
		if m.valid {
			r, ok := m.Poll()
			if !ok {
				t.Fatal("lost output")
			}
			return r
		}
	}
	t.Fatalf("timed out state %d stack %d top %d", m.state, len(m.stack), m.top)
	return 0
}
func TestDifferentialExamples(t *testing.T) {
	for _, name := range []string{"shared", "closure", "unused", "cycle", "productive", "squares"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("../../../examples/lazylang/" + name + ".lazy")
			if err != nil {
				t.Fatal(err)
			}
			ast, ds := syntax.Parse(context.Background(), string(source))
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			p, ds := check.Check(context.Background(), ast)
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			want, err := semantics.Observe(context.Background(), p, 8, 100000)
			if err != nil {
				t.Fatal(err)
			}
			m := build(t, string(source))
			r := demand(t, m, m.root)
			o := m.heap[r]
			switch want.Kind {
			case "INT":
				if o.Tag != ir.Int || int32(o.Payload) != want.Integer {
					t.Fatal(o, want)
				}
			case "ERROR":
				if o.Tag != ir.Error || o.Payload != want.Error {
					t.Fatal(o, want)
				}
			case "LIST":
				values := []int32{}
				for i := 0; i < 8; i++ {
					if o.Tag != ir.Cons {
						t.Fatal(o)
					}
					h := m.heap[demand(t, m, o.A)]
					if h.Tag != ir.Int {
						t.Fatal(h)
					}
					values = append(values, int32(h.Payload))
					if i < 7 {
						o = m.heap[demand(t, m, o.B)]
					}
				}
				if !reflect.DeepEqual(values, want.Values) {
					t.Fatal(values, want.Values)
				}
			default:
				t.Fatal(want)
			}
			got := semantics.Stats{Claims: m.counters.Claims, Updates: m.counters.Updates, Adds: m.counters.Adds, Subs: m.counters.Subtracts, Muls: m.counters.Multiplies, EQs: m.counters.Equalities, LEs: m.counters.ComparisonsLE}
			if got != want.Stats {
				t.Fatalf("semantic counters %+v want %+v", got, want.Stats)
			}
			if m.counters.Claims != m.counters.Updates {
				t.Fatal("outstanding update")
			}
			for i, o := range m.heap[:m.top] {
				if o.Tag == ir.Blackhole {
					t.Fatalf("blackhole left at %d", i)
				}
			}
		})
	}
}
func TestSharedFixtureAndStableOutput(t *testing.T) {
	m := build(t, `def main : Int = let x : Int = (fun (n : Int) -> n * 2) 21 in (x + x) + (x + x);`)
	if err := m.Demand(m.root); err != nil {
		t.Fatal(err)
	}
	for !m.valid {
		if err := m.Tick(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	}
	if m.result != 25 || m.heap[m.result].Payload != 168 || m.counters.Allocations != 9 || m.counters.Claims != 3 || m.counters.Updates != 3 || m.counters.Multiplies != 1 || m.counters.Adds != 3 {
		t.Fatalf("%+v", m.Snapshot())
	}
	before := m.Snapshot()
	if err := m.Demand(m.root); err == nil {
		t.Fatal("demand accepted with output pending")
	}
	if err := m.Tick(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	after := m.Snapshot()
	if after.ResultRef != before.ResultRef || !after.Valid || !reflect.DeepEqual(after.Heap, before.Heap) || after.Counters.OutputStalls != 7 {
		t.Fatal("unstable output")
	}
	if _, ok := m.Poll(); !ok {
		t.Fatal("missing output")
	}
	claims := m.counters.Claims
	r := demand(t, m, m.root)
	if m.counters.Claims != claims || r != 25 {
		t.Fatal("recomputed shared result")
	}
	before.Heap[0] = "changed"
	before.Trace[0].Kind = 99
	if m.Snapshot().Heap[0] == "changed" || m.trace[0].Kind == 99 {
		t.Fatal("snapshot aliases machine")
	}
}
func TestAllocationPauseAndPublication(t *testing.T) {
	m := build(t, `def main : ListInt = Cons(1, Nil);`)
	if err := m.Demand(m.root); err != nil {
		t.Fatal(err)
	}
	for m.state != Allocate {
		if err := m.Tick(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	}
	base := m.top
	for m.alloc.Stage < 2 {
		if m.top != base {
			t.Fatal("early publication")
		}
		if err := m.Demand(m.root); err == nil {
			t.Fatal("demand accepted mid-allocation")
		}
		if err := m.Tick(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	}
	if m.top != base+3 {
		t.Fatal("batch not committed")
	}
	for _, span := range m.provenance[base:m.top] {
		if span == 0 {
			t.Fatal("published without provenance")
		}
	}
	for !m.valid {
		if err := m.Tick(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	}
	for i := base; i < m.top; i++ {
		if m.heap[i].Tag == ir.Free {
			t.Fatal("published FREE")
		}
	}
}
func TestFaultUnwindAndLimits(t *testing.T) {
	for _, tc := range []struct {
		source string
		code   uint32
	}{
		{`def main : Int = let rec x : Int = x in x + (3 * 4);`, 3},
		{`def main : Int = let x : Int = 2147483647 + 1 in x;`, 6},
		{`def f : Int -> Int = fun (x : Int) -> 1 + f x; def main : Int = f 0;`, 2},
	} {
		m := build(t, tc.source)
		r := demand(t, m, m.root)
		if m.heap[r].Tag != ir.Error || m.heap[r].Payload != tc.code {
			t.Fatal(m.heap[r], tc.code)
		}
		if m.counters.Claims != m.counters.Updates {
			t.Fatal("failed to unwind")
		}
	}
	m := build(t, `def main : ListInt = let rec xs : ListInt = Cons(1, xs) in xs;`)
	// Force allocation failure with the immutable initial program intact.
	m.top = ir.Capacity - 1
	r := demand(t, m, m.root)
	if m.heap[r].Payload != 8 || m.counters.Allocations != 0 || m.counters.Claims != m.counters.Updates {
		t.Fatal(m.heap[r], m.counters)
	}
	m = build(t, `def main : Int = 1;`)
	r = demand(t, m, ir.Absent)
	if m.heap[r].Payload != 1 {
		t.Fatal(m.heap[r])
	}
	if err := m.Tick(context.Background(), 1000001); err == nil {
		t.Fatal("unbounded ticks")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := m.Tick(ctx, 1); err != context.Canceled {
		t.Fatal(err)
	}
}
func TestTracePrefixNeverStalls(t *testing.T) {
	source, err := os.ReadFile("../../../examples/lazylang/squares.lazy")
	if err != nil {
		t.Fatal(err)
	}
	m := build(t, string(source))
	r := demand(t, m, m.root)
	for i := 0; i < 8; i++ {
		o := m.heap[r]
		demand(t, m, o.A)
		if i < 7 {
			r = demand(t, m, o.B)
		}
	}
	if len(m.trace) != 64 || m.counters.TraceDropped == 0 {
		t.Fatal(len(m.trace), m.counters.TraceDropped)
	}
	first := m.trace[0]
	demand(t, m, m.root)
	if m.trace[0] != first {
		t.Fatal("overwrote prefix")
	}
}

func TestTickPartitionAndFaultDefenses(t *testing.T) {
	src := `def main : Int = let x : Int = 6 * 7 in x + x;`
	a, b := build(t, src), build(t, src)
	if err := a.Demand(a.root); err != nil {
		t.Fatal(err)
	}
	if err := b.Demand(b.root); err != nil {
		t.Fatal(err)
	}
	if err := a.Tick(context.Background(), 1000); err != nil {
		t.Fatal(err)
	}
	for _, n := range []uint32{1, 3, 16, 37, 243, 700} {
		if err := b.Tick(context.Background(), n); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(a.Snapshot(), b.Snapshot()) {
		t.Fatal("tick partition changes machine")
	}
	for _, tc := range []struct {
		name  string
		setup func(*Machine)
		fault uint32
	}{
		{"IND cycle", func(m *Machine) { m.heap[m.root] = ir.Object{Tag: ir.Ind, A: m.root} }, 4},
		{"bad ENV", func(m *Machine) {
			m.code = []ir.Code{{Op: ir.Var}}
			m.heap[m.root] = ir.Object{Tag: ir.Thunk, A: 0, B: ir.Absent}
		}, 10},
		{"bad code", func(m *Machine) { m.heap[m.root] = ir.Object{Tag: ir.Thunk, A: ir.Absent, B: ir.Absent} }, 9},
		{"type", func(m *Machine) { m.heap[m.root] = ir.Object{Tag: ir.Env} }, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := build(t, `def main : Int = 1;`)
			tc.setup(m)
			r := demand(t, m, m.root)
			if m.heap[r].Tag != ir.Error || m.heap[r].Payload != tc.fault {
				t.Fatal(m.heap[r])
			}
		})
	}
}
