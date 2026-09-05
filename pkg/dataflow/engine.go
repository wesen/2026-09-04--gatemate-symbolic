package dataflow

import (
	"context"
	"github.com/pkg/errors"
)

var ErrFull = errors.New("input queue is full; tick before retrying")
var ErrBlocked = errors.New("cancellation blocked by offered output or epoch wrap drain")

// Engine operations and snapshots require one external owner. Serial also
// serializes wire exchanges, including all pages of a snapshot.
type Engine interface {
	Execute(context.Context, Operation) (*Token, error)
	Snapshot(context.Context) (Snapshot, error)
	Close() error
}
type Operation struct {
	Debug   *DebugControl `json:"debug,omitempty"`
	Kind    string        `json:"kind"`
	Graph   *Graph        `json:"graph,omitempty"`
	Token   *Token        `json:"token,omitempty"`
	Context byte          `json:"context,omitempty"`
	Ticks   uint32        `json:"ticks,omitempty"`
	Config  *Config       `json:"config,omitempty"`
}

func (o Operation) Validate() error {
	if o.Graph != nil && o.Kind != "load" {
		return errors.New("graph is only valid on load")
	}
	if o.Debug != nil && o.Kind != "debug" {
		return errors.New("debug control is only valid on debug")
	}
	if o.Config != nil && o.Kind != "reset" {
		return errors.New("configuration is only valid on reset")
	}

	switch o.Kind {
	case "debug":
		if o.Debug == nil {
			return errors.New("debug requires control")
		}
		return o.Debug.Validate()
	case "load":
		if o.Graph == nil {
			return errors.New("load requires graph")
		}
		return o.Graph.Validate()
	case "reset":
		if o.Config != nil {
			_, err := NewTransaction(*o.Config)
			return err
		}
	case "inject":
		if o.Token == nil {
			return errors.New("inject requires token")
		}
		_, err := o.Token.Bytes()
		return err
	case "cancel":
		if o.Context >= Contexts {
			return errors.New("context must be 0..3")
		}
	case "tick":
		if o.Ticks > 1_000_000 {
			return errors.New("tick count must be at most 1000000")
		}
	case "poll":
	default:
		return errors.New("unknown engine operation")
	}
	if o.Config != nil && o.Kind != "reset" {
		return errors.New("configuration is only valid on reset")
	}
	return nil
}

type SlotSnapshot struct {
	Context byte     `json:"context"`
	Node    byte     `json:"node"`
	Values  [2]Value `json:"values"`
	Valid   byte     `json:"valid"`
	Issued  bool     `json:"issued"`
	Pending bool     `json:"pending"`
}
type IssueSnapshot struct {
	Token  Token    `json:"token"`
	Phase  int      `json:"phase"`
	Values [2]Value `json:"values"`
}
type RouterSnapshot struct {
	Token     Token `json:"token"`
	Delivered byte  `json:"delivered"`
}
type Snapshot struct {
	Debug      DebugSnapshot     `json:"debug"`
	Graph      Graph             `json:"graph"`
	Source     string            `json:"source"`
	Config     Config            `json:"config"`
	Epochs     [Contexts]byte    `json:"epochs"`
	Closed     [Contexts]bool    `json:"closed"`
	Slots      []SlotSnapshot    `json:"slots"`
	Issue      *IssueSnapshot    `json:"issue"`
	Router     *RouterSnapshot   `json:"router"`
	Mul        []*Token          `json:"mul"`
	ALU        []*Token          `json:"alu"`
	Input      []Token           `json:"input"`
	Completion []Token           `json:"completion"`
	Output     []Token           `json:"output"`
	Errors     [Contexts]*Token  `json:"errors"`
	Counters   map[string]uint32 `json:"counters"`
	Quiescent  bool              `json:"quiescent"`
}

var _ Engine = &Transaction{}

