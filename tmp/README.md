# tmp/ — scratch directory for transient test data

This folder is `.gitignored` (see root `.gitignore`). Use it as the single home for
**all** transient data produced while testing or developing locally.

## Guidance for test runs

- **Backend integration tests** that need a temporary NAS storage path: set
  `NAS_DATA_DIR=./tmp/nas` before running the suite. The backend will read/write
  there and you can wipe the whole `tmp/` between runs.
- **Playwright walkthroughs** that create test files: drop them under
  `./tmp/playwright-uploads/` or similar; do not pollute `data/nas/` (the dev
  default storage root).
- **TUS handler scratch**: the upload server keeps partial chunks at
  `backend/tests/integration/tus/`, which is now gitignored. If you want them
  somewhere else, pass `--tus-store=./tmp/tus` (or the equivalent env var).
- **Screenshots, console logs, network captures**: write under
  `./tmp/screenshots/`, `./tmp/logs/`, etc.

## Already-ignored fixed paths

Some tools write to hardcoded paths that we cannot redirect; they're still
gitignored so artifacts don't sneak in:

- `.playwright-mcp/` — Playwright MCP console-log cache
- `backend/tests/integration/tus/*` (except `.gitkeep`) — TUS scratch
- `data/nas/Music/`, `data/nas/Videos/` — past test upload spillover paths

If you find a new tool writing junk into the repo, prefer redirecting it here
(via env/CLI flag) over adding a per-tool ignore line.

## `recovered-from-parent/`

If you see this subfolder, it contains data that previously leaked **above**
the repo (siblings to the project root: `C:\Data\Git\ANXI\nas-data\`,
`nas-admin-data\`, `nas-db\`). Root cause: a backend instance was started
with relative env paths that traversed `..` out of `cwd` (e.g.
`NAS_DATA_DIR=../nas-data`).

Mitigations now in place:

- `backend/internal/config/config.go` logs `slog.Warn` at startup when any of
  `NAS_DATA_DIR`, `NAS_ADMIN_DATA_DIR`, `NAS_TEMP_DIR`, `DB_PATH`, `FRONTEND_DIR`
  resolves outside cwd. Look for `path resolves outside cwd` in the log.
- `.env.example` documents the safe defaults and the footgun.

Once you've reviewed the recovered contents and decided what (if anything) to
salvage, delete this subfolder — `tmp/*` is gitignored so it won't be tracked.
