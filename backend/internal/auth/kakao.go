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
	kakaoTokenURL = "https://kauth.kakao.com/oauth/token"
	kakaoUserURL  = "https://kapi.kakao.com/v2/user/me"
)

type kakaoTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type kakaoUserResponse struct {
	ID         int64 `json:"id"`
	Properties struct {
		Nickname string `json:"nickname"`
	} `json:"properties"`
}

func exchangeKakaoCode(client *http.Client, restAPIKey, clientSecret, redirectURI, code string) (string, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {restAPIKey},
		"redirect_uri":  {redirectURI},
		"client_secret": {clientSecret},
		"code":          {code},
	}
	req, err := http.NewRequest("POST", kakaoTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("kakao token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("kakao token status %d: %s", resp.StatusCode, string(body))
	}
	var tok kakaoTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", fmt.Errorf("decode token: %w", err)
	}
	return tok.AccessToken, nil
}

func fetchKakaoUser(client *http.Client, accessToken string) (*kakaoUserResponse, error) {
	req, err := http.NewRequest("GET", kakaoUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kakao user: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("kakao user status %d: %s", resp.StatusCode, string(body))
	}
	user := &kakaoUserResponse{}
	if err := json.NewDecoder(resp.Body).Decode(user); err != nil {
		return nil, fmt.Errorf("decode kakao user: %w", err)
	}
	return user, nil
}

func (h *Handlers) KakaoLogin(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code required", http.StatusBadRequest)
		return
	}
	if h.Config.KakaoRestAPIKey == "" || h.Config.KakaoRedirectURI == "" || h.Config.KakaoClientSecret == "" {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":   "OAuth configuration error",
			"message": "Kakao OAuth credentials not configured",
		})
		return
	}
	accessToken, err := exchangeKakaoCode(http.DefaultClient, h.Config.KakaoRestAPIKey,
		h.Config.KakaoClientSecret, h.Config.KakaoRedirectURI, code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":   "OAuth process failed",
			"message": "An error occurred during Kakao OAuth authentication",
		})
		return
	}
	user, err := fetchKakaoUser(http.DefaultClient, accessToken)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":   "OAuth process failed",
			"message": "An error occurred during Kakao OAuth authentication",
		})
		return
	}
	userIDStr := fmt.Sprintf("%d", user.ID)
	existing, err := db.GetUser(h.DB, userIDStr)
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
			"userId":   existing.UserID,
			"nickname": user.Properties.Nickname,
			"token":    token,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "new",
		"userId":   userIDStr,
		"nickname": user.Properties.Nickname,
	})
}

type kakaoRegisterRequest struct {
	UserID   string `json:"userId"`
	Nickname string `json:"nickname"`
	KrName   string `json:"krname"`
}

func (h *Handlers) KakaoRegister(w http.ResponseWriter, r *http.Request) {
	var req kakaoRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if _, err := db.SaveOAuthUser(h.DB, req.UserID, req.Nickname, req.Nickname, req.KrName); err != nil {
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
		"username":    req.Nickname,
		"global_name": req.Nickname,
		"token":       token,
	})
}
