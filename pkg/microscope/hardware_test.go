package microscope

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var hardwareDevice = flag.String("hardware-device", "", "Explicit FPGA serial device for opt-in physical tests")
var hardwareOut = flag.String("hardware-out", "", "Directory for physical wire logs and summary")

func TestPhysicalGraphEvents(t *testing.T) {
	if *hardwareDevice == "" {
		t.Skip("physical test requires -hardware-device")
	}
	out := *hardwareOut
	if out == "" {
		out = t.TempDir()
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	device, err := NewSerial(*hardwareDevice)
	if err != nil {
		t.Fatal(err)
	}
	defer device.Close()
	cases := []struct {
		Name  string
		Graph Graph
	}{
		{"triangle", Graph{3, 3, [][2]int{{0, 1}, {1, 2}, {0, 2}}, false}},
		{"unsatisfiable", Graph{3, 2, [][2]int{{0, 1}, {1, 2}, {0, 2}}, false}},
		{"path", Graph{4, 2, [][2]int{{0, 1}, {1, 2}, {2, 3}}, false}},
		{"first", Graph{3, 3, [][2]int{{0, 1}, {1, 2}, {0, 2}}, true}},
		{"root", Graph{2, 1, [][2]int{{0, 1}}, false}},
	}
	var results []map[string]any
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			log, err := os.Create(filepath.Join(out, "hardware-"+tc.Name+"-wire.log"))
			if err != nil {
				t.Fatal(err)
			}
			defer log.Close()
			device.Capture = log
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := device.Load(ctx, tc.Graph); err != nil {
				t.Fatal(err)
			}
			model := NewModel()
			_ = model.Load(ctx, tc.Graph)
			state := Initial(tc.Graph)
			events := 0
			for events < 10000 {
				actual, err := device.Step(ctx)
				if err != nil {
					t.Fatal(err)
				}
				want, err := model.Step(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if actual.Kind != Create && actual.Kind != Update {
					actual.ChoiceWord = want.ChoiceWord
				}
				if actual.Kind != Write && actual.Kind != Restore {
					actual.TrailWord = want.TrailWord
				}
				if actual != want {
					t.Fatalf("physical event mismatch\ngot %+v\nwant %+v", actual, want)
				}
				state, err = Apply(tc.Graph, state, actual)
				if err != nil {
					t.Fatal(err)
				}
				events++
				if actual.Terminal() {
					break
				}
			}
			if !state.Terminal() {
				t.Fatal("event watchdog")
			}
			results = append(results, map[string]any{"graph": tc.Name, "events": events, "solutions": state.Count, "terminal": state.Name, "match": true})
			if err := device.Reset(ctx); err != nil {
				t.Fatal(err)
			}
			first, err := device.Step(ctx)
			if err != nil || first.Sequence != 1 {
				t.Fatalf("reset: %+v %v", first, err)
			}
		})
	}
	data, _ := json.MarshalIndent(results, "", "  ")
	if err := os.WriteFile(filepath.Join(out, "hardware-summary.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}
