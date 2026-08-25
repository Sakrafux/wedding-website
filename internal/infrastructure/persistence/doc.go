// Package persistence holds the sqlx-backed stores and the embedded migrations.
//
// Stores are concrete types, not interfaces. SQL lives here and only here.
// Migrations are numbered .sql files applied in order at startup, forward-only.
package persistence
