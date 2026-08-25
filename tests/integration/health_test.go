package integration

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/httplog/v2"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web"
)

// newTestServer starts the production router on a random port.
//
// The logger is set to Error so a passing test run stays readable; a failing
// assertion says more than a request log line would.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	logger := httplog.NewLogger("wedding-test", httplog.Options{LogLevel: slog.LevelError})
	server := httptest.NewServer(web.NewRouter(logger))
	t.Cleanup(server.Close)

	return server
}

// get performs a GET and returns status, content type and trimmed body.
func get(t *testing.T, baseURL, path string) (int, string, string) {
	t.Helper()

	response, err := http.Get(baseURL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("reading body of GET %s: %v", path, err)
	}

	return response.StatusCode, response.Header.Get("Content-Type"), strings.TrimSpace(string(body))
}

func TestHealthReturnsOK(t *testing.T) {
	server := newTestServer(t)

	status, contentType, body := get(t, server.URL, "/api/health")

	if status != http.StatusOK {
		t.Errorf("status = %d, want %d", status, http.StatusOK)
	}
	if want := `{"status":"ok"}`; body != want {
		t.Errorf("body = %s, want %s", body, want)
	}
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
}

// TestUnknownAPIPathReturnsJSON guards the contract the frontend's fetch layer
// depends on: a miss under /api is JSON, never the SPA's index.html.
func TestUnknownAPIPathReturnsJSON(t *testing.T) {
	server := newTestServer(t)

	status, contentType, body := get(t, server.URL, "/api/does-not-exist")

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", status, http.StatusNotFound)
	}
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
	if !strings.Contains(body, `"code":"not_found"`) {
		t.Errorf("body = %s, want an error envelope with code not_found", body)
	}
}

func TestWrongMethodOnKnownPathReturnsJSON(t *testing.T) {
	server := newTestServer(t)

	response, err := http.Post(server.URL+"/api/health", "application/json", nil)
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
