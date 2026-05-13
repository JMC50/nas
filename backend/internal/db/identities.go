package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// Identity links a NAS user to an external sign-in source. Provider is one
// of "local", "discord", "google"; external_id is the user's identifier
// inside that provider (local userId, Discord snowflake, Google sub).
// Primary key is (provider, external_id) so the same external account can
// only ever map to one NAS user.
type Identity struct {
	UserID     int64
	Provider   string
	ExternalID string
}

func GetIdentity(conn *sql.DB, provider, externalID string) (*Identity, error) {
	row := conn.QueryRow(`
		SELECT user_id FROM user_identities
		WHERE provider = ? AND external_id = ?`, provider, externalID)
	var userID int64
	if err := row.Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan identity: %w", err)
	}
	return &Identity{UserID: userID, Provider: provider, ExternalID: externalID}, nil
}

func AddIdentity(conn *sql.DB, userID int64, provider, externalID string) error {
	_, err := conn.Exec(`
		INSERT INTO user_identities (user_id, provider, external_id)
		VALUES (?, ?, ?)`, userID, provider, externalID)
	if err != nil {
		return fmt.Errorf("insert identity: %w", err)
	}
	return nil
}

func ListIdentity(conn *sql.DB, userID int64) ([]Identity, error) {
	rows, err := conn.Query(`
		SELECT provider, external_id FROM user_identities
		WHERE user_id = ? ORDER BY provider`, userID)
	if err != nil {
		return nil, fmt.Errorf("query identities: %w", err)
	}
	defer rows.Close()
	out := []Identity{}
	for rows.Next() {
		entry := Identity{UserID: userID}
		if err := rows.Scan(&entry.Provider, &entry.ExternalID); err != nil {
			return nil, fmt.Errorf("scan identity row: %w", err)
		}
		out = append(out, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate identities: %w", err)
	}
	return out, nil
}

func DropIdentity(conn *sql.DB, userID int64, provider string) error {
	_, err := conn.Exec(`
		DELETE FROM user_identities
		WHERE user_id = ? AND provider = ?`, userID, provider)
	if err != nil {
		return fmt.Errorf("delete identity: %w", err)
	}
	return nil
}

// Backfill seeds user_identities from existing users for the local provider
// only. OAuth users are deliberately skipped — the wire-stored auth_type does
// not distinguish Discord from Google for legacy rows, so those identities are
// re-created lazily by the OAuth callbacks on first sign-in after this rolls
// out. Safe to call on every startup: INSERT OR IGNORE keeps it idempotent.
func Backfill(conn *sql.DB) error {
	_, err := conn.Exec(`
		INSERT OR IGNORE INTO user_identities (user_id, provider, external_id)
		SELECT id, 'local', userId FROM users
		WHERE auth_type = 'local'`)
	if err != nil {
		return fmt.Errorf("backfill identities: %w", err)
	}
	return nil
}
