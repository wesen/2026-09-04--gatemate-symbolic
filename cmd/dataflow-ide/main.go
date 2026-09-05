package main

import (
	"context"
	"net"
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
	ide "github.com/wesen/2026-09-04--gatemate-symbolic/internal/dataflowide"
	df "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/dataflow"
	"golang.org/x/sync/errgroup"
)

type Command struct{ *cmds.CommandDescription }
type Settings struct {
	Listen   string `glazed:"listen"`
	Engine   string `glazed:"engine"`
	Device   string `glazed:"device"`
	Projects string `glazed:"projects"`
	LogLevel string `glazed:"log-level"`
}

var _ cmds.BareCommand = &Command{}

func NewCommand() *Command {
	return &Command{cmds.NewCommandDescription("dataflow-ide", cmds.WithShort("Run a local dataflow scenario IDE; resets the selected engine on startup"), cmds.WithFlags(
		fields.New("listen", fields.TypeString, fields.WithDefault("127.0.0.1:8087")),
		fields.New("engine", fields.TypeChoice, fields.WithChoices("model", "serial"), fields.WithDefault("model")),
		fields.New("device", fields.TypeString, fields.WithDefault("/dev/ttyACM0")),
		fields.New("projects", fields.TypeString, fields.WithDefault("elastic_dataflow/projects")),
		fields.New("log-level", fields.TypeChoice, fields.WithChoices("debug", "info", "warn", "error"), fields.WithDefault("info")),
	))}
}
func (c *Command) Run(ctx context.Context, v *values.Values) error {
	var settings Settings
	if err := v.DecodeSectionInto(schema.DefaultSlug, &settings); err != nil {
		return err
	}
	host, _, err := net.SplitHostPort(settings.Listen)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return errors.New("the device IDE must listen on a loopback address")
	}
	level, err := zerolog.ParseLevel(settings.LogLevel)
	if err != nil {
		return err
	}
	logger := zerolog.New(os.Stderr).Level(level).With().Timestamp().Logger()
	projects, err := ide.NewProjects(settings.Projects)
	if err != nil {
		return err
	}
	defer projects.Close()
	var engine df.Engine
	if settings.Engine == "serial" {
		engine, err = df.NewSerial(settings.Device)
	} else {
		engine, err = df.NewTransaction(df.DefaultConfig())
	}
	if err != nil {
		return err
	}
	session, err := ide.NewSession(ctx, engine, logger)
	if err != nil {
		_ = engine.Close()
		return err
	}
	defer session.Close()
	server := &http.Server{Addr: settings.Listen, Handler: ide.NewHandler(session, projects), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	group, gctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		logger.Info().Str("listen", settings.Listen).Str("engine", settings.Engine).Msg("dataflow IDE started")
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})
	group.Go(func() error {
		<-gctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	})
	return group.Wait()
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
