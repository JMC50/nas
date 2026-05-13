package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/JMC50/nas/internal/db"
)

// Identities lists the current user's connected providers (no secrets).
func (h *Handlers) Identities(w http.ResponseWriter, r *http.Request) {
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
	identities, err := db.ListIdentity(h.DB, user.ID)
	if err != nil {
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}
	out := make([]map[string]string, 0, len(identities))
	for _, identity := range identities {
		out = append(out, map[string]string{
			"provider":   identity.Provider,
			"externalId": identity.ExternalID,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"identities": out})
}

// Unlink removes one provider link. Rejects the request when it would leave
// the user with zero sign-in methods.
func (h *Handlers) Unlink(w http.ResponseWriter, r *http.Request) {
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
	provider := chi.URLParam(r, "provider")
	if status, message := h.canUnlink(user.ID, provider); status != 0 {
		http.Error(w, message, status)
		return
	}
	if err := db.DropIdentity(h.DB, user.ID, provider); err != nil {
		http.Error(w, "remove failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed", "provider": provider})
}

func (h *Handlers) canUnlink(userID int64, provider string) (int, string) {
	identities, err := db.ListIdentity(h.DB, userID)
	if err != nil {
		return http.StatusInternalServerError, "list failed"
	}
	if len(identities) <= 1 {
		return http.StatusConflict, "Cannot remove your last sign-in method"
	}
	if !hasIdentity(identities, provider) {
		return http.StatusNotFound, "not linked"
	}
	return 0, ""
}

func hasIdentity(identities []db.Identity, provider string) bool {
	for _, identity := range identities {
		if identity.Provider == provider {
			return true
		}
	}
	return false
}
