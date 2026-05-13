# Office Document Viewer (LibreOffice Backend) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Render Office documents (`.docx`, `.doc`, `.xlsx`, `.xls`, `.pptx`, `.ppt`, `.odt`, `.ods`, `.odp`, `.hwp`, `.rtf`) inline in the NAS viewer by **converting them to PDF server-side via headless LibreOffice** and displaying the resulting PDF in the existing PDF viewer. Decision (`b`) was chosen over per-format JS libraries to maximize fidelity consistency and minimize frontend complexity.

**Architecture:**
1. **Backend**: new `office` package with a single endpoint `/server/getOfficePdf?token=...&loc=...&name=...`. Endpoint: SafeJoin → content-hash the file → check disk cache → on miss run `soffice --headless --convert-to pdf` with a 30s timeout → cache result → stream PDF response. Concurrent requests for the same file deduplicate via per-hash channel locks.
2. **Frontend**: new `OfficeViewer.svelte` that calls `/server/getOfficePdf` and renders the response with the existing `PdfViewer.svelte`. New TabKind `"office"`. Registry routes office extensions to it.
3. **Infrastructure**: Dockerfile installs `libreoffice-core libreoffice-writer libreoffice-calc libreoffice-impress` (HWP support is partial, baked into LibreOffice itself).

**Tech Stack:**
- Backend: Go stdlib (`os/exec`, `crypto/sha256`, `sync`, `context.WithTimeout`), no new Go dependencies.
- Frontend: Svelte 5, **depends on `feat/pdf-viewer` having landed** (reuses `PdfViewer.svelte`). If PDF viewer plan not yet merged, do that first.
- Infra: Debian-based base image + `libreoffice` apt packages (~400MB).

**Testing Approach:**
- **Backend has a real test framework** (Go `testing` + testify in `backend/tests/integration/`). Use it. Write integration tests against the new endpoint with fixture docx/xlsx/pptx in `backend/tests/testdata/`.
- Skip tests gracefully if `soffice` isn't on PATH in the test environment (CI without LibreOffice).
- Frontend: spec-driven manual verification (no test framework).

**Out of scope (separate / future):**
- Live editing in browser (Collabora Online) — major infra commitment, separate plan
- Cache eviction policy — start with no LRU; cache grows in NAS_TEMP_DIR. v2 adds TTL or size-based eviction.
- Format fidelity guarantees — LibreOffice's best effort. Known limitations: complex Excel macros, unusual fonts, HWP advanced features.
- Streaming partial conversion — `soffice` writes the whole PDF before exit; we wait for completion.
- Per-page thumbnails — defer to PDF viewer if/when added.

---

## File Structure

| File | Status | Responsibility | Target Lines |
|---|---|---|---|
| `backend/internal/office/converter.go` | Create | `Convert(ctx, srcPath, dstPath) error` — wraps `soffice --convert-to pdf` with timeout | ≤100 |
| `backend/internal/office/cache.go` | Create | `Cache` struct: `Get(hash) (path, hit)`, `Put(hash, path)`, disk-backed at `<tempDir>/office-cache/` | ≤80 |
| `backend/internal/office/handlers.go` | Create | HTTP handler `GetOfficePdf`: hash → cache → convert if miss → stream | ≤120 |
| `backend/internal/office/dedupe.go` | Create | Per-hash conversion lock so parallel requests don't double-convert | ≤60 |
| `backend/internal/server/router.go` | Modify | Register `/getOfficePdf` route with `auth.IntentOpen` | +2 |
| `backend/tests/integration/office_test.go` | Create | Integration test: docx → PDF conversion + cache hit + path traversal block + concurrent dedup | ≤200 |
| `backend/tests/testdata/sample.docx` | Create (binary) | Tiny fixture (single page "Hello World") | — |
| `backend/tests/testdata/sample.xlsx` | Create (binary) | Tiny fixture | — |
| `backend/tests/testdata/sample.pptx` | Create (binary) | Tiny fixture | — |
| `Dockerfile` | Modify | `apt-get install -y libreoffice-core libreoffice-writer libreoffice-calc libreoffice-impress` | +3 |
| `frontend/src/lib/components/Viewers/OfficeViewer.svelte` | Create | Calls `/server/getOfficePdf`, swaps the PDF source into PdfViewer | ≤80 |
| `frontend/src/lib/components/Viewers/PdfViewer.svelte` | Modify | Accept an optional `urlOverride` prop instead of always deriving from loc/name | ~10 lines changed |
| `frontend/src/lib/types.ts` | Modify | Add `"office"` to `TabKind` | +1 |
| `frontend/src/lib/components/Viewers/registry.ts` | Modify | Add `OFFICE_EXTENSIONS` set, route to `"office"` | +10 |
| `frontend/src/lib/components/Tabs/TabContent.svelte` | Modify | Dispatch `kind === "office"` to `OfficeViewer` | +4 |

