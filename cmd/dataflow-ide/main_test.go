package main

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"testing"
)

type capture struct {
	*cmds.CommandDescription
	settings Settings
}

func (c *capture) Run(_ context.Context, v *values.Values) error {
	return v.DecodeSectionInto(schema.DefaultSlug, &c.settings)
}

var _ cmds.BareCommand = &capture{}

func TestIDEFlags(t *testing.T) {
	c := &capture{CommandDescription: NewCommand().CommandDescription}
	root, err := cli.BuildCobraCommandFromCommand(c)
	if err != nil {
		t.Fatal(err)
	}
	root.SetArgs([]string{"--engine", "serial", "--listen", "127.0.0.1:9998", "--projects", "/tmp/example-projects", "--device", "/dev/example", "--log-level", "debug"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if c.settings != (Settings{"127.0.0.1:9998", "serial", "/dev/example", "/tmp/example-projects", "debug"}) {
		t.Fatal(c.settings)
	}
}
