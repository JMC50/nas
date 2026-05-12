package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFromEnv_RequiredFields(t *testing.T) {
	t.Setenv("PRIVATE_KEY", "test-key")
	t.Setenv("ADMIN_PASSWORD", "test-admin")
	t.Setenv("PORT", "7777")

	cfg, err := LoadFromEnv()
	require.NoError(t, err)
	require.Equal(t, "test-key", cfg.PrivateKey)
	require.Equal(t, "test-admin", cfg.AdminPassword)
	require.Equal(t, 7777, cfg.Port)
}

func TestLoadFromEnv_ProductionRequiresSecrets(t *testing.T) {
	t.Setenv("NODE_ENV", "production")
	os.Unsetenv("PRIVATE_KEY")
	_, err := LoadFromEnv()
	require.Error(t, err)
	require.Contains(t, err.Error(), "PRIVATE_KEY")
}

func TestLoadFromEnv_DockerSecretFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "secret")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())
	_, _ = tmp.WriteString("secret-from-file")
	tmp.Close()

	t.Setenv("PRIVATE_KEY_FILE", tmp.Name())
	t.Setenv("ADMIN_PASSWORD", "x")
	cfg, err := LoadFromEnv()
	require.NoError(t, err)
	require.Equal(t, "secret-from-file", cfg.PrivateKey)
}
