package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/db"
)

func TestExtractToken_QueryParamPriority(t *testing.T) {
	req := httptest.NewRequest("GET", "/?token=fromquery", nil)
	req.Header.Set("Authorization", "Bearer fromheader")
	require.Equal(t, "fromquery", ExtractToken(req))
}

func TestExtractToken_BearerFallback(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer mytoken")
	require.Equal(t, "mytoken", ExtractToken(req))
}

func TestExtractToken_NoneReturnsEmpty(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	require.Equal(t, "", ExtractToken(req))
}

func TestRequireToken_RejectsMissing(t *testing.T) {
	handler := RequireToken("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireToken_AttachesClaims(t *testing.T) {
	token, err := IssueToken("alice", "secret")
	require.NoError(t, err)

	var got string
	handler := RequireToken("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = ClaimsFromContext(r.Context()).UserID
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/?token="+token, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "alice", got)
}

func TestRequireIntent_AllowsWhenGranted(t *testing.T) {
	conn := setupTestDBForAuth(t)
	_, err := db.SaveLocalUser(conn, "alice", "alice", "h", "alice")
	require.NoError(t, err)
	require.NoError(t, db.ToggleIntent(conn, "alice", "UPLOAD"))

	token, err := IssueToken("alice", "secret")
	require.NoError(t, err)

	chain := RequireToken("secret")(RequireIntent(conn, IntentUpload)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	))
	req := httptest.NewRequest("GET", "/?token="+token, nil)
	w := httptest.NewRecorder()
	chain.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestRequireIntent_BlocksWhenMissing(t *testing.T) {
	conn := setupTestDBForAuth(t)
	_, err := db.SaveLocalUser(conn, "bob", "bob", "h", "bob")
	require.NoError(t, err)

	token, err := IssueToken("bob", "secret")
	require.NoError(t, err)

	chain := RequireToken("secret")(RequireIntent(conn, IntentDelete)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler must not be called")
		}),
	))
	req := httptest.NewRequest("GET", "/?token="+token, nil)
	w := httptest.NewRecorder()
	chain.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}
