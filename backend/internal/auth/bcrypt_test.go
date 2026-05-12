package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashAndVerify_RoundTrip(t *testing.T) {
	hash, err := HashPassword("test-password-123")
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.True(t, VerifyPassword("test-password-123", hash))
	require.False(t, VerifyPassword("wrong-password", hash))
}

// TestVerifyPassword_StoredHashFormat verifies an externally-generated bcrypt $2a$
// hash (the format bcryptjs in the legacy Node backend produces) is verifiable by
// golang.org/x/crypto/bcrypt. Both libraries follow the OpenBSD bcrypt spec so $2a$/$2b$
// hashes are mutually compatible.
func TestVerifyPassword_StoredHashFormat(t *testing.T) {
	// Pre-generated $2a$ hash for the password "password" at cost 10
	storedHash := "$2a$10$vRi5Usv2mbZffA.DCR0JPepvl.ZRO.yhBLXBItsD1Nfznc.wxsDMe"

	require.True(t, VerifyPassword("password", storedHash), "Go bcrypt must accept stored $2a$ hash")
	require.False(t, VerifyPassword("wrong", storedHash))
}

func TestVerifyPassword_2bFormat(t *testing.T) {
	// $2b$ is the format Go's bcrypt produces. Generate fresh and re-verify.
	hash, err := HashPassword("another-password")
	require.NoError(t, err)
	// Hash should start with $2a$ or $2b$ (both are bcrypt)
	require.Regexp(t, `^\$2[ab]\$10\$`, hash)
	require.True(t, VerifyPassword("another-password", hash))
}
