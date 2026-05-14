package config

import (
	"bytes"
	"log/slog"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func absoluteForOS(leaf string) string {
	if runtime.GOOS == "windows" {
		return `C:\mnt\` + leaf
	}
	return "/mnt/" + leaf
}

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

func TestWarnEscapingPaths(t *testing.T) {
	cases := []struct {
		name       string
		envValues  map[string]string
		expectWarn bool
		wantVar    string
	}{
		{
			name:      "relative path escaping cwd warns",
			envValues: map[string]string{"NAS_DATA_DIR": "../escaping-test"},
			expectWarn: true,
			wantVar:   "NAS_DATA_DIR",
		},
		{
			name:      "relative path inside cwd does not warn",
			envValues: map[string]string{"NAS_DATA_DIR": "./data/nas"},
			expectWarn: false,
		},
		{
			name:      "absolute path is trusted (no warn)",
			envValues: map[string]string{"NAS_DATA_DIR": absoluteForOS("nas-storage")},
			expectWarn: false,
		},
		{
			name:      "empty value is ignored",
			envValues: map[string]string{"NAS_DATA_DIR": ""},
			expectWarn: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			oldLogger := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
			defer slog.SetDefault(oldLogger)

			c := &Config{}
			for k, v := range tc.envValues {
				switch k {
				case "NAS_DATA_DIR":
					c.NASDataDir = v
				case "NAS_ADMIN_DATA_DIR":
					c.NASAdminDataDir = v
				case "DB_PATH":
					c.DBPath = v
				}
			}
			c.auditPaths()

			output := buf.String()
			if tc.expectWarn {
				require.Contains(t, output, "path resolves outside cwd")
				if tc.wantVar != "" {
					require.Contains(t, output, tc.wantVar)
				}
			} else {
				require.NotContains(t, output, "path resolves outside cwd")
			}
		})
	}
}
