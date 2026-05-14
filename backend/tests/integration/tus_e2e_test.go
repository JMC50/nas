package integration

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/JMC50/nas/internal/auth"
)

// tusClient drives the tus 1.0.0 protocol over an httptest.Server's URL.
// Methods return raw HTTP responses + parsed offsets so tests can verify each
// protocol step.
type tusClient struct {
	t       *testing.T
	baseURL string
	token   string
	client  *http.Client
}

func newTusClient(t *testing.T, baseURL, token string) *tusClient {
	return &tusClient{
		t:       t,
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *tusClient) create(totalSize int64, metadata map[string]string) (string, error) {
	c.t.Helper()
	req, err := http.NewRequest("POST", c.baseURL+"/files/", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", fmt.Sprintf("%d", totalSize))
	req.Header.Set("Upload-Metadata", encodeMetadata(metadata))
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("tus create status %d: %s", resp.StatusCode, string(body))
	}
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("tus create response missing Location header")
	}
	return location, nil
}

func (c *tusClient) patch(uploadURL string, offset int64, chunk []byte) (int64, error) {
	c.t.Helper()
	req, err := http.NewRequest("PATCH", uploadURL, bytes.NewReader(chunk))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Content-Type", "application/offset+octet-stream")
	req.Header.Set("Upload-Offset", fmt.Sprintf("%d", offset))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.ContentLength = int64(len(chunk))

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("tus patch status %d: %s", resp.StatusCode, string(body))
	}
	var newOffset int64
	if _, err := fmt.Sscanf(resp.Header.Get("Upload-Offset"), "%d", &newOffset); err != nil {
		return 0, fmt.Errorf("parse Upload-Offset: %w", err)
	}
	return newOffset, nil
}

func (c *tusClient) head(uploadURL string) (int64, int64, error) {
	c.t.Helper()
	req, err := http.NewRequest("HEAD", uploadURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, 0, fmt.Errorf("tus head status %d: %s", resp.StatusCode, string(body))
	}
	var offset, length int64
	fmt.Sscanf(resp.Header.Get("Upload-Offset"), "%d", &offset)
	fmt.Sscanf(resp.Header.Get("Upload-Length"), "%d", &length)
	return offset, length, nil
}

func encodeMetadata(values map[string]string) string {
	pairs := []string{}
	for key, value := range values {
		pairs = append(pairs, fmt.Sprintf("%s %s", key, base64.StdEncoding.EncodeToString([]byte(value))))
	}
	return joinComma(pairs)
}

func joinComma(parts []string) string {
	result := ""
	for i, segment := range parts {
		if i > 0 {
			result += ","
		}
		result += segment
	}
	return result
}

// waitForFile polls for a file to appear at the given path. Returns its bytes
// or fails the test after timeout. Used for waiting on the post-completion hook.
func waitForFile(t *testing.T, path string, timeout time.Duration) []byte {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			return data
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("file did not appear at %s within %s", path, timeout)
	return nil
}

// setupTusTestServer creates a live httptest.Server with the full router so
// tests exercise the real tus integration end to end (including the goroutine
// that watches CompleteUploads).
func setupTusTestServer(t *testing.T) (*httptest.Server, string, string) {
	t.Helper()
	router, cfg, _, dataDir := setupFilesTestServer(t)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	token, err := auth.IssueToken("admin1", cfg.PrivateKey)
	require.NoError(t, err)
	return srv, token, dataDir
}

func TestTus_SingleChunkUpload_FinalizesFile(t *testing.T) {
	srv, token, dataDir := setupTusTestServer(t)
	client := newTusClient(t, srv.URL, token)

	payload := []byte("hello tus world from a single chunk")
	uploadURL, err := client.create(int64(len(payload)), map[string]string{
		"filename": "single.txt",
		"loc":      "",
	})
	require.NoError(t, err)

	offset, err := client.patch(uploadURL, 0, payload)
	require.NoError(t, err)
	require.Equal(t, int64(len(payload)), offset)

	data := waitForFile(t, filepath.Join(dataDir, "single.txt"), 3*time.Second)
	require.Equal(t, payload, data)
}

func TestTus_MultiChunkUpload_FinalizesFile(t *testing.T) {
	srv, token, dataDir := setupTusTestServer(t)
	client := newTusClient(t, srv.URL, token)

	const totalSize = 256 * 1024 // 256KB
	const chunkSize = 32 * 1024
	payload := make([]byte, totalSize)
	for i := range payload {
		payload[i] = byte((i * 31) % 251) // deterministic but non-trivial
	}
	expectedHash := sha256.Sum256(payload)

	uploadURL, err := client.create(totalSize, map[string]string{
		"filename": "multi.bin",
		"loc":      "subdir",
	})
	require.NoError(t, err)

	var offset int64
	for offset < totalSize {
		end := offset + chunkSize
		if end > totalSize {
			end = totalSize
		}
		newOffset, err := client.patch(uploadURL, offset, payload[offset:end])
		require.NoError(t, err)
		require.Equal(t, end, newOffset)
		offset = newOffset
	}

	data := waitForFile(t, filepath.Join(dataDir, "subdir", "multi.bin"), 5*time.Second)
	require.Equal(t, totalSize, len(data))
	gotHash := sha256.Sum256(data)
	require.Equal(t, hex.EncodeToString(expectedHash[:]), hex.EncodeToString(gotHash[:]),
		"file integrity check failed after multi-chunk upload")
}

