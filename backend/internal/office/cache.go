package office

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

type Cache struct {
	dir string
}

func NewCache(rootDir string) (*Cache, error) {
	dir := filepath.Join(rootDir, "office-cache")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Cache{dir: dir}, nil
}

func (c *Cache) Path(hash string) string {
	return filepath.Join(c.dir, hash+".pdf")
}

func (c *Cache) Hit(hash string) bool {
	info, err := os.Stat(c.Path(hash))
	return err == nil && info.Size() > 0
}

func HashFile(srcPath string) (string, error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
