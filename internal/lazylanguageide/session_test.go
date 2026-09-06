package lazylanguageide

import (
	"context"
	"github.com/rs/zerolog"
	"net/http/httptest"
	"strings"
	"testing"
)

func control(t *testing.T, s *Session, kind string, ticks uint32) State {
	t.Helper()
	state := s.State()
	got, err := s.Control(context.Background(), Operation{Kind: kind, ExpectedID: state.Frame.ID, RunID: state.Frame.RunID, Ticks: ticks})
	if err != nil {
		t.Fatal(err)
	}
	return got
}
func loaded(t *testing.T, source string) *Session {
	t.Helper()
	s := NewSession(nil, zerolog.Nop())
	compiled := s.Compile(context.Background(), source, 7)
	if len(compiled.Diagnostics) > 0 {
		t.Fatal(compiled.Diagnostics)
	}
	state := s.State()
	if _, err := s.Control(context.Background(), Operation{Kind: "load", ExpectedID: state.Frame.ID, RunID: state.Frame.RunID, ArtifactID: compiled.ArtifactID}); err != nil {
		t.Fatal(err)
	}
	return s
}
func TestStreamDemandBoundaryAndResume(t *testing.T) {
	s := loaded(t, `def main : ListInt = Cons(7, let rec xs : ListInt = xs in xs);`)
	control(t, s, "stream-start", 0)
	paused := control(t, s, "stream-next", 1)
	if !paused.Frame.Stream.Pending {
		t.Fatal("missing pending request")
	}
	op := Operation{Kind: "stream-next", ExpectedID: paused.Frame.ID, RunID: paused.Frame.RunID, Ticks: 1000}
	if _, err := s.Control(context.Background(), op); err == nil {
		t.Fatal("duplicate next accepted")
	}
	state := control(t, s, "stream-resume", 1000)
	if state.Frame.Stream.Pending || len(state.Frame.Stream.Values) != 1 || state.Frame.Stream.Values[0] != 7 || state.Frame.Stream.Fault != 0 {
		t.Fatal(state.Frame.Stream)
	}
	if state.Frame.Snapshot.Counters.Claims != 2 {
		t.Fatal("forced final tail", state.Frame.Snapshot.Counters)
	}
	state = control(t, s, "stream-next", 1000)
	if !state.Frame.Stream.Complete || state.Frame.Stream.Fault != 3 {
		t.Fatal(state.Frame.Stream)
	}
}
func TestRunIdentityHistoryAndCompileIsolation(t *testing.T) {
	s := loaded(t, `def main : Int = 42;`)
	before := s.State()
	bad := s.Compile(context.Background(), `def main : Int = true;`, 9)
	if len(bad.Diagnostics) == 0 || bad.ClientRevision != 9 {
		t.Fatal(bad)
	}
	if s.State().Frame.ID != before.Frame.ID {
		t.Fatal("compile changed execution")
	}
	control(t, s, "tick", 0)
	if _, err := s.Control(context.Background(), Operation{Kind: "tick", ExpectedID: before.Frame.ID, RunID: before.Frame.RunID}); err == nil {
		t.Fatal("historical control accepted")
	}
	history, ok := s.History(before.Frame.ID)
	if !ok {
		t.Fatal("history missing")
	}
	history.Snapshot.Heap[0] = "mutated"
	again, _ := s.History(before.Frame.ID)
	if again.Snapshot.Heap[0] == "mutated" {
		t.Fatal("history aliases caller")
	}
	for i := 0; i < 140; i++ {
		control(t, s, "tick", 0)
	}
	if len(s.State().History) != 128 {
		t.Fatal("unbounded history")
	}
	reset := control(t, s, "reset", 0)
	if reset.Frame.RunID == before.Frame.RunID || reset.Frame.ArtifactID != "" {
		t.Fatal("reset identity")
	}
}
func TestHTTPBoundary(t *testing.T) {
	handler := NewHandler(NewSession(nil, zerolog.Nop()))
	for _, tc := range []struct {
		host, origin, body string
		want               int
	}{
		{"localhost:18091", "", `{"source":"def main : Int = 1;","clientRevision":1}`, 200},
		{"evil.example", "", `{}`, 403},
		{"localhost:18091", "http://evil.example", `{}`, 403},
		{"localhost:18091", "", `{} {}`, 400},
		{"localhost:18091", "", `{"unknown":1}`, 400},
	} {
		r := httptest.NewRequest("POST", "http://"+tc.host+"/api/language/compile", strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
