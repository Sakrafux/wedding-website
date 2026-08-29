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
	app := newTestApp(t)

	response := app.get("/")

	if response.Status != http.StatusOK {
		t.Errorf("status = %d, want %d", response.Status, http.StatusOK)
	}
	if !strings.HasPrefix(response.ContentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", response.ContentType)
	}
	if !strings.Contains(response.Body, `<div id="root">`) {
		t.Errorf("body is not the SPA shell: %q", response.Body)
	}
}

// TestClientSideRouteServesIndexHTML is the deep-link refresh: /rsvp exists only in
// the browser's router, and must answer the shell with 200 rather than a 404.
func TestClientSideRouteServesIndexHTML(t *testing.T) {
	requireBundle(t)
	app := newTestApp(t)

	response := app.get("/rsvp")

	if response.Status != http.StatusOK {
		t.Errorf("status = %d, want %d", response.Status, http.StatusOK)
	}
	if !strings.HasPrefix(response.ContentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", response.ContentType)
	}
	if !strings.Contains(response.Body, `<div id="root">`) {
		t.Errorf("body is not the SPA shell: %q", response.Body)
	}
}

// TestUnknownAPIPathStillReturnsJSONWithBundlePresent is the boundary that only
// becomes testable now: with an SPA fallback registered, /api must keep answering
// JSON instead of falling through to index.html. TestUnknownAPIPathReturnsJSON in
// health_test.go asserts the same thing without the bundle.
func TestUnknownAPIPathStillReturnsJSONWithBundlePresent(t *testing.T) {
	requireBundle(t)
	app := newTestApp(t)

	response := app.get("/api/unknown")

	if response.Status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", response.Status, http.StatusNotFound)
	}
	if !strings.HasPrefix(response.ContentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", response.ContentType)
	}
	if !strings.Contains(response.Body, `"code":"not_found"`) {
		t.Errorf("body = %q, want the error envelope", response.Body)
	}
}

func TestHashedAssetIsImmutableAndIndexIsNot(t *testing.T) {
	bundle := requireBundle(t)
	app := newTestApp(t)

	asset := firstHashedAsset(t, bundle)

	assetResponse, err := http.Get(app.URL + "/" + asset)
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

	indexResponse, err := http.Get(app.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer indexResponse.Body.Close()

	if want := "no-cache"; indexResponse.Header.Get("Cache-Control") != want {
		t.Errorf("Cache-Control for / = %q, want %q", indexResponse.Header.Get("Cache-Control"), want)
	}
}

// TestBundleReferencesThePublicBasePath guards the one thing the reverse proxy
// cannot fix. Caddy serves the app under /hochzeit and strips that prefix, so Go
// answers at the root either way — but the browser resolves the asset URLs in
// index.html against the public path, and a bundle built without Vite's `base` would
// point at /assets/…, which on that host belongs to another app entirely.
//
// The literal is repeated here on purpose: it is the deployment contract, and a test
// that imported it from the frontend build would follow a change instead of catching
// it. It must match PUBLIC_BASE_PATH and the handle_path rule in the Caddyfile.
func TestBundleReferencesThePublicBasePath(t *testing.T) {
	requireBundle(t)
	app := newTestApp(t)

	response := app.get("/")

	if !strings.Contains(response.Body, `src="/hochzeit/assets/`) {
		t.Errorf("index.html does not reference assets under the public base path:\n%s", response.Body)
	}
}

// TestPrefixedPathsAreServedWhenTheProxyDoesNotStrip covers the other half of the
// deployment: production runs behind handle_path, which strips /hochzeit, but a
// curl straight at the container or `make preview` does not. Both must answer the
// same thing, or the only way to check a container is through the proxy.
func TestPrefixedPathsAreServedWhenTheProxyDoesNotStrip(t *testing.T) {
	requireBundle(t)
	app := newTestAppWithBasePath(t, "/hochzeit", nil)

	for path, wantContentType := range map[string]string{
		"/hochzeit":            "text/html",
		"/hochzeit/":           "text/html",
		"/hochzeit/rsvp":       "text/html",
		"/hochzeit/api/health": "application/json",
	} {
		t.Run(path, func(t *testing.T) {
			response := app.get(path)

			if response.Status != http.StatusOK {
				t.Errorf("status = %d, want %d", response.Status, http.StatusOK)
			}
			if !strings.HasPrefix(response.ContentType, wantContentType) {
				t.Errorf("Content-Type = %q, want %q", response.ContentType, wantContentType)
			}
		})
	}
}

// TestUnknownAssetPathServesIndexHTML documents the accepted trade-off: outside
// /api nothing 404s, because the handler cannot tell a typo from a route the
// browser knows about.
func TestUnknownAssetPathServesIndexHTML(t *testing.T) {
	requireBundle(t)
	app := newTestApp(t)

	response := app.get("/assets/does-not-exist.js")

	if response.Status != http.StatusOK {
		t.Errorf("status = %d, want %d", response.Status, http.StatusOK)
	}
	if !strings.HasPrefix(response.ContentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", response.ContentType)
	}
}
