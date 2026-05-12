package integration

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/auth"
)

func TestZipFiles_CreatesArchive(t *testing.T) {
	router, cfg, _, dataDir := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(dataDir, "a.txt"), []byte("aaaa"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dataDir, "b.txt"), []byte("bbbb"), 0o644))

	body, _ := json.Marshal([]map[string]any{
		{"loc": "", "name": "a.txt", "isFolder": false},
		{"loc": "", "name": "b.txt", "isFolder": false},
	})
	req := httptest.NewRequest("POST", "/zipFiles?token="+token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotEmpty(t, resp["progressId"])
	require.NotEmpty(t, resp["zipPath"])

	// Verify the zip exists and contains both files
	reader, err := zip.OpenReader(resp["zipPath"])
	require.NoError(t, err)
	defer reader.Close()
	names := []string{}
	for _, entry := range reader.File {
		names = append(names, entry.Name)
	}
	require.ElementsMatch(t, []string{"a.txt", "b.txt"}, names)

	// Progress should report done
	progressReq := httptest.NewRequest("GET", "/progress?progressId="+resp["progressId"], nil)
	progressW := httptest.NewRecorder()
	router.ServeHTTP(progressW, progressReq)
	require.Equal(t, http.StatusOK, progressW.Code)
	var progress map[string]any
	require.NoError(t, json.Unmarshal(progressW.Body.Bytes(), &progress))
	require.Equal(t, "done", progress["status"])
	require.Equal(t, float64(100), progress["percent"])
}

func TestUnzipFile_ExtractsToSubfolder(t *testing.T) {
	router, cfg, _, dataDir := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	// Build a zip on disk
	buffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(buffer)
	entry, _ := zipWriter.Create("file.txt")
	entry.Write([]byte("payload"))
	zipWriter.Close()
	zipPath := filepath.Join(dataDir, "bundle.zip")
	require.NoError(t, os.WriteFile(zipPath, buffer.Bytes(), 0o644))

	body, _ := json.Marshal(map[string]string{
		"loc":        "",
		"name":       "bundle.zip",
		"extensions": "zip",
	})
	req := httptest.NewRequest("POST", "/unzipFile?token="+token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	extracted, err := os.ReadFile(filepath.Join(dataDir, "bundle_unzipped", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "payload", string(extracted))
}

func TestDownloadZip_ServesFile(t *testing.T) {
	router, cfg, _, dataDir := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	zipPath := filepath.Join(dataDir, "test.zip")
	require.NoError(t, os.WriteFile(zipPath, []byte("zip-bytes"), 0o644))

	req := httptest.NewRequest("GET",
		"/downloadZip?token="+token+"&zipPath="+url.QueryEscape(zipPath), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Equal(t, "zip-bytes", w.Body.String())
}

func TestDownloadZip_RejectsPathOutsideDataDir(t *testing.T) {
	router, cfg, _, _ := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	req := httptest.NewRequest("GET",
		"/downloadZip?token="+token+"&zipPath="+url.QueryEscape("/etc/passwd"), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSystemInfo_ReturnsSnapshot(t *testing.T) {
	router, _, _, _ := setupFilesTestServer(t)

	req := httptest.NewRequest("GET", "/getSystemInfo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Contains(t, resp, "cpu")
	require.Contains(t, resp, "memory")
	require.Contains(t, resp, "uptime")
	require.Contains(t, resp, "disk")

	diskInfo := resp["disk"].(map[string]any)
	require.Contains(t, diskInfo, "total")
}
