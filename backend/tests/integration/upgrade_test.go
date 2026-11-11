package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/server"
)

// v0.6.0의 db.InitSchema가 만든 스키마 그대로임
const schemaV060 = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	userId TEXT UNIQUE NOT NULL,
	username TEXT NOT NULL,
	global_name TEXT,
	krname TEXT,
	password TEXT,
	auth_type TEXT
);
CREATE TABLE IF NOT EXISTS user_intents (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	intent TEXT NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS log (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	activity TEXT NOT NULL,
	description TEXT,
	user_id INTEGER NOT NULL,
	time INTEGER NOT NULL,
	loc TEXT,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);
CREATE TABLE IF NOT EXISTS server_settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS user_identities (
	user_id INTEGER NOT NULL,
	provider TEXT NOT NULL,
	external_id TEXT NOT NULL,
	PRIMARY KEY (provider, external_id),
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_identities_user ON user_identities(user_id);
CREATE TABLE IF NOT EXISTS share_links (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	token_hash TEXT UNIQUE NOT NULL,
	user_id INTEGER NOT NULL,
	loc TEXT NOT NULL,
	name TEXT NOT NULL,
	is_folder INTEGER NOT NULL DEFAULT 0,
	created_at INTEGER NOT NULL,
	expires_at INTEGER NOT NULL,
	revoked_at INTEGER,
	download_count INTEGER NOT NULL DEFAULT 0,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_share_links_user ON share_links(user_id);
`

// v0.6.0이 쓰던 데이터베이스를 만든 뒤, 새 서버가 시작할 때 하는 순서대로 올림
func upgradedV060Server(t *testing.T, authType config.AuthType) (http.Handler, *sql.DB) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "nas.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	_, err = conn.Exec(schemaV060)
	require.NoError(t, err)

	hash, err := auth.HashPassword("old-password")
	require.NoError(t, err)
	_, err = conn.Exec(`
		INSERT INTO users (userId, username, global_name, krname, password, auth_type)
		VALUES ('alice', 'alice', 'Alice', '앨리스', ?, 'local'), ('google-555', 'Bob', 'Bob', '밥', NULL, NULL)`, hash)
	require.NoError(t, err)
	_, err = conn.Exec(`
		INSERT INTO user_identities (user_id, provider, external_id) VALUES (1, 'local', 'alice'), (2, 'google', 'google-555');
		INSERT INTO user_intents (user_id, intent) VALUES (1, 'ADMIN'), (2, 'VIEW'), (2, 'DOWNLOAD');`)
	require.NoError(t, err)

	require.NoError(t, db.InitSchema(conn))
	require.NoError(t, db.VerifySchema(conn))
	require.NoError(t, db.Backfill(conn))

	cfg := &config.Config{
		CorsOrigin:           "*",
		PrivateKey:           testPrivateKey,
		AdminPassword:        "admin-pass",
		AuthType:             authType,
		NASDataDir:           t.TempDir(),
		PasswordRequirements: config.PasswordRequirements{MinLength: 4},
	}
	return server.NewRouter(cfg, conn), conn
}

// v0.6.0의 IssueToken이 서명하던 내용임. issued_at이 없음
func tokenFromV060(t *testing.T, userID string) string {
	t.Helper()
	claims := jwt.MapClaims{"userId": userID, "expires_in": time.Now().Add(3 * 24 * time.Hour).UnixMilli()}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testPrivateKey))
	require.NoError(t, err)
	return signed
}

func postJSONBody(router http.Handler, path string, body any) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestUpgradeFromV060_TokensIssuedBeforeTheUpgradeStayValid(t *testing.T) {
	for _, authType := range []config.AuthType{config.AuthTypeBoth, config.AuthTypeLocal, config.AuthTypeOAuth} {
		router, _ := upgradedV060Server(t, authType)
		alice, bob := tokenFromV060(t, "alice"), tokenFromV060(t, "google-555")

		require.Equal(t, http.StatusOK, getAs(router, "/server/getAllUsers", alice).Code, authType)
		require.Equal(t, http.StatusOK, getAs(router, "/server/readFolder?loc=", bob).Code, authType)
		require.Equal(t, http.StatusForbidden, getAs(router, "/server/getAllUsers", bob).Code, authType)

		intents := getAs(router, "/server/getIntents?userId=google-555", bob)
		require.Equal(t, http.StatusOK, intents.Code, authType)
		require.JSONEq(t, `{"intents":["VIEW","DOWNLOAD"]}`, intents.Body.String())

		logged := postJSONBody(router, "/server/log", map[string]any{"activity": "OPEN", "description": "x", "token": bob, "time": 1, "loc": "/"})
		require.Equal(t, http.StatusOK, logged.Code, authType)
	}
}

func TestUpgradeFromV060_ExistingLocalAccountSignsIn(t *testing.T) {
	router, conn := upgradedV060Server(t, config.AuthTypeBoth)

	w := postJSONBody(router, "/server/auth/login", map[string]string{"userId": "alice", "password": "old-password"})
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var login struct {
		Success bool   `json:"success"`
		Token   string `json:"token"`
		User    struct {
			UserID  string   `json:"userId"`
			KrName  string   `json:"krname"`
			Intents []string `json:"intents"`
		} `json:"user"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &login))
	require.True(t, login.Success)
	require.Equal(t, "alice", login.User.UserID)
	require.Equal(t, "앨리스", login.User.KrName)
	require.Equal(t, []string{"ADMIN"}, login.User.Intents)
	require.Equal(t, http.StatusOK, getAs(router, "/server/getAllUsers", login.Token).Code)

	wrong := postJSONBody(router, "/server/auth/login", map[string]string{"userId": "alice", "password": "wrong"})
	require.Equal(t, http.StatusUnauthorized, wrong.Code)

	changed := postJSONBody(router, "/server/auth/change-password?token="+tokenFromV060(t, "alice"),
		map[string]string{"oldPassword": "old-password", "newPassword": "new-password"})
	require.Equal(t, http.StatusOK, changed.Code, changed.Body.String())

	registered := postJSONBody(router, "/server/auth/register", map[string]string{"userId": "carol", "username": "carol", "password": "strong-pass"})
	require.Equal(t, http.StatusOK, registered.Code, registered.Body.String())

	var restricted int
	require.NoError(t, conn.QueryRow("SELECT COUNT(*) FROM users WHERE tokens_valid_after IS NOT NULL").Scan(&restricted))
	require.Zero(t, restricted, "nothing sets a restriction outside oidc mode")
}

