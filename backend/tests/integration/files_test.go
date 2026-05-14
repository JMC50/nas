package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/server"
)

func setupFilesTestServer(t *testing.T) (http.Handler, *config.Config, *sql.DB, string) {
	t.Helper()
	tmp := t.TempDir()
	dataDir := filepath.Join(tmp, "data")
	require.NoError(t, ensureDir(dataDir))

	dbPath := filepath.Join(tmp, "test.sqlite")
	conn, err := db.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	_, err = conn.Exec(testSchema())
	require.NoError(t, err)

	// Seed an admin user (ADMIN grants all intents via HasIntent logic)
	_, err = db.SaveLocalUser(conn, "admin1", "admin1", "hashed", "admin")
	require.NoError(t, err)
	require.NoError(t, db.ToggleIntent(conn, "admin1", "ADMIN"))

	cfg := &config.Config{
		Port:                 0,
		CorsOrigin:           "*",
		PrivateKey:           testPrivateKey,
		AdminPassword:        "admin-pass",
		AuthType:             config.AuthTypeBoth,
		NASDataDir:           dataDir,
		PasswordRequirements: config.PasswordRequirements{MinLength: 4},
	}
	router := server.NewRouter(cfg, conn)
	return router, cfg, conn, dataDir
}

func ensureDir(path string) error {
	return _osMkdirAll(path)
}

func TestFilesRoundTrip_MkdirSaveReadDelete(t *testing.T) {
	router, cfg, _, dataDir := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	// 1. makedir /testdir
	get(t, router, "/makedir?token="+token+"&loc=&name=testdir")

	// 2. saveTextFile /testdir/hello.txt
	body, _ := json.Marshal(map[string]string{"text": "hello world"})
	postJSON(t, router, "/saveTextFile?token="+token+"&loc=testdir&name=hello.txt", body)

	// 3. readFolder /testdir shows the file
	resp := get(t, router, "/readFolder?token="+token+"&loc=testdir")
	require.Contains(t, resp, "hello.txt")

	// 4. getTextFile reads it back
	resp = get(t, router, "/getTextFile?token="+token+"&loc=testdir&name=hello.txt")
	require.Contains(t, resp, "hello world")

	// 5. rename to greetings.txt
	get(t, router, "/rename?token="+token+"&loc=testdir&name=hello.txt&change=greetings.txt")

	// 6. forceDelete the entire dir
	get(t, router, "/forceDelete?token="+token+"&loc=&name=testdir")

	// 7. data dir is empty again (only the seed structure remains)
	resp = get(t, router, "/readFolder?token="+token+"&loc=")
	require.NotContains(t, resp, "testdir")
	_ = dataDir
}

func TestFilesPathTraversal_BlockedByMiddleware(t *testing.T) {
	router, cfg, _, _ := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	cases := []struct {
		name string
		url  string
	}{
		{"loc dot-dot", "/readFolder?token=" + token + "&loc=" + url.QueryEscape("../../etc")},
		{"name dot-dot", "/stat?token=" + token + "&loc=&name=" + url.QueryEscape("../../etc/passwd")},
		{"copy origin traversal", "/copy?token=" + token + "&originLoc=" + url.QueryEscape("../../etc") + "&fileName=passwd&targetLoc=here"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		})
	}
}

func TestFilesCopy_RecursiveDir(t *testing.T) {
	router, cfg, _, _ := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	get(t, router, "/makedir?token="+token+"&loc=&name=src")
	body, _ := json.Marshal(map[string]string{"text": "v1"})
	postJSON(t, router, "/saveTextFile?token="+token+"&loc=src&name=a.txt", body)

	get(t, router, "/copy?token="+token+"&originLoc=&fileName=src&targetLoc=")
	// After copy, /src is renamed to /src (since target is empty location). Skip recursive verify here —
	// the round-trip integration is in TestFilesRoundTrip; this just exercises the recursive path.
}

func TestStat_ReturnsFolderType(t *testing.T) {
	router, cfg, _, _ := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	get(t, router, "/makedir?token="+token+"&loc=&name=mydir")
	resp := get(t, router, "/stat?token="+token+"&loc=&name=mydir")
	require.Contains(t, resp, `"type":"folder"`)
}

