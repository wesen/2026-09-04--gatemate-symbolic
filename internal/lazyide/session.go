package lazyide

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	l "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazy"
	"sync"
)

type Frame struct {
	ID       uint64     `json:"id"`
	Label    string     `json:"label"`
	Snapshot l.Snapshot `json:"snapshot"`
}
type State struct {
	Current    Frame    `json:"current"`
	Frames     []Frame  `json:"frames"`
	Results    []l.Word `json:"results"`
	NeedsReset bool     `json:"needsReset"`
}
type Session struct {
	mu         sync.Mutex
	engine     l.Engine
	logger     zerolog.Logger
	frames     []Frame
	results    []l.Word
	next       uint64
	needsReset bool
}

func NewSession(ctx context.Context, e l.Engine, logger zerolog.Logger) (*Session, error) {
	s := &Session{engine: e, logger: logger, results: []l.Word{}}
	if _, err := e.Execute(ctx, l.Operation{Kind: "reset"}); err != nil {
		return nil, err
	}
	if err := s.capture(ctx, "startup reset"); err != nil {
		return nil, err
	}
	return s, nil
}
func (s *Session) Close() error { return s.engine.Close() }
func clone(f Frame) Frame {
	v := &f.Snapshot
	v.Heap = append([]l.Word{}, v.Heap...)
	v.Stack = append([]l.Frame{}, v.Stack...)
	v.Trace = append([]l.Mutation{}, v.Trace...)
	return f
}
func (s *Session) state() State {
	frames := make([]Frame, len(s.frames))
	for n, f := range s.frames {
		frames[n] = clone(f)
	}
	return State{clone(s.frames[len(s.frames)-1]), frames, append([]l.Word{}, s.results...), s.needsReset}
}
func (s *Session) State() State { s.mu.Lock(); defer s.mu.Unlock(); return s.state() }
func (s *Session) capture(ctx context.Context, label string) error {
	snap, e := s.engine.Snapshot(ctx)
	if e != nil {
		s.needsReset = true
		return e
	}
	s.next++
	s.frames = append(s.frames, Frame{s.next, label, snap})
	if len(s.frames) > 128 {
		s.frames = append([]Frame{}, s.frames[len(s.frames)-128:]...)
	}
	return nil
}
func (s *Session) Control(ctx context.Context, expected uint64, o l.Operation) (State, error) {
	if e := o.Validate(); e != nil {
		return State{}, e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if expected != s.frames[len(s.frames)-1].ID {
		return State{}, errors.New("stale frame: return to live state")
	}
	if s.needsReset && o.Kind != "reset" && o.Kind != "load" {
		return State{}, errors.New("engine uncertain: reset required")
	}
	if o.Kind == "force" && s.frames[len(s.frames)-1].Snapshot.State != "idle" {
		return State{}, errors.New("finish execution and poll before forcing again")
	}
	v, e := s.engine.Execute(ctx, o)
	if e != nil {
		s.needsReset = true
		s.logger.Error().Err(e).Str("operation", o.Kind).Msg("lazy control failed")
		return State{}, e
	}
	previousFrames := s.frames
	if o.Kind == "reset" || o.Kind == "load" {
		s.frames = nil
		s.results = []l.Word{}
		s.needsReset = false
	}
	if v != nil {
		s.results = append(s.results, *v)
		if len(s.results) > 128 {
			s.results = append([]l.Word{}, s.results[len(s.results)-128:]...)
		}
	}
	if e = s.capture(ctx, o.Kind); e != nil {
		s.frames = previousFrames
		return State{}, e
	}
	s.logger.Debug().Str("operation", o.Kind).Uint64("frame", s.next).Msg("lazy state captured")
	return s.state(), nil
}
