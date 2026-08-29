// Package frontend embeds the built React bundle so the binary carries the whole
// app: one artefact, and a frontend/backend skew that cannot happen.
//
// The file lives in web/ rather than next to the handler that serves it because an
// embed directive cannot reach outside its own package directory — no "../.." —
// and dist/ is written here by Vite. The package name deliberately differs from the
// directory name so it does not collide with internal/infrastructure/web.
package frontend

import (
	"embed"
	"io/fs"
)

// dist is the Vite output. The all: prefix is required: without it go:embed skips
// entries starting with a dot, and on a clean checkout dist/ holds nothing but
// .gitkeep — the embed would then fail the build with "no matching files".
//
//go:embed all:dist
var dist embed.FS

// Bundle returns the built frontend rooted at dist/, so callers address files the
// way the browser does ("index.html", not "dist/index.html").
//
// It panics on failure, which can only mean the embed directive above stopped
// matching — a build-time mistake, not a runtime condition.
func Bundle() fs.FS {
	bundle, err := fs.Sub(dist, "dist")
	if err != nil {
		panic("embedded frontend bundle is missing its dist root: " + err.Error())
	}

	return bundle
}
