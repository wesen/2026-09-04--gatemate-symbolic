// Package compile lowers checked source into deterministic LFL1 artifacts.
package compile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/check"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/ir"
	s "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/syntax"
)

const Version = "lfl1-go-1"
const Profile = "LFL1-2048-512-64-v1"

type Artifact struct {
	ID          string          `json:"id"`
	Version     string          `json:"version"`
	Profile     string          `json:"profile"`
	Source      string          `json:"source"`
	EntryType   string          `json:"entryType"`
	Root        uint16          `json:"root"`
	ConstantEnd uint16          `json:"constantEnd"`
	Code        []string        `json:"code"`
	Heap        []string        `json:"heap"`
	Provenance  []uint16        `json:"provenance"`
	Spans       []s.Span        `json:"spans"`
	Bindings    []check.Binding `json:"bindings"`
	Uses        []check.Use     `json:"uses"`
	CodeTypes   []string        `json:"codeTypes"`
}

// Digest hashes compact encoding/json of this fixed struct with ID empty. Arrays
// are ordered, no maps participate, and UTF-8 source is preserved in Source.
func (a Artifact) Digest() string {
	a.ID = ""
	b, _ := json.Marshal(a)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

type builder struct {
	p         *check.Program
	a         Artifact
	spans     map[s.Span]uint16
	constants map[int32]uint16
}

func (b *builder) span(x s.Span) uint16 {
	if id, ok := b.spans[x]; ok {
		return id
	}
	id := uint16(len(b.a.Spans))
	b.a.Spans = append(b.a.Spans, x)
	b.spans[x] = id
	return id
}
func (b *builder) object(o ir.Object, span uint16) uint16 {
	id := uint16(len(b.a.Heap))
	b.a.Heap = append(b.a.Heap, o.Hex())
	b.a.Provenance = append(b.a.Provenance, span)
	return id
}
func (b *builder) expr(e s.Expr) uint16 {
	c := ir.Code{Span: b.span(e.SourceSpan())}
	switch n := e.(type) {
	case *s.GroupExpr:
		return b.expr(n.Inner)
	case *s.IntExpr:
		c.Op = ir.Const
		ref, ok := b.constants[n.Value]
		if !ok {
			ref = b.object(ir.Object{Tag: ir.Int, Payload: uint32(n.Value)}, c.Span)
			b.constants[n.Value] = ref
		}
		c.A = ref
	case *s.BoolExpr:
		c.Op = ir.Const
		c.A = 10
		if n.Value {
			c.A = 11
		}
	case *s.NilExpr:
		c.Op = ir.Const
		c.A = 12
	case *s.VarExpr:
		c.Op = ir.Var
		c.Immediate = uint32(b.p.Uses[n].Depth)
	case *s.LambdaExpr:
		c.Op = ir.Lambda
		c.A = b.expr(n.Body)
	case *s.ApplyExpr:
		c.Op = ir.App
		c.A = b.expr(n.Function)
		c.B = b.expr(n.Argument)
	case *s.LetExpr:
		c.Op = ir.Let
		if n.Recursive {
			c.Op = ir.Letrec
		}
		c.A = b.expr(n.Value)
		c.B = b.expr(n.Body)
	case *s.BinaryExpr:
		c.Op = ir.Prim
		c.A = b.expr(n.Left)
		c.B = b.expr(n.Right)
		c.Immediate = uint32(map[s.TokenKind]ir.Primitive{s.Plus: ir.Add, s.Minus: ir.Sub, s.Star: ir.Mul, s.Equal: ir.EQ, s.LessEqual: ir.LE}[n.Operator])
	case *s.IfExpr:
		c.Op = ir.If
		c.A = b.expr(n.Condition)
		c.B = b.expr(n.Then)
		c.C = b.expr(n.Else)
	case *s.ConsExpr:
		c.Op = ir.MakeCons
		c.A = b.expr(n.Head)
		c.B = b.expr(n.Tail)
	case *s.CaseExpr:
		c.Op = ir.Case
		c.A = b.expr(n.Scrutinee)
		c.B = b.expr(n.NilBranch)
		c.C = b.expr(n.ConsBranch)
	}
	id := uint16(len(b.a.Code))
	b.a.Code = append(b.a.Code, c.Hex())
	b.a.CodeTypes = append(b.a.CodeTypes, check.TypeName(b.p.Types[e]))
	return id
}
func Compile(ctx context.Context, source string) (*Artifact, []s.Diagnostic) {
	ast, ds := s.Parse(ctx, source)
	if len(ds) > 0 {
		return nil, ds
	}
	p, ds := check.Check(ctx, ast)
	if len(ds) > 0 {
		return nil, ds
	}
	b := builder{p: p, spans: map[s.Span]uint16{}, constants: map[int32]uint16{}, a: Artifact{Version: Version, Profile: Profile, Source: source, Spans: []s.Span{{}}, Bindings: p.Bindings, Uses: []check.Use{}}}
	for i := uint32(1); i <= 10; i++ {
		b.object(ir.Object{Tag: ir.Error, Payload: i}, 0)
	}
	b.object(ir.Object{Tag: ir.Bool}, 0)
	b.object(ir.Object{Tag: ir.Bool, Payload: 1}, 0)
	b.object(ir.Object{Tag: ir.Nil}, 0)
	bodies := make([]uint16, len(ast.Definitions))
	for i, d := range ast.Definitions {
		bodies[i] = b.expr(d.Value)
	}
	b.a.ConstantEnd = uint16(len(b.a.Heap))
	head := ir.Absent
	// Top-level slots alternate THUNK, ENV in source order. All thunks capture
	// the final environment, supporting mutual recursion with fixed addresses.
	final := uint16(len(b.a.Heap) + 2*len(bodies) - 1)
	for i, d := range ast.Definitions {
		ref := b.object(ir.Object{Tag: ir.Thunk, A: bodies[i], B: final}, b.span(d.Value.SourceSpan()))
		head = b.object(ir.Object{Tag: ir.Env, A: ref, B: head}, b.span(d.Name.Span))
		if p.Declarations[d.Name.Span] == p.Main {
			b.a.Root = ref
			b.a.EntryType = check.TypeName(d.Annotation)
		}
	}
	for _, use := range p.Uses {
		b.a.Uses = append(b.a.Uses, use)
	}
	sort.Slice(b.a.Uses, func(i, j int) bool { return b.a.Uses[i].Span.Start < b.a.Uses[j].Span.Start })
	b.a.ID = b.a.Digest()
	if err := b.a.Validate(); err != nil {
		return nil, []s.Diagnostic{{Code: "ARTIFACT_LIMIT", Message: err.Error(), Span: ast.Span}}
	}
	if err := ctx.Err(); err != nil {
		return nil, []s.Diagnostic{{Code: "CANCELED", Message: err.Error(), Span: ast.Span}}
	}
	return &b.a, []s.Diagnostic{}
}

// Validate rejects malformed external artifacts before a runtime is mutated.
func (a Artifact) Validate() error {
	if a.Version != Version || a.Profile != Profile {
		return errors.New("unsupported compiler or profile")
	}
	if a.ID != a.Digest() {
		return errors.New("artifact digest mismatch")
	}
	if len(a.Code) == 0 || len(a.Code) > ir.Capacity || len(a.Heap) > ir.Capacity || len(a.Heap) < 13 || len(a.Provenance) != len(a.Heap) || len(a.CodeTypes) != len(a.Code) {
		return errors.New("invalid code or heap dimensions")
	}
	if a.ConstantEnd < 13 || int(a.ConstantEnd) > len(a.Heap) || int(a.Root) >= len(a.Heap) || len(a.Spans) == 0 || len(a.Spans) > 65535 || len(a.Source) > s.MaxSourceBytes {
		return errors.New("invalid artifact bounds")
	}
	for _, span := range a.Spans {
		if span.Start < 0 || span.End < span.Start || span.End > len(a.Source) {
			return errors.New("invalid source span")
		}
	}
	heap := make([]ir.Object, len(a.Heap))
	for i, packed := range a.Heap {
		o, err := ir.ParseObject(packed)
		if err != nil {
			return err
		}
		heap[i] = o
	}
	for i, o := range heap {
		if o.Flags != 0 || int(a.Provenance[i]) >= len(a.Spans) {
			return fmt.Errorf("invalid heap flags/provenance at %d", i)
		}
		if i < 10 {
			if o != (ir.Object{Tag: ir.Error, Payload: uint32(i + 1)}) {
				return errors.New("invalid pinned error")
			}
			continue
		}
		if i == 10 || i == 11 {
			if o != (ir.Object{Tag: ir.Bool, Payload: uint32(i - 10)}) {
				return errors.New("invalid pinned boolean")
			}
			continue
		}
		if i == 12 {
			if o != (ir.Object{Tag: ir.Nil}) {
				return errors.New("invalid pinned NIL")
			}
			continue
		}
		if i < int(a.ConstantEnd) {
			if o.Tag != ir.Int || o.A != 0 || o.B != 0 {
				return errors.New("invalid constant")
			}
			continue
		}
		switch o.Tag {
		case ir.Thunk:
			if int(o.A) >= len(a.Code) || int(o.B) >= len(heap) || heap[o.B].Tag != ir.Env || o.Payload != 0 {
				return errors.New("invalid initial thunk")
			}
		case ir.Env:
			if int(o.A) >= len(heap) || o.B != ir.Absent && (int(o.B) >= i || heap[o.B].Tag != ir.Env) || o.Payload != 0 {
				return errors.New("invalid initial ENV")
			}
		default:
			return errors.New("initial mutable heap must contain only THUNK and ENV")
		}
	}
	if heap[a.Root].Tag != ir.Thunk {
		return errors.New("root must be a thunk")
	}
	for i, packed := range a.Code {
		c, err := ir.ParseCode(packed)
		if err != nil {
			return err
		}
		if c.Flags != 0 || c.Reserved != 0 || int(c.Span) >= len(a.Spans) {
			return errors.New("invalid code flags/span")
		}
		child := func(x uint16) bool { return int(x) < i }
		valid := false
		switch c.Op {
		case ir.Const:
			valid = c.A < a.ConstantEnd && c.B == 0 && c.C == 0 && c.Immediate == 0
		case ir.Var:
			valid = c.A == 0 && c.B == 0 && c.C == 0 && c.Immediate < ir.Capacity
		case ir.Lambda:
			valid = child(c.A) && c.B == 0 && c.C == 0 && c.Immediate == 0
		case ir.App, ir.Let, ir.Letrec, ir.MakeCons:
			valid = child(c.A) && child(c.B) && c.C == 0 && c.Immediate == 0
		case ir.Prim:
			valid = child(c.A) && child(c.B) && c.C == 0 && c.Immediate <= uint32(ir.LE)
		case ir.If, ir.Case:
			valid = child(c.A) && child(c.B) && child(c.C) && c.Immediate == 0
		}
		if !valid {
			return fmt.Errorf("invalid code operands at %d", i)
		}
	}
	return nil
}
