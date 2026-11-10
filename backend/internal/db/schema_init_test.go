package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// v0.6.0이 만든 데이터베이스의 users 표
const usersBeforeTokensValidAfter = `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userId TEXT UNIQUE NOT NULL,
		username TEXT NOT NULL,
		global_name TEXT,
		krname TEXT,
		password TEXT,
		auth_type TEXT
	);
`

func openEmptyDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestInitSchema_AddsTokensValidAfterToOlderDatabase(t *testing.T) {
	conn := openEmptyDB(t)
	_, err := conn.Exec(usersBeforeTokensValidAfter)
	require.NoError(t, err)
	_, err = conn.Exec(`INSERT INTO users (userId, username, password, auth_type) VALUES ('alice', 'Alice', 'hash', 'local')`)
	require.NoError(t, err)

	require.NoError(t, InitSchema(conn))
	require.NoError(t, InitSchema(conn), "a second start must not fail on the existing column")
	require.NoError(t, VerifySchema(conn))

	validAfter, found, err := TokensValidAfter(conn, "alice")
	require.NoError(t, err)
	require.True(t, found)
	require.Zero(t, validAfter, "an upgraded user has no restriction")

	user, err := GetUser(conn, "alice")
	require.NoError(t, err)
	require.Equal(t, "Alice", user.Username)
	require.Equal(t, "hash", user.Password)
}

func TestInitSchema_NewDatabaseHasTokensValidAfter(t *testing.T) {
	conn := openEmptyDB(t)
	require.NoError(t, InitSchema(conn))
	require.NoError(t, VerifySchema(conn))

	_, err := SaveLocalUser(conn, "bob", "bob", "hash", "")
	require.NoError(t, err)
	validAfter, found, err := TokensValidAfter(conn, "bob")
	require.NoError(t, err)
	require.True(t, found)
	require.Zero(t, validAfter)
}

func TestTokensValidAfter_SetAndMissingUser(t *testing.T) {
	conn := openEmptyDB(t)
	require.NoError(t, InitSchema(conn))
	id, err := SaveLocalUser(conn, "carol", "carol", "hash", "")
	require.NoError(t, err)

	require.NoError(t, SetTokensValidAfter(conn, id, 1_800_000_000_000))
	validAfter, found, err := TokensValidAfter(conn, "carol")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(1_800_000_000_000), validAfter)

	_, found, err = TokensValidAfter(conn, "nobody")
	require.NoError(t, err)
	require.False(t, found)
}
