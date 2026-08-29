package integration

import (
	"io/fs"
	"net/http"
	"strings"
	"testing"

	frontend "github.com/Sakrafux/wedding-website/web"
)

// requireBundle skips the test when the binary was built without a frontend build.
//
// The embedded bundle is fixed at compile time, and on a clean checkout web/dist
// holds only .gitkeep. Skipping keeps `go test ./...` usable for backend work
// without running pnpm first; `make build` runs the frontend build ahead of the
// binary, so the real artefact is never the skipped case.
func requireBundle(t *testing.T) fs.FS {
	t.Helper()

	bundle := frontend.Bundle()
	if _, err := fs.Stat(bundle, "index.html"); err != nil {
		t.Skip("no frontend bundle embedded; run `make build-web` first")
	}

	return bundle
}

// firstHashedAsset returns the name of any file under assets/, since the names are
// content-hashed and no test can hardcode one.
func firstHashedAsset(t *testing.T, bundle fs.FS) string {
	t.Helper()

	entries, err := fs.ReadDir(bundle, "assets")
	if err != nil {
		t.Fatalf("reading embedded assets/: %v", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			return "assets/" + entry.Name()
		}
	}

	t.Fatal("embedded assets/ contains no files")
	return ""
}

func TestRootServesIndexHTML(t *testing.T) {
	requireBundle(t)
	server := newTestServer(t)

	status, contentType, body := get(t, server.URL, "/")

	if status != http.StatusOK {
		t.Errorf("status = %d, want %d", status, http.StatusOK)
	}
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", contentType)
	}
	if !strings.Contains(body, `<div id="root">`) {
		t.Errorf("body is not the SPA shell: %q", body)
	}
}

// TestClientSideRouteServesIndexHTML is the deep-link refresh: /rsvp exists only in
// the browser's router, and must answer the shell with 200 rather than a 404.
func TestClientSideRouteServesIndexHTML(t *testing.T) {
	requireBundle(t)
	server := newTestServer(t)

	status, contentType, body := get(t, server.URL, "/rsvp")

	if status != http.StatusOK {
		t.Errorf("status = %d, want %d", status, http.StatusOK)
	}
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", contentType)
	}
	if !strings.Contains(body, `<div id="root">`) {
		t.Errorf("body is not the SPA shell: %q", body)
	}
}

// TestUnknownAPIPathStillReturnsJSONWithBundlePresent is the boundary that only
// becomes testable now: with an SPA fallback registered, /api must keep answering
// JSON instead of falling through to index.html. TestUnknownAPIPathReturnsJSON in
// health_test.go asserts the same thing without the bundle.
func TestUnknownAPIPathStillReturnsJSONWithBundlePresent(t *testing.T) {
	requireBundle(t)
	server := newTestServer(t)

	status, contentType, body := get(t, server.URL, "/api/unknown")

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", status, http.StatusNotFound)
	}
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
	if !strings.Contains(body, `"code":"not_found"`) {
		t.Errorf("body = %q, want the error envelope", body)
	}
}

func TestHashedAssetIsImmutableAndIndexIsNot(t *testing.T) {
	bundle := requireBundle(t)
	server := newTestServer(t)

	asset := firstHashedAsset(t, bundle)

	assetResponse, err := http.Get(server.URL + "/" + asset)
	if err != nil {
		t.Fatalf("GET /%s: %v", asset, err)
	}
	defer assetResponse.Body.Close()

	if assetResponse.StatusCode != http.StatusOK {
		t.Errorf("status for /%s = %d, want %d", asset, assetResponse.StatusCode, http.StatusOK)
	}
	if want := "public, max-age=31536000, immutable"; assetResponse.Header.Get("Cache-Control") != want {
		t.Errorf("Cache-Control for /%s = %q, want %q", asset, assetResponse.Header.Get("Cache-Control"), want)
	}

	indexResponse, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer indexResponse.Body.Close()

	if want := "no-cache"; indexResponse.Header.Get("Cache-Control") != want {
		t.Errorf("Cache-Control for / = %q, want %q", indexResponse.Header.Get("Cache-Control"), want)
	}
}

// TestUnknownAssetPathServesIndexHTML documents the accepted trade-off: outside
// /api nothing 404s, because the handler cannot tell a typo from a route the
// browser knows about.
func TestUnknownAssetPathServesIndexHTML(t *testing.T) {
	requireBundle(t)
	server := newTestServer(t)

	status, contentType, _ := get(t, server.URL, "/assets/does-not-exist.js")

	if status != http.StatusOK {
		t.Errorf("status = %d, want %d", status, http.StatusOK)
	}
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", contentType)
	}
}