func TestTus_ResumeAfterInterruption(t *testing.T) {
	srv, token, dataDir := setupTusTestServer(t)
	client := newTusClient(t, srv.URL, token)

	const totalSize = 100 * 1024 // 100KB
	const chunkSize = 16 * 1024
	payload := make([]byte, totalSize)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	uploadURL, err := client.create(totalSize, map[string]string{
		"filename": "resumable.bin",
		"loc":      "",
	})
	require.NoError(t, err)

	// Upload first 3 chunks (~48KB) then "pause"
	var offset int64
	for i := 0; i < 3; i++ {
		end := offset + chunkSize
		newOffset, err := client.patch(uploadURL, offset, payload[offset:end])
		require.NoError(t, err)
		offset = newOffset
	}
	require.Equal(t, int64(3*chunkSize), offset)

	// HEAD verifies offset survived
	resumeOffset, totalReported, err := client.head(uploadURL)
	require.NoError(t, err)
	require.Equal(t, offset, resumeOffset)
	require.Equal(t, int64(totalSize), totalReported)

	// New client (simulating client process restart) resumes from reported offset
	resumeClient := newTusClient(t, srv.URL, token)
	for offset < totalSize {
		end := offset + chunkSize
		if end > totalSize {
			end = totalSize
		}
		newOffset, err := resumeClient.patch(uploadURL, offset, payload[offset:end])
		require.NoError(t, err)
		offset = newOffset
	}

	data := waitForFile(t, filepath.Join(dataDir, "resumable.bin"), 5*time.Second)
	require.Equal(t, payload, data)
}

func TestTus_ConcurrentUploads(t *testing.T) {
	srv, token, dataDir := setupTusTestServer(t)

	const fileCount = 5
	const fileSize = 64 * 1024
	const chunkSize = 16 * 1024

	var waitGroup sync.WaitGroup
	hashes := make(map[string][32]byte)
	var hashMutex sync.Mutex
	errors := make(chan error, fileCount)

	for index := 0; index < fileCount; index++ {
		waitGroup.Add(1)
		go func(idx int) {
			defer waitGroup.Done()
			client := newTusClient(t, srv.URL, token)
			payload := make([]byte, fileSize)
			for i := range payload {
				payload[i] = byte((i + idx*7) % 256)
			}
			filename := fmt.Sprintf("concurrent-%d.bin", idx)

			uploadURL, err := client.create(fileSize, map[string]string{
				"filename": filename,
				"loc":      "concurrent",
			})
			if err != nil {
				errors <- fmt.Errorf("create %d: %w", idx, err)
				return
			}

			var offset int64
			for offset < fileSize {
				end := offset + chunkSize
				if end > fileSize {
					end = fileSize
				}
				newOffset, err := client.patch(uploadURL, offset, payload[offset:end])
				if err != nil {
					errors <- fmt.Errorf("patch %d offset %d: %w", idx, offset, err)
					return
				}
				offset = newOffset
			}

			hashMutex.Lock()
			hashes[filename] = sha256.Sum256(payload)
			hashMutex.Unlock()
		}(index)
	}

	waitGroup.Wait()
	close(errors)
	for err := range errors {
		t.Fatalf("concurrent upload error: %v", err)
	}

	// Verify each file landed and matches expected hash
	for filename, expectedHash := range hashes {
		data := waitForFile(t, filepath.Join(dataDir, "concurrent", filename), 10*time.Second)
		gotHash := sha256.Sum256(data)
		require.Equal(t, hex.EncodeToString(expectedHash[:]), hex.EncodeToString(gotHash[:]),
			"hash mismatch for %s", filename)
	}
}

func TestTus_RejectsBadOffset(t *testing.T) {
	srv, token, _ := setupTusTestServer(t)
	client := newTusClient(t, srv.URL, token)

	uploadURL, err := client.create(100, map[string]string{
		"filename": "bad-offset.bin",
		"loc":      "",
	})
	require.NoError(t, err)

	// PATCH at offset=50 when server expects offset=0
	req, _ := http.NewRequest("PATCH", uploadURL, bytes.NewReader([]byte("data")))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Content-Type", "application/offset+octet-stream")
	req.Header.Set("Upload-Offset", "50")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusConflict, resp.StatusCode, "tus must return 409 for offset mismatch")
}

