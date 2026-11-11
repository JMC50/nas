package integration

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/server"
)

const (
	oidcClientID     = "nas"
	oidcClientSecret = "issuer-test-secret"
	oidcRedirectURI  = "https://nas.example/auth/oidc/callback"
	oidcLoginCookie  = "__Host-nas_oidc_login"
	oidcTokenCookie  = "__Host-nas_oidc_session"
	logoutEvent      = "http://schemas.openid.net/event/backchannel-logout"
)

var (
	issuerRSAKey = sync.OnceValue(func() *rsa.PrivateKey {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
		return key
	})
	// 발급자의 JWKS에 없는 키임. 위조한 토큰을 만들 때 씀
	strangerRSAKey = sync.OnceValue(func() *rsa.PrivateKey {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
		return key
	})
	issuerEdKey = sync.OnceValue(func() ed25519.PrivateKey {
		_, key, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			panic(err)
		}
		return key
	})
)

// 시험 안에서만 도는 OIDC 발급자임. discovery, JWKS, 토큰 주소를 실제 HTTP로 내줌
type fakeIssuer struct {
	t          *testing.T
	server     *httptest.Server
	endSession bool

	// 발급자 이름이 /로 끝나는 발급자를 흉내 냄. 주소들은 그대로임
	trailingSlash bool

	mu         sync.Mutex
	codes      map[string]grantedCode
	tokenCalls int
}

type grantedCode struct {
	challenge string
	idToken   string
}

func newFakeIssuer(t *testing.T, endSession bool) *fakeIssuer {
	t.Helper()
	issuer := &fakeIssuer{t: t, endSession: endSession, codes: map[string]grantedCode{}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/openid-configuration", issuer.discovery)
	mux.HandleFunc("GET /jwks", issuer.jwks)
	mux.HandleFunc("POST /token", issuer.token)
	issuer.server = httptest.NewServer(mux)
	t.Cleanup(issuer.server.Close)
	return issuer
}

func (f *fakeIssuer) url() string { return f.server.URL }

func (f *fakeIssuer) name() string {
	if f.trailingSlash {
		return f.url() + "/"
	}
	return f.url()
}

func (f *fakeIssuer) discovery(w http.ResponseWriter, r *http.Request) {
	document := map[string]any{
		"issuer":                                f.name(),
		"authorization_endpoint":                f.url() + "/authorize",
		"token_endpoint":                        f.url() + "/token",
		"jwks_uri":                              f.url() + "/jwks",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256", "EdDSA"},
	}
	if f.endSession {
		document["end_session_endpoint"] = f.url() + "/logout?from=nas"
	}
	writeTestJSON(w, http.StatusOK, document)
}

func (f *fakeIssuer) jwks(w http.ResponseWriter, r *http.Request) {
	public := issuerRSAKey().PublicKey
	writeTestJSON(w, http.StatusOK, map[string]any{"keys": []map[string]string{
		{
			"kty": "RSA", "use": "sig", "alg": "RS256", "kid": "rsa-1",
			"n": base64.RawURLEncoding.EncodeToString(public.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(public.E)).Bytes()),
		},
		{
			"kty": "OKP", "use": "sig", "alg": "EdDSA", "kid": "ed-1", "crv": "Ed25519",
			"x": base64.RawURLEncoding.EncodeToString(issuerEdKey().Public().(ed25519.PublicKey)),
		},
	}})
}

// 실제 발급자처럼 클라이언트 비밀, redirect_uri, PKCE verifier가 맞을 때만 id_token을 내줌
func (f *fakeIssuer) token(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tokenCalls++
	if err := r.ParseForm(); err != nil {
		writeTestJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		clientID, clientSecret = r.PostForm.Get("client_id"), r.PostForm.Get("client_secret")
	}
	if clientID != oidcClientID || clientSecret != oidcClientSecret {
		writeTestJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_client"})
		return
	}
	code := r.PostForm.Get("code")
	granted, known := f.codes[code]
	delete(f.codes, code)
	challenge := sha256.Sum256([]byte(r.PostForm.Get("code_verifier")))
	if !known || r.PostForm.Get("grant_type") != "authorization_code" ||
		r.PostForm.Get("redirect_uri") != oidcRedirectURI ||
		base64.RawURLEncoding.EncodeToString(challenge[:]) != granted.challenge {
		writeTestJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	writeTestJSON(w, http.StatusOK, map[string]any{
		"access_token": "access-" + code, "token_type": "Bearer", "expires_in": 300, "id_token": granted.idToken,
	})
}

func (f *fakeIssuer) tokenRequests() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.tokenCalls
}

// 사용자가 발급자에서 로그인을 마친 것과 같음. 콜백에 실을 인가 코드를 돌려줌
func (f *fakeIssuer) grant(authURL *url.URL, idToken string) string {
	code := randomHex(f.t)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.codes[code] = grantedCode{challenge: authURL.Query().Get("code_challenge"), idToken: idToken}
	return code
}

func (f *fakeIssuer) idClaims(authURL *url.URL, subject string, groups any) jwt.MapClaims {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":                f.name(),
		"aud":                oidcClientID,
		"sub":                subject,
		"iat":                now.Unix(),
		"exp":                now.Add(5 * time.Minute).Unix(),
		"nonce":              authURL.Query().Get("nonce"),
		"name":               "Person " + subject,
		"preferred_username": "user-" + subject,
		"email":              subject + "@example.com",
	}
	if groups != nil {
		claims["groups"] = groups
	}
	return claims
}

