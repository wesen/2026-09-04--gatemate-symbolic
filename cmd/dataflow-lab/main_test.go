package main

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
)

type rowCapture struct {
	middlewares.Processor
	rows []types.Row
	err  error
}

func (p *rowCapture) AddRow(_ context.Context, r types.Row) error {
	p.rows = append(p.rows, r)
	return p.err
}

var _ middlewares.Processor = &rowCapture{}

type commandCapture struct {
	*Command
	settings  Settings
	processor rowCapture
}

func (c *commandCapture) RunIntoGlazeProcessor(ctx context.Context, v *values.Values, _ middlewares.Processor) error {
	if err := v.DecodeSectionInto(schema.DefaultSlug, &c.settings); err != nil {
		return err
	}
	return c.Command.RunIntoGlazeProcessor(ctx, v, &c.processor)
}
func TestExamplesAndOutputFlags(t *testing.T) {
	for _, example := range []string{"book", "copy", "fault", "cancel"} {
		t.Run(example, func(t *testing.T) {
			capture := &commandCapture{Command: NewCommand()}
			command, err := cli.BuildCobraCommandFromCommand(capture)
			if err != nil {
				t.Fatal(err)
			}
			for _, flag := range []string{"format", "output-fields", "max-output-rows", "engine", "example", "log-level"} {
				if command.Flags().Lookup(flag) == nil {
					t.Fatal(flag)
				}
			}
			command.SetArgs([]string{"--example", example, "--format", "json", "--log-level", "error"})
			if err := command.ExecuteContext(context.Background()); err != nil {
				t.Fatal(err)
			}
			if capture.settings.Example != example || capture.settings.Engine != "model" || len(capture.processor.rows) != 1 {
				t.Fatalf("settings=%+v rows=%d", capture.settings, len(capture.processor.rows))
			}
		})
	}
}
func TestCanceledExampleStopsBeforeEmission(t *testing.T) {
	capture := &commandCapture{Command: NewCommand()}
	command, err := cli.BuildCobraCommandFromCommand(capture)
	if err != nil {
		t.Fatal(err)
	}
	command.SetArgs([]string{"--log-level", "error"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Glazed deliberately normalizes context.Canceled to successful shutdown.
	if err := command.ExecuteContext(ctx); err != nil {
		t.Fatal(err)
	}
	if len(capture.processor.rows) != 0 {
		t.Fatal("canceled command emitted results")
	}
}
func TestProcessorFailurePropagates(t *testing.T) {
	sentinel := errors.New("sink failed")
	capture := &commandCapture{Command: NewCommand(), processor: rowCapture{err: sentinel}}
	command, err := cli.BuildCobraCommandFromCommand(capture)
	if err != nil {
		t.Fatal(err)
	}
	command.SetArgs([]string{"--log-level", "error"})
	if err := command.ExecuteContext(context.Background()); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
}
