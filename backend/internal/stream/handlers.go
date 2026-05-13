package stream

import (
	"database/sql"
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/files"
)

//go:embed icons/*.png
var iconsFS embed.FS

type Handlers struct {
	Config *config.Config
	DB     *sql.DB
}

// serveFile is the shared core: SafeJoin → Stat → http.ServeContent for range support.
func (h *Handlers) serveFile(w http.ResponseWriter, r *http.Request, contentType string) {
	loc := r.URL.Query().Get("loc")
	name := r.URL.Query().Get("name")
	target, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(loc), name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	file, err := os.Open(target)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		http.Error(w, "stat failed", http.StatusInternalServerError)
		return
	}
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	http.ServeContent(w, r, name, info.ModTime(), file)
}

func (h *Handlers) Video(w http.ResponseWriter, r *http.Request) {
	h.serveFile(w, r, contentTypeFor(r.URL.Query().Get("name"), "video/mp4"))
}

func (h *Handlers) Audio(w http.ResponseWriter, r *http.Request) {
	h.serveFile(w, r, contentTypeFor(r.URL.Query().Get("name"), "audio/mpeg"))
}

func (h *Handlers) Image(w http.ResponseWriter, r *http.Request) {
	h.serveFile(w, r, contentTypeFor(r.URL.Query().Get("name"), "application/octet-stream"))
}

// Download serves the file with Content-Disposition: attachment so the browser
// triggers a save dialog. Range requests still work via http.ServeContent.
func (h *Handlers) Download(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	w.Header().Set("Content-Disposition", files.ContentDispositionAttachment(name))
	h.serveFile(w, r, "")
}

// Img serves bundled file-type icons from embedded FS. Falls back to file.png
// if the requested type has no specific icon.
func (h *Handlers) Img(w http.ResponseWriter, r *http.Request) {
	iconType := r.URL.Query().Get("type")
	data, err := iconsFS.ReadFile("icons/" + iconType + ".png")
	if err != nil {
		data, err = iconsFS.ReadFile("icons/file.png")
		if err != nil {
			http.Error(w, "no fallback icon", http.StatusNotFound)
			return
		}
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

func contentTypeFor(name, fallback string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".flac":
		return "audio/flac"
	}
	return fallback
}

// IconNames returns the list of bundled icon types (without .png suffix).
// Useful for diagnostics — not currently routed.
func IconNames() []string {
	names := []string{}
	_ = fs.WalkDir(iconsFS, "icons", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		base := strings.TrimSuffix(filepath.Base(path), ".png")
		names = append(names, base)
		return nil
	})
	return names
}
