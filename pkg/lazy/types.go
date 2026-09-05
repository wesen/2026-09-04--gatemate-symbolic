// Package lazy implements the bounded, single-evaluator lazy graph laboratory.
package lazy

import (
	"context"
	"github.com/pkg/errors"
	"math"
)

const HeapCapacity = 1024
const StackCapacity = 512
const IndirectionLimit = 1024
const TraceCapacity = 64

type Tag byte

const (
	Int       Tag = 0
	Add       Tag = 1
	Mul       Tag = 2
	Thunk     Tag = 3
	Ind       Tag = 4
	Blackhole Tag = 5
	Error     Tag = 13
	Free      Tag = 15
)
const (
	AddressFault uint32 = iota + 1
	StackOverflow
	CyclicThunk
	IndirectionCycle
	TypeFault
	ArithmeticOverflow
	OwnershipFault
)

var FaultNames = map[uint32]string{AddressFault: "HEAP_ADDRESS", StackOverflow: "CONTINUATION_OVERFLOW", CyclicThunk: "CYCLIC_THUNK", IndirectionCycle: "INDIRECTION_CYCLE", TypeFault: "TYPE_FAULT", ArithmeticOverflow: "ARITHMETIC_OVERFLOW", OwnershipFault: "OWNERSHIP_FAULT"}

type Word uint64

func Integer(v int32) Word         { return Word(uint32(v)) }
func Node(t Tag, a, b uint16) Word { return Word(t)<<36 | Word(a)<<16 | Word(b) }
func Fault(code uint32) Word       { return Word(Error)<<36 | Word(code) }
func (w Word) Tag() Tag            { return Tag(w>>36) & 15 }
func (w Word) A() uint16           { return uint16(w >> 16) }
func (w Word) B() uint16           { return uint16(w) }
func (w Word) Int32() int32        { return int32(w) }
func (w Word) Canonical() bool {
	if w>>40 != 0 || (w>>32)&15 != 0 {
		return false
	}
	switch w.Tag() {
	case Int, Add, Mul, Error:
		return true
	case Thunk, Ind:
		return w.B() == 0
	case Blackhole, Free:
		return uint32(w) == 0
	}
	return false
}
func Evaluate(op Tag, a, b Word) Word {
	if a.Tag() == Error {
		return a
	}
	if b.Tag() == Error {
		return b
	}
	if !a.Canonical() || !b.Canonical() || a.Tag() != Int || b.Tag() != Int {
		return Fault(TypeFault)
	}
	x, y := int64(a.Int32()), int64(b.Int32())
	var z int64
	switch op {
	case Add:
		z = x + y
	case Mul:
		z = x * y
	default:
		return Fault(TypeFault)
	}
	if z < math.MinInt32 || z > math.MaxInt32 {
		return Fault(ArithmeticOverflow)
	}
	return Integer(int32(z))
}

type Image struct {
	Nodes []Word `json:"nodes"`
	Root  uint16 `json:"root"`
}

func (i Image) Validate() error {
	if len(i.Nodes) == 0 || len(i.Nodes) > HeapCapacity || int(i.Root) >= len(i.Nodes) {
		return errors.New("image requires 1..1024 nodes and an in-range root")
	}
	for n, w := range i.Nodes {
		if !w.Canonical() || w.Tag() == Blackhole {
			return errors.Errorf("node %d has invalid encoding or external blackhole", n)
		}
	}
	return nil
}

type Frame struct {
	Kind    byte   `json:"kind"`
	Op      Tag    `json:"op"`
	Address uint16 `json:"address"`
	Value   Word   `json:"value"`
}

const (
	EvalRight byte = 1
	Apply     byte = 2
	Update    byte = 3
)

type Counters struct {
	Cycles       uint32 `json:"cycles"`
	Reads        uint32 `json:"reads"`
	Writes       uint32 `json:"writes"`
	Claims       uint32 `json:"claims"`
	Updates      uint32 `json:"updates"`
	Muls         uint32 `json:"muls"`
	Adds         uint32 `json:"adds"`
	Indirections uint32 `json:"indirections"`
	Blackholes   uint32 `json:"blackholes"`
	MaxStack     uint32 `json:"maxStack"`
	Stalls       uint32 `json:"stalls"`
	Faults       uint32 `json:"faults"`
}
type Mutation struct {
	Cycle   uint32 `json:"cycle"`
	Address uint16 `json:"address"`
	Old     Word   `json:"old"`
	New     Word   `json:"new"`
}
type Snapshot struct {
	Source   string     `json:"source"`
	State    string     `json:"state"`
	Address  uint16     `json:"address"`
	Result   Word       `json:"result"`
	Valid    bool       `json:"valid"`
	Heap     []Word     `json:"heap"`
	Stack    []Frame    `json:"stack"`
	Counters Counters   `json:"counters"`
	Trace    []Mutation `json:"trace"`
	Dropped  uint32     `json:"dropped"`
}
type Operation struct {
	Kind  string `json:"kind"`
	Image *Image `json:"image,omitempty"`
	Root  uint16 `json:"root,omitempty"`
	Ticks uint32 `json:"ticks,omitempty"`
}

func (o Operation) Validate() error {
	if o.Image != nil && o.Kind != "load" || o.Ticks != 0 && o.Kind != "tick" || o.Root != 0 && o.Kind != "force" {
		return errors.New("misplaced operation fields")
	}
	switch o.Kind {
	case "reset", "poll":
		return nil
	case "force":
		if o.Root >= HeapCapacity {
			return errors.New("root exceeds heap capacity")
		}
		return nil
	case "tick":
		if o.Ticks > 1000000 {
			return errors.New("tick batch exceeds one million")
		}
		return nil
	case "load":
		if o.Image == nil {
			return errors.New("load requires image")
		}
		return o.Image.Validate()
	}
	return errors.New("unknown operation")
}

type Engine interface {
	Execute(context.Context, Operation) (*Word, error)
	Snapshot(context.Context) (Snapshot, error)
	Close() error
}

func Example(name string) (Image, error) {
	switch name {
	case "shared":
		return Image{[]Word{Integer(21), Integer(2), Node(Mul, 0, 1), Node(Thunk, 2, 0), Node(Add, 3, 3), Node(Add, 3, 3), Node(Add, 4, 5)}, 6}, nil
	case "cycle":
		return Image{[]Word{Integer(1), Node(Thunk, 2, 0), Node(Add, 0, 1)}, 1}, nil
	case "indirection":
		return Image{[]Word{Integer(-42), Node(Ind, 0, 0), Node(Ind, 1, 0)}, 2}, nil
	case "indirection-cycle":
		return Image{[]Word{Node(Ind, 1, 0), Node(Ind, 0, 0)}, 0}, nil
	case "overflow":
		return Image{[]Word{Integer(math.MaxInt32), Integer(2), Node(Mul, 0, 1), Node(Thunk, 2, 0)}, 3}, nil
	}
	return Image{}, errors.New("unknown example")
}
