package integration

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/server"
)

type zipItem struct {
	Loc  string `json:"loc"`
	Name string `json:"name"`
}

type preparedZip struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Entries int    `json:"entries"`
	Bytes   int64  `json:"bytes"`
}

// 같은 데이터 폴더와 데이터베이스에 zip 한도만 다른 서버를 하나 더 만듦
func (f *shareFixture) withZipLimits(maxFiles int, maxBytes int64) *shareFixture {
	cfg := &config.Config{
		CorsOrigin:           "*",
		PrivateKey:           testPrivateKey,
		AuthType:             config.AuthTypeBoth,
		NASDataDir:           f.dataDir,
		PasswordRequirements: config.PasswordRequirements{MinLength: 4},
		ZipDownloadMaxFiles:  maxFiles,
		ZipDownloadMaxBytes:  maxBytes,
	}
	limited := *f
	limited.router = server.NewRouter(cfg, f.conn)
	return &limited
}

func (f *shareFixture) prepareZip(user string, items ...zipItem) *httptest.ResponseRecorder {
	return f.do("POST", "/server/zip-downloads", user, map[string]any{"items": items}, nil)
}

func (f *shareFixture) mustPrepareZip(t *testing.T, user string, items ...zipItem) preparedZip {
	t.Helper()
	w := f.prepareZip(user, items...)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var prepared preparedZip
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &prepared))
	return prepared
}

// 브라우저처럼 token 쿼리로 받음
func (f *shareFixture) zipURL(prepared preparedZip, user string) string {
	return "/server" + prepared.Path + "?token=" + f.tokens[user]
}

func TestZipDownload_KeepsFolderStructureEmptyFoldersAndKoreanNames(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "자료/사진/바다.png", "png bytes")
	f.write(t, "자료/사진/여행/메모.txt", "memo")
	f.write(t, "자료/보고서.txt", "report")
	f.write(t, "자료/other.txt", "not selected")
	require.NoError(t, os.MkdirAll(filepath.Join(f.dataDir, "자료", "사진", "빈 폴더"), 0o755))

	prepared := f.mustPrepareZip(t, "alice", zipItem{"/자료", "사진"}, zipItem{"/자료", "보고서.txt"})
	require.Equal(t, "자료.zip", prepared.Name)
	require.Equal(t, 6, prepared.Entries)
	require.Equal(t, int64(len("png bytes")+len("memo")+len("report")), prepared.Bytes)

	w := f.do("GET", f.zipURL(prepared, "alice"), "", nil, nil)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/zip", w.Header().Get("Content-Type"))
	require.Equal(t, `attachment; filename="__.zip"; filename*=UTF-8''%EC%9E%90%EB%A3%8C.zip`, w.Header().Get("Content-Disposition"))
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	body := w.Body.Bytes()
	entries := readZip(t, body)
	require.Equal(t, map[string]string{
		"사진/":          "",
		"사진/바다.png":    "png bytes",
		"사진/빈 폴더/":     "",
		"사진/여행/":       "",
		"사진/여행/메모.txt": "memo",
		"보고서.txt":      "report",
	}, entries)

	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	require.NoError(t, err)
	for _, file := range reader.File {
		require.NotZero(t, file.Flags&0x800, "%s must carry the UTF-8 name flag", file.Name)
	}

	var description, loc string
	require.NoError(t, f.conn.QueryRow(`SELECT description, loc FROM log WHERE activity = 'DOWNLOAD'`).Scan(&description, &loc))
	require.Equal(t, "DOWNLOAD [ZIP] 2 ITEMS AT /자료", description)
	require.Equal(t, "/자료", loc)
}

