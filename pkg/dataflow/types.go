// Package dataflow specifies the fixed elastic-expression laboratory.
package dataflow

import (
	"encoding/binary"
	"math"

	"github.com/pkg/errors"
)

const Contexts = 4
const Nodes = 7

const (
	BadDestination byte = iota + 1
	DuplicateOperand
	BadTag
	IntegerOverflow
	BadDescriptor
)

type Value uint64

func Int(x int32) Value { return Value(uint32(x)) }
func Bool(x bool) Value {
	if x {
		return Value(1)<<36 | 1
	}
	return Value(1) << 36
}
func ErrorValue(code byte) Value { return Value(13)<<36 | Value(code) }
func (v Value) Tag() byte        { return byte(v>>36) & 15 }
func (v Value) Int32() int32     { return int32(v) }
func (v Value) Canonical() bool {
	return uint64(v)>>40 == 0 && (v>>32)&15 == 0 && (v.Tag() == 0 || v.Tag() == 1 && uint32(v) <= 1)
}

type Opcode byte

const (
	Mul Opcode = iota
	Add
	Less
	BoolToInt
	Copy
	Sub
)

func Evaluate(op Opcode, a, b Value) Value {
	if !a.Canonical() || (op != BoolToInt && op != Copy && !b.Canonical()) {
		return ErrorValue(BadTag)
	}
	if op == Copy {
		return a
	}
	if op == BoolToInt {
		if a.Tag() != 1 {
			return ErrorValue(BadTag)
		}
		return Int(a.Int32())
	}
	if a.Tag() != 0 || b.Tag() != 0 {
		return ErrorValue(BadTag)
	}
	x, y := int64(a.Int32()), int64(b.Int32())
	var result int64
	switch op {
	case Mul:
		if x < math.MinInt16 || x > math.MaxInt16 || y < math.MinInt16 || y > math.MaxInt16 {
			return ErrorValue(IntegerOverflow)
		}
		result = x * y
	case Add:
		result = x + y
	case Sub:
		result = x - y
	case Less:
		return Bool(x < y)
	default:
		return ErrorValue(BadDescriptor)
	}
	if result < math.MinInt32 || result > math.MaxInt32 {
		return ErrorValue(IntegerOverflow)
	}
	return Int(int32(result))
}

type Destination struct{ Node, Port byte }
type Descriptor struct {
	Op           Opcode
	Required     byte
	Destinations [2]Destination
	Count        byte
	Final        bool
}

type Token struct {
	Context  byte  `json:"context"`
	Epoch    byte  `json:"epoch"`
	Node     byte  `json:"node"`
	Port     byte  `json:"port"`
	Final    bool  `json:"final"`
	Producer byte  `json:"producer"`
	Subtype  byte  `json:"subtype"`
	Value    Value `json:"value"`
}

func (t Token) Bytes() ([10]byte, error) {
	var p [10]byte
	if t.Node > 63 || t.Port > 1 || uint64(t.Value)>>40 != 0 {
		return p, errors.New("token field exceeds envelope width")
	}
	p[0], p[1], p[2], p[3], p[4] = t.Context, t.Epoch, t.Node<<2|t.Port<<1, t.Producer, t.Subtype
	if t.Final {
		p[2] |= 1
	}
	p[5] = byte(t.Value >> 32)
	binary.BigEndian.PutUint32(p[6:], uint32(t.Value))
	return p, nil
}
func DecodeToken(p []byte) (Token, error) {
	if len(p) != 10 {
		return Token{}, errors.New("token must contain 10 bytes")
	}
	return Token{p[0], p[1], p[2] >> 2, p[2] >> 1 & 1, p[2]&1 != 0, p[3], p[4], Value(p[5])<<32 | Value(binary.BigEndian.Uint32(p[6:]))}, nil
}
func Source(context, epoch, node, port byte, v Value) Token {
	return Token{Context: context, Epoch: epoch, Node: node, Port: port, Producer: 255, Value: v}
}
func Inputs(context, epoch byte, values [6]int32) []Token {
	return []Token{Source(context, epoch, 0, 0, Int(values[0])), Source(context, epoch, 0, 1, Int(values[1])),
		Source(context, epoch, 1, 0, Int(values[2])), Source(context, epoch, 1, 1, Int(values[3])),
		Source(context, epoch, 3, 0, Int(values[4])), Source(context, epoch, 3, 1, Int(values[5]))}
}

type slot struct {
	Values          [2]Value
	Valid           byte
	Issued, Pending bool
}
type state struct {
	Graph  *Graph
	Epoch  [Contexts]byte
	Closed [Contexts]bool
	Slots  [Contexts][Nodes]slot
}

func (s *state) invalidate(c byte) { s.Slots[c] = [Nodes]slot{} }
func (s *state) accept(t Token) byte {
	if t.Node >= s.graph().Count || t.Port > 1 || s.descriptor(t.Node).Required&(1<<t.Port) == 0 || t.Final || t.Subtype != 0 {
		return BadDestination
	}
	v := &s.Slots[t.Context][t.Node]
	if v.Valid&(1<<t.Port) != 0 {
		return DuplicateOperand
	}
	v.Values[t.Port] = t.Value
	v.Valid |= 1 << t.Port
	if v.Valid&s.descriptor(t.Node).Required == s.descriptor(t.Node).Required && !v.Issued {
		v.Issued = true
		v.Pending = true
	}
	return 0
}
func fault(c, e, n, code byte) Token {
	return Token{Context: c, Epoch: e, Node: n, Producer: n, Final: true, Subtype: code, Value: ErrorValue(code)}
}
