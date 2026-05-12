package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/JMC50/nas/internal/db"
)

const discordUserURL = "https://discord.com/api/users/@me"

type DiscordUser struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name"`
}

func fetchDiscordUser(client *http.Client, accessToken string) (*DiscordUser, error) {
	req, err := http.NewRequest("GET", discordUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discord api: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord api status %d: %s", resp.StatusCode, string(body))
	}
	user := &DiscordUser{}
	if err := json.NewDecoder(resp.Body).Decode(user); err != nil {
		return nil, fmt.Errorf("decode discord user: %w", err)
	}
	return user, nil
}

// DiscordLogin handles GET /login?access_token=… (Discord OAuth callback).
// If user exists, returns token + user. If new, returns {status: "new", userId, ...}.
func (h *Handlers) DiscordLogin(w http.ResponseWriter, r *http.Request) {
	accessToken := r.URL.Query().Get("access_token")
	if accessToken == "" {
		http.Error(w, "access_token required", http.StatusBadRequest)
		return
	}
	user, err := fetchDiscordUser(http.DefaultClient, accessToken)
	if err != nil {
		http.Error(w, "discord lookup failed", http.StatusBadGateway)
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
			"id":          existing.ID,
			"userId":      existing.UserID,
			"username":    existing.Username,
			"global_name": existing.GlobalName,
			"krname":      existing.KrName,
			"intents":     existing.Intents,
			"token":       token,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "new",
		"userId":      user.ID,
		"username":    user.Username,
		"global_name": user.GlobalName,
	})
}

type discordRegisterRequest struct {
	AccessToken string `json:"access_token"`
	KrName      string `json:"krname"`
}

// DiscordRegister handles POST /register (after user completes /login as new).
func (h *Handlers) DiscordRegister(w http.ResponseWriter, r *http.Request) {
	var req discordRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	user, err := fetchDiscordUser(http.DefaultClient, req.AccessToken)
	if err != nil {
		http.Error(w, "discord lookup failed", http.StatusBadGateway)
		return
	}
	if _, err := db.SaveOAuthUser(h.DB, user.ID, user.Username, user.GlobalName, req.KrName); err != nil {
		http.Error(w, "save user failed", http.StatusInternalServerError)
		return
	}
	token, err := IssueToken(user.ID, h.Config.PrivateKey)
	if err != nil {
		http.Error(w, "token issue failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "complete",
		"userId":      user.ID,
		"username":    user.Username,
		"global_name": user.GlobalName,
		"token":       token,
	})
}
