package integration

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/server"
)

type shareFixture struct {
	router  http.Handler
	conn    *sql.DB
	dataDir string
	outside string
	tokens  map[string]string
}

// newShareFixture는 실제 스키마(db.InitSchema)로 서버를 만들고 권한이 다른 사용자 네 명을 넣는다.
func newShareFixture(t *testing.T) *shareFixture {
	t.Helper()
	tmp := t.TempDir()
	dataDir := filepath.Join(tmp, "data")
	outside := filepath.Join(tmp, "outside")
	require.NoError(t, os.MkdirAll(dataDir, 0o755))
	require.NoError(t, os.MkdirAll(outside, 0o755))

	conn, err := db.Open(filepath.Join(tmp, "test.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	require.NoError(t, db.InitSchema(conn))

	users := map[string][]string{
		"admin": {"ADMIN"},
		"alice": {"VIEW", "OPEN", "DOWNLOAD"},
		"bob":   {"VIEW"},
		"carol": {"DOWNLOAD"},
	}
	tokens := map[string]string{}
	for userID, intents := range users {
		_, err := db.SaveLocalUser(conn, userID, userID, "hashed", userID)
		require.NoError(t, err)
		for _, intent := range intents {
			require.NoError(t, db.ToggleIntent(conn, userID, intent))
		}
		token, err := auth.IssueToken(userID, testPrivateKey)
		require.NoError(t, err)
		tokens[userID] = token
	}

	cfg := &config.Config{
		CorsOrigin:           "*",
		PrivateKey:           testPrivateKey,
		AuthType:             config.AuthTypeBoth,
		NASDataDir:           dataDir,
		PasswordRequirements: config.PasswordRequirements{MinLength: 4},
	}
	return &shareFixture{router: server.NewRouter(cfg, conn), conn: conn, dataDir: dataDir, outside: outside, tokens: tokens}
}

func (f *shareFixture) write(t *testing.T, rel, content string) {
	t.Helper()
	path := filepath.Join(f.dataDir, filepath.FromSlash(rel))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func (f *shareFixture) do(method, path, user string, body any, header map[string]string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		encoded, _ := json.Marshal(body)
		reader = bytes.NewReader(encoded)
	}
	req := httptest.NewRequest(method, path, reader)
	if user != "" {
		req.Header.Set("Authorization", "Bearer "+f.tokens[user])
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

type createdLink struct {
	ID        int64  `json:"id"`
	Path      string `json:"path"`
	IsFolder  bool   `json:"isFolder"`
	ExpiresAt int64  `json:"expiresAt"`
}

func (f *shareFixture) create(t *testing.T, user, loc, name, expiresIn string) createdLink {
	t.Helper()
	w := f.do("POST", "/server/share-links", user, map[string]string{"loc": loc, "name": name, "expiresIn": expiresIn}, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var link createdLink
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &link))
	require.True(t, strings.HasPrefix(link.Path, "/d/"))
	return link
}

func TestShareLink_CreateAndDownloadFile(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "문서/report.txt", "quarterly numbers")

	before := time.Now()
	link := f.create(t, "alice", "/문서", "report.txt", "7d")
	require.False(t, link.IsFolder)
	require.InDelta(t, before.Add(7*24*time.Hour).UnixMilli(), link.ExpiresAt, float64(time.Minute.Milliseconds()))

	token := strings.TrimPrefix(link.Path, "/d/")
	require.GreaterOrEqual(t, len(token), 43, "token must carry at least 256 bits of base64url text")

	w := f.do("GET", link.Path, "", nil, nil)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "quarterly numbers", w.Body.String())
	require.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	require.Contains(t, w.Header().Get("Content-Disposition"), "report.txt")
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	var stored string
	var count int
	require.NoError(t, f.conn.QueryRow(`SELECT token_hash, download_count FROM share_links WHERE id = ?`, link.ID).Scan(&stored, &count))
	require.NotEqual(t, token, stored)
	require.NotContains(t, stored, token)
	require.Equal(t, 1, count)

	var description string
	require.NoError(t, f.conn.QueryRow(`SELECT description FROM log WHERE activity = 'DOWNLOAD'`).Scan(&description))
	require.Equal(t, "DOWNLOAD [LINK] AT /문서/report.txt", description)
}

func TestShareLink_CreateRequiresDownloadPermissionAndValidInput(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "a.txt", "a")

	require.Equal(t, http.StatusUnauthorized, f.do("POST", "/server/share-links", "", map[string]string{"loc": "/", "name": "a.txt", "expiresIn": "1d"}, nil).Code)
	require.Equal(t, http.StatusForbidden, f.do("POST", "/server/share-links", "bob", map[string]string{"loc": "/", "name": "a.txt", "expiresIn": "1d"}, nil).Code)
	require.Equal(t, http.StatusBadRequest, f.do("POST", "/server/share-links", "alice", map[string]string{"loc": "/", "name": "a.txt", "expiresIn": "2h"}, nil).Code)
	require.Equal(t, http.StatusNotFound, f.do("POST", "/server/share-links", "alice", map[string]string{"loc": "/", "name": "missing.txt", "expiresIn": "1d"}, nil).Code)
	for _, expiresIn := range []string{"1h", "1d", "7d", "30d"} {
		f.create(t, "alice", "/", "a.txt", expiresIn)
	}
}

