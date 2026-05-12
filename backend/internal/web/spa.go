package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// MountSPA serves the SvelteKit/Vite build at `staticDir`.
// Unknown paths (non-file, non-API) fall back to index.html for client-side routing.
// Returns nil handler if the static dir doesn't exist (allows backend to run without frontend).
func MountSPA(staticDir string) http.Handler {
	info, err := os.Stat(staticDir)
	if err != nil || !info.IsDir() {
		return nil
	}
	fileServer := http.FileServer(http.Dir(staticDir))
	indexPath := filepath.Join(staticDir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API and known backend endpoints are handled before this; if we got here
		// it's a static asset or SPA route.
		requested := filepath.Join(staticDir, strings.TrimPrefix(r.URL.Path, "/"))
		fileInfo, statErr := os.Stat(requested)
		if statErr == nil && !fileInfo.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		// Fallback: serve index.html (SPA routing)
		http.ServeFile(w, r, indexPath)
	})
}
