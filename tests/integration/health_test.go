package integration

import (
	"net/http"
	"strings"
	"testing"
)

func TestHealthReturnsOK(t *testing.T) {
	app := newTestApp(t)

	response := app.get("/api/health")

	if response.Status != http.StatusOK {
		t.Errorf("status = %d, want %d", response.Status, http.StatusOK)
	}
	if want := `{"status":"ok"}`; response.Body != want {
		t.Errorf("body = %s, want %s", response.Body, want)
	}
	if !strings.HasPrefix(response.ContentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", response.ContentType)
	}
}

// TestUnknownAPIPathReturnsJSON guards the contract the frontend's fetch layer
// depends on: a miss under /api is JSON, never the SPA's index.html.
func TestUnknownAPIPathReturnsJSON(t *testing.T) {
	app := newTestApp(t)

	response := app.get("/api/does-not-exist")

	if response.Status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", response.Status, http.StatusNotFound)
	}
	if !strings.HasPrefix(response.ContentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", response.ContentType)
	}
	if !strings.Contains(response.Body, `"code":"not_found"`) {
		t.Errorf("body = %s, want an error envelope with code not_found", response.Body)
	}
}

func TestWrongMethodOnKnownPathReturnsJSON(t *testing.T) {
	app := newTestApp(t)

	response, err := http.Post(app.URL+"/api/health", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/health: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", response.StatusCode, http.StatusMethodNotAllowed)
	}
	if contentType := response.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
}
