// Package configuration parses the environment, opens the database with the right
// pragmas and pool limits, and wires the dependency graph.
//
// Configuration is environment variables only — no config file. A missing required
// variable is a hard failure at startup, never a silent default, because a wedding
// app that quietly comes up with the wrong DB_PATH looks healthy while losing data.
package configuration
