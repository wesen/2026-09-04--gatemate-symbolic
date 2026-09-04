package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	service "github.com/wesen/2026-09-04--gatemate-symbolic/internal/microscope"
	domain "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/microscope"
	"golang.org/x/sync/errgroup"
)

type ServeCommand struct{ *cmds.CommandDescription }
type Settings struct {
	Listen   string `glazed:"listen"`
	Engine   string `glazed:"engine"`
	Device   string `glazed:"device"`
	LogLevel string `glazed:"log-level"`
}

var _ cmds.BareCommand = &ServeCommand{}

func NewServeCommand() *ServeCommand {
	return &ServeCommand{cmds.NewCommandDescription("search-microscope", cmds.WithShort("Inspect graph-coloring propagation and rollback on FPGA or model"), cmds.WithLong("Load graphs and control a single instrument through the local Go HTTP API and React UI. Serial mode uses real FPGA events; model mode is an explicitly labeled simulator. Build the frontend with make frontend before serving the UI."), cmds.WithFlags(
		fields.New("listen", fields.TypeString, fields.WithDefault("127.0.0.1:8086"), fields.WithHelp("HTTP listening address")),
		fields.New("engine", fields.TypeChoice, fields.WithChoices("model", "serial"), fields.WithDefault("model"), fields.WithHelp("Explicit execution source")),
		fields.New("device", fields.TypeString, fields.WithDefault("/dev/ttyACM0"), fields.WithHelp("FPGA UART device in serial mode")),
		fields.New("log-level", fields.TypeChoice, fields.WithChoices("debug", "info", "warn", "error"), fields.WithDefault("info"), fields.WithHelp("Structured logging level")),
	))}
}
func (c *ServeCommand) Run(ctx context.Context, parsed *values.Values) error {
	var settings Settings
	if err := parsed.DecodeSectionInto(schema.DefaultSlug, &settings); err != nil {
		return err
	}
	level, err := zerolog.ParseLevel(settings.LogLevel)
	if err != nil {
		return err
	}
	logger := zerolog.New(os.Stderr).Level(level).With().Timestamp().Logger()
	var engine domain.Engine = domain.NewModel()
	if settings.Engine == "serial" {
		device, err := domain.NewSerial(settings.Device)
		if err != nil {
			return err
		}
		engine = device
	}
	session := service.NewSession(ctx, engine, settings.Engine, logger)
	defer session.Close()
	server := &http.Server{Addr: settings.Listen, Handler: service.NewHandler(session), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	group, groupctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		logger.Info().Str("listen", settings.Listen).Str("engine", settings.Engine).Msg("search microscope started")
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})
	group.Go(func() error {
		<-groupctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	})
	return group.Wait()
}
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	command, err := cli.BuildCobraCommandFromCommand(NewServeCommand())
	if err == nil {
		err = command.ExecuteContext(ctx)
	}
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
