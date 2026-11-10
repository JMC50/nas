package share

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/files"
)

var errLinkClosed = errors.New("link closed")

type sharedTarget struct {
	link      *db.ShareLink
	creatorID string
	root      *os.Root
	rel       string
}

// Download는 로그인 없이 링크의 파일이나 폴더 zip을 보내고, 받을 수 없는 링크에는 이유 없이 404로 답합니다.
func (h *Handlers) Download(w http.ResponseWriter, r *http.Request) {
	target, err := h.openShared(chi.URLParam(r, "token"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer target.root.Close()
	release, status := h.limits.acquire(target.link.ID, target.link.IsFolder)
	if status != 0 {
		w.Header().Set("Retry-After", retryAfterSeconds)
		http.Error(w, http.StatusText(status), status)
		return
	}
	defer release()
	guard := newStreamGuard(w, func() error { return h.recheck(target.link.ID) })
	defer guard.finish()
	stopWatch := guard.watch()
	defer stopWatch()

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if target.link.IsFolder {
		h.serveFolder(guard, target)
		return
	}
	h.serveFile(guard, r, target)
}

// openShared는 링크를 찾아 받을 수 있는지 보고 데이터 폴더의 os.Root를 연 채로 돌려줌. root는 부른 쪽이 닫음
func (h *Handlers) openShared(token string) (*sharedTarget, error) {
	link, err := db.GetShareLinkByHash(h.DB, hashToken(token))
	if err != nil {
		return nil, err
	}
	creatorID, err := h.usable(link)
	if err != nil {
		return nil, err
	}
	rel, err := targetPath(link.Loc, link.Name)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(h.Config.NASDataDir)
	if err != nil {
		return nil, err
	}
	info, err := statTarget(root, rel)
	if err != nil || info.IsDir() != link.IsFolder {
		root.Close()
		return nil, errLinkClosed
	}
	return &sharedTarget{link: link, creatorID: creatorID, root: root, rel: rel}, nil
}

func (h *Handlers) serveFolder(guard *streamGuard, target *sharedTarget) {
	h.recordDownload(target.link, target.creatorID)
	items := []zipItem{{rel: target.rel, name: target.link.Name}}
	if err := streamZip(guard, target.root, items, target.link.Name+".zip", zipLimits{}); err != nil {
		slog.Warn("share folder zip stopped", "link", target.link.ID, "err", err)
		// 끝맺지 않은 zip을 정상 응답처럼 마치지 않도록 연결을 끊음
		panic(http.ErrAbortHandler)
	}
}

func (h *Handlers) serveFile(guard *streamGuard, r *http.Request, target *sharedTarget) {
	file, info, err := openRegular(target.root, target.rel)
	if err != nil {
		http.NotFound(guard, r)
		return
	}
	defer file.Close()
	// 이어받기 요청마다 세지 않도록 처음부터 받는 요청만 셈
	if rangeHeader := r.Header.Get("Range"); rangeHeader == "" || strings.HasPrefix(rangeHeader, "bytes=0-") {
		h.recordDownload(target.link, target.creatorID)
	}
	guard.Header().Set("Content-Type", "application/octet-stream")
	guard.Header().Set("Content-Disposition", files.ContentDispositionAttachment(target.link.Name))
	http.ServeContent(guard, r, target.link.Name, info.ModTime(), file)
}

// usable은 끊김, 만료, 만든 사람의 존재와 다운로드 권한을 보고 만든 사람의 userId를 돌려줌
func (h *Handlers) usable(link *db.ShareLink) (string, error) {
	if link == nil || link.RevokedAt.Valid || time.Now().UnixMilli() >= link.ExpiresAt {
		return "", errLinkClosed
	}
	creator, err := db.GetUserByPrimaryKey(h.DB, link.UserID)
	if err != nil || creator == nil {
		return "", errLinkClosed
	}
	allowed, err := db.HasIntent(h.DB, creator.UserID, string(auth.IntentDownload))
	if err != nil || !allowed {
		return "", errLinkClosed
	}
	return creator.UserID, nil
}

func (h *Handlers) recheck(linkID int64) error {
	link, err := db.GetShareLink(h.DB, linkID)
	if err != nil {
		return err
	}
	_, err = h.usable(link)
	return err
}

func (h *Handlers) recordDownload(link *db.ShareLink, creatorID string) {
	now := time.Now()
	if !h.limits.shouldRecord(link.ID, now) {
		return
	}
	if err := db.CountShareDownload(h.DB, link.ID); err != nil {
		slog.Warn("share download count failed", "link", link.ID, "err", err)
	}
	description := "DOWNLOAD [LINK] AT " + strings.TrimSuffix(link.Loc, "/") + "/" + link.Name
	if err := db.InsertLog(h.DB, creatorID, "DOWNLOAD", description, link.Loc, now.UnixMilli()); err != nil {
		slog.Warn("share download log failed", "link", link.ID, "err", err)
	}
}
