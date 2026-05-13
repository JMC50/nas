# Go Backend Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the existing Node.js/Express backend entirely with a Go backend that (1) ships large-file uploads reliably via tus protocol, (2) preserves all existing data — SQLite, password hashes, JWT tokens, OAuth IDs, file storage — with zero user-visible regression, and (3) deploys as a single Docker image.

**Architecture:** Single static Go binary serving HTTP API + (Phase 7+) embedded SvelteKit static build via `embed.FS`. Backed by the existing `nas.sqlite` file with no schema-breaking changes. Auth via `golang-jwt/jwt` (HS256, same `PRIVATE_KEY` — existing tokens survive) and `golang.org/x/crypto/bcrypt` (binary-compatible with `bcryptjs` hashes). Resumable uploads via the official `tus/tusd` server, mounted as a sub-router. Docker multi-stage build with ffmpeg/libreoffice for Phase 3+ media/document preview.

**Tech Stack:**
- **Go 1.23+**
- **HTTP**: `go-chi/chi v5` — lightweight, idiomatic router
- **DB**: `modernc.org/sqlite` — pure-Go SQLite (no CGO, simpler cross-compile)
- **JWT**: `golang-jwt/jwt/v5`
- **Bcrypt**: `golang.org/x/crypto/bcrypt`
- **Uploads**: `tus/tusd/pkg/handler` + filestore
- **Archives**: `archive/zip` stdlib (stream-friendly) + `mholt/archiver/v4` if needed for tar
- **Validation**: `go-playground/validator/v10`
- **Logging**: `log/slog` stdlib
- **Tests**: stdlib `testing` + `stretchr/testify`
- **E2E**: `@playwright/test` Node package (separate from Go)
- **System info**: `shirou/gopsutil/v3` (cross-platform disk/cpu/mem)

---

## Migration Strategy

### Data continuity — guaranteed compatible
| Asset | Storage | Compatibility | Action needed |
|-------|---------|---------------|---------------|
| User files | `../../nas-data/`, `../../nas-data-admin/` | Just files on disk | None — Go reads same paths |
| SQLite DB | `backend/db/nas.sqlite` | `modernc.org/sqlite` reads same file | Move to `data/db/` Docker volume |
| `users` table | bcryptjs hashes (`$2a$/$2b$`) | Identical to Go bcrypt | None |
| OAuth IDs | TEXT column | Just strings | None |
| `user_intents` | TEXT permissions | Same schema | None |
| `log` table | INTEGER timestamps | Same schema | None |
| JWT tokens | HS256 + `PRIVATE_KEY` | `golang-jwt/jwt` HS256 compatible | None — issued tokens stay valid |

**Net result:** No forced re-login, no password reset, no data migration script needed.

### Safety net: rename, don't delete
1. `git mv backend backend.legacy` — current Node backend stays accessible as a frozen reference
2. New Go code goes into a fresh `backend/` directory
3. `docker-compose.yml` is rewritten for the Go service only (no parallel legacy service — too much complexity for marginal benefit)
4. Rollback is via `git checkout` of the pre-migration commit + rebuild — DB and file volumes (`./data/`) persist across the rollback, so no data loss
5. After 2 weeks of stable Go in production: `git rm -r backend.legacy/`

### Defaults for unanswered questions
| Question | Default chosen | Override how |
|----------|----------------|--------------|
| Env var management | Hybrid — Docker secrets in production (`*_FILE` pattern, mirrors current `getSecret()`), `.env` in dev | Set `DOCKER_SECRETS_DIR` env var or use plain `PRIVATE_KEY=…` in compose |
| OAuth redirect URI | Keep port `7777`, paths unchanged (`/login`, `/kakaoLogin`) | No app-console changes needed |
| Docker host | Docker Desktop on Windows (volume binds use Linux paths inside container) | WSL2 backend works identically |

---

## Current Backend API Inventory

All endpoints currently use **query parameters for token + intent check pattern**. We mirror them exactly in Go to keep frontend changes minimal in Phase 1, then phase in REST-cleanup in a later plan.

| Method | Path | Intent | Notes for Go port |
|--------|------|--------|-------------------|
| GET | `/` | — | Health |
| GET | `/getSystemInfo` | — | Use `gopsutil`; replace shelling out to `df` |
| GET | `/stat` | — | `os.Stat`, ISO-format dates |
| GET | `/download` | DOWNLOAD | `http.ServeContent` (range + sendfile) |
| GET | `/getTextFile` | OPEN | `os.ReadFile` UTF-8 |
| POST | `/saveTextFile` | UPLOAD | `os.WriteFile` |
| GET | `/getVideoData` | OPEN | `http.ServeContent` (handles range) |
| GET | `/getAudioData` | OPEN | `http.ServeContent` |
| GET | `/getImageData` | OPEN | Set Content-Type by extension, `http.ServeFile` |
| GET | `/forceDelete` | DELETE | `os.RemoveAll` |
| GET | `/copy` | COPY | `cp -r` equivalent (recursive) |
| GET | `/move` | COPY | `os.Rename` (cross-FS fallback: copy+delete) |
| GET | `/rename` | RENAME | `os.Rename` |
| POST | `/zipFiles` | UPLOAD | `archive/zip` streaming + progress |
| POST | `/unzipFile` | UPLOAD | `archive/zip` reader + progress |
| GET | `/progress` | — | Replace tmpfile-based progress with in-memory `sync.Map` |
| GET | `/makedir` | UPLOAD | `os.MkdirAll` |
| GET | `/readFolder` | VIEW | `os.ReadDir` |
| GET | `/searchInAllFiles` | VIEW | `filepath.WalkDir` |
| GET | `/img` | — | Bundled static icons (embed.FS) |
| POST | `/input` | UPLOAD | **Replace with tus** — but keep this endpoint as a fallback wrapper for backward compat in Phase 1 |
| POST | `/inputZip` | UPLOAD | tus upload + post-process unzip |
| GET | `/login` | — | Discord OAuth callback |
| GET | `/kakaoLogin` | — | Kakao OAuth callback |
| POST | `/register` | — | Discord register |
| POST | `/registerKakao` | — | Kakao register |
| GET | `/auth/config` | — | Static config |
| POST | `/auth/register` | — | Local register (bcrypt) |
| POST | `/auth/login` | — | Local login (bcrypt verify) |
| POST | `/auth/change-password` | — | Bcrypt rehash |
| GET | `/getIntents` | — | SELECT |
| GET | `/checkAdmin` | — | SELECT |
| GET | `/getAllUsers` | — | SELECT + intent map |
| GET | `/getActivityLog` | — | SELECT JOIN |
| GET | `/checkIntent` | — | SELECT |
| GET | `/authorize` | ADMIN | toggle intent |
| GET | `/unauthorize` | ADMIN | toggle intent |
| POST | `/requestAdminIntent` | — | Verify admin password |
| POST | `/log` | — | Insert log |
| GET | `/downloadZip` | — | `http.ServeContent` of tmp zip |
| GET | `/deleteTempZip` | — | `os.Remove` of tmp zip |

