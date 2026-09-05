package dataflow

import (
	"bufio"
	"encoding/hex"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var rtlLogDir = flag.String("rtl-log-dir", "", "Directory containing dataflow RTL stress logs for differential validation")

func TestRTLDifferential(t *testing.T) {
	if *rtlLogDir == "" {
		t.Skip("run RTL stress script then pass -rtl-log-dir")
	}
	files, err := filepath.Glob(filepath.Join(*rtlLogDir, "depth-*-latency-*.log"))
	if err != nil || len(files) != 12 {
		t.Fatalf("need twelve RTL matrix logs: %d %v", len(files), err)
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			var source, outputs, activations []Token
			cases := 0
			active := false
			passed := false
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				fields := strings.Fields(line)
				if len(fields) == 0 {
					continue
				}
				switch fields[0] {
				case "CASE":
					if active {
						t.Fatal("unterminated case")
					}
					active = true
					source = nil
					outputs = nil
					activations = nil
				case "SRC", "OUT", "ACT":
					if !active || len(fields) != 2 {
						t.Fatalf("invalid record %q", line)
					}
					p, err := hex.DecodeString(fields[1])
					if err != nil {
						t.Fatal(err)
					}
					token, err := DecodeToken(p)
					if err != nil {
						t.Fatal(err)
					}
					switch fields[0] {
					case "SRC":
						source = append(source, token)
					case "OUT":
						outputs = append(outputs, token)
					case "ACT":
						activations = append(activations, token)
					}
				case "END":
					if !active || len(source) != 24 || len(outputs) != 4 || len(activations) != 24 {
						t.Fatalf("incomplete case src=%d out=%d act=%d", len(source), len(outputs), len(activations))
					}
					m := newModel(t, DefaultConfig())
					want := run(t, m, source, int64(cases+1))
					expected := map[byte]Token{}
					for _, token := range want {
						expected[token.Context] = token
					}
					for _, token := range outputs {
						if expected[token.Context] != token {
							t.Fatalf("RTL result %+v expected %+v", token, expected[token.Context])
						}
						delete(expected, token.Context)
					}
					if len(expected) != 0 {
						t.Fatal("missing result")
					}
					key := func(c, e, n byte) [3]byte { return [3]byte{c, e, n} }
					expectedActs := map[[3]byte]Value{}
					for _, a := range m.Trace {
						expectedActs[key(a.Context, a.Epoch, a.Node)] = a.Value
					}
					for _, a := range activations {
						k := key(a.Context, a.Epoch, a.Node)
						v, ok := expectedActs[k]
						if !ok || v != a.Value {
							t.Fatalf("RTL activation %+v expected value=%x present=%v", a, v, ok)
						}
						delete(expectedActs, k)
					}
					if len(expectedActs) != 0 {
						t.Fatal("missing activation")
					}
					active = false
					cases++
				case "PASS":
					passed = true
				case "FATAL:":
					t.Fatal(line)
				}
			}
			if err := scanner.Err(); err != nil {
				t.Fatal(err)
			}
			if active || cases != 50 || !passed {
				t.Fatalf("incomplete run: cases=%d active=%v passed=%v", cases, active, passed)
			}
		})
	}
}
