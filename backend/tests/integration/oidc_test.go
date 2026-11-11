package integration

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
)

var memberIntents = []string{"VIEW", "OPEN", "DOWNLOAD", "UPLOAD", "COPY", "RENAME"}

func TestOIDCLogin_SendsCodeFlowWithPKCEStateAndNonce(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	authURL, cookie := f.start()

	require.Equal(t, f.issuer.url()+"/authorize", authURL.Scheme+"://"+authURL.Host+authURL.Path)
	query := authURL.Query()
	require.Equal(t, "code", query.Get("response_type"))
	require.Equal(t, oidcClientID, query.Get("client_id"))
	require.Equal(t, oidcRedirectURI, query.Get("redirect_uri"))
	require.Equal(t, "openid profile email groups", query.Get("scope"))
	require.Equal(t, "S256", query.Get("code_challenge_method"))
	require.Len(t, query.Get("code_challenge"), 43)
	require.Equal(t, "upstream-google", query.Get("idp_hint"))
	require.GreaterOrEqual(t, len(query.Get("state")), 32)
	require.GreaterOrEqual(t, len(query.Get("nonce")), 32)
	require.NotEqual(t, query.Get("state"), query.Get("nonce"))
	require.NotContains(t, authURL.String(), oidcClientSecret)

	other, _ := f.start()
	require.NotEqual(t, query.Get("state"), other.Query().Get("state"))
	require.NotEqual(t, query.Get("code_challenge"), other.Query().Get("code_challenge"))

	parts := strings.Split(cookie.Value, ".")
	require.Len(t, parts, 3)
	require.Equal(t, query.Get("state"), parts[0])
}

func TestOIDCCookies_AreSecureHttpOnlyAndShortLived(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	_, login := f.start()
	handoff := setCookie(f.finishAs("sub-cookie", []string{"nas_member"}), oidcTokenCookie)
	require.NotNil(t, handoff)

	for cookie, maxAge := range map[*http.Cookie]int{login: 600, handoff: 60} {
		require.True(t, strings.HasPrefix(cookie.Name, "__Host-"), cookie.Name)
		require.True(t, cookie.Secure, cookie.Name)
		require.True(t, cookie.HttpOnly, cookie.Name)
		require.Equal(t, http.SameSiteLaxMode, cookie.SameSite, cookie.Name)
		require.Equal(t, "/", cookie.Path, cookie.Name)
		require.Empty(t, cookie.Domain, cookie.Name)
		require.Equal(t, maxAge, cookie.MaxAge, cookie.Name)
	}
}

func TestOIDCCallback_CreatesMemberAndHandsOverToken(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	w := f.finishAs("sub-member", []string{"other_group", "nas_member"})

	require.Equal(t, http.StatusFound, w.Code)
	require.Equal(t, "/oidcLogin", w.Header().Get("Location"))
	require.NotContains(t, w.Header().Get("Location"), "token")
	cleared := setCookie(w, oidcLoginCookie)
	require.NotNil(t, cleared)
	require.Negative(t, cleared.MaxAge)
	handoff := setCookie(w, oidcTokenCookie)
	require.NotNil(t, handoff)

	answer := f.request("POST", "/server/auth/oidc/session", "", handoff)
	require.Equal(t, http.StatusOK, answer.Code, answer.Body.String())
	require.Equal(t, "no-store", answer.Header().Get("Cache-Control"))
	require.Negative(t, setCookie(answer, oidcTokenCookie).MaxAge)
	var session oidcSession
	require.NoError(t, json.Unmarshal(answer.Body.Bytes(), &session))
	require.Equal(t, "sub-member", session.UserID)
	require.Equal(t, "user-sub-member", session.Username)
	require.Equal(t, "Person sub-member", session.GlobalName)

	user, err := db.GetUser(f.conn, "sub-member")
	require.NoError(t, err)
	require.Equal(t, "oidc", user.AuthType)
	require.ElementsMatch(t, memberIntents, user.Intents)
	identity, err := db.GetIdentity(f.conn, "oidc", "sub-member")
	require.NoError(t, err)
	require.Equal(t, user.ID, identity.UserID)

	require.Equal(t, http.StatusOK, f.tokenStatus(session.Token))
	require.Equal(t, http.StatusForbidden, getAs(f.router, "/server/getAllUsers", session.Token).Code)
}

