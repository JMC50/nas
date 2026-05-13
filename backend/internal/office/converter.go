package office

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const ConvertTimeout = 30 * time.Second

// Convert runs LibreOffice headless to produce a PDF.
// It writes to a fresh temp subdir under `workDir` to avoid basename collisions
// between concurrent conversions of different files that share a filename, then
// returns the path to the generated PDF (caller is responsible for moving/renaming).
func Convert(ctx context.Context, srcPath, workDir string) (string, error) {
	bin, err := exec.LookPath("soffice")
	if err != nil {
		return "", fmt.Errorf("soffice not found: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, ConvertTimeout)
	defer cancel()

	// Per-call temp dir prevents basename collisions when two requests with
	// different content but identical filenames run in parallel (same package,
	// different folders — dedupe only collapses same-hash, not same-name).
	tmpDir, err := os.MkdirTemp(workDir, "conv-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}
	// Caller cleans up tmpDir after moving the PDF; we don't defer-remove here
	// because the caller needs the output path to still exist after we return.

	cmd := exec.CommandContext(ctx, bin,
		"--headless",
		"--convert-to", "pdf",
		"--outdir", tmpDir,
		srcPath,
	)
	// LibreOffice writes both progress and errors to stderr; capture for diagnostics.
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("soffice timeout after %v: %s", ConvertTimeout, stderr.String())
		}
		return "", fmt.Errorf("soffice failed: %w: %s", err, stderr.String())
	}

	base := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	outPath := filepath.Join(tmpDir, base+".pdf")
	if _, err := os.Stat(outPath); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("expected output not found: %w", err)
	}
	return outPath, nil
}
