package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// ShareLink는 공개 다운로드 링크 한 개입니다. 토큰 원문은 저장하지 않고 해시만 저장합니다.
type ShareLink struct {
	ID            int64
	TokenHash     string
	UserID        int64
	CreatorUserID string
	CreatorName   string
	Loc           string
	Name          string
	IsFolder      bool
	CreatedAt     int64
	ExpiresAt     int64
	RevokedAt     sql.NullInt64
	DownloadCount int64
}

const shareLinkColumns = `
	share_links.id, share_links.token_hash, share_links.user_id,
	COALESCE(users.userId, ''),
	COALESCE(NULLIF(users.krname, ''), NULLIF(users.global_name, ''), users.username, ''),
	share_links.loc, share_links.name, share_links.is_folder,
	share_links.created_at, share_links.expires_at, share_links.revoked_at, share_links.download_count`

func scanShareLink(row interface{ Scan(...any) error }) (*ShareLink, error) {
	link := &ShareLink{}
	var isFolder int64
	err := row.Scan(&link.ID, &link.TokenHash, &link.UserID, &link.CreatorUserID, &link.CreatorName,
		&link.Loc, &link.Name, &isFolder, &link.CreatedAt, &link.ExpiresAt, &link.RevokedAt, &link.DownloadCount)
	if err != nil {
		return nil, err
	}
	link.IsFolder = isFolder != 0
	return link, nil
}

func CreateShareLink(conn *sql.DB, link ShareLink) (int64, error) {
	isFolder := 0
	if link.IsFolder {
		isFolder = 1
	}
	result, err := conn.Exec(`
		INSERT INTO share_links (token_hash, user_id, loc, name, is_folder, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		link.TokenHash, link.UserID, link.Loc, link.Name, isFolder, link.CreatedAt, link.ExpiresAt)
	if err != nil {
		return 0, fmt.Errorf("insert share link: %w", err)
	}
	return result.LastInsertId()
}

// GetShareLinkByHash는 링크가 없으면 nil과 nil 오류를 돌려줍니다.
func GetShareLinkByHash(conn *sql.DB, tokenHash string) (*ShareLink, error) {
	return getShareLink(conn, "share_links.token_hash = ?", tokenHash)
}

// GetShareLink는 링크가 없으면 nil과 nil 오류를 돌려줍니다.
func GetShareLink(conn *sql.DB, id int64) (*ShareLink, error) {
	return getShareLink(conn, "share_links.id = ?", id)
}

func getShareLink(conn *sql.DB, where string, arg any) (*ShareLink, error) {
	row := conn.QueryRow(`SELECT `+shareLinkColumns+`
		FROM share_links LEFT JOIN users ON users.id = share_links.user_id
		WHERE `+where, arg)
	link, err := scanShareLink(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan share link: %w", err)
	}
	return link, nil
}

// ListShareLinks는 creatorID가 0이면 모든 링크를, 아니면 그 사용자가 만든 링크만 최근 순으로 돌려줍니다.
func ListShareLinks(conn *sql.DB, creatorID int64) ([]*ShareLink, error) {
	rows, err := conn.Query(`SELECT `+shareLinkColumns+`
		FROM share_links LEFT JOIN users ON users.id = share_links.user_id
		WHERE ? = 0 OR share_links.user_id = ?
		ORDER BY share_links.created_at DESC`, creatorID, creatorID)
	if err != nil {
		return nil, fmt.Errorf("query share links: %w", err)
	}
	defer rows.Close()
	links := []*ShareLink{}
	for rows.Next() {
		link, err := scanShareLink(rows)
		if err != nil {
			return nil, fmt.Errorf("scan share link: %w", err)
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

func RevokeShareLink(conn *sql.DB, id, revokedAt int64) error {
	_, err := conn.Exec(`UPDATE share_links SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL`, revokedAt, id)
	if err != nil {
		return fmt.Errorf("revoke share link: %w", err)
	}
	return nil
}

func CountShareDownload(conn *sql.DB, id, at int64) error {
	_, err := conn.Exec(`UPDATE share_links SET download_count = download_count + 1, last_download_at = ? WHERE id = ?`, at, id)
	if err != nil {
		return fmt.Errorf("count share download: %w", err)
	}
	return nil
}
