package check_test

import (
	"context"
	"testing"

	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/check"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/syntax"
)

func TestDiagnostics(t *testing.T) {
	for _, tc := range []struct{ source, code string }{
		{`def x : Int = 1;`, "MISSING_MAIN"},
		{`def main : Int = x;`, "UNBOUND_NAME"},
		{`def main : Int = true;`, "TYPE_MISMATCH"},
		{`def main : Int = 1 2;`, "NOT_A_FUNCTION"},
		{`def main : Int = let x : Int = x in x;`, "UNBOUND_NAME"},
		{`def main : Int = if 1 then 2 else 3;`, "TYPE_MISMATCH"},
		{`def main : Int = (fun (x : Int) -> x) true;`, "TYPE_MISMATCH"},
		{`def main : Int = case Nil of { Nil -> 0; Cons(x, x) -> 1 };`, "DUPLICATE_PATTERN"},
		{`def main : Int = 1; def main : Int = 2;`, "DUPLICATE_BINDING"},
	} {
		t.Run(tc.code+tc.source, func(t *testing.T) {
			ast, ds := syntax.Parse(context.Background(), tc.source)
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			p, ds := check.Check(context.Background(), ast)
			if p != nil {
				t.Fatal("invalid program accepted")
			}
			for _, d := range ds {
				if d.Code == tc.code {
					return
				}
			}
			t.Fatalf("want %s: %+v", tc.code, ds)
		})
	}
}

func TestBindingDepthAndShadowing(t *testing.T) {
	ast, ds := syntax.Parse(context.Background(), `def x : Int = 8; def main : Int = let x : Int = x in (fun (y : Int) -> x + y) 1;`)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	p, ds := check.Check(context.Background(), ast)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	counts := map[int]int{}
	depths := map[int]uint16{}
	for _, use := range p.Uses {
		counts[use.Binding]++
		depths[use.Binding] = use.Depth
	}
	for id, depth := range map[int]uint16{0: 1, 2: 1, 3: 0} {
		if counts[id] != 1 || depths[id] != depth {
			t.Fatalf("binding %d: counts=%v depths=%v", id, counts, depths)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, ds := check.Check(ctx, ast); len(ds) == 0 || ds[0].Code != "CANCELED" {
		t.Fatal(ds)
	}
}
