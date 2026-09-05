package lazy

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"
)

var device = flag.String("lazy-physical-device", "", "UART for opt-in lazy FPGA qualification")
var wirelog = flag.String("lazy-wire-log", "", "Optional qualification wire log")

func TestPhysicalLazyQualification(t *testing.T) {
	if *device == "" {
		t.Skip("requires lazy FPGA device")
	}
	e, err := NewSerial(*device)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if *wirelog != "" {
		f, err := os.Create(*wirelog)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		e.Capture = f
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	execute := func(o Operation) *Word {
		t.Helper()
		v, err := e.Execute(ctx, o)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	snap := func() Snapshot {
		t.Helper()
		s, err := e.Snapshot(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	for _, name := range []string{"shared", "cycle", "indirection", "indirection-cycle", "overflow"} {
		if _, _, err := RunExample(ctx, e, name); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	i, _ := Example("shared")
	execute(Operation{Kind: "load", Image: &i})
	execute(Operation{Kind: "force", Root: 6})
	execute(Operation{Kind: "tick", Ticks: 9})
	claimed := snap()
	if claimed.Counters.Claims != 1 || claimed.Counters.Updates != 0 || claimed.Heap[3].Tag() != Blackhole || len(claimed.Stack) != 3 {
		t.Fatalf("claim snapshot %+v", claimed)
	}
	again := snap()
	if !reflect.DeepEqual(claimed, again) {
		t.Fatal("inspection mutated state")
	}
	execute(Operation{Kind: "tick", Ticks: 1000})
	held := snap()
	if !held.Valid || held.Result != 168 || held.Counters.Muls != 1 {
		t.Fatal("shared result")
	}
	execute(Operation{Kind: "tick", Ticks: 20})
	if s := snap(); s.Result != held.Result || !s.Valid || s.Counters.Stalls != held.Counters.Stalls+20 {
		t.Fatal("held output stability")
	}
	if v := execute(Operation{Kind: "poll"}); v == nil || *v != 168 {
		t.Fatal("poll")
	}
	execute(Operation{Kind: "force", Root: 6})
	execute(Operation{Kind: "tick", Ticks: 1000})
	if s := snap(); s.Result != 168 || s.Counters.Muls != 1 || s.Counters.Claims != 1 {
		t.Fatal("repeat force recomputed body")
	}
	r := rand.New(rand.NewSource(9011))
	for trial := 0; trial < 60; trial++ {
		nodes := []Word{Integer(int32(r.Intn(200001) - 100000)), Integer(int32(r.Intn(200001) - 100000))}
		for n := 2; n < 18; n++ {
			a, b := uint16(r.Intn(n)), uint16(r.Intn(n))
			switch r.Intn(4) {
			case 0:
				nodes = append(nodes, Node(Thunk, a, 0))
			case 1:
				nodes = append(nodes, Node(Mul, a, b))
			default:
				nodes = append(nodes, Node(Add, a, b))
			}
		}
		image := Image{nodes, 17}
		want, heap := Reference(image, StackCapacity)
		execute(Operation{Kind: "load", Image: &image})
		execute(Operation{Kind: "force", Root: 17})
		execute(Operation{Kind: "tick", Ticks: 20000})
		s := snap()
		if !s.Valid || s.Result != want || !reflect.DeepEqual(s.Heap, heap) {
			t.Fatalf("physical graph %d mismatch result %x want %x", trial, s.Result, want)
		}
	}
	nodes := []Word{Integer(1)}
	for n := 1; n < 80; n++ {
		nodes = append(nodes, Node(Thunk, uint16(n-1), 0))
	}
	deep := Image{nodes, 79}
	execute(Operation{Kind: "load", Image: &deep})
	execute(Operation{Kind: "force", Root: 79})
	execute(Operation{Kind: "tick", Ticks: 20000})
	s := snap()
	if s.Result != 1 || s.Counters.Claims != 79 || s.Counters.Updates != 79 || len(s.Trace) != 64 || s.Dropped != 94 {
		t.Fatalf("trace overflow %+v", s)
	}
	overflow := Image{[]Word{Node(Thunk, 1, 0), Node(Add, 1, 1)}, 0}
	execute(Operation{Kind: "load", Image: &overflow})
	execute(Operation{Kind: "force"})
	execute(Operation{Kind: "tick", Ticks: 20000})
	s = snap()
	if s.Result != Fault(StackOverflow) || s.Heap[0] != Fault(StackOverflow) || s.Counters.Claims != 1 || s.Counters.Updates != 1 || s.Counters.MaxStack != 512 {
		t.Fatalf("stack overflow unwind %+v", s)
	}
	t.Log("PASS 60 randomized physical graphs, five directed examples, live claim/stack inspection, held result, repeated force, 79 nested thunk updates, trace overflow and 512-frame overflow unwind")
}
