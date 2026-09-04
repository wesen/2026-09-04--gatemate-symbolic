package microscope

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	domain "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/microscope"
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
		if err != nil || u.Host != r.Host {
			return errors.New("cross-origin device mutation is not allowed")
		}
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return errors.Wrap(err, "invalid JSON")
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("expected one JSON object")
	}
	return nil
}
func NewHandler(session *Session) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) { respond(w, 200, session.State()) })
	mux.HandleFunc("POST /api/graph", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Vertices  int     `json:"vertices"`
			Colors    int     `json:"colors"`
			Edges     [][]int `json:"edges"`
			FirstOnly bool    `json:"firstOnly"`
		}
		if err := decode(w, r, &input); err != nil {
			problem(w, 400, err)
			return
		}
		g := domain.Graph{Vertices: input.Vertices, Colors: input.Colors, FirstOnly: input.FirstOnly, Edges: [][2]int{}}
		for _, edge := range input.Edges {
			if len(edge) != 2 {
				problem(w, 400, errors.New("each edge must contain exactly two vertices"))
				return
			}
			g.Edges = append(g.Edges, [2]int{edge[0], edge[1]})
		}
		if _, err := g.Adjacency(); err != nil {
			problem(w, 400, err)
			return
		}
		if err := session.Load(r.Context(), g); err != nil {
			code := 502
			if errors.Is(err, ErrConflict) {
				code = 409
			}
			problem(w, code, err)
			return
		}
		respond(w, 200, session.State())
	})
	mux.HandleFunc("POST /api/control", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Action string `json:"action"`
		}
		if err := decode(w, r, &input); err != nil {
			problem(w, 400, err)
			return
		}
		if input.Action != "step" && input.Action != "run" && input.Action != "pause" && input.Action != "reset" {
			problem(w, 400, errors.New("action must be step, run, pause, or reset"))
			return
		}
		if err := session.Control(r.Context(), input.Action); err != nil {
			code := 502
			if errors.Is(err, ErrConflict) {
				code = 409
			}
			problem(w, code, err)
			return
		}
		respond(w, 200, session.State())
	})
	mux.HandleFunc("GET /api/events/{sequence}", func(w http.ResponseWriter, r *http.Request) {
		sequence, err := strconv.ParseUint(r.PathValue("sequence"), 10, 32)
		if err != nil {
			problem(w, 400, errors.New("invalid event sequence"))
			return
		}
		generation, err := strconv.ParseUint(r.URL.Query().Get("generation"), 10, 64)
		if err != nil {
			problem(w, 400, errors.New("generation is required"))
			return
		}
		snapshot, err := session.Event(uint32(sequence), generation)
		if err != nil {
			code := 404
			if errors.Is(err, ErrConflict) {
				code = 409
			}
			problem(w, code, err)
			return
		}
		respond(w, 200, snapshot)
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) { problem(w, 404, errors.New("unknown API endpoint")) })
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		body, err := asset("index.html")
		if err != nil {
			http.Error(w, "Frontend is not built. Run make frontend.", 503)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(body)
	})
	mux.HandleFunc("GET /static/{path...}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("path")
		if name != "app.js" && name != "app.css" {
			http.NotFound(w, r)
			return
		}
		body, err := asset(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if name == "app.js" {
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		}
		_, _ = w.Write(body)
	})
	return mux
}
