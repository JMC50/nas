package files

import (
	"errors"
	"path/filepath"
	"strings"
)

// ErrUnsafePath is returned when a join would escape the base directory.
var ErrUnsafePath = errors.New("unsafe path")

// SafeJoin returns base+segments joined and verified to stay strictly inside base.
// Blocks: ../ traversal, absolute paths, URL-encoded escapes (caller must decode first),
// symlink escapes (only after Stat — SafeJoin itself does lexical check; symlink
// targets are filesystem state and must be checked by callers if they care).
func SafeJoin(base string, segments ...string) (string, error) {
	cleanedBase, err := filepath.Abs(filepath.Clean(base))
	if err != nil {
		return "", err
	}
	for _, segment := range segments {
		if strings.ContainsRune(segment, 0) {
			return "", ErrUnsafePath
		}
		if filepath.IsAbs(segment) {
			return "", ErrUnsafePath
		}
	}
	joined := filepath.Join(append([]string{cleanedBase}, segments...)...)
	cleaned := filepath.Clean(joined)
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", err
	}
	if abs != cleanedBase && !strings.HasPrefix(abs, cleanedBase+string(filepath.Separator)) {
		return "", ErrUnsafePath
	}
	return abs, nil
}

// TrimLeadingSlash mirrors `loc.replace(/^\/+/, "")` from the legacy Node code.
// Used before SafeJoin so segments are treated as relative.
func TrimLeadingSlash(s string) string {
	return strings.TrimLeft(s, "/\\")
}
