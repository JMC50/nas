package share

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/db"
)

const (
	maxBundleBodyBytes = 1 << 20
	bundleFallbackName = "nas-download"
)

type bundleRequest struct {
	Items []struct {
		Loc  string `json:"loc"`
		Name string `json:"name"`
	} `json:"items"`
}

// PrepareZip은 로그인한 사용자가 고른 파일과 폴더를 검사하고, zip 하나로 받을 짧은 수명의 주소를 돌려줍니다.
func (h *Handlers) PrepareZip(w http.ResponseWriter, r *http.Request) {
	rels, err := decodeBundle(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	caller, _, err := h.caller(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	root, err := os.OpenRoot(h.Config.NASDataDir)
	if err != nil {
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	defer root.Close()
	items, err := bundleItems(root, rels)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	usage, err := measureZip(r.Context(), root, items, h.bundleZipLimits())
	if err != nil {
		refuseBundle(w, err, usage.limits)
		return
	}
	if status := h.bundleLimits.zipBusy(caller.ID); status != 0 {
		w.Header().Set("Retry-After", retryAfterSeconds)
		http.Error(w, http.StatusText(status), status)
		return
	}
	loc, _ := splitTarget(rels[0])
	ticket := bundleTicket{userID: caller.ID, items: items, filename: bundleFilename(items), loc: loc}
	id, err := h.bundleTickets.issue(ticket)
	if err != nil {
		http.Error(w, "too many pending zip downloads", http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"path":    "/zip-downloads/" + id,
		"name":    ticket.filename,
		"entries": usage.entries,
		"bytes":   usage.bytes,
	})
}

// DownloadZip은 PrepareZip이 내준 주소로 zip을 바로 써서 보내고, 다른 사용자의 주소에는 없는 주소처럼 404로 답합니다.
func (h *Handlers) DownloadZip(w http.ResponseWriter, r *http.Request) {
	caller, _, err := h.caller(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ticket, ok := h.bundleTickets.find(chi.URLParam(r, "id"))
	if !ok || ticket.userID != caller.ID {
		http.NotFound(w, r)
		return
	}
	root, err := os.OpenRoot(h.Config.NASDataDir)
	if err != nil {
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	defer root.Close()
	release, status := h.bundleLimits.acquire(caller.ID, true)
	if status != 0 {
		w.Header().Set("Retry-After", retryAfterSeconds)
		http.Error(w, http.StatusText(status), status)
		return
	}
	defer release()
	guard := newStreamGuard(w, func() error { return h.bundleAllowed(r.Context(), caller.UserID) })
	defer guard.finish()
	stopWatch := guard.watch()
	defer stopWatch()

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	h.recordBundle(caller.UserID, ticket)
	if err := streamZip(guard, root, ticket.items, ticket.filename, h.bundleZipLimits()); err != nil {
		slog.Warn("zip download stopped", "user", caller.ID, "err", err)
		// 끝맺지 않은 zip을 정상 응답처럼 마치지 않도록 연결을 끊음
		panic(http.ErrAbortHandler)
	}
}

// decodeBundle은 요청한 항목을 데이터 폴더 기준의 정리된 상대 경로로 바꾸고 같은 경로는 한 번만 남김
func decodeBundle(w http.ResponseWriter, r *http.Request) ([]string, error) {
	var req bundleRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBundleBodyBytes)).Decode(&req); err != nil {
		return nil, errors.New("invalid body")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("items must not be empty")
	}
	seen := map[string]bool{}
	rels := make([]string, 0, len(req.Items))
	for _, item := range req.Items {
		rel, err := targetPath(item.Loc, item.Name)
		if err != nil {
			return nil, err
		}
		if !seen[rel] {
			seen[rel] = true
			rels = append(rels, rel)
		}
	}
	return rels, nil
}

// bundleItems는 모든 대상이 데이터 폴더 안의 보통 파일이나 폴더인지 보고 zip 맨 위 이름을 겹치지 않게 정함
func bundleItems(root *os.Root, rels []string) ([]zipItem, error) {
	used := map[string]bool{}
	items := make([]zipItem, 0, len(rels))
	for _, rel := range rels {
		info, err := statTarget(root, rel)
		if err != nil {
			return nil, err
		}
		items = append(items, zipItem{rel: rel, name: uniqueName(used, path.Base(rel), info.IsDir())})
	}
	return items, nil
}

// Windows와 macOS는 대소문자만 다른 이름을 같은 이름으로 풀므로 소문자로 비교함
func uniqueName(used map[string]bool, name string, folder bool) string {
	stem, ext := name, ""
	if dot := strings.LastIndex(name, "."); !folder && dot > 0 {
		stem, ext = name[:dot], name[dot:]
	}
	candidate := name
	for n := 2; used[strings.ToLower(candidate)]; n++ {
		candidate = fmt.Sprintf("%s (%d)%s", stem, n, ext)
	}
	used[strings.ToLower(candidate)] = true
	return candidate
}

func bundleFilename(items []zipItem) string {
	if len(items) == 1 {
		return items[0].name + ".zip"
	}
	loc := path.Dir(items[0].rel)
	for _, item := range items[1:] {
		if path.Dir(item.rel) != loc {
			return bundleFallbackName + ".zip"
		}
	}
	if loc == "." {
		return bundleFallbackName + ".zip"
	}
	return path.Base(loc) + ".zip"
}

func (h *Handlers) bundleZipLimits() zipLimits {
	return zipLimits{entries: h.Config.ZipDownloadMaxFiles, bytes: h.Config.ZipDownloadMaxBytes}
}

func refuseBundle(w http.ResponseWriter, err error, limits zipLimits) {
	switch {
	case errors.Is(err, errTooManyEntries):
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{"error": "too_many_entries", "limit": limits.entries})
	case errors.Is(err, errTooManyBytes):
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{"error": "too_large", "limit": limits.bytes})
	default:
		http.Error(w, "scan failed", http.StatusInternalServerError)
	}
}

// 받는 쪽이 연결을 끊었거나 받는 중에 다운로드 권한을 잃으면 보내기를 멈춤
func (h *Handlers) bundleAllowed(ctx context.Context, userID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	allowed, err := db.HasIntent(h.DB, userID, string(auth.IntentDownload))
	if err != nil {
		return err
	}
	if !allowed {
		return errLinkClosed
	}
	return nil
}

func (h *Handlers) recordBundle(userID string, ticket bundleTicket) {
	description := fmt.Sprintf("DOWNLOAD [ZIP] %d ITEMS AT %s", len(ticket.items), ticket.loc)
	if len(ticket.items) == 1 {
		description = "DOWNLOAD [ZIP] AT /" + ticket.items[0].rel
	}
	if err := db.InsertLog(h.DB, userID, "DOWNLOAD", description, ticket.loc, time.Now().UnixMilli()); err != nil {
		slog.Warn("zip download log failed", "err", err)
	}
}
