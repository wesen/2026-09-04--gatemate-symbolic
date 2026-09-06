package lazylanguageide

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/compile"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/ir"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/machine"
	wire "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/serial"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/syntax"
)

type Stream struct {
	Cursor     uint16  `json:"cursor"`
	Phase      string  `json:"phase"`
	Pending    bool    `json:"pending"`
	SentDemand bool    `json:"sentDemand"`
	Values     []int32 `json:"values"`
	Complete   bool    `json:"complete"`
	Fault      uint32  `json:"fault"`
}
type ObjectView struct {
	Address   uint16 `json:"address"`
	Tag       string `json:"tag"`
	A         uint16 `json:"a"`
	B         uint16 `json:"b"`
	Payload   uint32 `json:"payload"`
	Integer   int32  `json:"integer"`
	Span      uint16 `json:"span"`
	Committed bool   `json:"committed"`
}
type CodeView struct {
	ID        uint16 `json:"id"`
	Op        string `json:"op"`
	A         uint16 `json:"a"`
	B         uint16 `json:"b"`
	C         uint16 `json:"c"`
	Immediate uint32 `json:"immediate"`
	Span      uint16 `json:"span"`
	Type      string `json:"type"`
}
type ArtifactView struct {
	Artifact *compile.Artifact `json:"artifact"`
	Code     []CodeView        `json:"code"`
}
type Frame struct {
	ID         uint64            `json:"id"`
	RunID      string            `json:"runId"`
	ArtifactID string            `json:"artifactId"`
	Backend    string            `json:"backend"`
	Operation  string            `json:"operation"`
	NeedsReset bool              `json:"needsReset"`
	Error      string            `json:"error"`
	Snapshot   *machine.Snapshot `json:"snapshot"`
	Heap       []ObjectView      `json:"heap"`
	Stream     *Stream           `json:"stream"`
	Polled     *uint16           `json:"polled"`
}
type HistoryItem struct {
	ID         uint64 `json:"id"`
	RunID      string `json:"runId"`
	ArtifactID string `json:"artifactId"`
	Operation  string `json:"operation"`
}
type State struct {
	Frame   Frame         `json:"frame"`
	History []HistoryItem `json:"history"`
}
type Operation struct {
	Kind       string `json:"kind"`
	ExpectedID uint64 `json:"expectedId"`
	RunID      string `json:"runId"`
	ArtifactID string `json:"artifactId,omitempty"`
	Ref        uint16 `json:"ref,omitempty"`
	Ticks      uint32 `json:"ticks,omitempty"`
}
type CompileResult struct {
	ClientRevision uint64              `json:"clientRevision"`
	ArtifactID     string              `json:"artifactId"`
	Diagnostics    []syntax.Diagnostic `json:"diagnostics"`
}
type Session struct {
	mu            sync.Mutex
	model         *machine.Machine
	serial        *wire.Client
	logger        zerolog.Logger
	artifacts     map[string]*compile.Artifact
	artifactOrder []string
	history       []Frame
	current       Frame
	stream        *Stream
	streamTail    uint16
}

