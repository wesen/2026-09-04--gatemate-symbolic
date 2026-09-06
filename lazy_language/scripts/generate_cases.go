//go:build ignore

// Generate compiled program vectors and independently model-qualified results.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/compile"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/ir"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/machine"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	must(os.MkdirAll("build", 0755))
	f, err := os.Create("build/generated_cases.svh")
	must(err)
	defer f.Close()
	sources := []string{
		`def main : Int = let x : Int = (fun (n : Int) -> n * 2) 21 in (x + x) + (x + x);`,
		`def main : Int = -2147483648 * (-1);`,
		`def main : Int = (-3) * 7;`,
		`def main : Int = 2147483647 + 1;`,
		`def main : Int = -2147483648 - 1;`,
		`def main : Bool = 42 == 42;`,
		`def main : Bool = -1 <= 0;`,
		`def main : Int = if false then 1 else 2;`,
		`def main : Int = case Nil of { Nil -> 5; Cons(h, t) -> h };`,
	}
	files, err := filepath.Glob("../examples/lazylang/*.lazy")
	must(err)
	for _, path := range files {
		b, err := os.ReadFile(path)
		must(err)
		sources = append(sources, string(b))
	}
	for trial, source := range sources {
		a, ds := compile.Compile(context.Background(), source)
		if len(ds) > 0 {
			panic(ds)
		}
		m, err := machine.New(a)
		must(err)
		fmt.Fprintf(f, "reset(); // compiled program %d\n", trial)
		fmt.Fprintf(f, "load(0,0,128'h%016x%04x%04x%04x%04x);\n", 0, len(a.Code), len(a.Heap), a.Root, a.ConstantEnd)
		for i, c := range a.Code {
			fmt.Fprintf(f, "load(1,%d,128'h%s);\n", i, c)
		}
		for i, h := range a.Heap {
			fmt.Fprintf(f, "load(2,%d,128'h%s);\n", i, h)
		}
		for i, p := range a.Provenance {
			fmt.Fprintf(f, "load(3,%d,128'h%04x);\n", i, p)
		}
		fmt.Fprintln(f, "load(4,0,0);")
		demand := func(ref uint16) ir.Object {
			must(m.Demand(ref))
			for i := 0; i < 100000; i++ {
				must(m.Tick(context.Background(), 1))
				if m.Snapshot().Valid {
					break
				}
			}
			r, ok := m.Poll()
			if !ok {
				panic("model timeout")
			}
			snap := m.Snapshot()
			o, err := ir.ParseObject(snap.Heap[r])
			must(err)
			fmt.Fprintf(f, "force_node(%d);finish(%d,80'h%s);poll();\n", ref, r, o.Hex())
			for i, h := range snap.Heap {
				fmt.Fprintf(f, "if(dut.heap_ram.mem[%d]!==80'h%s)$fatal(1,\"program %d heap %d\");\n", i, h, trial, i)
			}
			c := snap.Counters
			counts := []uint32{c.Allocations, c.AllocatedThunks, c.AllocatedFunctions, c.AllocatedEnvironments, c.AllocatedCons, c.AllocatedIntegers, c.Claims, c.Updates, c.Adds, c.Subtracts, c.Multiplies, c.Equalities, c.ComparisonsLE, c.Indirections, c.EnvironmentSteps, c.MaxStack, c.MaxHeap}
			for i, count := range counts {
				fmt.Fprintf(f, "if(dut.counters[%d]!==32'd%d)$fatal(1,\"program %d counter %d got %%d want %d\",dut.counters[%d]);\n", i+3, count, trial, i+3, count, i+3)
			}
			return o
		}
		o := demand(a.Root)
		if o.Tag == ir.Cons {
			for i := 0; i < 8; i++ {
				demand(o.A)
				if i < 7 {
					o = demand(o.B)
					if o.Tag != ir.Cons {
						break
					}
				}
			}
		}
	}
	fmt.Fprintf(f, "$display(\"PASS %d compiled programs, full heaps and semantic counters\");\n", len(sources))
}
