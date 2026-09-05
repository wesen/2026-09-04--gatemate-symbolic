package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/dataflow"
)

type Command struct{ *cmds.CommandDescription }
type Settings struct {
	Engine   string `glazed:"engine"`
	Device   string `glazed:"device"`
	Example  string `glazed:"example"`
	LogLevel string `glazed:"log-level"`
	WireLog  string `glazed:"wire-log"`
}

var _ cmds.GlazeCommand = &Command{}

func NewCommand() *Command {
	return &Command{cmds.NewCommandDescription("dataflow-lab", cmds.WithShort("Reset and verify an elastic dataflow example on the model or FPGA"), cmds.WithFlags(
		fields.New("engine", fields.TypeChoice, fields.WithChoices("model", "serial"), fields.WithDefault("model")),
		fields.New("device", fields.TypeString, fields.WithDefault("/dev/ttyACM0")),
		fields.New("example", fields.TypeChoice, fields.WithChoices("book", "copy", "fault", "cancel"), fields.WithDefault("book")),
		fields.New("log-level", fields.TypeChoice, fields.WithChoices("debug", "info", "warn", "error"), fields.WithDefault("info")),
		fields.New("wire-log", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Optional file for serial request/response evidence")),
	))}
}
func (c *Command) RunIntoGlazeProcessor(ctx context.Context, v *values.Values, p middlewares.Processor) error {
	var settings Settings
	if err := v.DecodeSectionInto(schema.DefaultSlug, &settings); err != nil {
		return err
	}
	level, err := zerolog.ParseLevel(settings.LogLevel)
	if err != nil {
		return err
	}
	logger := zerolog.New(os.Stderr).Level(level).With().Timestamp().Logger()
	var engine dataflow.Engine
	if settings.Engine == "serial" {
		s, err := dataflow.NewSerial(settings.Device)
		if err != nil {
			return err
		}
		engine = s
		if settings.WireLog != "" {
			f, err := os.Create(settings.WireLog)
			if err != nil {
				_ = engine.Close()
				return err
			}
			defer f.Close()
			s.Capture = f
		}
	} else {
		engine, err = dataflow.NewTransaction(dataflow.DefaultConfig())
		if err != nil {
			return err
		}
	}
	defer engine.Close()
	logger.Info().Str("engine", settings.Engine).Str("example", settings.Example).Msg("resetting engine and checking example")
	outputs, snapshot, err := dataflow.RunExample(ctx, engine, settings.Example)
	if err != nil {
		return err
	}
	return p.AddRow(ctx, types.NewRow(types.MRP("example", settings.Example), types.MRP("source", settings.Engine), types.MRP("outputs", outputs), types.MRP("snapshot", snapshot)))
}
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	command, err := cli.BuildCobraCommandFromCommand(NewCommand())
	if err == nil {
		err = command.ExecuteContext(ctx)
	}
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