func (m *Transaction) Execute(ctx context.Context, o Operation) (*Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := o.Validate(); err != nil {
		return nil, err
	}
	switch o.Kind {
	case "debug":
		m.debugControl(*o.Debug)
	case "load":
		return nil, m.LoadGraph(*o.Graph)
	case "reset":
		c := m.Config
		if o.Config != nil {
			c = *o.Config
		}
		fresh, err := NewTransaction(c)
		if err != nil {
			return nil, err
		}
		*m = *fresh
	case "inject":
		if !m.Inject(*o.Token) {
			return nil, ErrFull
		}
	case "cancel":
		if !m.Cancel(o.Context) {
			return nil, ErrBlocked
		}
	case "tick":
		for i := uint32(0); i < o.Ticks && !m.Debug.Halted; i++ {
			if err := m.Tick(ctx); err != nil {
				return nil, err
			}
		}
	case "poll":
		return m.Poll(), nil
	}
	return nil, nil
}
func cloneTokens(in []*Token) []*Token {
	out := make([]*Token, len(in))
	for i, t := range in {
		if t != nil {
			copy := *t
			out[i] = &copy
		}
	}
	return out
}
func (m *Transaction) Snapshot(ctx context.Context) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	s := Snapshot{Debug: m.Debug.Clone(), Graph: m.graph(), Source: "model", Config: m.Config, Epochs: m.Epoch, Closed: m.Closed, Quiescent: m.Quiescent(),
		Mul: cloneTokens(m.mul), ALU: cloneTokens(m.alu), Input: append([]Token{}, m.input...), Completion: append([]Token{}, m.completed...), Output: append([]Token{}, m.output...),
		Counters: map[string]uint32{"cycles": m.Metrics.Cycles, "source": m.Metrics.Source, "activations": m.Metrics.Activations, "mul": m.Metrics.Mul, "alu": m.Metrics.ALU,
			"stale":      m.Metrics.StaleInput + m.Metrics.StaleIssue + m.Metrics.StaleCompletion + m.Metrics.StaleRouter + m.Metrics.StaleOutput,
			"duplicates": m.Metrics.Duplicates, "invalid": m.Metrics.Invalid, "unitBusy": m.Metrics.UnitBusy, "unitBlocked": m.Metrics.UnitBlocked, "routerBlocked": m.Metrics.RouterBlocked,
			"inputHigh": uint32(m.Metrics.InputHigh), "readyHigh": uint32(m.Metrics.ReadyHigh), "completionHigh": uint32(m.Metrics.CompletionHigh), "outputHigh": uint32(m.Metrics.OutputHigh)}}
	for c := 0; c < Contexts; c++ {
		for n := 0; n < Nodes; n++ {
			v := m.Slots[c][n]
			s.Slots = append(s.Slots, SlotSnapshot{byte(c), byte(n), v.Values, v.Valid, v.Issued, v.Pending})
		}
		if m.pendingError[c] != nil {
			copy := *m.pendingError[c]
			s.Errors[c] = &copy
		}
	}
	if m.issue != nil {
		s.Issue = &IssueSnapshot{Token: m.completion(m.issue.Context, m.issue.Epoch, m.issue.Node, 0), Phase: m.issue.Wait + 1, Values: m.issue.Values}
	}
	if m.router != nil {
		s.Router = &RouterSnapshot{m.router.Token, m.router.Delivered}
	}
	return s, nil
}
func (m *Transaction) Close() error { return nil }

// Clone detaches all mutable snapshot containers for history and API consumers.
func (s Snapshot) Clone() Snapshot {
	s.Debug = s.Debug.Clone()
	s.Slots = append([]SlotSnapshot{}, s.Slots...)
	s.Input = append([]Token{}, s.Input...)
	s.Completion = append([]Token{}, s.Completion...)
	s.Output = append([]Token{}, s.Output...)
	s.Mul = cloneTokens(s.Mul)
	s.ALU = cloneTokens(s.ALU)
	counters := map[string]uint32{}
	for k, v := range s.Counters {
		counters[k] = v
	}
	s.Counters = counters
	for c, t := range s.Errors {
		if t != nil {
			copy := *t
			s.Errors[c] = &copy
		}
	}
	if s.Issue != nil {
		copy := *s.Issue
		s.Issue = &copy
	}
	if s.Router != nil {
		copy := *s.Router
		s.Router = &copy
	}
	return s
}
