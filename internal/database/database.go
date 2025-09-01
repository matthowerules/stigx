package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
	path string
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open SQLite database
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure SQLite settings
	if _, err := conn.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA cache_size = 1000;
		PRAGMA foreign_keys = ON;
		PRAGMA temp_store = memory;
	`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to configure database: %w", err)
	}

	db := &DB{
		conn: conn,
		path: dbPath,
	}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// GetConnection returns the underlying SQL connection
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

// initSchema loads and executes the schema
func (db *DB) initSchema() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	if _, err := db.conn.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// Health checks if the database is healthy
func (db *DB) Health() error {
	return db.conn.Ping()
}

// GetStats returns database statistics
func (db *DB) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get table counts
	queries := map[string]string{
		"stigs_count":      "SELECT COUNT(*) FROM stigs",
		"rules_count":      "SELECT COUNT(*) FROM stig_rules", 
		"checklists_count": "SELECT COUNT(*) FROM checklists",
		"bundles_count":    "SELECT COUNT(*) FROM bundles",
	}

	for key, query := range queries {
		var count int
		if err := db.conn.QueryRow(query).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", key, err)
		}
		stats[key] = count
	}

	// Get database file info
	if fileInfo, err := os.Stat(db.path); err == nil {
		stats["db_size_bytes"] = fileInfo.Size()
		stats["db_path"] = db.path
	}

	return stats, nil
}