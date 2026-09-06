// Package check resolves lexical bindings and checks the monomorphic source language.
package check

import (
	"context"
	"fmt"
	s "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/syntax"
	"sort"
)

type Binding struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Span s.Span `json:"span"`
	Type string `json:"type"`
}
type Use struct {
	Binding int    `json:"binding"`
	Depth   uint16 `json:"depth"`
	Span    s.Span `json:"span"`
}
type Program struct {
	Syntax       *s.Program
	Bindings     []Binding
	Uses         map[*s.VarExpr]Use
	Declarations map[s.Span]int
	Types        map[s.Expr]*s.Type
	Main         int
}

func TypeName(t *s.Type) string {
	if t == nil {
		return "<invalid>"
	}
	if t.Kind != s.FunctionType {
		return string(t.Kind)
	}
	left := TypeName(t.Parameter)
	if t.Parameter.Kind == s.FunctionType {
		left = "(" + left + ")"
	}
	return left + " -> " + TypeName(t.Result)
}
func Same(a, b *s.Type) bool {
	if a == nil || b == nil {
		return false
	}
	if a.Kind != b.Kind {
		return false
	}
	return a.Kind != s.FunctionType || Same(a.Parameter, b.Parameter) && Same(a.Result, b.Result)
}

type entry struct {
	id   int
	name string
	typ  *s.Type
}
type checker struct {
	ctx context.Context
	out *Program
	ds  []s.Diagnostic
}

func (c *checker) diagnose(code, message string, span s.Span) {
	if len(c.ds) < s.MaxDiagnostics {
		c.ds = append(c.ds, s.Diagnostic{Code: code, Message: message, Span: span})
	}
}
func (c *checker) bind(name s.Identifier, t *s.Type) entry {
	id := len(c.out.Bindings)
	c.out.Bindings = append(c.out.Bindings, Binding{id, name.Name, name.Span, TypeName(t)})
	c.out.Declarations[name.Span] = id
	return entry{id, name.Name, t}
}
func extend(env []entry, e entry) []entry { return append(append([]entry{}, env...), e) }
func (c *checker) expect(got, want *s.Type, span s.Span) {
	if got != nil && want != nil && !Same(got, want) {
		c.diagnose("TYPE_MISMATCH", fmt.Sprintf("expected %s, found %s", TypeName(want), TypeName(got)), span)
	}
}

