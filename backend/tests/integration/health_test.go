package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/server"
)

func TestHealthz_ReturnsOK(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	conn, err := db.Open(dbPath)
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.Exec(testSchema())
	require.NoError(t, err)

	cfg := &config.Config{Port: 0, CorsOrigin: "*"}
	r := server.NewRouter(cfg, conn)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Equal(t, "ok", body["status"])
	require.Equal(t, "connected", body["db"])
	require.Equal(t, "valid", body["schema"])
}

func testSchema() string {
	return `
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
}