**Defects observed in current code** (carry forward as Go improvements, not bugs to preserve):
- `getDiskUsage()` uses `df -h` — fails on Windows; replace with `gopsutil`
- `searchFilesInDir()` parses path via `split("/nas-data")` — Windows-broken; fix with `filepath.Rel`
- `/forceDelete`, `/copy`, `/move`, `/rename` are GET — keep paths/methods identical in Phase 1 for FE compat; clean up in a separate "REST cleanup" plan
- Path traversal: current code uses `path.resolve()` but doesn't verify the resolved path stays inside `nas-data/` — Go port MUST add `strings.HasPrefix(resolved, baseDir+sep)` check
- `/zipFiles` writes progress to `/tmp/nas-progress-*.json` — fails on Windows; replace with in-memory store

---

## Go Project File Structure

```
backend/
├── go.mod                          # Module declaration + dependencies
├── go.sum
├── Dockerfile                      # Multi-stage: builder (alpine+go) → runtime (alpine+ffmpeg+libreoffice)
├── .dockerignore
├── cmd/
│   └── server/
│       └── main.go                 # Entry: load config, wire deps, start http.Server
├── internal/
│   ├── config/
│   │   ├── config.go               # struct Config, LoadFromEnv()
│   │   └── paths.go                # platform-aware path resolution
│   ├── auth/
│   │   ├── jwt.go                  # Issue/Verify HS256, claim struct matches Node payload
│   │   ├── bcrypt.go               # Hash/Verify wrappers
│   │   ├── intents.go              # type Intent string; constants ADMIN/VIEW/…
│   │   ├── password.go             # Validate per requirements
│   │   ├── discord.go              # OAuth fetch + register
│   │   ├── kakao.go                # OAuth fetch + register
│   │   ├── middleware.go           # RequireToken, RequireIntent
│   │   └── handlers.go             # /auth/* HTTP handlers
│   ├── db/
│   │   ├── sqlite.go               # Open existing nas.sqlite; verify schema (no destructive ALTER)
│   │   ├── users.go                # GetUser, GetAllUsers, SaveUser, UpdateUser
│   │   ├── intents.go              # GetIntents, EditIntent, HasIntent
│   │   ├── logs.go                 # InsertLog, GetActivityLogs
│   │   └── schema_verify.go        # On startup: assert tables/columns exist; FAIL FAST if mismatch
│   ├── files/
│   │   ├── safepath.go             # SafeJoin(base, untrusted) — blocks ../ traversal
│   │   ├── browse.go               # ReadFolder, Stat, Search
│   │   ├── crud.go                 # MkDir, Rename, Copy, Move, Delete
│   │   ├── text.go                 # GetText, SaveText
│   │   └── handlers.go
│   ├── stream/
│   │   ├── range.go                # Generic http.ServeContent wrapper with intent check
│   │   └── handlers.go             # /getVideoData, /getAudioData, /getImageData, /img
│   ├── upload/
│   │   ├── tus.go                  # tusd handler creation + filestore wiring
│   │   ├── legacy.go               # POST /input wrapper (backward compat — pipes to disk)
│   │   ├── post_hook.go            # Move from tus staging → nas-data, log activity
│   │   └── handlers.go
│   ├── archive/
│   │   ├── zip.go                  # Stream zip with progress events
│   │   ├── unzip.go                # Stream unzip with progress events
│   │   ├── progress.go             # type Tracker struct{sync.Map}; Tracker.Set/Get
│   │   └── handlers.go
│   ├── admin/
│   │   ├── users.go                # /getAllUsers, /authorize, /unauthorize, /requestAdminIntent
│   │   ├── logs.go                 # /getActivityLog
│   │   └── handlers.go
│   ├── system/
│   │   └── info.go                 # gopsutil-based CPU/mem/disk/uptime
│   ├── web/
│   │   ├── embed.go                # Phase 7: embed.FS for SvelteKit static build
│   │   └── spa.go                  # SPA fallback router (index.html for unknown paths)
│   └── server/
│       ├── router.go               # chi router assembly
│       ├── middleware.go           # CORS, recovery, request logging
│       └── server.go               # http.Server lifecycle (graceful shutdown)
├── tests/
│   ├── integration/
│   │   ├── auth_test.go            # Login flow, JWT round-trip
│   │   ├── files_test.go           # CRUD + path traversal blocking
│   │   ├── stream_test.go          # Range request behavior
│   │   ├── upload_test.go          # tus protocol compliance
│   │   ├── archive_test.go         # zip/unzip
│   │   └── helpers.go              # Shared fixtures, test DB setup
│   └── fixtures/
│       ├── bcrypt_hashes.txt       # Sample hashes from existing DB for compatibility tests
│       └── sample_files/           # Tiny test files (image, video, text)
└── e2e/                            # Phase 7 — Playwright TS, NOT Go tests
    ├── package.json
    ├── playwright.config.ts
    └── tests/
        ├── auth.spec.ts
        ├── browse.spec.ts
        ├── upload-large.spec.ts    # Multi-GB upload via tus
        ├── preview.spec.ts         # Video/PDF/image
        └── admin.spec.ts
```

**Decomposition rationale:** Each `internal/<domain>/handlers.go` owns its HTTP surface; `db/` is a thin repo layer (no business logic); `auth/middleware.go` is the only place tokens get verified. Files stay under ~300 lines each.