// TestTus_PostHookSkipsWithoutFilenameMetadata: when client omits filename,
// the post-hook should log an error but not crash. File stays in staging.
func TestTus_PostHookHandlesMissingMetadata(t *testing.T) {
	srv, token, dataDir := setupTusTestServer(t)
	client := newTusClient(t, srv.URL, token)

	payload := []byte("orphan")
	uploadURL, err := client.create(int64(len(payload)), map[string]string{
		"loc": "no-filename",
		// intentionally no "filename"
	})
	require.NoError(t, err)
	_, err = client.patch(uploadURL, 0, payload)
	require.NoError(t, err)

	// The post-hook should NOT create a file at any expected path. Wait briefly
	// then confirm no file exists.
	time.Sleep(200 * time.Millisecond)
	entries, err := os.ReadDir(filepath.Join(dataDir, "no-filename"))
	if err == nil {
		require.Empty(t, entries, "missing-filename upload should not finalize to nas-data")
	}
}

// TestTusUploadCreatesNestedFolders: tus uploads with multi-level `loc` metadata
// must auto-create intermediate directories under NAS_DATA_DIR via the post-hook
// (`os.MkdirAll(filepath.Dir(target), 0o755)` in finalizeUpload). Each sub-test
// runs a distinct upload through the same server to confirm no interference.
func TestTusUploadCreatesNestedFolders(t *testing.T) {
	srv, token, dataDir := setupTusTestServer(t)

	// expectedDirMode is the permission bits MkdirAll is invoked with in
	// finalizeUpload. On Windows the actual perm bits as returned by Stat
	// can differ from the requested mode; we only assert on the unix-mask
	// of executable+read bits we actually request.
	const expectedDirMode os.FileMode = 0o755

	type subCase struct {
		name          string
		loc           string
		filename      string
		payload       []byte
		intermediates []string // relative paths under dataDir that must exist
		finalDir      string   // relative path under dataDir where file lives
	}
	cases := []subCase{
		{
			name:          "three-level",
			loc:           "/A/B/C",
			filename:      "nested.txt",
			payload:       []byte("three deep"),
			intermediates: []string{"A", filepath.Join("A", "B"), filepath.Join("A", "B", "C")},
			finalDir:      filepath.Join("A", "B", "C"),
		},
		{
			name:     "five-level",
			loc:      "/one/two/three/four/five",
			filename: "deep.bin",
			payload:  []byte("five deep payload"),
			intermediates: []string{
				"one",
				filepath.Join("one", "two"),
				filepath.Join("one", "two", "three"),
				filepath.Join("one", "two", "three", "four"),
				filepath.Join("one", "two", "three", "four", "five"),
			},
			finalDir: filepath.Join("one", "two", "three", "four", "five"),
		},
		{
			name:          "root-level",
			loc:           "",
			filename:      "root.txt",
			payload:       []byte("at root"),
			intermediates: nil,
			finalDir:      "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := newTusClient(t, srv.URL, token)
			uploadURL, err := client.create(int64(len(tc.payload)), map[string]string{
				"filename": tc.filename,
				"loc":      tc.loc,
			})
			require.NoError(t, err)

			offset, err := client.patch(uploadURL, 0, tc.payload)
			require.NoError(t, err)
			require.Equal(t, int64(len(tc.payload)), offset)

			finalPath := filepath.Join(dataDir, tc.finalDir, tc.filename)
			data := waitForFile(t, finalPath, 5*time.Second)
			require.Equal(t, tc.payload, data, "uploaded content mismatch for %s", tc.name)

			for _, rel := range tc.intermediates {
				dir := filepath.Join(dataDir, rel)
				stat, err := os.Stat(dir)
				require.NoError(t, err, "intermediate dir missing: %s", rel)
				require.True(t, stat.IsDir(), "expected directory at %s", rel)
				perm := stat.Mode().Perm()
				if runtime.GOOS == "windows" {
					// On Windows os.MkdirAll mode bits are best-effort and
					// Stat reports synthesized bits (typically 0o777). Assert
					// owner readable+writable+executable as the invariant.
					require.Equal(t, os.FileMode(0o700), perm&0o700,
						"directory %s should be owner-rwx, got %#o", rel, perm)
				} else {
					require.Equal(t, expectedDirMode, perm,
						"directory %s should be %#o, got %#o", rel, expectedDirMode, perm)
				}
			}
		})
	}
}

// TestTus_PathTraversalBlocked: a malicious filename or loc that escapes
// NAS_DATA_DIR must be rejected by the post-hook via SafeJoin.
func TestTus_PathTraversalBlocked(t *testing.T) {
	srv, token, dataDir := setupTusTestServer(t)
	client := newTusClient(t, srv.URL, token)

	payload := []byte("malicious")
	uploadURL, err := client.create(int64(len(payload)), map[string]string{
		"filename": "../../../etc/passwd",
		"loc":      "",
	})
	require.NoError(t, err)
	_, err = client.patch(uploadURL, 0, payload)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	// The traversed path must NOT exist
	parent := filepath.Dir(filepath.Dir(filepath.Dir(dataDir)))
	_, err = os.Stat(filepath.Join(parent, "passwd"))
	require.True(t, os.IsNotExist(err), "path traversal must be blocked")
}