func TestOIDCSession_NeedsTheHandoffCookie(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	require.Equal(t, http.StatusUnauthorized, f.request("POST", "/auth/oidc/session", "").Code)

	forged := &http.Cookie{Name: oidcTokenCookie, Value: "not-a-token"}
	w := f.request("POST", "/auth/oidc/session", "", forged)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.NotContains(t, w.Body.String(), "token\"")
	require.Equal(t, http.StatusMethodNotAllowed, f.request("GET", "/auth/oidc/session", "").Code)
}

func TestOIDCCallback_AdminGroupGetsAdmin(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	admin := f.signIn("sub-admin", []string{"nas_admin"})
	both := f.signIn("sub-both", []string{"nas_member", "nas_admin"})

	user, err := db.GetUser(f.conn, "sub-admin")
	require.NoError(t, err)
	require.Equal(t, []string{"ADMIN"}, user.Intents)
	user, err = db.GetUser(f.conn, "sub-both")
	require.NoError(t, err)
	require.ElementsMatch(t, append([]string{"ADMIN"}, memberIntents...), user.Intents)

	require.Equal(t, http.StatusOK, getAs(f.router, "/server/getAllUsers", admin.Token).Code)
	require.Equal(t, http.StatusOK, getAs(f.router, "/server/getAllUsers", both.Token).Code)
}

func TestOIDCCallback_ExistingUserKeepsAccountAndIntents(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	id, err := db.SaveOAuthUser(f.conn, "google-123", "Old Name", "Old Name", "홍길동")
	require.NoError(t, err)
	require.NoError(t, db.AddIdentity(f.conn, id, "google", "google-123"))
	require.NoError(t, db.AddIdentity(f.conn, id, "oidc", "sub-linked"))
	require.NoError(t, db.ToggleIntent(f.conn, "google-123", "VIEW"))

	session := f.signIn("sub-linked", []string{"nas_member", "nas_admin"})
	require.Equal(t, "google-123", session.UserID)
	require.Equal(t, "Old Name", session.Username)
	require.Equal(t, 1, f.userCount())

	user, err := db.GetUserByPrimaryKey(f.conn, id)
	require.NoError(t, err)
	require.Equal(t, []string{"VIEW"}, user.Intents, "groups set intents only when the user is created")
	require.Equal(t, "홍길동", user.KrName)

	again := f.signIn("sub-linked", []string{"nas_member"})
	require.Equal(t, "google-123", again.UserID)
	require.Equal(t, 1, f.userCount())
}

func TestOIDCCallback_RefusesUserWithoutGroupAndCreatesNoAccount(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	for name, groups := range map[string]any{
		"no claim":         nil,
		"empty list":       []string{},
		"other groups":     []string{"git_member", "nas_members", "nas"},
		"string not array": "nas_member",
		"substring":        []string{"xnas_memberx"},
	} {
		w := f.finishAs("sub-outsider", groups)
		require.Equal(t, http.StatusFound, w.Code, name)
		require.Equal(t, "/oidcLogin?error=denied", w.Header().Get("Location"), name)
		require.Nil(t, setCookie(w, oidcTokenCookie), name)
	}
	require.Zero(t, f.userCount())
	identity, err := db.GetIdentity(f.conn, "oidc", "sub-outsider")
	require.NoError(t, err)
	require.Nil(t, identity)
}

func TestOIDCCallback_RefusesExistingUserWhoLostTheGroup(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	f.signIn("sub-leaver", []string{"nas_member"})

	w := f.finishAs("sub-leaver", []string{})
	require.Equal(t, "/oidcLogin?error=denied", w.Header().Get("Location"))
	require.Nil(t, setCookie(w, oidcTokenCookie))
}

