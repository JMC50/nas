package db

import (
	"database/sql"
	"fmt"
)

// tableSchema is one required table with its columns.
// Slice (not map) so verification order is deterministic and tests are stable.
type tableSchema struct {
	name    string
	columns []string
}

// requiredSchema is checked in order on startup.
// We only verify presence — destructive ALTER is forbidden (data continuity guarantee).
// Column types intentionally NOT verified — bcryptjs hashes were stored as TEXT
// and remain TEXT under Go bcrypt. If a future migration changes types, update this slice.
var requiredSchema = []tableSchema{
	{"users", []string{"id", "userId", "username", "global_name", "krname", "password", "auth_type"}},
	{"user_intents", []string{"id", "user_id", "intent"}},
	{"log", []string{"id", "activity", "description", "user_id", "time", "loc"}},
	{"server_settings", []string{"key", "value"}},
}

func VerifySchema(conn *sql.DB) error {
	for _, table := range requiredSchema {
		columns, err := tableColumns(conn, table.name)
		if err != nil {
			return fmt.Errorf("introspect %s: %w", table.name, err)
		}
		for _, required := range table.columns {
			if _, ok := columns[required]; !ok {
				return fmt.Errorf("table %s missing column %s", table.name, required)
			}
		}
	}
	return nil
}

func tableColumns(conn *sql.DB, table string) (map[string]struct{}, error) {
	rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := map[string]struct{}{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		columns[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("table %s does not exist", table)
	}
	return columns, nil
}
