package microscope

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func request(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestHTTPGraphValidationAndControl(t *testing.T) {
	s := newTestSession(t)
	h := NewHandler(s)
	for _, body := range []string{`{"vertices":3,"colors":2,"edges":[[0,1,2]]}`, `{"vertices":3,"colors":2,"extra":true}`, `{"vertices":3,"colors":2} {}`, `{"vertices":3,"colors":2,"edges":[[0,0]]}`, strings.Repeat("x", 5000)} {
		if w := request(h, "POST", "/api/graph", body); w.Code != 400 {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
	}
	if s.State().Loaded {
		t.Fatal("invalid input changed device")
	}
	w := request(h, "POST", "/api/graph", `{"vertices":3,"colors":3,"edges":[[0,1],[1,2],[0,2]],"firstOnly":true}`)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	for !s.State().Latest.Terminal() {
		if w = request(h, "POST", "/api/control", `{"action":"step"}`); w.Code != 200 {
			t.Fatal(w.Body.String())
		}
	}
	if s.State().Latest.Count != 1 {
		t.Fatal("cut")
	}
	if w = request(h, "POST", "/api/control", `{"action":"step"}`); w.Code != 409 {
		t.Fatal(w.Code)
	}
	if w = request(h, "GET", "/api/events/1?generation=1", ""); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w = request(h, "GET", "/api/events/1?generation=999", ""); w.Code != 409 {
		t.Fatal(w.Code)
	}
	if w = request(h, "GET", "/api/events/1", ""); w.Code != 400 {
		t.Fatal(w.Code)
	}
	if w = request(h, "GET", "/api/no-such-path", ""); w.Code != 404 || !strings.Contains(w.Header().Get("Content-Type"), "json") {
		t.Fatal(w.Code)
	}
	if w = request(h, "GET", "/static/../../go.mod", ""); w.Code == 200 {
		t.Fatal("static traversal")
	}
	w = request(h, "GET", "/api/state", "")
	var state State
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil || state.Engine != "model" {
		t.Fatal(err, state)
	}
	r := httptest.NewRequest("POST", "/api/control", strings.NewReader(`{"action":"reset"}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", "http://other.invalid")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal("cross-origin mutation", w.Code)
	}
}
