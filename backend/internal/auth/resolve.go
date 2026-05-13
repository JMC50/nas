package auth

import (
	"github.com/JMC50/nas/internal/db"
)

// resolveLocal looks up a user by their local sign-in identifier. Prefers the
// identity table; if missing (e.g., legacy account whose backfill row was lost)
// falls back to users.userId and lazily seeds the identity row so subsequent
// sign-ins use the fast path.
func (h *Handlers) resolveLocal(userID string) (*db.User, error) {
	identity, err := db.GetIdentity(h.DB, "local", userID)
	if err != nil {
		return nil, err
	}
	if identity != nil {
		return db.GetUserByPrimaryKey(h.DB, identity.UserID)
	}
	user, err := db.GetUser(h.DB, userID)
	if err != nil || user == nil {
		return user, err
	}
	if user.AuthType != "local" {
		return nil, nil
	}
	if err := db.AddIdentity(h.DB, user.ID, "local", userID); err != nil {
		return nil, err
	}
	return user, nil
}

// resolveOAuth mirrors resolveLocal for OAuth callbacks. Pre-update users
// stored the provider snowflake/sub as users.userId; this lazily migrates them
// into user_identities the first time they sign in after the schema change.
func (h *Handlers) resolveOAuth(provider, externalID string) (*db.User, error) {
	identity, err := db.GetIdentity(h.DB, provider, externalID)
	if err != nil {
		return nil, err
	}
	if identity != nil {
		return db.GetUserByPrimaryKey(h.DB, identity.UserID)
	}
	user, err := db.GetUser(h.DB, externalID)
	if err != nil || user == nil {
		return user, err
	}
	if user.AuthType == "local" {
		return nil, nil
	}
	if err := db.AddIdentity(h.DB, user.ID, provider, externalID); err != nil {
		return nil, err
	}
	return user, nil
}
