package dataflowide

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	df "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/dataflow"
)

func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func problem(w http.ResponseWriter, status int, err error) {
	respond(w, status, map[string]string{"error": err.Error()})
}
func decode(w http.ResponseWriter, r *http.Request, target any) error {
	if strings.Split(r.Header.Get("Content-Type"), ";")[0] != "application/json" {
		return errors.New("Content-Type must be application/json")
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		if err != nil || u.Host != r.Host || (u.Scheme != "http" && u.Scheme != "https") {
			return errors.New("cross-origin mutation refused")
		}
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxSourceBytes*2))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return errors.Wrap(err, "invalid JSON")
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return errors.New("expected one JSON object")
	}
	return nil
}
func operationStatus(err error) int {
	if errors.Is(err, ErrConflict) || errors.Is(err, df.ErrFull) || errors.Is(err, df.ErrBlocked) || errors.Is(err, ErrAssertion) {
		return 409
	}
	return 502
}
func NewHandler(s *Session, p *Projects) http.Handler {
	mux := http.NewServeMux()
	programRoutes(mux, s, p)
	mux.HandleFunc("GET /api/dataflow/state", func(w http.ResponseWriter, r *http.Request) { respond(w, 200, s.State()) })
	mux.HandleFunc("GET /api/dataflow/examples", func(w http.ResponseWriter, r *http.Request) { respond(w, 200, Examples()) })
	mux.HandleFunc("POST /api/dataflow/control", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ExpectedID uint64       `json:"expectedId"`
			Operation  df.Operation `json:"operation"`
		}
		if err := decode(w, r, &input); err != nil {
			problem(w, 400, err)
			return
		}
		if err := input.Operation.Validate(); err != nil {
			problem(w, 400, err)
			return
		}
		if err := s.Control(r.Context(), input.ExpectedID, input.Operation); err != nil {
			problem(w, operationStatus(err), err)
			return
		}
		respond(w, 200, s.State())
	})
	mux.HandleFunc("POST /api/dataflow/pause", func(w http.ResponseWriter, r *http.Request) {
		var input struct{}
		if err := decode(w, r, &input); err != nil {
			problem(w, 400, err)
			return
		}
		s.Pause()
		respond(w, 200, s.State())
	})
	mux.HandleFunc("POST /api/dataflow/scenario/validate", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Source string `json:"source"`
		}
		if err := decode(w, r, &input); err != nil {
			problem(w, 400, err)
			return
		}
		scenario, d := ParseScenario(input.Source)
		respond(w, 200, map[string]any{"diagnostics": d, "formatted": FormatScenario(scenario), "actions": len(scenario.Actions)})
	})
	mux.HandleFunc("POST /api/dataflow/scenario/run", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Source     string `json:"source"`
			Mode       string `json:"mode"`
			ExpectedID uint64 `json:"expectedId"`
		}
		if err := decode(w, r, &input); err != nil {
			problem(w, 400, err)
			return
		}
		scenario, d := ParseScenario(input.Source)
		if len(d) > 0 {
			respond(w, 400, map[string]any{"error": "invalid scenario", "diagnostics": d})
			return
		}
		var err error
		switch input.Mode {
		case "step":
			err = s.StepScenario(r.Context(), input.ExpectedID, scenario)
		case "run":
			err = s.RunScenario(input.ExpectedID, scenario)
		default:
			problem(w, 400, errors.New("mode must be step or run"))
			return
		}
		if err != nil {
			problem(w, operationStatus(err), err)
			return
		}
		respond(w, 200, s.State())
	})
	mux.HandleFunc("GET /api/dataflow/history/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, e1 := strconv.ParseUint(r.PathValue("id"), 10, 64)
		generation, e2 := strconv.ParseUint(r.URL.Query().Get("generation"), 10, 64)
		if e1 != nil || e2 != nil {
			problem(w, 400, errors.New("history requires numeric id and generation"))
			return
		}
		frame, ok := s.Historical(id, generation)
		if !ok {
			problem(w, 404, errors.New("snapshot expired or belongs to another reset generation"))
			return
		}
		respond(w, 200, frame)
	})
	mux.HandleFunc("GET /api/dataflow/projects", func(w http.ResponseWriter, r *http.Request) {
		ids, err := p.List()
		if err != nil {
			problem(w, 500, err)
			return
		}
		respond(w, 200, ids)
	})
	mux.HandleFunc("GET /api/dataflow/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		source, err := p.Read(r.PathValue("id"))
		if err != nil {
			status := 400
			if errors.Is(err, os.ErrNotExist) {
				status = 404
			}
			problem(w, status, err)
			return
		}
		respond(w, 200, map[string]string{"source": source})
	})
	mux.HandleFunc("PUT /api/dataflow/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Source string `json:"source"`
		}
		if err := decode(w, r, &input); err != nil {
			problem(w, 400, err)
			return
		}
		if err := p.Write(r.PathValue("id"), input.Source); err != nil {
			problem(w, 400, err)
			return
		}
		respond(w, 200, map[string]bool{"saved": true})
	})
	mux.HandleFunc("GET /static/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name != "app.js" && name != "app.css" {
			http.NotFound(w, r)
			return
		}
		data, err := asset(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if name == "app.js" {
			w.Header().Set("Content-Type", "text/javascript")
		} else {
			w.Header().Set("Content-Type", "text/css")
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		data, err := asset("index.html")
		if err != nil {
			http.Error(w, "Build the dataflow frontend with make dataflow-frontend", 503)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})
	return mux
}
