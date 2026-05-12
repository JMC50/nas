package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-private-key"

func TestIssueAndParse_RoundTrip(t *testing.T) {
	token, err := IssueToken("user123", testSecret)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := ParseToken(token, testSecret)
	require.NoError(t, err)
	require.Equal(t, "user123", claims.UserID)
	require.Greater(t, claims.ExpiresIn, time.Now().UnixMilli())
}

func TestParseToken_RejectsWrongSecret(t *testing.T) {
	token, err := IssueToken("user123", testSecret)
	require.NoError(t, err)

	_, err = ParseToken(token, "wrong-secret")
	require.Error(t, err)
}

func TestParseToken_RejectsExpired(t *testing.T) {
	claims := NodeCompatClaims{
		UserID:    "user123",
		ExpiresIn: time.Now().UnixMilli() - 1000,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expired, err := tok.SignedString([]byte(testSecret))
	require.NoError(t, err)

	_, err = ParseToken(expired, testSecret)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

// TestParseToken_NodeStylePayload verifies that a token signed with the same payload
// shape Node's jsonwebtoken produces ({userId, expires_in}) is accepted.
func TestParseToken_NodeStylePayload(t *testing.T) {
	// Simulate Node: jwt.sign({userId, expires_in: Date.now()+7d}, private_key)
	exp := time.Now().UnixMilli() + 7*24*60*60*1000
	nodeClaims := jwt.MapClaims{
		"userId":     "discord-user-456",
		"expires_in": exp,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, nodeClaims)
	signed, err := tok.SignedString([]byte(testSecret))
	require.NoError(t, err)

	claims, err := ParseToken(signed, testSecret)
	require.NoError(t, err)
	require.Equal(t, "discord-user-456", claims.UserID)
	require.Equal(t, exp, claims.ExpiresIn)
}

// TestParseToken_LegacyTokensWithoutExpiry guards against null/zero expires_in:
// the legacy Node code's jwt.verify did NOT enforce expiry. Ensure tokens without
// an expires_in claim still parse (claims.ExpiresIn == 0 → expiry check bypassed).
func TestParseToken_LegacyTokensWithoutExpiry(t *testing.T) {
	nodeClaims := jwt.MapClaims{"userId": "old-user"}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, nodeClaims)
	signed, err := tok.SignedString([]byte(testSecret))
	require.NoError(t, err)

	claims, err := ParseToken(signed, testSecret)
	require.NoError(t, err)
	require.Equal(t, "old-user", claims.UserID)
	require.Equal(t, int64(0), claims.ExpiresIn)
}
