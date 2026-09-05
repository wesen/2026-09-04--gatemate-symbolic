package syntax

import (
	"context"
	"fmt"
	"sort"
	"strconv"
)

// Parse returns complete definitions and source-ordered diagnostics. Malformed
// definitions are omitted. Callers must not execute a program with diagnostics.
// Names, main's existence, and type correctness are deliberately not checked here.
func Parse(ctx context.Context, source string) (*Program, []Diagnostic) {
	tokens, diagnostics := Lex(ctx, source)
	program := &Program{Definitions: []Definition{}, Span: Span{0, len(source)}}
	if err := ctx.Err(); err != nil {
		if len(diagnostics) == 0 {
			diagnostics = append(diagnostics, Diagnostic{"CANCELED", err.Error(), Span{0, 0}})
		}
		return program, diagnostics
	}
	// A truncated token stream cannot be interpreted as a complete source file.
	for _, d := range diagnostics {
		if d.Code == "SOURCE_LIMIT" || d.Code == "TOKEN_LIMIT" {
			return program, diagnostics
		}
	}
	p := parser{ctx: ctx, tokens: tokens, heights: map[Expr]int{}}
	for p.peek().Kind != EOF && len(diagnostics) < MaxDiagnostics {
		start := p.i
		p.problem = nil
		def := p.definition()
		if p.problem != nil {
			diagnostics = append(diagnostics, *p.problem)
			if p.problem.Code == "CANCELED" {
				break
			}
			p.recoverDefinition(start)
		} else {
			program.Definitions = append(program.Definitions, *def)
		}
	}
	if len(program.Definitions) == 0 && len(diagnostics) == 0 {
		diagnostics = append(diagnostics, Diagnostic{"EXPECTED_DEFINITION", "expected at least one definition", p.peek().Span})
	}
	sort.SliceStable(diagnostics, func(i, j int) bool { return diagnostics[i].Span.Start < diagnostics[j].Span.Start })
	return program, diagnostics
}

type parser struct {
	ctx      context.Context
	tokens   []Token
	i, depth int
	problem  *Diagnostic
	heights  map[Expr]int
}

func (p *parser) peek() Token { return p.tokens[p.i] }
func (p *parser) next() Token {
	t := p.peek()
	if t.Kind != EOF {
		p.i++
	}
	return t
}
func (p *parser) fail(code, message string, span Span) {
	if p.problem == nil {
		p.problem = &Diagnostic{code, message, span}
	}
}
func (p *parser) expect(kind TokenKind) Token {
	if p.problem != nil {
		return Token{}
	}
	t := p.peek()
	if t.Kind != kind {
		p.fail("EXPECTED_TOKEN", fmt.Sprintf("expected %s, found %s", kind, t.Kind), t.Span)
		return Token{}
	}
	return p.next()
}
func (p *parser) accept(kind TokenKind) bool {
	if p.problem == nil && p.peek().Kind == kind {
		p.next()
		return true
	}
	return false
}
func (p *parser) enter() bool {
	if p.problem != nil {
		return false
	}
	if err := p.ctx.Err(); err != nil {
		p.fail("CANCELED", err.Error(), p.peek().Span)
		return false
	}
	if p.depth == MaxNesting {
		p.fail("NESTING_LIMIT", "syntax nesting exceeds limit", p.peek().Span)
		return false
	}
	p.depth++
	return true
}
func (p *parser) finish(e Expr, children ...Expr) Expr {
	if p.problem != nil {
		return nil
	}
	height := 1
	for _, c := range children {
		if p.heights[c]+1 > height {
			height = p.heights[c] + 1
		}
	}
	if height > MaxNesting {
		p.fail("NESTING_LIMIT", "expression tree exceeds nesting limit", e.SourceSpan())
		return nil
	}
	p.heights[e] = height
	return e
}
func identifier(t Token) Identifier { return Identifier{t.Text, t.Span} }
func node(start, end int) Node      { return Node{Range: Span{start, end}} }
func (p *parser) definition() *Definition {
	start := p.expect(Def)
	name := p.expect(Ident)
	p.expect(Colon)
	annotation := p.typ()
	p.expect(Assign)
	value := p.expr(0)
	end := p.expect(Semicolon)
	if p.problem != nil {
		return nil
	}
	return &Definition{identifier(name), annotation, value, Span{start.Span.Start, end.Span.End}}
}

