package config

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

var knownAuthTypes = []AuthType{AuthTypeLocal, AuthTypeOAuth, AuthTypeBoth, AuthTypeOIDC}

// An unknown AUTH_TYPE starts the server with password sign-in open, so a value
// that was meant as oidc must stop the start instead.
func (c *Config) authTypeProblem() string {
	if slices.Contains(knownAuthTypes, c.AuthType) {
		return ""
	}
	if cleanAuthType(string(c.AuthType)) == string(AuthTypeOIDC) {
		return fmt.Sprintf("AUTH_TYPE=%q must be exactly oidc, with no space, quote, upper-case letter or comment", string(c.AuthType))
	}
	if c.OIDC.configured() {
		return fmt.Sprintf("AUTH_TYPE=%q is not one of local, oauth, both, oidc, but OIDC_* settings are present", string(c.AuthType))
	}
	return ""
}

func (c *Config) warnUnknownAuthType() {
	if !slices.Contains(knownAuthTypes, c.AuthType) {
		slog.Warn("AUTH_TYPE is not one of local, oauth, both, oidc", "value", string(c.AuthType))
	}
}

// docker run --env-file keeps quotes, trailing spaces and a comment after the value
func cleanAuthType(value string) string {
	value, _, _ = strings.Cut(value, "#")
	value = strings.Trim(strings.TrimSpace(value), `"'`)
	return strings.ToLower(strings.TrimSpace(value))
}
