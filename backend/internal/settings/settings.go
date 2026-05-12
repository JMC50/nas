// Package settings exposes a key/value store backed by the server_settings table.
// Used by admin features (e.g. OAuth provider credentials) to override boot-time
// env values without redeploy. Values are stored verbatim; callers handle masking
// for any secret fields.
package settings

import (
	"database/sql"
	"errors"
	"fmt"
)

func Get(conn *sql.DB, key string) (string, error) {
	row := conn.QueryRow(`SELECT value FROM server_settings WHERE key = ?`, key)
	var value string
	if err := row.Scan(&value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("scan setting %s: %w", key, err)
	}
	return value, nil
}

func Set(conn *sql.DB, key, value string) error {
	_, err := conn.Exec(`
		INSERT INTO server_settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	if err != nil {
		return fmt.Errorf("upsert setting %s: %w", key, err)
	}
	return nil
}

func GetAll(conn *sql.DB) (map[string]string, error) {
	rows, err := conn.Query(`SELECT key, value FROM server_settings`)
	if err != nil {
		return nil, fmt.Errorf("query settings: %w", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan setting row: %w", err)
		}
		out[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate settings: %w", err)
	}
	return out, nil
}
