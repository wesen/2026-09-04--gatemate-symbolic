package microscope

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGraphRTLSerialEvents(t *testing.T) {
	if _, err := exec.LookPath("iverilog"); err != nil {
		t.Skip("source OSS CAD Suite environment for RTL UART tests")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "graph.vvp")
	files := []string{"queens_rollback/rtl/queens_types_pkg.sv", "symbolic_eval/rtl/sync_sdp_ram.sv", "symbolic_eval/rtl/uart_tx.sv", "graph_microscope/rtl/graph_core.sv", "graph_microscope/rtl/uart_rx.sv", "graph_microscope/rtl/graph_link.sv", "graph_microscope/sim/tb_graph_link.sv"}
	args := []string{"-g2012", "-s", "tb_graph_link", "-o", binary}
	for _, f := range files {
		args = append(args, filepath.Join(root, f))
	}
	if out, err := exec.Command("iverilog", args...).CombinedOutput(); err != nil {
		t.Fatalf("compile: %v\n%s", err, out)
	}
	cases := []struct {
		name              string
		g                 Graph
		negative, restart bool
	}{
		{"triangle", Graph{3, 3, [][2]int{{0, 1}, {1, 2}, {0, 2}}, false}, true, false},
		{"unsatisfiable", Graph{3, 2, [][2]int{{0, 1}, {1, 2}, {0, 2}}, false}, false, false},
		{"path", Graph{4, 2, [][2]int{{0, 1}, {1, 2}, {2, 3}}, false}, false, true},
		{"cut", Graph{3, 3, [][2]int{{0, 1}, {1, 2}, {0, 2}}, true}, false, false},
		{"root_contradiction", Graph{2, 1, [][2]int{{0, 1}}, false}, false, false},
		{"root_solution", Graph{8, 1, nil, false}, false, false},
		{"eight_colors", Graph{1, 8, nil, false}, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			command, _ := EncodeLoad(tc.g)
			var encoded strings.Builder
			for _, b := range []byte(command) {
				fmt.Fprintf(&encoded, "%02x\n", b)
			}
			load := filepath.Join(dir, tc.name+".hex")
			if err := os.WriteFile(load, []byte(encoded.String()), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			negative, restart := 0, 0
			if tc.negative {
				negative = 1
			}
			if tc.restart {
				restart = 1
			}
			out, err := exec.CommandContext(ctx, "vvp", binary, "+load="+load, fmt.Sprintf("+negative=%d", negative), fmt.Sprintf("+restart=%d", restart)).CombinedOutput()
			if err != nil {
				t.Fatalf("simulation: %v\n%s", err, out)
			}
			if !strings.Contains(string(out), "DONE") {
				t.Fatalf("no completion: %s", out)
			}
			m := NewModel()
			_ = m.Load(ctx, tc.g)
			snapshot := Initial(tc.g)
			count := 0
			for _, line := range strings.Split(string(out), "\n") {
				if line == "RESET" {
					_ = m.Reset(ctx)
					snapshot = Initial(tc.g)
					continue
				}
				if !strings.HasPrefix(line, "E") {
					continue
				}
				e, err := DecodeEvent(line + "\n")
				if err != nil {
					t.Fatal(err, line)
				}
				want, err := m.Step(ctx)
				if err != nil {
					t.Fatal(err)
				}
				// Staging words outside these events are intentionally unspecified.
				if e.Kind != Create && e.Kind != Update {
					e.ChoiceWord = want.ChoiceWord
				}
				if e.Kind != Write && e.Kind != Restore {
					e.TrailWord = want.TrailWord
				}
				if e != want {
					t.Fatalf("event mismatch\ngot  %+v\nwant %+v", e, want)
				}
				snapshot, err = Apply(tc.g, snapshot, e)
				if err != nil {
					t.Fatal(err)
				}
				count++
			}
			if count == 0 || !snapshot.Terminal() {
				t.Fatal("missing terminal event")
			}
		})
	}
}
