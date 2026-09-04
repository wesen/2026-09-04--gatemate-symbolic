package main

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

type captureCommand struct {
	*cmds.CommandDescription
	settings Settings
}

var _ cmds.BareCommand = &captureCommand{}

func (c *captureCommand) Run(ctx context.Context, parsed *values.Values) error {
	return parsed.DecodeSectionInto(schema.DefaultSlug, &c.settings)
}
func TestCLIFieldsDecode(t *testing.T) {
	capture := &captureCommand{CommandDescription: NewServeCommand().CommandDescription}
	root, err := cli.BuildCobraCommandFromCommand(capture)
	if err != nil {
		t.Fatal(err)
	}
	root.SetArgs([]string{"--listen", "127.0.0.1:9999", "--engine", "serial", "--device", "/dev/test", "--log-level", "debug"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if capture.settings != (Settings{"127.0.0.1:9999", "serial", "/dev/test", "debug"}) {
		t.Fatal(capture.settings)
	}
	for _, name := range []string{"listen", "engine", "device", "log-level"} {
		if root.Flags().Lookup(name) == nil {
			t.Fatal(name)
		}
	}
}
