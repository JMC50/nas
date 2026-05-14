package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type AuthType string

const (
	AuthTypeOAuth AuthType = "oauth"
	AuthTypeLocal AuthType = "local"
	AuthTypeBoth  AuthType = "both"
)

type PasswordRequirements struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

type Config struct {
	NodeEnv      string
	Port         int
	Host         string
	IsProduction bool

	PrivateKey    string
	AdminPassword string
	AuthType      AuthType
	JWTExpiry     string

	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURI  string
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURI   string

	NASDataDir      string
	NASAdminDataDir string
	NASTempDir      string
	DBPath          string
	FrontendDir     string

	PasswordRequirements PasswordRequirements
	CorsOrigin           string
	MaxFileSizeBytes     int64
	MediaLibraryLimit    int
}

func LoadFromEnv() (*Config, error) {
	c := &Config{
		NodeEnv: getEnv("NODE_ENV", "development"),
	}
	c.IsProduction = c.NodeEnv == "production"

	c.Port = getEnvInt("PORT", 7777)
	c.Host = getEnv("HOST", "0.0.0.0")
	c.AuthType = AuthType(getEnv("AUTH_TYPE", "both"))
	c.JWTExpiry = getEnv("JWT_EXPIRY", "168h")

	c.PrivateKey = getSecret("PRIVATE_KEY", "")
	c.AdminPassword = getSecret("ADMIN_PASSWORD", "")

	c.DiscordClientID = getSecret("DISCORD_CLIENT_ID", "")
	c.DiscordClientSecret = getSecret("DISCORD_CLIENT_SECRET", "")
	c.DiscordRedirectURI = getEnv("DISCORD_REDIRECT_URI", "")
	c.GoogleClientID = getSecret("GOOGLE_CLIENT_ID", "")
	c.GoogleClientSecret = getSecret("GOOGLE_CLIENT_SECRET", "")
	c.GoogleRedirectURI = getEnv("GOOGLE_REDIRECT_URI", "")

	c.NASDataDir = getEnv("NAS_DATA_DIR", "")
	c.NASAdminDataDir = getEnv("NAS_ADMIN_DATA_DIR", "")
	c.NASTempDir = getEnv("NAS_TEMP_DIR", os.TempDir())
	c.DBPath = getEnv("DB_PATH", "")
	c.FrontendDir = getEnv("FRONTEND_DIR", "")

	c.PasswordRequirements = PasswordRequirements{
		MinLength:        getEnvInt("PASSWORD_MIN_LENGTH", 8),
		RequireUppercase: getEnvBool("PASSWORD_REQUIRE_UPPERCASE", false),
		RequireLowercase: getEnvBool("PASSWORD_REQUIRE_LOWERCASE", false),
		RequireNumber:    getEnvBool("PASSWORD_REQUIRE_NUMBER", false),
		RequireSpecial:   getEnvBool("PASSWORD_REQUIRE_SPECIAL", false),
	}

	c.CorsOrigin = getEnv("CORS_ORIGIN", "*")
	maxSize, err := parseSize(getEnv("MAX_FILE_SIZE", "50gb"))
	if err != nil {
		return nil, fmt.Errorf("MAX_FILE_SIZE: %w", err)
	}
	c.MaxFileSizeBytes = maxSize
	c.MediaLibraryLimit = getEnvInt("MEDIA_LIB_LIMIT", 5000)

	if err := c.validate(); err != nil {
		return nil, err
	}
	c.auditPaths()
	return c, nil
}

func (c *Config) validate() error {
	var errs []string
	if c.IsProduction {
		if c.PrivateKey == "" {
			errs = append(errs, "PRIVATE_KEY required in production")
		}
		if c.AdminPassword == "" {
			errs = append(errs, "ADMIN_PASSWORD required in production")
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// auditPaths logs a warning when a relative path env var resolves outside
// the current working directory. Absolute paths are trusted (production NAS_DATA_DIR
// commonly points at /mnt/* or similar). Only catches the "../parent" footgun that
// caused stray nas-data, nas-admin-data, nas-db dirs to materialize above the repo.
func (c *Config) auditPaths() {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	checks := map[string]string{
		"NAS_DATA_DIR":       c.NASDataDir,
		"NAS_ADMIN_DATA_DIR": c.NASAdminDataDir,
		"NAS_TEMP_DIR":       c.NASTempDir,
		"DB_PATH":            c.DBPath,
		"FRONTEND_DIR":       c.FrontendDir,
	}
	for name, value := range checks {
		if value == "" || filepath.IsAbs(value) {
			continue
		}
		abs, err := filepath.Abs(value)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(cwd, abs)
		if err != nil {
			continue
		}
		if strings.HasPrefix(filepath.ToSlash(rel), "../") || rel == ".." {
			slog.Warn("path resolves outside cwd — data will land above the project",
				"var", name, "value", value, "resolved", abs)
		}
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		return v == "true"
	}
	return def
}

func getSecret(key, def string) string {
	if path, ok := os.LookupEnv(key + "_FILE"); ok {
		if data, err := os.ReadFile(path); err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return getEnv(key, def)
}

func parseSize(s string) (int64, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	var mult int64 = 1
	switch {
	case strings.HasSuffix(s, "gb"):
		mult = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "gb")
	case strings.HasSuffix(s, "mb"):
		mult = 1024 * 1024
		s = strings.TrimSuffix(s, "mb")
	case strings.HasSuffix(s, "kb"):
		mult = 1024
		s = strings.TrimSuffix(s, "kb")
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}
	return n * mult, nil
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{Env=%s, Port=%d, AuthType=%s, IsProd=%t}", c.NodeEnv, c.Port, c.AuthType, c.IsProduction)
}
