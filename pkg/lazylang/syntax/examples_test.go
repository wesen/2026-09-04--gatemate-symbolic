package syntax

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRepositoryExamples(t *testing.T) {
	paths, err := filepath.Glob("../../../examples/lazylang/*.lazy")
	if err != nil || len(paths) != 6 {
		t.Fatalf("examples: %v %v", paths, err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			p := parsed(t, string(data))
			last := p.Definitions[len(p.Definitions)-1]
			if last.Name.Name != "main" {
				t.Fatal(last.Name)
			}
			checkProgramSpans(t, string(data), p)
		})
	}
}
func ExampleParse() {
	program, diagnostics := Parse(context.Background(), "def main : Int = let x : Int = 21 * 2 in x + x;")
	fmt.Println(len(diagnostics))
	fmt.Println(program.Definitions[0].Name.Name)
	fmt.Printf("%T\n", program.Definitions[0].Value)
	// Output:
	// 0
	// main
	// *syntax.LetExpr
}
