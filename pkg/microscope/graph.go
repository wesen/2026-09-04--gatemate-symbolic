package microscope

import (
	"context"
	"io"
	"math/bits"

	"github.com/pkg/errors"
)

// Graph uses labeled vertices and colors; edges are undirected and unique.
type Graph struct {
	Vertices  int      `json:"vertices"`
	Colors    int      `json:"colors"`
	Edges     [][2]int `json:"edges"`
	FirstOnly bool     `json:"firstOnly"`
}

func (g Graph) Adjacency() ([8]byte, error) {
	var a [8]byte
	if g.Vertices < 1 || g.Vertices > 8 || g.Colors < 1 || g.Colors > 8 {
		return a, errors.New("vertices and colors must be in 1..8")
	}
	for _, e := range g.Edges {
		u, v := e[0], e[1]
		if u < 0 || v < 0 || u >= g.Vertices || v >= g.Vertices || u == v {
			return a, errors.New("edges need two distinct active vertices")
		}
		if a[u]&(1<<v) != 0 {
			return a, errors.New("duplicate undirected edge")
		}
		a[u] |= 1 << v
		a[v] |= 1 << u
	}
	return a, nil
}

func (g Graph) Domains() [8]byte {
	var d [8]byte
	for i := range d {
		d[i] = 1
		if i < g.Vertices {
			d[i] = byte((1 << g.Colors) - 1)
		}
	}
	return d
}

const (
	Create byte = iota + 1
	Update
	Write
	Propagated
	Contradiction
	Restore
	Restored
	Pop
	Output
	Complete
	Fault
)

var eventNames = [...]string{"", "CREATE", "UPDATE", "WRITE", "PROPAGATED", "CONTRADICTION", "RESTORE", "RESTORED", "POP", "OUTPUT", "COMPLETE", "FAULT"}

func EventName(kind byte) string {
	if int(kind) < len(eventNames) {
		return eventNames[kind]
	}
	return "INVALID"
}

type Choice struct {
	Vertex     int  `json:"vertex"`
	Remaining  byte `json:"remaining"`
	Mark       int  `json:"mark"`
	Propagated byte `json:"propagated"`
}

func (c Choice) Word() uint64 {
	return uint64(c.Vertex)<<37 | uint64(c.Remaining)<<29 | uint64(c.Mark)<<22 | uint64(c.Propagated)<<14
}
func DecodeChoice(w uint64) Choice {
	return Choice{int(w >> 37 & 7), byte(w >> 29), int(w >> 22 & 127), byte(w >> 14)}
}

type TrailEntry struct {
	Vertex int  `json:"vertex"`
	Old    byte `json:"old"`
	Level  int  `json:"level"`
}

func (t TrailEntry) Word() uint32 {
	return uint32(t.Vertex)<<17 | uint32(t.Old)<<9 | uint32(t.Level)<<4
}
func DecodeTrail(w uint32) TrailEntry {
	return TrailEntry{int(w >> 17 & 7), byte(w >> 9), int(w >> 4 & 31)}
}

type Event struct {
	Sequence   uint32  `json:"sequence"`
	Kind       byte    `json:"kind"`
	Domains    [8]byte `json:"domains"`
	Propagated byte    `json:"propagated"`
	ChoiceTop  int     `json:"choiceTop"`
	TrailTop   int     `json:"trailTop"`
	Base       int     `json:"base"`
	Count      uint32  `json:"count"`
	Result     uint32  `json:"result"`
	Fault      byte    `json:"fault"`
	ChoiceWord uint64  `json:"choiceWord"`
	TrailWord  uint32  `json:"trailWord"`
}

func (e Event) Terminal() bool { return e.Kind == Complete || e.Kind == Fault }
func (e Event) Rows(n int) []int {
	r := make([]int, n)
	for i := range r {
		r[i] = int(e.Result >> uint(3*i) & 7)
	}
	return r
}
func Singleton(v byte) bool { return v != 0 && v&(v-1) == 0 }
func first(v byte) byte     { return v & -v }
func packedResult(d [8]byte) uint32 {
	var r uint32
	for i, v := range d {
		r |= uint32(bits.TrailingZeros8(v)) << uint(3*i)
	}
	return r
}

// Engine is a new instrument contract, shared by the explicit simulator and device.
type Engine interface {
	Load(context.Context, Graph) error
	Step(context.Context) (Event, error)
	Reset(context.Context) error
	Close() error
}

var ErrTerminal = io.EOF
