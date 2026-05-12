package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/auth"
)

func TestVideoStream_HandlesRangeRequest(t *testing.T) {
	router, cfg, _, dataDir := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	// Create a fake video file with predictable bytes
	target := filepath.Join(dataDir, "movie.mp4")
	payload := make([]byte, 1024)
	for i := range payload {
		payload[i] = byte(i % 256)
	}
	require.NoError(t, os.WriteFile(target, payload, 0o644))

	// Full request
	req := httptest.NewRequest("GET", "/getVideoData?token="+token+"&loc=&name=movie.mp4", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 1024, w.Body.Len())

	// Range request: bytes 100-199
	req = httptest.NewRequest("GET", "/getVideoData?token="+token+"&loc=&name=movie.mp4", nil)
	req.Header.Set("Range", "bytes=100-199")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusPartialContent, w.Code)
	require.Equal(t, 100, w.Body.Len())
	require.Equal(t, "bytes 100-199/1024", w.Header().Get("Content-Range"))

	// Verify the actual byte content
	for i := 0; i < 100; i++ {
		require.Equal(t, byte((100+i)%256), w.Body.Bytes()[i],
			fmt.Sprintf("byte at offset %d should be %d", i, (100+i)%256))
	}
}

func TestImg_ReturnsEmbeddedIcon(t *testing.T) {
	router, _, _, _ := setupFilesTestServer(t)

	req := httptest.NewRequest("GET", "/img?type=mp4", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "image/png", w.Header().Get("Content-Type"))
	require.Greater(t, w.Body.Len(), 100, "PNG should be > 100 bytes")
	// PNG magic bytes
	require.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47}, w.Body.Bytes()[:4])
}

func TestImg_FallsBackToFilePng(t *testing.T) {
	router, _, _, _ := setupFilesTestServer(t)

	req := httptest.NewRequest("GET", "/img?type=nonexistent-type-xyz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47}, w.Body.Bytes()[:4])
}

func TestDownload_SetsContentDisposition(t *testing.T) {
	router, cfg, _, dataDir := setupFilesTestServer(t)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(dataDir, "report.pdf"), []byte("pdf-content"), 0o644))

	req := httptest.NewRequest("GET", "/download?token="+token+"&loc=&name=report.pdf", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Disposition"), `filename="report.pdf"`)
	require.Equal(t, "pdf-content", w.Body.String())
}
