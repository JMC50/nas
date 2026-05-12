package archive

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/files"
)

type Handlers struct {
	Config  *config.Config
	DB      *sql.DB
	Tracker *Tracker
}

type zipRequestItem struct {
	Loc      string `json:"loc"`
	Name     string `json:"name"`
	IsFolder bool   `json:"isFolder"`
}

func (h *Handlers) ZipFiles(w http.ResponseWriter, r *http.Request) {
	var items []zipRequestItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if len(items) == 0 {
		http.Error(w, "no files to zip", http.StatusBadRequest)
		return
	}

	zipName := fmt.Sprintf("archive_%d.zip", time.Now().UnixMilli())
	zipPath, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(items[0].Loc), zipName)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}

	progressID := uuid.NewString()
	h.Tracker.Set(progressID, Progress{Percent: 0, Status: "zipping"})

	output, err := os.Create(zipPath)
	if err != nil {
		http.Error(w, "create zip failed", http.StatusInternalServerError)
		return
	}
	defer output.Close()
	zipWriter := zip.NewWriter(output)
	defer zipWriter.Close()

	totalEntries := len(items)
	completed := 0
	for _, item := range items {
		source, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(item.Loc), item.Name)
		if err != nil {
			h.Tracker.Set(progressID, Progress{Percent: 100, Status: "error"})
			http.Error(w, "unsafe source path", http.StatusBadRequest)
			return
		}
		if item.IsFolder {
			if err := addDirectory(zipWriter, source, item.Name); err != nil {
				h.Tracker.Set(progressID, Progress{Percent: 100, Status: "error"})
				http.Error(w, "failed to zip files", http.StatusInternalServerError)
				return
			}
		} else {
			if err := addFile(zipWriter, source, item.Name); err != nil {
				h.Tracker.Set(progressID, Progress{Percent: 100, Status: "error"})
				http.Error(w, "failed to zip files", http.StatusInternalServerError)
				return
			}
		}
		completed++
		h.Tracker.Set(progressID, Progress{
			Percent: completed * 100 / totalEntries,
			Status:  "zipping",
		})
	}

	h.Tracker.Set(progressID, Progress{Percent: 100, Status: "done"})
	writeJSON(w, http.StatusOK, map[string]string{
		"zipPath":    zipPath,
		"progressId": progressID,
	})
}

type unzipRequest struct {
	Loc        string `json:"loc"`
	Name       string `json:"name"`
	Extensions string `json:"extensions"`
}

func (h *Handlers) UnzipFile(w http.ResponseWriter, r *http.Request) {
	var req unzipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Extensions != "zip" {
		http.Error(w, "invalid file", http.StatusBadRequest)
		return
	}

	zipPath, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(req.Loc), req.Name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	stem := strings.TrimSuffix(req.Name, filepath.Ext(req.Name))
	extractDir, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(req.Loc), stem+"_unzipped")
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}

	progressID := uuid.NewString()
	h.Tracker.Set(progressID, Progress{Percent: 0, Status: "unzipping"})

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		h.Tracker.Set(progressID, Progress{Percent: 100, Status: "error"})
		http.Error(w, "failed to open zip", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		h.Tracker.Set(progressID, Progress{Percent: 100, Status: "error"})
		http.Error(w, "mkdir failed", http.StatusInternalServerError)
		return
	}

	total := len(reader.File)
	for index, entry := range reader.File {
		if err := extractEntry(extractDir, entry); err != nil {
			h.Tracker.Set(progressID, Progress{Percent: 100, Status: "error"})
			http.Error(w, "zip processing error", http.StatusInternalServerError)
			return
		}
		h.Tracker.Set(progressID, Progress{
			Percent: (index + 1) * 100 / total,
			Status:  "unzipping",
		})
	}

	h.Tracker.Set(progressID, Progress{Percent: 100, Status: "done"})
	writeJSON(w, http.StatusOK, map[string]string{
		"extractedPath": extractDir,
		"progressId":    progressID,
	})
}

func (h *Handlers) Progress(w http.ResponseWriter, r *http.Request) {
	progressID := r.URL.Query().Get("progressId")
	if progressID == "" {
		http.Error(w, "progressId required", http.StatusBadRequest)
		return
	}
	progress, ok := h.Tracker.Get(progressID)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, progress)
}

// DownloadZip serves the temp zip produced by /zipFiles.
// Verifies zipPath stays under NAS_DATA_DIR.
func (h *Handlers) DownloadZip(w http.ResponseWriter, r *http.Request) {
	rawPath, err := url.QueryUnescape(r.URL.Query().Get("zipPath"))
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	abs, err := filepath.Abs(rawPath)
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	baseAbs, err := filepath.Abs(h.Config.NASDataDir)
	if err != nil {
		http.Error(w, "config error", http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(abs, baseAbs+string(filepath.Separator)) && abs != baseAbs {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	file, err := os.Open(abs)
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
	w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(abs)+`"`)
	http.ServeContent(w, r, filepath.Base(abs), info.ModTime(), file)
}

func (h *Handlers) DeleteTempZip(w http.ResponseWriter, r *http.Request) {
	rawPath, err := url.QueryUnescape(r.URL.Query().Get("path"))
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	abs, err := filepath.Abs(rawPath)
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	baseAbs, err := filepath.Abs(h.Config.NASDataDir)
	if err != nil {
		http.Error(w, "config error", http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(abs, baseAbs+string(filepath.Separator)) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	if err := os.Remove(abs); err != nil {
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("complete"))
}

// --- zip helpers ---

func addFile(writer *zip.Writer, source, archiveName string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()
	entry, err := writer.Create(archiveName)
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, file)
	return err
}

func addDirectory(writer *zip.Writer, source, archiveName string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		archivePath := filepath.ToSlash(filepath.Join(archiveName, relative))
		if info.IsDir() {
			if archivePath == archiveName {
				return nil
			}
			_, err := writer.Create(archivePath + "/")
			return err
		}
		return addFile(writer, path, archivePath)
	})
}

func extractEntry(extractDir string, entry *zip.File) error {
	target, err := files.SafeJoin(extractDir, entry.Name)
	if err != nil {
		return fmt.Errorf("zip slip blocked for %q: %w", entry.Name, err)
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
	_, err = io.Copy(out, source)
	return err
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
