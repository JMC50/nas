package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const verifiedGoogleToken = "token-from-google"

// fakeGoogle answers the Google userinfo call made through http.DefaultClient.
type fakeGoogle struct{}

func (fakeGoogle) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != "www.googleapis.com" || req.Header.Get("Authorization") != "Bearer "+verifiedGoogleToken {
		return &http.Response{StatusCode: http.StatusUnauthorized, Body: io.NopCloser(strings.NewReader(`{}`)), Header: http.Header{}, Request: req}, nil
	}
	body := `{"id":"google-777","email":"real@example.com","name":"Real Person"}`
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}, Request: req}, nil
}

func useFakeGoogle(t *testing.T) {
	t.Helper()
	previous := http.DefaultClient.Transport
	http.DefaultClient.Transport = fakeGoogle{}
	t.Cleanup(func() { http.DefaultClient.Transport = previous })
}

func postGoogleRegister(router http.Handler, body map[string]string) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/server/auth/google/register", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestGoogleRegister_RejectsUnverifiedToken(t *testing.T) {
	useFakeGoogle(t)
	router, cfg := setupAuthTestServer(t)
	victimToken := registerAndLogin(t, router, "victim")
	grantAdmin(t, router, victimToken, cfg.AdminPassword)

	for _, body := range []map[string]string{
		{"userId": "victim", "access_token": "forged", "krname": "x"},
		{"userId": "victim", "krname": "x"},
	} {
		w := postGoogleRegister(router, body)
		require.NotEqual(t, http.StatusOK, w.Code, w.Body.String())
		require.NotContains(t, w.Body.String(), "token")
	}

	w := getAs(router, "/server/getAllUsers", victimToken)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"userId":"victim"`)
	require.NotContains(t, w.Body.String(), "google-777")
}

func TestGoogleRegister_UsesGoogleIdentityNotBody(t *testing.T) {
	useFakeGoogle(t)
	router, cfg := setupAuthTestServer(t)
	adminToken := registerAndLogin(t, router, "victim")
	grantAdmin(t, router, adminToken, cfg.AdminPassword)

	w := postGoogleRegister(router, map[string]string{"userId": "victim", "access_token": verifiedGoogleToken, "krname": "홍길동"})
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "google-777", resp["userId"])
	newToken, _ := resp["token"].(string)
	require.NotEmpty(t, newToken)

	// The token must belong to the Google user, which has no ADMIN intent.
	require.Equal(t, http.StatusForbidden, getAs(router, "/server/getAllUsers", newToken).Code)
	require.Equal(t, http.StatusOK, getAs(router, "/server/getAllUsers", adminToken).Code)
}
