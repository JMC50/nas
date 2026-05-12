package config

import (
	"os"
	"path/filepath"
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
	return nil
}

func defaultDataDir(isProd bool, sub string) string {
	if isProd {
		return filepath.Join("/app", sub)
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "..", "..", "nas-"+sub)
}
