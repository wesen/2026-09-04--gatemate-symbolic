package dataflow

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

func TestPhysicalWorkbench(t *testing.T) {
	if *physicalDevice == "" {
		t.Skip("requires explicit -physical-device")
	}
	engine, err := NewSerial(*physicalDevice)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	execute := func(o Operation) *Token {
		t.Helper()
		v, e := engine.Execute(ctx, o)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	snap := func() Snapshot {
		t.Helper()
		s, e := engine.Snapshot(ctx)
		if e != nil {
			t.Fatal(e)
		}
		return s
	}
	rng := rand.New(rand.NewSource(8008))
	sources := []string{
		"input a: int16; let p=a*a; output p+p",
		"input a,b,c: int16; let square=a*a; let offset=b*c; output square+offset+int(a<b)",
		"input a: int16; let p=a*a; output p+p+p+p",
	}
	expressions := 0
	for which, source := range sources {
		p, e := Compile(source)
		if e != nil {
			t.Fatal(e)
		}
		execute(Operation{Kind: "reset"})
		execute(Operation{Kind: "load", Graph: &p.Graph})
		s := snap()
		if s.Graph != p.Graph {
			t.Fatal("physical descriptor mismatch")
		}
		for round := 0; round < 8; round++ {
			want := [Contexts]int32{}
			for c := byte(0); c < Contexts; c++ {
				if round > 0 {
					execute(Operation{Kind: "cancel", Context: c})
				}
				a := int64(rng.Intn(101) - 50)
				values := map[string]int64{"a": a}
				answer := a * a * 2
				if which == 1 {
					b, z := int64(rng.Intn(101)-50), int64(rng.Intn(101)-50)
					values["b"] = b
					values["c"] = z
					answer = a*a + b*z
					if a < b {
						answer++
					}
				}
				if which == 2 {
					answer = a * a * 4
				}
				want[c] = int32(answer)
				tokens, e := p.Tokens(c, byte(round), values)
				if e != nil {
					t.Fatal(e)
				}
				for _, token := range tokens {
					for retries := 0; ; retries++ {
						if retries > 100 {
							t.Fatal("input credit did not recover")
						}
						_, e := engine.Execute(ctx, Operation{Kind: "inject", Token: &token})
						if e == nil {
							break
						}
						if e != ErrFull {
							t.Fatal(e)
						}
						execute(Operation{Kind: "tick", Ticks: 1})
					}
				}
			}
			execute(Operation{Kind: "tick", Ticks: 400})
			seen := [Contexts]bool{}
			for n := 0; n < Contexts; n++ {
				result := execute(Operation{Kind: "poll"})
				if result == nil || result.Context >= Contexts || seen[result.Context] || result.Value != Int(want[result.Context]) || result.Epoch != byte(round) {
					t.Fatalf("physical result %+v want %v", result, want)
				}
				seen[result.Context] = true
				expressions++
			}
			if execute(Operation{Kind: "poll"}) != nil {
				t.Fatal("extra output")
			}
		}
	}
	p, _ := Compile(sources[0])
	execute(Operation{Kind: "reset"})
	execute(Operation{Kind: "load", Graph: &p.Graph})
	execute(Operation{Kind: "debug", Debug: &DebugControl{Mask: 1, Node: 0, Context: 255, Clear: true}})
	tokens, _ := p.Tokens(2, 0, map[string]int64{"a": 7})
	for _, token := range tokens {
		execute(Operation{Kind: "inject", Token: &token})
	}
	execute(Operation{Kind: "tick", Ticks: 100})
	stopped := snap()
	if !stopped.Debug.Halted || stopped.Debug.Reason != 1 || stopped.Counters["cycles"] >= 100 {
		t.Fatalf("physical breakpoint %+v", stopped.Debug)
	}
	found := false
	for _, e := range stopped.Debug.Events {
		if e.Kind == EventIssue && e.Token.Context == 2 && e.Token.Value == 49 {
			found = true
		}
	}
	if !found {
		t.Fatal("physical issue trace missing")
	}
	execute(Operation{Kind: "tick", Ticks: 100})
	held := snap()
	if held.Counters["cycles"] != stopped.Counters["cycles"] {
		t.Fatal("hardware progressed while halted")
	}
	execute(Operation{Kind: "debug", Debug: &DebugControl{Node: 255, Context: 255, Resume: true}})
	execute(Operation{Kind: "tick", Ticks: 100})
	if result := execute(Operation{Kind: "poll"}); result == nil || result.Value != 98 {
		t.Fatal("resume failed", result)
	}
	for n := 0; n < 40; n++ {
		execute(Operation{Kind: "cancel", Context: 2})
	}
	full := snap()
	if len(full.Debug.Events) != TraceCapacity || full.Debug.Dropped == 0 {
		t.Fatal("physical overflow accounting")
	}
	execute(Operation{Kind: "debug", Debug: &DebugControl{Node: 255, Context: 255, Clear: true}})
	if len(snap().Debug.Events) != 0 {
		t.Fatal("clear failed")
	}
	t.Logf("PASS %d compiled physical expressions across three graphs, descriptor readback, issue stop cycle %d, held halt, resume, trace overflow and clear", expressions, stopped.Counters["cycles"])
}
