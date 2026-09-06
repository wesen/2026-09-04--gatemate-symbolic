package main

import (
	"bytes"
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/compile"
	"os"
	"path/filepath"
	"testing"
)

func TestCompileCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.lazy")
	if err := os.WriteFile(path, []byte("def main : Int = 42;"), 0600); err != nil {
		t.Fatal(err)
	}
	c := NewCommand()
	var out bytes.Buffer
	c.Output = &out
	cmd, err := cli.BuildCobraCommandFromCommand(c)
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs([]string{"--source", path, "--action", "compile"})
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var a compile.Artifact
	if err = json.Unmarshal(out.Bytes(), &a); err != nil {
		t.Fatal(err)
	}
	if err = a.Validate(); err != nil {
		t.Fatal(err)
	}
}
