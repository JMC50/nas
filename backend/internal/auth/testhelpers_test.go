package auth

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/db"
)

const authTestSchema = `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userId TEXT UNIQUE NOT NULL,
		username TEXT NOT NULL,
		global_name TEXT,
		krname TEXT,
		password TEXT,
		auth_type TEXT
	);
	CREATE TABLE user_intents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		intent TEXT NOT NULL
	);
	CREATE TABLE log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		activity TEXT NOT NULL,
		description TEXT,
		user_id INTEGER NOT NULL,
		time INTEGER NOT NULL,
		loc TEXT
	);
`

func setupTestDBForAuth(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	_, err = conn.Exec(authTestSchema)
	require.NoError(t, err)
	return conn
}
