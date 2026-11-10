package auth

import (
	"cmp"
	"encoding/json"
	"log/slog"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/JMC50/nas/internal/db"
)

var oidcMemberIntents = []Intent{IntentView, IntentOpen, IntentDownload, IntentUpload, IntentCopy, IntentRename}

// 빈 결과면 로그인을 거부함. 권한은 사용자를 처음 만들 때만 쓰고 이미 있는 사용자의 권한은 바꾸지 않음
func (h *OIDCHandlers) firstIntents(idToken *oidc.IDToken) []string {
	var claims map[string]json.RawMessage
	if err := idToken.Claims(&claims); err != nil {
		return nil
	}
	// 문자열 하나로 온 값은 부분 문자열이 맞을 수 있어서 배열만 받음
	var groups []string
	if err := json.Unmarshal(claims[h.Config.OIDC.GroupsClaim], &groups); err != nil {
		return nil
	}
	var intents []string
	if slices.Contains(groups, h.Config.OIDC.MemberGroup) {
		for _, intent := range oidcMemberIntents {
			intents = append(intents, string(intent))
		}
	}
	if slices.Contains(groups, h.Config.OIDC.AdminGroup) {
		intents = append(intents, string(IntentAdmin))
	}
	return intents
}

func (h *OIDCHandlers) userFor(idToken *oidc.IDToken, intents []string) (*db.User, string) {
	user, err := h.findUser(idToken.Subject)
	if err == nil && user == nil {
		user, err = h.createUser(idToken, intents)
	}
	if err == nil {
		return user, ""
	}
	// 같은 사람의 첫 로그인 두 번이 겹치면 늦은 쪽은 먼저 만들어진 사용자를 씀
	if existing, findErr := h.findUser(idToken.Subject); findErr == nil && existing != nil {
		return existing, ""
	}
	slog.Error("oidc user lookup or creation failed", "err", err)
	if strings.Contains(err.Error(), "UNIQUE") {
		return nil, oidcConflict
	}
	return nil, oidcFailed
}

func (h *OIDCHandlers) findUser(subject string) (*db.User, error) {
	identity, err := db.GetIdentity(h.DB, oidcIdentity, subject)
	if err != nil || identity == nil {
		return nil, err
	}
	return db.GetUserByPrimaryKey(h.DB, identity.UserID)
}

func (h *OIDCHandlers) createUser(idToken *oidc.IDToken, intents []string) (*db.User, error) {
	var profile struct {
		Username string `json:"preferred_username"`
		Name     string `json:"name"`
		Email    string `json:"email"`
	}
	if err := idToken.Claims(&profile); err != nil {
		return nil, err
	}
	username := cmp.Or(profile.Username, profile.Email, profile.Name, idToken.Subject)
	id, err := db.CreateIdentityUser(h.DB, db.User{
		UserID:     idToken.Subject,
		Username:   username,
		GlobalName: cmp.Or(profile.Name, username),
		AuthType:   oidcIdentity,
		Intents:    intents,
	}, oidcIdentity, idToken.Subject)
	if err != nil {
		return nil, err
	}
	return db.GetUserByPrimaryKey(h.DB, id)
}
