package auth

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
)

var (
	bothConfig = &config.Config{PrivateKey: testSecret, AuthType: config.AuthTypeBoth}
	oidcConfig = &config.Config{PrivateKey: testSecret, AuthType: config.AuthTypeOIDC}
)

// 실제 스키마(db.InitSchema)로 만든 데이터베이스
func newSchemaDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	require.NoError(t, db.InitSchema(conn))
	return conn
}

// v0.6.0이 내준 토큰과 같은 모양임. issued_at이 없음
func tokenWithoutIssuedAt(t *testing.T, userID string) string {
	t.Helper()
	claims := jwt.MapClaims{"userId": userID, "expires_in": time.Now().UnixMilli() + tokenLifetimeMs}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	require.NoError(t, err)
	return signed
}

func TestIssueToken_CarriesIssuedAt(t *testing.T) {
	before := time.Now().UnixMilli()
	token, err := IssueToken("alice", testSecret)
	require.NoError(t, err)

	claims, err := ParseToken(token, testSecret)
	require.NoError(t, err)
	require.GreaterOrEqual(t, claims.IssuedAtMs, before)
	require.LessOrEqual(t, claims.IssuedAtMs, time.Now().UnixMilli())
}

func TestAuthenticate_AcceptsTokenWithoutIssuedAtWhileNoRestriction(t *testing.T) {
	conn := newSchemaDB(t)
	_, err := db.SaveLocalUser(conn, "alice", "alice", "hash", "")
	require.NoError(t, err)

	for _, cfg := range []*config.Config{bothConfig, oidcConfig} {
		claims, err := Authenticate(cfg, conn, tokenWithoutIssuedAt(t, "alice"))
		require.NoError(t, err, cfg.AuthType)
		require.Equal(t, "alice", claims.UserID)
		require.Zero(t, claims.IssuedAtMs)
	}
}

func TestAuthenticate_RejectsTokensIssuedBeforeTokensValidAfter(t *testing.T) {
	conn := newSchemaDB(t)
	id, err := db.SaveLocalUser(conn, "bob", "bob", "hash", "")
	require.NoError(t, err)
	issued, err := IssueToken("bob", testSecret)
	require.NoError(t, err)
	claims, err := ParseToken(issued, testSecret)
	require.NoError(t, err)

	require.NoError(t, db.SetTokensValidAfter(conn, id, claims.IssuedAtMs+1))
	for _, token := range []string{issued, tokenWithoutIssuedAt(t, "bob")} {
		_, err := Authenticate(bothConfig, conn, token)
		require.ErrorIs(t, err, ErrInvalidToken)
	}

	require.NoError(t, db.SetTokensValidAfter(conn, id, claims.IssuedAtMs))
	_, err = Authenticate(bothConfig, conn, issued)
	require.NoError(t, err, "a token issued at the same millisecond or later stays valid")
}

func TestAuthenticate_UserWithoutRow(t *testing.T) {
	conn := newSchemaDB(t)
	token, err := IssueToken("ghost", testSecret)
	require.NoError(t, err)

	claims, err := Authenticate(bothConfig, conn, token)
	require.NoError(t, err)
	require.Equal(t, "ghost", claims.UserID)

	_, err = Authenticate(oidcConfig, conn, token)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestAuthenticate_RejectsBadSignatureAsInvalid(t *testing.T) {
	conn := newSchemaDB(t)
	token, err := IssueToken("alice", "another-secret")
	require.NoError(t, err)
	_, err = Authenticate(bothConfig, conn, token)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestRequireToken_RejectsTokenIssuedBeforeTokensValidAfter(t *testing.T) {
	conn := newSchemaDB(t)
	id, err := db.SaveLocalUser(conn, "carol", "carol", "hash", "")
	require.NoError(t, err)
	old, err := IssueToken("carol", testSecret)
	require.NoError(t, err)
	oldClaims, err := ParseToken(old, testSecret)
	require.NoError(t, err)
	require.NoError(t, db.SetTokensValidAfter(conn, id, oldClaims.IssuedAtMs+1))

	handler := RequireToken(bothConfig, conn)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rejected := httptest.NewRecorder()
	handler.ServeHTTP(rejected, httptest.NewRequest("GET", "/?token="+old, nil))
	require.Equal(t, http.StatusUnauthorized, rejected.Code)

	require.NoError(t, db.SetTokensValidAfter(conn, id, oldClaims.IssuedAtMs))
	accepted := httptest.NewRecorder()
	handler.ServeHTTP(accepted, httptest.NewRequest("GET", "/?token="+old, nil))
	require.Equal(t, http.StatusOK, accepted.Code)
}

func TestRequireToken_ReportsFailedLookupAsServerError(t *testing.T) {
	conn := newSchemaDB(t)
	token, err := IssueToken("dave", testSecret)
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	handler := RequireToken(bothConfig, conn)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler must not be called")
	}))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/?token="+token, nil))
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
