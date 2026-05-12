package auth

import (
	"database/sql"
	"net/url"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/settings"
)

// OAuthCreds holds effective OAuth provider credentials. Each field is sourced
// from server_settings first, falling back to the env-loaded Config value.
// Building this struct centralizes the override layer so callbacks and the
// public auth-config response see the same effective values.
type OAuthCreds struct {
	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURI  string
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURI   string
}

const (
	keyDiscordClientID     = "oauth.discord.client_id"
	keyDiscordClientSecret = "oauth.discord.client_secret"
	keyDiscordRedirectURI  = "oauth.discord.redirect_uri"
	keyGoogleClientID      = "oauth.google.client_id"
	keyGoogleClientSecret  = "oauth.google.client_secret"
	keyGoogleRedirectURI   = "oauth.google.redirect_uri"
)

func ResolveCreds(cfg *config.Config, conn *sql.DB) OAuthCreds {
	creds := envCreds(cfg)
	overrides, err := settings.GetAll(conn)
	if err != nil {
		return creds
	}
	applyOver(&creds, overrides)
	return creds
}

func envCreds(cfg *config.Config) OAuthCreds {
	return OAuthCreds{
		DiscordClientID:     cfg.DiscordClientID,
		DiscordClientSecret: cfg.DiscordClientSecret,
		DiscordRedirectURI:  cfg.DiscordRedirectURI,
		GoogleClientID:      cfg.GoogleClientID,
		GoogleClientSecret:  cfg.GoogleClientSecret,
		GoogleRedirectURI:   cfg.GoogleRedirectURI,
	}
}

func applyOver(creds *OAuthCreds, overrides map[string]string) {
	targets := map[string]*string{
		keyDiscordClientID:     &creds.DiscordClientID,
		keyDiscordClientSecret: &creds.DiscordClientSecret,
		keyDiscordRedirectURI:  &creds.DiscordRedirectURI,
		keyGoogleClientID:      &creds.GoogleClientID,
		keyGoogleClientSecret:  &creds.GoogleClientSecret,
		keyGoogleRedirectURI:   &creds.GoogleRedirectURI,
	}
	for key, target := range targets {
		if value, ok := overrides[key]; ok && value != "" {
			*target = value
		}
	}
}

// discordOK reports whether enough credentials are configured to start a
// Discord OAuth flow. Discord uses implicit grant for sign-in, so client
// secret is not strictly required at the start of the flow — but we require
// it anyway so the provider stays in a consistent "configured" state.
func discordOK(c OAuthCreds) bool {
	return c.DiscordClientID != "" && c.DiscordRedirectURI != ""
}

func googleOK(c OAuthCreds) bool {
	return c.GoogleClientID != "" && c.GoogleClientSecret != "" && c.GoogleRedirectURI != ""
}

func discordURL(clientID, redirectURI string) string {
	q := url.Values{
		"client_id":     {clientID},
		"response_type": {"token"},
		"redirect_uri":  {redirectURI},
		"scope":         {"identify"},
	}
	return "https://discord.com/oauth2/authorize?" + q.Encode()
}

func googleURL(clientID, redirectURI string) string {
	q := url.Values{
		"response_type": {"code"},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"scope":         {"openid email profile"},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + q.Encode()
}

// ProviderView is the admin-facing safe shape for a single OAuth provider.
// The client secret is intentionally omitted — only its presence is reported.
type ProviderView struct {
	ClientID    string `json:"clientId"`
	RedirectURI string `json:"redirectUri"`
	HasSecret   bool   `json:"hasSecret"`
}

type OAuthCredsView struct {
	Discord ProviderView `json:"discord"`
	Google  ProviderView `json:"google"`
}

type OAuthCredsUpdate struct {
	DiscordClientID     string `json:"discordClientId"`
	DiscordClientSecret string `json:"discordClientSecret"`
	DiscordRedirectURI  string `json:"discordRedirectUri"`
	GoogleClientID      string `json:"googleClientId"`
	GoogleClientSecret  string `json:"googleClientSecret"`
	GoogleRedirectURI   string `json:"googleRedirectUri"`
}

func ViewCreds(cfg *config.Config, conn *sql.DB) OAuthCredsView {
	creds := ResolveCreds(cfg, conn)
	return OAuthCredsView{
		Discord: ProviderView{
			ClientID:    creds.DiscordClientID,
			RedirectURI: creds.DiscordRedirectURI,
			HasSecret:   creds.DiscordClientSecret != "",
		},
		Google: ProviderView{
			ClientID:    creds.GoogleClientID,
			RedirectURI: creds.GoogleRedirectURI,
			HasSecret:   creds.GoogleClientSecret != "",
		},
	}
}

// WriteCreds persists provider credentials to server_settings. Empty client_id
// or redirect_uri disables the provider on next AuthConfig fetch. Empty secret
// is treated as "preserve existing" — callers must send the literal secret only
// when changing it, never to confirm an unchanged value.
func WriteCreds(conn *sql.DB, update OAuthCredsUpdate) error {
	if err := writeFields(conn, update); err != nil {
		return err
	}
	return writeSecrets(conn, update)
}

func writeFields(conn *sql.DB, update OAuthCredsUpdate) error {
	pairs := map[string]string{
		keyDiscordClientID:    update.DiscordClientID,
		keyDiscordRedirectURI: update.DiscordRedirectURI,
		keyGoogleClientID:     update.GoogleClientID,
		keyGoogleRedirectURI:  update.GoogleRedirectURI,
	}
	for key, value := range pairs {
		if err := settings.Set(conn, key, value); err != nil {
			return err
		}
	}
	return nil
}

func writeSecrets(conn *sql.DB, update OAuthCredsUpdate) error {
	if update.DiscordClientSecret != "" {
		if err := settings.Set(conn, keyDiscordClientSecret, update.DiscordClientSecret); err != nil {
			return err
		}
	}
	if update.GoogleClientSecret != "" {
		return settings.Set(conn, keyGoogleClientSecret, update.GoogleClientSecret)
	}
	return nil
}