func TestOIDCCallback_RejectsStateMismatchAndMissingCookie(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	authURL, cookie := f.start()
	code := f.issuer.grant(authURL, f.issuer.sign(f.issuer.idClaims(authURL, "sub-state", []string{"nas_member"})))
	state := authURL.Query().Get("state")
	_, otherCookie := f.start()

	for name, w := range map[string]*httptest.ResponseRecorder{
		"wrong state":           f.callback(code, "forged-state", cookie),
		"empty state":           f.callback(code, "", cookie),
		"no cookie":             f.callback(code, state),
		"cookie of other login": f.callback(code, state, otherCookie),
		"malformed cookie":      f.callback(code, state, &http.Cookie{Name: oidcLoginCookie, Value: state}),
	} {
		require.Equal(t, http.StatusFound, w.Code, name)
		require.Equal(t, "/oidcLogin?error=failed", w.Header().Get("Location"), name)
		require.Nil(t, setCookie(w, oidcTokenCookie), name)
	}
	require.Zero(t, f.issuer.tokenRequests(), "the code must not be redeemed before the state matches")
	require.Zero(t, f.userCount())
}

func TestOIDCCallback_RejectsWrongPKCEVerifier(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	authURL, cookie := f.start()
	code := f.issuer.grant(authURL, f.issuer.sign(f.issuer.idClaims(authURL, "sub-pkce", []string{"nas_member"})))
	parts := strings.Split(cookie.Value, ".")
	tampered := &http.Cookie{Name: oidcLoginCookie, Value: parts[0] + "." + parts[1] + ".another-verifier-another-verifier-another-ver"}

	w := f.callback(code, authURL.Query().Get("state"), tampered)
	require.Equal(t, "/oidcLogin?error=failed", w.Header().Get("Location"))
	require.Positive(t, f.issuer.tokenRequests(), "the issuer is the one that refuses the verifier")
	require.Zero(t, f.userCount())
}

func TestOIDCCallback_RejectsInvalidIDTokens(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	base := func(authURL *url.URL) jwt.MapClaims {
		return f.issuer.idClaims(authURL, "sub-invalid", []string{"nas_member", "nas_admin"})
	}
	cases := map[string]func(authURL *url.URL) string{
		"nonce mismatch": func(authURL *url.URL) string {
			claims := base(authURL)
			claims["nonce"] = "another-nonce"
			return f.issuer.sign(claims)
		},
		"nonce missing": func(authURL *url.URL) string {
			claims := base(authURL)
			delete(claims, "nonce")
			return f.issuer.sign(claims)
		},
		"wrong issuer": func(authURL *url.URL) string {
			claims := base(authURL)
			claims["iss"] = "https://evil.example"
			return f.issuer.sign(claims)
		},
		"wrong audience": func(authURL *url.URL) string {
			claims := base(authURL)
			claims["aud"] = "another-client"
			return f.issuer.sign(claims)
		},
		"expired": func(authURL *url.URL) string {
			claims := base(authURL)
			claims["exp"] = time.Now().Add(-time.Minute).Unix()
			return f.issuer.sign(claims)
		},
		"signed by another key": func(authURL *url.URL) string {
			return signJWT(t, jwt.SigningMethodRS256, strangerRSAKey(), "rsa-1", base(authURL))
		},
		"signed with the client secret": func(authURL *url.URL) string {
			return signJWT(t, jwt.SigningMethodHS256, []byte(oidcClientSecret), "rsa-1", base(authURL))
		},
		"unsigned": func(authURL *url.URL) string {
			return signJWT(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, "rsa-1", base(authURL))
		},
		"no subject": func(authURL *url.URL) string {
			claims := base(authURL)
			delete(claims, "sub")
			return f.issuer.sign(claims)
		},
		"not a token": func(*url.URL) string { return "" },
	}
	for name, idToken := range cases {
		w := f.finish(idToken)
		require.Equal(t, http.StatusFound, w.Code, name)
		require.Equal(t, "/oidcLogin?error=failed", w.Header().Get("Location"), name)
		require.Nil(t, setCookie(w, oidcTokenCookie), name)
	}
	require.Zero(t, f.userCount())
}

