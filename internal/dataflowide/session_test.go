package dataflowide

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	df "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/dataflow"
)

func testSession(t *testing.T) *Session {
	t.Helper()
	m, err := df.NewTransaction(df.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewSession(context.Background(), m, zerolog.Nop())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}
func TestScenarioExamplesAndStep(t *testing.T) {
	for id, source := range Examples() {
		t.Run(id, func(t *testing.T) {
			s := testSession(t)
			scenario, d := ParseScenario(source)
			if len(d) != 0 {
				t.Fatal(d)
			}
			for range scenario.Actions {
				if err := s.StepScenario(context.Background(), s.State().Current.ID, scenario); err != nil {
					t.Fatal(err)
				}
			}
			state := s.State()
			if state.Step != len(scenario.Actions) || len(state.Results) == 0 || state.Error != "" {
				t.Fatalf("bad final state %+v", state)
			}
		})
	}
}
func TestHistoryGenerationBoundsAndOwnership(t *testing.T) {
	s := testSession(t)
	ctx := context.Background()
	initial := s.State().Current
	for n := 0; n < 150; n++ {
		if err := s.Control(ctx, s.State().Current.ID, df.Operation{Kind: "tick", Ticks: 1}); err != nil {
			t.Fatal(err)
		}
	}
	state := s.State()
	if len(state.History) != HistoryLimit {
		t.Fatal(len(state.History))
	}
	if _, ok := s.Historical(initial.ID, initial.Generation); ok {
		t.Fatal("expired frame retained")
	}
	state.Current.Snapshot.Counters["cycles"] = 999
	if s.State().Current.Snapshot.Counters["cycles"] != 150 {
		t.Fatal("state snapshot aliases history")
	}
	old := s.State().Current
	if err := s.Control(ctx, old.ID, df.Operation{Kind: "reset"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Historical(old.ID, old.Generation); ok {
		t.Fatal("old generation retained")
	}
	if err := s.Control(ctx, old.ID, df.Operation{Kind: "tick", Ticks: 1}); !errors.Is(err, ErrConflict) {
		t.Fatal("stale frame mutated engine", err)
	}
}
func TestRunPauseAndConflicts(t *testing.T) {
	s := testSession(t)
	scenario := Scenario{Version: 1, Name: "pause", Actions: []Action{{Kind: "reset"}}}
	for n := 0; n < 100; n++ {
		scenario.Actions = append(scenario.Actions, Action{Kind: "tick", Ticks: 1})
	}
	if err := s.RunScenario(s.State().Current.ID, scenario); err != nil {
		t.Fatal(err)
	}
	if err := s.Control(context.Background(), s.State().Current.ID, df.Operation{Kind: "tick", Ticks: 1}); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	s.Pause()
	deadline := time.Now().Add(time.Second)
	for s.State().Running && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	state := s.State()
	if state.Running || state.Step >= len(scenario.Actions) {
		t.Fatal("pause did not stop at boundary")
	}
}
func TestScenarioValidationAndAssertions(t *testing.T) {
	for _, source := range []string{`{}`, `{"version":1,"name":"bad","actions":[{"kind":"poll"}]}`, `{"version":1,"name":"bad","actions":[{"kind":"reset"},{"kind":"inputs","values":[1]}]}`, `{"version":1,"name":"bad","actions":[{"kind":"reset"}],"unknown":1}`} {
		if _, d := ParseScenario(source); len(d) == 0 {
			t.Fatal("accepted", source)
		}
	}
	s := testSession(t)
	scenario := Scenario{Version: 1, Name: "assert", Actions: []Action{{Kind: "reset"}, {Kind: "expect", Value: value(58)}}}
	if err := s.StepScenario(context.Background(), s.State().Current.ID, scenario); err != nil {
		t.Fatal(err)
	}
	if err := s.StepScenario(context.Background(), s.State().Current.ID, scenario); !errors.Is(err, ErrAssertion) {
		t.Fatal(err)
	}
}
func TestProjectsAtomicValidationAndConfinement(t *testing.T) {
	root := t.TempDir()
	p, err := NewProjects(root)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	source := Examples()["book"]
	if err := p.Write("book", source); err != nil {
		t.Fatal(err)
	}
	got, err := p.Read("book")
	if err != nil || got != source {
		t.Fatal(err)
	}
	if err := p.Write("book", `{}`); err == nil {
		t.Fatal("invalid source overwrote project")
	}
	got, _ = p.Read("book")
	if got != source {
		t.Fatal("project changed")
	}
	if err := p.Write("../escape", source); err == nil {
		t.Fatal("path traversal accepted")
	}
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Read("link"); err == nil {
		t.Fatal("followed symlink outside project root")
	}
	ids, err := p.List()
	if err != nil || len(ids) != 2 {
		t.Fatal(ids, err)
	}
}
