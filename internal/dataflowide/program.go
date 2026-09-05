package dataflowide

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	df "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/dataflow"
)

func (s *Session) LoadProgram(ctx context.Context, expected uint64, source string) error {
	p, err := df.Compile(source)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err = s.guardLocked(expected); err != nil {
		return err
	}
	if _, err = s.executeLocked(ctx, df.Operation{Kind: "reset"}, "program load: explicit reset"); err != nil {
		return err
	}
	if _, err = s.executeLocked(ctx, df.Operation{Kind: "load", Graph: &p.Graph}, "program descriptors activated"); err != nil {
		return err
	}
	last := &s.frames[len(s.frames)-1]
	if last.Snapshot.Graph != p.Graph {
		s.needsReset = true
		return errors.New("active graph readback differs from compiled program")
	}
	s.program = p.Clone()
	last.Program = p.Clone()
	s.scenarioHash = [32]byte{}
	s.scenario = Scenario{}
	s.step = 0
	s.eventLocked("program load", "compiled graph verified against engine readback", false)
	return nil
}
func (s *Session) ProgramInputs(ctx context.Context, expected uint64, c byte, values map[string]int64) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err = s.guardLocked(expected); err != nil {
		return err
	}
	if s.program == nil {
		return errors.New("load a compiled program first")
	}
	if c >= df.Contexts {
		return errors.New("context must be 0..3")
	}
	if s.needsReset {
		return errors.New("engine state uncertain; reset required")
	}
	frame := s.frames[len(s.frames)-1]
	if frame.Snapshot.Closed[c] {
		return errors.Wrap(ErrConflict, "cancel closed context before supplying new inputs")
	}
	if frame.Snapshot.Debug.Halted {
		return errors.Wrap(ErrConflict, "resume before supplying grouped inputs")
	}
	tokens, err := s.program.Tokens(c, frame.Snapshot.Epochs[c], values)
	if err != nil {
		return err
	}
	// Values are validated before the first mutation. A failure still captures the
	// actual partial injection state and makes transport uncertainty explicit.
	defer func() {
		if err != nil && !errors.Is(err, df.ErrFull) && !errors.Is(err, df.ErrBlocked) {
			s.needsReset = true
		}
		captureErr := s.captureLocked(ctx, "named program inputs")
		if err == nil {
			err = captureErr
		}
		if err != nil {
			s.lastError = err.Error()
			s.eventLocked("program inputs", err.Error(), true)
		} else {
			s.lastError = ""
			s.eventLocked("program inputs", "all named values supplied", false)
		}
		s.scenarioHash = [32]byte{}
		s.step = 0
	}()
	for _, token := range tokens {
		accepted := false
		for attempt := 0; attempt < 128; attempt++ {
			_, err = s.engine.Execute(ctx, df.Operation{Kind: "inject", Token: &token})
			if err == nil {
				accepted = true
				break
			}
			if !errors.Is(err, df.ErrFull) {
				return err
			}
			if _, err = s.engine.Execute(ctx, df.Operation{Kind: "tick", Ticks: 1}); err != nil {
				return err
			}
		}
		if !accepted {
			return df.ErrFull
		}
	}
	return nil
}
func programRoutes(mux *http.ServeMux, s *Session, p *Projects) {
	mux.HandleFunc("POST /api/dataflow/program/compile", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Source string `json:"source"`
		}
		if err := decode(w, r, &in); err != nil {
			problem(w, 400, err)
			return
		}
		program, err := df.Compile(in.Source)
		if err != nil {
			respond(w, 200, map[string]any{"diagnostics": []string{err.Error()}, "program": nil})
			return
		}
		respond(w, 200, map[string]any{"diagnostics": []string{}, "program": program})
	})
	mux.HandleFunc("POST /api/dataflow/program/load", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Source     string `json:"source"`
			ExpectedID uint64 `json:"expectedId"`
		}
		if err := decode(w, r, &in); err != nil {
			problem(w, 400, err)
			return
		}
		if _, err := df.Compile(in.Source); err != nil {
			problem(w, 400, err)
			return
		}
		if err := s.LoadProgram(r.Context(), in.ExpectedID, in.Source); err != nil {
			problem(w, operationStatus(err), err)
			return
		}
		respond(w, 200, s.State())
	})
	mux.HandleFunc("POST /api/dataflow/program/inputs", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			ExpectedID uint64           `json:"expectedId"`
			Context    byte             `json:"context"`
			Values     map[string]int64 `json:"values"`
		}
		if err := decode(w, r, &in); err != nil {
			problem(w, 400, err)
			return
		}
		if err := s.ProgramInputs(r.Context(), in.ExpectedID, in.Context, in.Values); err != nil {
			problem(w, operationStatus(err), err)
			return
		}
		respond(w, 200, s.State())
	})
	mux.HandleFunc("GET /api/dataflow/programs", func(w http.ResponseWriter, r *http.Request) {
		ids, err := p.list(".df")
		if err != nil {
			problem(w, 500, err)
			return
		}
		respond(w, 200, ids)
	})
	mux.HandleFunc("GET /api/dataflow/programs/{id}", func(w http.ResponseWriter, r *http.Request) {
		source, err := p.read(r.PathValue("id"), ".df")
		if err != nil {
			problem(w, 400, err)
			return
		}
		respond(w, 200, map[string]string{"source": source})
	})
	mux.HandleFunc("PUT /api/dataflow/programs/{id}", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Source string `json:"source"`
		}
		if err := decode(w, r, &in); err != nil {
			problem(w, 400, err)
			return
		}
		if _, err := df.Compile(in.Source); err != nil {
			problem(w, 400, err)
			return
		}
		if err := p.write(r.PathValue("id"), in.Source, ".df"); err != nil {
			problem(w, 400, err)
			return
		}
		respond(w, 200, map[string]bool{"saved": true})
	})
}