// Recovery scans from the definition's beginning, so a case-alternative
// semicolon is never mistaken for the terminating definition semicolon. A new
// def is also a recovery anchor, including after unmatched opening braces.
func (p *parser) recoverDefinition(start int) {
	braces := 0
	for j := start; j < len(p.tokens); j++ {
		t := p.tokens[j]
		if t.Kind == EOF || j > start && t.Kind == Def {
			p.i = j
			return
		}
		switch t.Kind {
		case LBrace:
			braces++
		case RBrace:
			if braces > 0 {
				braces--
			}
		case Semicolon:
			if braces == 0 {
				p.i = j + 1
				return
			}
		}
	}
}
func (p *parser) typ() *Type {
	if !p.enter() {
		return nil
	}
	defer func() { p.depth-- }()
	t := p.next()
	var left *Type
	switch t.Kind {
	case TInt, TBool, TListInt:
		left = &Type{Kind: TypeKind(t.Kind), Span: t.Span}
	case LParen:
		left = p.typ()
		end := p.expect(RParen)
		if p.problem != nil {
			return nil
		}
		left.Span = Span{t.Span.Start, end.Span.End}
	default:
		p.fail("EXPECTED_TYPE", "expected Int, Bool, ListInt, or a parenthesized type", t.Span)
		return nil
	}
	if p.accept(Arrow) {
		right := p.typ()
		if p.problem != nil {
			return nil
		}
		return &Type{Kind: FunctionType, Parameter: left, Result: right, Span: Span{left.Span.Start, right.Span.End}}
	}
	return left
}
func precedence(k TokenKind) int {
	switch k {
	case Equal, LessEqual:
		return 10
	case Plus, Minus:
		return 20
	case Star:
		return 30
	}
	return -1
}
func comparison(k TokenKind) bool { return k == Equal || k == LessEqual }
func argumentStart(k TokenKind) bool {
	switch k {
	case Ident, Integer, True, False, Nil, Cons, LParen:
		return true
	}
	return false
}
func (p *parser) expr(min int) Expr {
	if !p.enter() {
		return nil
	}
	defer func() { p.depth-- }()
	left := p.prefix()
	if left == nil {
		return nil
	}
	for p.problem == nil {
		if err := p.ctx.Err(); err != nil {
			p.fail("CANCELED", err.Error(), p.peek().Span)
			return nil
		}
		// Adjacency is left-associative application. Bare minus is always infix
		// here; a negative function argument must be parenthesized.
		if argumentStart(p.peek().Kind) && 40 >= min {
			right := p.expr(41)
			if right == nil {
				return nil
			}
			left = p.finish(&ApplyExpr{node(left.SourceSpan().Start, right.SourceSpan().End), left, right}, left, right)
			if left == nil {
				return nil
			}
			continue
		}
		op := p.peek()
		bp := precedence(op.Kind)
		if bp < min {
			break
		}
		if comparison(op.Kind) {
			if prev, ok := left.(*BinaryExpr); ok && comparison(prev.Operator) {
				p.fail("CHAINED_COMPARISON", "comparisons cannot chain; group the intended comparison explicitly", op.Span)
				return nil
			}
		}
		p.next()
		right := p.expr(bp + 1)
		if right == nil {
			return nil
		}
		left = p.finish(&BinaryExpr{node(left.SourceSpan().Start, right.SourceSpan().End), op.Kind, left, right}, left, right)
		if left == nil {
			return nil
		}
	}
	return left
}
func (p *parser) prefix() Expr {
	if p.problem != nil {
		return nil
	}
	t := p.peek()
	switch t.Kind {
	case Integer, Minus:
		p.next()
		text := t.Text
		end := t.Span.End
		if t.Kind == Minus {
			digits := p.expect(Integer)
			if p.problem != nil {
				return nil
			}
			text = "-" + digits.Text
			end = digits.Span.End
		}
		value, err := strconv.ParseInt(text, 10, 32)
		if err != nil {
			p.fail("INTEGER_RANGE", "integer literal must fit signed int32", Span{t.Span.Start, end})
			return nil
		}
		return p.finish(&IntExpr{node(t.Span.Start, end), int32(value)})
	case Ident:
		p.next()
		return p.finish(&VarExpr{Node: node(t.Span.Start, t.Span.End), Name: identifier(t)})
	case True, False:
		p.next()
		return p.finish(&BoolExpr{node(t.Span.Start, t.Span.End), t.Kind == True})
	case Nil:
		p.next()
		return p.finish(&NilExpr{node(t.Span.Start, t.Span.End)})
	case LParen:
		p.next()
		inner := p.expr(0)
		end := p.expect(RParen)
		if p.problem != nil {
			return nil
		}
		return p.finish(&GroupExpr{node(t.Span.Start, end.Span.End), inner}, inner)
	case Fun:
		return p.lambda()
	case Let:
		return p.let()
	case If:
		return p.conditional()
	case Cons:
		return p.cons()
	case Case:
		return p.caseExpr()
	default:
		p.fail("EXPECTED_EXPRESSION", fmt.Sprintf("expected expression, found %s", t.Kind), t.Span)
		return nil
	}
}
func (p *parser) lambda() Expr {
	start := p.expect(Fun)
	lp := p.expect(LParen)
	name := p.expect(Ident)
	p.expect(Colon)
	annotation := p.typ()
	rp := p.expect(RParen)
	p.expect(Arrow)
	body := p.expr(0)
	if p.problem != nil {
		return nil
	}
	param := Parameter{identifier(name), annotation, Span{lp.Span.Start, rp.Span.End}}
	return p.finish(&LambdaExpr{node(start.Span.Start, body.SourceSpan().End), param, body}, body)
}
func (p *parser) let() Expr {
	start := p.expect(Let)
	recursive := p.accept(Rec)
	name := p.expect(Ident)
	p.expect(Colon)
	annotation := p.typ()
	p.expect(Assign)
	value := p.expr(0)
	p.expect(In)
	body := p.expr(0)
	if p.problem != nil {
		return nil
	}
	return p.finish(&LetExpr{node(start.Span.Start, body.SourceSpan().End), recursive, identifier(name), annotation, value, body}, value, body)
}
func (p *parser) conditional() Expr {
	start := p.expect(If)
	condition := p.expr(0)
	p.expect(Then)
	yes := p.expr(0)
	p.expect(Else)
	no := p.expr(0)
	if p.problem != nil {
		return nil
	}
	return p.finish(&IfExpr{node(start.Span.Start, no.SourceSpan().End), condition, yes, no}, condition, yes, no)
}
func (p *parser) cons() Expr {
	start := p.expect(Cons)
	p.expect(LParen)
	head := p.expr(0)
	p.expect(Comma)
	tail := p.expr(0)
	end := p.expect(RParen)
	if p.problem != nil {
		return nil
	}
	return p.finish(&ConsExpr{node(start.Span.Start, end.Span.End), head, tail}, head, tail)
}
func (p *parser) caseExpr() Expr {
	start := p.expect(Case)
	value := p.expr(0)
	p.expect(Of)
	p.expect(LBrace)
	p.expect(Nil)
	p.expect(Arrow)
	empty := p.expr(0)
	p.expect(Semicolon)
	p.expect(Cons)
	p.expect(LParen)
	head := p.expect(Ident)
	p.expect(Comma)
	tail := p.expect(Ident)
	p.expect(RParen)
	p.expect(Arrow)
	nonempty := p.expr(0)
	end := p.expect(RBrace)
	if p.problem != nil {
		return nil
	}
	return p.finish(&CaseExpr{node(start.Span.Start, end.Span.End), value, empty, identifier(head), identifier(tail), nonempty}, value, empty, nonempty)
}
