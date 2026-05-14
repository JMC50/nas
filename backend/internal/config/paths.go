package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolvePaths fills in NASDataDir/NASAdminDataDir/DBPath defaults based on platform + env.
// Explicit env values always win.
func (c *Config) ResolvePaths() error {
	if c.NASDataDir == "" {
		c.NASDataDir = defaultDataDir(c.IsProduction, "data")
	}
	if c.NASAdminDataDir == "" {
		c.NASAdminDataDir = defaultDataDir(c.IsProduction, "admin-data")
	}
	if c.DBPath == "" {
		base := defaultDataDir(c.IsProduction, "db")
		c.DBPath = filepath.Join(base, "nas.sqlite")
	}

	for _, dir := range []string{c.NASDataDir, c.NASAdminDataDir, filepath.Dir(c.DBPath), c.NASTempDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	// Resolve the app version from the repo-root VERSION file. Search order:
	//   1. APP_VERSION env (set by Docker/CI build)
	//   2. ./VERSION (binary cwd — works for `go run ./cmd/server` from backend/)
	//   3. ../VERSION (repo root from backend/)
	// Falls back to "0.0.0-unknown" so the server still starts.
	if c.Version == "" {
		if v := os.Getenv("APP_VERSION"); v != "" {
			c.Version = strings.TrimSpace(v)
		} else {
			for _, candidate := range []string{"VERSION", filepath.Join("..", "VERSION")} {
				if data, err := os.ReadFile(candidate); err == nil {
					c.Version = strings.TrimSpace(string(data))
					break
				}
			}
		}
		if c.Version == "" {
			c.Version = "0.0.0-unknown"
		}
	}
	return nil
}

func defaultDataDir(isProd bool, sub string) string {
	if isProd {
		return filepath.Join("/app", sub)
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "..", "..", "nas-"+sub)
}
