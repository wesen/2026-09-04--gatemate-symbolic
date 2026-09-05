package dataflowide

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPControlValidationAndIsolation(t *testing.T) {
	s := testSession(t)
	p, err := NewProjects(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	h := NewHandler(s, p)
	post := func(path, body, origin string) int {
		r := httptest.NewRequest("POST", path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	body := fmt.Sprintf(`{"expectedId":%d,"operation":{"kind":"tick","ticks":1}}`, s.State().Current.ID)
	if status := post("/api/dataflow/control", body, "https://unrelated.example"); status != 400 {
		t.Fatal(status)
	}
	if status := post("/api/dataflow/control", body, ""); status != 200 {
		t.Fatal(status)
	}
	if status := post("/api/dataflow/control", body, ""); status != 409 {
		t.Fatal(status)
	}
	if status := post("/api/dataflow/control", `{"operation":{"kind":"bad"}}`, ""); status != 400 {
		t.Fatal(status)
	}
	if status := post("/api/dataflow/control", body+`{}`, ""); status != 400 {
		t.Fatal(status)
	}
	for _, path := range []string{"/static/missing.js", "/api/dataflow/missing", "/static/../../go.mod"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code == 200 {
			t.Fatal("unexpectedly served", path)
		}
	}
	source, _ := json.Marshal(map[string]string{"source": Examples()["book"]})
	if status := post("/api/dataflow/scenario/validate", string(source), ""); status != 200 {
		t.Fatal(status)
	}
}