// Check accepts parser-produced syntax. It returns no checked program on errors.
func Check(ctx context.Context, ast *s.Program) (*Program, []s.Diagnostic) {
	if ast == nil {
		return nil, []s.Diagnostic{{Code: "INVALID_PROGRAM", Message: "missing syntax tree"}}
	}
	c := &checker{ctx: ctx, out: &Program{Syntax: ast, Bindings: []Binding{}, Uses: map[*s.VarExpr]Use{}, Declarations: map[s.Span]int{}, Types: map[s.Expr]*s.Type{}, Main: -1}, ds: []s.Diagnostic{}}
	env := []entry{}
	seen := map[string]bool{}
	for _, d := range ast.Definitions {
		if seen[d.Name.Name] {
			c.diagnose("DUPLICATE_BINDING", "duplicate top-level binding "+d.Name.Name, d.Name.Span)
		}
		seen[d.Name.Name] = true
		e := c.bind(d.Name, d.Annotation)
		env = append(env, e)
		if d.Name.Name == "main" {
			c.out.Main = e.id
		}
	}
	if c.out.Main < 0 {
		c.diagnose("MISSING_MAIN", "program requires a main definition", ast.Span)
	}
	for _, d := range ast.Definitions {
		c.expect(c.expr(d.Value, env), d.Annotation, d.Value.SourceSpan())
	}
	sort.SliceStable(c.ds, func(i, j int) bool { return c.ds[i].Span.Start < c.ds[j].Span.Start })
	if len(c.ds) > 0 {
		return nil, c.ds
	}
	return c.out, c.ds
}
func (c *checker) expr(e s.Expr, env []entry) *s.Type {
	if err := c.ctx.Err(); err != nil {
		c.diagnose("CANCELED", err.Error(), e.SourceSpan())
		return nil
	}
	if len(c.ds) >= s.MaxDiagnostics {
		return nil
	}
	integer := &s.Type{Kind: s.IntType}
	boolean := &s.Type{Kind: s.BoolType}
	list := &s.Type{Kind: s.ListIntType}
	var out *s.Type
	switch n := e.(type) {
	case *s.IntExpr:
		out = integer
	case *s.BoolExpr:
		out = boolean
	case *s.NilExpr:
		out = list
	case *s.GroupExpr:
		out = c.expr(n.Inner, env)
	case *s.VarExpr:
		for i := len(env) - 1; i >= 0; i-- {
			if env[i].name == n.Name.Name {
				depth := len(env) - 1 - i
				if depth >= 2048 {
					c.diagnose("ENVIRONMENT_LIMIT", "lexical depth exceeds profile", n.Name.Span)
					return nil
				}
				c.out.Uses[n] = Use{env[i].id, uint16(depth), n.Name.Span}
				out = env[i].typ
				break
			}
		}
		if out == nil {
			c.diagnose("UNBOUND_NAME", "unbound name "+n.Name.Name, n.Name.Span)
		}
	case *s.LambdaExpr:
		body := c.expr(n.Body, extend(env, c.bind(n.Parameter.Name, n.Parameter.Annotation)))
		if body != nil {
			out = &s.Type{Kind: s.FunctionType, Parameter: n.Parameter.Annotation, Result: body}
		}
	case *s.ApplyExpr:
		fn, arg := c.expr(n.Function, env), c.expr(n.Argument, env)
		if fn != nil {
			if fn.Kind != s.FunctionType {
				c.diagnose("NOT_A_FUNCTION", "cannot apply "+TypeName(fn), n.Function.SourceSpan())
			} else {
				c.expect(arg, fn.Parameter, n.Argument.SourceSpan())
				out = fn.Result
			}
		}
	case *s.LetExpr:
		binding := c.bind(n.Name, n.Annotation)
		scope := extend(env, binding)
		rhsEnv := env
		if n.Recursive {
			rhsEnv = scope
		}
		c.expect(c.expr(n.Value, rhsEnv), n.Annotation, n.Value.SourceSpan())
		out = c.expr(n.Body, scope)
	case *s.BinaryExpr:
		c.expect(c.expr(n.Left, env), integer, n.Left.SourceSpan())
		c.expect(c.expr(n.Right, env), integer, n.Right.SourceSpan())
		out = integer
		if n.Operator == s.Equal || n.Operator == s.LessEqual {
			out = boolean
		}
	case *s.IfExpr:
		c.expect(c.expr(n.Condition, env), boolean, n.Condition.SourceSpan())
		a, b := c.expr(n.Then, env), c.expr(n.Else, env)
		c.expect(b, a, n.Else.SourceSpan())
		out = a
	case *s.ConsExpr:
		c.expect(c.expr(n.Head, env), integer, n.Head.SourceSpan())
		c.expect(c.expr(n.Tail, env), list, n.Tail.SourceSpan())
		out = list
	case *s.CaseExpr:
		c.expect(c.expr(n.Scrutinee, env), list, n.Scrutinee.SourceSpan())
		a := c.expr(n.NilBranch, env)
		if n.Head.Name == n.Tail.Name {
			c.diagnose("DUPLICATE_PATTERN", "head and tail bindings must differ", n.Tail.Span)
		}
		scope := extend(env, c.bind(n.Head, integer))
		scope = extend(scope, c.bind(n.Tail, list))
		b := c.expr(n.ConsBranch, scope)
		c.expect(b, a, n.ConsBranch.SourceSpan())
		out = a
	default:
		c.diagnose("INVALID_EXPRESSION", "unsupported syntax node", e.SourceSpan())
	}
	if out != nil {
		c.out.Types[e] = out
	}
	return out
}
