package lazylanguageide

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	examples "github.com/wesen/2026-09-04--gatemate-symbolic/examples/lazylang"
)

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		respond(w, 415, map[string]string{"error": "application/json required"})
		return false
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		respond(w, 400, map[string]string{"error": err.Error()})
		return false
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		respond(w, 400, map[string]string{"error": "expected one JSON object"})
		return false
	}
	return true
}
func NewHandler(s *Session) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/language/state", func(w http.ResponseWriter, r *http.Request) { respond(w, 200, s.State()) })
	mux.HandleFunc("GET /api/language/examples", func(w http.ResponseWriter, r *http.Request) {
		respond(w, 200, examples.Sources())
	})
	mux.HandleFunc("POST /api/language/compile", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Source         string `json:"source"`
			ClientRevision uint64 `json:"clientRevision"`
		}
		if !decode(w, r, &in) {
			return
		}
		respond(w, 200, s.Compile(r.Context(), in.Source, in.ClientRevision))
	})
	mux.HandleFunc("GET /api/language/artifacts/{id}", func(w http.ResponseWriter, r *http.Request) {
		v, ok := s.Artifact(r.PathValue("id"))
		if !ok {
			respond(w, 404, map[string]string{"error": "artifact not found"})
			return
		}
		respond(w, 200, v)
	})
	mux.HandleFunc("GET /api/language/history/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
		if err != nil {
			respond(w, 400, map[string]string{"error": "invalid frame ID"})
			return
		}
		f, ok := s.History(id)
		if !ok {
			respond(w, 404, map[string]string{"error": "history frame not retained"})
			return
		}
		respond(w, 200, f)
	})
	mux.HandleFunc("POST /api/language/control", func(w http.ResponseWriter, r *http.Request) {
		var op Operation
		if !decode(w, r, &op) {
			return
		}
		state, err := s.Control(r.Context(), op)
		if err != nil {
			respond(w, 409, struct {
				Error string `json:"error"`
				State State  `json:"state"`
			}{err.Error(), state})
			return
		}
		respond(w, 200, state)
	})
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
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
			http.Error(w, "Build the language frontend with go generate ./internal/lazylanguageide", 503)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		ip := net.ParseIP(host)
		if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			respond(w, 403, map[string]string{"error": "loopback host required"})
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || u.Host != r.Host || (u.Scheme != "http" && u.Scheme != "https") {
				respond(w, 403, map[string]string{"error": "cross-origin access rejected"})
				return
			}
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		mux.ServeHTTP(w, r)
	})
}
