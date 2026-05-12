package upload

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/JMC50/nas/internal/files"
)

// extractZipInto unzips zipPath into extractDir. Returns error if any entry
// would escape extractDir (zip-slip protection).
func extractZipInto(zipPath, extractDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer reader.Close()

	for _, entry := range reader.File {
		if err := extractOneEntry(entry, extractDir); err != nil {
			return err
		}
	}
	return nil
}

func extractOneEntry(entry *zip.File, extractDir string) error {
	// Block zip-slip: reject entries whose path escapes extractDir.
	cleaned := filepath.Clean(entry.Name)
	if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
		return fmt.Errorf("zip entry %q escapes extract dir", entry.Name)
	}
	target, err := files.SafeJoin(extractDir, cleaned)
	if err != nil {
		return fmt.Errorf("zip entry %q: %w", entry.Name, err)
	}
	if entry.FileInfo().IsDir() {
		return os.MkdirAll(target, entry.Mode())
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	source, err := entry.Open()
	if err != nil {
		return err
	}
	defer source.Close()
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, entry.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, source); err != nil {
		return err
	}
	return nil
}
