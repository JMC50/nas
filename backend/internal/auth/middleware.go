package auth

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/JMC50/nas/internal/db"
)

type contextKey int

const (
	ctxClaims contextKey = iota
)

// ExtractToken pulls the JWT from (1) `?token=…` query param (legacy compat)
// or (2) `Authorization: Bearer …` header (new clients). Empty string if neither.
func ExtractToken(r *http.Request) string {
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// ClaimsFromContext returns the verified claims attached by RequireToken middleware.
// Returns nil if RequireToken did not run before this handler.
func ClaimsFromContext(ctx context.Context) *NodeCompatClaims {
	claims, _ := ctx.Value(ctxClaims).(*NodeCompatClaims)
	return claims
}

// RequireToken parses + verifies the JWT and attaches claims to context.
// Returns 401 (not 500) on missing/invalid token — matches REST conventions, departs
// from legacy's `res.status(500).end("wtf is this token")` which was a defect.
func RequireToken(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := ExtractToken(r)
			if raw == "" {
				http.Error(w, "token required", http.StatusUnauthorized)
				return
			}
			claims, err := ParseToken(raw, secret)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ctxClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireIntent enforces that the requesting user has `intent` (or ADMIN).
// Must run after RequireToken — looks up claims from context.
func RequireIntent(conn *sql.DB, intent Intent) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				http.Error(w, "auth context missing", http.StatusInternalServerError)
				return
			}
			ok, err := db.HasIntent(conn, claims.UserID, string(intent))
			if err != nil {
				http.Error(w, "intent check failed", http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
