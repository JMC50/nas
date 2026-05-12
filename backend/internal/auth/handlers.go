package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
)

type Handlers struct {
	Config *config.Config
	DB     *sql.DB
}

type registerRequest struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Password string `json:"password"`
	KrName   string `json:"krname"`
}

type loginRequest struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type authResponse struct {
	Success bool     `json:"success"`
	Token   string   `json:"token,omitempty"`
	User    *userDTO `json:"user,omitempty"`
	Message string   `json:"message,omitempty"`
}

type userDTO struct {
	UserID     string   `json:"userId"`
	Username   string   `json:"username"`
	KrName     string   `json:"krname,omitempty"`
	GlobalName string   `json:"global_name,omitempty"`
	Intents    []string `json:"intents,omitempty"`
	AuthType   string   `json:"auth_type"`
}

func (h *Handlers) AuthConfig(w http.ResponseWriter, r *http.Request) {
	authType := string(h.Config.AuthType)
	resp := map[string]any{
		"authType":         authType,
		"localAuthEnabled": authType == "local" || authType == "both",
		"oauthEnabled":     authType == "oauth" || authType == "both",
		"passwordRequirements": map[string]any{
			"minLength":        h.Config.PasswordRequirements.MinLength,
			"requireUppercase": h.Config.PasswordRequirements.RequireUppercase,
			"requireLowercase": h.Config.PasswordRequirements.RequireLowercase,
			"requireNumber":    h.Config.PasswordRequirements.RequireNumber,
			"requireSpecial":   h.Config.PasswordRequirements.RequireSpecial,
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) RegisterLocal(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: "invalid request body"})
		return
	}
	if req.UserID == "" || req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: "User ID, username, and password are required"})
		return
	}
	if h.Config.AuthType == config.AuthTypeOAuth {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: "Local authentication is disabled. Please use OAuth."})
		return
	}

	existing, err := db.GetUser(h.DB, req.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, authResponse{Message: "lookup failed"})
		return
	}
	if existing != nil {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: "User ID already exists"})
		return
	}

	if err := ValidatePassword(req.Password, h.Config.PasswordRequirements); err != nil {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: err.Error()})
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, authResponse{Message: "hash failed"})
		return
	}
	if _, err := db.SaveLocalUser(h.DB, req.UserID, req.Username, hash, req.KrName); err != nil {
		writeJSON(w, http.StatusInternalServerError, authResponse{Message: "Failed to create user"})
		return
	}
	token, err := IssueToken(req.UserID, h.Config.PrivateKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, authResponse{Message: "token issue failed"})
		return
	}
	writeJSON(w, http.StatusOK, authResponse{
		Success: true,
		Token:   token,
		User: &userDTO{
			UserID:     req.UserID,
			Username:   req.Username,
			KrName:     req.KrName,
			GlobalName: req.Username,
			AuthType:   "local",
		},
	})
}

func (h *Handlers) LoginLocal(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: "invalid request body"})
		return
	}
	if req.UserID == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: "User ID and password are required"})
		return
	}
	if h.Config.AuthType == config.AuthTypeOAuth {
		writeJSON(w, http.StatusUnauthorized, authResponse{Message: "Local authentication is disabled. Please use OAuth."})
		return
	}

	user, err := db.GetUser(h.DB, req.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, authResponse{Message: "lookup failed"})
		return
	}
	if user == nil || user.AuthType != "local" || !VerifyPassword(req.Password, user.Password) {
		writeJSON(w, http.StatusUnauthorized, authResponse{Message: "Invalid user ID or password"})
		return
	}

	token, err := IssueToken(user.UserID, h.Config.PrivateKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, authResponse{Message: "token issue failed"})
		return
	}
	writeJSON(w, http.StatusOK, authResponse{
		Success: true,
		Token:   token,
		User: &userDTO{
			UserID:     user.UserID,
			Username:   user.Username,
			KrName:     user.KrName,
			GlobalName: orString(user.GlobalName, user.Username),
			Intents:    user.Intents,
			AuthType:   "local",
		},
	})
}

// ChangePassword requires a valid JWT (extracted from query or Bearer).
// Looks up the user, verifies old password, rehashes new.
func (h *Handlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	raw := ExtractToken(r)
	if raw == "" {
		writeJSON(w, http.StatusUnauthorized, authResponse{Message: "Token required"})
		return
	}
	claims, err := ParseToken(raw, h.Config.PrivateKey)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, authResponse{Message: "Invalid token"})
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: "invalid request body"})
		return
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: "Old password and new password are required"})
		return
	}

	user, err := db.GetUser(h.DB, claims.UserID)
	if err != nil || user == nil || user.AuthType != "local" {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: "User not found"})
		return
	}
	if !VerifyPassword(req.OldPassword, user.Password) {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: "Invalid old password"})
		return
	}
	if err := ValidatePassword(req.NewPassword, h.Config.PasswordRequirements); err != nil {
		writeJSON(w, http.StatusBadRequest, authResponse{Message: err.Error()})
		return
	}
	newHash, err := HashPassword(req.NewPassword)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, authResponse{Message: "hash failed"})
		return
	}
	if err := db.UpdatePassword(h.DB, claims.UserID, newHash); err != nil {
		writeJSON(w, http.StatusInternalServerError, authResponse{Message: "Failed to change password"})
		return
	}
	writeJSON(w, http.StatusOK, authResponse{Success: true, Message: "Password changed successfully"})
}

// GetIntents returns the list of intents for a user (no auth required, matches legacy).
func (h *Handlers) GetIntents(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	user, err := db.GetUser(h.DB, userID)
	if err != nil || user == nil {
		writeJSON(w, http.StatusOK, map[string]any{"intents": []string{}})
		return
	}
	intents := user.Intents
	if intents == nil {
		intents = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"intents": intents})
}

// CheckAdmin returns {"isAdmin": bool} for the requesting user (token-bound).
func (h *Handlers) CheckAdmin(w http.ResponseWriter, r *http.Request) {
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
	isAdmin := false
	for _, intent := range user.Intents {
		if intent == "ADMIN" {
			isAdmin = true
			break
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"isAdmin": isAdmin})
}

// CheckIntent returns {"status": bool} for whether the requester has the given intent.
func (h *Handlers) CheckIntent(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	intent := r.URL.Query().Get("intent")
	has, err := db.HasIntent(h.DB, claims.UserID, intent)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"status": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": has})
}

// GetAllUsers returns all users with their intents. Note: legacy version had no auth gate.
// Phase 1 mirrors that behavior; admin gating moves in via REST cleanup phase.
func (h *Handlers) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := db.GetAllUsers(h.DB)
	if err != nil {
		http.Error(w, "lookup failed", http.StatusInternalServerError)
		return
	}
	dtos := make([]userDTO, 0, len(users))
	for _, user := range users {
		dtos = append(dtos, userDTO{
			UserID:     user.UserID,
			Username:   user.Username,
			KrName:     user.KrName,
			GlobalName: user.GlobalName,
			Intents:    user.Intents,
			AuthType:   user.AuthType,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": dtos})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func orString(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
