package dataflowide

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	df "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/dataflow"
	"golang.org/x/sync/errgroup"
)

var ErrConflict = errors.New("state changed or scenario running; refresh the live state")
var ErrAssertion = errors.New("scenario assertion failed")

const HistoryLimit = 128

type FrameInfo struct {
	ID         uint64 `json:"id"`
	Generation uint64 `json:"generation"`
	Label      string `json:"label"`
	Cycle      uint32 `json:"cycle"`
	At         string `json:"at"`
}
type Frame struct {
	Program *df.Program `json:"program"`
	FrameInfo
	Snapshot df.Snapshot `json:"snapshot"`
}
type Event struct {
	Label  string `json:"label"`
	Detail string `json:"detail"`
	At     string `json:"at"`
	Error  bool   `json:"error"`
}
type State struct {
	Current    Frame       `json:"current"`
	History    []FrameInfo `json:"history"`
	Results    []df.Token  `json:"results"`
	Events     []Event     `json:"events"`
	Running    bool        `json:"running"`
	Scenario   string      `json:"scenario"`
	Step       int         `json:"step"`
	Actions    int         `json:"actions"`
	Error      string      `json:"error"`
	NeedsReset bool        `json:"needsReset"`
}
type Session struct {
	program            *df.Program
	mu                 sync.Mutex
	engine             df.Engine
	ctx                context.Context
	cancel             context.CancelFunc
	group              *errgroup.Group
	logger             zerolog.Logger
	frames             []Frame
	results            []df.Token
	events             []Event
	nextID, generation uint64
	running            bool
	runCancel          context.CancelFunc
	scenario           Scenario
	scenarioHash       [32]byte
	step               int
	lastError          string
	needsReset         bool
}

