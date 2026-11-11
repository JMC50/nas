package config

import (
	"net"
	"net/url"
	"os"
	"slices"
	"strings"
)

// OIDCConfig is used only when AUTH_TYPE=oidc.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	IDPHint      string
	Scopes       []string
	GroupsClaim  string
	MemberGroup  string
	AdminGroup   string
	LogoutURL    string
}

func loadOIDC() OIDCConfig {
	return OIDCConfig{
		Issuer:       getEnv("OIDC_ISSUER", ""),
		ClientID:     getSecret("OIDC_CLIENT_ID", ""),
		ClientSecret: getSecret("OIDC_CLIENT_SECRET", ""),
		RedirectURI:  getEnv("OIDC_REDIRECT_URI", ""),
		IDPHint:      getEnv("OIDC_IDP_HINT", ""),
		Scopes:       strings.Fields(getEnvNonEmpty("OIDC_SCOPES", "openid profile email groups")),
		GroupsClaim:  getEnvNonEmpty("OIDC_GROUPS_CLAIM", "groups"),
		MemberGroup:  getEnvNonEmpty("OIDC_MEMBER_GROUP", "nas_member"),
		AdminGroup:   getEnvNonEmpty("OIDC_ADMIN_GROUP", "nas_admin"),
		LogoutURL:    getEnv("OIDC_LOGOUT_URL", ""),
	}
}

// problems names settings only. It never includes a value, because the list goes to the startup log.
func (o OIDCConfig) problems() []string {
	var problems []string
	if missing := o.missing(); len(missing) > 0 {
		problems = append(problems, "AUTH_TYPE=oidc requires "+strings.Join(missing, ", "))
	}
	addresses := [][2]string{
		{"OIDC_ISSUER", o.Issuer},
		{"OIDC_REDIRECT_URI", o.RedirectURI},
		{"OIDC_LOGOUT_URL", o.LogoutURL},
	}
	for _, setting := range addresses {
		if setting[1] != "" && !isHTTPURL(setting[1]) {
			problems = append(problems, setting[0]+" must be an absolute http or https URL")
		}
	}
	// discovery와 JWKS를 이 주소에서 받으므로 평문 http면 오가는 길에서 서명 키를 바꿔치기할 수 있음
	if isHTTPURL(o.Issuer) && !isHTTPSOrLoopback(o.Issuer) {
		problems = append(problems, "OIDC_ISSUER must use https, except on a loopback host")
	}
	if !slices.Contains(o.Scopes, "openid") {
		problems = append(problems, "OIDC_SCOPES must include openid")
	}
	return problems
}

func (o OIDCConfig) required() [][2]string {
	return [][2]string{
		{"OIDC_ISSUER", o.Issuer},
		{"OIDC_CLIENT_ID", o.ClientID},
		{"OIDC_CLIENT_SECRET", o.ClientSecret},
		{"OIDC_REDIRECT_URI", o.RedirectURI},
	}
}

func (o OIDCConfig) configured() bool {
	return len(o.missing()) < len(o.required())
}

func (o OIDCConfig) missing() []string {
	var missing []string
	for _, setting := range o.required() {
		if setting[1] == "" {
			missing = append(missing, setting[0])
		}
	}
	return missing
}

func isHTTPURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func isHTTPSOrLoopback(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	if parsed.Scheme == "https" {
		return true
	}
	host := parsed.Hostname()
	address := net.ParseIP(host)
	return strings.EqualFold(host, "localhost") || (address != nil && address.IsLoopback())
}

func getEnvNonEmpty(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
