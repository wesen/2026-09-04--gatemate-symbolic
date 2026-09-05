package dataflow

import (
	"context"
	"testing"
)

func TestModelIssueBreakpointAndTrace(t *testing.T) {
	ctx := context.Background()
	p, _ := Compile("input a: int16; let p=a*a; output p+p")
	m, _ := NewTransaction(DefaultConfig())
	if err := m.LoadGraph(p.Graph); err != nil {
		t.Fatal(err)
	}
	_, err := m.Execute(ctx, Operation{Kind: "debug", Debug: &DebugControl{Mask: 1, Node: 0, Context: 2, Clear: true}})
	if err != nil {
		t.Fatal(err)
	}
	tokens, _ := p.Tokens(2, 0, map[string]int64{"a": 7})
	for _, v := range tokens {
		m.Inject(v)
	}
	_, _ = m.Execute(ctx, Operation{Kind: "tick", Ticks: 100})
	s, _ := m.Snapshot(ctx)
	if !s.Debug.Halted || s.Debug.Reason != 1 || s.Counters["cycles"] >= 100 {
		t.Fatalf("not halted: %+v", s.Debug)
	}
	found := false
	for _, e := range s.Debug.Events {
		if e.Kind == EventIssue && e.Token.Node == 0 && e.Token.Value == 49 {
			found = true
		}
	}
	if !found {
		t.Fatal("issue event missing")
	}
	before := m.Metrics.Cycles
	_, _ = m.Execute(ctx, Operation{Kind: "tick", Ticks: 100})
	if m.Metrics.Cycles != before {
		t.Fatal("halted cycles advanced")
	}
	s.Debug.Events[0].Token.Value = 999
	again, _ := m.Snapshot(ctx)
	if again.Debug.Events[0].Token.Value == 999 {
		t.Fatal("debug snapshot aliases state")
	}
	_, _ = m.Execute(ctx, Operation{Kind: "debug", Debug: &DebugControl{Node: 255, Context: 255, Resume: true}})
	_, _ = m.Execute(ctx, Operation{Kind: "tick", Ticks: 100})
	result := m.Poll()
	if result == nil || result.Value != 98 {
		t.Fatalf("result %+v", result)
	}
	for n := 0; n < 40; n++ {
		if !m.Cancel(2) {
			t.Fatal("cancel rejected")
		}
	}
	if len(m.Debug.Events) != TraceCapacity || m.Debug.Dropped == 0 {
		t.Fatal("trace not bounded with explicit loss")
	}
	_, _ = m.Execute(ctx, Operation{Kind: "debug", Debug: &DebugControl{Node: 255, Context: 255, Clear: true}})
	if len(m.Debug.Events) != 0 || m.Debug.Dropped != 0 {
		t.Fatal("trace clear")
	}
}
func TestModelStaleBreakpoint(t *testing.T) {
	m, _ := NewTransaction(DefaultConfig())
	m.debugControl(DebugControl{Mask: 4, Node: 255, Context: 255})
	if !m.Cancel(0) {
		t.Fatal("cancel")
	}
	m.Inject(Source(0, 0, 0, 0, 7))
	_ = m.Tick(context.Background())
	if !m.Debug.Halted || m.Debug.Reason != 4 {
		t.Fatal("stale breakpoint")
	}
}
