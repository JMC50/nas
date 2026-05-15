package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	_, err = conn.Exec(fullSchema)
	require.NoError(t, err)
	return conn
}

func TestSaveOAuthUser_InsertsThenUpdates(t *testing.T) {
	conn := setupTestDB(t)

	id1, err := SaveOAuthUser(conn, "discord-123", "alice", "Alice", "앨리스")
	require.NoError(t, err)
	require.Greater(t, id1, int64(0))

	id2, err := SaveOAuthUser(conn, "discord-123", "alice2", "Alice2", "앨리스2")
	require.NoError(t, err)
	require.Equal(t, id1, id2, "second call must update, not insert")

	user, err := GetUser(conn, "discord-123")
	require.NoError(t, err)
	require.Equal(t, "alice2", user.Username)
	require.Equal(t, "앨리스2", user.KrName)
}

func TestGetUser_ReturnsNilWhenMissing(t *testing.T) {
	conn := setupTestDB(t)
	user, err := GetUser(conn, "nonexistent")
	require.NoError(t, err)
	require.Nil(t, user)
}

func TestSaveLocalUser_StoresPasswordAndAuthType(t *testing.T) {
	conn := setupTestDB(t)

	_, err := SaveLocalUser(conn, "local-1", "bob", "$2a$10$hash", "밥")
	require.NoError(t, err)

	user, err := GetUser(conn, "local-1")
	require.NoError(t, err)
	require.Equal(t, "$2a$10$hash", user.Password)
	require.Equal(t, "local", user.AuthType)
	require.Equal(t, "밥", user.KrName)
}

func TestUpdatePassword(t *testing.T) {
	conn := setupTestDB(t)
	_, err := SaveLocalUser(conn, "local-2", "carol", "$2a$10$old", "캐롤")
	require.NoError(t, err)

	require.NoError(t, UpdatePassword(conn, "local-2", "$2a$10$new"))

	user, err := GetUser(conn, "local-2")
	require.NoError(t, err)
	require.Equal(t, "$2a$10$new", user.Password)
}

func TestDeleteUser_CascadesIntentsAndIdentities(t *testing.T) {
	conn := setupTestDB(t)
	id, err := SaveLocalUser(conn, "local-del", "del", "h", "del")
	require.NoError(t, err)
	require.NoError(t, AddIdentity(conn, id, "local", "local-del"))
	require.NoError(t, ToggleIntent(conn, "local-del", "UPLOAD"))
	require.NoError(t, InsertLog(conn, "local-del", "act", "desc", "/", 0))

	require.NoError(t, DeleteUser(conn, "local-del"))

	user, err := GetUser(conn, "local-del")
	require.NoError(t, err)
	require.Nil(t, user)

	var intentCount, identityCount int
	require.NoError(t, conn.QueryRow("SELECT COUNT(*) FROM user_intents WHERE user_id = ?", id).Scan(&intentCount))
	require.Equal(t, 0, intentCount, "user_intents must cascade delete")
	require.NoError(t, conn.QueryRow("SELECT COUNT(*) FROM user_identities WHERE user_id = ?", id).Scan(&identityCount))
	require.Equal(t, 0, identityCount, "user_identities must cascade delete")

	var logCount int
	require.NoError(t, conn.QueryRow("SELECT COUNT(*) FROM log WHERE activity = ?", "act").Scan(&logCount))
	require.Equal(t, 0, logCount, "log rows for deleted user must be removed")
}

func TestDeleteUser_ReturnsErrUserNotFoundWhenMissing(t *testing.T) {
	conn := setupTestDB(t)
	err := DeleteUser(conn, "nonexistent")
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestGetAllUsers_AggregatesIntents(t *testing.T) {
	conn := setupTestDB(t)
	_, err := SaveLocalUser(conn, "local-a", "a", "h", "a")
	require.NoError(t, err)
	_, err = SaveOAuthUser(conn, "oauth-b", "b", "B", "b")
	require.NoError(t, err)

	// Give user-a UPLOAD intent
	require.NoError(t, ToggleIntent(conn, "local-a", "UPLOAD"))

	users, err := GetAllUsers(conn)
	require.NoError(t, err)
	require.Len(t, users, 2)

	intentByUser := map[string][]string{}
	for _, user := range users {
		intentByUser[user.UserID] = user.Intents
	}
	require.Equal(t, []string{"UPLOAD"}, intentByUser["local-a"])
	require.Empty(t, intentByUser["oauth-b"])
}