---

## Per-Phase Quality Gate

**MANDATORY after every Task that writes/edits code, in EVERY phase (1 through 8):**
1. Run `go test ./... -race` — must pass
2. Run `go vet ./...` — must pass
3. Run `gofmt -l . | wc -l` — must be 0 (PowerShell: `(gofmt -l .).Count` — must be 0)
4. **Invoke `code-review` skill on the changed files** (per global CLAUDE.md rule — NOT optional)
5. Fix every ❌ Critical from review before next Task; ⚠️ Warnings get surfaced to user

**For Phases 2-8:** This quality gate is the implicit final step of every Task in the bulleted task lists, even when not spelled out. Skipping it violates the user's global rule.

Phase 1 spells the gate out explicitly so the pattern is clear; subsequent phases assume it.

---

# PHASE 1: Project Scaffold + DB + Health (Day 1-2)

**Goal:** A Go binary that starts, opens existing `nas.sqlite`, verifies schema, exposes `GET /healthz` and `GET /`, has working tests, builds in Docker.

**Done when:**
- `go test ./... -race` passes
- `docker compose build` succeeds
- `curl http://localhost:7777/healthz` returns `{"status":"ok","db":"connected","schema":"valid"}`

### Task 1.1: Branch + backup current backend

**Files:**
- Rename: `backend/` → `backend.legacy/`

- [ ] **Step 1: Create feature branch**

```bash
git checkout -b feature/go-backend-migration
```

- [ ] **Step 2: Rename backend directory (preserves git history)**

```bash
git mv backend backend.legacy
```

- [ ] **Step 3: Verify rename preserved structure (no npm install — don't mutate legacy)**

```bash
test -f backend.legacy/src/index.ts && echo "OK: index.ts present"
test -f backend.legacy/db/nas.sqlite && echo "OK: existing DB intact"
git status  # should show ONLY the rename, no untracked node_modules etc.
```

Expected: Both `OK:` lines + clean `git status` (besides the rename itself).

- [ ] **Step 4: Commit**

```bash
git commit -m "[refactor] move backend to backend.legacy for Go migration

[Body]
- Preserve current Node.js backend as backend.legacy for emergency reference
- Will be removed after Go backend reaches 2-week stable production state
- Rollback is via git checkout of this commit (not via running legacy service)"
```

### Task 1.2: Initialize Go module

**Files:**
- Create: `backend/go.mod`
- Create: `backend/.gitignore`

- [ ] **Step 1: Create directory and init module**

```bash
mkdir -p backend/cmd/server backend/internal backend/tests/integration backend/tests/fixtures
cd backend && go mod init github.com/kiwooriHS/nas
# ⚠️ Module name is a guess based on the git user `kiwooriHS`. If your fork
# lives elsewhere (e.g. github.com/<other-account>/nas), run:
#   go mod edit -module=github.com/<your-account>/nas
# All import paths in this plan use this name — keep them in sync after change.
```

- [ ] **Step 2: Add dependencies**

```bash
go get github.com/go-chi/chi/v5@latest
go get modernc.org/sqlite@latest
go get github.com/golang-jwt/jwt/v5@latest
go get golang.org/x/crypto/bcrypt
go get github.com/tus/tusd/v2/pkg/handler@latest
go get github.com/tus/tusd/v2/pkg/filestore@latest
go get github.com/stretchr/testify@latest
go get github.com/shirou/gopsutil/v3@latest
go get github.com/go-playground/validator/v10@latest
go mod tidy
```

- [ ] **Step 3: Create `.gitignore`**

```gitignore
# Binaries
/server
*.exe
*.test

# Coverage
*.out
coverage.html

# Go workspace
go.work
go.work.sum

# Test artifacts
tests/fixtures/temp/
```

- [ ] **Step 4: Run `go vet ./...` to verify the module compiles**

Expected: No output (clean).

- [ ] **Step 5: Commit**

```bash
git add backend/
git commit -m "[chore] initialize Go module with core dependencies

[Body]
- chi v5 for routing
- modernc.org/sqlite (pure Go, no CGO)
- golang-jwt/jwt v5 (HS256 compatibility with existing Node tokens)
- golang.org/x/crypto/bcrypt (compatibility with bcryptjs hashes)
- tus/tusd v2 for resumable uploads
- gopsutil for cross-platform system info"
```

### Task 1.3: Config loading (mirror environment.ts)

**Files:**
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/config/paths.go`
- Test: `backend/internal/config/config_test.go`

- [ ] **Step 1: Write failing test for Config.LoadFromEnv**

`backend/internal/config/config_test.go`:
```go
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFromEnv_RequiredFields(t *testing.T) {
	t.Setenv("PRIVATE_KEY", "test-key")
	t.Setenv("ADMIN_PASSWORD", "test-admin")
	t.Setenv("PORT", "7777")

	cfg, err := LoadFromEnv()
	require.NoError(t, err)
	require.Equal(t, "test-key", cfg.PrivateKey)
	require.Equal(t, "test-admin", cfg.AdminPassword)
	require.Equal(t, 7777, cfg.Port)
}

func TestLoadFromEnv_ProductionRequiresSecrets(t *testing.T) {
	t.Setenv("NODE_ENV", "production")
	os.Unsetenv("PRIVATE_KEY")
	_, err := LoadFromEnv()
	require.Error(t, err)
	require.Contains(t, err.Error(), "PRIVATE_KEY")
}

func TestLoadFromEnv_DockerSecretFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "secret")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())
	_, _ = tmp.WriteString("secret-from-file")
	tmp.Close()

	t.Setenv("PRIVATE_KEY_FILE", tmp.Name())
	t.Setenv("ADMIN_PASSWORD", "x")
	cfg, err := LoadFromEnv()
	require.NoError(t, err)
	require.Equal(t, "secret-from-file", cfg.PrivateKey)
}
```

- [ ] **Step 2: Run test, confirm it fails**

```bash
cd backend && go test ./internal/config/ -v
```

Expected: FAIL — `LoadFromEnv` undefined.

- [ ] **Step 3: Implement `config.go`**

`backend/internal/config/config.go`:
```go
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type AuthType string

