// Package ir defines the fixed-width LFL1 code, heap and stack records.
package ir

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	Capacity             = 2048
	StackCapacity        = 512
	Absent        uint16 = 0xffff
)

type Tag uint8

const (
	Int Tag = iota
	Bool
	Nil
	Cons
	Fun
	Thunk
	Ind
	Blackhole
	Env
	Error Tag = 13
	Free  Tag = 15
)

type Opcode uint8

const (
	Const Opcode = iota
	Var
	Lambda
	App
	Let
	Letrec
	Prim
	If
	MakeCons
	Case
)

type Primitive uint32

const (
	Add Primitive = iota
	Sub
	Mul
	EQ
	LE
)

type Object struct {
	Tag     Tag
	Flags   uint8
	A, B    uint16
	Payload uint32
}

func (o Object) Bytes() [10]byte {
	var b [10]byte
	b[0] = byte(o.Tag)
	b[1] = o.Flags
	binary.BigEndian.PutUint16(b[2:4], o.A)
	binary.BigEndian.PutUint16(b[4:6], o.B)
	binary.BigEndian.PutUint32(b[6:10], o.Payload)
	return b
}
func (o Object) Hex() string { b := o.Bytes(); return hex.EncodeToString(b[:]) }
func ParseObject(s string) (Object, error) {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 10 {
		return Object{}, fmt.Errorf("object requires 20 hexadecimal characters")
	}
	return Object{Tag: Tag(b[0]), Flags: b[1], A: binary.BigEndian.Uint16(b[2:4]), B: binary.BigEndian.Uint16(b[4:6]), Payload: binary.BigEndian.Uint32(b[6:10])}, nil
}
func (o Object) Terminal() bool {
	switch o.Tag {
	case Int, Bool, Nil, Cons, Fun, Error:
		return true
	}
	return false
}

type Code struct {
	Op             Opcode
	Flags          uint8
	A, B, C        uint16
	Immediate      uint32
	Span, Reserved uint16
}

func (c Code) Bytes() [16]byte {
	var b [16]byte
	b[0] = byte(c.Op)
	b[1] = c.Flags
	binary.BigEndian.PutUint16(b[2:4], c.A)
	binary.BigEndian.PutUint16(b[4:6], c.B)
	binary.BigEndian.PutUint16(b[6:8], c.C)
	binary.BigEndian.PutUint32(b[8:12], c.Immediate)
	binary.BigEndian.PutUint16(b[12:14], c.Span)
	binary.BigEndian.PutUint16(b[14:16], c.Reserved)
	return b
}
func (c Code) Hex() string { b := c.Bytes(); return hex.EncodeToString(b[:]) }
func ParseCode(s string) (Code, error) {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 16 {
		return Code{}, fmt.Errorf("code requires 32 hexadecimal characters")
	}
	return Code{Op: Opcode(b[0]), Flags: b[1], A: binary.BigEndian.Uint16(b[2:4]), B: binary.BigEndian.Uint16(b[4:6]), C: binary.BigEndian.Uint16(b[6:8]), Immediate: binary.BigEndian.Uint32(b[8:12]), Span: binary.BigEndian.Uint16(b[12:14]), Reserved: binary.BigEndian.Uint16(b[14:16])}, nil
}

type FrameKind uint8

const (
	Arg FrameKind = iota + 1
	PrimRight
	PrimApply
	Branch
	Match
	Update
)

type Frame struct {
	Kind                                 FrameKind
	Op                                   uint8
	A, B, C, D, SavedRef, Span, Reserved uint16
}

func (f Frame) Bytes() [16]byte {
	var b [16]byte
	b[0] = byte(f.Kind)
	b[1] = f.Op
	for i, v := range []uint16{f.A, f.B, f.C, f.D, f.SavedRef, f.Span, f.Reserved} {
		binary.BigEndian.PutUint16(b[2+2*i:4+2*i], v)
	}
	return b
}
func (f Frame) Hex() string { b := f.Bytes(); return hex.EncodeToString(b[:]) }
