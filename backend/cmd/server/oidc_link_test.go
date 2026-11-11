package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/db"
)

// 구글 사용자 두 명이 있는 데이터베이스 파일과 CSV 파일의 경로를 돌려줌
func linkFixture(t *testing.T, csvText string) (dbPath, mapPath string) {
	t.Helper()
	dir := t.TempDir()
	dbPath = filepath.Join(dir, "nas.sqlite")
	conn, err := db.Open(dbPath)
	require.NoError(t, err)
	defer conn.Close()
	require.NoError(t, db.InitSchema(conn))
	for _, googleID := range []string{"google-1", "google-2"} {
		id, err := db.SaveOAuthUser(conn, googleID, "name of "+googleID, "", "")
		require.NoError(t, err)
		require.NoError(t, db.AddIdentity(conn, id, "google", googleID))
	}
	mapPath = filepath.Join(dir, "map.csv")
	require.NoError(t, os.WriteFile(mapPath, []byte(csvText), 0o600))
	return dbPath, mapPath
}

func linkedSubjects(t *testing.T, dbPath string) []string {
	t.Helper()
	conn, err := db.Open(dbPath)
	require.NoError(t, err)
	defer conn.Close()
	rows, err := conn.Query("SELECT external_id FROM user_identities WHERE provider = 'oidc' ORDER BY external_id")
	require.NoError(t, err)
	defer rows.Close()
	subjects := []string{}
	for rows.Next() {
		var subject string
		require.NoError(t, rows.Scan(&subject))
		subjects = append(subjects, subject)
	}
	return subjects
}

func TestOIDCLink_DryRunThenRealRunThenRepeat(t *testing.T) {
	dbPath, mapPath := linkFixture(t, "# old google id, new sub\ngoogle-1,sub-1\ngoogle-2, sub-2\ngoogle-unknown,sub-3\n")

	require.Equal(t, 0, runOIDCLink([]string{"--db", dbPath, "--map", mapPath, "--dry-run"}))
	require.Empty(t, linkedSubjects(t, dbPath))

	require.Equal(t, 0, runOIDCLink([]string{"--db", dbPath, "--map", mapPath}))
	require.Equal(t, []string{"sub-1", "sub-2"}, linkedSubjects(t, dbPath))

	require.Equal(t, 0, runOIDCLink([]string{"--db", dbPath, "--map", mapPath}))
	require.Equal(t, []string{"sub-1", "sub-2"}, linkedSubjects(t, dbPath))
}

func TestOIDCLink_IgnoresBOMAndSpaceAroundValues(t *testing.T) {
	for name, csvText := range map[string]string{
		"BOM before a comment": "\xEF\xBB\xBF# old google id, new sub\r\ngoogle-1 , sub-1 \r\n\tgoogle-2\t,\tsub-2\t\r\n",
		"BOM before a value":   "\xEF\xBB\xBFgoogle-1,sub-1\ngoogle-2,sub-2 \n",
	} {
		t.Run(name, func(t *testing.T) {
			dbPath, mapPath := linkFixture(t, csvText)
			require.Equal(t, 0, runOIDCLink([]string{"--db", dbPath, "--map", mapPath}))
			require.Equal(t, []string{"sub-1", "sub-2"}, linkedSubjects(t, dbPath))
		})
	}
}

func TestOIDCLink_ConflictGivesExitCodeOne(t *testing.T) {
	dbPath, mapPath := linkFixture(t, "google-1,sub-1\ngoogle-2,sub-1\n")
	require.Equal(t, 1, runOIDCLink([]string{"--db", dbPath, "--map", mapPath}))
	require.Equal(t, []string{"sub-1"}, linkedSubjects(t, dbPath))
}

func TestOIDCLink_ReadsDatabasePathFromEnvironment(t *testing.T) {
	dbPath, mapPath := linkFixture(t, "google-1,sub-1\n")
	t.Setenv("DB_PATH", dbPath)
	require.Equal(t, 0, runOIDCLink([]string{"--map", mapPath}))
	require.Equal(t, []string{"sub-1"}, linkedSubjects(t, dbPath))
}

func TestOIDCLink_RefusesBadInput(t *testing.T) {
	dbPath, mapPath := linkFixture(t, "google-1,sub-1\n")
	missing := filepath.Join(t.TempDir(), "missing.sqlite")
	t.Setenv("DB_PATH", "")

	require.Equal(t, 2, runOIDCLink([]string{"--map", mapPath}))
	require.Equal(t, 2, runOIDCLink([]string{"--db", dbPath}))
	require.Equal(t, 2, runOIDCLink([]string{"--db", dbPath, "--map", mapPath, "--from", "kakao"}))
	require.Equal(t, 1, runOIDCLink([]string{"--db", missing, "--map", mapPath}))
	require.NoFileExists(t, missing)

	_, malformed := linkFixture(t, "google-1\n")
	require.Equal(t, 1, runOIDCLink([]string{"--db", dbPath, "--map", malformed}))
	_, blank := linkFixture(t, "google-1,\n")
	require.Equal(t, 1, runOIDCLink([]string{"--db", dbPath, "--map", blank}))
	_, spaces := linkFixture(t, "google-1,   \n")
	require.Equal(t, 1, runOIDCLink([]string{"--db", dbPath, "--map", spaces}))
	require.Empty(t, linkedSubjects(t, dbPath))
}