const (
	AuthTypeOAuth AuthType = "oauth"
	AuthTypeLocal AuthType = "local"
	AuthTypeBoth  AuthType = "both"
)

type PasswordRequirements struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

type Config struct {
	NodeEnv       string
	Port          int
	Host          string
	IsProduction  bool

	PrivateKey    string
	AdminPassword string
	AuthType      AuthType
	JWTExpiry     string // duration string, parsed at use site

	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURI  string
	KakaoRestAPIKey     string
	KakaoClientSecret   string
	KakaoRedirectURI    string

	NASDataDir      string
	NASAdminDataDir string
	NASTempDir      string
	DBPath          string

	PasswordRequirements PasswordRequirements
	CorsOrigin           string
	MaxFileSizeBytes     int64
}

func LoadFromEnv() (*Config, error) {
	c := &Config{
		NodeEnv: getEnv("NODE_ENV", "development"),
	}
	c.IsProduction = c.NodeEnv == "production"

	c.Port = getEnvInt("PORT", 7777)
	c.Host = getEnv("HOST", "0.0.0.0")
	c.AuthType = AuthType(getEnv("AUTH_TYPE", "both"))
	c.JWTExpiry = getEnv("JWT_EXPIRY", "168h") // 7 days

	c.PrivateKey = getSecret("PRIVATE_KEY", "")
	c.AdminPassword = getSecret("ADMIN_PASSWORD", "")

	c.DiscordClientID = getSecret("DISCORD_CLIENT_ID", "")
	c.DiscordClientSecret = getSecret("DISCORD_CLIENT_SECRET", "")
	c.DiscordRedirectURI = getEnv("DISCORD_REDIRECT_URI", "")
	c.KakaoRestAPIKey = getSecret("KAKAO_REST_API_KEY", "")
	c.KakaoClientSecret = getSecret("KAKAO_CLIENT_SECRET", "")
	c.KakaoRedirectURI = getEnv("KAKAO_REDIRECT_URI", "")

	c.NASDataDir = getEnv("NAS_DATA_DIR", "")
	c.NASAdminDataDir = getEnv("NAS_ADMIN_DATA_DIR", "")
	c.NASTempDir = getEnv("NAS_TEMP_DIR", os.TempDir())
	c.DBPath = getEnv("DB_PATH", "")

	c.PasswordRequirements = PasswordRequirements{
		MinLength:        getEnvInt("PASSWORD_MIN_LENGTH", 8),
		RequireUppercase: getEnvBool("PASSWORD_REQUIRE_UPPERCASE", false),
		RequireLowercase: getEnvBool("PASSWORD_REQUIRE_LOWERCASE", false),
		RequireNumber:    getEnvBool("PASSWORD_REQUIRE_NUMBER", false),
		RequireSpecial:   getEnvBool("PASSWORD_REQUIRE_SPECIAL", false),
	}

	c.CorsOrigin = getEnv("CORS_ORIGIN", "*")
	c.MaxFileSizeBytes = parseSizeOrDefault(getEnv("MAX_FILE_SIZE", "50gb"), 50*1024*1024*1024)

	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) validate() error {
	var errs []string
	if c.IsProduction {
		if c.PrivateKey == "" {
			errs = append(errs, "PRIVATE_KEY required in production")
		}
		if c.AdminPassword == "" {
			errs = append(errs, "ADMIN_PASSWORD required in production")
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		return v == "true"
	}
	return def
}

func getSecret(key, def string) string {
	if path, ok := os.LookupEnv(key + "_FILE"); ok {
		if data, err := os.ReadFile(path); err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return getEnv(key, def)
}

func parseSizeOrDefault(s string, def int64) int64 {
	s = strings.ToLower(strings.TrimSpace(s))
	var mult int64 = 1
	switch {
	case strings.HasSuffix(s, "gb"):
		mult = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "gb")
	case strings.HasSuffix(s, "mb"):
		mult = 1024 * 1024
		s = strings.TrimSuffix(s, "mb")
	case strings.HasSuffix(s, "kb"):
		mult = 1024
		s = strings.TrimSuffix(s, "kb")
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return def
	}
	return n * mult
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{Env=%s, Port=%d, AuthType=%s, IsProd=%t}", c.NodeEnv, c.Port, c.AuthType, c.IsProduction)
}
```

- [ ] **Step 4: Implement `paths.go`**

`backend/internal/config/paths.go`:
```go
package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// ResolvePaths fills in DataDir/AdminDataDir/DBPath defaults based on platform + env.
// Explicit env values always win.
func (c *Config) ResolvePaths() error {
	if c.NASDataDir == "" {
		c.NASDataDir = defaultDataDir(c.IsProduction, "data")
	}
	if c.NASAdminDataDir == "" {
		c.NASAdminDataDir = defaultDataDir(c.IsProduction, "admin-data")
	}
	if c.DBPath == "" {
		base := defaultDataDir(c.IsProduction, "db")
		c.DBPath = filepath.Join(base, "nas.sqlite")
	}

	// Ensure parents exist
	for _, dir := range []string{c.NASDataDir, c.NASAdminDataDir, filepath.Dir(c.DBPath), c.NASTempDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func defaultDataDir(isProd bool, sub string) string {
	if isProd {
		// In Docker: NAS_DATA_DIR is set via compose. Fallback to /app/<sub>.
		return filepath.Join("/app", sub)
	}
	// Dev: relative to cwd, two levels up (matches current Node behavior)
	cwd, _ := os.Getwd()
	if runtime.GOOS == "windows" {
		return filepath.Join(cwd, "..", "..", "nas-"+sub)
	}
	return filepath.Join(cwd, "..", "..", "nas-"+sub)
}
```

- [ ] **Step 5: Run tests — should pass**

```bash
go test ./internal/config/ -v -race
```

Expected: All PASS.

- [ ] **Step 6: Run quality gate**

```bash
go vet ./...
gofmt -l .
```

Expected: Empty output.

- [ ] **Step 7: Invoke `code-review` skill on changed files**

Files to review: `internal/config/config.go`, `internal/config/paths.go`, `internal/config/config_test.go`. Fix every ❌ Critical before proceeding.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/config/
git commit -m "[feat] add config loader with env vars + Docker secrets support

[Body]
- Mirror Node Environment class: NODE_ENV, PORT, PRIVATE_KEY, ADMIN_PASSWORD, AUTH_TYPE
- Docker secrets via *_FILE pattern (production-first)
- Cross-platform path resolution (NAS_DATA_DIR, DB_PATH)
- Password requirement struct
- Unit tests: env vars, secret file, production validation"
```

### Task 1.4: SQLite connection + schema verification

**Files:**
- Create: `backend/internal/db/sqlite.go`
- Create: `backend/internal/db/schema_verify.go`
- Test: `backend/internal/db/schema_verify_test.go`

- [ ] **Step 1: Write failing test for schema verification**

`backend/internal/db/schema_verify_test.go`:
```go
package db

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

const fullSchema = `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userId TEXT UNIQUE NOT NULL,
		username TEXT NOT NULL,
		global_name TEXT,
		krname TEXT,
		password TEXT,
		auth_type TEXT
	);
	CREATE TABLE user_intents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		intent TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE TABLE log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		activity TEXT NOT NULL,
		description TEXT,
		user_id INTEGER NOT NULL,
		time INTEGER NOT NULL,
		loc TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
	);
`

func TestVerifySchema_AcceptsCurrentSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	conn, err := Open(dbPath)
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.Exec(fullSchema)
	require.NoError(t, err)

	err = VerifySchema(conn)
	require.NoError(t, err)
}

func TestVerifySchema_FailsOnMissingTable(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	conn, err := Open(dbPath)
	require.NoError(t, err)
	defer conn.Close()
	// Don't create any tables.
	// Verification checks 'users' first (slice order is deterministic).

	err = VerifySchema(conn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "users")
}

func TestVerifySchema_FailsOnMissingColumn(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	conn, err := Open(dbPath)
	require.NoError(t, err)
	defer conn.Close()
	// Create users table missing `auth_type` column
	_, err = conn.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		userId TEXT,
		username TEXT,
		global_name TEXT,
		krname TEXT,
		password TEXT
	);
	CREATE TABLE user_intents (id INTEGER PRIMARY KEY, user_id INTEGER, intent TEXT);
	CREATE TABLE log (id INTEGER PRIMARY KEY, activity TEXT, description TEXT, user_id INTEGER, time INTEGER, loc TEXT);`)
	require.NoError(t, err)

	err = VerifySchema(conn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "auth_type")
}
```

- [ ] **Step 2: Run test — should fail (functions undefined)**

```bash
go test ./internal/db/ -v -run TestVerifySchema
```

Expected: FAIL.

- [ ] **Step 3: Implement `sqlite.go`**

`backend/internal/db/sqlite.go`:
```go
package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return conn, nil
}
```

- [ ] **Step 4: Implement `schema_verify.go`**

`backend/internal/db/schema_verify.go`:
```go
package db

