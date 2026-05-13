package office

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/files"
)

type Handlers struct {
	Config *config.Config
	DB     *sql.DB
	Cache  *Cache
	Dedupe *Dedupe
}

func NewHandlers(cfg *config.Config, db *sql.DB) (*Handlers, error) {
	cache, err := NewCache(cfg.NASTempDir)
	if err != nil {
		return nil, err
	}
	return &Handlers{
		Config: cfg,
		DB:     db,
		Cache:  cache,
		Dedupe: &Dedupe{},
	}, nil
}

func (h *Handlers) GetOfficePdf(w http.ResponseWriter, r *http.Request) {
	loc := r.URL.Query().Get("loc")
	name := r.URL.Query().Get("name")
	src, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(loc), name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(src); err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	hash, err := HashFile(src)
	if err != nil {
		http.Error(w, "hash failed", http.StatusInternalServerError)
		return
	}

	pdfPath := h.Cache.Path(hash)
	if !h.Cache.Hit(hash) {
		release, isLeader := h.Dedupe.Acquire(hash)
		if isLeader {
			outPath, err := Convert(r.Context(), src, h.Cache.dir)
			if err != nil {
				release()
				http.Error(w, "conversion failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			// soffice writes to a per-call temp dir; move into the cache as <hash>.pdf.
			// Write to .tmp first then atomic-rename so partial writes don't poison the cache.
			tmpFinal := pdfPath + ".tmp"
			if err := os.Rename(outPath, tmpFinal); err != nil {
				os.RemoveAll(filepath.Dir(outPath))
				release()
				http.Error(w, "cache write failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			os.RemoveAll(filepath.Dir(outPath))
			if err := os.Rename(tmpFinal, pdfPath); err != nil {
				release()
				http.Error(w, "cache finalize failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			// Release immediately after the cache file is materialized — followers
			// can now serve from cache in parallel with the leader's response stream.
			release()
		}
		// follower (or post-leader): re-check cache existence
		if !h.Cache.Hit(hash) {
			http.Error(w, "conversion produced no output", http.StatusInternalServerError)
			return
		}
	}

	file, err := os.Open(pdfPath)
	if err != nil {
		http.Error(w, "open cached pdf failed", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(w, "stat cached pdf failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	http.ServeContent(w, r, name+".pdf", info.ModTime().UTC().Truncate(time.Second), file)
}
