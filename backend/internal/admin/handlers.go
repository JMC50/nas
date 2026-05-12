package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
)

type Handlers struct {
	Config *config.Config
	DB     *sql.DB
}

// Authorize/Unauthorize toggle an intent for a target user. Caller must have ADMIN.
// Legacy semantics: both endpoints call editIntent (toggle), so they're functionally
// equivalent. Frontend distinguishes them by intent meaning, not API behavior.
func (h *Handlers) ToggleIntent(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	caller, err := db.GetUser(h.DB, claims.UserID)
	if err != nil || caller == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !hasAdmin(caller.Intents) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	target := r.URL.Query().Get("userId")
	intent := r.URL.Query().Get("intent")
	if target == "" || intent == "" {
		http.Error(w, "userId and intent required", http.StatusBadRequest)
		return
	}
	if err := db.ToggleIntent(h.DB, target, intent); err != nil {
		http.Error(w, "toggle failed", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("complete"))
}

type requestAdminBody struct {
	Password string `json:"pwd"`
}

func (h *Handlers) RequestAdminIntent(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req requestAdminBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Password != h.Config.AdminPassword {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	if err := db.ToggleIntent(h.DB, claims.UserID, "ADMIN"); err != nil {
		http.Error(w, "toggle failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("complete"))
}

func (h *Handlers) GetActivityLog(w http.ResponseWriter, r *http.Request) {
	logs, err := db.GetActivityLogs(h.DB)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": logs})
}

type logRequest struct {
	Activity    string `json:"activity"`
	Description string `json:"description"`
	Token       string `json:"token"`
	Time        int64  `json:"time"`
	Loc         string `json:"loc"`
}

// InsertLog is called by client to record activity. Legacy accepted token in body.
// Phase 1 mirrors that behavior, parsing the token to find user_id.
func (h *Handlers) InsertLog(w http.ResponseWriter, r *http.Request) {
	var req logRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	claims, err := auth.ParseToken(req.Token, h.Config.PrivateKey)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	if err := db.InsertLog(h.DB, claims.UserID, req.Activity, req.Description, req.Loc, req.Time); err != nil {
		http.Error(w, "insert failed", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("complete"))
}

func hasAdmin(intents []string) bool {
	for _, intent := range intents {
		if intent == "ADMIN" {
			return true
		}
	}
	return false
}
