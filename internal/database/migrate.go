package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB is the package-level database handle. Initialised by Open().
var DB *sql.DB

// Open opens (or creates) the SQLite database at the given path,
// sets the required PRAGMAs, and runs any pending migrations.
func Open(dbPath string) error {
	// CGo SQLite driver — the DSN supports query parameters for PRAGMAs
	// that must be set at connection time (e.g. _foreign_keys).
	dsn := fmt.Sprintf(
		"%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000",
		dbPath,
	)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	// Single writer; multiple readers via WAL. Keep one connection in the
	// write pool so we never get "database is locked" on the Pi.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	DB = db

	if err := runMigrations(db, migrationsDir()); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	log.Println("database ready:", dbPath)
	return nil
}

// Close closes the database. Call via defer after Open() succeeds.
func Close() {
	if DB != nil {
		DB.Close()
	}
}

// migrationsDir returns the path to the migrations folder relative to the
// binary's working directory. Override by setting MIGRATIONS_DIR env var.
func migrationsDir() string {
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir
	}
	return "db/migrations"
}

// ============================================================
// Migration runner
// ============================================================

// migration represents a single .sql file to be applied.
type migration struct {
	version  string // e.g. "001"
	filename string // e.g. "001_initial_schema.sql"
	path     string // absolute or relative path
}

// runMigrations ensures the schema_migrations table exists, then applies
// any SQL files in migrationsDir that have not yet been recorded.
// Files are applied in lexicographic order — the numeric prefix
// (001_, 002_, …) is what guarantees correct ordering.
func runMigrations(db *sql.DB, dir string) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	files, err := findMigrationFiles(dir)
	if err != nil {
		return err
	}

	applied, err := appliedMigrations(db)
	if err != nil {
		return err
	}

	pending := filterPending(files, applied)
	if len(pending) == 0 {
		log.Println("migrations: nothing to apply")
		return nil
	}

	for _, m := range pending {
		if err := applyMigration(db, m); err != nil {
			return fmt.Errorf("apply %s: %w", m.filename, err)
		}
	}

	return nil
}

// ensureMigrationsTable creates the tracking table if it does not exist.
func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     TEXT PRIMARY KEY,
			filename    TEXT    NOT NULL,
			applied_at  TEXT    NOT NULL DEFAULT (datetime('now'))
		)
	`)
	return err
}

// findMigrationFiles returns all *.sql files in dir, sorted by filename.
func findMigrationFiles(dir string) ([]migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	var migrations []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		version := extractVersion(e.Name())
		if version == "" {
			log.Printf("migrations: skipping %s (no numeric prefix)", e.Name())
			continue
		}
		migrations = append(migrations, migration{
			version:  version,
			filename: e.Name(),
			path:     filepath.Join(dir, e.Name()),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].filename < migrations[j].filename
	})

	return migrations, nil
}

// extractVersion pulls the leading numeric segment from a filename.
// "001_initial_schema.sql" → "001"
// "002_add_tags.sql"       → "002"
// "no_prefix.sql"          → ""
func extractVersion(filename string) string {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) == 0 {
		return ""
	}
	v := parts[0]
	for _, c := range v {
		if c < '0' || c > '9' {
			return ""
		}
	}
	if len(v) == 0 {
		return ""
	}
	return v
}

// appliedMigrations returns the set of version strings already in
// schema_migrations.
func appliedMigrations(db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = struct{}{}
	}
	return applied, rows.Err()
}

// filterPending returns migrations whose version is not in applied.
func filterPending(files []migration, applied map[string]struct{}) []migration {
	var pending []migration
	for _, f := range files {
		if _, ok := applied[f.version]; !ok {
			pending = append(pending, f)
		}
	}
	return pending
}

// applyMigration reads the SQL file and executes it inside a transaction.
// On success it records the version in schema_migrations.
func applyMigration(db *sql.DB, m migration) error {
	log.Printf("migrations: applying %s", m.filename)

	content, err := os.ReadFile(m.path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	// Always rollback if we return early; no-op after Commit().
	defer tx.Rollback()

	// Execute the full file. SQLite's Exec handles multiple statements
	// separated by semicolons when using the go-sqlite3 driver.
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("exec sql: %w", err)
	}

	// Record the migration inside the same transaction so that a failure
	// to record also rolls back the schema changes.
	_, err = tx.Exec(`
		INSERT INTO schema_migrations (version, filename, applied_at)
		VALUES (?, ?, ?)
	`, m.version, m.filename, time.Now().UTC().Format("2006-01-02T15:04:05"))
	if err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Printf("migrations: applied %s ✓", m.filename)
	return nil
}
