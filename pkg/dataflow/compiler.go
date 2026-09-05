package dataflow

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/pkg/errors"
)

type InputBinding struct {
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	Destinations []Destination `json:"destinations"`
}
type ConstantBinding struct {
	Destination Destination `json:"destination"`
	Value       Value       `json:"value"`
}
type NodeInfo struct {
	Node byte   `json:"node"`
	Name string `json:"name"`
	Line int    `json:"line"`
	Type string `json:"type"`
}
type Program struct {
	Source    string            `json:"source"`
	Graph     Graph             `json:"graph"`
	Inputs    []InputBinding    `json:"inputs"`
	Constants []ConstantBinding `json:"constants"`
	Nodes     []NodeInfo        `json:"nodes"`
}

// Tokens validates the complete input set before returning any source tokens.
func (p Program) Tokens(c, e byte, values map[string]int64) ([]Token, error) {
	if c >= Contexts {
		return nil, errors.New("context must be 0..3")
	}
	if err := p.Graph.Validate(); err != nil {
		return nil, err
	}
	if len(values) != len(p.Inputs) {
		return nil, errors.New("provide exactly the compiled named inputs")
	}
	out := []Token{}
	for _, in := range p.Inputs {
		v, ok := values[in.Name]
		if !ok {
			return nil, errors.Errorf("missing input %s", in.Name)
		}
		if v < math.MinInt32 || v > math.MaxInt32 || in.Type == "int16" && (v < math.MinInt16 || v > math.MaxInt16) || in.Type == "bool" && (v < 0 || v > 1) {
			return nil, errors.Errorf("input %s is outside %s range", in.Name, in.Type)
		}
		value := Int(int32(v))
		if in.Type == "bool" {
			value = Bool(v == 1)
		}
		for _, d := range in.Destinations {
			out = append(out, Source(c, e, d.Node, d.Port, value))
		}
	}
	for _, v := range p.Constants {
		out = append(out, Source(c, e, v.Destination.Node, v.Destination.Port, v.Value))
	}
	return out, nil
}

type expression struct {
	op        Opcode
	args      []*expression
	typ, name string
	line      int
	external  bool
	literal   *Value
}
type compiler struct {
	s             scanner.Scanner
	tok           rune
	text          string
	line          int
	names         map[string]*expression
	inputs        []*expression
	depth, tokens int
	scanErr       error
}

