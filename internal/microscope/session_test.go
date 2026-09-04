package microscope

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	domain "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/microscope"
	"golang.org/x/sync/errgroup"
)

func newTestSession(t *testing.T) *Session {
	t.Helper()
	s := NewSession(context.Background(), domain.NewModel(), "model", zerolog.Nop())
	t.Cleanup(func() { _ = s.Close() })
	return s
}
func TestSessionRunPauseResetAndHistory(t *testing.T) {
	s := newTestSession(t)
	ctx := context.Background()
	g := domain.Graph{Vertices: 4, Colors: 3}
	if err := s.Control(ctx, "step"); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	if err := s.Load(ctx, g); err != nil {
		t.Fatal(err)
	}
	if err := s.Control(ctx, "run"); err != nil {
		t.Fatal(err)
	}
	if err := s.Load(ctx, g); !errors.Is(err, ErrConflict) {
		t.Fatal("load while running", err)
	}
	deadline := time.Now().Add(time.Second)
	for s.State().Latest.Sequence == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if err := s.Control(ctx, "pause"); err != nil {
		t.Fatal(err)
	}
	seq := s.State().Latest.Sequence
	if seq == 0 {
		t.Fatal("run did not advance")
	}
	time.Sleep(40 * time.Millisecond)
	if s.State().Latest.Sequence != seq {
		t.Fatal("advanced after pause")
	}
	for s.State().Latest.Sequence < HistoryLimit+20 {
		if err := s.Control(ctx, "step"); err != nil {
			t.Fatal(err)
		}
	}
	state := s.State()
	if len(state.History) != HistoryLimit {
		t.Fatal("unbounded history")
	}
	if _, err := s.Event(1, state.Generation); err == nil {
		t.Fatal("old history retained")
	}
	recorded, err := s.Event(state.Latest.Sequence, state.Generation)
	if err != nil {
		t.Fatal(err)
	}
	recorded.Domains[0] = 0
	if len(recorded.Choices) > 0 {
		recorded.Choices[0].Remaining = 255
	}
	if s.State().Latest.Domains[0] == 0 {
		t.Fatal("snapshot alias")
	}
	generation := state.Generation
	if err := s.Control(ctx, "reset"); err != nil {
		t.Fatal(err)
	}
	if s.State().Generation != generation+1 || s.State().Latest.Sequence != 0 || len(s.State().History) != 0 {
		t.Fatal("reset state")
	}
	if _, err := s.Event(1, generation); !errors.Is(err, ErrConflict) {
		t.Fatal("stale generation", err)
	}
}
func TestSessionSerializesConcurrentControls(t *testing.T) {
	s := newTestSession(t)
	ctx := context.Background()
	_ = s.Load(ctx, domain.Graph{Vertices: 4, Colors: 3})
	group, ctx := errgroup.WithContext(ctx)
	for i := 0; i < 20; i++ {
		group.Go(func() error { _ = s.State(); return s.Control(ctx, "step") })
	}
	if err := group.Wait(); err != nil {
		t.Fatal(err)
	}
	if s.State().Latest.Sequence != 20 {
		t.Fatal("lost event")
	}
}

type failingEngine struct {
	*domain.Model
	fail bool
}

var _ domain.Engine = &failingEngine{}

func (e *failingEngine) Step(ctx context.Context) (domain.Event, error) {
	if e.fail {
		return domain.Event{}, errors.New("device disconnected")
	}
	return e.Model.Step(ctx)
}
func TestSessionErrorRequiresExplicitRecovery(t *testing.T) {
	engine := &failingEngine{domain.NewModel(), true}
	s := NewSession(context.Background(), engine, "serial", zerolog.Nop())
	defer s.Close()
	ctx := context.Background()
	_ = s.Load(ctx, domain.Graph{Vertices: 2, Colors: 2})
	if err := s.Control(ctx, "step"); err == nil {
		t.Fatal("missing device error")
	}
	if s.State().Error != "device disconnected" || s.State().Running {
		t.Fatal(s.State())
	}
	if err := s.Control(ctx, "run"); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	engine.fail = false
	if err := s.Control(ctx, "reset"); err != nil {
		t.Fatal(err)
	}
	if err := s.Control(ctx, "step"); err != nil {
		t.Fatal(err)
	}
}
