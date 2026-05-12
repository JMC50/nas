package integration

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/db"
)

func TestLegacyInput_StreamsBodyToDisk(t *testing.T) {
	router, cfg, _, dataDir := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	payload := []byte("hello from /input endpoint")
	req := httptest.NewRequest("POST",
		"/input?token="+token+"&loc=&name=upload.txt",
		bytes.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	written, err := os.ReadFile(filepath.Join(dataDir, "upload.txt"))
	require.NoError(t, err)
	require.Equal(t, payload, written)
}

func TestLegacyInputZip_ExtractsAfterUpload(t *testing.T) {
	router, cfg, _, dataDir := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	buffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(buffer)
	entryWriter, err := zipWriter.Create("inner.txt")
	require.NoError(t, err)
	_, err = entryWriter.Write([]byte("zip content"))
	require.NoError(t, err)
	require.NoError(t, zipWriter.Close())

	req := httptest.NewRequest("POST",
		"/inputZip?token="+token+"&loc=&name=bundle.zip",
		bytes.NewReader(buffer.Bytes()))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	extracted, err := os.ReadFile(filepath.Join(dataDir, "inner.txt"))
	require.NoError(t, err)
	require.Equal(t, []byte("zip content"), extracted)
	_, err = os.Stat(filepath.Join(dataDir, "bundle.zip"))
	require.True(t, os.IsNotExist(err), "temp zip should be removed after extraction")
}

func TestTusUpload_RequiresBearerToken(t *testing.T) {
	router, _, _, _ := setupFilesTestServer(t)

	req := httptest.NewRequest("POST", "/files/", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", "100")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code, w.Body.String())
}

func TestTusUpload_RejectsMissingIntent(t *testing.T) {
	router, cfg, conn, _ := setupFilesTestServer(t)

	// Seed a user with NO intents
	_, err := db.SaveLocalUser(conn, "no-upload", "no-upload", "h", "no-upload")
	require.NoError(t, err)

	token, err := auth.IssueToken("no-upload", cfg.PrivateKey)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/files/", strings.NewReader(""))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", "5")
	req.Header.Set("Upload-Metadata", "filename "+b64("test.txt")+",loc "+b64(""))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
}

func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
