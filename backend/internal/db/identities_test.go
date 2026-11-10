package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateIdentityUser_WritesUserIdentityAndIntents(t *testing.T) {
	conn := openEmptyDB(t)
	require.NoError(t, InitSchema(conn))

	id, err := CreateIdentityUser(conn, User{
		UserID: "sub-1", Username: "dave", GlobalName: "Dave", AuthType: "oidc", Intents: []string{"VIEW", "OPEN"},
	}, "oidc", "sub-1")
	require.NoError(t, err)

	user, err := GetUserByPrimaryKey(conn, id)
	require.NoError(t, err)
	require.Equal(t, "sub-1", user.UserID)
	require.Equal(t, "oidc", user.AuthType)
	require.ElementsMatch(t, []string{"VIEW", "OPEN"}, user.Intents)

	identity, err := GetIdentity(conn, "oidc", "sub-1")
	require.NoError(t, err)
	require.Equal(t, id, identity.UserID)
}

func TestCreateIdentityUser_LeavesNothingWhenUserIDIsTaken(t *testing.T) {
	conn := openEmptyDB(t)
	require.NoError(t, InitSchema(conn))
	_, err := SaveLocalUser(conn, "erin", "erin", "hash", "")
	require.NoError(t, err)

	_, err = CreateIdentityUser(conn, User{UserID: "erin", Username: "other", Intents: []string{"ADMIN"}}, "oidc", "erin")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "UNIQUE"), err.Error())

	identity, err := GetIdentity(conn, "oidc", "erin")
	require.NoError(t, err)
	require.Nil(t, identity)
	isAdmin, err := HasIntent(conn, "erin", "ADMIN")
	require.NoError(t, err)
	require.False(t, isAdmin)
}
