package share

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/files"
)

type Handlers struct {
	Config *config.Config
	DB     *sql.DB
}

var expiryChoices = map[string]time.Duration{
	"1h":  time.Hour,
	"1d":  24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
}

const tokenBytes = 32

type createRequest struct {
	Loc       string `json:"loc"`
	Name      string `json:"name"`
	ExpiresIn string `json:"expiresIn"`
}

type linkView struct {
	ID            int64  `json:"id"`
	Loc           string `json:"loc"`
	Name          string `json:"name"`
	IsFolder      bool   `json:"isFolder"`
	CreatorID     string `json:"creatorId"`
	CreatorName   string `json:"creatorName"`
	CreatedAt     int64  `json:"createdAt"`
	ExpiresAt     int64  `json:"expiresAt"`
	RevokedAt     *int64 `json:"revokedAt"`
	DownloadCount int64  `json:"downloadCount"`
	Status        string `json:"status"`
}

// Create는 요청한 사용자가 다운로드 권한을 가진 파일이나 폴더에 공개 링크를 만듭니다. 토큰 원문은 이 응답에서만 돌려줍니다.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	lifetime, ok := expiryChoices[req.ExpiresIn]
	if !ok {
		http.Error(w, "expiresIn must be one of 1h, 1d, 7d, 30d", http.StatusBadRequest)
		return
	}
	if req.Name == "" || strings.ContainsAny(req.Name, `/\`) {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}
	creator, err := db.GetUser(h.DB, claims.UserID)
	if err != nil || creator == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	target, err := h.resolve(req.Loc, req.Name)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		http.Error(w, "token generation failed", http.StatusInternalServerError)
		return
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	now := time.Now()
	link := db.ShareLink{
		TokenHash: hashToken(token),
		UserID:    creator.ID,
		Loc:       normalizeLoc(req.Loc),
		Name:      req.Name,
		IsFolder:  target.info.IsDir(),
		CreatedAt: now.UnixMilli(),
		ExpiresAt: now.Add(lifetime).UnixMilli(),
	}
	id, err := db.CreateShareLink(h.DB, link)
	if err != nil {
		http.Error(w, "create failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":        id,
		"path":      "/d/" + token,
		"name":      link.Name,
		"isFolder":  link.IsFolder,
		"expiresAt": link.ExpiresAt,
	})
}

// List는 관리자에게 모든 링크를, 다른 사용자에게 자기 링크만 돌려줍니다.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	caller, admin, err := h.caller(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	creatorFilter := caller.ID
	if admin {
		creatorFilter = 0
	}
	links, err := db.ListShareLinks(h.DB, creatorFilter)
	if err != nil {
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}
	now := time.Now().UnixMilli()
	views := make([]linkView, 0, len(links))
	for _, link := range links {
		views = append(views, toView(link, now))
	}
	writeJSON(w, http.StatusOK, map[string]any{"links": views, "all": admin})
}

// Revoke는 링크를 만든 사용자나 관리자만 링크를 끊게 합니다. 남의 링크는 없는 링크처럼 404로 답합니다.
func (h *Handlers) Revoke(w http.ResponseWriter, r *http.Request) {
	caller, admin, err := h.caller(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	link, err := db.GetShareLink(h.DB, id)
	if err != nil {
		http.Error(w, "lookup failed", http.StatusInternalServerError)
		return
	}
	if link == nil || (!admin && link.UserID != caller.ID) {
		http.NotFound(w, r)
		return
	}
	if err := db.RevokeShareLink(h.DB, id, time.Now().UnixMilli()); err != nil {
		http.Error(w, "revoke failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// Download는 로그인 없이 링크의 파일을 보내거나 폴더를 zip으로 보냅니다.
// 받을 때마다 만료, 끊김, 만든 사람의 다운로드 권한, 실제 경로가 데이터 폴더 안인지를 다시 확인하고, 하나라도 어긋나면 이유 없이 404로 답합니다.
func (h *Handlers) Download(w http.ResponseWriter, r *http.Request) {
	link, err := db.GetShareLinkByHash(h.DB, hashToken(chi.URLParam(r, "token")))
	now := time.Now()
	if err != nil || link == nil || link.RevokedAt.Valid || now.UnixMilli() >= link.ExpiresAt {
		http.NotFound(w, r)
		return
	}
	creator, err := db.GetUserByPrimaryKey(h.DB, link.UserID)
	if err != nil || creator == nil {
		http.NotFound(w, r)
		return
	}
	allowed, err := db.HasIntent(h.DB, creator.UserID, string(auth.IntentDownload))
	if err != nil || !allowed {
		http.NotFound(w, r)
		return
	}
	target, err := h.resolve(link.Loc, link.Name)
	if err != nil || target.info.IsDir() != link.IsFolder {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if link.IsFolder {
		h.record(link, creator.UserID, now)
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
		h.record(link, creator.UserID, now)
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", files.ContentDispositionAttachment(link.Name))
	http.ServeContent(w, r, link.Name, target.info.ModTime(), file)
}

func (h *Handlers) record(link *db.ShareLink, creatorID string, at time.Time) {
	if err := db.CountShareDownload(h.DB, link.ID, at.UnixMilli()); err != nil {
		slog.Warn("share download count failed", "link", link.ID, "err", err)
	}
	description := "DOWNLOAD [LINK] AT " + strings.TrimSuffix(link.Loc, "/") + "/" + link.Name
	if err := db.InsertLog(h.DB, creatorID, "DOWNLOAD", description, link.Loc, at.UnixMilli()); err != nil {
		slog.Warn("share download log failed", "link", link.ID, "err", err)
	}
}

type resolvedTarget struct {
	path string
	root string
	info os.FileInfo
}

// resolve는 loc과 name을 데이터 폴더 기준 실제 경로로 바꿉니다. 심볼릭 링크를 따라간 실제 경로가 데이터 폴더 밖이거나, 대상이 보통 파일이나 폴더가 아니면 오류입니다.
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

var errNotShareable = errors.New("not a regular file or folder")

func insideRoot(root, path string) bool {
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}

func (h *Handlers) caller(r *http.Request) (*db.User, bool, error) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		return nil, false, errors.New("no claims")
	}
	user, err := db.GetUser(h.DB, claims.UserID)
	if err != nil || user == nil {
		return nil, false, errors.New("unknown user")
	}
	admin, err := db.HasIntent(h.DB, user.UserID, string(auth.IntentAdmin))
	if err != nil {
		return nil, false, err
	}
	return user, admin, nil
}

func toView(link *db.ShareLink, now int64) linkView {
	view := linkView{
		ID:            link.ID,
		Loc:           link.Loc,
		Name:          link.Name,
		IsFolder:      link.IsFolder,
		CreatorID:     link.CreatorUserID,
		CreatorName:   link.CreatorName,
		CreatedAt:     link.CreatedAt,
		ExpiresAt:     link.ExpiresAt,
		DownloadCount: link.DownloadCount,
		Status:        "active",
	}
	switch {
	case link.RevokedAt.Valid:
		revoked := link.RevokedAt.Int64
		view.RevokedAt = &revoked
		view.Status = "revoked"
	case now >= link.ExpiresAt:
		view.Status = "expired"
	}
	return view
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func normalizeLoc(loc string) string {
	return "/" + strings.Trim(strings.ReplaceAll(loc, `\`, "/"), "/")
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