func TestShareLink_ExpiredLinkIsNotFound(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "a.txt", "a")
	link := f.create(t, "alice", "/", "a.txt", "1h")

	_, err := f.conn.Exec(`UPDATE share_links SET expires_at = ? WHERE id = ?`, time.Now().Add(-time.Second).UnixMilli(), link.ID)
	require.NoError(t, err)

	w := f.do("GET", link.Path, "", nil, nil)
	require.Equal(t, http.StatusNotFound, w.Code)
	require.NotContains(t, strings.ToLower(w.Body.String()), "expire")
}

func TestShareLink_RevokeByCreatorOrAdminOnly(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "a.txt", "a")
	aliceLink := f.create(t, "alice", "/", "a.txt", "1d")
	carolLink := f.create(t, "carol", "/", "a.txt", "1d")

	require.Equal(t, http.StatusNotFound, f.do("DELETE", fmt.Sprintf("/server/share-links/%d", aliceLink.ID), "carol", nil, nil).Code)
	require.Equal(t, http.StatusOK, f.do("GET", aliceLink.Path, "", nil, nil).Code)

	require.Equal(t, http.StatusOK, f.do("DELETE", fmt.Sprintf("/server/share-links/%d", aliceLink.ID), "alice", nil, nil).Code)
	require.Equal(t, http.StatusNotFound, f.do("GET", aliceLink.Path, "", nil, nil).Code)

	require.Equal(t, http.StatusOK, f.do("DELETE", fmt.Sprintf("/server/share-links/%d", carolLink.ID), "admin", nil, nil).Code)
	require.Equal(t, http.StatusNotFound, f.do("GET", carolLink.Path, "", nil, nil).Code)
}