func (f *fakeIssuer) logoutClaims(subject string) jwt.MapClaims {
	now := time.Now()
	return jwt.MapClaims{
		"iss":    f.name(),
		"aud":    oidcClientID,
		"sub":    subject,
		"iat":    now.Unix(),
		"exp":    now.Add(2 * time.Minute).Unix(),
		"jti":    randomHex(f.t),
		"events": map[string]any{logoutEvent: map[string]any{}},
	}
}

func (f *fakeIssuer) sign(claims jwt.MapClaims) string {
	return signJWT(f.t, jwt.SigningMethodRS256, issuerRSAKey(), "rsa-1", claims)
}

func signJWT(t *testing.T, method jwt.SigningMethod, key any, keyID string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = keyID
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func randomHex(t *testing.T) string {
	t.Helper()
	raw := make([]byte, 16)
	_, err := rand.Read(raw)
	require.NoError(t, err)
	return hex.EncodeToString(raw)
}

func writeTestJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

// AUTH_TYPE=oidc로 만든 실제 라우터와 가짜 발급자를 묶음
type oidcFixture struct {
	t      *testing.T
	router http.Handler
	conn   *sql.DB
	issuer *fakeIssuer
}

func newOIDCFixture(t *testing.T, issuer *fakeIssuer, adjust ...func(*config.OIDCConfig)) *oidcFixture {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	require.NoError(t, db.InitSchema(conn))

	cfg := &config.Config{
		CorsOrigin:           "*",
		PrivateKey:           testPrivateKey,
		AdminPassword:        "admin-pass",
		AuthType:             config.AuthTypeOIDC,
		PasswordRequirements: config.PasswordRequirements{MinLength: 4},
		OIDC: config.OIDCConfig{
			Issuer:       issuer.name(),
			ClientID:     oidcClientID,
			ClientSecret: oidcClientSecret,
			RedirectURI:  oidcRedirectURI,
			IDPHint:      "upstream-google",
			Scopes:       []string{"openid", "profile", "email", "groups"},
			GroupsClaim:  "groups",
			MemberGroup:  "nas_member",
			AdminGroup:   "nas_admin",
		},
	}
	for _, change := range adjust {
		change(&cfg.OIDC)
	}
	return &oidcFixture{t: t, router: server.NewRouter(cfg, conn), conn: conn, issuer: issuer}
}

func (f *oidcFixture) request(method, target string, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for _, cookie := range cookies {
		req.AddCookie(&http.Cookie{Name: cookie.Name, Value: cookie.Value})
	}
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

func setCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// 로그인 시작 경로를 부르고, 발급자로 가는 주소와 브라우저가 받은 쿠키를 돌려줌
func (f *oidcFixture) start() (*url.URL, *http.Cookie) {
	f.t.Helper()
	w := f.request("GET", "/auth/oidc/login", "")
	require.Equal(f.t, http.StatusFound, w.Code, w.Body.String())
	authURL, err := url.Parse(w.Header().Get("Location"))
	require.NoError(f.t, err)
	cookie := setCookie(w, oidcLoginCookie)
	require.NotNil(f.t, cookie)
	return authURL, cookie
}

func (f *oidcFixture) callback(code, state string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	query := url.Values{"code": {code}, "state": {state}}
	return f.request("GET", "/auth/oidc/callback?"+query.Encode(), "", cookies...)
}

// 발급자가 edit이 만든 id_token으로 로그인을 끝낸 뒤의 콜백 응답을 돌려줌
func (f *oidcFixture) finish(edit func(authURL *url.URL) string) *httptest.ResponseRecorder {
	f.t.Helper()
	authURL, cookie := f.start()
	code := f.issuer.grant(authURL, edit(authURL))
	return f.callback(code, authURL.Query().Get("state"), cookie)
}

func (f *oidcFixture) finishAs(subject string, groups any) *httptest.ResponseRecorder {
	return f.finish(func(authURL *url.URL) string {
		return f.issuer.sign(f.issuer.idClaims(authURL, subject, groups))
	})
}

type oidcSession struct {
	UserID     string `json:"userId"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name"`
	Token      string `json:"token"`
}

// 로그인 흐름 전체를 지나 프런트엔드가 받는 값을 돌려줌
func (f *oidcFixture) signIn(subject string, groups any) oidcSession {
	f.t.Helper()
	w := f.finishAs(subject, groups)
	require.Equal(f.t, http.StatusFound, w.Code)
	require.Equal(f.t, "/oidcLogin", w.Header().Get("Location"))
	handoff := setCookie(w, oidcTokenCookie)
	require.NotNil(f.t, handoff)

	answer := f.request("POST", "/server/auth/oidc/session", "", handoff)
	require.Equal(f.t, http.StatusOK, answer.Code, answer.Body.String())
	var session oidcSession
	require.NoError(f.t, json.Unmarshal(answer.Body.Bytes(), &session))
	require.NotEmpty(f.t, session.Token)
	return session
}

func (f *oidcFixture) backchannel(logoutToken string) *httptest.ResponseRecorder {
	return f.request("POST", "/auth/oidc/backchannel", url.Values{"logout_token": {logoutToken}}.Encode())
}

func (f *oidcFixture) userCount() int {
	f.t.Helper()
	var count int
	require.NoError(f.t, f.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count))
	return count
}

func (f *oidcFixture) tokenStatus(token string) int {
	return getAs(f.router, "/server/checkIntent?intent=VIEW", token).Code
}
