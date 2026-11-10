package share

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/files"
)

var (
	errLinkClosed   = errors.New("link closed")
	errNotShareable = errors.New("not a regular file or folder")
)

type resolvedTarget struct {
	path string
	root string
	info os.FileInfo
}

// Download는 로그인 없이 링크의 파일이나 폴더 zip을 보내고, 받을 수 없는 링크에는 이유 없이 404로 답합니다.
func (h *Handlers) Download(w http.ResponseWriter, r *http.Request) {
	link, creatorID, target, err := h.openLink(chi.URLParam(r, "token"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if link.IsFolder {
		h.recordDownload(link, creatorID)
		streamFolder(w, target, link.Name)
		return
	}
	file, err := os.Open(target.path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	// 이어받기 요청마다 세지 않도록 처음부터 받는 요청만 셈
	if rangeHeader := r.Header.Get("Range"); rangeHeader == "" || strings.HasPrefix(rangeHeader, "bytes=0-") {
		h.recordDownload(link, creatorID)
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", files.ContentDispositionAttachment(link.Name))
	http.ServeContent(w, r, link.Name, target.info.ModTime(), file)
}

func (h *Handlers) openLink(token string) (*db.ShareLink, string, resolvedTarget, error) {
	link, err := db.GetShareLinkByHash(h.DB, hashToken(token))
	if err != nil {
		return nil, "", resolvedTarget{}, err
	}
	if link == nil || link.RevokedAt.Valid || time.Now().UnixMilli() >= link.ExpiresAt {
		return nil, "", resolvedTarget{}, errLinkClosed
	}
	creator, err := db.GetUserByPrimaryKey(h.DB, link.UserID)
	if err != nil || creator == nil {
		return nil, "", resolvedTarget{}, errLinkClosed
	}
	allowed, err := db.HasIntent(h.DB, creator.UserID, string(auth.IntentDownload))
	if err != nil || !allowed {
		return nil, "", resolvedTarget{}, errLinkClosed
	}
	target, err := h.resolve(link.Loc, link.Name)
	if err != nil || target.info.IsDir() != link.IsFolder {
		return nil, "", resolvedTarget{}, errLinkClosed
	}
	return link, creator.UserID, target, nil
}

func (h *Handlers) recordDownload(link *db.ShareLink, creatorID string) {
	if err := db.CountShareDownload(h.DB, link.ID); err != nil {
		slog.Warn("share download count failed", "link", link.ID, "err", err)
	}
	description := "DOWNLOAD [LINK] AT " + strings.TrimSuffix(link.Loc, "/") + "/" + link.Name
	if err := db.InsertLog(h.DB, creatorID, "DOWNLOAD", description, link.Loc, time.Now().UnixMilli()); err != nil {
		slog.Warn("share download log failed", "link", link.ID, "err", err)
	}
}

func (h *Handlers) resolve(loc, name string) (resolvedTarget, error) {
	joined, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(loc), name)
	if err != nil {
		return resolvedTarget{}, err
	}
	root, err := filepath.EvalSymlinks(h.Config.NASDataDir)
	if err != nil {
		return resolvedTarget{}, err
	}
	resolved, err := filepath.EvalSymlinks(joined)
	if err != nil {
		return resolvedTarget{}, err
	}
	if !insideRoot(root, resolved) || resolved == root {
		return resolvedTarget{}, files.ErrUnsafePath
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return resolvedTarget{}, err
	}
	if !info.Mode().IsRegular() && !info.IsDir() {
		return resolvedTarget{}, errNotShareable
	}
	return resolvedTarget{path: resolved, root: root, info: info}, nil
}

func insideRoot(root, path string) bool {
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}
