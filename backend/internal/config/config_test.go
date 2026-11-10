package config

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
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

func TestLoadFromEnv_ZipDownloadLimits(t *testing.T) {
	cfg, err := LoadFromEnv()
	require.NoError(t, err)
	require.Equal(t, 20000, cfg.ZipDownloadMaxFiles)
	require.Equal(t, int64(50)<<30, cfg.ZipDownloadMaxBytes)

	t.Setenv("ZIP_DOWNLOAD_MAX_FILES", "0")
	t.Setenv("ZIP_DOWNLOAD_MAX_SIZE", "2gb")
	cfg, err = LoadFromEnv()
	require.NoError(t, err)
	require.Equal(t, 0, cfg.ZipDownloadMaxFiles)
	require.Equal(t, int64(2)<<30, cfg.ZipDownloadMaxBytes)

	t.Setenv("ZIP_DOWNLOAD_MAX_SIZE", "lots")
	_, err = LoadFromEnv()
	require.ErrorContains(t, err, "ZIP_DOWNLOAD_MAX_SIZE")
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

func setOIDCEnv(t *testing.T) {
	t.Helper()
	t.Setenv("AUTH_TYPE", "oidc")
	t.Setenv("OIDC_ISSUER", "https://login.example/auth/v1")
	t.Setenv("OIDC_CLIENT_ID", "nas")
	t.Setenv("OIDC_CLIENT_SECRET", "client-secret")
	t.Setenv("OIDC_REDIRECT_URI", "https://nas.example/auth/oidc/callback")
}

func TestLoadFromEnv_OIDCSettingsAreIgnoredOutsideOIDCMode(t *testing.T) {
	cfg, err := LoadFromEnv()
	require.NoError(t, err)
	require.Equal(t, AuthTypeBoth, cfg.AuthType)
	require.Empty(t, cfg.OIDC.Issuer)

	for _, authType := range []string{"local", "oauth", "both"} {
		t.Setenv("AUTH_TYPE", authType)
		t.Setenv("OIDC_ISSUER", "not a url")
		t.Setenv("OIDC_SCOPES", "profile")
		_, err := LoadFromEnv()
		require.NoError(t, err, authType)
	}
}

func TestLoadFromEnv_OIDCNamesMissingSettings(t *testing.T) {
	t.Setenv("AUTH_TYPE", "oidc")
	_, err := LoadFromEnv()
	require.ErrorContains(t, err, "AUTH_TYPE=oidc requires OIDC_ISSUER, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, OIDC_REDIRECT_URI")

	t.Setenv("OIDC_ISSUER", "https://login.example/auth/v1")
	t.Setenv("OIDC_CLIENT_ID", "nas")
	_, err = LoadFromEnv()
	require.ErrorContains(t, err, "AUTH_TYPE=oidc requires OIDC_CLIENT_SECRET, OIDC_REDIRECT_URI")
	require.NotContains(t, err.Error(), "OIDC_ISSUER")
}

func TestLoadFromEnv_OIDCDefaults(t *testing.T) {
	setOIDCEnv(t)
	t.Setenv("OIDC_MEMBER_GROUP", "")
	cfg, err := LoadFromEnv()
	require.NoError(t, err)
	require.Equal(t, AuthTypeOIDC, cfg.AuthType)
	require.Equal(t, []string{"openid", "profile", "email", "groups"}, cfg.OIDC.Scopes)
	require.Equal(t, "groups", cfg.OIDC.GroupsClaim)
	require.Equal(t, "nas_member", cfg.OIDC.MemberGroup)
	require.Equal(t, "nas_admin", cfg.OIDC.AdminGroup)
	require.Empty(t, cfg.OIDC.IDPHint)
	require.Empty(t, cfg.OIDC.LogoutURL)
}

func TestLoadFromEnv_OIDCRejectsInvalidValues(t *testing.T) {
	for key, value := range map[string]string{
		"OIDC_ISSUER":       "login.example",
		"OIDC_REDIRECT_URI": "/auth/oidc/callback",
		"OIDC_LOGOUT_URL":   "javascript:alert(1)",
		"OIDC_SCOPES":       "profile email",
	} {
		t.Run(key, func(t *testing.T) {
			setOIDCEnv(t)
			t.Setenv(key, value)
			_, err := LoadFromEnv()
			require.ErrorContains(t, err, key)
			require.NotContains(t, err.Error(), "client-secret")
		})
	}
}

func TestLoadFromEnv_OIDCClientSecretFromFile(t *testing.T) {
	setOIDCEnv(t)
	t.Setenv("OIDC_CLIENT_SECRET", "")
	path := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.WriteFile(path, []byte("secret-from-file\n"), 0o600))
	t.Setenv("OIDC_CLIENT_SECRET_FILE", path)

	cfg, err := LoadFromEnv()
	require.NoError(t, err)
	require.Equal(t, "secret-from-file", cfg.OIDC.ClientSecret)
	require.NotContains(t, cfg.String(), "secret-from-file")
}
