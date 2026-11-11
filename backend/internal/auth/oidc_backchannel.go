package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/JMC50/nas/internal/db"
)

const (
	logoutTokenMaxAge = 5 * time.Minute
	logoutClockSkew   = 30 * time.Second
	// go-jose가 4096비트보다 큰 RSA 키를 받지 않아서 서명은 683자를 넘지 않음. 그 키로 서명한 logout_token이 1.3 KiB쯤임
	logoutFormLimit = 8 << 10
)

var errIssuerUnavailable = errors.New("issuer unavailable")

// BackchannelLogout은 발급자가 보낸 logout_token을 확인하고, 그 사용자가 이미 받은 NAS 토큰을 모두 무효로 만듭니다.
func (h *OIDCHandlers) BackchannelLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	r.Body = http.MaxBytesReader(w, r.Body, logoutFormLimit)
	token, err := h.readLogoutToken(r.Context(), r.PostFormValue("logout_token"))
	if errors.Is(err, errIssuerUnavailable) {
		// 토큰이 틀린 것이 아니라 확인하지 못한 것이므로 일시 장애로 답함
		slog.Error("oidc logout_token not checked", "err", err)
		http.Error(w, "issuer unavailable", http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		slog.Warn("oidc logout_token refused", "err", err)
		http.Error(w, "invalid logout_token", http.StatusBadRequest)
		return
	}
	if err := h.endSessions(token.Subject); err != nil {
		// 발급자가 같은 토큰으로 다시 보낼 수 있게 사용 기록을 되돌림
		h.spent.release(token.TokenID)
		slog.Error("oidc back-channel logout failed", "err", err)
		http.Error(w, "logout failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *OIDCHandlers) readLogoutToken(ctx context.Context, raw string) (*oidc.LogoutToken, error) {
	if raw == "" {
		return nil, errors.New("logout_token missing")
	}
	issuer, err := h.discover()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errIssuerUnavailable, err)
	}
	token, err := issuer.verifier.VerifyLogout(oidc.ClientContext(ctx, h.client), raw)
	if err != nil && keysUnavailable(err) {
		return nil, fmt.Errorf("%w: %v", errIssuerUnavailable, err)
	}
	if err != nil {
		return nil, err
	}
	// NAS는 발급자의 세션 ID를 기록하지 않으므로 sub가 없으면 누구인지 알 수 없음
	if token.Subject == "" {
		return nil, errors.New("logout token has no sub")
	}
	age := time.Since(token.IssuedAt)
	if token.IssuedAt.IsZero() || age > logoutTokenMaxAge || age < -logoutClockSkew {
		return nil, errors.New("logout token iat is missing or outside the accepted window")
	}
	if !h.spent.claim(token.TokenID, token.IssuedAt.Add(logoutTokenMaxAge+logoutClockSkew)) {
		return nil, errors.New("logout token was already used")
	}
	return token, nil
}

// go-oidc는 JWKS를 읽지 못한 오류를 타입으로 알리지 않고 문구로만 알림. 요청이 정하는 alg 값이 오류 문구에 들어갈 수 있어서 앞부분만 비교함
func keysUnavailable(err error) bool {
	return strings.HasPrefix(err.Error(), "failed to verify signature: fetching keys")
}

func (h *OIDCHandlers) endSessions(subject string) error {
	identity, err := db.GetIdentity(h.DB, oidcIdentity, subject)
	if err != nil || identity == nil {
		return err
	}
	// 같은 밀리초에 먼저 나간 토큰도 무효가 되게 1을 더함
	return db.SetTokensValidAfter(h.DB, identity.UserID, time.Now().UnixMilli()+1)
}

// 받은 logout_token의 jti를 그 토큰이 다시 통과할 수 있는 시간 동안 기억함
type spentTokens struct {
	mu    sync.Mutex
	until map[string]time.Time
}

func (s *spentTokens) claim(id string, until time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for seen, expires := range s.until {
		if now.After(expires) {
			delete(s.until, seen)
		}
	}
	if _, used := s.until[id]; used {
		return false
	}
	s.until[id] = until
	return true
}

func (s *spentTokens) release(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.until, id)
}