func TestZipDownload_NamesTheFileAfterTheSelection(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "docs/a.txt", "a")
	f.write(t, "docs/b.txt", "b")
	f.write(t, "top.txt", "t")
	f.write(t, "other.txt", "o")

	require.Equal(t, "docs.zip", f.mustPrepareZip(t, "alice", zipItem{"/", "docs"}).Name)
	require.Equal(t, "docs.zip", f.mustPrepareZip(t, "alice", zipItem{"/docs", "a.txt"}, zipItem{"/docs", "b.txt"}).Name)
	require.Equal(t, "nas-download.zip", f.mustPrepareZip(t, "alice", zipItem{"/", "top.txt"}, zipItem{"/", "other.txt"}).Name)
	require.Equal(t, "nas-download.zip", f.mustPrepareZip(t, "alice", zipItem{"/", "top.txt"}, zipItem{"/docs", "a.txt"}).Name)
}

func TestZipDownload_SameNamesGetASuffixInsteadOfOverwriting(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "a/x.txt", "from a")
	f.write(t, "b/x.txt", "from b")
	f.write(t, "c/X.TXT", "from c")
	f.write(t, "a/set/one.txt", "a one")
	f.write(t, "b/set/one.txt", "b one")

	prepared := f.mustPrepareZip(t, "alice",
		zipItem{"/a", "x.txt"}, zipItem{"/b", "x.txt"}, zipItem{"/c", "X.TXT"}, zipItem{"/a", "x.txt"},
		zipItem{"/a", "set"}, zipItem{"/b", "set"})
	w := f.do("GET", f.zipURL(prepared, "alice"), "", nil, nil)
	require.Equal(t, http.StatusOK, w.Code)

	require.Equal(t, map[string]string{
		"x.txt":           "from a",
		"x (2).txt":       "from b",
		"X (3).TXT":       "from c",
		"set/":            "",
		"set/one.txt":     "a one",
		"set (2)/":        "",
		"set (2)/one.txt": "b one",
	}, readZip(t, w.Body.Bytes()))
}

func TestZipDownload_RequiresLoginAndDownloadPermission(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "a.txt", "a")
	item := zipItem{"/", "a.txt"}

	require.Equal(t, http.StatusUnauthorized, f.prepareZip("", item).Code)
	require.Equal(t, http.StatusForbidden, f.prepareZip("bob", item).Code)

	prepared := f.mustPrepareZip(t, "alice", item)
	require.Equal(t, http.StatusUnauthorized, f.do("GET", "/server"+prepared.Path, "", nil, nil).Code)
	require.Equal(t, http.StatusUnauthorized, f.do("GET", "/server"+prepared.Path+"?token=not-a-token", "", nil, nil).Code)
	require.Equal(t, http.StatusForbidden, f.do("GET", f.zipURL(prepared, "bob"), "", nil, nil).Code)
	require.Equal(t, http.StatusNotFound, f.do("GET", f.zipURL(prepared, "carol"), "", nil, nil).Code, "another user's address must look missing")
	require.Equal(t, http.StatusNotFound, f.do("GET", "/server/zip-downloads/unknown", "alice", nil, nil).Code)

	require.Equal(t, http.StatusOK, f.do("GET", f.zipURL(prepared, "alice"), "", nil, nil).Code)
	require.Equal(t, http.StatusOK, f.do("GET", f.zipURL(f.mustPrepareZip(t, "admin", item), "admin"), "", nil, nil).Code)

	require.NoError(t, db.ToggleIntent(f.conn, "alice", "DOWNLOAD"))
	require.Equal(t, http.StatusForbidden, f.do("GET", f.zipURL(prepared, "alice"), "", nil, nil).Code)
}

func TestZipDownload_RejectsBadInputAndPathsOutsideTheDataFolder(t *testing.T) {
	f := newShareFixture(t)
	require.NoError(t, os.WriteFile(filepath.Join(f.outside, "secret.txt"), []byte("secret"), 0o644))
	f.write(t, "a.txt", "a")

	require.Equal(t, http.StatusBadRequest, f.prepareZip("alice").Code)
	require.Equal(t, http.StatusBadRequest, f.do("POST", "/server/zip-downloads", "alice", "not an object", nil).Code)
	for _, name := range []string{"", ".", "..", "../outside/secret.txt", `a\b`} {
		require.Equal(t, http.StatusBadRequest, f.prepareZip("alice", zipItem{"/", name}).Code, name)
	}
	for _, item := range []zipItem{
		{"/../outside", "secret.txt"},
		{"/../../outside", "secret.txt"},
		{f.outside, "secret.txt"},
		{"/", "missing.txt"},
	} {
		w := f.prepareZip("alice", zipItem{"/", "a.txt"}, item)
		require.Equal(t, http.StatusNotFound, w.Code, item)
		require.NotContains(t, w.Body.String(), f.dataDir)
	}
}