import (
	"database/sql"
	"fmt"
)

// tableSchema is one required table with its columns.
// Slice (not map) so verification order is deterministic and tests are stable.
type tableSchema struct {
	name    string
	columns []string
}

// requiredSchema is checked in order on startup.
// We only verify presence — destructive ALTER is forbidden (data continuity guarantee).
// Note: column types are intentionally NOT verified — bcryptjs hashes were stored as TEXT
// and remain TEXT under Go bcrypt. If a future migration changes types, update this slice.
var requiredSchema = []tableSchema{
	{"users", []string{"id", "userId", "username", "global_name", "krname", "password", "auth_type"}},
	{"user_intents", []string{"id", "user_id", "intent"}},
	{"log", []string{"id", "activity", "description", "user_id", "time", "loc"}},
}

func VerifySchema(conn *sql.DB) error {
	for _, t := range requiredSchema {
		got, err := tableColumns(conn, t.name)
		if err != nil {
			return fmt.Errorf("introspect %s: %w", t.name, err)
		}
		for _, want := range t.columns {
			if _, ok := got[want]; !ok {
				return fmt.Errorf("table %s missing column %s", t.name, want)
			}
		}
	}
	return nil
}

func tableColumns(conn *sql.DB, table string) (map[string]struct{}, error) {
	rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := map[string]struct{}{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols[name] = struct{}{}
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("table %s does not exist", table)
	}
	return cols, nil
}
```

- [ ] **Step 5: Run tests — should pass**

```bash
go test ./internal/db/ -v -race
```

- [ ] **Step 6: Quality gate**

```bash
go vet ./...
gofmt -l .
```

- [ ] **Step 7: Invoke `code-review` skill on `internal/db/*.go`**

- [ ] **Step 8: Commit**

```bash
git add backend/internal/db/
git commit -m "[feat] add SQLite connection with non-destructive schema verification

[Body]
- modernc.org/sqlite pure Go driver (no CGO)
- WAL mode + foreign keys enabled
- VerifySchema asserts required columns exist; fails fast on mismatch
- No destructive ALTER — existing data preserved as-is
- Tests cover happy path and missing-table failure"
```

### Task 1.5: HTTP server + health endpoint

**Files:**
- Create: `backend/internal/server/router.go`
- Create: `backend/internal/server/server.go`
- Create: `backend/cmd/server/main.go`
- Test: `backend/tests/integration/health_test.go`

- [ ] **Step 1: Write failing integration test**

`backend/tests/integration/health_test.go`:
```go
package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/kiwooriHS/nas/internal/config"
	"github.com/kiwooriHS/nas/internal/db"
	"github.com/kiwooriHS/nas/internal/server"
	"github.com/stretchr/testify/require"
)

func TestHealthz_ReturnsOK(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	conn, err := db.Open(dbPath)
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.Exec(testSchema())
	require.NoError(t, err)

	cfg := &config.Config{Port: 0}
	r := server.NewRouter(cfg, conn)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Equal(t, "ok", body["status"])
	require.Equal(t, "connected", body["db"])
	require.Equal(t, "valid", body["schema"])
}

func testSchema() string {
	return `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			userId TEXT UNIQUE NOT NULL,
			username TEXT NOT NULL,
			global_name TEXT,
			krname TEXT,
			password TEXT,
			auth_type TEXT
		);
		CREATE TABLE user_intents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			intent TEXT NOT NULL
		);
		CREATE TABLE log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity TEXT NOT NULL,
			description TEXT,
			user_id INTEGER NOT NULL,
			time INTEGER NOT NULL,
			loc TEXT
		);
	`
}
```

- [ ] **Step 2: Run test — should fail**

```bash
go test ./tests/integration/ -v -run TestHealthz
```

Expected: FAIL — `server.NewRouter` undefined.

- [ ] **Step 2.5: Add `go-chi/cors` dependency**

```bash
go get github.com/go-chi/cors@latest
go mod tidy
```

- [ ] **Step 3: Implement `server/router.go`**

`backend/internal/server/router.go`:
```go
package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/kiwooriHS/nas/internal/config"
	"github.com/kiwooriHS/nas/internal/db"
)

func NewRouter(cfg *config.Config, conn *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	// CORS: dev frontend runs on :5050. Production should set CORS_ORIGIN explicitly.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CorsOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Tus-Resumable", "Upload-Length", "Upload-Metadata", "Upload-Offset"},
		ExposedHeaders:   []string{"Tus-Resumable", "Upload-Offset", "Upload-Length", "Location"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("server is running :D"))
	})

	r.Get("/healthz", healthzHandler(conn))
	return r
}

func healthzHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"status": "ok"}
		if err := conn.Ping(); err != nil {
			resp["status"] = "degraded"
			resp["db"] = "disconnected"
		} else {
			resp["db"] = "connected"
		}
		if err := db.VerifySchema(conn); err != nil {
			resp["schema"] = "invalid: " + err.Error()
		} else {
			resp["schema"] = "valid"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
```

- [ ] **Step 4: Implement `server/server.go` + `cmd/server/main.go`**

`backend/internal/server/server.go`:
```go
package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/kiwooriHS/nas/internal/config"
)

func Run(cfg *config.Config, conn *sql.DB) error {
	r := NewRouter(cfg, conn)
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,  // slowloris protection
		IdleTimeout:       120 * time.Second,
		// Intentionally NO ReadTimeout/WriteTimeout — large uploads/downloads need indefinite duration.
		// Per-route timeouts (e.g. 30s for /auth/*) can be added via middleware.
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", srv.Addr, "config", cfg.String())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-errCh:
		slog.Error("server failed", "err", err)
		return err
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return srv.Shutdown(shutCtx)
}
```

`backend/cmd/server/main.go`:
```go
package main

import (
	"log/slog"
	"os"

	"github.com/kiwooriHS/nas/internal/config"
	"github.com/kiwooriHS/nas/internal/db"
	"github.com/kiwooriHS/nas/internal/server"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("config error", "err", err)
		os.Exit(1)
	}
	if err := cfg.ResolvePaths(); err != nil {
		slog.Error("path resolution failed", "err", err)
		os.Exit(1)
	}

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("db open failed", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	if err := db.VerifySchema(conn); err != nil {
		slog.Error("schema verification failed", "err", err)
		os.Exit(1)
	}

	if err := server.Run(cfg, conn); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 5: Run integration test — should pass**

```bash
go test ./tests/integration/ -v -race
```

- [ ] **Step 6: Build and smoke-test manually**

**Bash (Linux/Mac/Git Bash on Windows):**
```bash
go build -o ./bin/server ./cmd/server
DB_PATH="../backend.legacy/db/nas.sqlite" \
PRIVATE_KEY=devkey \
ADMIN_PASSWORD=devadmin \
PORT=7778 \
./bin/server &
sleep 1
curl -s http://localhost:7778/healthz | jq .
kill %1
```

**PowerShell (Windows native):**
```powershell
go build -o ./bin/server.exe ./cmd/server
$env:DB_PATH = "..\backend.legacy\db\nas.sqlite"
$env:PRIVATE_KEY = "devkey"
$env:ADMIN_PASSWORD = "devadmin"
$env:PORT = "7778"
$proc = Start-Process -PassThru .\bin\server.exe
Start-Sleep -Seconds 1
Invoke-RestMethod http://localhost:7778/healthz | ConvertTo-Json
Stop-Process -Id $proc.Id
```

Expected: `{"status":"ok","db":"connected","schema":"valid"}`

- [ ] **Step 7: Quality gate**

```bash
go vet ./...
gofmt -l .
go test ./... -race -cover
```

- [ ] **Step 8: Invoke `code-review` skill on all new files**

- [ ] **Step 9: Commit**

```bash
git add backend/
git commit -m "[feat] add HTTP server with /healthz and graceful shutdown

[Body]
- chi v5 router with Recoverer/RealIP/Logger middleware
- /healthz reports DB connectivity + schema validity
- Graceful shutdown on SIGINT/SIGTERM (30s timeout)
- Main entry wires config → db → server with fail-fast on schema mismatch
- Integration test using httptest"
```

### Task 1.6: Dockerfile (Go builder + minimal runtime)

**Files:**
- Create: `backend/Dockerfile`
- Modify: `docker-compose.yml` (add Go service, demote Node to legacy profile)

- [ ] **Step 1: Create Dockerfile**

`backend/Dockerfile`:
```dockerfile
# syntax=docker/dockerfile:1.6

FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata
# Phase 3+ will add: ffmpeg libreoffice
WORKDIR /app
COPY --from=build /out/server /app/server

ENV NODE_ENV=production
ENV PORT=7777
ENV HOST=0.0.0.0
ENV NAS_DATA_DIR=/data/nas
ENV NAS_ADMIN_DATA_DIR=/data/nas-admin
ENV DB_PATH=/data/db/nas.sqlite

EXPOSE 7777

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:7777/healthz || exit 1

CMD ["/app/server"]
```

- [ ] **Step 2: Update `docker-compose.yml`**

Replace top-level structure (Go-only; rollback is via `git checkout phase1-complete~1` + rebuild, not a running legacy service):

```yaml
services:
  nas-app:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: nas-go
    ports:
      - "7777:7777"
    volumes:
      - ./data/nas:/data/nas
      - ./data/nas-admin:/data/nas-admin
      - ./data/db:/data/db
    environment:
      - NODE_ENV=production
      - PRIVATE_KEY=${PRIVATE_KEY}
      - ADMIN_PASSWORD=${ADMIN_PASSWORD}
      - AUTH_TYPE=${AUTH_TYPE:-both}
      - CORS_ORIGIN=${CORS_ORIGIN:-*}
      - MAX_FILE_SIZE=${MAX_FILE_SIZE:-50gb}
      - PASSWORD_MIN_LENGTH=${PASSWORD_MIN_LENGTH:-8}
      - DISCORD_CLIENT_ID=${DISCORD_CLIENT_ID:-}
      - DISCORD_CLIENT_SECRET=${DISCORD_CLIENT_SECRET:-}
      - KAKAO_REST_API_KEY=${KAKAO_REST_API_KEY:-}
      - KAKAO_CLIENT_SECRET=${KAKAO_CLIENT_SECRET:-}
    restart: unless-stopped
```

**Rollback procedure (if Go backend breaks):**
```bash
# Stop Go service
docker compose down
# Reset to pre-migration commit
git checkout main  # or specific known-good commit
# Use the OLD docker-compose.yml (the one in backend.legacy/Dockerfile path)
docker compose up -d
```

The DB and file volumes (`./data/`) persist across this — no data loss.

- [ ] **Step 3: Test Docker build**

```bash
docker compose build nas-app
```

Expected: Build succeeds, image size <30MB (alpine + Go binary, no ffmpeg yet).

- [ ] **Step 4: Smoke test Docker run**

```bash
mkdir -p data/db
cp backend.legacy/db/nas.sqlite data/db/  # Seed with existing DB
PRIVATE_KEY=devkey ADMIN_PASSWORD=devadmin docker compose up -d nas-app
sleep 3
curl -s http://localhost:7777/healthz | jq .
docker compose logs nas-app --tail 20
docker compose down
```

Expected: `{"status":"ok","db":"connected","schema":"valid"}`

- [ ] **Step 5: Invoke `code-review` skill on Dockerfile + compose changes**

- [ ] **Step 6: Commit**

```bash
git add backend/Dockerfile docker-compose.yml
git commit -m "[feat] add Docker build for Go backend; legacy Node moved to rollback profile

[Body]
- Multi-stage build: golang:1.23-alpine → alpine:3.20 runtime
- Final image <30MB (no CGO, static binary)
- Healthcheck hits /healthz every 30s
- Volume mounts: ./data/{nas,nas-admin,db} to /data inside container
- Existing nas.sqlite is reused from mounted db volume
- Legacy Node service kept under 'rollback-only' profile"
```

### Task 1.7: Phase 1 closeout

- [ ] **Step 1: Run full Phase 1 quality gate**

```bash
cd backend
go test ./... -race -cover
go vet ./...
gofmt -l . | wc -l  # must be 0
```

- [ ] **Step 2: Run final `code-review` over entire `backend/` directory**

Fix any ❌ Critical findings.

- [ ] **Step 3: Real-DB smoke test (data continuity verification)**

This is the moment the "data continuity guarantee" goes from claim to fact.

```bash
# Copy the actual production-equivalent DB into the new volume location
mkdir -p data/db
cp backend.legacy/db/nas.sqlite data/db/nas.sqlite

# Run via Docker
PRIVATE_KEY=$(grep PRIVATE_KEY .env | cut -d= -f2) \
ADMIN_PASSWORD=$(grep ADMIN_PASSWORD .env | cut -d= -f2) \
docker compose up -d nas-app

# Verify health + schema
curl -s http://localhost:7777/healthz | jq .

# Inspect DB contents through container to confirm rows are readable
docker compose exec nas-app sh -c 'sqlite3 /data/db/nas.sqlite "SELECT count(*) as users FROM users; SELECT count(*) as intents FROM user_intents; SELECT count(*) as logs FROM log;"' || true

docker compose down
```

Expected: All three counts match the legacy Node backend's row counts.

- [ ] **Step 4: Tag the Phase 1 milestone**

```bash
git tag -a phase1-complete -m "Go backend MVP: config + DB + health + Docker"
```

- [ ] **Step 5: Update this plan**

Mark all Phase 1 tasks complete, write a brief retrospective at the bottom of the plan: "What surprised me", "Adjustments for Phase 2".

---

# PHASE 2: Auth Layer (Day 3-4)

**Goal:** Full parity with current auth API: JWT issuance/verification (compatible with existing tokens), bcrypt password verify, intents check middleware, OAuth (Discord/Kakao), local auth (register/login/change-password).

**Done when:**
- An existing JWT issued by Node backend verifies successfully in Go
- An existing bcrypt hash from the DB verifies via Go
- All `/auth/*` endpoints return identical shapes to Node version
- Discord/Kakao OAuth round-trip works (manual test with real credentials)

### Critical compatibility notes for Phase 2

**JWT claim structure**: Node code signs `{userId, expires_in}` — note that `expires_in` is NOT the standard JWT `exp` claim (Unix seconds). It's a Node `Date.now() + N` millisecond timestamp. Go side must define a matching custom claim struct:

```go
type NodeCompatClaims struct {
    UserID    string `json:"userId"`
    ExpiresIn int64  `json:"expires_in"`  // milliseconds since epoch, NOT standard exp
    jwt.RegisteredClaims                   // for future migration to standard claims
}
```

Token expiration check is manual (compare `ExpiresIn` to `time.Now().UnixMilli()`), NOT delegated to `jwt.RegisteredClaims.VerifyExpiresAt`. New tokens issued by Go should keep using this format until a coordinated frontend update can move to standard `exp`.

**Bcrypt cost**: Use `bcrypt.DefaultCost` (=10) explicitly in code with a comment — matches bcryptjs default that produced existing hashes, so re-hashing during password change won't drift.

### Tasks (TDD pattern same as Phase 1; quality gate after each)
- **2.1** Bcrypt wrapper + compatibility test against real hash extracted from DB
- **2.2** JWT issue/verify with `NodeCompatClaims` + test loading a real Node-signed token
- **2.3** User repo (GetUser, GetAllUsers, SaveUser, UpdateUser)
- **2.4** Intent repo (HasIntent, EditIntent)
- **2.5** Auth middleware (RequireToken, RequireIntent) — uses manual expiry check
- **2.6** `POST /auth/register` (local)
- **2.7** `POST /auth/login` (local)
- **2.8** `POST /auth/change-password`
- **2.9** `GET /auth/config`
- **2.10** Discord OAuth — `GET /login`, `POST /register`
- **2.11** Kakao OAuth — `GET /kakaoLogin`, `POST /registerKakao`
- **2.12** `GET /getIntents`, `/checkAdmin`, `/checkIntent`, `/getAllUsers`
- **2.13** `GET /authorize`, `/unauthorize` (ADMIN required)
- **2.14** `POST /requestAdminIntent`
- **2.15** Phase 2 closeout: integration smoke test logs in with a real existing user from the DB (no password change) and verifies their JWT works against a protected endpoint

---

# PHASE 3: File Browsing + CRUD (Day 5-6)

**Goal:** File operations parity. Path traversal blocked. Activity logging works.

### Critical: Path safety upfront

`internal/files/safepath.go` is the most security-sensitive file in the project. Test it adversarially:
- `../../../etc/passwd`
- URL-encoded `..%2F`
- Null bytes
- Symlinks pointing outside data dir
- Absolute paths

Every file handler must call `SafeJoin(cfg.NASDataDir, untrustedLoc, untrustedName)` and reject any result outside the base.

### Tasks
- **3.1** `SafeJoin` with traversal tests (adversarial)
- **3.2** `GET /readFolder`
- **3.3** `GET /stat`
- **3.4** `GET /getTextFile`, `POST /saveTextFile`
- **3.5** `GET /makedir`
- **3.6** `GET /forceDelete` (with activity log)
- **3.7** `GET /copy` (recursive)
- **3.8** `GET /move`
- **3.9** `GET /rename`
- **3.10** `GET /searchInAllFiles` (using `filepath.WalkDir`)
- **3.11** `POST /log`, `GET /getActivityLog`
- **3.12** Phase 3 closeout

---

# PHASE 4: Streaming + Download (Day 7)

**Goal:** Video/audio/image streaming with range requests handled by `http.ServeContent` (which does proper Range/If-Modified-Since/sendfile).

### Tasks
- **4.1** Generic stream handler with intent check
- **4.2** `GET /getVideoData`
- **4.3** `GET /getAudioData`
- **4.4** `GET /getImageData`
- **4.5** `GET /download`
- **4.6** `GET /img` (static icons via embed.FS)
- **4.7** Integration test: range request returns 206 with correct bytes
- **4.8** Phase 4 closeout

---

# PHASE 5: tus Upload + Legacy Wrapper (Day 8-9)

**Goal:** Resumable uploads via tus protocol mounted at `/files/`. Legacy `POST /input` endpoint kept as a wrapper that internally streams to disk for old frontend code paths.

### Tasks
- **5.1** Mount `tusd` at `/files/` with filestore in `NAS_TEMP_DIR/tus`
- **5.2** Post-upload hook: move from tus staging → `NAS_DATA_DIR/<loc>/<name>`, write activity log
- **5.3** Auth on tus: pre-create hook verifies JWT + UPLOAD intent
- **5.4** Legacy `POST /input` wrapper (raw stream → disk, for backward compat)
- **5.5** `POST /inputZip` (tus upload + post-process unzip)
- **5.6** Integration test: full tus lifecycle — POST creation, PATCH chunks, HEAD status, file lands at final path
- **5.7** Phase 5 closeout

---

# PHASE 6: Archives (Day 10)

**Goal:** Zip/unzip streaming with progress tracking via in-memory store.

### Tasks
- **6.1** In-memory progress tracker (`sync.Map`-backed, TTL cleanup)
- **6.2** `POST /zipFiles` (streaming `archive/zip`)
- **6.3** `POST /unzipFile`
- **6.4** `GET /progress`
- **6.5** `GET /downloadZip`, `GET /deleteTempZip`
- **6.6** Phase 6 closeout

---

# PHASE 7: System Info + Admin (Day 11)

### Tasks
- **7.1** `GET /getSystemInfo` via gopsutil
- **7.2** Phase 7 closeout

---

# PHASE 8: Playwright E2E + Cutover (Day 12-14)

**Goal:** End-to-end test suite running real frontend against new Go backend. Production cutover.

### Tasks
- **8.1** Initialize Playwright project in `backend/e2e/`
- **8.2** E2E test: auth flow (register, login, JWT round-trip)
- **8.3** E2E test: browse, mkdir, rename, delete
- **8.4** E2E test: large file upload (1GB+) via tus + verify integrity
- **8.5** E2E test: video streaming (range request behavior)
- **8.6** E2E test: zip/unzip with progress
- **8.7** E2E test: admin user management
- **8.8** Performance baseline: 10x parallel downloads, 5x parallel uploads
- **8.9** Cutover: stop legacy Node service, ensure data volumes intact
- **8.10** Monitor for 48h; document any regressions in plan retrospective
- **8.11** Phase 8 closeout + ready for "remove backend.legacy/" decision

---

## Out of Scope (Future Plans)

These are explicitly NOT in this plan:
1. **Frontend SvelteKit migration** — separate plan (current Svelte 5 + Vite SPA stays, only API client URL changes)
2. **HLS transcoding for non-web video** — Phase 2 plan, after backend cutover stable
3. **Office document preview (LibreOffice)** — Phase 2 plan
4. **REST API cleanup** (turning `GET /forceDelete` into `DELETE /files/...`) — Phase 3 plan, requires coordinated frontend changes
5. **Prometheus metrics, structured logging beyond slog** — Phase 4 plan
6. **OAuth refresh tokens, session management improvements** — Phase 5 plan

---

## Plan Retrospective

(Fill in after each phase)

### Phase 1 — TBD
- What surprised me:
- Adjustments for next phase:

### Phase 2 — TBD
- …

---

## Execution Handoff

After plan sign-off, two execution modes available:

**Subagent-Driven (recommended for radical pace)** — fresh subagent per task, two-stage review, parallel where possible.

**Inline Execution** — execute in this session with checkpoints between Phases.
