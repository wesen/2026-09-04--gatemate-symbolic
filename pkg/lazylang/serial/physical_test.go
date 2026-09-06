package serial

import (
	"context"
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/compile"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/ir"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/machine"
)

var physicalDevice = flag.String("lfl-device", "", "explicit LFL1 physical UART path; empty skips hardware test")
var physicalCapture = flag.String("lfl-capture", "", "optional UART evidence log path")

func TestPhysicalPrograms(t *testing.T) {
	if *physicalDevice == "" {
		t.Skip("requires explicit -lfl-device")
	}
	c, err := Open(*physicalDevice)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if *physicalCapture != "" {
		f, err := os.Create(*physicalCapture)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		c.Capture = f
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	for _, name := range []string{"shared", "closure", "unused", "cycle", "productive", "squares"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("../../../examples/lazylang/" + name + ".lazy")
			if err != nil {
				t.Fatal(err)
			}
			a, ds := compile.Compile(ctx, string(source))
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			if err := c.Load(ctx, a); err != nil {
				t.Fatal(err)
			}
			m, err := machine.New(a)
			if err != nil {
				t.Fatal(err)
			}
			demand := func(ref uint16) ir.Object {
				t.Helper()
				if err := m.Demand(ref); err != nil {
					t.Fatal(err)
				}
				if err := m.Tick(ctx, 100000); err != nil {
					t.Fatal(err)
				}
				want, ok := m.Poll()
				if !ok {
					t.Fatal("model timeout")
				}
				if err := c.Demand(ctx, ref); err != nil {
					t.Fatal(err)
				}
				if err := c.Tick(ctx, 100000); err != nil {
					t.Fatal(err)
				}
				got, valid, err := c.Poll(ctx)
				if err != nil || !valid || got != want {
					t.Fatalf("output %d/%v want %d: %v", got, valid, want, err)
				}
				o, err := c.ReadObject(ctx, got)
				if err != nil {
					t.Fatal(err)
				}
				expected, _ := ir.ParseObject(m.Snapshot().Heap[want])
				if o != expected {
					t.Fatal(o, expected)
				}
				return o
			}
			o := demand(a.Root)
			values := []int32{}
			if o.Tag == ir.Cons {
				for i := 0; i < 8; i++ {
					h := demand(o.A)
					values = append(values, int32(h.Payload))
					if i < 7 {
						o = demand(o.B)
					}
				}
			}
			got, err := c.Snapshot(ctx)
			if err != nil {
				t.Fatal(err)
			}
			want := m.Snapshot()
			if !reflect.DeepEqual(got.Heap, want.Heap) || !reflect.DeepEqual(got.Provenance, want.Provenance) {
				t.Fatal("physical heap or provenance differs from model")
			}
			actual, expected := got.Counters, want.Counters
			// Synchronous hardware rereads terminal values/primitive operands; host tick
			// budgets also differ in output-stall duration. Semantic counts must match.
			actual.Cycles = 0
			actual.HeapReads = 0
			actual.OutputStalls = 0
			expected.Cycles = 0
			expected.HeapReads = 0
			expected.OutputStalls = 0
			if actual != expected {
				t.Fatal("semantic counters", actual, expected)
			}
			t.Logf("%s artifact=%s heap=%d result=%+v prefix=%v counters=%+v", name, a.ID, len(got.Heap), o, values, got.Counters)
			if len(got.Trace) != len(want.Trace) {
				t.Fatal("trace length")
			}
			for i, event := range got.Trace {
				expected := want.Trace[i]
				event.Cycle = 0
				expected.Cycle = 0
				if event != expected {
					t.Fatal(fmt.Sprintf("trace %d", i), event, expected)
				}
			}
		})
	}
}