func TestZipDownload_SymlinksNeverLeaveTheDataFolder(t *testing.T) {
	f := newShareFixture(t)
	secret := filepath.Join(f.outside, "secret.txt")
	require.NoError(t, os.WriteFile(secret, []byte("secret"), 0o644))
	f.write(t, "공유/a.txt", "a")
	f.write(t, "다른 폴더/b.txt", "b")
	require.NoError(t, os.Symlink(secret, filepath.Join(f.dataDir, "escape.txt")))
	require.NoError(t, os.Symlink(f.outside, filepath.Join(f.dataDir, "escape-dir")))
	require.NoError(t, os.Symlink(secret, filepath.Join(f.dataDir, "공유", "escape.txt")))
	require.NoError(t, os.Symlink(f.outside, filepath.Join(f.dataDir, "공유", "escape-dir")))
	require.NoError(t, os.Symlink(filepath.Join("..", "다른 폴더", "b.txt"), filepath.Join(f.dataDir, "공유", "inside.txt")))
	require.NoError(t, os.Symlink(filepath.Join(f.dataDir, "다른 폴더", "b.txt"), filepath.Join(f.dataDir, "공유", "absolute.txt")))

	require.Equal(t, http.StatusNotFound, f.prepareZip("alice", zipItem{"/", "escape.txt"}).Code)
	require.Equal(t, http.StatusNotFound, f.prepareZip("alice", zipItem{"/", "escape-dir"}).Code)

	prepared := f.mustPrepareZip(t, "alice", zipItem{"/", "공유"})
	w := f.do("GET", f.zipURL(prepared, "alice"), "", nil, nil)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, map[string]string{
		"공유/":           "",
		"공유/a.txt":      "a",
		"공유/inside.txt": "b",
	}, readZip(t, w.Body.Bytes()))
	require.NotContains(t, w.Body.String(), "secret")
}

func TestZipDownload_ATargetSwappedForAnOutsideLinkIsLeftOut(t *testing.T) {
	f := newShareFixture(t)
	secret := filepath.Join(f.outside, "secret.txt")
	require.NoError(t, os.WriteFile(secret, []byte("secret"), 0o644))
	f.write(t, "swap.txt", "inside")
	f.write(t, "keep.txt", "keep")

	prepared := f.mustPrepareZip(t, "alice", zipItem{"/", "swap.txt"}, zipItem{"/", "keep.txt"})
	require.NoError(t, os.Remove(filepath.Join(f.dataDir, "swap.txt")))
	require.NoError(t, os.Symlink(secret, filepath.Join(f.dataDir, "swap.txt")))

	w := f.do("GET", f.zipURL(prepared, "alice"), "", nil, nil)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, map[string]string{"keep.txt": "keep"}, readZip(t, w.Body.Bytes()))
}