func runID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
func NewSession(device *wire.Client, logger zerolog.Logger) *Session {
	backend := "model"
	if device != nil {
		backend = "serial"
	}
	return &Session{serial: device, logger: logger, artifacts: map[string]*compile.Artifact{}, history: []Frame{}, current: Frame{ID: 1, RunID: runID(), Backend: backend, Operation: "startup", Heap: []ObjectView{}}}
}
func (s *Session) Close() error {
	if s.serial != nil {
		return s.serial.Close()
	}
	return nil
}
func (s *Session) Compile(ctx context.Context, source string, revision uint64) CompileResult {
	a, ds := compile.Compile(ctx, source)
	result := CompileResult{ClientRevision: revision, Diagnostics: ds}
	if a == nil {
		return result
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.artifacts[a.ID]; !exists {
		// Keep artifacts referenced by history; only evict unreferenced compilations.
		if len(s.artifacts) >= 160 {
			for index, id := range s.artifactOrder {
				used := s.current.ArtifactID == id
				for _, f := range s.history {
					used = used || f.ArtifactID == id
				}
				if !used {
					delete(s.artifacts, id)
					s.artifactOrder = append(s.artifactOrder[:index], s.artifactOrder[index+1:]...)
					break
				}
			}
		}
		if len(s.artifacts) >= 160 {
			result.Diagnostics = []syntax.Diagnostic{{Code: "ARTIFACT_CACHE_FULL", Message: "artifact cache is full; start a new IDE session"}}
			return result
		}
		s.artifacts[a.ID] = a
		s.artifactOrder = append(s.artifactOrder, a.ID)
	}
	result.ArtifactID = a.ID
	return result
}
func (s *Session) Artifact(id string) (ArtifactView, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.artifacts[id]
	if !ok {
		return ArtifactView{}, false
	}
	// Compile artifacts are immutable internally. Copy all slices before exposing.
	b := *a
	b.Code = append([]string{}, a.Code...)
	b.Heap = append([]string{}, a.Heap...)
	b.Provenance = append([]uint16{}, a.Provenance...)
	b.Spans = append([]syntax.Span{}, a.Spans...)
	b.Bindings = append(b.Bindings[:0:0], a.Bindings...)
	b.Uses = append(b.Uses[:0:0], a.Uses...)
	b.CodeTypes = append([]string{}, a.CodeTypes...)
	out := ArtifactView{Artifact: &b, Code: []CodeView{}}
	ops := []string{"CONST", "VAR", "LAMBDA", "APP", "LET", "LETREC", "PRIM", "IF", "CONS", "CASE"}
	for i, packed := range a.Code {
		c, _ := ir.ParseCode(packed)
		out.Code = append(out.Code, CodeView{ID: uint16(i), Op: ops[c.Op], A: c.A, B: c.B, C: c.C, Immediate: c.Immediate, Span: c.Span, Type: a.CodeTypes[i]})
	}
	return out, true
}
func cloneFrame(f Frame) Frame {
	f.Heap = append([]ObjectView{}, f.Heap...)
	if f.Snapshot != nil {
		v := *f.Snapshot
		v.Heap = append([]string{}, v.Heap...)
		v.Provenance = append([]uint16{}, v.Provenance...)
		v.Stack = append([]string{}, v.Stack...)
		v.Trace = append([]machine.Event{}, v.Trace...)
		f.Snapshot = &v
	}
	if f.Stream != nil {
		v := *f.Stream
		v.Values = append([]int32{}, v.Values...)
		f.Stream = &v
	}
	if f.Polled != nil {
		v := *f.Polled
		f.Polled = &v
	}
	return f
}
func (s *Session) state() State {
	out := State{Frame: cloneFrame(s.current), History: []HistoryItem{}}
	for _, f := range s.history {
		out.History = append(out.History, HistoryItem{f.ID, f.RunID, f.ArtifactID, f.Operation})
	}
	return out
}
func (s *Session) State() State { s.mu.Lock(); defer s.mu.Unlock(); return s.state() }
func (s *Session) History(id uint64) (Frame, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.history {
		if f.ID == id {
			return cloneFrame(f), true
		}
	}
	return Frame{}, false
}
func (s *Session) demand(ctx context.Context, ref uint16) error {
	if s.serial != nil {
		return s.serial.Demand(ctx, ref)
	}
	return s.model.Demand(ref)
}
func (s *Session) tick(ctx context.Context, n uint32) error {
	if s.serial != nil {
		return s.serial.Tick(ctx, n)
	}
	return s.model.Tick(ctx, n)
}
func (s *Session) poll(ctx context.Context) (uint16, bool, error) {
	if s.serial != nil {
		return s.serial.Poll(ctx)
	}
	r, ok := s.model.Poll()
	return r, ok, nil
}
func (s *Session) object(ctx context.Context, ref uint16) (ir.Object, error) {
	if s.serial != nil {
		return s.serial.ReadObject(ctx, ref)
	}
	snap := s.model.Snapshot()
	if int(ref) >= len(snap.Heap) {
		return ir.Object{}, errors.New("object reference outside heap")
	}
	return ir.ParseObject(snap.Heap[ref])
}
func (s *Session) snapshot(ctx context.Context) (machine.Snapshot, error) {
	if s.serial != nil {
		return s.serial.Snapshot(ctx)
	}
	return s.model.Snapshot(), nil
}
func (s *Session) advanceStream(ctx context.Context, budget uint32) error {
	stream := s.stream
	for stream.Pending {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !stream.SentDemand {
			if budget == 0 {
				return nil
			}
			if err := s.demand(ctx, stream.Cursor); err != nil {
				return err
			}
			stream.SentDemand = true
		}
		ref, valid, err := s.poll(ctx)
		if err != nil {
			return err
		}
		if !valid {
			if budget == 0 {
				return nil
			}
			n := uint32(64)
			if n > budget {
				n = budget
			}
			if err = s.tick(ctx, n); err != nil {
				return err
			}
			budget -= n
			continue
		}
		stream.SentDemand = false
		o, err := s.object(ctx, ref)
		if err != nil {
			return err
		}
		if o.Tag == ir.Error {
			stream.Fault = o.Payload
			stream.Pending = false
			stream.Complete = true
			return nil
		}
		if stream.Phase == "NeedConstructor" {
			if o.Tag == ir.Nil {
				stream.Complete = true
				stream.Pending = false
				return nil
			}
			if o.Tag != ir.Cons {
				return errors.New("stream constructor is not CONS or NIL")
			}
			stream.Cursor = o.A
			stream.Phase = "NeedHead"
			s.streamTail = o.B
		} else {
			if o.Tag != ir.Int {
				return errors.New("stream head is not INT")
			}
			stream.Values = append(stream.Values, int32(o.Payload))
			stream.Cursor = s.streamTail
			stream.Phase = "NeedConstructor"
			stream.Pending = false
		}
	}
	return nil
}
func (s *Session) Control(ctx context.Context, op Operation) (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if op.ExpectedID != s.current.ID || op.RunID != s.current.RunID {
		return s.state(), errors.New("stale frame or run: refresh state before acting")
	}
	if op.Ticks > 1000000 {
		return s.state(), errors.New("tick budget exceeds 1000000")
	}
	if s.current.NeedsReset && op.Kind != "reset" && op.Kind != "load" {
		return s.state(), errors.New("reset required before further execution")
	}
	if op.Kind != "load" && op.Kind != "reset" && s.current.ArtifactID == "" {
		return s.state(), errors.New("load a compiled artifact first")
	}
	if s.stream != nil && op.Kind != "stream-next" && op.Kind != "stream-resume" && op.Kind != "reset" && op.Kind != "load" {
		return s.state(), errors.New("stream session owns execution; reset to abort")
	}
	var err error
	var polled *uint16
	artifactID := s.current.ArtifactID
	nextRun := s.current.RunID
	switch op.Kind {
	case "load":
		a, ok := s.artifacts[op.ArtifactID]
		if !ok {
			return s.state(), errors.New("unknown artifact")
		}
		nextRun = runID()
		s.stream = nil
		if s.serial != nil {
			err = s.serial.Load(ctx, a)
		} else {
			s.model, err = machine.New(a)
		}
		artifactID = a.ID
	case "reset":
		nextRun = runID()
		s.stream = nil
		s.model = nil
		artifactID = ""
		if s.serial != nil {
			err = s.serial.Reset(ctx)
		}
	case "force":
		err = s.demand(ctx, op.Ref)
	case "tick":
		err = s.tick(ctx, op.Ticks)
	case "poll":
		var ref uint16
		var valid bool
		ref, valid, err = s.poll(ctx)
		if valid {
			polled = &ref
		}
	case "stream-start":
		if s.artifacts[artifactID].EntryType != "ListInt" {
			return s.state(), errors.New("stream requires ListInt entry")
		}
		s.stream = &Stream{Cursor: s.artifacts[artifactID].Root, Phase: "NeedConstructor", Values: []int32{}}
	case "stream-next":
		if s.stream == nil || s.stream.Pending || s.stream.Complete {
			return s.state(), errors.New("stream next requires an idle unfinished stream")
		}
		s.stream.Pending = true
		err = s.advanceStream(ctx, op.Ticks)
	case "stream-resume":
		if s.stream == nil || !s.stream.Pending {
			return s.state(), errors.New("no pending stream demand")
		}
		err = s.advanceStream(ctx, op.Ticks)
	default:
		return s.state(), errors.New("unknown operation")
	}
	f := Frame{ID: s.current.ID + 1, RunID: nextRun, ArtifactID: artifactID, Backend: s.current.Backend, Operation: op.Kind, Heap: []ObjectView{}, Polled: polled, Stream: s.stream}
	if err == nil && artifactID != "" {
		var snap machine.Snapshot
		snap, err = s.snapshot(ctx)
		if err == nil {
			f.Snapshot = &snap
			tags := map[ir.Tag]string{ir.Int: "INT", ir.Bool: "BOOL", ir.Nil: "NIL", ir.Cons: "CONS", ir.Fun: "FUN", ir.Thunk: "THUNK", ir.Ind: "IND", ir.Blackhole: "BLACKHOLE", ir.Env: "ENV", ir.Error: "ERROR", ir.Free: "FREE"}
			for i, packed := range snap.Heap {
				o, e := ir.ParseObject(packed)
				if e != nil {
					err = e
					break
				}
				f.Heap = append(f.Heap, ObjectView{Address: uint16(i), Tag: tags[o.Tag], A: o.A, B: o.B, Payload: o.Payload, Integer: int32(o.Payload), Span: snap.Provenance[i], Committed: i < int(snap.CommittedTop)})
			}
		}
	}
	if err != nil {
		f.NeedsReset = true
		f.Error = err.Error()
		s.logger.Error().Err(err).Str("operation", op.Kind).Msg("language operation failed")
	} else {
		s.logger.Debug().Str("operation", op.Kind).Uint64("frame", f.ID).Msg("language operation complete")
	}
	s.current = cloneFrame(f)
	s.history = append(s.history, cloneFrame(f))
	if len(s.history) > 128 {
		s.history = s.history[len(s.history)-128:]
	}
	return s.state(), err
}