func TestShareLink_ListShowsOwnLinksOrAllForAdmin(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "a.txt", "a")
	f.create(t, "alice", "/", "a.txt", "1d")
	f.create(t, "carol", "/", "a.txt", "1d")

	list := func(user string) (bool, []string) {
		w := f.do("GET", "/server/share-links", user, nil, nil)
		require.Equal(t, http.StatusOK, w.Code)
		var body struct {
			All   bool `json:"all"`
			Links []struct {
				CreatorID     string `json:"creatorId"`
				Status        string `json:"status"`
				DownloadCount int    `json:"downloadCount"`
			} `json:"links"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		creators := []string{}
		for _, link := range body.Links {
			creators = append(creators, link.CreatorID)
			require.Equal(t, "active", link.Status)
		}
		sort.Strings(creators)
		return body.All, creators
	}

	all, creators := list("alice")
	require.False(t, all)
	require.Equal(t, []string{"alice"}, creators)
	all, creators = list("admin")
	require.True(t, all)
	require.Equal(t, []string{"alice", "carol"}, creators)
	require.Equal(t, http.StatusUnauthorized, f.do("GET", "/server/share-links", "", nil, nil).Code)
}

func TestShareLink_CreatorLosingPermissionClosesLink(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "a.txt", "a")
	link := f.create(t, "alice", "/", "a.txt", "1d")
	require.Equal(t, http.StatusOK, f.do("GET", link.Path, "", nil, nil).Code)

	require.NoError(t, db.ToggleIntent(f.conn, "alice", "DOWNLOAD"))
	require.Equal(t, http.StatusNotFound, f.do("GET", link.Path, "", nil, nil).Code)

	require.NoError(t, db.ToggleIntent(f.conn, "alice", "DOWNLOAD"))
	require.Equal(t, http.StatusOK, f.do("GET", link.Path, "", nil, nil).Code)

	require.NoError(t, db.DeleteUser(f.conn, "alice"))
	require.Equal(t, http.StatusNotFound, f.do("GET", link.Path, "", nil, nil).Code)
}

func TestShareLink_TraversalIsRejected(t *testing.T) {
	f := newShareFixture(t)
	require.NoError(t, os.WriteFile(filepath.Join(f.outside, "secret.txt"), []byte("secret"), 0o644))
	f.write(t, "a.txt", "a")

	for _, req := range []map[string]string{
		{"loc": "/../outside", "name": "secret.txt", "expiresIn": "1d"},
		{"loc": "/", "name": "../outside/secret.txt", "expiresIn": "1d"},
		{"loc": "/", "name": "..", "expiresIn": "1d"},
	} {
		w := f.do("POST", "/server/share-links", "alice", req, nil)
		require.Contains(t, []int{http.StatusBadRequest, http.StatusNotFound}, w.Code, req)
	}

	link := f.create(t, "alice", "/", "a.txt", "1d")
	_, err := f.conn.Exec(`UPDATE share_links SET loc = '/../outside', name = 'secret.txt' WHERE id = ?`, link.ID)
	require.NoError(t, err)
	w := f.do("GET", link.Path, "", nil, nil)
	require.Equal(t, http.StatusNotFound, w.Code)
	require.NotContains(t, w.Body.String(), "secret")
}

func TestShareLink_SymlinkEscapeIsRejected(t *testing.T) {
	f := newShareFixture(t)
	secret := filepath.Join(f.outside, "secret.txt")
	require.NoError(t, os.WriteFile(secret, []byte("secret"), 0o644))
	require.NoError(t, os.Symlink(secret, filepath.Join(f.dataDir, "escape.txt")))
	require.NoError(t, os.Symlink(f.outside, filepath.Join(f.dataDir, "escape-dir")))

	require.Equal(t, http.StatusNotFound, f.do("POST", "/server/share-links", "alice", map[string]string{"loc": "/", "name": "escape.txt", "expiresIn": "1d"}, nil).Code)
	require.Equal(t, http.StatusNotFound, f.do("POST", "/server/share-links", "alice", map[string]string{"loc": "/", "name": "escape-dir", "expiresIn": "1d"}, nil).Code)

	f.write(t, "swap.txt", "inside")
	link := f.create(t, "alice", "/", "swap.txt", "1d")
	require.NoError(t, os.Remove(filepath.Join(f.dataDir, "swap.txt")))
	require.NoError(t, os.Symlink(secret, filepath.Join(f.dataDir, "swap.txt")))
	w := f.do("GET", link.Path, "", nil, nil)
	require.Equal(t, http.StatusNotFound, w.Code)
	require.NotContains(t, w.Body.String(), "secret")
}

func TestShareLink_RangeDownload(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "digits.txt", "0123456789")
	link := f.create(t, "alice", "/", "digits.txt", "1d")

	w := f.do("GET", link.Path, "", nil, map[string]string{"Range": "bytes=2-5"})
	require.Equal(t, http.StatusPartialContent, w.Code)
	require.Equal(t, "2345", w.Body.String())
	require.Equal(t, "bytes 2-5/10", w.Header().Get("Content-Range"))

	var count int
	require.NoError(t, f.conn.QueryRow(`SELECT download_count FROM share_links WHERE id = ?`, link.ID).Scan(&count))
	require.Equal(t, 0, count, "a resumed range request is not a new download")
}

func readZip(t *testing.T, body []byte) map[string]string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	require.NoError(t, err)
	entries := map[string]string{}
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, "/") {
			entries[file.Name] = ""
			continue
		}
		rc, err := file.Open()
		require.NoError(t, err)
		content, err := io.ReadAll(rc)
		require.NoError(t, err)
		rc.Close()
		entries[file.Name] = string(content)
	}
	return entries
}

func TestShareLink_FolderDownloadsAsZip(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "사진/바다.png", "png bytes")
	f.write(t, "사진/여행/메모.txt", "memo")
	require.NoError(t, os.MkdirAll(filepath.Join(f.dataDir, "사진", "빈 폴더"), 0o755))

	link := f.create(t, "alice", "/", "사진", "1d")
	require.True(t, link.IsFolder)

	w := f.do("GET", link.Path, "", nil, nil)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/zip", w.Header().Get("Content-Type"))
	require.Contains(t, w.Header().Get("Content-Disposition"), ".zip")

	entries := readZip(t, w.Body.Bytes())
	require.Equal(t, "png bytes", entries["사진/바다.png"])
	require.Equal(t, "memo", entries["사진/여행/메모.txt"])
	require.Contains(t, entries, "사진/빈 폴더/")
}

func TestShareLink_FolderZipSkipsSymlinkEscapes(t *testing.T) {
	f := newShareFixture(t)
	secret := filepath.Join(f.outside, "secret.txt")
	require.NoError(t, os.WriteFile(secret, []byte("secret"), 0o644))
	f.write(t, "공유/a.txt", "a")
	f.write(t, "다른 폴더/b.txt", "b")
	require.NoError(t, os.Symlink(secret, filepath.Join(f.dataDir, "공유", "escape.txt")))
	require.NoError(t, os.Symlink(f.outside, filepath.Join(f.dataDir, "공유", "escape-dir")))
	require.NoError(t, os.Symlink(filepath.Join(f.dataDir, "다른 폴더", "b.txt"), filepath.Join(f.dataDir, "공유", "inside.txt")))

	link := f.create(t, "alice", "/", "공유", "1d")
	w := f.do("GET", link.Path, "", nil, nil)
	require.Equal(t, http.StatusOK, w.Code)

	entries := readZip(t, w.Body.Bytes())
	require.Equal(t, "a", entries["공유/a.txt"])
	require.Equal(t, "b", entries["공유/inside.txt"])
	require.NotContains(t, entries, "공유/escape.txt")
	require.NotContains(t, entries, "공유/escape-dir/")
	require.NotContains(t, entries, "공유/escape-dir/secret.txt")
	require.NotContains(t, w.Body.String(), "secret")
}
