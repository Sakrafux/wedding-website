package middleware

import (
	"net/http"
	"strings"
)

// StripPublicBasePath removes the deployment's public path prefix from the request
// path when the request still carries it.
//
// In production it never fires: Caddy's `handle_path /hochzeit*` strips the
// prefix before the request arrives, which is why the prefix has to be configured
// rather than inferred. It exists so the binary is reachable at the same URL with or
// without that proxy — `make preview`, a curl straight at the container, and a
// future proxy rule that forwards the prefix all behave like production instead of
// serving index.html and then 404-ing every asset under it.
//
// basePath is the normalized form from configuration: leading slash, no trailing
// slash, or "/" for a deployment at the root, in which case this is a no-op.
func StripPublicBasePath(basePath string) func(http.Handler) http.Handler {
	if basePath == "" || basePath == "/" {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			remainder, found := strings.CutPrefix(request.URL.Path, basePath)
			// Not prefixed (the proxy already stripped it), or a path that merely
			// starts with the same letters — /hochzeitcake is somebody else's.
			if !found || (remainder != "" && !strings.HasPrefix(remainder, "/")) {
				next.ServeHTTP(writer, request)
				return
			}
			// "/hochzeit" with no trailing slash still means the app root; an
			// empty path would reach chi as a route nothing matches.
			if remainder == "" {
				remainder = "/"
			}

			stripped := request.Clone(request.Context())
			stripped.URL.Path = remainder
			// RawPath is only set when it differs from the decoded path; leaving a
			// stale one behind would make the two disagree.
			stripped.URL.RawPath = ""

			next.ServeHTTP(writer, stripped)
		})
	}
}
