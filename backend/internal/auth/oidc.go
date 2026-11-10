package auth

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/JMC50/nas/internal/config"
)

const (
	oidcIdentity = "oidc"

	// 프런트엔드가 /auth/config에서 읽는 주소임
	oidcLoginPath  = "/server/auth/oidc/login"
	oidcLogoutPath = "/server/auth/oidc/logout"

	oidcResultPath    = "/oidcLogin"
	oidcSignedOutPath = "/localLogin?signedout=1"

	oidcLoginCookie   = "__Host-nas_oidc_login"
	oidcSessionCookie = "__Host-nas_oidc_session"
	oidcLoginTTL      = 10 * time.Minute
	oidcSessionTTL    = time.Minute
	oidcHTTPTimeout   = 10 * time.Second

	oidcFailed      = "failed"
	oidcDenied      = "denied"
	oidcConflict    = "conflict"
	oidcUnavailable = "unavailable"
)

// OIDCHandlers는 AUTH_TYPE=oidc일 때만 있는 경로를 처리합니다.
type OIDCHandlers struct {
	Config *config.Config
	DB     *sql.DB

	client *http.Client
	spent  spentTokens

	mu     sync.Mutex
	issuer *oidcIssuer
}

type oidcIssuer struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	// 토큰 주소가 받는 클라이언트 인증 방식을 한 번 알아낸 뒤 기억하므로 요청마다 새로 만들지 않음
	oauth *oauth2.Config
}

func NewOIDCHandlers(cfg *config.Config, conn *sql.DB) *OIDCHandlers {
	return &OIDCHandlers{
		Config: cfg,
		DB:     conn,
		client: &http.Client{Timeout: oidcHTTPTimeout},
		spent:  spentTokens{until: map[string]time.Time{}},
	}
}

// 발급자가 멈춰 있어도 서버가 뜨도록 처음 쓸 때 읽음. 실패하면 다음 요청에서 다시 읽음
func (h *OIDCHandlers) discover() (*oidcIssuer, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.issuer != nil {
		return h.issuer, nil
	}
	provider, err := oidc.NewProvider(oidc.ClientContext(context.Background(), h.client), h.Config.OIDC.Issuer)
	if err != nil {
		slog.Error("oidc discovery failed", "err", err)
		return nil, err
	}
	h.issuer = &oidcIssuer{
		provider: provider,
		verifier: provider.Verifier(&oidc.Config{ClientID: h.Config.OIDC.ClientID}),
		oauth: &oauth2.Config{
			ClientID:     h.Config.OIDC.ClientID,
			ClientSecret: h.Config.OIDC.ClientSecret,
			RedirectURL:  h.Config.OIDC.RedirectURI,
			Endpoint:     provider.Endpoint(),
			Scopes:       h.Config.OIDC.Scopes,
		},
	}
	return h.issuer, nil
}

// Login은 인가 코드 흐름을 시작합니다. state, nonce, PKCE verifier는 콜백까지 이 브라우저의 쿠키에 둡니다.
func (h *OIDCHandlers) Login(w http.ResponseWriter, r *http.Request) {
	issuer, err := h.discover()
	if err != nil {
		finishOIDC(w, r, oidcUnavailable)
		return
	}
	state, stateErr := makeNonce()
	nonce, nonceErr := makeNonce()
	if stateErr != nil || nonceErr != nil {
		finishOIDC(w, r, oidcFailed)
		return
	}
	verifier := oauth2.GenerateVerifier()
	setOIDCCookie(w, oidcLoginCookie, strings.Join([]string{state, nonce, verifier}, "."), oidcLoginTTL)

	options := []oauth2.AuthCodeOption{oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier)}
	if hint := h.Config.OIDC.IDPHint; hint != "" {
		options = append(options, oauth2.SetAuthURLParam("idp_hint", hint))
	}
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, issuer.oauth.AuthCodeURL(state, options...), http.StatusFound)
}

type oidcPending struct {
	state    string
	nonce    string
	verifier string
}

func readPending(r *http.Request) (oidcPending, bool) {
	cookie, err := r.Cookie(oidcLoginCookie)
	if err != nil {
		return oidcPending{}, false
	}
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 3 || slices.Contains(parts, "") {
		return oidcPending{}, false
	}
	return oidcPending{state: parts[0], nonce: parts[1], verifier: parts[2]}, true
}

// Callback은 흐름을 끝냅니다. NAS 토큰은 주소에 싣지 않고 수명이 짧은 쿠키로 프런트엔드에 넘깁니다.
func (h *OIDCHandlers) Callback(w http.ResponseWriter, r *http.Request) {
	pending, bound := readPending(r)
	clearOIDCCookie(w, oidcLoginCookie)
	query := r.URL.Query()
	if problem := callbackProblem(query, pending, bound); problem != "" {
		finishOIDC(w, r, problem)
		return
	}
	token, problem := h.signIn(r.Context(), query.Get("code"), pending)
	if problem == "" {
		setOIDCCookie(w, oidcSessionCookie, token, oidcSessionTTL)
	}
	finishOIDC(w, r, problem)
}

func callbackProblem(query url.Values, pending oidcPending, bound bool) string {
	switch {
	case !bound || !sameText(query.Get("state"), pending.state):
		return oidcFailed
	case query.Get("error") == "access_denied":
		return oidcDenied
	case query.Get("code") == "":
		return oidcFailed
	}
	return ""
}

func (h *OIDCHandlers) signIn(ctx context.Context, code string, pending oidcPending) (token, problem string) {
	idToken, problem := h.redeem(ctx, code, pending)
	if problem != "" {
		return "", problem
	}
	intents := h.firstIntents(idToken)
	if len(intents) == 0 {
		slog.Warn("oidc sign-in refused: no NAS group", "sub", idToken.Subject)
		return "", oidcDenied
	}
	user, problem := h.userFor(idToken, intents)
	if problem != "" {
		return "", problem
	}
	token, err := IssueToken(user.UserID, h.Config.PrivateKey)
	if err != nil {
		slog.Error("oidc sign-in failed: token not issued", "err", err)
		return "", oidcFailed
	}
	return token, ""
}

func (h *OIDCHandlers) redeem(ctx context.Context, code string, pending oidcPending) (*oidc.IDToken, string) {
	issuer, err := h.discover()
	if err != nil {
		return nil, oidcUnavailable
	}
	ctx = oidc.ClientContext(ctx, h.client)
	token, err := issuer.oauth.Exchange(ctx, code, oauth2.VerifierOption(pending.verifier))
	if err != nil {
		slog.Warn("oidc code exchange failed", "err", err)
		return nil, oidcFailed
	}
	rawIDToken, _ := token.Extra("id_token").(string)
	idToken, err := issuer.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		slog.Warn("oidc id_token refused", "err", err)
		return nil, oidcFailed
	}
	if idToken.Subject == "" || !sameText(idToken.Nonce, pending.nonce) {
		slog.Warn("oidc id_token refused: nonce or sub does not match")
		return nil, oidcFailed
	}
	return idToken, ""
}

func finishOIDC(w http.ResponseWriter, r *http.Request, problem string) {
	target := oidcResultPath
	if problem != "" {
		target += "?error=" + problem
	}
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, target, http.StatusFound)
}

func setOIDCCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearOIDCCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func sameText(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
