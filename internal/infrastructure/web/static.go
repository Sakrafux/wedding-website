package web

import (
	"io/fs"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/go-chi/httplog/v2"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

const (
	// indexFile is both the entry point and the fallback: a client-side route like
	// /rsvp has no file of its own, and the router in the browser resolves it once
	// the shell has loaded.
	indexFile = "index.html"

	// hashedAssetPrefix is Vite's output directory for content-hashed files. The
	// hash is in the filename, so a change produces a new URL and the old one can be
	// cached forever.
	hashedAssetPrefix = "assets/"

	immutableCacheControl = "public, max-age=31536000, immutable"

	// index.html must never be cached: it names the hashed assets, so a stale copy
	// after a redeploy points at files that no longer exist and the app fails to
	// boot for exactly the guests whose browser kept it. "no-cache" still allows a
	// revalidated 304, it only forbids using the copy without asking.
	indexCacheControl = "no-cache"
)

// contentTypesByExtension covers every extension the Vite build can emit here.
//
// Set explicitly rather than left to mime.TypeByExtension, whose answers come from
// the container's /etc/mime.types and can therefore differ between the dev machine
// and the deployed image — a .js served as text/plain is a blank page, and with
// X-Content-Type-Options: nosniff the browser will not rescue it.
var contentTypesByExtension = map[string]string{
	".html":        "text/html; charset=utf-8",
	".css":         "text/css; charset=utf-8",
	".js":          "text/javascript; charset=utf-8",
	".json":        "application/json; charset=utf-8",
	".txt":         "text/plain; charset=utf-8",
	".svg":         "image/svg+xml",
	".ico":         "image/x-icon",
	".png":         "image/png",
	".jpg":         "image/jpeg",
	".jpeg":        "image/jpeg",
	".webp":        "image/webp",
	".avif":        "image/avif",
	".woff2":       "font/woff2",
	".webmanifest": "application/manifest+json",
}

// staticHandler serves the embedded SPA: a real file when the path names one,
// index.html otherwise.
type staticHandler struct {
	bundle fs.FS
	// index is read once at construction because it answers most requests; nil when
	// the binary was built without a frontend build, which every request then reports.
	index []byte
}

// newStaticHandler reads index.html out of the embedded bundle and returns the
// handler for every path outside /api.
//
// A missing index.html is not fatal at startup: `go build` on a clean checkout
// embeds only web/dist/.gitkeep, and the API — including /api/health — must still
// come up so that the backend can be worked on without running the frontend build.
// The failure surfaces per request instead.
func newStaticHandler(bundle fs.FS) *staticHandler {
	index, err := fs.ReadFile(bundle, indexFile)
	if err != nil {
		index = nil
	}

	return &staticHandler{bundle: bundle, index: index}
}

func (handler *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		httpio.RespondError(w, r, httpio.ErrMethodNotAllowed)
		return
	}

	name := path.Clean(strings.TrimPrefix(r.URL.Path, "/"))

	// fs.ValidPath rejects the traversal attempts ("..", absolute paths) that
	// path.Clean leaves behind; "." is what an empty path cleans to, i.e. the root.
	if name == "." || name == indexFile || !fs.ValidPath(name) {
		handler.serveIndex(w, r)
		return
	}

	// A directory, like a missing file, reads as an error here and falls through to
	// the shell — there are no directory listings to serve.
	content, err := fs.ReadFile(handler.bundle, name)
	if err != nil {
		handler.serveIndex(w, r)
		return
	}

	cacheControl := indexCacheControl
	if strings.HasPrefix(name, hashedAssetPrefix) {
		cacheControl = immutableCacheControl
	}

	writeFile(w, r, name, content, cacheControl)
}

// serveIndex answers with the SPA shell and status 200.
//
// 200 and not 404: the path is not missing, it is a client-side route. A 404 would
// make the browser's history and any error reporting treat a working page as broken.
func (handler *staticHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	if handler.index == nil {
		httplog.LogEntry(r.Context()).Error("no frontend bundle embedded, run the frontend build before go build", "path", r.URL.Path)
		httpio.RespondError(w, r, httpio.ErrInternal)
		return
	}

	writeFile(w, r, indexFile, handler.index, indexCacheControl)
}

// writeFile writes content with an explicit content type, length and cache policy.
func writeFile(w http.ResponseWriter, r *http.Request, name string, content []byte, cacheControl string) {
	contentType, isKnownExtension := contentTypesByExtension[path.Ext(name)]
	if !isKnownExtension {
		// Unknown means the build started emitting something the table does not list.
		// application/octet-stream makes the browser download rather than misinterpret
		// it, and the log line says which extension to add.
		httplog.LogEntry(r.Context()).Warn("no content type mapped for asset", "asset", name)
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Cache-Control", cacheControl)
	w.WriteHeader(http.StatusOK)

	// net/http discards the body for a HEAD request, so this needs no special case.
	if _, err := w.Write(content); err != nil {
		httplog.LogEntry(r.Context()).Debug("static response write failed", "asset", name, "error", err)
	}
}
