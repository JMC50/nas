package auth

import (
	"database/sql"

	"github.com/JMC50/nas/internal/db"
)

type LinkStatus string

const (
	LinkAdded    LinkStatus = "linked"
	LinkPresent  LinkStatus = "already linked"
	LinkNoUser   LinkStatus = "no NAS user"
	LinkConflict LinkStatus = "conflict"
)

type LinkResult struct {
	Status   LinkStatus
	Username string
	Detail   string
}

// LinkOIDCIdentity는 예전 로그인 수단(fromProvider, oldID)의 사용자에게 새 발급자의 sub를 oidc 신원으로 더합니다.
// 같은 값으로 다시 불러도 결과가 같습니다. dryRun이면 아무것도 쓰지 않습니다.
func LinkOIDCIdentity(conn *sql.DB, fromProvider, oldID, subject string, dryRun bool) (LinkResult, error) {
	user, err := identityUser(conn, fromProvider, oldID)
	if err != nil {
		return LinkResult{}, err
	}
	if user == nil {
		return LinkResult{Status: LinkNoUser}, nil
	}
	result := LinkResult{Username: user.Username}
	result.Status, result.Detail, err = linkState(conn, user.ID, subject)
	if err != nil || result.Status != LinkAdded || dryRun {
		return result, err
	}
	return result, db.AddIdentity(conn, user.ID, oidcIdentity, subject)
}

// LinkAdded는 아직 연결이 없고 더해도 되는 상태임
func linkState(conn *sql.DB, userID int64, subject string) (LinkStatus, string, error) {
	owner, err := db.GetIdentity(conn, oidcIdentity, subject)
	if err != nil {
		return "", "", err
	}
	if owner != nil && owner.UserID == userID {
		return LinkPresent, "", nil
	}
	if owner != nil {
		return LinkConflict, "the new subject belongs to another NAS user", nil
	}
	identities, err := db.ListIdentity(conn, userID)
	if err != nil {
		return "", "", err
	}
	if hasIdentity(identities, oidcIdentity) {
		return LinkConflict, "the NAS user has another OIDC subject", nil
	}
	return LinkAdded, "", nil
}

func identityUser(conn *sql.DB, provider, externalID string) (*db.User, error) {
	identity, err := db.GetIdentity(conn, provider, externalID)
	if err != nil {
		return nil, err
	}
	if identity != nil {
		return db.GetUserByPrimaryKey(conn, identity.UserID)
	}
	// 신원 표가 생기기 전에 가입한 OAuth 사용자는 외부 ID가 users.userId에만 있음
	user, err := db.GetUser(conn, externalID)
	if err != nil || user == nil || provider == "local" || user.AuthType == "local" || user.AuthType == oidcIdentity {
		return nil, err
	}
	return user, nil
}
