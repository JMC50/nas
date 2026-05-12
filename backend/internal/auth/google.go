package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/JMC50/nas/internal/db"
)

const (
	googleTokenURL = "https://oauth2.googleapis.com/token"
	googleUserURL  = "https://www.googleapis.com/oauth2/v2/userinfo"
)

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type googleUserResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func exchangeGoogleCode(client *http.Client, clientID, clientSecret, redirectURI, code string) (string, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"code":          {code},
	}
	req, err := http.NewRequest("POST", googleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("google token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("google token status %d: %s", resp.StatusCode, string(body))
	}
	var tok googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", fmt.Errorf("decode token: %w", err)
	}
	return tok.AccessToken, nil
}

func fetchGoogleUser(client *http.Client, accessToken string) (*googleUserResponse, error) {
	req, err := http.NewRequest("GET", googleUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google user: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google user status %d: %s", resp.StatusCode, string(body))
	}
	user := &googleUserResponse{}
	if err := json.NewDecoder(resp.Body).Decode(user); err != nil {
		return nil, fmt.Errorf("decode google user: %w", err)
	}
	return user, nil
}

func (h *Handlers) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code required", http.StatusBadRequest)
		return
	}
	creds := ResolveCreds(h.Config, h.DB)
	if !googleOK(creds) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":   "OAuth configuration error",
			"message": "Google OAuth credentials not configured",
		})
		return
	}
	accessToken, err := exchangeGoogleCode(http.DefaultClient, creds.GoogleClientID,
		creds.GoogleClientSecret, creds.GoogleRedirectURI, code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":   "OAuth process failed",
			"message": "An error occurred during Google OAuth authentication",
		})
		return
	}
	user, err := fetchGoogleUser(http.DefaultClient, accessToken)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":   "OAuth process failed",
			"message": "An error occurred during Google OAuth authentication",
		})
		return
	}
	existing, err := db.GetUser(h.DB, user.ID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if existing != nil {
		token, err := IssueToken(existing.UserID, h.Config.PrivateKey)
		if err != nil {
			http.Error(w, "token issue failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"userId":      existing.UserID,
			"username":    user.Name,
			"email":       user.Email,
			"global_name": user.Name,
			"token":       token,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "new",
		"userId":      user.ID,
		"username":    user.Name,
		"email":       user.Email,
		"global_name": user.Name,
	})
}

type googleRegisterRequest struct {
	UserID     string `json:"userId"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	GlobalName string `json:"global_name"`
	KrName     string `json:"krname"`
}

func (h *Handlers) GoogleRegister(w http.ResponseWriter, r *http.Request) {
	var req googleRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	globalName := req.GlobalName
	if globalName == "" {
		globalName = req.Username
	}
	if _, err := db.SaveOAuthUser(h.DB, req.UserID, req.Username, globalName, req.KrName); err != nil {
		http.Error(w, "save user failed", http.StatusInternalServerError)
		return
	}
	token, err := IssueToken(req.UserID, h.Config.PrivateKey)
	if err != nil {
		http.Error(w, "token issue failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "complete",
		"userId":      req.UserID,
		"username":    req.Username,
		"email":       req.Email,
		"global_name": globalName,
		"token":       token,
	})
}
