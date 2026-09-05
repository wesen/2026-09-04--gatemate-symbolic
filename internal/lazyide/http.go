package lazyide

import (
	"encoding/json"
	l "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazy"
	"io"
	"net/http"
	"net/url"
)

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func NewHandler(s *Session) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.HandleFunc("GET /api/lazy/state", func(w http.ResponseWriter, r *http.Request) { respond(w, 200, s.State()) })
	mux.HandleFunc("GET /api/lazy/examples", func(w http.ResponseWriter, r *http.Request) {
		out := map[string]l.Image{}
		for _, name := range []string{"shared", "cycle", "indirection", "indirection-cycle", "overflow"} {
			out[name], _ = l.Example(name)
		}
		respond(w, 200, out)
	})
	mux.HandleFunc("POST /api/lazy/control", func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			u, e := url.Parse(origin)
			if e != nil || u.Host != r.Host || (u.Scheme != "http" && u.Scheme != "https") {
				respond(w, 403, map[string]string{"error": "cross-origin mutation rejected"})
				return
			}
		}
		var in struct {
			ExpectedID uint64      `json:"expectedId"`
			Operation  l.Operation `json:"operation"`
		}
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128*1024))
		d.DisallowUnknownFields()
		if e := d.Decode(&in); e != nil {
			respond(w, 400, map[string]string{"error": e.Error()})
			return
		}
		if e := d.Decode(&struct{}{}); e != io.EOF {
			respond(w, 400, map[string]string{"error": "expected one JSON object"})
			return
		}
		state, e := s.Control(r.Context(), in.ExpectedID, in.Operation)
		if e != nil {
			respond(w, 409, map[string]string{"error": e.Error()})
			return
		}
		respond(w, 200, state)
	})
	mux.HandleFunc("GET /static/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name != "app.js" && name != "app.css" {
			http.NotFound(w, r)
			return
		}
		b, e := asset(name)
		if e != nil {
			http.NotFound(w, r)
			return
		}
		if name == "app.js" {
			w.Header().Set("Content-Type", "text/javascript")
		} else {
			w.Header().Set("Content-Type", "text/css")
		}
		_, _ = w.Write(b)
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		b, e := asset("index.html")
		if e != nil {
			http.Error(w, "Build with make lazy-frontend", 503)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(b)
	})
	return mux
}
