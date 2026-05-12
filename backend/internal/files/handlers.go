package files

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
)

type Handlers struct {
	Config *config.Config
	DB     *sql.DB
}

type fileEntry struct {
	Name       string `json:"name"`
	IsFolder   bool   `json:"isFolder"`
	Extensions string `json:"extensions"`
	Loc        string `json:"loc,omitempty"`
}

func (h *Handlers) ReadFolder(w http.ResponseWriter, r *http.Request) {
	loc := decode(r.URL.Query().Get("loc"))
	target, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(loc))
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		http.Error(w, "read folder failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	results := make([]fileEntry, 0, len(entries))
	for _, entry := range entries {
		extension := extensionOf(entry.Name())
		results = append(results, fileEntry{
			Name:       entry.Name(),
			IsFolder:   entry.IsDir(),
			Extensions: extension,
		})
	}
	writeJSON(w, http.StatusOK, results)
}

type statResponse struct {
	Name       string `json:"name"`
	Size       string `json:"size,omitempty"`
	Type       string `json:"type"`
	CreatedAt  string `json:"createdAt"`
	ModifiedAt string `json:"modifiedAt"`
}

func (h *Handlers) Stat(w http.ResponseWriter, r *http.Request) {
	loc := decode(r.URL.Query().Get("loc"))
	name := decode(r.URL.Query().Get("name"))
	target, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(loc), name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	info, err := os.Stat(target)
	if err != nil {
		http.Error(w, "stat failed", http.StatusNotFound)
		return
	}
	resp := statResponse{
		Name:       name,
		Type:       "file",
		CreatedAt:  info.ModTime().Format(time.RFC3339),
		ModifiedAt: info.ModTime().Format(time.RFC3339),
	}
	if info.IsDir() {
		resp.Type = "folder"
	} else {
		resp.Size = formatSize(info.Size())
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) GetTextFile(w http.ResponseWriter, r *http.Request) {
	loc := decode(r.URL.Query().Get("loc"))
	name := decode(r.URL.Query().Get("name"))
	target, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(loc), name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	content, err := os.ReadFile(target)
	if err != nil {
		http.Error(w, "read failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"name": name, "content": string(content)})
}

type saveTextBody struct {
	Text string `json:"text"`
}

func (h *Handlers) SaveTextFile(w http.ResponseWriter, r *http.Request) {
	loc := decode(r.URL.Query().Get("loc"))
	name := decode(r.URL.Query().Get("name"))
	target, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(loc), name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	var body saveTextBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := os.WriteFile(target, []byte(body.Text), 0o644); err != nil {
		http.Error(w, "write failed", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("complete"))
}

func (h *Handlers) MakeDir(w http.ResponseWriter, r *http.Request) {
	loc := decode(r.URL.Query().Get("loc"))
	name := decode(r.URL.Query().Get("name"))
	target, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(loc), name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	if _, err := os.Stat(target); err == nil {
		w.Write([]byte("failed")) // already exists — matches legacy "failed"
		return
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		http.Error(w, "mkdir failed", http.StatusInternalServerError)
		return
	}
	h.logActivity(r, "UPLOAD", fmt.Sprintf("CREATE [FOLDER] AT /%s", convertLoc(loc+"/"+name)), "/"+loc)
	w.Write([]byte("complete"))
}

func (h *Handlers) ForceDelete(w http.ResponseWriter, r *http.Request) {
	loc := decode(r.URL.Query().Get("loc"))
	name := decode(r.URL.Query().Get("name"))
	target, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(loc), name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	h.logActivity(r, "DELETE", fmt.Sprintf("DELETE [FILE] AT /%s", convertLoc(loc+"/"+name)), "/"+loc)
	if err := os.RemoveAll(target); err != nil {
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("complete"))
}

func (h *Handlers) Copy(w http.ResponseWriter, r *http.Request) {
	originLoc := decode(r.URL.Query().Get("originLoc"))
	fileName := decode(r.URL.Query().Get("fileName"))
	targetLoc := decode(r.URL.Query().Get("targetLoc"))
	source, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(originLoc), fileName)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	target, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(targetLoc), fileName)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	h.logActivity(r, "COPY",
		fmt.Sprintf("COPY [FILE] FROM /%s TO /%s",
			convertLoc(originLoc+"/"+fileName),
			convertLoc(targetLoc+"/"+fileName)),
		"/"+targetLoc)
	if err := copyRecursive(source, target); err != nil {
		http.Error(w, "copy failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("complete"))
}

func (h *Handlers) Move(w http.ResponseWriter, r *http.Request) {
	originLoc := decode(r.URL.Query().Get("originLoc"))
	fileName := decode(r.URL.Query().Get("fileName"))
	targetLoc := decode(r.URL.Query().Get("targetLoc"))
	source, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(originLoc), fileName)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	target, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(targetLoc), fileName)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	h.logActivity(r, "MOVE",
		fmt.Sprintf("MOVE [FILE] FROM /%s TO /%s",
			convertLoc(originLoc+"/"+fileName),
			convertLoc(targetLoc+"/"+fileName)),
		"/"+targetLoc)
	if err := os.Rename(source, target); err != nil {
		// Cross-FS fallback: copy then delete
		if err := copyRecursive(source, target); err != nil {
			http.Error(w, "move failed", http.StatusInternalServerError)
			return
		}
		if err := os.RemoveAll(source); err != nil {
			http.Error(w, "move cleanup failed", http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("complete"))
}

func (h *Handlers) Rename(w http.ResponseWriter, r *http.Request) {
	loc := decode(r.URL.Query().Get("loc"))
	name := decode(r.URL.Query().Get("name"))
	change := decode(r.URL.Query().Get("change"))
	source, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(loc), name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	target, err := SafeJoin(h.Config.NASDataDir, TrimLeadingSlash(loc), change)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}
	h.logActivity(r, "RENAME",
		fmt.Sprintf("RENAME [FILE] AT /%s TO /%s",
			convertLoc(loc+"/"+name), convertLoc(loc+"/"+change)),
		"/"+loc)
	if err := os.Rename(source, target); err != nil {
		http.Error(w, "rename failed", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("complete"))
}

func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(decode(r.URL.Query().Get("query")))
	if query == "" {
		writeJSON(w, http.StatusOK, []fileEntry{})
		return
	}
	results := []fileEntry{}
	baseDir := h.Config.NASDataDir
	err := filepath.WalkDir(baseDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == baseDir {
			return nil
		}
		name := entry.Name()
		if !strings.Contains(strings.ToLower(name), query) {
			return nil
		}
		relative, err := filepath.Rel(baseDir, filepath.Dir(path))
		if err != nil {
			return nil
		}
		results = append(results, fileEntry{
			Name:       name,
			IsFolder:   entry.IsDir(),
			Extensions: extensionOf(name),
			Loc:        "/" + filepath.ToSlash(relative),
		})
		return nil
	})
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

// logActivity is fire-and-forget — failures are logged but do not block the request.
func (h *Handlers) logActivity(r *http.Request, activity, description, loc string) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		return
	}
	_ = db.InsertLog(h.DB, claims.UserID, activity, description, loc, time.Now().UnixMilli())
}

func copyRecursive(source, target string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return copyFile(source, target, info.Mode())
	}
	if err := os.MkdirAll(target, info.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		targetPath := filepath.Join(target, entry.Name())
		if err := copyRecursive(sourcePath, targetPath); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(source, target string, mode os.FileMode) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

func decode(s string) string {
	// Legacy uses decodeURIComponent. chi has already URL-decoded query params,
	// so this is a passthrough — kept as a hook for future re-encoding.
	return s
}

func extensionOf(name string) string {
	if idx := strings.LastIndex(name, "."); idx > 0 && idx < len(name)-1 {
		return name[idx+1:]
	}
	return "file"
}

// convertLoc mirrors Node's convertLoc: trim leading slashes from joined path.
func convertLoc(loc string) string {
	parts := strings.Split(loc, "/")
	out := []string{}
	skipLeading := true
	for _, segment := range parts {
		if skipLeading && segment == "" {
			continue
		}
		skipLeading = false
		out = append(out, segment)
	}
	return strings.Join(out, "/")
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func formatSize(bytes int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

