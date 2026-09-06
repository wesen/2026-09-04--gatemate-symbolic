package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	ide "github.com/wesen/2026-09-04--gatemate-symbolic/internal/lazylanguageide"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/check"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/compile"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/semantics"
	wire "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/serial"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/syntax"
)

type Command struct {
	*cmds.CommandDescription
	Output io.Writer
}
type Settings struct {
	Source   string `glazed:"source"`
	Action   string `glazed:"action"`
	Engine   string `glazed:"engine"`
	Device   string `glazed:"device"`
	Prefix   int    `glazed:"prefix"`
	Ticks    int    `glazed:"ticks"`
	LogLevel string `glazed:"log-level"`
}

var _ cmds.BareCommand = &Command{}

func NewCommand() *Command {
	return &Command{CommandDescription: cmds.NewCommandDescription("lazy-language", cmds.WithShort("Compile, observe or execute typed LFL1 source and emit JSON"), cmds.WithFlags(
		fields.New("source", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Source .lazy file")),
		fields.New("action", fields.TypeChoice, fields.WithChoices("compile", "reference", "run"), fields.WithDefault("run")),
		fields.New("engine", fields.TypeChoice, fields.WithChoices("model", "serial"), fields.WithDefault("model")),
		fields.New("device", fields.TypeString, fields.WithDefault("/dev/ttyACM0")),
		fields.New("prefix", fields.TypeInteger, fields.WithDefault(8)),
		fields.New("ticks", fields.TypeInteger, fields.WithDefault(100000)),
		fields.New("log-level", fields.TypeChoice, fields.WithChoices("debug", "info", "warn", "error"), fields.WithDefault("info")),
	)), Output: os.Stdout}
}
func (c *Command) Run(ctx context.Context, v *values.Values) error {
	var settings Settings
	if err := v.DecodeSectionInto(schema.DefaultSlug, &settings); err != nil {
		return err
	}
	if settings.Prefix < 0 || settings.Prefix > 2048 || settings.Ticks < 1 || settings.Ticks > 1000000 {
		return errors.New("prefix must be 0..2048 and ticks 1..1000000")
	}
	f, err := os.Open(settings.Source)
	if err != nil {
		return errors.Wrap(err, "open source")
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, syntax.MaxSourceBytes+1))
	if err != nil {
		return err
	}
	a, ds := compile.Compile(ctx, string(data))
	if len(ds) > 0 {
		_ = json.NewEncoder(c.Output).Encode(map[string]any{"diagnostics": ds})
		return errors.New("source compilation failed")
	}
	if settings.Action == "compile" {
		return json.NewEncoder(c.Output).Encode(a)
	}
	if settings.Action == "reference" {
		ast, _ := syntax.Parse(ctx, string(data))
		p, _ := check.Check(ctx, ast)
		out, err := semantics.Observe(ctx, p, settings.Prefix, uint64(settings.Ticks))
		if err != nil {
			return err
		}
		return json.NewEncoder(c.Output).Encode(out)
	}
	level, err := zerolog.ParseLevel(settings.LogLevel)
	if err != nil {
		return err
	}
	logger := zerolog.New(os.Stderr).Level(level).With().Timestamp().Logger()
	var device *wire.Client
	if settings.Engine == "serial" {
		device, err = wire.Open(settings.Device)
		if err != nil {
			return err
		}
	}
	session := ide.NewSession(device, logger)
	defer session.Close()
	compiled := session.Compile(ctx, string(data), 1)
	act := func(kind string) error {
		state := session.State()
		_, err := session.Control(ctx, ide.Operation{Kind: kind, ExpectedID: state.Frame.ID, RunID: state.Frame.RunID, ArtifactID: compiled.ArtifactID, Ref: a.Root, Ticks: uint32(settings.Ticks)})
		return err
	}
	if err = act("load"); err != nil {
		return err
	}
	if a.EntryType == "ListInt" {
		if err = act("stream-start"); err != nil {
			return err
		}
		for i := 0; i < settings.Prefix; i++ {
			if err = act("stream-next"); err != nil {
				return err
			}
			stream := session.State().Frame.Stream
			if stream.Pending || stream.Complete {
				break
			}
		}
	} else {
		if err = act("force"); err != nil {
			return err
		}
		if err = act("tick"); err != nil {
			return err
		}
	}
	return json.NewEncoder(c.Output).Encode(session.State())
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