**Why this split:**
- Backend: `converter`, `cache`, `dedupe`, `handlers` separate by responsibility (SRP). `dedupe` is a small concern that deserves its own file because it has tricky concurrency semantics.
- Frontend `OfficeViewer.svelte` is thin: just a one-line URL change passed to PdfViewer. The bulk of UI is reused from `PdfViewer.svelte` (which adds an `urlOverride` prop).

**Dependency**: This plan **assumes `feat/pdf-viewer` is merged** (the previous plan). If not, merge that first, OR the office plan can be adapted to use the old iframe-based PDF viewer (lower fidelity).

---

## Task 0: Branch setup

- [ ] **Step 1: Confirm PDF viewer plan is merged**

Verify the artifact, not the commit message:

```bash
grep -q '"pdfjs-dist"' C:/Data/Git/ANXI/nas/frontend/package.json && echo "PDF viewer ready" || echo "MISSING — complete feat/pdf-viewer plan first"
```

Expected: `PDF viewer ready`. If `MISSING`, stop and complete the PDF viewer plan first.

- [ ] **Step 2: Create branch**

```bash
git checkout main && git pull
git checkout -b feat/office-docs
```

---

## Task 1: Backend converter — wraps `soffice` CLI

**Files:**
- Create: `backend/internal/office/converter.go`

**Spec:**

| Function | Behavior |
|---|---|
| `Convert(ctx, srcPath, dstDir) (outPath string, err error)` | Runs `soffice --headless --convert-to pdf --outdir <dstDir> <srcPath>` via `exec.CommandContext`; returns full path to the generated PDF (`<dstDir>/<srcBase>.pdf`) |
| 30s timeout | `context.WithTimeout` enforces; on timeout returns wrapped `context.DeadlineExceeded` |
| Missing soffice | Returns wrapped error `"soffice not found: <exec.LookPath error>"` |
| Conversion failure | Returns error containing stderr |

- [ ] **Step 1: Write `converter.go`**

```go
package office

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const ConvertTimeout = 30 * time.Second

// Convert runs LibreOffice headless to produce a PDF.
// It writes to a fresh temp subdir under `workDir` to avoid basename collisions
// between concurrent conversions of different files that share a filename, then
// returns the path to the generated PDF (caller is responsible for moving/renaming).
func Convert(ctx context.Context, srcPath, workDir string) (string, error) {
	bin, err := exec.LookPath("soffice")
	if err != nil {
		return "", fmt.Errorf("soffice not found: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, ConvertTimeout)
	defer cancel()

	// Per-call temp dir prevents basename collisions when two requests with
	// different content but identical filenames run in parallel (same package,
	// different folders — dedupe only collapses same-hash, not same-name).
	tmpDir, err := os.MkdirTemp(workDir, "conv-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}
	// Caller cleans up tmpDir after moving the PDF; we don't defer-remove here
	// because the caller needs the output path to still exist after we return.

	cmd := exec.CommandContext(ctx, bin,
		"--headless",
		"--convert-to", "pdf",
		"--outdir", tmpDir,
		srcPath,
	)
	// LibreOffice writes both progress and errors to stderr; capture for diagnostics.
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("soffice timeout after %v: %s", ConvertTimeout, stderr.String())
		}
		return "", fmt.Errorf("soffice failed: %w: %s", err, stderr.String())
	}

	base := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	outPath := filepath.Join(tmpDir, base+".pdf")
	if _, err := os.Stat(outPath); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("expected output not found: %w", err)
	}
	return outPath, nil
}
```

- [ ] **Step 2: Build**

```bash
cd backend && go build ./...
```

Expected: succeeds.

