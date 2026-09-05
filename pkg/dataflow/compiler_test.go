package dataflow

import (
	"context"
	"strings"
	"testing"
)

func runProgram(t *testing.T, p Program, values map[string]int64, want int32) {
	t.Helper()
	m, _ := NewTransaction(Config{1, 1, 1, 4, 8})
	if err := m.LoadGraph(p.Graph); err != nil {
		t.Fatal(err)
	}
	semantic := &Semantic{}
	if err := semantic.LoadGraph(p.Graph); err != nil {
		t.Fatal(err)
	}
	var immediate []Token
	for c := byte(0); c < Contexts; c++ {
		tokens, err := p.Tokens(c, 0, values)
		if err != nil {
			t.Fatal(err)
		}
		for _, token := range tokens {
			immediate = append(immediate, semantic.Deliver(token)...)
			for !m.Inject(token) {
				if err := m.Tick(context.Background()); err != nil {
					t.Fatal(err)
				}
			}
		}
		// Drain each context with tiny queues, then reuse shared units for the next.
		var result *Token
		for n := 0; n < 500; n++ {
			_ = m.Tick(context.Background())
			if result = m.Poll(); result != nil {
				break
			}
		}
		if result == nil || result.Value != Int(want) || result.Context != c || !result.Final {
			t.Fatalf("context %d result %+v want %d", c, result, want)
		}
	}
	if len(immediate) != 4 {
		t.Fatalf("semantic outputs %v", immediate)
	}
	for _, v := range immediate {
		if v.Value != Int(want) || !v.Final {
			t.Fatalf("semantic result %+v", v)
		}
	}
}
func TestCompilerExecution(t *testing.T) {
	cases := []struct {
		source string
		values map[string]int64
		want   int32
	}{
		{"input a, b, c: int16\nlet square = a * a\nlet offset = b * c\nlet selected = int(a < b)\noutput square + offset + selected", map[string]int64{"a": 3, "b": 4, "c": 5}, 30},
		{"input a: int16; let p = a*a; output p+p+p", map[string]int64{"a": 7}, 147},
		{"input a: int16; let p = a*a; output p+p+p+p", map[string]int64{"a": -3}, 36},
		{"input a: int32; output a", map[string]int64{"a": -2147483648}, -2147483648},
		{"output 2+3*4-5", map[string]int64{}, 9},
		{"input flag: bool; output int(flag)+10", map[string]int64{"flag": 1}, 11},
		{"output -2147483648", map[string]int64{}, -2147483648},
	}
	for _, tc := range cases {
		t.Run(tc.source, func(t *testing.T) {
			p, err := Compile(tc.source)
			if err != nil {
				t.Fatal(err)
			}
			runProgram(t, p, tc.values, tc.want)
		})
	}
}
func TestCompilerDiagnostics(t *testing.T) {
	for _, source := range []string{"output missing", "input a: bool; output a+1", "input a: int32; output a*a", "input a: int16; let a=2; output a", "output 1; output 2", "let a=2", "output int(1)", "output 2147483648", "output 1+2+3+4+5+6+7+8+9", "input a: wat; output a", "output 1 < 2 < 3"} {
		if _, err := Compile(source); err == nil {
			t.Errorf("accepted %q", source)
		}
	}
	p, err := Compile("input a: int16; output a*a")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []map[string]int64{{}, {"a": 32768}, {"a": 1, "b": 2}} {
		if _, err := p.Tokens(0, 0, v); err == nil {
			t.Errorf("accepted inputs %v", v)
		}
	}
	if _, err := Compile("output " + strings.Repeat("(", 130) + "1" + strings.Repeat(")", 130)); err == nil {
		t.Fatal("accepted excessive nesting")
	}
}
func TestGraphValidationAndOwnership(t *testing.T) {
	p, err := Compile("input a: int16; let p=a*a; output p+p")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*Graph){
		func(g *Graph) { g.Count = 0 }, func(g *Graph) { g.Descriptors[0].Destinations[0].Node = 0 }, func(g *Graph) { g.Descriptors[0].Destinations[1] = g.Descriptors[0].Destinations[0] }, func(g *Graph) { g.Descriptors[0].Op = 7 }, func(g *Graph) { g.Descriptors[1].Final = false }, func(g *Graph) { g.Descriptors[6].Final = true },
	} {
		g := p.Graph
		mutate(&g)
		if g.Validate() == nil {
			t.Fatal("accepted malformed graph")
		}
	}
	m, _ := NewTransaction(DefaultConfig())
	if err := m.LoadGraph(p.Graph); err != nil {
		t.Fatal(err)
	}
	snap, _ := m.Snapshot(context.Background())
	snap.Graph.Descriptors[0].Op = Copy
	snap2, _ := m.Snapshot(context.Background())
	if snap2.Graph.Descriptors[0].Op != Mul {
		t.Fatal("snapshot aliases graph")
	}
	_ = m.Tick(context.Background())
	if m.LoadGraph(p.Graph) == nil {
		t.Fatal("loaded graph after tick")
	}
	_, _ = m.Execute(context.Background(), Operation{Kind: "reset"})
	if m.graph() != ResetGraph() {
		t.Fatal("reset graph missing")
	}
}
