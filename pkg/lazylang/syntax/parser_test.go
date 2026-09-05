package syntax

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func parsed(t *testing.T, s string) *Program {
	t.Helper()
	p, ds := Parse(context.Background(), s)
	if len(ds) > 0 {
		t.Fatalf("%s: %+v", s, ds)
	}
	return p
}
func expression(t *testing.T, s string) Expr {
	t.Helper()
	return parsed(t, "def main : Int = "+s+";").Definitions[0].Value
}
func shape(e Expr) string {
	switch n := e.(type) {
	case *IntExpr:
		return fmt.Sprint(n.Value)
	case *VarExpr:
		return n.Name.Name
	case *ApplyExpr:
		return "app(" + shape(n.Function) + "," + shape(n.Argument) + ")"
	case *BinaryExpr:
		return string(n.Operator) + "(" + shape(n.Left) + "," + shape(n.Right) + ")"
	case *GroupExpr:
		return "group(" + shape(n.Inner) + ")"
	}
	return fmt.Sprintf("%T", e)
}
func TestPrecedenceAndApplication(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{"f x + y * 2", "+(app(f,x),*(y,2))"},
		{"f x y", "app(app(f,x),y)"},
		{"a - b - c", "-(-(a,b),c)"},
		{"a + b <= c * d", "<=(+(a,b),*(c,d))"},
		{"n-1", "-(n,1)"}, {"n - 1", "-(n,1)"}, {"n*-2", "*(n,-2)"},
		{"f (-1)", "app(f,group(-1))"}, {"(a <= b) == c", "==(group(<=(a,b)),c)"},
	} {
		t.Run(tc.source, func(t *testing.T) {
			if got := shape(expression(t, tc.source)); got != tc.want {
				t.Fatalf("%s != %s", got, tc.want)
			}
		})
	}
}
func TestTypesAndScopeSyntax(t *testing.T) {
	p := parsed(t, "def add : Int -> Int -> Int = fun (x : Int) -> fun (y : Int) -> x+y;")
	typ := p.Definitions[0].Annotation
	if typ.Kind != FunctionType || typ.Parameter.Kind != IntType || typ.Result.Kind != FunctionType {
		t.Fatal(typ)
	}
	p = parsed(t, "def map : (Int -> Int) -> ListInt -> ListInt = fun (f : Int -> Int) -> fun (xs : ListInt) -> case xs of { Nil -> Nil; Cons(h,t) -> Cons(f h, map f t) };")
	if p.Definitions[0].Annotation.Parameter.Kind != FunctionType {
		t.Fatal("parenthesized function type")
	}
	c := p.Definitions[0].Value.(*LambdaExpr).Body.(*LambdaExpr).Body.(*CaseExpr)
	if c.Head.Name != "h" || c.Tail.Name != "t" {
		t.Fatal(c)
	}
	cons := c.ConsBranch.(*ConsExpr)
	if shape(cons.Tail) != "app(app(map,f),t)" {
		t.Fatal(shape(cons.Tail))
	}
	e := expression(t, "let rec x : Int = x in if true then x else 0").(*LetExpr)
	if !e.Recursive || e.Body.(*IfExpr).Condition.(*BoolExpr).Value != true {
		t.Fatal(e)
	}
}
func TestSpansAndSignedRange(t *testing.T) {
	src := "// 🙂\r\ndef main : Int = (-2147483648);"
	p := parsed(t, src)
	d := p.Definitions[0]
	g := d.Value.(*GroupExpr)
	n := g.Inner.(*IntExpr)
	if n.Value != -2147483648 || src[n.Range.Start:n.Range.End] != "-2147483648" || src[g.Range.Start:g.Range.End] != "(-2147483648)" {
		t.Fatal(n, g)
	}
	if src[d.Name.Span.Start:d.Name.Span.End] != "main" || src[d.Annotation.Span.Start:d.Annotation.Span.End] != "Int" {
		t.Fatal(d)
	}
	if src[d.Span.Start:d.Span.End] != "def main : Int = (-2147483648);" {
		t.Fatal(d.Span)
	}
	for _, v := range []string{"2147483648", "-2147483649"} {
		_, ds := Parse(context.Background(), "def main : Int = "+v+";")
		if len(ds) != 1 || ds[0].Code != "INTEGER_RANGE" {
			t.Fatal(ds)
		}
	}
}
func TestDiagnosticsAndRecovery(t *testing.T) {
	for _, bad := range []string{
		"def bad : Int = ;", "def bad : Int = 1 < 2;", "def bad : Int = a <= b == c;",
		"def bad : Int = -(x);", "def bad : Int = Cons(1);", "def bad : Int = fun x -> x;",
		"def bad : Int = case x of { Nil -> ; Cons(h,t) -> h };",
		"def bad : Int = case x of { Nil -> 1; Cons(h,t) -> ;", // unmatched brace
		"def bad : Int = 1", // missing definition separator
	} {
		t.Run(bad, func(t *testing.T) {
			src := bad + " def good : Int = 7;"
			p, ds := Parse(context.Background(), src)
			if len(ds) == 0 || len(p.Definitions) != 1 || p.Definitions[0].Name.Name != "good" {
				t.Fatalf("defs=%+v diagnostics=%+v", p.Definitions, ds)
			}
			for _, d := range ds {
				if d.Span.Start < 0 || d.Span.End > len(src) || d.Span.Start > d.Span.End {
					t.Fatal(d)
				}
			}
		})
	}
	_, ds := Parse(context.Background(), "def main : Int = 1")
	if ds[0].Span != (Span{18, 18}) {
		t.Fatalf("EOF location: %+v", ds)
	}
}
func TestParserBoundsAndCancellation(t *testing.T) {
	for _, expr := range []string{strings.Repeat("(", MaxNesting+1) + "1" + strings.Repeat(")", MaxNesting+1), strings.Repeat("1+", MaxNesting) + "1"} {
		_, ds := Parse(context.Background(), "def main : Int = "+expr+";")
		if len(ds) == 0 || ds[0].Code != "NESTING_LIMIT" {
			t.Fatal(ds)
		}
	}
	_, ds := Parse(context.Background(), strings.Repeat("def x : Int = ;", 100))
	if len(ds) != MaxDiagnostics {
		t.Fatal(len(ds))
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, ds = Parse(ctx, "")
	if len(ds) != 1 || ds[0].Code != "CANCELED" {
		t.Fatal(ds)
	}
	_, ds = Parse(context.Background(), "// only a comment")
	if len(ds) != 1 || ds[0].Code != "EXPECTED_DEFINITION" {
		t.Fatal(ds)
	}
}
