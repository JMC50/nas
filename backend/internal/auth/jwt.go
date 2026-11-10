package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
)

// tokenLifetimeMs matches Node legacy: 7 days in milliseconds (Date.now() + 7d).
// Stored as a custom `expires_in` claim (NOT the standard JWT `exp` in seconds).
const tokenLifetimeMs int64 = 7 * 24 * 60 * 60 * 1000

// NodeCompatClaims matches the legacy Node payload exactly:
//
//	jwt.sign({ userId, expires_in: Date.now() + 7d }, private_key)
//
// expires_in is milliseconds-since-epoch, NOT the standard JWT `exp` (seconds).
// Manual expiry check required at verify time.
//
// issued_at uses the same unit. Tokens from before v0.7.0 do not have it.
type NodeCompatClaims struct {
	UserID     string `json:"userId"`
	ExpiresIn  int64  `json:"expires_in"`
	IssuedAtMs int64  `json:"issued_at,omitempty"`
	jwt.RegisteredClaims
}

// ErrInvalidToken is returned by Authenticate for a token that must not be accepted.
var ErrInvalidToken = errors.New("invalid token")

func IssueToken(userID, secret string) (string, error) {
	now := time.Now().UnixMilli()
	claims := NodeCompatClaims{
		UserID:     userID,
		ExpiresIn:  now + tokenLifetimeMs,
		IssuedAtMs: now,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// ParseToken verifies the signature and the custom `expires_in` claim.
// Returns the parsed claims on success.
func ParseToken(raw, secret string) (*NodeCompatClaims, error) {
	claims := &NodeCompatClaims{}
	parsed, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !parsed.Valid {
		return nil, errors.New("token invalid")
	}
	if claims.ExpiresIn > 0 && time.Now().UnixMilli() >= claims.ExpiresIn {
		return nil, errors.New("token expired")
	}
	return claims, nil
}

// Authenticate runs ParseToken and then rejects a token that was issued before
// the user's tokens_valid_after time. Any other error is a failed lookup.
func Authenticate(cfg *config.Config, conn *sql.DB, raw string) (*NodeCompatClaims, error) {
	claims, err := ParseToken(raw, cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	validAfter, found, err := db.TokensValidAfter(conn, claims.UserID)
	if err != nil {
		return nil, err
	}
	if !found {
		// oidc 모드에서는 지워진 사용자의 토큰을 받지 않음. 다른 모드는 통과시키고 권한 확인에 맡김
		if cfg.AuthType == config.AuthTypeOIDC {
			return nil, fmt.Errorf("%w: user removed", ErrInvalidToken)
		}
		return claims, nil
	}
	if claims.IssuedAtMs < validAfter {
		return nil, fmt.Errorf("%w: issued before tokens_valid_after", ErrInvalidToken)
	}
	return claims, nil
}