func TestOIDCCallback_AcceptsEdDSASignedIDToken(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	w := f.finish(func(authURL *url.URL) string {
		claims := f.issuer.idClaims(authURL, "sub-eddsa", []string{"nas_member"})
		return signJWT(t, jwt.SigningMethodEdDSA, issuerEdKey(), "ed-1", claims)
	})
	require.Equal(t, "/oidcLogin", w.Header().Get("Location"))
	require.Equal(t, 1, f.userCount())
}

func TestOIDCCallback_AuthorizationCodeWorksOnce(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	authURL, cookie := f.start()
	code := f.issuer.grant(authURL, f.issuer.sign(f.issuer.idClaims(authURL, "sub-once", []string{"nas_member"})))
	state := authURL.Query().Get("state")

	require.Equal(t, "/oidcLogin", f.callback(code, state, cookie).Header().Get("Location"))
	replay := f.callback(code, state, cookie)
	require.Equal(t, "/oidcLogin?error=failed", replay.Header().Get("Location"))
	require.Nil(t, setCookie(replay, oidcTokenCookie))
}

func TestOIDCCallback_ReportsIssuerDenialAndErrors(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	authURL, cookie := f.start()
	state := url.QueryEscape(authURL.Query().Get("state"))

	denied := f.request("GET", "/auth/oidc/callback?error=access_denied&state="+state, "", cookie)
	require.Equal(t, "/oidcLogin?error=denied", denied.Header().Get("Location"))

	other := f.request("GET", "/auth/oidc/callback?error=server_error&error_description=%3Cscript%3E&state="+state, "", cookie)
	require.Equal(t, "/oidcLogin?error=failed", other.Header().Get("Location"))

	unbound := f.request("GET", "/auth/oidc/callback?error=access_denied", "")
	require.Equal(t, "/oidcLogin?error=failed", unbound.Header().Get("Location"))
}

func TestOIDC_RedirectTargetsCannotBeChangedByTheRequest(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	hostile := url.Values{
		"return":       {"https://evil.example/"},
		"next":         {"//evil.example"},
		"redirect_uri": {"https://evil.example/callback"},
		"redirect":     {"https://evil.example/"},
	}.Encode()

	w := f.request("GET", "/auth/oidc/login?"+hostile, "")
	require.Equal(t, http.StatusFound, w.Code)
	authURL, err := url.Parse(w.Header().Get("Location"))
	require.NoError(t, err)
	require.Equal(t, f.issuer.url()+"/authorize", authURL.Scheme+"://"+authURL.Host+authURL.Path)
	require.Equal(t, []string{oidcRedirectURI}, authURL.Query()["redirect_uri"])
	require.NotContains(t, w.Header().Get("Location"), "evil.example")

	cookie := setCookie(w, oidcLoginCookie)
	code := f.issuer.grant(authURL, f.issuer.sign(f.issuer.idClaims(authURL, "sub-redirect", []string{"nas_member"})))
	target := "/auth/oidc/callback?code=" + code + "&state=" + url.QueryEscape(authURL.Query().Get("state")) + "&" + hostile
	done := f.request("GET", target, "", cookie)
	require.Equal(t, "/oidcLogin", done.Header().Get("Location"))

	failed := f.request("GET", "/auth/oidc/callback?"+hostile, "")
	require.Equal(t, "/oidcLogin?error=failed", failed.Header().Get("Location"))
}

func TestOIDCLogin_ReportsUnreachableIssuer(t *testing.T) {
	issuer := newFakeIssuer(t, true)
	f := newOIDCFixture(t, issuer, func(oidc *config.OIDCConfig) { oidc.Issuer = "http://127.0.0.1:1/unreachable" })

	w := f.request("GET", "/auth/oidc/login", "")
	require.Equal(t, http.StatusFound, w.Code)
	require.Equal(t, "/oidcLogin?error=unavailable", w.Header().Get("Location"))
	require.Nil(t, setCookie(w, oidcLoginCookie))
	require.Equal(t, http.StatusServiceUnavailable, f.backchannel(issuer.sign(issuer.logoutClaims("sub-any"))).Code)
}

