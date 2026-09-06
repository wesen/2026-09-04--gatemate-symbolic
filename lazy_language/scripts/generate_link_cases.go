//go:build ignore

package main

import (
	"context"
	"fmt"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/compile"
	"os"
)

func main() {
	a, ds := compile.Compile(context.Background(), `def main : Int = let x : Int = (fun (n : Int) -> n * 2) 21 in (x + x) + (x + x);`)
	if len(ds) > 0 {
		panic(ds)
	}
	f, err := os.Create("build/link_cases.svh")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fmt.Fprintf(f, "request(\"B\",8,144'h%04x%04x%04x%04x,0);ack();\n", len(a.Code), len(a.Heap), a.Root, a.ConstantEnd)
	fmt.Fprintln(f, `request("K",0,0,0);rejected(); request("C",18,144'h000100000000000000000000000000000000,0);rejected();`)
	for i, c := range a.Code {
		fmt.Fprintf(f, "request(\"C\",18,144'h%04x%s,0);ack();\n", i, c)
	}
	for i, h := range a.Heap {
		fmt.Fprintf(f, "request(\"H\",12,144'h%04x%s,0);ack();\n", i, h)
	}
	for i, p := range a.Provenance {
		fmt.Fprintf(f, "request(\"V\",4,144'h%04x%04x,0);ack();\n", i, p)
	}
	for i, c := range a.Code {
		fmt.Fprintf(f, "request(\"Q\",2,144'h%04x,0);if(record!==128'h%s)$fatal(1,\"code readback %d\");\n", 0x2000+i, c, i)
	}
	for i, h := range a.Heap {
		fmt.Fprintf(f, "request(\"Q\",2,144'h%04x,0);if(record[79:0]!==80'h%s)$fatal(1,\"heap readback %d\");\n", 0x1000+i, h, i)
	}
	for i, p := range a.Provenance {
		fmt.Fprintf(f, "request(\"Q\",2,144'h%04x,0);if(record[15:0]!==16'h%04x)$fatal(1,\"provenance readback %d\");\n", 0x4000+i, p, i)
	}
	fmt.Fprintln(f, `request("K",0,0,0);ack();`)
}
