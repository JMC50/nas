package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/config"
)

func TestValidatePassword_LengthEnforced(t *testing.T) {
	requirements := config.PasswordRequirements{MinLength: 8}
	require.Error(t, ValidatePassword("short", requirements))
	require.NoError(t, ValidatePassword("longenough", requirements))
}

func TestValidatePassword_AllRequirements(t *testing.T) {
	requirements := config.PasswordRequirements{
		MinLength:        10,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSpecial:   true,
	}
	require.Error(t, ValidatePassword("AAAAAAAAAA", requirements), "no lower/digit/special")
	require.Error(t, ValidatePassword("Aa1!short", requirements), "too short")
	require.NoError(t, ValidatePassword("StrongP@ssw0rd", requirements))
}
