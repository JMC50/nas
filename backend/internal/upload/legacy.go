package upload

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/files"
)

type Handlers struct {
	Config *config.Config
	DB     *sql.DB
}

// LegacyInput mirrors the Node POST /input endpoint: raw body streamed to disk.
// Frontend code that hasn't switched to tus yet keeps working.
func (h *Handlers) LegacyInput(w http.ResponseWriter, r *http.Request) {
	loc := r.URL.Query().Get("loc")
	name := r.URL.Query().Get("name")
	target, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(loc), name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		http.Error(w, "mkdir failed", http.StatusInternalServerError)
		return
	}
	out, err := os.Create(target)
	if err != nil {
		http.Error(w, "create failed", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, r.Body); err != nil {
		http.Error(w, "write failed", http.StatusInternalServerError)
		return
	}
	h.logActivity(r, "UPLOAD", fmt.Sprintf("UPLOAD [FILE] AT /%s", convertLoc(loc+"/"+name)), "/"+loc)
	w.Write([]byte("complete"))
}

// LegacyInputZip mirrors POST /inputZip: stream body to disk, then unzip in place.
func (h *Handlers) LegacyInputZip(w http.ResponseWriter, r *http.Request) {
	loc := r.URL.Query().Get("loc")
	name := r.URL.Query().Get("name")
	target, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(loc), name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	extractDir, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(loc))
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		http.Error(w, "mkdir failed", http.StatusInternalServerError)
		return
	}
	out, err := os.Create(target)
	if err != nil {
		http.Error(w, "create failed", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(out, r.Body); err != nil {
		out.Close()
		http.Error(w, "write failed", http.StatusInternalServerError)
		return
	}
	out.Close()

	if err := extractZipInto(target, extractDir); err != nil {
		http.Error(w, "extract failed", http.StatusInternalServerError)
		return
	}
	_ = os.Remove(target) // remove the temp zip
	h.logActivity(r, "UPLOAD", fmt.Sprintf("UPLOAD [FOLDER] AT /%s", convertLoc(loc+"/"+name)), "/"+loc)
	w.Write([]byte("complete"))
}

func (h *Handlers) logActivity(r *http.Request, activity, description, loc string) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		return
	}
	_ = db.InsertLog(h.DB, claims.UserID, activity, description, loc, time.Now().UnixMilli())
}

// convertLoc copy from files package to avoid cyclic imports
func convertLoc(loc string) string {
	cleaned := loc
	for len(cleaned) > 0 && (cleaned[0] == '/' || cleaned[0] == '\\') {
		cleaned = cleaned[1:]
	}
	return cleaned
}
