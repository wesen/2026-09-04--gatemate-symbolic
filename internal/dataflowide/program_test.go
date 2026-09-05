package dataflowide

import (
	"context"
	df "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/dataflow"
	"testing"
)

func TestProgramSessionOwnershipAndExecution(t *testing.T) {
	s := testSession(t)
	ctx := context.Background()
	source := "input a: int16; let square=a*a; output square+square"
	old := s.State().Current
	if err := s.LoadProgram(ctx, old.ID, source); err != nil {
		t.Fatal(err)
	}
	loaded := s.State().Current
	if loaded.Program == nil || loaded.Snapshot.Graph != loaded.Program.Graph || loaded.Generation == old.Generation {
		t.Fatal("program graph or generation")
	}
	loaded.Program.Inputs[0].Destinations[0].Node = 6
	if s.State().Current.Program.Inputs[0].Destinations[0].Node == 6 {
		t.Fatal("program aliases history")
	}
	before := s.State().Current.ID
	if err := s.ProgramInputs(ctx, before, 0, map[string]int64{"a": 32768}); err == nil {
		t.Fatal("invalid input accepted")
	}
	if s.State().Current.ID != before || s.State().NeedsReset {
		t.Fatal("validation mutated engine")
	}
	if err := s.ProgramInputs(ctx, before, 0, map[string]int64{"a": 7}); err != nil {
		t.Fatal(err)
	}
	if err := s.Control(ctx, s.State().Current.ID, df.Operation{Kind: "tick", Ticks: 100}); err != nil {
		t.Fatal(err)
	}
	if err := s.Control(ctx, s.State().Current.ID, df.Operation{Kind: "poll"}); err != nil {
		t.Fatal(err)
	}
	state := s.State()
	if len(state.Results) != 1 || state.Results[0].Value != 98 {
		t.Fatalf("result %+v", state.Results)
	}
	if err := s.LoadProgram(ctx, state.Current.ID, "output undefined"); err == nil {
		t.Fatal("invalid source loaded")
	}
	if s.State().Current.Program.Source != source {
		t.Fatal("failed compile changed loaded source")
	}
	if err := s.LoadProgram(ctx, state.Current.ID, "output 12"); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Historical(loaded.ID, loaded.Generation); ok {
		t.Fatal("old generation retained across load")
	}
}
func TestProgramPersistenceIsSeparateAndConfined(t *testing.T) {
	p, err := NewProjects(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if err = p.write("sample", "output 12", ".df"); err != nil {
		t.Fatal(err)
	}
	if got, err := p.read("sample", ".df"); err != nil || got != "output 12" {
		t.Fatal(got, err)
	}
	ids, _ := p.list(".df")
	old, _ := p.List()
	if len(ids) != 1 || len(old) != 0 {
		t.Fatal("program and scenario namespaces mixed")
	}
	if p.write("../escape", "output 12", ".df") == nil {
		t.Fatal("invalid ID accepted")
	}
}