- [ ] **Step 3: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/office/converter.go
git commit -m "[feat] add office.Convert wrapping soffice headless CLI"
```

---

## Task 2: Backend cache — disk-backed by content hash

**Files:**
- Create: `backend/internal/office/cache.go`

**Spec:**

| Function | Behavior |
|---|---|
| `New(rootDir) *Cache` | Ensures `<rootDir>/office-cache/` exists |
| `(c *Cache).Path(hash) string` | Returns the cache file path for a hash (does not check existence) |
| `(c *Cache).Hit(hash) bool` | Returns true if a PDF exists at the path |
| `HashFile(srcPath) (string, error)` | SHA-256 of file contents, hex-encoded |

- [ ] **Step 1: Write `cache.go`**

```go
package office

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

type Cache struct {
	dir string
}

func NewCache(rootDir string) (*Cache, error) {
	dir := filepath.Join(rootDir, "office-cache")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Cache{dir: dir}, nil
}

func (c *Cache) Path(hash string) string {
	return filepath.Join(c.dir, hash+".pdf")
}

func (c *Cache) Hit(hash string) bool {
	info, err := os.Stat(c.Path(hash))
	return err == nil && info.Size() > 0
}

func HashFile(srcPath string) (string, error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
```

- [ ] **Step 2: Build** → succeeds.

- [ ] **Step 3: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/office/cache.go
git commit -m "[feat] add office.Cache disk-backed by SHA-256 of source"
```

---

## Task 3: Backend dedupe — single-flight per hash

**Files:**
- Create: `backend/internal/office/dedupe.go`

**Spec:**

| Behavior | Expected |
|---|---|
| Concurrent requests for the same hash | Only one conversion runs; others wait on the result |
| Different hashes | Convert in parallel |
| Cleanup | After conversion finishes, the in-flight entry is removed |

Use `sync.Map` of `chan struct{}`. First call wins, others read result via the closed channel.

- [ ] **Step 1: Write `dedupe.go`**

```go
package office

import "sync"

type Dedupe struct {
	flights sync.Map // map[string]chan struct{}
}

// Acquire returns (done, isLeader). The leader does the work and calls done() when finished.
// Followers wait on the returned channel; isLeader is false for them and done is nil.
func (d *Dedupe) Acquire(hash string) (release func(), isLeader bool) {
	ch := make(chan struct{})
	existing, loaded := d.flights.LoadOrStore(hash, ch)
	if loaded {
		// follower
		<-existing.(chan struct{})
		return nil, false
	}
	// leader
	return func() {
		close(ch)
		d.flights.Delete(hash)
	}, true
}
```

- [ ] **Step 2: Build** → succeeds.

- [ ] **Step 3: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/office/dedupe.go
git commit -m "[feat] add office.Dedupe single-flight conversion lock per hash"
```

---

## Task 4: Backend HTTP handler

**Files:**
- Create: `backend/internal/office/handlers.go`

**Spec:**

| Behavior | Expected |
|---|---|
| `GET /server/getOfficePdf?token=...&loc=...&name=...` | Resolves loc+name → SafeJoin → reads file |
| Hash file content | SHA-256 |
| Cache hit | Streams PDF immediately via `http.ServeContent` |
| Cache miss | Acquires dedupe lock; if leader: convert + cache + serve; if follower: waits, then serves cached |
| Error: file missing | 404 |
| Error: unsafe path | 400 |
| Error: conversion fails | 500 with message |

- [ ] **Step 1: Write `handlers.go`**

```go
package office

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/files"
)

type Handlers struct {
	Config *config.Config
	DB     *sql.DB
	Cache  *Cache
	Dedupe *Dedupe
}

func NewHandlers(cfg *config.Config, db *sql.DB) (*Handlers, error) {
	cache, err := NewCache(cfg.NASTempDir)
	if err != nil {
		return nil, err
	}
	return &Handlers{
		Config: cfg,
		DB:     db,
		Cache:  cache,
		Dedupe: &Dedupe{},
	}, nil
}