func NewSession(ctx context.Context, e df.Engine, logger zerolog.Logger) (*Session, error) {
	life, cancel := context.WithCancel(ctx)
	group, gctx := errgroup.WithContext(life)
	s := &Session{engine: e, ctx: gctx, cancel: cancel, group: group, logger: logger, results: []df.Token{}, events: []Event{}}
	if _, err := s.executeLocked(ctx, df.Operation{Kind: "reset"}, "session reset"); err != nil {
		cancel()
		return nil, err
	}
	return s, nil
}
func (s *Session) Close() error { s.cancel(); _ = s.group.Wait(); return s.engine.Close() }
func (s *Session) stateLocked() State {
	state := State{Results: append([]df.Token{}, s.results...), Events: append([]Event{}, s.events...), History: []FrameInfo{}, Running: s.running, Scenario: s.scenario.Name, Step: s.step, Actions: len(s.scenario.Actions), Error: s.lastError, NeedsReset: s.needsReset}
	if len(s.frames) > 0 {
		state.Current = s.frames[len(s.frames)-1]
		state.Current.Snapshot = state.Current.Snapshot.Clone()
		state.Current.Program = state.Current.Program.Clone()
	}
	for _, f := range s.frames {
		state.History = append(state.History, f.FrameInfo)
	}
	return state
}
func (s *Session) State() State { s.mu.Lock(); defer s.mu.Unlock(); return s.stateLocked() }
func (s *Session) Historical(id, generation uint64) (Frame, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.frames {
		if f.ID == id && f.Generation == generation {
			f.Snapshot = f.Snapshot.Clone()
			f.Program = f.Program.Clone()
			return f, true
		}
	}
	return Frame{}, false
}
func (s *Session) eventLocked(label, detail string, failed bool) {
	s.events = append(s.events, Event{label, detail, time.Now().UTC().Format(time.RFC3339Nano), failed})
	if len(s.events) > 128 {
		s.events = s.events[len(s.events)-128:]
	}
}
func (s *Session) captureLocked(ctx context.Context, label string) error {
	snap, err := s.engine.Snapshot(ctx)
	if err != nil {
		s.needsReset = true
		return err
	}
	s.nextID++
	s.frames = append(s.frames, Frame{Program: s.program.Clone(), FrameInfo: FrameInfo{s.nextID, s.generation, label, snap.Counters["cycles"], time.Now().UTC().Format(time.RFC3339Nano)}, Snapshot: snap})
	if len(s.frames) > HistoryLimit {
		s.frames = s.frames[len(s.frames)-HistoryLimit:]
	}
	return nil
}
func (s *Session) executeLocked(ctx context.Context, o df.Operation, label string) (*df.Token, error) {
	if s.needsReset && o.Kind != "reset" {
		return nil, errors.New("engine state uncertain; reset required")
	}
	token, err := s.engine.Execute(ctx, o)
	if err != nil {
		if !errors.Is(err, df.ErrFull) && !errors.Is(err, df.ErrBlocked) {
			s.needsReset = true
		}
		return nil, err
	}
	if o.Kind == "load" {
		s.program = nil
	}
	if o.Kind == "reset" {
		s.program = nil
		s.generation++
		s.frames = nil
		s.results = []df.Token{}
		s.events = []Event{}
		s.needsReset = false
		s.lastError = ""
	}
	if token != nil {
		s.results = append(s.results, *token)
		if len(s.results) > 256 {
			s.results = s.results[len(s.results)-256:]
		}
	}
	if err := s.captureLocked(ctx, label); err != nil {
		return token, err
	}
	return token, nil
}
func (s *Session) guardLocked(expected uint64) error {
	if s.running || len(s.frames) == 0 || s.frames[len(s.frames)-1].ID != expected {
		return ErrConflict
	}
	return nil
}
func (s *Session) Control(ctx context.Context, expected uint64, o df.Operation) error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.Config != nil && (o.Config.InputDepth > 8 || o.Config.CompletionDepth > 8 || o.Config.OutputDepth > 8) {
		return errors.New("IDE queue depths must be 1..8")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.guardLocked(expected); err != nil {
		return err
	}
	_, err := s.executeLocked(ctx, o, o.Kind)
	s.scenarioHash = [32]byte{}
	s.step = 0
	if err != nil {
		s.lastError = err.Error()
		s.eventLocked(o.Kind, err.Error(), true)
		return err
	}
	s.lastError = ""
	s.eventLocked(o.Kind, "operation completed", false)
	return nil
}
func (s *Session) actionLocked(ctx context.Context, a Action) error {
	label := fmt.Sprintf("action %d: %s", s.step+1, a.Kind)
	if a.Kind == "expect" {
		for n := len(s.results) - 1; n >= 0; n-- {
			r := s.results[n]
			if r.Context == a.Context && (a.Epoch == nil || r.Epoch == *a.Epoch) {
				if r.Value != *a.Value {
					return errors.Wrapf(ErrAssertion, "context %d: got %d, expected %d", a.Context, r.Value, *a.Value)
				}
				s.eventLocked(label, fmt.Sprintf("context %d value %d passed", a.Context, *a.Value), false)
				return s.captureLocked(ctx, label)
			}
		}
		return errors.Wrapf(ErrAssertion, "no polled result for context %d", a.Context)
	}
	if a.Kind == "inputs" {
		if s.needsReset {
			return errors.New("engine state uncertain; reset required")
		}
		epoch := s.frames[len(s.frames)-1].Snapshot.Epochs[a.Context]
		for _, token := range df.Inputs(a.Context, epoch, [6]int32(a.Values)) {
			for attempts := 0; ; attempts++ {
				_, err := s.engine.Execute(ctx, df.Operation{Kind: "inject", Token: &token})
				if err == nil {
					break
				}
				if !errors.Is(err, df.ErrFull) || attempts >= 1000 {
					s.needsReset = true
					return errors.Wrap(err, "input action may be partially applied; reset before rerunning")
				}
				if _, err := s.engine.Execute(ctx, df.Operation{Kind: "tick", Ticks: 1}); err != nil {
					s.needsReset = true
					return err
				}
			}
		}
		return s.captureLocked(ctx, label+" (includes credit ticks if needed)")
	}
	_, err := s.executeLocked(ctx, df.Operation{Kind: a.Kind, Context: a.Context, Token: a.Token, Ticks: a.Ticks, Config: a.Config}, label)
	return err
}
func (s *Session) setScenarioLocked(scenario Scenario) {
	s.scenario = scenario
	s.scenarioHash = sha256.Sum256(canonicalScenario(scenario))
	s.step = 0
	s.lastError = ""
}
func (s *Session) StepScenario(ctx context.Context, expected uint64, scenario Scenario) error {
	if err := validateScenario(scenario); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.guardLocked(expected); err != nil {
		return err
	}
	hash := sha256.Sum256(canonicalScenario(scenario))
	if s.scenarioHash != hash || s.step >= len(s.scenario.Actions) {
		s.setScenarioLocked(scenario)
	}
	err := s.actionLocked(ctx, s.scenario.Actions[s.step])
	if err != nil {
		s.lastError = err.Error()
		s.eventLocked("scenario stopped", err.Error(), true)
		return err
	}
	if s.scenario.Actions[s.step].Kind != "expect" {
		s.eventLocked(fmt.Sprintf("action %d: %s", s.step+1, s.scenario.Actions[s.step].Kind), "completed", false)
	}
	s.step++
	return nil
}
func (s *Session) RunScenario(expected uint64, scenario Scenario) error {
	if err := validateScenario(scenario); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.guardLocked(expected); err != nil {
		return err
	}
	s.setScenarioLocked(scenario)
	runctx, cancel := context.WithCancel(s.ctx)
	s.runCancel = cancel
	s.running = true
	s.group.Go(func() error {
		defer cancel()
		for {
			s.mu.Lock()
			if err := runctx.Err(); err != nil {
				s.running = false
				s.eventLocked("paused", "execution stopped at an action boundary", false)
				s.mu.Unlock()
				return nil
			}
			if s.step >= len(s.scenario.Actions) {
				s.running = false
				s.eventLocked("scenario complete", s.scenario.Name, false)
				s.mu.Unlock()
				return nil
			}
			err := s.actionLocked(runctx, s.scenario.Actions[s.step])
			if err != nil {
				s.running = false
				s.lastError = err.Error()
				s.eventLocked("scenario stopped", err.Error(), true)
				s.logger.Warn().Err(err).Msg("scenario stopped")
				s.mu.Unlock()
				return nil
			}
			if s.scenario.Actions[s.step].Kind != "expect" {
				s.eventLocked(fmt.Sprintf("action %d: %s", s.step+1, s.scenario.Actions[s.step].Kind), "completed", false)
			}
			s.step++
			s.mu.Unlock()
			// Yield between actions so pause and state requests can observe boundaries.
			timer := time.NewTimer(15 * time.Millisecond)
			select {
			case <-runctx.Done():
				timer.Stop()
			case <-timer.C:
			}
		}
	})
	return nil
}
func (s *Session) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.runCancel != nil {
		s.runCancel()
	}
}
