package syntax

import (
	"context"
	"reflect"
	"testing"
)

func checkProgramSpans(t *testing.T, source string, p *Program) {
	t.Helper()
	check := func(s, parent Span) {
		t.Helper()
		if s.Start < parent.Start || s.End > parent.End || s.Start > s.End {
			t.Fatalf("invalid nested span %+v in %+v", s, parent)
		}
	}
	var typ func(*Type, Span)
	typ = func(n *Type, parent Span) {
		t.Helper()
		if n == nil {
			t.Fatal("nil type in complete definition")
		}
		check(n.Span, parent)
		if n.Kind == FunctionType {
			typ(n.Parameter, n.Span)
			typ(n.Result, n.Span)
		}
	}
	var expr func(Expr, Span)
	expr = func(e Expr, parent Span) {
		t.Helper()
		if e == nil {
			t.Fatal("nil expression")
		}
		s := e.SourceSpan()
		check(s, parent)
		switch n := e.(type) {
		case *IntExpr, *BoolExpr, *NilExpr:
		case *VarExpr:
			check(n.Name.Span, s)
			if source[n.Name.Span.Start:n.Name.Span.End] != n.Name.Name {
				t.Fatal("identifier slice")
			}
		case *LambdaExpr:
			check(n.Parameter.Span, s)
			check(n.Parameter.Name.Span, n.Parameter.Span)
			typ(n.Parameter.Annotation, n.Parameter.Span)
			expr(n.Body, s)
		case *ApplyExpr:
			expr(n.Function, s)
			expr(n.Argument, s)
		case *LetExpr:
			check(n.Name.Span, s)
			typ(n.Annotation, s)
			expr(n.Value, s)
			expr(n.Body, s)
		case *BinaryExpr:
			expr(n.Left, s)
			expr(n.Right, s)
		case *IfExpr:
			expr(n.Condition, s)
			expr(n.Then, s)
			expr(n.Else, s)
		case *ConsExpr:
			expr(n.Head, s)
			expr(n.Tail, s)
		case *CaseExpr:
			expr(n.Scrutinee, s)
			expr(n.NilBranch, s)
			check(n.Head.Span, s)
			check(n.Tail.Span, s)
			expr(n.ConsBranch, s)
		case *GroupExpr:
			expr(n.Inner, s)
		default:
			t.Fatalf("unknown node %T", e)
		}
	}
	check(p.Span, Span{0, len(source)})
	for _, d := range p.Definitions {
		check(d.Span, p.Span)
		check(d.Name.Span, d.Span)
		typ(d.Annotation, d.Span)
		expr(d.Value, d.Span)
	}
}
func FuzzParse(f *testing.F) {
	for _, s := range []string{"", "def main : Int = f x + y * 2;", "// λ🙂\ndef main : Int = -2147483648;", "def x : Int = case xs of { Nil -> ; Cons(h,t) -> h }; def good : Int = 7;", "def main : ListInt = let rec xs : ListInt = Cons(1,xs) in xs;", "\xff\x00"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, source string) {
		p, ds := Parse(context.Background(), source)
		again, other := Parse(context.Background(), source)
		if !reflect.DeepEqual(p, again) || !reflect.DeepEqual(ds, other) {
			t.Fatal("nondeterministic parse")
		}
		if len(ds) > MaxDiagnostics {
			t.Fatal("unbounded diagnostics")
		}
		last := -1
		for _, d := range ds {
			if d.Span.Start < 0 || d.Span.End > len(source) || d.Span.Start > d.Span.End || d.Span.Start < last {
				t.Fatalf("bad diagnostic %+v", d)
			}
			last = d.Span.Start
		}
		checkProgramSpans(t, source, p)
		tokens, _ := Lex(context.Background(), source)
		if len(tokens) > MaxTokens+1 || tokens[len(tokens)-1].Kind != EOF {
			t.Fatal("bad token bound/EOF")
		}
		previous := 0
		for _, tok := range tokens {
			if tok.Span.Start < previous || tok.Span.End > len(source) || tok.Span.End < tok.Span.Start {
				t.Fatal(tok)
			}
			if tok.Text != source[tok.Span.Start:tok.Span.End] {
				t.Fatal("token slice")
			}
			previous = tok.Span.End
		}
	})
}