// Google의 토큰 주소와 사용자 정보 주소를 대신함. 업그레이드 전에 가입한 google-555로 답함
type googleOfExistingUser struct{}

func (googleOfExistingUser) RoundTrip(req *http.Request) (*http.Response, error) {
	body := `{}`
	switch req.URL.Host {
	case "oauth2.googleapis.com":
		body = `{"access_token":"google-access"}`
	case "www.googleapis.com":
		body = `{"id":"google-555","email":"bob@example.com","name":"Bob"}`
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}, Request: req}, nil
}

func TestUpgradeFromV060_ExistingGoogleAccountSignsIn(t *testing.T) {
	previous := http.DefaultClient.Transport
	http.DefaultClient.Transport = googleOfExistingUser{}
	t.Cleanup(func() { http.DefaultClient.Transport = previous })

	router, conn := upgradedV060Server(t, config.AuthTypeBoth)
	_, err := conn.Exec(`INSERT INTO server_settings (key, value) VALUES
		('oauth.google.client_id', 'id'), ('oauth.google.client_secret', 'secret'), ('oauth.google.redirect_uri', 'https://nas.example/googleLogin')`)
	require.NoError(t, err)

	w := getAs(router, "/server/auth/google/callback?code=from-google", "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var login map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &login))
	require.Equal(t, "google-555", login["userId"])
	require.NotContains(t, login, "status", "an existing account is not sent to registration")
	token, _ := login["token"].(string)
	require.NotEmpty(t, token)

	intents := getAs(router, "/server/getIntents?userId=google-555", token)
	require.Equal(t, http.StatusOK, intents.Code)
	require.JSONEq(t, `{"intents":["VIEW","DOWNLOAD"]}`, intents.Body.String())
	identities := getAs(router, "/server/auth/identities", token)
	require.JSONEq(t, `{"identities":[{"provider":"google","externalId":"google-555"}]}`, identities.Body.String())
}

func TestLegacyModes_AuthConfigHasTheV060Fields(t *testing.T) {
	for authType, local := range map[config.AuthType]bool{config.AuthTypeBoth: true, config.AuthTypeLocal: true, config.AuthTypeOAuth: false} {
		router, _ := upgradedV060Server(t, authType)
		w := getAs(router, "/server/auth/config", "")
		require.Equal(t, http.StatusOK, w.Code)

		var body map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		keys := []string{}
		for key := range body {
			keys = append(keys, key)
		}
		require.ElementsMatch(t, []string{
			"authType", "localAuthEnabled", "discordEnabled", "discordLoginUrl",
			"googleEnabled", "googleLoginUrl", "oauthEnabled", "passwordRequirements",
		}, keys, authType)
		require.JSONEq(t, `"`+string(authType)+`"`, string(body["authType"]))
		require.Equal(t, local, string(body["localAuthEnabled"]) == "true", authType)
		require.NotContains(t, w.Body.String(), "oidc")
	}
}

func TestLegacyModes_KeepTheAdminPasswordRoute(t *testing.T) {
	for _, authType := range []config.AuthType{config.AuthTypeBoth, config.AuthTypeLocal, config.AuthTypeOAuth} {
		router, conn := upgradedV060Server(t, authType)
		token := tokenFromV060(t, "google-555")

		wrong := postJSONBody(router, "/requestAdminIntent?token="+token, map[string]string{"pwd": "not-the-password"})
		require.Equal(t, http.StatusInternalServerError, wrong.Code, authType)
		right := postJSONBody(router, "/server/requestAdminIntent?token="+token, map[string]string{"pwd": "admin-pass"})
		require.Equal(t, http.StatusOK, right.Code, authType)
		require.Equal(t, "complete", right.Body.String(), authType)

		user, err := db.GetUser(conn, "google-555")
		require.NoError(t, err)
		require.Contains(t, user.Intents, "ADMIN", authType)
	}
}

func TestLegacyModes_HaveNoOIDCRoutes(t *testing.T) {
	for _, authType := range []config.AuthType{config.AuthTypeBoth, config.AuthTypeLocal, config.AuthTypeOAuth} {
		router, _ := upgradedV060Server(t, authType)
		for _, route := range [][2]string{
			{"GET", "/auth/oidc/login"},
			{"GET", "/auth/oidc/callback?code=x&state=y"},
			{"POST", "/auth/oidc/session"},
			{"GET", "/auth/oidc/logout"},
			{"POST", "/auth/oidc/backchannel"},
		} {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(route[0], route[1], nil))
			require.Equal(t, http.StatusNotFound, w.Code, string(authType)+" "+route[1])
			require.Empty(t, w.Header().Values("Set-Cookie"))
		}
	}
}
