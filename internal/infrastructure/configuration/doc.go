// Package configuration parses the environment and opens the database with the
// right pragmas and pool limits.
//
// Configuration is environment variables only — no config file. A missing required
// variable is a hard failure at startup, never a silent default, because a wedding
// app that quietly comes up with the wrong DB_PATH looks healthy while losing data.
//
// The package imports no other internal package and is meant to stay that way: it
// sits at the bottom of the graph so anything may depend on it. Wiring the
// dependency graph is cmd/wedding's job, not this package's — doing it here would
// mean importing web, and then web could not import this.
package configuration
