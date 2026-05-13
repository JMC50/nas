package integration

import (
	"bytes"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/server"
)

func skipIfNoSoffice(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("soffice"); err != nil {
		t.Skip("soffice not installed; skipping office conversion tests")
	}
}

// setupOfficeTestServer mirrors setupFilesTestServer but also sets NASTempDir
// (office.NewHandlers requires it for the conversion cache).
func setupOfficeTestServer(t *testing.T) (http.Handler, *config.Config, *sql.DB, string) {
	t.Helper()
	tmp := t.TempDir()
	dataDir := filepath.Join(tmp, "data")
	tempDir := filepath.Join(tmp, "tmp")
	require.NoError(t, os.MkdirAll(dataDir, 0o755))
	require.NoError(t, os.MkdirAll(tempDir, 0o755))

	dbPath := filepath.Join(tmp, "test.sqlite")
	conn, err := db.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	_, err = conn.Exec(testSchema())
	require.NoError(t, err)

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
		NASTempDir:           tempDir,
		PasswordRequirements: config.PasswordRequirements{MinLength: 4},
	}
	router := server.NewRouter(cfg, conn)
	return router, cfg, conn, dataDir
}

func copyFixture(t *testing.T, dst, fixtureName string) {
	t.Helper()
	src := filepath.Join("..", "testdata", fixtureName)
	srcFile, err := os.Open(src)
	require.NoError(t, err)
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	require.NoError(t, err)
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	require.NoError(t, err)
}

func TestOfficeConvert_DocxToPdf(t *testing.T) {
	skipIfNoSoffice(t)
	router, cfg, _, dataDir := setupOfficeTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	copyFixture(t, filepath.Join(dataDir, "sample.docx"), "sample.docx")

	req := httptest.NewRequest("GET", "/getOfficePdf?token="+token+"&loc=&name=sample.docx", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	require.True(t, bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF-")), "response body must start with %PDF- magic")
}

func TestOfficeConvert_CacheHit(t *testing.T) {
	skipIfNoSoffice(t)
	router, cfg, _, dataDir := setupOfficeTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	copyFixture(t, filepath.Join(dataDir, "sample.docx"), "sample.docx")

	url := "/getOfficePdf?token=" + token + "&loc=&name=sample.docx"

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, httptest.NewRequest("GET", url, nil))
	require.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, httptest.NewRequest("GET", url, nil))
	require.Equal(t, http.StatusOK, w2.Code)
	require.Equal(t, w1.Body.Len(), w2.Body.Len(), "cached response should be identical size")
}

func TestOfficeConvert_PathTraversal(t *testing.T) {
	skipIfNoSoffice(t)
	router, cfg, _, _ := setupOfficeTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/getOfficePdf?token="+token+"&loc=..%2F..%2Fetc&name=passwd", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOfficeConvert_ConcurrentDedup(t *testing.T) {
	skipIfNoSoffice(t)
	router, cfg, _, dataDir := setupOfficeTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	copyFixture(t, filepath.Join(dataDir, "sample.docx"), "sample.docx")

	url := "/getOfficePdf?token=" + token + "&loc=&name=sample.docx"

	const N = 10
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest("GET", url, nil))
			require.Equal(t, http.StatusOK, w.Code, strings.TrimSpace(w.Body.String()))
		}()
	}
	wg.Wait()
	cacheDir := filepath.Join(cfg.NASTempDir, "office-cache")
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	require.Equal(t, 1, len(entries), "exactly one cached pdf expected, got %d", len(entries))
}
