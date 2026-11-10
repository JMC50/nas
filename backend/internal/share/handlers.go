package share

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
)

type Handlers struct {
	Config *config.Config
	DB     *sql.DB
	limits *limiter
}

func NewHandlers(cfg *config.Config, conn *sql.DB) *Handlers {
	return &Handlers{Config: cfg, DB: conn, limits: newLimiter()}
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

// Create는 파일이나 폴더의 공개 링크를 만들고, 토큰 원문은 이 응답에서만 돌려줍니다.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	rel, lifetime, err := decodeCreate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	creator, _, err := h.caller(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	info, err := h.statShared(rel)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	token, err := newToken()
	if err != nil {
		http.Error(w, "token generation failed", http.StatusInternalServerError)
		return
	}
	now := time.Now()
	loc, name := splitTarget(rel)
	link := db.ShareLink{
		TokenHash: hashToken(token),
		UserID:    creator.ID,
		Loc:       loc,
		Name:      name,
		IsFolder:  info.IsDir(),
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

// decodeCreate는 요청 값을 검사하고 데이터 폴더 기준 상대 경로로 바꿈. 저장도 이 경로로 해서 검사한 값과 저장한 값이 같음
func decodeCreate(r *http.Request) (string, time.Duration, error) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", 0, errors.New("invalid body")
	}
	lifetime, ok := expiryChoices[req.ExpiresIn]
	if !ok {
		return "", 0, errors.New("expiresIn must be one of 1h, 1d, 7d, 30d")
	}
	rel, err := targetPath(req.Loc, req.Name)
	if err != nil {
		return "", 0, err
	}
	return rel, lifetime, nil
}

func (h *Handlers) statShared(rel string) (os.FileInfo, error) {
	root, err := os.OpenRoot(h.Config.NASDataDir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return statTarget(root, rel)
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

// Revoke는 링크를 만든 사용자나 관리자만 링크를 끊게 하고, 남의 링크에는 없는 링크처럼 404로 답합니다.
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

func newToken() (string, error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
