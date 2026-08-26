package configuration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	// Pure-Go SQLite driver, registered under the name "sqlite". No CGO, so the
	// binary stays statically linked and the Docker image needs no build toolchain.
	_ "modernc.org/sqlite"
)

const (
	driverName = "sqlite"

	// maxReadConnections is a handful, not a pool sized for a web scale that does
	// not exist here: 60 households, and WAL lets readers run while the single
	// writer works.
	maxReadConnections = 4

	// databaseDirPermissions is restrictive because the file below it is a secret:
	// household login codes are stored in plaintext so they can be reprinted.
	databaseDirPermissions = 0o700

	// pingTimeout bounds the readiness check. Shorter than busy_timeout (5s) on
	// purpose — a readiness probe should answer, not queue behind a busy writer.
	pingTimeout = 2 * time.Second
)

// Database holds the two handles on the one SQLite file.
//
// Writes and reads are separated because SQLite has exactly one writer: the write
// pool is capped at a single connection so that concurrent RSVP submissions queue
// in Go instead of racing in SQLite and surfacing as SQLITE_BUSY. Reads run on
// their own pool, which WAL mode allows to proceed while a write is in flight.
type Database struct {
	// Write is the only handle allowed to mutate. Capped at one connection.
	Write *sqlx.DB
	// Read is opened read-only, so a query that tries to write fails loudly here
	// rather than quietly stealing the writer slot.
	Read *sqlx.DB
}

// OpenDatabase opens the write and read handles on config.DatabasePath, creating the
// parent directory if it is missing.
//
// Both handles are verified with a ping before returning, so a bad path or an
// unmounted volume fails at startup rather than on the first guest request. On any
// error nothing is left open.
func OpenDatabase(config Config) (*Database, error) {
	// A fresh volume mount is an empty directory; requiring a manual mkdir before
	// the first start is a deploy step nobody remembers.
	if err := os.MkdirAll(filepath.Dir(config.DatabasePath), databaseDirPermissions); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	writePool, err := openPool(writeDataSourceName(config.DatabasePath), 1)
	if err != nil {
		return nil, fmt.Errorf("opening write pool: %w", err)
	}

	// The write pool is opened and pinged first because it creates the database file
	// and, by switching on WAL, its -wal and -shm siblings. A read-only handle
	// cannot create any of them.
	readPool, err := openPool(readDataSourceName(config.DatabasePath), maxReadConnections)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("opening read pool: %w", err), writePool.Close())
	}

	return &Database{Write: writePool, Read: readPool}, nil
}

// Ping reports whether both handles answer within pingTimeout.
//
// Both are checked: a read-only handle failing while writes work is exactly the
// asymmetric failure a single check would miss — the read pool needs the -shm file,
// so a permissions or volume problem can hit it alone.
func (database *Database) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := database.Write.PingContext(pingCtx); err != nil {
		return fmt.Errorf("write pool: %w", err)
	}
	if err := database.Read.PingContext(pingCtx); err != nil {
		return fmt.Errorf("read pool: %w", err)
	}
	return nil
}

// Close closes both handles, reporting every failure rather than the first: with
// WAL, a failure to close is worth seeing in the shutdown log.
func (database *Database) Close() error {
	return errors.Join(database.Write.Close(), database.Read.Close())
}

// openPool opens one handle and proves it works.
//
// database/sql connects lazily, so without the ping a wrong path or a corrupt file
// would first surface inside an unrelated request handler.
func openPool(dataSourceName string, maxOpenConnections int) (*sqlx.DB, error) {
	pool, err := sqlx.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	pool.SetMaxOpenConns(maxOpenConnections)
	// Idle equals open: connections to a local file are cheap to hold and expensive
	// to re-establish only in that each new one re-runs the pragma list.
	pool.SetMaxIdleConns(maxOpenConnections)
	// No ConnMaxLifetime: recycling exists for network databases that drop
	// connections and for failing over to a new host. Neither applies to a file.

	pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := pool.PingContext(pingCtx); err != nil {
		return nil, errors.Join(err, pool.Close())
	}
	return pool, nil
}

// writeDataSourceName builds the DSN for the writing handle.
//
// The pragmas are carried in the DSN, not executed after opening: the driver applies
// them to every connection it creates, whereas an "PRAGMA ..." after sql.Open
// configures whichever single connection served that statement and leaves the rest
// of the pool with foreign_keys = OFF — a bug that surfaces as orphaned rows months
// later, if at all.
func writeDataSourceName(path string) string {
	return dataSourceName(path,
		// WAL: readers never block the writer, and the write pool is the only handle
		// that may set it, since switching journal mode writes to the file header.
		"_pragma=journal_mode(WAL)",
		// NORMAL fsyncs at checkpoints rather than every commit. With WAL the risk is
		// losing the last commits on OS or hardware failure, not corruption, and this
		// is RSVP data on a single server, not a ledger.
		"_pragma=synchronous(NORMAL)",
		"_pragma=foreign_keys(1)",
		"_pragma=busy_timeout(5000)",
		// IMMEDIATE takes the write lock when the transaction starts instead of on its
		// first write. A deferred transaction that upgrades mid-flight can fail with
		// SQLITE_BUSY that busy_timeout cannot resolve, because the upgrade may
		// require another connection to release a lock it is still using.
		"_txlock=immediate",
	)
}

// readDataSourceName builds the DSN for the read-only handle.
//
// journal_mode is deliberately absent: setting it writes to the file header, which a
// read-only connection cannot do. The mode is a property of the file, already set by
// the write handle, and reading `PRAGMA journal_mode` here still reports `wal`.
func readDataSourceName(path string) string {
	return dataSourceName(path,
		"mode=ro",
		"_pragma=synchronous(NORMAL)",
		// Harmless on a read-only connection today, and correct the moment someone
		// runs a query through the read pool expecting the same semantics.
		"_pragma=foreign_keys(1)",
		"_pragma=busy_timeout(5000)",
	)
}

func dataSourceName(path string, parameters ...string) string {
	// The file: prefix is what makes the driver read the query parameters at all;
	// a bare path is taken literally, pragmas included.
	return "file:" + path + "?" + strings.Join(parameters, "&")
}
