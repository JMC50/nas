package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInsertAndGetLogs(t *testing.T) {
	conn := setupTestDB(t)
	_, err := SaveLocalUser(conn, "user-1", "alice", "h", "alice")
	require.NoError(t, err)

	require.NoError(t, InsertLog(conn, "user-1", "UPLOAD", "uploaded test.txt", "/dir", time.Now().UnixMilli()))
	require.NoError(t, InsertLog(conn, "user-1", "DELETE", "deleted test.txt", "/dir", time.Now().UnixMilli()+1))

	logs, err := GetActivityLogs(conn)
	require.NoError(t, err)
	require.Len(t, logs, 2)
	// Logs are sorted by time DESC, so most recent first
	require.Equal(t, "DELETE", logs[0].Activity)
	require.Equal(t, "UPLOAD", logs[1].Activity)
	require.Equal(t, "alice", logs[0].Username)
}

func TestInsertLog_RejectsUnknownUser(t *testing.T) {
	conn := setupTestDB(t)
	err := InsertLog(conn, "ghost", "UPLOAD", "test", "/", time.Now().UnixMilli())
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}
