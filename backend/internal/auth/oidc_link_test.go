package auth

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/db"
)

func oidcSubjects(t *testing.T, conn *sql.DB, userID int64) []string {
	t.Helper()
	identities, err := db.ListIdentity(conn, userID)
	require.NoError(t, err)
	subjects := []string{}
	for _, identity := range identities {
		if identity.Provider == "oidc" {
			subjects = append(subjects, identity.ExternalID)
		}
	}
	return subjects
}

func TestLinkOIDCIdentity_AddsIdentityToGoogleUser(t *testing.T) {
	conn := newSchemaDB(t)
	id, err := db.SaveOAuthUser(conn, "google-111", "Alice", "Alice", "앨리스")
	require.NoError(t, err)
	require.NoError(t, db.AddIdentity(conn, id, "google", "google-111"))
	require.NoError(t, db.ToggleIntent(conn, "google-111", "UPLOAD"))

	result, err := LinkOIDCIdentity(conn, "google", "google-111", "sub-aaa", false)
	require.NoError(t, err)
	require.Equal(t, LinkAdded, result.Status)
	require.Equal(t, "Alice", result.Username)
	require.Equal(t, []string{"sub-aaa"}, oidcSubjects(t, conn, id))

	user, err := db.GetUserByPrimaryKey(conn, id)
	require.NoError(t, err)
	require.Equal(t, "google-111", user.UserID)
	require.Equal(t, []string{"UPLOAD"}, user.Intents)
}

func TestLinkOIDCIdentity_RunsAgainWithoutChange(t *testing.T) {
	conn := newSchemaDB(t)
	id, err := db.SaveOAuthUser(conn, "google-222", "Bob", "Bob", "")
	require.NoError(t, err)
	require.NoError(t, db.AddIdentity(conn, id, "google", "google-222"))

	for _, want := range []LinkStatus{LinkAdded, LinkPresent, LinkPresent} {
		result, err := LinkOIDCIdentity(conn, "google", "google-222", "sub-bbb", false)
		require.NoError(t, err)
		require.Equal(t, want, result.Status)
	}
	require.Equal(t, []string{"sub-bbb"}, oidcSubjects(t, conn, id))
}

func TestLinkOIDCIdentity_DryRunWritesNothing(t *testing.T) {
	conn := newSchemaDB(t)
	id, err := db.SaveOAuthUser(conn, "google-333", "Carol", "Carol", "")
	require.NoError(t, err)
	require.NoError(t, db.AddIdentity(conn, id, "google", "google-333"))

	result, err := LinkOIDCIdentity(conn, "google", "google-333", "sub-ccc", true)
	require.NoError(t, err)
	require.Equal(t, LinkAdded, result.Status)
	require.Empty(t, oidcSubjects(t, conn, id))
}

func TestLinkOIDCIdentity_FindsUserFromBeforeTheIdentityTable(t *testing.T) {
	conn := newSchemaDB(t)
	id, err := db.SaveOAuthUser(conn, "google-444", "Dave", "Dave", "")
	require.NoError(t, err)

	result, err := LinkOIDCIdentity(conn, "google", "google-444", "sub-ddd", false)
	require.NoError(t, err)
	require.Equal(t, LinkAdded, result.Status)
	require.Equal(t, []string{"sub-ddd"}, oidcSubjects(t, conn, id))
}

func TestLinkOIDCIdentity_ReportsUnknownUser(t *testing.T) {
	conn := newSchemaDB(t)
	_, err := db.SaveLocalUser(conn, "google-555", "local user with the same text", "hash", "")
	require.NoError(t, err)

	for _, oldID := range []string{"google-000", "google-555"} {
		result, err := LinkOIDCIdentity(conn, "google", oldID, "sub-eee", false)
		require.NoError(t, err)
		require.Equal(t, LinkNoUser, result.Status, oldID)
	}
	identity, err := db.GetIdentity(conn, "oidc", "sub-eee")
	require.NoError(t, err)
	require.Nil(t, identity)
}

func TestLinkOIDCIdentity_ReportsConflicts(t *testing.T) {
	conn := newSchemaDB(t)
	first, err := db.SaveOAuthUser(conn, "google-666", "Erin", "Erin", "")
	require.NoError(t, err)
	require.NoError(t, db.AddIdentity(conn, first, "google", "google-666"))
	second, err := db.SaveOAuthUser(conn, "google-777", "Frank", "Frank", "")
	require.NoError(t, err)
	require.NoError(t, db.AddIdentity(conn, second, "google", "google-777"))
	_, err = LinkOIDCIdentity(conn, "google", "google-666", "sub-fff", false)
	require.NoError(t, err)

	taken, err := LinkOIDCIdentity(conn, "google", "google-777", "sub-fff", false)
	require.NoError(t, err)
	require.Equal(t, LinkConflict, taken.Status)
	require.Empty(t, oidcSubjects(t, conn, second))

	other, err := LinkOIDCIdentity(conn, "google", "google-666", "sub-ggg", false)
	require.NoError(t, err)
	require.Equal(t, LinkConflict, other.Status)
	require.Equal(t, []string{"sub-fff"}, oidcSubjects(t, conn, first))
}

func TestLinkOIDCIdentity_LinksLocalUser(t *testing.T) {
	conn := newSchemaDB(t)
	id, err := db.SaveLocalUser(conn, "grace", "Grace", "hash", "")
	require.NoError(t, err)
	require.NoError(t, db.AddIdentity(conn, id, "local", "grace"))

	result, err := LinkOIDCIdentity(conn, "local", "grace", "sub-hhh", false)
	require.NoError(t, err)
	require.Equal(t, LinkAdded, result.Status)
	require.Equal(t, []string{"sub-hhh"}, oidcSubjects(t, conn, id))
}
