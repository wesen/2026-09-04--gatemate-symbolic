// Package semantics evaluates checked source directly, independently of compiler
// records and the hardware-style allocated machine.
package semantics

import (
	"context"
	"github.com/pkg/errors"
	c "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/check"
	s "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/syntax"
	"math"
)

type Stats struct{ Claims, Updates, Adds, Subs, Muls, EQs, LEs uint32 }
type Observation struct {
	Kind     string  `json:"kind"`
	Integer  int32   `json:"integer"`
	Boolean  bool    `json:"boolean"`
	Values   []int32 `json:"values"`
	Complete bool    `json:"complete"`
	Error    uint32  `json:"error"`
	Stats    Stats   `json:"stats"`
}
type environment map[int]*cell
type cell struct {
	state byte
	expr  s.Expr
	env   environment
	value *value
}
type value struct {
	kind       string
	integer    int32
	boolean    bool
	code       uint32
	body       s.Expr
	env        environment
	parameter  int
	head, tail *cell
}
type evaluator struct {
	ctx   context.Context
	p     *c.Program
	fuel  uint64
	stats Stats
}

func clone(env environment) environment {
	out := environment{}
	for k, v := range env {
		out[k] = v
	}
	return out
}
func fault(code uint32) *value { return &value{kind: "ERROR", code: code} }
func (e *evaluator) step() error {
	if err := e.ctx.Err(); err != nil {
		return err
	}
	if e.fuel == 0 {
		return errors.New("reference fuel exhausted: observation inconclusive")
	}
	e.fuel--
	return nil
}
func (e *evaluator) force(x *cell) (*value, error) {
	if err := e.step(); err != nil {
		return nil, err
	}
	if x.state == 2 {
		return x.value, nil
	}
	if x.state == 1 {
		return fault(3), nil
	}
	x.state = 1
	e.stats.Claims++
	v, err := e.eval(x.expr, x.env)
	if err != nil {
		return nil, err
	}
	x.value = v
	x.state = 2
	e.stats.Updates++
	return v, nil
}
func (e *evaluator) eval(expr s.Expr, env environment) (*value, error) {
	if err := e.step(); err != nil {
		return nil, err
	}
	switch n := expr.(type) {
	case *s.IntExpr:
		return &value{kind: "INT", integer: n.Value}, nil
	case *s.BoolExpr:
		return &value{kind: "BOOL", boolean: n.Value}, nil
	case *s.NilExpr:
		return &value{kind: "NIL"}, nil
	case *s.GroupExpr:
		return e.eval(n.Inner, env)
	case *s.VarExpr:
		return e.force(env[e.p.Uses[n].Binding])
	case *s.LambdaExpr:
		return &value{kind: "FUN", body: n.Body, env: env, parameter: e.p.Declarations[n.Parameter.Name.Span]}, nil
	case *s.LetExpr:
		next := clone(env)
		x := &cell{expr: n.Value, env: env}
		next[e.p.Declarations[n.Name.Span]] = x
		if n.Recursive {
			x.env = next
		}
		return e.eval(n.Body, next)
	case *s.ApplyExpr:
		fn, err := e.eval(n.Function, env)
		if err != nil || fn.kind == "ERROR" {
			return fn, err
		}
		if fn.kind != "FUN" {
			return fault(5), nil
		}
		next := clone(fn.env)
		next[fn.parameter] = &cell{expr: n.Argument, env: env}
		return e.eval(fn.body, next)
	case *s.ConsExpr:
		return &value{kind: "CONS", head: &cell{expr: n.Head, env: env}, tail: &cell{expr: n.Tail, env: env}}, nil
	case *s.IfExpr:
		condition, err := e.eval(n.Condition, env)
		if err != nil || condition.kind == "ERROR" {
			return condition, err
		}
		if condition.kind != "BOOL" {
			return fault(5), nil
		}
		if condition.boolean {
			return e.eval(n.Then, env)
		}
		return e.eval(n.Else, env)
	case *s.CaseExpr:
		list, err := e.eval(n.Scrutinee, env)
		if err != nil || list.kind == "ERROR" {
			return list, err
		}
		if list.kind == "NIL" {
			return e.eval(n.NilBranch, env)
		}
		if list.kind != "CONS" {
			return fault(5), nil
		}
		next := clone(env)
		next[e.p.Declarations[n.Head.Span]] = list.head
		next[e.p.Declarations[n.Tail.Span]] = list.tail
		return e.eval(n.ConsBranch, next)
	case *s.BinaryExpr:
		left, err := e.eval(n.Left, env)
		if err != nil || left.kind == "ERROR" {
			return left, err
		}
		right, err := e.eval(n.Right, env)
		if err != nil || right.kind == "ERROR" {
			return right, err
		}
		if left.kind != "INT" || right.kind != "INT" {
			return fault(5), nil
		}
		a, b := int64(left.integer), int64(right.integer)
		var z int64
		switch n.Operator {
		case s.Plus:
			e.stats.Adds++
			z = a + b
		case s.Minus:
			e.stats.Subs++
			z = a - b
		case s.Star:
			e.stats.Muls++
			z = a * b
		case s.Equal:
			e.stats.EQs++
			return &value{kind: "BOOL", boolean: a == b}, nil
		case s.LessEqual:
			e.stats.LEs++
			return &value{kind: "BOOL", boolean: a <= b}, nil
		}
		if z < math.MinInt32 || z > math.MaxInt32 {
			return fault(6), nil
		}
		return &value{kind: "INT", integer: int32(z)}, nil
	}
	return nil, errors.New("unsupported reference expression")
}

// Observe demands main and at most prefix list heads. It does not force the tail
// after the final requested head. Fuel/cancellation are Go errors, not memoized faults.
func Observe(ctx context.Context, p *c.Program, prefix int, fuel uint64) (Observation, error) {
	out := Observation{Values: []int32{}}
	if p == nil || prefix < 0 || prefix > 2048 || fuel == 0 || fuel > 1000000 {
		return out, errors.New("invalid reference observation arguments")
	}
	e := &evaluator{ctx: ctx, p: p, fuel: fuel}
	env := environment{}
	for _, d := range p.Syntax.Definitions {
		env[p.Declarations[d.Name.Span]] = &cell{expr: d.Value, env: env}
	}
	v, err := e.force(env[p.Main])
	if err != nil {
		return out, err
	}
	out.Kind = v.kind
	if v.kind == "INT" {
		out.Integer = v.integer
	}
	if v.kind == "BOOL" {
		out.Boolean = v.boolean
	}
	if v.kind == "ERROR" {
		out.Error = v.code
	}
	if v.kind == "CONS" || v.kind == "NIL" {
		out.Kind = "LIST"
		for i := 0; i < prefix; i++ {
			if v.kind == "NIL" {
				out.Complete = true
				break
			}
			if v.kind == "ERROR" {
				out.Error = v.code
				break
			}
			head, err := e.force(v.head)
			if err != nil {
				return out, err
			}
			if head.kind == "ERROR" {
				out.Error = head.code
				break
			}
			out.Values = append(out.Values, head.integer)
			if i+1 < prefix {
				v, err = e.force(v.tail)
				if err != nil {
					return out, err
				}
			}
		}
		if v.kind == "NIL" {
			out.Complete = true
		}
	}
	out.Stats = e.stats
	return out, nil
}