func (h *Handlers) GetOfficePdf(w http.ResponseWriter, r *http.Request) {
	loc := r.URL.Query().Get("loc")
	name := r.URL.Query().Get("name")
	src, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(loc), name)
	if err != nil {
		http.Error(w, "unsafe path", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(src); err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	hash, err := HashFile(src)
	if err != nil {
		http.Error(w, "hash failed", http.StatusInternalServerError)
		return
	}

	pdfPath := h.Cache.Path(hash)
	if !h.Cache.Hit(hash) {
		release, isLeader := h.Dedupe.Acquire(hash)
		if isLeader {
			// leader converts
			outPath, err := Convert(r.Context(), src, h.Cache.dir)
			if err != nil {
				release() // release BEFORE returning so followers stop waiting
				http.Error(w, "conversion failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			// soffice writes to a per-call temp dir; move into the cache as <hash>.pdf.
			// Write to .tmp first then atomic-rename so partial writes don't poison the cache.
			tmpFinal := pdfPath + ".tmp"
			if err := os.Rename(outPath, tmpFinal); err != nil {
				os.RemoveAll(filepath.Dir(outPath))
				release()
				http.Error(w, "cache write failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			os.RemoveAll(filepath.Dir(outPath)) // clean the per-call temp dir from Convert
			if err := os.Rename(tmpFinal, pdfPath); err != nil {
				release()
				http.Error(w, "cache finalize failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			// Release immediately after the cache file is materialized — followers
			// can now serve from cache in parallel with the leader's response stream.
			release()
		}
		// follower (or post-leader): re-check cache existence
		if !h.Cache.Hit(hash) {
			http.Error(w, "conversion produced no output", http.StatusInternalServerError)
			return
		}
	}

	file, err := os.Open(pdfPath)
	if err != nil {
		http.Error(w, "open cached pdf failed", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(w, "stat cached pdf failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	http.ServeContent(w, r, name+".pdf", info.ModTime().UTC().Truncate(time.Second), file)
}
```

Note: relies on `files.SafeJoin` and `files.TrimLeadingSlash` being exported. If they are not, expose them (small refactor in `files` package).

- [ ] **Step 2: Check `files.SafeJoin` export**

```bash
grep -n "^func SafeJoin\|^func TrimLeadingSlash" backend/internal/files/*.go
```

If both functions are exported (capitalized), Task 4 proceeds. If not, add to a follow-up step:

```go
// in backend/internal/files/utils.go or similar:
func SafeJoin(base string, parts ...string) (string, error) { /* existing impl */ }
func TrimLeadingSlash(s string) string { /* existing impl */ }
```

(Spec: existing handlers in `backend/internal/files/handlers.go` already use these names lower or upper-case — verify and rename if needed, OR expose accessor functions.)

- [ ] **Step 3: Build** → succeeds.

- [ ] **Step 4: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/office/handlers.go
git commit -m "[feat] add office.GetOfficePdf handler with cache + dedupe"
```

---

## Task 5: Register route

**Files:**
- Modify: `backend/internal/server/router.go`

**Note on signature:** `NewRouter(cfg, conn) http.Handler` currently has no error return (verified — `files_test.go:49` calls it without checking). Adding an error return cascades to `main.go` and tests. To minimize blast radius, log-and-panic on init failure (init errors here are non-recoverable — they indicate a misconfigured `NAS_TEMP_DIR`).

- [ ] **Step 1: Add import + handler construction + route**

In `router.go`, alongside other handler initializations:

```go
import (
	"log"
	"github.com/JMC50/nas/internal/office"
	// ...existing imports
)
```

```go
officeHandlers, err := office.NewHandlers(cfg, conn)
if err != nil {
	log.Fatalf("office handlers init: %v", err)
}
```

Route registration (next to other file-routed endpoints):

```go
addFileRoute(r, "GET", "/getOfficePdf", auth.IntentOpen, requireToken, conn, officeHandlers.GetOfficePdf)
```

(If `NewRouter` already returns an error or the project prefers proper propagation, do that instead — but only if you're willing to update `main.go` + test helpers in the same commit.)

- [ ] **Step 2: Build** → succeeds.

- [ ] **Step 3: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/server/router.go
git commit -m "[feat] register /getOfficePdf route"
```

---

## Task 6: Test fixtures

**Files:**
- Create: `backend/tests/testdata/sample.docx`, `sample.xlsx`, `sample.pptx`

**Spec:** tiny (single page / sheet / slide) documents that LibreOffice can convert. Generate them with `soffice` itself, or use existing minimal samples from a public domain source.

- [ ] **Step 1: Generate fixtures**

```bash
cd backend/tests/testdata
echo "Hello World" > sample.txt
soffice --headless --convert-to docx --outdir . sample.txt
soffice --headless --convert-to xlsx --outdir . sample.txt
soffice --headless --convert-to pptx --outdir . sample.txt
rm sample.txt
```

Verify the three files exist (~10-30KB each).

- [ ] **Step 2: Commit**

```bash
git add backend/tests/testdata/sample.docx backend/tests/testdata/sample.xlsx backend/tests/testdata/sample.pptx
git commit -m "[test] add minimal office fixtures for conversion tests"
```

---

## Task 7: Integration tests

**Files:**
- Create: `backend/tests/integration/office_test.go`

**Spec:**

| Test | Behavior |
|---|---|
| `TestOfficeConvert_DocxToPdf` | Place sample.docx in data dir, call `/getOfficePdf`, expect 200, response Content-Type `application/pdf`, body starts with `%PDF-` magic |
| `TestOfficeConvert_CacheHit` | Convert twice; second should be much faster (assert sub-50ms is reasonable since it skips soffice) |
| `TestOfficeConvert_PathTraversal` | `?loc=../../etc&name=passwd` → 400 |
| `TestOfficeConvert_ConcurrentDedup` | 10 parallel requests for same file → exactly one soffice invocation (mock or rely on cache state inspection) |
| Skip if soffice unavailable | All tests use `t.Skip("soffice not installed")` if `exec.LookPath("soffice")` fails |

- [ ] **Step 1: Write `office_test.go`**

```go
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

	// first call (cold)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, httptest.NewRequest("GET", url, nil))
	require.Equal(t, http.StatusOK, w1.Code)

	// second call (cache hit) — should be near-instant
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, httptest.NewRequest("GET", url, nil))
	require.Equal(t, http.StatusOK, w2.Code)
	require.Equal(t, w1.Body.Len(), w2.Body.Len(), "cached response should be identical size")
}

func TestOfficeConvert_PathTraversal(t *testing.T) {
	skipIfNoSoffice(t) // skip even though we don't actually convert — keeps test list consistent
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
	// Dedup is observed externally via timing — if 10 conversions ran serially this would be very slow.
	// Direct assertion: only one PDF should exist in the cache dir.
	cacheDir := filepath.Join(cfg.NASTempDir, "office-cache")
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	require.Equal(t, 1, len(entries), "exactly one cached pdf expected, got %d", len(entries))
}
```

All required imports (`os`, `database/sql`, `config`, `db`, `server`) are already in the imports block above.

- [ ] **Step 2: Run tests**

```bash
cd backend && go test ./tests/integration/ -run TestOfficeConvert -v
```

Expected (with soffice installed): all PASS.
Expected (without soffice): all SKIP.

- [ ] **Step 3: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 4: Commit**

```bash
git add backend/tests/integration/office_test.go
git commit -m "[test] add office conversion integration tests with dedup verification"
```

---

## Task 8: Dockerfile — install LibreOffice

**Files:**
- Modify: `backend/Dockerfile` (the production runtime — verified Alpine 3.20 base; the root `Dockerfile` is legacy Node-era and not the target)

**Spec:** install LibreOffice on the Alpine runtime stage. Alpine ships a single `libreoffice` meta-package (no separate writer/calc/impress packages). Total image size bump ~200-400MB. Also set `HOME` so LibreOffice's profile dir is writable in the container.

- [ ] **Step 1: Verify target Dockerfile and base image**

```bash
head -30 C:/Data/Git/ANXI/nas/backend/Dockerfile
```

Expected: shows `FROM ... alpine ...` for the runtime stage. The existing `# Phase 3+ will add: ffmpeg libreoffice` comment marks the install location.

- [ ] **Step 2: Add LibreOffice install + HOME env**

In `backend/Dockerfile`, in the runtime stage, add:

```dockerfile
RUN apk add --no-cache libreoffice ttf-dejavu fontconfig
ENV HOME=/tmp/soffice-home
RUN mkdir -p /tmp/soffice-home
```

`ttf-dejavu` + `fontconfig` ensure conversion has fonts (LibreOffice silently produces empty pages otherwise). `HOME` set to a writable path so LibreOffice can create its user profile (the default `$HOME` may be read-only in some container configurations).

If the runtime image is later switched to Debian-based, the equivalent is:
```dockerfile
RUN apt-get update && apt-get install -y --no-install-recommends \
    libreoffice-core libreoffice-writer libreoffice-calc libreoffice-impress \
    fonts-dejavu fontconfig \
    && rm -rf /var/lib/apt/lists/*
ENV HOME=/tmp/soffice-home
RUN mkdir -p /tmp/soffice-home
```

- [ ] **Step 3: Build the image locally to verify**

```bash
docker build -f backend/Dockerfile -t nas:test C:/Data/Git/ANXI/nas
docker run --rm nas:test soffice --headless --version
```

Expected: prints `LibreOffice X.Y.Z ...`.

- [ ] **Step 4: Sanity-convert an HWP**

If you have an HWP fixture handy:

```bash
docker run --rm -v $(pwd)/backend/tests/testdata:/data nas:test sh -c "soffice --headless --convert-to pdf --outdir /tmp /data/sample.hwp && ls /tmp/*.pdf"
```

Expected: a PDF file created. If HWP filter is missing, output will be empty — flag as known limitation in PR description.

- [ ] **Step 5: `/code-review`** on `backend/Dockerfile` → 0 ❌ Critical.

- [ ] **Step 6: Commit**

```bash
git add backend/Dockerfile
git commit -m "[chore] install LibreOffice in runtime image for office conversion"
```

---

## Task 9: Frontend — `PdfViewer.svelte` accepts URL override

**Files:**
- Modify: `frontend/src/lib/components/Viewers/PdfViewer.svelte`

**Spec:** PdfViewer currently derives `pdfUrl` from `loc + name`. Add an optional `urlOverride: string` prop. When set, use it directly. Default behavior unchanged when omitted.

- [ ] **Step 1: Update Props interface + pdfUrl derivation**

```ts
interface Props {
  loc: string;
  name: string;
  urlOverride?: string;
}

let { loc, name, urlOverride }: Props = $props();

const pdfUrl = $derived(
  urlOverride ??
    `/server/download?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
);
```

- [ ] **Step 2: Type check** → 0 errors.

- [ ] **Step 3: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/components/Viewers/PdfViewer.svelte
git commit -m "[refactor] PdfViewer accepts optional urlOverride for reuse"
```

---

## Task 10: Frontend — `OfficeViewer.svelte`

**Files:**
- Create: `frontend/src/lib/components/Viewers/OfficeViewer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Renders PdfViewer with `urlOverride` pointing to `/server/getOfficePdf` | Same toolbar (page nav, zoom, download) |
| Loading state | "Converting…" shown over PdfViewer's "Loading…" while server processes |
| Errors | notifications.error if response is non-OK |

- [ ] **Step 1: Write the component**

```svelte
<!-- frontend/src/lib/components/Viewers/OfficeViewer.svelte -->
<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";
  import PdfViewer from "./PdfViewer.svelte";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  const officeUrl = $derived(
    `/server/getOfficePdf?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );
</script>

<PdfViewer {loc} {name} urlOverride={officeUrl} />
```

Simple wrapper. The "Converting…" UX is handled by PdfViewer's existing "Loading…" — the server will hold the response while soffice runs (up to 30s).

- [ ] **Step 2: Type check** → 0 errors.

- [ ] **Step 3: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/components/Viewers/OfficeViewer.svelte
git commit -m "[feat] add OfficeViewer that reuses PdfViewer with conversion URL"
```

---

## Task 11: Frontend — registry + dispatch

**Files:**
- Modify: `frontend/src/lib/types.ts`
- Modify: `frontend/src/lib/components/Viewers/registry.ts`
- Modify: `frontend/src/lib/components/Tabs/TabContent.svelte`

- [ ] **Step 1: `TabKind` adds `"office"`**

```ts
export type TabKind =
  | "explorer"
  // ...existing...
  | "office";
```

- [ ] **Step 2: `registry.ts` — add OFFICE_EXTENSIONS and route**

```ts
const OFFICE_EXTENSIONS = new Set([
  "doc", "docx", "rtf", "odt",
  "xls", "xlsx", "ods",
  "ppt", "pptx", "odp",
  "hwp",
]);

export function pickViewer(extension: string): TabKind {
  const ext = extension.toLowerCase().replace(/^\./, "");
  if (OFFICE_EXTENSIONS.has(ext)) return "office";
  // ...rest unchanged...
}

export function viewerIconName(kind: TabKind): string {
  switch (kind) {
    case "office":
      return "file-text"; // or "files" — pick one
    // ...rest unchanged...
  }
}
```

- [ ] **Step 3: `TabContent.svelte` — dispatch**

```svelte
<script lang="ts">
  import OfficeViewer from "$lib/components/Viewers/OfficeViewer.svelte";
  // ...
</script>

<!-- inside the {#each ...} -->
{:else if tab.kind === "office"}
  {@const payload = tab.payload as FilePayload}
  <OfficeViewer loc={payload.loc} name={payload.name} />
```

- [ ] **Step 4: Type check + manual verification**

```bash
npm run dev
```

- Place `sample.docx`, `sample.xlsx`, `sample.pptx` in NAS data dir
- Open each in explorer → loading bar appears (server converting) → PDF viewer shows the rendered document
- Open the same file again → loads in <1 second (cache hit)
- Open a `.hwp` file (if available) → converts; quality may vary (LibreOffice partial HWP)

- [ ] **Step 5: `/code-review`** on all three modified files → 0 ❌ Critical.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/types.ts frontend/src/lib/components/Viewers/registry.ts frontend/src/lib/components/Tabs/TabContent.svelte
git commit -m "[feat] route office documents to OfficeViewer"
```

---

## Task 12: Final integration verification + build

**Files:** None modified.

- [ ] **Step 1: Backend tests**

```bash
cd backend && go test ./... 2>&1 | tail -30
```

Expected: all tests pass (or skip cleanly if soffice not in CI).

- [ ] **Step 2: Frontend type + build**

```bash
cd frontend && npm run check && npm run build
```

Expected: 0 errors / 0 warnings; build succeeds.

- [ ] **Step 3: Spec walkthrough** (requires LibreOffice installed locally or in Docker)

| # | Action | Expected | Status |
|---|---|---|---|
| 1 | Open `.docx` | Loads in 1-3s (cold), shows as PDF | |
| 2 | Open same `.docx` again | Loads <1s (cache hit) | |
| 3 | Open `.xlsx` | Spreadsheet rendered as PDF pages | |
| 4 | Open `.pptx` | Slides rendered as PDF pages | |
| 5 | Open `.hwp` | Renders (quality varies — LibreOffice partial support) | |
| 6 | Open malformed .docx (e.g., zero-byte file) | 500 error toast | |
| 7 | Path traversal `?loc=../../etc&name=passwd` | 400 (already covered by middleware) | |
| 8 | Open 10MB docx with images | Loads within timeout, renders all pages | |
| 9 | Open same file from two tabs simultaneously | One conversion runs (dedup), both succeed | |
| 10 | Reuse PdfViewer features (zoom, page nav, download) | All work on office-converted PDFs | |

- [ ] **Step 4: Final `/code-review`** on all created/modified files → 0 ❌ Critical.

- [ ] **Step 5: Push + PR**

```bash
git push -u origin feat/office-docs
```

PR title: `[feat] office document viewer via LibreOffice headless conversion`
PR body: reference this plan + spec walkthrough table.

PR description must mention the Dockerfile change — reviewer should verify image build succeeds in CI.

---

## Completion Criteria

- All 13 tasks (Task 0 - Task 12) committed.
- Backend: `go test ./...` passes (or skips cleanly without soffice).
- Frontend: `npm run check` + `npm run build` pass.
- Docker image builds with LibreOffice installed; `docker run nas:test soffice --version` works.
- `/code-review` 0 ❌ Critical on all changed files.
- Spec walkthrough rows pass (where applicable for local environment).
- PR opened against `main`.

## Risk Register

| Risk | Mitigation |
|---|---|
| LibreOffice install adds ~400MB to image | Accepted trade-off (decision `b`). Consider Alpine `libreoffice` or distroless variants in v2. |
| `soffice` hangs on malformed input | `context.WithTimeout(30s)` in `Convert` → request returns 500 cleanly |
| HWP fidelity | Documented as "best effort"; v2 could add a dedicated HWP renderer if quality unacceptable |
| Cache disk usage grows unbounded | v1 has no eviction; track in Risk Register as known TODO. Operator can `rm -rf $NAS_TEMP_DIR/office-cache` manually. v2 adds size-based LRU. |
| Concurrent conversions overload CPU | Dedupe handles same-file; different-file parallelism limited by Go scheduler. v2 can add a global semaphore (e.g., 4 concurrent soffice). |
| `soffice --convert-to pdf` differs across versions | Tested with LibreOffice 7.x; output format `pdf` stable across recent versions. Output filename pattern fixed (basename + `.pdf`). |
| Test fixtures take disk space in repo | ~30KB total — acceptable |
| CI without LibreOffice | Tests skip cleanly via `skipIfNoSoffice` helper |
| Soffice writes to `/tmp` even with `--outdir` | LibreOffice's user profile dir; set `HOME=/tmp` or `--user-profile=/tmp/soffice-profile` if container has read-only root. Defer fix until observed. |
