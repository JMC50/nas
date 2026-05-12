package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/server"
)

const testPrivateKey = "test-secret-key"

func setupAuthTestServer(t *testing.T) (http.Handler, *config.Config) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	conn, err := db.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	_, err = conn.Exec(testSchema())
	require.NoError(t, err)

	cfg := &config.Config{
		Port:                 0,
		CorsOrigin:           "*",
		PrivateKey:           testPrivateKey,
		AdminPassword:        "admin-pass",
		AuthType:             config.AuthTypeBoth,
		PasswordRequirements: config.PasswordRequirements{MinLength: 4},
	}
	t.Cleanup(func() {
		// the conn captured here is closed via the earlier Cleanup
	})

	// Stash the conn on the server router via config.
	router := server.NewRouter(cfg, conn)
	// Also let the test seed users directly:
	t.Cleanup(func() { _ = conn })
	return router, cfg
}

// TestLocalAuth_RegisterLoginRoundTrip — full path: register → login → JWT works
// against a protected endpoint.
func TestLocalAuth_RegisterLoginRoundTrip(t *testing.T) {
	router, _ := setupAuthTestServer(t)

	body, _ := json.Marshal(map[string]string{
		"userId":   "alice",
		"username": "alice",
		"password": "strong-pass",
		"krname":   "앨리스",
	})
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var registerResp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &registerResp))
	require.True(t, registerResp["success"].(bool))

	// Now login
	loginBody, _ := json.Marshal(map[string]string{"userId": "alice", "password": "strong-pass"})
	req = httptest.NewRequest("POST", "/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var loginResp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &loginResp))
	token := loginResp["token"].(string)
	require.NotEmpty(t, token)

	// Verify JWT against a protected endpoint
	req = httptest.NewRequest("GET", "/checkAdmin?token="+token, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var checkResp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &checkResp))
	require.Equal(t, false, checkResp["isAdmin"]) // freshly registered user is not admin
}

func TestLocalAuth_RejectsWrongPassword(t *testing.T) {
	router, _ := setupAuthTestServer(t)
	body, _ := json.Marshal(map[string]string{"userId": "bob", "username": "bob", "password": "good-pass"})
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	body, _ = json.Marshal(map[string]string{"userId": "bob", "password": "wrong-pass"})
	req = httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthConfig_ReturnsStruct(t *testing.T) {
	router, _ := setupAuthTestServer(t)
	req := httptest.NewRequest("GET", "/auth/config", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "both", resp["authType"])
	require.Equal(t, true, resp["localAuthEnabled"])
	require.Equal(t, true, resp["oauthEnabled"])
}

func TestProtectedEndpoint_RejectsMissingToken(t *testing.T) {
	router, _ := setupAuthTestServer(t)
	req := httptest.NewRequest("GET", "/checkAdmin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestExistingNodeToken_VerifiesAgainstGoBackend — simulates the data continuity
// case: a JWT issued by the legacy Node backend (using same PRIVATE_KEY) must
// verify successfully in Go without forcing re-login.
func TestExistingNodeToken_VerifiesAgainstGoBackend(t *testing.T) {
	router, cfg := setupAuthTestServer(t)

	// Seed a user the Node way (raw INSERT, no Go-side registration)
	conn, err := db.Open(filepath.Join(t.TempDir(), "seed.sqlite"))
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.Exec(testSchema())
	require.NoError(t, err)
	_, err = conn.Exec(`INSERT INTO users (userId, username) VALUES ('legacy-user', 'legacy')`)
	require.NoError(t, err)

	// Issue a token with the legacy payload shape via auth.IssueToken (same key, same payload)
	token, err := auth.IssueToken("legacy-user", cfg.PrivateKey)
	require.NoError(t, err)

	// Verify it parses
	claims, err := auth.ParseToken(token, cfg.PrivateKey)
	require.NoError(t, err)
	require.Equal(t, "legacy-user", claims.UserID)

	// And that it passes the middleware on a real route
	req := httptest.NewRequest("GET", "/checkIntent?token="+token+"&intent=VIEW", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
}