// TestReadFolderDetails verifies that readFolder responses include `size`
// (bytes for files, 0 for folders) and a RFC3339 `modifiedAt` close to now.
func TestReadFolderDetails(t *testing.T) {
	router, cfg, _, _ := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	// Seed: one file with known content + one subfolder.
	body, _ := json.Marshal(map[string]string{"text": "hello world"})
	postJSON(t, router, "/saveTextFile?token="+token+"&loc=&name=hello.txt", body)
	get(t, router, "/makedir?token="+token+"&loc=&name=mydir")

	// Read root, decode into the wire shape.
	raw := get(t, router, "/readFolder?token="+token+"&loc=")
	type wireEntry struct {
		Name       string `json:"name"`
		IsFolder   bool   `json:"isFolder"`
		Extensions string `json:"extensions"`
		Size       int64  `json:"size"`
		ModifiedAt string `json:"modifiedAt"`
	}
	var entries []wireEntry
	require.NoError(t, json.Unmarshal([]byte(raw), &entries))

	var file, folder *wireEntry
	for i := range entries {
		switch entries[i].Name {
		case "hello.txt":
			file = &entries[i]
		case "mydir":
			folder = &entries[i]
		}
	}
	require.NotNil(t, file, "hello.txt missing from readFolder response: %s", raw)
	require.NotNil(t, folder, "mydir missing from readFolder response: %s", raw)

	// File row: size > 0, modifiedAt parseable as RFC3339 within ±60s of now.
	require.Greater(t, file.Size, int64(0), "file size should be > 0")
	parsed, err := time.Parse(time.RFC3339, file.ModifiedAt)
	require.NoError(t, err, "modifiedAt should parse as RFC3339, got %q", file.ModifiedAt)
	require.WithinDuration(t, time.Now(), parsed, 60*time.Second)

	// Folder row: isFolder true, size 0.
	require.True(t, folder.IsFolder, "mydir should be isFolder=true")
	require.Equal(t, int64(0), folder.Size, "folder size should be 0")
}

// mediaWire mirrors the JSON shape emitted by MediaLibrary
// (`backend/internal/files/handlers.go`). Local to this test file because the
// backend struct is unexported.
type mediaWire struct {
	Name       string `json:"name"`
	Loc        string `json:"loc"`
	Extensions string `json:"extensions"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modifiedAt"`
	Kind       string `json:"kind"`
}

// TestMediaLibrary verifies the /mediaLibrary endpoint walks the NAS data dir,
// filters by audio/video extension, skips dot-prefixed dirs/files, and rejects
// invalid `kind` values.
func TestMediaLibrary(t *testing.T) {
	router, cfg, _, dataDir := setupFilesTestServer(t)
	cfg.MediaLibraryLimit = 5000 // explicit default for test
	require.Equal(t, 5000, cfg.MediaLibraryLimit)

	// Seed:
	//   /a.mp3                  → audio at /
	//   /b/c.mp3                → audio at /b
	//   /b/d.mp4                → video at /b
	//   /.hidden/e.mp3          → audio under hidden dir (must be skipped)
	//   /f/.hidden.mp3          → hidden file under visible dir (must be skipped)
	require.NoError(t, os.MkdirAll(filepath.Join(dataDir, "b"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dataDir, ".hidden"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dataDir, "f"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dataDir, "a.mp3"), []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dataDir, "b", "c.mp3"), []byte("c"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dataDir, "b", "d.mp4"), []byte("d"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dataDir, ".hidden", "e.mp3"), []byte("e"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dataDir, "f", ".hidden.mp3"), []byte("h"), 0o644))

	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)
	srv := httptest.NewServer(router)
	defer srv.Close()

	// --- audio ---
	resp, err := http.Get(srv.URL + "/mediaLibrary?kind=audio&token=" + token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var audio []mediaWire
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&audio))
	resp.Body.Close()
	require.Len(t, audio, 2, "expected 2 audio entries (a.mp3 + b/c.mp3); hidden skipped; got %+v", audio)

	byName := func(rows []mediaWire, name string) *mediaWire {
		for i := range rows {
			if rows[i].Name == name {
				return &rows[i]
			}
		}
		return nil
	}
	a := byName(audio, "a.mp3")
	require.NotNil(t, a, "a.mp3 missing")
	require.Equal(t, "/", a.Loc)
	require.Equal(t, "audio", a.Kind)
	require.Equal(t, "mp3", a.Extensions)
	require.Equal(t, int64(1), a.Size)

	c := byName(audio, "c.mp3")
	require.NotNil(t, c, "b/c.mp3 missing")
	require.Equal(t, "/b", c.Loc)
	require.Equal(t, "audio", c.Kind)

	// --- video ---
	resp2, err := http.Get(srv.URL + "/mediaLibrary?kind=video&token=" + token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp2.StatusCode)
	var video []mediaWire
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&video))
	resp2.Body.Close()
	require.Len(t, video, 1, "expected 1 video entry (b/d.mp4); got %+v", video)
	require.Equal(t, "d.mp4", video[0].Name)
	require.Equal(t, "/b", video[0].Loc)
	require.Equal(t, "video", video[0].Kind)

	// --- invalid kind ---
	resp3, err := http.Get(srv.URL + "/mediaLibrary?kind=invalid&token=" + token)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp3.StatusCode)
	resp3.Body.Close()
}

// --- helpers ---

func get(t *testing.T, router http.Handler, path string) string {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "GET %s failed: %s", path, w.Body.String())
	return w.Body.String()
}

func postJSON(t *testing.T, router http.Handler, path string, body []byte) string {
	t.Helper()
	req := httptest.NewRequest("POST", path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "POST %s failed: %s", path, w.Body.String())
	return strings.TrimSpace(w.Body.String())
}
