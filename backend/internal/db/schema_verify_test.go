package db

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const fullSchema = `
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
		intent TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE TABLE log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		activity TEXT NOT NULL,
		description TEXT,
		user_id INTEGER NOT NULL,
		time INTEGER NOT NULL,
		loc TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
	);
`

func TestVerifySchema_AcceptsCurrentSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	conn, err := Open(dbPath)
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.Exec(fullSchema)
	require.NoError(t, err)

	err = VerifySchema(conn)
	require.NoError(t, err)
}

func TestVerifySchema_FailsOnMissingTable(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	conn, err := Open(dbPath)
	require.NoError(t, err)
	defer conn.Close()

	err = VerifySchema(conn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "users")
}

func TestVerifySchema_FailsOnMissingColumn(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	conn, err := Open(dbPath)
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		userId TEXT,
		username TEXT,
		global_name TEXT,
		krname TEXT,
		password TEXT
	);
	CREATE TABLE user_intents (id INTEGER PRIMARY KEY, user_id INTEGER, intent TEXT);
	CREATE TABLE log (id INTEGER PRIMARY KEY, activity TEXT, description TEXT, user_id INTEGER, time INTEGER, loc TEXT);`)
	require.NoError(t, err)

	err = VerifySchema(conn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "auth_type")
}
