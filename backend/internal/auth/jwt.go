package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
type NodeCompatClaims struct {
	UserID    string `json:"userId"`
	ExpiresIn int64  `json:"expires_in"`
	jwt.RegisteredClaims
}

func IssueToken(userID, secret string) (string, error) {
	claims := NodeCompatClaims{
		UserID:    userID,
		ExpiresIn: time.Now().UnixMilli() + tokenLifetimeMs,
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
