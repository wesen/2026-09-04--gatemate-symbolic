package microscope

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	domain "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/microscope"
	"golang.org/x/sync/errgroup"
)

var ErrConflict = errors.New("session conflict")

const HistoryLimit = 256

type Summary struct {
	Sequence  uint32 `json:"sequence"`
	Kind      byte   `json:"kind"`
	Name      string `json:"name"`
	Count     uint32 `json:"count"`
	ChoiceTop int    `json:"choiceTop"`
	TrailTop  int    `json:"trailTop"`
}
type State struct {
	Engine     string          `json:"engine"`
	Loaded     bool            `json:"loaded"`
	Graph      domain.Graph    `json:"graph"`
	Generation uint64          `json:"generation"`
	Running    bool            `json:"running"`
	Error      string          `json:"error"`
	Latest     domain.Snapshot `json:"latest"`
	History    []Summary       `json:"history"`
}
type Session struct {
	mu      sync.Mutex
	engine  domain.Engine
	state   State
	history []domain.Snapshot
	cancel  context.CancelFunc
	group   *errgroup.Group
	logger  zerolog.Logger
	closed  bool
}

func NewSession(ctx context.Context, engine domain.Engine, kind string, logger zerolog.Logger) *Session {
	ctx, cancel := context.WithCancel(ctx)
	group, ctx := errgroup.WithContext(ctx)
	s := &Session{engine: engine, state: State{Engine: kind, History: []Summary{}}, cancel: cancel, group: group, logger: logger}
	group.Go(func() error {
		ticker := time.NewTicker(25 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				s.mu.Lock()
				if s.state.Running && !s.closed {
					_ = s.stepLocked(ctx)
				}
				s.mu.Unlock()
			}
		}
	})
	return s
}
func (s *Session) failLocked(err error) error {
	s.state.Running = false
	s.state.Error = err.Error()
	s.logger.Error().Err(err).Msg("instrument operation failed")
	return err
}
func (s *Session) initializeLocked(g domain.Graph) {
	g.Edges = append([][2]int{}, g.Edges...)
	s.state.Graph = g
	s.state.Loaded = true
	s.state.Generation++
	s.state.Running = false
	s.state.Error = ""
	s.state.Latest = domain.Initial(g)
	s.history = nil
}
func (s *Session) Load(ctx context.Context, g domain.Graph) error {
	if _, err := g.Adjacency(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.state.Running {
		return errors.Wrap(ErrConflict, "pause before loading a graph")
	}
	if err := s.engine.Load(ctx, g); err != nil {
		return s.failLocked(err)
	}
	s.initializeLocked(g)
	s.logger.Info().Int("vertices", g.Vertices).Int("colors", g.Colors).Msg("graph loaded")
	return nil
}
func (s *Session) stepLocked(ctx context.Context) error {
	e, err := s.engine.Step(ctx)
	if err != nil {
		return s.failLocked(err)
	}
	snapshot, err := domain.Apply(s.state.Graph, s.state.Latest, e)
	if err != nil {
		return s.failLocked(err)
	}
	s.state.Latest = snapshot
	if len(s.history) == HistoryLimit {
		copy(s.history, s.history[1:])
		s.history = s.history[:HistoryLimit-1]
	}
	s.history = append(s.history, snapshot.Clone())
	if e.Terminal() {
		s.state.Running = false
		s.logger.Info().Str("terminal", snapshot.Name).Uint32("solutions", e.Count).Msg("search terminated")
	}
	return nil
}
func (s *Session) Control(ctx context.Context, action string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.Wrap(ErrConflict, "session is closed")
	}
	if action == "pause" {
		s.state.Running = false
		return nil
	}
	if !s.state.Loaded {
		return errors.Wrap(ErrConflict, "load a graph first")
	}
	if s.state.Running {
		return errors.Wrap(ErrConflict, "pause before changing execution")
	}
	if action == "reset" {
		if err := s.engine.Reset(ctx); err != nil {
			return s.failLocked(err)
		}
		s.initializeLocked(s.state.Graph)
		return nil
	}
	if s.state.Error != "" {
		return errors.Wrap(ErrConflict, "reset or reload after an instrument error")
	}
	if s.state.Latest.Terminal() {
		return errors.Wrap(ErrConflict, "search is terminal; reset to run again")
	}
	switch action {
	case "step":
		return s.stepLocked(ctx)
	case "run":
		s.state.Running = true
		return nil
	default:
		return errors.New("unknown control action")
	}
}
func (s *Session) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.state
	state.Graph.Edges = append([][2]int{}, s.state.Graph.Edges...)
	state.Latest = s.state.Latest.Clone()
	state.History = make([]Summary, 0, len(s.history))
	for _, e := range s.history {
		state.History = append(state.History, Summary{e.Sequence, e.Kind, e.Name, e.Count, e.ChoiceTop, e.TrailTop})
	}
	return state
}
func (s *Session) Event(sequence uint32, generation uint64) (domain.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if generation != s.state.Generation {
		return domain.Snapshot{}, errors.Wrap(ErrConflict, "history belongs to a different run")
	}
	for _, snapshot := range s.history {
		if snapshot.Sequence == sequence {
			return snapshot.Clone(), nil
		}
	}
	return domain.Snapshot{}, errors.New("event is not retained")
}
func (s *Session) Close() error {
	s.cancel()
	_ = s.group.Wait()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	s.state.Running = false
	return s.engine.Close()
}
