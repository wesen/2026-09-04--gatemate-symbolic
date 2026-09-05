package dataflow

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
)

var physicalDevice = flag.String("physical-device", "", "Explicit FPGA UART path; enables destructive-reset physical qualification")
var physicalWireLog = flag.String("physical-wire-log", "", "Optional wire evidence file for physical qualification")

func TestPhysicalQualification(t *testing.T) {
	if *physicalDevice == "" {
		t.Skip("requires explicit -physical-device")
	}
	s, err := NewSerial(*physicalDevice)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if *physicalWireLog != "" {
		f, err := os.Create(*physicalWireLog)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		s.Capture = f
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	rng := rand.New(rand.NewSource(54831))
	execute := func(o Operation) *Token {
		token, err := s.Execute(ctx, o)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	snapshot := func() Snapshot {
		v, err := s.Snapshot(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	for trial := 0; trial < 32; trial++ {
		execute(Operation{Kind: "reset"})
		var source []Token
		want := map[byte]int32{}
		for c := byte(0); c < Contexts; c++ {
			var v [6]int32
			for n := range v {
				v[n] = int32(rng.Intn(201) - 100)
			}
			result := v[0]*v[1] + v[2]*v[3]
			if v[4] < v[5] {
				result++
			}
			want[c] = result
			source = append(source, Inputs(c, 0, v)...)
		}
		rng.Shuffle(len(source), func(i, j int) { source[i], source[j] = source[j], source[i] })
		for _, token := range source {
			for attempts := 0; ; attempts++ {
				_, err := s.Execute(ctx, Operation{Kind: "inject", Token: &token})
				if err == nil {
					break
				}
				if attempts > 1000 || !errors.Is(err, ErrFull) {
					t.Fatal(err)
				}
				execute(Operation{Kind: "tick", Ticks: uint32(rng.Intn(4) + 1)})
			}
		}
		outputs := []Token{}
		for ticks := 0; ticks < 1000 && len(outputs) < 4; ticks++ {
			execute(Operation{Kind: "tick", Ticks: uint32(rng.Intn(16) + 1)})
			if rng.Intn(3) == 0 {
				if token := execute(Operation{Kind: "poll"}); token != nil {
					outputs = append(outputs, *token)
				}
			}
		}
		assertOutputs(t, outputs, want)
		state := snapshot()
		if !state.Quiescent || state.Counters["activations"] != 24 {
			t.Fatalf("trial %d final %+v", trial, state.Counters)
		}
	}
	execute(Operation{Kind: "reset"})
	for _, token := range Inputs(0, 0, [6]int32{7, 6, 3, 5, 2, 9}) {
		execute(Operation{Kind: "inject", Token: &token})
	}
	execute(Operation{Kind: "tick", Ticks: 128})
	before := snapshot()
	if len(before.Output) != 1 {
		t.Fatal("missing offered output")
	}
	if _, err := s.Execute(ctx, Operation{Kind: "cancel", Context: 0}); !errors.Is(err, ErrBlocked) {
		t.Fatal(err)
	}
	execute(Operation{Kind: "tick", Ticks: 64})
	after := snapshot()
	if len(after.Output) != 1 || after.Output[0] != before.Output[0] {
		t.Fatal("offered output changed")
	}
	result := execute(Operation{Kind: "poll"})
	if result == nil || result.Value != 58 {
		t.Fatal(result)
	}
	for n := 0; n < 256; n++ {
		execute(Operation{Kind: "cancel", Context: 0})
	}
	after = snapshot()
	if after.Epochs[0] != 0 {
		t.Fatal("physical epoch did not wrap after drained cancellations")
	}
	t.Log("PASS 32 randomized four-context physical expressions, held-output cancellation guard, and 256-step drained epoch wrap")
}
