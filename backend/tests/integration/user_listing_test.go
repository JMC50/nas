package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func registerAndLogin(t *testing.T, router http.Handler, userID string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"userId": userID, "username": userID, "password": "strong-pass"})
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	token, _ := resp["token"].(string)
	require.NotEmpty(t, token)
	return token
}

func grantAdmin(t *testing.T, router http.Handler, token, adminPassword string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"pwd": adminPassword})
	req := httptest.NewRequest("POST", "/requestAdminIntent", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
}

func getAs(router http.Handler, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestGetAllUsers_RejectsAnonymous(t *testing.T) {
	router, _ := setupAuthTestServer(t)
	registerAndLogin(t, router, "alice")

	for _, path := range []string{"/getAllUsers", "/server/getAllUsers"} {
		w := getAs(router, path, "")
		require.Equal(t, http.StatusUnauthorized, w.Code, path)
		require.NotContains(t, w.Body.String(), "alice", path)
	}
}

func TestGetAllUsers_RejectsNonAdmin(t *testing.T) {
	router, _ := setupAuthTestServer(t)
	token := registerAndLogin(t, router, "bob")

	w := getAs(router, "/server/getAllUsers", token)
	require.Equal(t, http.StatusForbidden, w.Code)
	require.NotContains(t, w.Body.String(), "bob")
}

func TestGetAllUsers_AdminGetsList(t *testing.T) {
	router, cfg := setupAuthTestServer(t)
	adminToken := registerAndLogin(t, router, "carol")
	registerAndLogin(t, router, "dave")
	grantAdmin(t, router, adminToken, cfg.AdminPassword)

	w := getAs(router, "/server/getAllUsers", adminToken)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp struct {
		Users []struct {
			UserID  string   `json:"userId"`
			Intents []string `json:"intents"`
		} `json:"users"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	ids := []string{}
	for _, user := range resp.Users {
		ids = append(ids, user.UserID)
	}
	require.ElementsMatch(t, []string{"carol", "dave"}, ids)
}

func TestGetIntents_RequiresTokenAndOwnUser(t *testing.T) {
	router, cfg := setupAuthTestServer(t)
	erinToken := registerAndLogin(t, router, "erin")
	frankToken := registerAndLogin(t, router, "frank")

	require.Equal(t, http.StatusUnauthorized, getAs(router, "/server/getIntents?userId=erin", "").Code)
	require.Equal(t, http.StatusOK, getAs(router, "/server/getIntents?userId=erin", erinToken).Code)
	require.Equal(t, http.StatusForbidden, getAs(router, "/server/getIntents?userId=erin", frankToken).Code)

	grantAdmin(t, router, frankToken, cfg.AdminPassword)
	require.Equal(t, http.StatusOK, getAs(router, "/server/getIntents?userId=erin", frankToken).Code)
}

func TestGetActivityLog_RequiresToken(t *testing.T) {
	router, _ := setupAuthTestServer(t)
	token := registerAndLogin(t, router, "grace")

	for _, path := range []string{"/getActivityLog", "/server/getActivityLog"} {
		require.Equal(t, http.StatusUnauthorized, getAs(router, path, "").Code, path)
	}
	require.Equal(t, http.StatusOK, getAs(router, "/server/getActivityLog", token).Code)
}
