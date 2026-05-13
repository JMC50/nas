package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/JMC50/nas/internal/db"
)

const linkStatePrefix = "link."

type linkStartBody struct {
	Provider string `json:"provider"`
}

type linkStartResponse struct {
	AuthorizeURL string `json:"authorizeUrl"`
}

type linkDiscordBody struct {
	State       string `json:"state"`
	AccessToken string `json:"access_token"`
}

type linkGoogleBody struct {
	State string `json:"state"`
	Code  string `json:"code"`
}

// LinkStart issues a state nonce and returns an OAuth authorize URL whose
// state param marks this flow as "link to current user" instead of sign-in.
func (h *Handlers) LinkStart(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := db.GetUser(h.DB, claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var body linkStartBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	creds := ResolveCreds(h.Config, h.DB)
	authorizeURL, err := h.buildLinkURL(user.ID, body.Provider, creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, linkStartResponse{AuthorizeURL: authorizeURL})
}

func (h *Handlers) buildLinkURL(userID int64, provider string, creds OAuthCreds) (string, error) {
	nonce, err := h.Links.Issue(userID, provider)
	if err != nil {
		return "", errors.New("nonce issue failed")
	}
	state := linkStatePrefix + nonce
	switch provider {
	case "discord":
		if !discordOK(creds) {
			return "", errors.New("Discord OAuth not configured")
		}
		return discordURL(creds.DiscordClientID, creds.DiscordRedirectURI, state), nil
	case "google":
		if !googleOK(creds) {
			return "", errors.New("Google OAuth not configured")
		}
		return googleURL(creds.GoogleClientID, creds.GoogleRedirectURI, state), nil
	default:
		return "", errors.New("unknown provider")
	}
}

// LinkDiscord finalizes a Discord link round-trip. The SPA callback at /login
// detects the state prefix and POSTs the access_token here instead of running
// the sign-in flow.
func (h *Handlers) LinkDiscord(w http.ResponseWriter, r *http.Request) {
	var body linkDiscordBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	userID, ok := h.consumeLink(body.State, "discord")
	if !ok {
		http.Error(w, "invalid or expired link state", http.StatusBadRequest)
		return
	}
	user, err := fetchDiscordUser(http.DefaultClient, body.AccessToken)
	if err != nil {
		http.Error(w, "discord lookup failed", http.StatusBadGateway)
		return
	}
	if err := db.AddIdentity(h.DB, userID, "discord", user.ID); err != nil {
		linkError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "linked", "provider": "discord"})
}

func (h *Handlers) LinkGoogle(w http.ResponseWriter, r *http.Request) {
	var body linkGoogleBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	userID, ok := h.consumeLink(body.State, "google")
	if !ok {
		http.Error(w, "invalid or expired link state", http.StatusBadRequest)
		return
	}
	user, err := h.googleUser(body.Code)
	if err != nil {
		http.Error(w, "google lookup failed", http.StatusBadGateway)
		return
	}
	if err := db.AddIdentity(h.DB, userID, "google", user.ID); err != nil {
		linkError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "linked", "provider": "google"})
}

func (h *Handlers) googleUser(code string) (*googleUserResponse, error) {
	creds := ResolveCreds(h.Config, h.DB)
	accessToken, err := exchangeGoogleCode(http.DefaultClient,
		creds.GoogleClientID, creds.GoogleClientSecret, creds.GoogleRedirectURI, code)
	if err != nil {
		return nil, err
	}
	return fetchGoogleUser(http.DefaultClient, accessToken)
}

func (h *Handlers) consumeLink(state, provider string) (int64, bool) {
	if !strings.HasPrefix(state, linkStatePrefix) {
		return 0, false
	}
	nonce := state[len(linkStatePrefix):]
	userID, recordedProvider, ok := h.Links.Consume(nonce)
	if !ok || recordedProvider != provider {
		return 0, false
	}
	return userID, true
}

func linkError(w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), "UNIQUE") {
		http.Error(w, "This provider account is already linked to another user", http.StatusConflict)
		return
	}
	http.Error(w, "save failed", http.StatusInternalServerError)
}

