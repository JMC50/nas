package admin

import (
	"encoding/json"
	"net/http"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/db"
)

// GetOAuthConfig returns the admin-facing OAuth credentials view. Caller must
// hold ADMIN intent. Secrets are masked — only the presence flag is returned.
func (h *Handlers) GetOAuthConfig(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	writeView(w, auth.ViewCreds(h.Config, h.DB))
}

// UpdateOAuthConfig persists OAuth provider credentials. Empty client_id or
// redirect_uri disables the provider; empty secret preserves the stored value.
func (h *Handlers) UpdateOAuthConfig(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	var update auth.OAuthCredsUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := auth.WriteCreds(h.DB, update); err != nil {
		http.Error(w, "save failed", http.StatusInternalServerError)
		return
	}
	writeView(w, auth.ViewCreds(h.Config, h.DB))
}

func (h *Handlers) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	caller, err := db.GetUser(h.DB, claims.UserID)
	if err != nil || caller == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return false
	}
	if !hasAdmin(caller.Intents) {
		http.Error(w, "not found", http.StatusNotFound)
		return false
	}
	return true
}

func writeView(w http.ResponseWriter, view auth.OAuthCredsView) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(view)
}