func TestZipDownload_RefusesSelectionsOverTheConfiguredLimits(t *testing.T) {
	f := newShareFixture(t)
	for index := range 4 {
		f.write(t, fmt.Sprintf("many/%d.txt", index), "0123456789")
	}

	refusal := func(limited *shareFixture) (int, string, int64) {
		w := limited.prepareZip("alice", zipItem{"/", "many"})
		var body struct {
			Error string `json:"error"`
			Limit int64  `json:"limit"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body), w.Body.String())
		return w.Code, body.Error, body.Limit
	}

	code, reason, limit := refusal(f.withZipLimits(4, 0))
	require.Equal(t, http.StatusRequestEntityTooLarge, code)
	require.Equal(t, "too_many_entries", reason, "the folder itself counts, so four files make five entries")
	require.Equal(t, int64(4), limit)

	code, reason, limit = refusal(f.withZipLimits(0, 39))
	require.Equal(t, http.StatusRequestEntityTooLarge, code)
	require.Equal(t, "too_large", reason)
	require.Equal(t, int64(39), limit)

	exact := f.withZipLimits(5, 40)
	prepared := exact.mustPrepareZip(t, "alice", zipItem{"/", "many"})
	w := exact.do("GET", exact.zipURL(prepared, "alice"), "", nil, nil)
	require.Equal(t, http.StatusOK, w.Code)
	require.Len(t, readZip(t, w.Body.Bytes()), 5)
}

func TestZipDownload_StopsWhenTheSelectionGrowsPastTheLimitAfterTheCheck(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "grow/a.txt", "a")
	limited := f.withZipLimits(3, 0)
	prepared := limited.mustPrepareZip(t, "alice", zipItem{"/", "grow"})
	for index := range 5 {
		f.write(t, fmt.Sprintf("grow/later%d.txt", index), "later")
	}

	// 아직 아무것도 보내지 않았으면 서버는 응답 없이 연결을 끊음
	resp, err := http.Get(limited.serve(t).URL + limited.zipURL(prepared, "alice"))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	require.Error(t, err, "an unfinished zip must not end like a complete response")
}

func TestZipDownload_LosingPermissionStopsARunningDownload(t *testing.T) {
	f := newShareFixture(t)
	var size int64
	for part := range 4 {
		size += f.writeBig(t, fmt.Sprintf("big/part%d.mp4", part), 12)
	}
	prepared := f.mustPrepareZip(t, "alice", zipItem{"/", "big"})

	resp := openStream(t, f.serve(t), f.zipURL(prepared, "alice"))
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, err := io.CopyN(io.Discard, resp.Body, 1024)
	require.NoError(t, err)

	require.NoError(t, db.ToggleIntent(f.conn, "alice", "DOWNLOAD"))
	rest, err := io.Copy(io.Discard, resp.Body)
	require.Error(t, err)
	require.Less(t, rest+1024, size)
}

func TestZipDownload_AUserRunsTwoZipsAndADisconnectFreesTheSlot(t *testing.T) {
	f := newShareFixture(t)
	for part := range 4 {
		f.writeBig(t, fmt.Sprintf("big/part%d.mp4", part), 12)
	}
	item := zipItem{"/", "big"}
	prepared := f.mustPrepareZip(t, "alice", item)
	adminPrepared := f.mustPrepareZip(t, "admin", item)
	srv := f.serve(t)

	var running []*http.Response
	for range 2 {
		resp := openStream(t, srv, f.zipURL(prepared, "alice"))
		require.Equal(t, http.StatusOK, resp.StatusCode)
		_, err := io.CopyN(io.Discard, resp.Body, 1024)
		require.NoError(t, err)
		running = append(running, resp)
	}
	extra := openStream(t, srv, f.zipURL(prepared, "alice"))
	extra.Body.Close()
	require.Equal(t, http.StatusTooManyRequests, extra.StatusCode)
	require.Equal(t, "30", extra.Header.Get("Retry-After"))
	busy := f.prepareZip("alice", item)
	require.Equal(t, http.StatusTooManyRequests, busy.Code, "the check before the download reports the same refusal")

	other := openStream(t, srv, f.zipURL(f.mustPrepareZip(t, "carol", item), "carol"))
	require.Equal(t, http.StatusOK, other.StatusCode, "one user's downloads do not use up another user's slots")
	full := openStream(t, srv, f.zipURL(adminPrepared, "admin"))
	full.Body.Close()
	require.Equal(t, http.StatusServiceUnavailable, full.StatusCode, "three zips run on the whole server at once")
	require.Equal(t, http.StatusServiceUnavailable, f.prepareZip("admin", item).Code)
	other.Body.Close()

	for _, resp := range running {
		resp.Body.Close()
	}
	requireSlotFreed(t, srv, f.zipURL(prepared, "alice"))
}