func TestOIDCBackchannel_AnswersUnavailableWhileTheKeysCannotBeRead(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	f.issuer.setKeysDown(true)
	logoutToken := f.issuer.sign(f.issuer.logoutClaims("sub-later"))

	w := f.backchannel(logoutToken)
	require.Equal(t, http.StatusServiceUnavailable, w.Code, w.Body.String())
	require.Equal(t, http.StatusBadRequest, f.backchannel("abc.def.ghi").Code, "a malformed token stays a bad request")

	f.issuer.setKeysDown(false)
	require.Equal(t, http.StatusOK, f.backchannel(logoutToken).Code, "the same token is accepted when the issuer sends it again")
}

// go-oidc의 오류 문구에 요청이 정한 alg가 들어가도 503이 되지 않아야 함
func TestOIDCBackchannel_AlgorithmTextCannotTurnRefusalIntoUnavailable(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	encode := func(text string) string { return base64.RawURLEncoding.EncodeToString([]byte(text)) }
	forged := encode(`{"alg":"failed to verify signature: fetching keys","kid":"rsa-1"}`) + "." + encode(`{"sub":"sub-any"}`) + "." + encode("signature")

	require.Equal(t, http.StatusBadRequest, f.backchannel(forged).Code)
}

func TestOIDCMode_ClosesOtherSignInAndAdminPasswordRoutes(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	session := f.signIn("sub-closed", []string{"nas_member"})
	user, err := db.GetUser(f.conn, "sub-closed")
	require.NoError(t, err)
	require.NoError(t, db.AddIdentity(f.conn, user.ID, "google", "google-closed"))

	closed := []struct{ method, path, body string }{
		{"POST", "/auth/register", `{"userId":"mallory","username":"mallory","password":"strong-pass"}`},
		{"POST", "/auth/login", `{"userId":"mallory","password":"strong-pass"}`},
		{"POST", "/auth/change-password?token=" + session.Token, `{"oldPassword":"a","newPassword":"strong-pass"}`},
		{"GET", "/auth/discord/callback?access_token=anything", ""},
		{"GET", "/auth/google/callback?code=anything", ""},
		{"POST", "/auth/discord/register", `{"access_token":"anything","krname":"x"}`},
		{"POST", "/auth/google/register", `{"access_token":"anything","krname":"x"}`},
		{"POST", "/auth/link/start?token=" + session.Token, `{"provider":"google"}`},
		{"POST", "/auth/link/discord/complete?token=" + session.Token, `{"state":"link.x","access_token":"x"}`},
		{"POST", "/auth/link/google/complete?token=" + session.Token, `{"state":"link.x","code":"x"}`},
		{"DELETE", "/auth/identities/oidc?token=" + session.Token, ""},
		{"DELETE", "/auth/identities/google?token=" + session.Token, ""},
		{"POST", "/requestAdminIntent?token=" + session.Token, `{"pwd":"admin-pass"}`},
	}
	for _, route := range closed {
		for _, prefix := range []string{"", "/server"} {
			req := httptest.NewRequest(route.method, prefix+route.path, strings.NewReader(route.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			f.router.ServeHTTP(w, req)
			require.Equal(t, http.StatusForbidden, w.Code, route.method+" "+prefix+route.path)
			require.Contains(t, w.Body.String(), "disabled", route.path)
			require.NotContains(t, w.Body.String(), `"token"`, route.path)
		}
	}
	require.Equal(t, 1, f.userCount())
	identities, err := db.ListIdentity(f.conn, user.ID)
	require.NoError(t, err)
	require.Len(t, identities, 2)
	user, err = db.GetUser(f.conn, "sub-closed")
	require.NoError(t, err)
	require.ElementsMatch(t, memberIntents, user.Intents, "the admin password must not give ADMIN")
	require.Equal(t, http.StatusForbidden, getAs(f.router, "/server/getAllUsers", session.Token).Code)
}

func TestOIDCMode_AuthConfigReportsOnlyOIDC(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	_, err := f.conn.Exec(`INSERT INTO server_settings (key, value) VALUES
		('oauth.google.client_id', 'id'), ('oauth.google.client_secret', 'secret'), ('oauth.google.redirect_uri', 'https://nas.example/googleLogin'),
		('oauth.discord.client_id', 'id'), ('oauth.discord.redirect_uri', 'https://nas.example/login')`)
	require.NoError(t, err)

	w := f.request("GET", "/server/auth/config", "")
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, "oidc", body["authType"])
	require.Equal(t, true, body["oidcEnabled"])
	require.Equal(t, "/server/auth/oidc/login", body["oidcLoginUrl"])
	require.Equal(t, "/server/auth/oidc/logout", body["oidcLogoutUrl"])
	for _, key := range []string{"localAuthEnabled", "discordEnabled", "googleEnabled", "oauthEnabled"} {
		require.Equal(t, false, body[key], key)
	}
	require.Empty(t, body["discordLoginUrl"])
	require.Empty(t, body["googleLoginUrl"])
	require.NotContains(t, w.Body.String(), oidcClientSecret)
}

func TestOIDCBackchannel_EndsTheTokensOfThatUser(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	target := f.signIn("sub-target", []string{"nas_member"})
	bystander := f.signIn("sub-bystander", []string{"nas_member"})
	require.Equal(t, http.StatusOK, f.tokenStatus(target.Token))

	w := f.backchannel(f.issuer.sign(f.issuer.logoutClaims("sub-target")))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	require.Equal(t, http.StatusUnauthorized, f.tokenStatus(target.Token))
	require.Equal(t, http.StatusUnauthorized, getAs(f.router, "/server/readFolder?loc=", target.Token).Code)
	require.Equal(t, http.StatusOK, f.tokenStatus(bystander.Token))

	// 무효 시각은 밀리초 단위라서 다음 로그인이 그 뒤가 되게 함
	time.Sleep(5 * time.Millisecond)
	again := f.signIn("sub-target", []string{"nas_member"})
	require.Equal(t, http.StatusOK, f.tokenStatus(again.Token))
	require.Equal(t, http.StatusUnauthorized, f.tokenStatus(target.Token))
}

func TestOIDCBackchannel_AcceptsUnknownSubjectWithoutChange(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	session := f.signIn("sub-known", []string{"nas_member"})

	require.Equal(t, http.StatusOK, f.backchannel(f.issuer.sign(f.issuer.logoutClaims("sub-never-seen"))).Code)
	require.Equal(t, http.StatusOK, f.tokenStatus(session.Token))
	require.Equal(t, 1, f.userCount())
}

func TestOIDCBackchannel_RejectsForgedAndMalformedLogoutTokens(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	session := f.signIn("sub-victim", []string{"nas_member"})
	authURL, _ := f.start()
	valid := func() jwt.MapClaims { return f.issuer.logoutClaims("sub-victim") }
	change := func(key string, value any) string {
		claims := valid()
		if value == nil {
			delete(claims, key)
		} else {
			claims[key] = value
		}
		return f.issuer.sign(claims)
	}

	cases := map[string]string{
		"signed by another key":         signJWT(t, jwt.SigningMethodRS256, strangerRSAKey(), "rsa-1", valid()),
		"signed with the client secret": signJWT(t, jwt.SigningMethodHS256, []byte(oidcClientSecret), "rsa-1", valid()),
		"unsigned":                      signJWT(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, "rsa-1", valid()),
		"wrong issuer":                  change("iss", "https://evil.example"),
		"wrong audience":                change("aud", "another-client"),
		"no events":                     change("events", nil),
		"other event":                   change("events", map[string]any{"http://example.com/event": map[string]any{}}),
		"has nonce":                     change("nonce", "abc"),
		"no jti":                        change("jti", nil),
		"no exp":                        change("exp", nil),
		"expired":                       change("exp", time.Now().Add(-time.Minute).Unix()),
		"no iat":                        change("iat", nil),
		"old iat":                       change("iat", time.Now().Add(-10*time.Minute).Unix()),
		"future iat":                    change("iat", time.Now().Add(5*time.Minute).Unix()),
		"no sub":                        change("sub", nil),
		"larger than 8 KiB":             change("padding", strings.Repeat("x", 8<<10)),
		"id_token of the user":          f.issuer.sign(f.issuer.idClaims(authURL, "sub-victim", []string{"nas_member"})),
		"empty":                         "",
		"not a token":                   "abc.def.ghi",
	}
	for name, logoutToken := range cases {
		w := f.backchannel(logoutToken)
		require.Equal(t, http.StatusBadRequest, w.Code, name)
	}
	require.Equal(t, http.StatusBadRequest, f.request("POST", "/auth/oidc/backchannel", "").Code)
	require.Equal(t, http.StatusMethodNotAllowed, f.request("GET", "/auth/oidc/backchannel", "").Code)
	require.Equal(t, http.StatusOK, f.tokenStatus(session.Token), "a refused logout_token must not end a session")
}

func TestOIDCBackchannel_RejectsReplayedLogoutToken(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	f.signIn("sub-replay", []string{"nas_member"})
	logoutToken := f.issuer.sign(f.issuer.logoutClaims("sub-replay"))
	require.Equal(t, http.StatusOK, f.backchannel(logoutToken).Code)

	time.Sleep(5 * time.Millisecond)
	again := f.signIn("sub-replay", []string{"nas_member"})
	require.Equal(t, http.StatusBadRequest, f.backchannel(logoutToken).Code)
	require.Equal(t, http.StatusOK, f.tokenStatus(again.Token), "a replayed logout_token must not end the new session")
}

func TestOIDCMode_RejectsTokenOfRemovedUser(t *testing.T) {
	f := newOIDCFixture(t, newFakeIssuer(t, true))
	session := f.signIn("sub-removed", []string{"nas_member"})
	require.NoError(t, db.DeleteUser(f.conn, "sub-removed"))

	require.Equal(t, http.StatusUnauthorized, f.tokenStatus(session.Token))
	require.Equal(t, http.StatusUnauthorized, getAs(f.router, "/server/getActivityLog", session.Token).Code)
}

func TestOIDCLogout_RedirectsToTheIssuerEndSession(t *testing.T) {
	issuer := newFakeIssuer(t, true)
	f := newOIDCFixture(t, issuer)
	w := f.request("GET", "/server/auth/oidc/logout", "")
	require.Equal(t, http.StatusFound, w.Code)
	target, err := url.Parse(w.Header().Get("Location"))
	require.NoError(t, err)
	require.Equal(t, issuer.url()+"/logout", target.Scheme+"://"+target.Host+target.Path)
	require.Equal(t, oidcClientID, target.Query().Get("client_id"))
	require.Equal(t, "nas", target.Query().Get("from"))

	fixed := newOIDCFixture(t, issuer, func(oidc *config.OIDCConfig) { oidc.LogoutURL = "https://login.example/logout" })
	require.Equal(t, "https://login.example/logout", fixed.request("GET", "/auth/oidc/logout", "").Header().Get("Location"))

	plain := newOIDCFixture(t, newFakeIssuer(t, false))
	require.Equal(t, "/localLogin?signedout=1", plain.request("GET", "/auth/oidc/logout", "").Header().Get("Location"))
}

func TestOIDC_WorksWithIssuerNameThatEndsWithSlash(t *testing.T) {
	issuer := newFakeIssuer(t, true)
	issuer.trailingSlash = true
	f := newOIDCFixture(t, issuer)

	session := f.signIn("sub-slash", []string{"nas_member"})
	require.Equal(t, http.StatusOK, f.tokenStatus(session.Token))
	require.Equal(t, http.StatusOK, f.backchannel(issuer.sign(issuer.logoutClaims("sub-slash"))).Code)
	require.Equal(t, http.StatusUnauthorized, f.tokenStatus(session.Token))

	// 설정에서 끝의 /만 빠져도 발급자가 알린 이름과 달라서 로그인을 시작하지 않음
	mismatched := newOIDCFixture(t, issuer, func(oidc *config.OIDCConfig) { oidc.Issuer = issuer.url() })
	w := mismatched.request("GET", "/auth/oidc/login", "")
	require.Equal(t, "/oidcLogin?error=unavailable", w.Header().Get("Location"))
	require.Nil(t, setCookie(w, oidcLoginCookie))
}
