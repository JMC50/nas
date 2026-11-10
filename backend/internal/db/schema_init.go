package db

import (
	"database/sql"
	"fmt"
)

// schemaDDL is the canonical CREATE TABLE statement for each table.
// Uses CREATE TABLE IF NOT EXISTS so an existing nas.sqlite is preserved untouched.
// Columns are kept identical to the legacy Node entity definitions so bcryptjs
// hashes and OAuth IDs remain compatible.
const schemaDDL = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	userId TEXT UNIQUE NOT NULL,
	username TEXT NOT NULL,
	global_name TEXT,
	krname TEXT,
	password TEXT,
	auth_type TEXT,
	tokens_valid_after INTEGER
);
CREATE TABLE IF NOT EXISTS user_intents (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	intent TEXT NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS log (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	activity TEXT NOT NULL,
	description TEXT,
	user_id INTEGER NOT NULL,
	time INTEGER NOT NULL,
	loc TEXT,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);
CREATE TABLE IF NOT EXISTS server_settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS user_identities (
	user_id INTEGER NOT NULL,
	provider TEXT NOT NULL,
	external_id TEXT NOT NULL,
	PRIMARY KEY (provider, external_id),
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_identities_user ON user_identities(user_id);
CREATE TABLE IF NOT EXISTS share_links (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	token_hash TEXT UNIQUE NOT NULL,
	user_id INTEGER NOT NULL,
	loc TEXT NOT NULL,
	name TEXT NOT NULL,
	is_folder INTEGER NOT NULL DEFAULT 0,
	created_at INTEGER NOT NULL,
	expires_at INTEGER NOT NULL,
	revoked_at INTEGER,
	download_count INTEGER NOT NULL DEFAULT 0,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_share_links_user ON share_links(user_id);
`

// InitSchema runs CREATE TABLE IF NOT EXISTS for the required tables, then adds
// the columns that an older database does not have yet.
// Safe to call on a populated DB — existing rows are preserved.
func InitSchema(conn *sql.DB) error {
	if _, err := conn.Exec(schemaDDL); err != nil {
		return fmt.Errorf("create tables: %w", err)
	}
	return addColumn(conn, "users", "tokens_valid_after", "INTEGER")
}

func addColumn(conn *sql.DB, table, column, columnType string) error {
	columns, err := tableColumns(conn, table)
	if err != nil {
		return fmt.Errorf("introspect %s: %w", table, err)
	}
	if _, ok := columns[column]; ok {
		return nil
	}
	if _, err := conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, columnType)); err != nil {
		return fmt.Errorf("add column %s.%s: %w", table, column, err)
	}
	return nil
}
