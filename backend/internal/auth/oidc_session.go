package auth

import (
	"net/http"
	"net/url"

	"github.com/JMC50/nas/internal/db"
)

// Session은 Callback이 쿠키에 둔 토큰을 프런트엔드에 내주고 그 쿠키를 지웁니다.
func (h *OIDCHandlers) Session(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(oidcSessionCookie)
	clearOIDCCookie(w, oidcSessionCookie)
	w.Header().Set("Cache-Control", "no-store")
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, authResponse{Message: "No sign-in is waiting. Sign in again."})
		return
	}
	var user *db.User
	claims, err := Authenticate(h.Config, h.DB, cookie.Value)
	if err == nil {
		user, err = db.GetUser(h.DB, claims.UserID)
	}
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, authResponse{Message: "The sign-in is no longer valid. Sign in again."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"userId":      user.UserID,
		"username":    user.Username,
		"krname":      user.KrName,
		"global_name": orString(user.GlobalName, user.Username),
		"token":       cookie.Value,
	})
}

// Logout은 브라우저를 발급자의 세션을 끝내는 주소로 보냅니다. NAS 토큰은 프런트엔드가 먼저 지웁니다.
func (h *OIDCHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, h.logoutURL(), http.StatusFound)
}

func (h *OIDCHandlers) logoutURL() string {
	if h.Config.OIDC.LogoutURL != "" {
		return h.Config.OIDC.LogoutURL
	}
	issuer, err := h.discover()
	if err != nil {
		return oidcSignedOutPath
	}
	var metadata struct {
		EndSession string `json:"end_session_endpoint"`
	}
	if err := issuer.provider.Claims(&metadata); err != nil || metadata.EndSession == "" {
		return oidcSignedOutPath
	}
	endSession, err := url.Parse(metadata.EndSession)
	if err != nil {
		return oidcSignedOutPath
	}
	query := endSession.Query()
	query.Set("client_id", h.Config.OIDC.ClientID)
	endSession.RawQuery = query.Encode()
	return endSession.String()
}

// RefuseOtherSignIn은 AUTH_TYPE=oidc에서 닫는 로그인 경로의 응답입니다.
func RefuseOtherSignIn(http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusForbidden, authResponse{
			Message: "This sign-in method is disabled. Sign in through single sign-on.",
		})
	})
}
