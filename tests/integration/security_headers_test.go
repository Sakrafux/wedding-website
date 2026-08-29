package integration

import (
	"net/http"
	"strings"
	"testing"
)

// wantSecurityHeaders mirrors the table in 06-privacy-security. Duplicated here on
// purpose rather than imported from the middleware: a test that reads the same map
// the code writes asserts only that a map exists, not that its values are the ones
// that were reviewed.
var wantSecurityHeaders = map[string]string{
	"Content-Security-Policy": "default-src 'self'; img-src 'self' data:; style-src 'self'; font-src 'self'; script-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'",
	"X-Content-Type-Options":  "nosniff",
	"Referrer-Policy":         "no-referrer",
	"X-Frame-Options":         "DENY",
	"Permissions-Policy":      "geolocation=(), microphone=(), camera=()",
	"X-Robots-Tag":            "noindex, nofollow",
}

func assertSecurityHeaders(t *testing.T, header http.Header, path string) {
	t.Helper()

	for name, want := range wantSecurityHeaders {
		if got := header.Get(name); got != want {
			t.Errorf("%s on %s = %q, want %q", name, path, got, want)
		}
		// A second value would be intersected by the browser, which is how a page
		// silently stops loading after a proxy adds its own CSP.
		if values := header.Values(name); len(values) > 1 {
			t.Errorf("%s on %s set %d times, want 1", name, path, len(values))
		}
	}
}

func TestSecurityHeadersOnAPIResponse(t *testing.T) {
	app := newTestApp(t)

	response, err := http.Get(app.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET /api/health: %v", err)
	}
	defer response.Body.Close()

	assertSecurityHeaders(t, response.Header, "/api/health")
}

// TestSecurityHeadersOnErrorResponse guards the case that matters most in practice:
// an error path that writes its own response must not bypass the headers.
func TestSecurityHeadersOnErrorResponse(t *testing.T) {
	app := newTestApp(t)

	response, err := http.Get(app.URL + "/api/does-not-exist")
	if err != nil {
		t.Fatalf("GET /api/does-not-exist: %v", err)
	}
	defer response.Body.Close()

	assertSecurityHeaders(t, response.Header, "/api/does-not-exist")
}

// TestSecurityHeadersOnNonAPIResponse covers the SPA fallback: index.html carries
// the same headers as the API, which is where the CSP actually matters.
func TestSecurityHeadersOnNonAPIResponse(t *testing.T) {
	app := newTestApp(t)

	response, err := http.Get(app.URL + "/rsvp")
	if err != nil {
		t.Fatalf("GET /rsvp: %v", err)
	}
	defer response.Body.Close()

	assertSecurityHeaders(t, response.Header, "/rsvp")
}

func TestRobotsTxtDisallowsEverything(t *testing.T) {
	app := newTestApp(t)

	response := app.get("/robots.txt")

	if response.Status != http.StatusOK {
		t.Errorf("status = %d, want %d", response.Status, http.StatusOK)
	}
	if !strings.HasPrefix(response.ContentType, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", response.ContentType)
	}
	if want := "User-agent: *\nDisallow: /"; response.Body != want {
		t.Errorf("body = %q, want %q", response.Body, want)
	}
}
