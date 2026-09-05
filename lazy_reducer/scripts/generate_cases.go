//go:build ignore

// Generate model-qualified graph vectors for the independent RTL controller.
package main

import (
	"context"
	"fmt"
	l "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazy"
	"math/rand"
	"os"
	"reflect"
)

func main() {
	f, e := os.Create("build/generated_cases.svh")
	if e != nil {
		panic(e)
	}
	defer f.Close()
	r := rand.New(rand.NewSource(9010))
	ctx := context.Background()
	for trial := 0; trial < 120; trial++ {
		nodes := []l.Word{l.Integer(int32(r.Intn(100001) - 50000)), l.Integer(int32(r.Intn(100001) - 50000))}
		for n := 2; n < 24; n++ {
			a, b := uint16(r.Intn(n)), uint16(r.Intn(n))
			switch r.Intn(5) {
			case 0:
				nodes = append(nodes, l.Node(l.Thunk, a, 0))
			case 1:
				nodes = append(nodes, l.Node(l.Ind, a, 0))
			case 2:
				nodes = append(nodes, l.Node(l.Mul, a, b))
			default:
				nodes = append(nodes, l.Node(l.Add, a, b))
			}
		}
		i := l.Image{Nodes: nodes, Root: 23}
		m := l.NewModelWithStack(8)
		_, e = m.Execute(ctx, l.Operation{Kind: "load", Image: &i})
		if e != nil {
			panic(e)
		}
		_, _ = m.Execute(ctx, l.Operation{Kind: "force", Root: i.Root})
		_, _ = m.Execute(ctx, l.Operation{Kind: "tick", Ticks: 10000})
		s, _ := m.Snapshot(ctx)
		want, heap := l.Reference(i, 8)
		if !s.Valid || s.Result != want || !reflect.DeepEqual(s.Heap, heap) {
			panic("model/reference disagreement")
		}
		fmt.Fprintf(f, "reset(); // generated graph %d\n", trial)
		for a, w := range nodes {
			fmt.Fprintf(f, "load(%d,40'h%010x);\n", a, w)
		}
		fmt.Fprintf(f, "force_node(23);finish(40'h%010x);\n", want)
		for a, w := range heap {
			fmt.Fprintf(f, "if(dut.heap.mem[%d]!==40'h%010x)$fatal(1,\"graph %d heap %d\");\n", a, w, trial, a)
		}
		fmt.Fprintf(f, "if(dut.counters[3]!=%d||dut.counters[4]!=%d||dut.counters[5]!=%d||dut.counters[6]!=%d)$fatal(1,\"graph %d counters\");\n", s.Counters.Claims, s.Counters.Updates, s.Counters.Muls, s.Counters.Adds, trial)
		for n, event := range s.Trace {
			fmt.Fprintf(f, "if(dut.trace_ram.mem[%d][79:0]!==80'h%010x%010x)$fatal(1,\"graph %d mutation %d\");\n", n, event.Old, event.New, trial, n)
		}
	}
	fmt.Fprintln(f, "$display(\"PASS 120 RTL graphs against recursive reference and transaction heap/mutation/counter results\");")
}
