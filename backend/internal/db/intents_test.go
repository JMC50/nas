package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToggleIntent_AddsThenRemoves(t *testing.T) {
	conn := setupTestDB(t)
	_, err := SaveLocalUser(conn, "user-x", "x", "hash", "x")
	require.NoError(t, err)

	has, err := HasIntent(conn, "user-x", "UPLOAD")
	require.NoError(t, err)
	require.False(t, has)

	require.NoError(t, ToggleIntent(conn, "user-x", "UPLOAD"))
	has, err = HasIntent(conn, "user-x", "UPLOAD")
	require.NoError(t, err)
	require.True(t, has)

	require.NoError(t, ToggleIntent(conn, "user-x", "UPLOAD"))
	has, err = HasIntent(conn, "user-x", "UPLOAD")
	require.NoError(t, err)
	require.False(t, has)
}

func TestHasIntent_AdminGrantsAll(t *testing.T) {
	conn := setupTestDB(t)
	_, err := SaveLocalUser(conn, "admin-user", "admin", "hash", "admin")
	require.NoError(t, err)
	require.NoError(t, ToggleIntent(conn, "admin-user", "ADMIN"))

	for _, intent := range []string{"UPLOAD", "DELETE", "RENAME", "COPY", "VIEW", "DOWNLOAD", "OPEN"} {
		has, err := HasIntent(conn, "admin-user", intent)
		require.NoError(t, err)
		require.True(t, has, "admin must have %s", intent)
	}
}

func TestHasIntent_NonexistentUser(t *testing.T) {
	conn := setupTestDB(t)
	has, err := HasIntent(conn, "ghost", "VIEW")
	require.NoError(t, err)
	require.False(t, has)
}
