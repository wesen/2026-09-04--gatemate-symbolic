package lazyide

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	l "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazy"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSessionOwnershipAndRepeatForce(t *testing.T) {
	ctx := context.Background()
	s, e := NewSession(ctx, l.NewModel(), zerolog.Nop())
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	i, _ := l.Example("shared")
	state, e := s.Control(ctx, s.State().Current.ID, l.Operation{Kind: "load", Image: &i})
	if e != nil {
		t.Fatal(e)
	}
	loaded := state.Current.ID
	state, e = s.Control(ctx, loaded, l.Operation{Kind: "force", Root: 6})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Control(ctx, loaded, l.Operation{Kind: "tick", Ticks: 10}); e == nil {
		t.Fatal("stale frame accepted")
	}
	state, e = s.Control(ctx, state.Current.ID, l.Operation{Kind: "tick", Ticks: 1000})
	if e != nil || state.Current.Snapshot.Result != 168 {
		t.Fatal(e)
	}
	state.Current.Snapshot.Heap[3] = 0
	state.Frames[0].Snapshot.Heap[0] = 0
	if s.State().Current.Snapshot.Heap[3] != 42 {
		t.Fatal("mutable frame alias")
	}
	state, e = s.Control(ctx, s.State().Current.ID, l.Operation{Kind: "poll"})
	if e != nil || len(state.Results) != 1 || state.Results[0] != 168 {
		t.Fatal(e)
	}
	state, e = s.Control(ctx, state.Current.ID, l.Operation{Kind: "force", Root: 6})
	if e != nil {
		t.Fatal(e)
	}
	state, e = s.Control(ctx, state.Current.ID, l.Operation{Kind: "tick", Ticks: 1000})
	if e != nil || state.Current.Snapshot.Counters.Muls != 1 {
		t.Fatal("repeat body")
	}
}
func TestHTTPRejectsCrossOriginAndMalformedMutation(t *testing.T) {
	s, _ := NewSession(context.Background(), l.NewModel(), zerolog.Nop())
	defer s.Close()
	h := NewHandler(s)
	for _, tc := range []struct {
		body, origin string
		code         int
	}{{`{"expectedId":1,"operation":{"kind":"reset"}}`, "http://evil.example", 403}, {`{"expectedId":1,"operation":{"kind":"reset"},"unknown":1}`, "", 400}, {`{"expectedId":1,"operation":{"kind":"reset"}} {}`, "", 400}, {`{"expectedId":0,"operation":{"kind":"reset"}}`, "", 409}} {
		req := httptest.NewRequest("POST", "http://localhost/api/lazy/control", strings.NewReader(tc.body))
		req.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != tc.code {
			t.Fatalf("%s got %d", tc.body, w.Code)
		}
	}
	r := httptest.NewRequest("POST", "http://localhost/api/lazy/control", strings.NewReader(fmt.Sprintf(`{"expectedId":%d,"operation":{"kind":"reset"}}`, s.State().Current.ID)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
}

type brokenSnapshot struct {
	*l.Model
	fail bool
}

func (b *brokenSnapshot) Snapshot(ctx context.Context) (l.Snapshot, error) {
	if b.fail {
		return l.Snapshot{}, errors.New("snapshot unavailable")
	}
	return b.Model.Snapshot(ctx)
}
func TestFailedResetCapturePreservesLastObservation(t *testing.T) {
	b := &brokenSnapshot{Model: l.NewModel()}
	s, _ := NewSession(context.Background(), b, zerolog.Nop())
	before := s.State().Current.ID
	b.fail = true
	if _, e := s.Control(context.Background(), before, l.Operation{Kind: "reset"}); e == nil {
		t.Fatal("failure hidden")
	}
	after := s.State()
	if !after.NeedsReset || after.Current.ID != before {
		t.Fatal("lost last observation")
	}
}
