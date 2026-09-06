package compile_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	c "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/compile"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/ir"
)

func artifact(t *testing.T, source string) *c.Artifact {
	t.Helper()
	a, ds := c.Compile(context.Background(), source)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	return a
}
func TestExamplesDeterministic(t *testing.T) {
	files, err := filepath.Glob("../../../examples/lazylang/*.lazy")
	if err != nil || len(files) != 6 {
		t.Fatal(files, err)
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			a := artifact(t, string(src))
			b := artifact(t, string(src))
			if !reflect.DeepEqual(a, b) {
				t.Fatal("nondeterministic artifact")
			}
			data, err := json.Marshal(a)
			if err != nil {
				t.Fatal(err)
			}
			var restored c.Artifact
			if err = json.Unmarshal(data, &restored); err != nil {
				t.Fatal(err)
			}
			if err = restored.Validate(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestSharedLayout(t *testing.T) {
	a := artifact(t, `def main : Int = let x : Int = (fun (n : Int) -> n * 2) 21 in (x + x) + (x + x);`)
	if a.Root != 15 || a.ConstantEnd != 15 || len(a.Heap) != 17 || len(a.Code) != 14 {
		t.Fatalf("%+v", a)
	}
	for i, want := range []ir.Object{{Tag: ir.Int, Payload: 2}, {Tag: ir.Int, Payload: 21}, {Tag: ir.Thunk, A: 13, B: 16}, {Tag: ir.Env, A: 15, B: ir.Absent}} {
		got, err := ir.ParseObject(a.Heap[13+i])
		if err != nil || got != want {
			t.Fatalf("heap %d: %+v want %+v: %v", 13+i, got, want, err)
		}
	}
	wantOps := []ir.Opcode{ir.Var, ir.Const, ir.Prim, ir.Lambda, ir.Const, ir.App, ir.Var, ir.Var, ir.Prim, ir.Var, ir.Var, ir.Prim, ir.Prim, ir.Let}
	for i, want := range wantOps {
		code, err := ir.ParseCode(a.Code[i])
		if err != nil || code.Op != want {
			t.Fatalf("code %d: %+v %v", i, code, err)
		}
		span := a.Spans[code.Span]
		if span.Start >= span.End {
			t.Fatalf("missing source span at %d", i)
		}
	}
}
func TestRejectCorruption(t *testing.T) {
	for name, mutate := range map[string]func(*c.Artifact){
		"digest":       func(a *c.Artifact) { a.Source += " " },
		"code cycle":   func(a *c.Artifact) { a.Code[0] = (ir.Code{Op: ir.Lambda, A: 0}).Hex(); a.ID = a.Digest() },
		"pinned error": func(a *c.Artifact) { a.Heap[0] = (ir.Object{Tag: ir.Int}).Hex(); a.ID = a.Digest() },
		"env cycle": func(a *c.Artifact) {
			i := len(a.Heap) - 1
			a.Heap[i] = (ir.Object{Tag: ir.Env, A: a.Root, B: uint16(i)}).Hex()
			a.ID = a.Digest()
		},
		"span":     func(a *c.Artifact) { a.Spans[0].End = len(a.Source) + 1; a.ID = a.Digest() },
		"encoding": func(a *c.Artifact) { a.Code[0] = "xyz"; a.ID = a.Digest() },
	} {
		t.Run(name, func(t *testing.T) {
			a := artifact(t, `def main : Int = 1;`)
			mutate(a)
			if a.Validate() == nil {
				t.Fatal("corruption accepted")
			}
		})
	}
	if a, ds := c.Compile(context.Background(), `def main : Int = true;`); a != nil || len(ds) == 0 {
		t.Fatal("type error accepted")
	}
	var src strings.Builder
	for i := 0; i < 1100; i++ {
		src.WriteString("def ")
		src.WriteString(strings.Repeat("x", i/26+1))
		src.WriteByte(byte('a' + i%26))
		src.WriteString(" : Int = 0;\n")
	}
	src.WriteString("def main : Int = 0;")
	if a, ds := c.Compile(context.Background(), src.String()); a != nil || len(ds) == 0 || ds[0].Code != "ARTIFACT_LIMIT" {
		t.Fatal("heap profile limit", ds)
	}
}
