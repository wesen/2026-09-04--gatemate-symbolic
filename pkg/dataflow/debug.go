package dataflow

import "github.com/pkg/errors"

const TraceCapacity = 32
const (
	EventOperand byte = iota + 1
	EventIssue
	EventCompletion
	EventRoute
	EventOutput
	EventCancel
	EventStale
)

type DebugControl struct {
	Mask    byte `json:"mask"`
	Node    byte `json:"node"`
	Context byte `json:"context"`
	Resume  bool `json:"resume"`
	Clear   bool `json:"clear"`
}

func (d DebugControl) Validate() error {
	if d.Mask > 7 || d.Node != 255 && d.Node >= Nodes || d.Context != 255 && d.Context >= Contexts {
		return errors.New("debug mask must be 0..7, node 0..6 or 255, context 0..3 or 255")
	}
	return nil
}

type DebugEvent struct {
	Cycle uint32 `json:"cycle"`
	Kind  byte   `json:"kind"`
	Token Token  `json:"token"`
}
type DebugSnapshot struct {
	Halted    bool         `json:"halted"`
	Reason    byte         `json:"reason"`
	StopCycle uint32       `json:"stopCycle"`
	Mask      byte         `json:"mask"`
	Node      byte         `json:"node"`
	Context   byte         `json:"context"`
	Dropped   uint32       `json:"dropped"`
	Events    []DebugEvent `json:"events"`
}

func (m *Transaction) debugEvent(kind byte, t Token) {
	event := DebugEvent{m.Metrics.Cycles, kind, t}
	if len(m.Debug.Events) < TraceCapacity {
		m.Debug.Events = append(m.Debug.Events, event)
	} else {
		m.Debug.Dropped++
	}
	selected := (m.Debug.Node == 255 || m.Debug.Node == t.Node) && (m.Debug.Context == 255 || m.Debug.Context == t.Context)
	if kind == EventIssue && m.Debug.Mask&1 != 0 && selected {
		m.halt(1)
	}
	if kind == EventStale && m.Debug.Mask&4 != 0 {
		m.halt(4)
	}
}
func (m *Transaction) halt(reason byte) {
	m.Debug.Halted = true
	m.Debug.Reason |= reason
	m.Debug.StopCycle = m.Metrics.Cycles
}
func (m *Transaction) debugControl(d DebugControl) {
	m.Debug.Mask = d.Mask
	m.Debug.Node = d.Node
	m.Debug.Context = d.Context
	if d.Resume {
		m.Debug.Halted = false
		m.Debug.Reason = 0
	}
	if d.Clear {
		m.Debug.Events = []DebugEvent{}
		m.Debug.Dropped = 0
	}
}
func (d DebugSnapshot) Clone() DebugSnapshot {
	d.Events = append([]DebugEvent{}, d.Events...)
	return d
}
