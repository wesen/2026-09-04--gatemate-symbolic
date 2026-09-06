package semantics_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/check"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/semantics"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/syntax"
)

func program(t *testing.T, source string) *check.Program {
	t.Helper()
	ast, ds := syntax.Parse(context.Background(), source)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	p, ds := check.Check(context.Background(), ast)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	return p
}
func observe(t *testing.T, source string, prefix int) semantics.Observation {
	t.Helper()
	o, err := semantics.Observe(context.Background(), program(t, source), prefix, 100000)
	if err != nil {
		t.Fatal(err)
	}
	return o
}
func TestExamples(t *testing.T) {
	for _, tc := range []struct {
		name, kind string
		integer    int32
		fault      uint32
		values     []int32
	}{
		{"shared", "INT", 168, 0, nil}, {"closure", "INT", 42, 0, nil}, {"unused", "INT", 7, 0, nil},
		{"cycle", "ERROR", 0, 3, nil}, {"productive", "LIST", 0, 0, []int32{1, 1, 1, 1, 1, 1, 1, 1}},
		{"squares", "LIST", 0, 0, []int32{1, 4, 9, 16, 25, 36, 49, 64}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source, err := os.ReadFile("../../../examples/lazylang/" + tc.name + ".lazy")
			if err != nil {
				t.Fatal(err)
			}
			o := observe(t, string(source), 8)
			if o.Kind != tc.kind || o.Integer != tc.integer || o.Error != tc.fault || len(o.Values) > 0 && !reflect.DeepEqual(o.Values, tc.values) {
				t.Fatalf("%+v", o)
			}
			if tc.name == "shared" && (o.Stats.Muls != 1 || o.Stats.Adds != 3) {
				t.Fatal(o.Stats)
			}
			if tc.name == "squares" {
				if o.Complete {
					t.Fatal("forced ninth constructor")
				}
				if !observe(t, string(source), 9).Complete {
					t.Fatal("missing NIL")
				}
			}
		})
	}
}
func TestDemandAndFaults(t *testing.T) {
	for _, tc := range []struct {
		source  string
		integer int32
		fault   uint32
	}{
		{`def main : Int = let x : Int = 2147483647 + 1 in 7;`, 7, 0},
		{`def main : Int = if true then 7 else 2147483647 + 1;`, 7, 0},
		{`def main : Int = 2147483647 + 1;`, 0, 6},
		{`def main : Int = -2147483648 - 1;`, 0, 6},
		{`def main : Int = 100000 * 100000;`, 0, 6},
		{`def main : Int = let rec x : Int = x in x + (3 * 4);`, 0, 3},
		{`def main : Int = let x : Int = 4 in let x : Int = x + 1 in x;`, 5, 0},
		{`def f : Int -> Int = fun (n : Int) -> if n <= 0 then 7 else g (n - 1); def g : Int -> Int = fun (n : Int) -> f n; def main : Int = f 4;`, 7, 0},
	} {
		o := observe(t, tc.source, 0)
		if o.Integer != tc.integer || o.Error != tc.fault {
			t.Fatalf("%s: %+v", tc.source, o)
		}
		if tc.fault == 3 && o.Stats.Muls != 0 {
			t.Fatal("evaluated right operand after left fault")
		}
	}
	o := observe(t, `def main : ListInt = Cons(1, let rec xs : ListInt = xs in xs);`, 1)
	if !reflect.DeepEqual(o.Values, []int32{1}) || o.Error != 0 {
		t.Fatal(o)
	}
	o = observe(t, `def main : ListInt = Cons(1, let rec xs : ListInt = xs in xs);`, 2)
	if o.Error != 3 {
		t.Fatal(o)
	}
	p := program(t, `def main : Int = 1;`)
	if _, err := semantics.Observe(context.Background(), p, 0, 1); err == nil {
		t.Fatal("missing fuel error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := semantics.Observe(ctx, p, 0, 100); err != context.Canceled {
		t.Fatal(err)
	}
}