func (p *compiler) next() {
	p.tok = p.s.Scan()
	p.text = p.s.TokenText()
	p.line = p.s.Position.Line
	p.tokens++
}
func (p *compiler) err(message string) error {
	return errors.Errorf("line %d column %d: %s", p.line, p.s.Position.Column, message)
}
func (p *compiler) take(text string) error {
	if p.text != text {
		return p.err("expected " + text)
	}
	p.next()
	return nil
}
func (p *compiler) identifier() (string, error) {
	if p.tok != scanner.Ident {
		return "", p.err("expected identifier")
	}
	n := p.text
	if n == "input" || n == "let" || n == "output" || n == "int" || n == "true" || n == "false" {
		return "", p.err("reserved identifier " + n)
	}
	p.next()
	return n, nil
}
func precedence(tok rune) int {
	switch tok {
	case '<':
		return 1
	case '+', '-':
		return 2
	case '*':
		return 3
	}
	return 0
}
func (p *compiler) expr(min int) (*expression, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > 128 || p.tokens > 4096 {
		return nil, p.err("expression complexity limit exceeded")
	}
	line := p.line
	var left *expression
	switch {
	case p.text == "int":
		p.next()
		if err := p.take("("); err != nil {
			return nil, err
		}
		a, err := p.expr(1)
		if err != nil {
			return nil, err
		}
		if err = p.take(")"); err != nil {
			return nil, err
		}
		if a.typ != "bool" {
			return nil, p.err("int conversion requires bool")
		}
		left = &expression{op: BoolToInt, args: []*expression{a}, typ: "int32", name: "int", line: line}
	case p.tok == '(':
		p.next()
		var err error
		left, err = p.expr(1)
		if err != nil {
			return nil, err
		}
		if err = p.take(")"); err != nil {
			return nil, err
		}
	case p.tok == scanner.Int || p.tok == '-':
		sign := int64(1)
		if p.tok == '-' {
			sign = -1
			p.next()
			if p.tok != scanner.Int {
				return nil, p.err("unary minus requires an integer literal")
			}
		}
		n, err := strconv.ParseInt(p.text, 10, 64)
		if err != nil {
			return nil, p.err("invalid decimal integer")
		}
		n *= sign
		if n < math.MinInt32 || n > math.MaxInt32 {
			return nil, p.err("literal outside int32")
		}
		v := Int(int32(n))
		typ := "int32"
		if n >= math.MinInt16 && n <= math.MaxInt16 {
			typ = "int16"
		}
		left = &expression{external: true, literal: &v, typ: typ, name: fmt.Sprint(n), line: line}
		p.next()
	case p.text == "true" || p.text == "false":
		v := Bool(p.text == "true")
		left = &expression{external: true, literal: &v, typ: "bool", name: p.text, line: line}
		p.next()
	case p.tok == scanner.Ident:
		var ok bool
		left, ok = p.names[p.text]
		if !ok {
			return nil, p.err("undefined name " + p.text)
		}
		p.next()
	default:
		return nil, p.err("expected expression")
	}
	for prec := precedence(p.tok); prec >= min && prec != 0; prec = precedence(p.tok) {
		op := p.tok
		p.next()
		right, err := p.expr(prec + 1)
		if err != nil {
			return nil, err
		}
		if left.typ == "bool" || right.typ == "bool" {
			return nil, p.err("arithmetic and comparison require integers; use int(bool)")
		}
		code, typ := Add, "int32"
		switch op {
		case '-':
			code = Sub
		case '*':
			code = Mul
			if left.typ != "int16" || right.typ != "int16" {
				return nil, p.err("MUL operands must have statically known int16 range")
			}
		case '<':
			code = Less
			typ = "bool"
		}
		left = &expression{op: code, args: []*expression{left, right}, typ: typ, name: string(op), line: line}
	}
	return left, nil
}
func Compile(source string) (Program, error) {
	out := Program{Source: source, Inputs: []InputBinding{}, Constants: []ConstantBinding{}, Nodes: []NodeInfo{}}
	if len(source) > 128*1024 {
		return out, errors.New("source exceeds 128 KiB")
	}
	p := compiler{names: map[string]*expression{}}
	p.s.Init(strings.NewReader(source))
	p.s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanComments | scanner.SkipComments
	p.s.Whitespace = 1<<' ' | 1<<'\t' | 1<<'\r'
	p.s.Error = func(s *scanner.Scanner, msg string) { p.scanErr = errors.Errorf("line %d: %s", s.Position.Line, msg) }
	p.next()
	var root *expression
	for p.tok != scanner.EOF {
		if p.tok == '\n' || p.tok == ';' {
			p.next()
			continue
		}
		if p.tokens > 4096 {
			return out, p.err("source token limit exceeded")
		}
		kind := p.text
		p.next()
		switch kind {
		case "input":
			names := []string{}
			for {
				n, err := p.identifier()
				if err != nil {
					return out, err
				}
				names = append(names, n)
				if p.tok != ',' {
					break
				}
				p.next()
			}
			if err := p.take(":"); err != nil {
				return out, err
			}
			typ := p.text
			if typ != "int16" && typ != "int32" && typ != "bool" {
				return out, p.err("input type must be int16, int32 or bool")
			}
			p.next()
			for _, n := range names {
				if _, ok := p.names[n]; ok {
					return out, p.err("duplicate name " + n)
				}
				x := &expression{external: true, typ: typ, name: n, line: p.line}
				p.names[n] = x
				p.inputs = append(p.inputs, x)
			}
		case "let":
			n, err := p.identifier()
			if err != nil {
				return out, err
			}
			if _, ok := p.names[n]; ok {
				return out, p.err("duplicate name " + n)
			}
			if err = p.take("="); err != nil {
				return out, err
			}
			x, err := p.expr(1)
			if err != nil {
				return out, err
			}
			if !x.external {
				x.name = n
			}
			p.names[n] = x
		case "output":
			if root != nil {
				return out, p.err("exactly one output statement is allowed")
			}
			var err error
			root, err = p.expr(1)
			if err != nil {
				return out, err
			}
		default:
			return out, p.err("expected input, let or output statement")
		}
		if p.tok != scanner.EOF && p.tok != '\n' && p.tok != ';' {
			return out, p.err("expected newline or semicolon")
		}
	}
	if p.scanErr != nil {
		return out, p.scanErr
	}
	if root == nil {
		return out, errors.New("program requires output")
	}
	if root.external {
		root = &expression{op: Copy, args: []*expression{root}, typ: root.typ, name: "output", line: root.line}
	}
	type use struct {
		node *expression
		port int
	}
	uses := map[*expression][]use{}
	seen := map[*expression]bool{}
	ops := []*expression{}
	var visit func(*expression)
	visit = func(x *expression) {
		if seen[x] || x.external {
			return
		}
		seen[x] = true
		for port, a := range x.args {
			uses[a] = append(uses[a], use{x, port})
			visit(a)
		}
		ops = append(ops, x)
	}
	visit(root)
	// Distribute internal values with >2 consumers using explicit COPY nodes.
	for _, x := range ops {
		u := uses[x]
		for len(u) > 2 {
			tail := u[len(u)-2:]
			cp := &expression{op: Copy, args: []*expression{x}, typ: x.typ, name: x.name + " fanout", line: x.line}
			for _, target := range tail {
				target.node.args[target.port] = cp
			}
			u = append(u[:len(u)-2], use{cp, 0})
		}
	}
	order := []*expression{}
	seen = map[*expression]bool{}
	var sort func(*expression)
	sort = func(x *expression) {
		if x.external || seen[x] {
			return
		}
		seen[x] = true
		for _, a := range x.args {
			sort(a)
		}
		order = append(order, x)
	}
	sort(root)
	if len(order) > Nodes {
		return out, errors.Errorf("compiled graph requires %d nodes including fanout; hardware capacity is %d", len(order), Nodes)
	}
	ids := map[*expression]byte{}
	for n, x := range order {
		ids[x] = byte(n)
	}
	bindings := map[*expression][]Destination{}
	out.Graph.Count = byte(len(order))
	for n, x := range order {
		d := &out.Graph.Descriptors[n]
		d.Op = x.op
		d.Required = required(x.op)
		d.Final = x == root
		out.Nodes = append(out.Nodes, NodeInfo{byte(n), x.name, x.line, x.typ})
		for port, a := range x.args {
			dest := Destination{byte(n), byte(port)}
			if a.external {
				bindings[a] = append(bindings[a], dest)
				continue
			}
			producer := &out.Graph.Descriptors[ids[a]]
			if producer.Count >= 2 {
				return out, errors.New("internal fanout lowering error")
			}
			producer.Destinations[producer.Count] = dest
			producer.Count++
		}
	}
	for _, in := range p.inputs {
		if dest := bindings[in]; len(dest) > 0 {
			out.Inputs = append(out.Inputs, InputBinding{in.name, in.typ, dest})
		}
	}
	// Emit constants in destination order, avoiding nondeterministic map iteration.
	for n, x := range order {
		for port, a := range x.args {
			if a.literal != nil {
				out.Constants = append(out.Constants, ConstantBinding{Destination{byte(n), byte(port)}, *a.literal})
			}
		}
	}
	if err := out.Graph.Validate(); err != nil {
		return out, errors.Wrap(err, "compiled graph")
	}
	return out, nil
}
