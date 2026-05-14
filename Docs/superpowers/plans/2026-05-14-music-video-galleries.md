# Music + Video Library Tabs — Implementation Plan

> **For agentic workers:** Use `superpowers:subagent-driven-development` or `superpowers:executing-plans`. Steps use `- [ ]` syntax.

**Goal:** Add two new top-level tabs — `Music` (Spotify-style grid of audio files grouped/sorted) and `Videos` (YouTube-style poster grid) — that index the NAS recursively, render a virtualized gallery, and play tracks/videos via the existing `MediaViewer`.

**Architecture:**
- Backend: new endpoint `GET /server/mediaLibrary?kind=audio|video` that walks `NASDataDir`, filters by extension, returns `{ name, loc, size, modifiedAt, durationMs?, posterUrl? }` JSON. Phase 1 returns filename-only metadata (no ID3, no ffmpeg) and a deterministic icon fallback. ID3 tag extraction and ffmpeg thumbnail are explicit follow-up phases gated behind a build-flag check.
- Frontend: new `TabKind` values `music-library` and `video-library`, new sidebar nav entries, new `MusicLibrary.svelte` / `VideoLibrary.svelte` components. Click-to-play opens a viewer tab using the existing `MediaViewer` and `tabs.open` plumbing.
- Pagination: virtual scrolling via `@tanstack/svelte-virtual` (already used elsewhere? — confirm in Phase 0; if not, plan adds dep). If huge library is unrealistic for the user's data set, fall back to plain CSS grid + lazy-render via IntersectionObserver pattern already proven by PdfViewer.

**Tech Stack:** Go for backend recursive walk; Svelte 5 runes, Tailwind 4, lucide-svelte for frontend; optional: virtual-scroll lib.

---

## Context

- [backend/internal/files/handlers.go:250](backend/internal/files/handlers.go:250) — existing `Search` handler uses `filepath.WalkDir` recursively; reference for the new media-walk endpoint (use `filepath.WalkDir` for parity, not `filepath.Walk`).
- [backend/internal/server/router.go:117](backend/internal/server/router.go:117) — route wiring (`addFileRoute`).
- [frontend/src/lib/types.ts:37-49](frontend/src/lib/types.ts:37) — `TabKind` union.
- [backend/internal/config/config.go](backend/internal/config/config.go) — register new `MediaLibraryLimit` field for `MEDIA_LIB_LIMIT` env var.
- [frontend/src/lib/components/Shell/VerticalNav.svelte](frontend/src/lib/components/Shell/VerticalNav.svelte) — sidebar nav (current items: Files, Activity, Users, Settings, System).
- [frontend/src/lib/components/Tabs/TabContent.svelte](frontend/src/lib/components/Tabs/TabContent.svelte) — tab routing.
- [frontend/src/lib/components/Viewers/MediaViewer.svelte](frontend/src/lib/components/Viewers/MediaViewer.svelte) — existing audio/video player.
- [frontend/src/lib/components/Viewers/registry.ts](frontend/src/lib/components/Viewers/registry.ts) — `VIDEO_EXTENSIONS`, `AUDIO_EXTENSIONS` to reuse.

**Confirmed absence:** No ffmpeg invocation in `backend/` per project search; thumbnail generation is a Phase 3 add-on, not assumed.

---

## Phase 0 — Pre-flight & dependency choice

### Task 0.1: Audit deps

- [ ] **Step 1** — `cat frontend/package.json | grep -E "virtual|tanstack"` to confirm whether a virtual-scroll lib is present.
- [ ] **Step 2** — Decide: if present, reuse; if not, use IntersectionObserver lazy-render (proven pattern in [PdfViewer.svelte](frontend/src/lib/components/Viewers/PdfViewer.svelte)). Plan defaults to IntersectionObserver to avoid new dep until measured need.
- [ ] **Verify:** Decision noted in execution md trail.

---

## Phase 1 — Backend media-library endpoint

### Task 1.0: Register `MEDIA_LIB_LIMIT` in config

**Files:**
- Modify: `backend/internal/config/config.go`

- [ ] **Step 1** — Add field `MediaLibraryLimit int` to the `Config` struct.
- [ ] **Step 2** — In the loader, `c.MediaLibraryLimit = getEnvInt("MEDIA_LIB_LIMIT", 5000)` (mirror existing `getEnvInt`/`getEnv` helper pattern in same file).
- [ ] **Verify:** `cd backend && go build ./...` succeeds.

### Task 1.1: New handler `MediaLibrary`

**Files:**
- Modify: `backend/internal/files/handlers.go` (add after `Search`)

- [ ] **Step 1** — Add `mediaEntry` struct with **separate `loc` (directory) and `name` (filename)** to match the existing `FilePayload` shape on the frontend (TabContent.svelte's `FilePayload` is `{ loc: string; name: string }`):
```go
type mediaEntry struct {
    Name       string `json:"name"`       // filename only, e.g. "song.mp3"
    Loc        string `json:"loc"`        // directory loc, e.g. "/Music/2024" (no trailing slash, no filename)
    Extensions string `json:"extensions"`
    Size       int64  `json:"size"`
    ModifiedAt string `json:"modifiedAt"`
    Kind       string `json:"kind"` // "audio" | "video"
}
```
- [ ] **Step 2** — Add `audioExt` and `videoExt` maps inside the package, mirroring frontend [registry.ts](frontend/src/lib/components/Viewers/registry.ts) lists (note: those constants are not exported; replicate the values, do not import).
- [ ] **Step 3** — Implement `func (h *Handlers) MediaLibrary(w, r)`:
  - Parse `kind` from query (`audio` or `video`; default error 400).
  - `filepath.WalkDir` from `h.Config.NASDataDir` (matches `Search` handler convention at handlers.go:250). Skip `.`-prefixed directories using `fs.SkipDir`.
  - For each file matching the extension set: compute `Loc` = `"/" + filepath.ToSlash(filepath.Dir(rel))` where `rel` is path relative to NASDataDir, and `Name` = `filepath.Base(path)`. Stat the entry for size/modtime.
  - Cap result count at `h.Config.MediaLibraryLimit` (default 5000 via Task 1.0); if exceeded, break the walk early and set `X-Library-Truncated: true` response header.
  - Write JSON.
- [ ] **Verify:** `go build ./...` succeeds.

### Task 1.2: Route wiring

**Files:**
- Modify: `backend/internal/server/router.go`

- [ ] **Step 1** — Add:
```go
addFileRoute(r, "GET", "/mediaLibrary", auth.IntentView, requireToken, conn, fileHandlers.MediaLibrary)
```
- [ ] **Verify:** `go build ./...`; `curl 'http://localhost:7777/mediaLibrary?kind=audio&token=<t>'` returns JSON array.

### Task 1.3: Integration test

**Files:**
- Modify: `backend/tests/integration/files_test.go`

- [ ] **Step 1** — Seed test NAS with `a.mp3`, `b/c.mp3`, `b/d.mp4`, `.hidden/e.mp3`.
- [ ] **Step 2** — `GET /mediaLibrary?kind=audio` → entries for `a.mp3` and `b/c.mp3` (hidden skipped).
- [ ] **Step 3** — `GET /mediaLibrary?kind=video` → only `b/d.mp4`.
- [ ] **Verify:** `go test ./tests/integration/ -run TestMediaLibrary -v` → PASS.

### Task 1.4: Commit

- [ ] `git commit -m "[feat] mediaLibrary endpoint walks NAS for audio/video entries"`

---

## Phase 2 — Frontend types + nav

### Task 2.1: Extend TabKind

**Files:**
- Modify: `frontend/src/lib/types.ts`

- [ ] **Step 1** — Add `"music-library"` and `"video-library"` to the `TabKind` union.
- [ ] **Step 2** — Add `interface MediaEntry { name; loc; extensions; size; modifiedAt; kind: "audio"|"video"; }`.
- [ ] **Verify:** `npm run check` → fails (TabContent must handle new kinds; expected — proceed to next task).

### Task 2.2: Wire TabContent

**Files:**
- Modify: `frontend/src/lib/components/Tabs/TabContent.svelte`

- [ ] **Step 1** — Add branches: `{:else if tab.kind === "music-library"} <MusicLibrary />` and same for `"video-library"` with `<VideoLibrary />`. Stub the imports (components created in Phase 3).
- [ ] **Verify:** Will fail until Phase 3 creates components; acceptable.

### Task 2.3: Add nav entries

**Files:**
- Modify: `frontend/src/lib/components/Shell/VerticalNav.svelte`

- [ ] **Step 1** — Add two entries: "Music" (lucide `music` icon) and "Videos" (lucide `film` icon). Click handler: `tabs.open({ kind: "music-library", title: "Music", icon: "music", payload: null, closable: false })` (and same for videos).
- [ ] **Verify:** Defer until Phase 3 components exist.

### Task 2.4: TabBar icon mapping

**Files:**
- Modify: `frontend/src/lib/components/Tabs/TabBar.svelte`

- [ ] **Step 1** — Add `"music-library": Music, "video-library": Film` to `KIND_TO_ICON`.

---

## Phase 3 — Library components

### Task 3.1: Shared library hook

**Files:**
- Create: `frontend/src/lib/components/Library/loader.ts`

- [ ] **Step 1** — Export `async function loadLibrary(kind: "audio"|"video"): Promise<MediaEntry[]>` calling `/server/mediaLibrary?kind=...&token=<auth.token>`.
- [ ] **Step 2** — Throw on non-200; surface error.
- [ ] **Verify:** `npm run check` → 0/0 after types are in.

### Task 3.2: MusicLibrary.svelte

**Files:**
- Create: `frontend/src/lib/components/Library/MusicLibrary.svelte`

- [ ] **Step 1** — Load on mount; render compact list grouped by folder (`entry.loc` IS the directory loc — see Task 1.1 Step 1 — group rows by it directly to form "album" sections). Each row: lucide `music` icon · filename (stripped of extension) · folder · size · `formatRelTime(modifiedAt)`. Click → `tabs.open({ kind: "audio", title: entry.name, icon: "audio", payload: { loc: entry.loc, name: entry.name }, closable: true })` (payload shape is now native `{loc, name}` because Task 1.1 split them; no path splitting needed on frontend).
- [ ] **Step 2** — Search box at top (text filter on filename + folder).
- [ ] **Step 3** — Performance: render only first 200 entries initially; render more on scroll near bottom via IntersectionObserver (sentinel element).
- [ ] **Verify:** Playwright: list scrolls smoothly; click row → audio viewer tab opens; search "wav" filters.

### Task 3.3: VideoLibrary.svelte

**Files:**
- Create: `frontend/src/lib/components/Library/VideoLibrary.svelte`

- [ ] **Step 1** — Grid layout (`grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-4 p-6`). Each tile: 16:9 dark placeholder showing lucide `film` icon center (until thumbnails added in a future plan), filename overlay bottom (truncate), size · duration (`—` until ID3/ffmpeg), folder label small muted.
- [ ] **Step 2** — Click tile → `tabs.open({ kind: "video", ... })`.
- [ ] **Step 3** — Search box top + IntersectionObserver lazy reveal (load 60 tiles initially, +60 per sentinel).
- [ ] **Verify:** Playwright: click tile → video viewer tab; search filters; many entries render without lag.

### Task 3.4: Commit Phase 3

- [ ] `git commit -m "[feat] music and video library tabs with click-to-play"`

---

## Phase 4 — `/code-review` + Playwright + merge

### Task 4.1: code-review

- [ ] Run on every new/changed file. Auto-fix Critical.

### Task 4.2: Playwright walkthrough

| # | Action | Expected |
|---|---|---|
| 1 | Open Music tab (sidebar) | List of all .mp3/.wav/etc. across NAS appears |
| 2 | Click first track | New `audio` tab opens with `MediaViewer` playing |
| 3 | Search "track" | List filters |
| 4 | Switch to Videos tab | Tile grid; placeholders visible |
| 5 | Click tile | `video` tab opens with `MediaViewer` |
| 6 | Scroll to bottom of long list | More entries lazy-load via IntersectionObserver |
| 7 | Theme toggle | Both libraries respect tokens |
| 8 | Open existing Files tab | Still works (no regression in multi-Explorer tabs) |

### Task 4.3: Pre-merge + merge

- [ ] RULE 6 checklist. Merge `feat/media-libraries` → main. Push origin.

---

## Phase 5 (FUTURE, gated) — Thumbnails & ID3

Out of scope for the initial merge. Documented for the follow-up plan:

- Backend ffmpeg dependency check → poster generation on demand (cache in `data/tmp/posters/<hash>.jpg`).
- ID3 tag parsing (Go lib `dhowden/tag` or similar) → album/artist/cover in JSON response.
- Backend can stay deterministic; if ffmpeg absent, return `posterUrl: null` and frontend shows lucide placeholder.

---

## Risks & open questions

| Risk | Mitigation |
|---|---|
| `filepath.WalkDir` over a 100k-file NAS → multi-second latency | Cap at `h.Config.MediaLibraryLimit` (default 5000) with truncation header. Frontend shows banner "Library truncated; refine via search". Early-break out of walk once limit hit. |
| Out-of-memory on backend with deeply nested data | `filepath.WalkDir` is iterator-style; memory bounded by entries count, which is capped. |
| `.`-prefixed dir skip is heuristic — user may want to see them | Document; future flag if requested. |
| Click-to-play needs full `loc` string and `name` matching MediaViewer payload | `MediaEntry.loc` is the directory loc (no filename), matching existing FilePayload shape (`{loc, name}` separately). Confirmed in TabContent.svelte. |
| Theme contrast on placeholder tile | Use `bg-bg-elevated` + `text-fg-muted` — same tokens as ForbiddenPanel |
| Search complexity on 5000 rows | Plain JS `.filter` is sub-millisecond at this size; no virtualization library needed yet |
| New nav entries may clutter sidebar | Stays at 7 items after adding 2 (Files / Music / Videos / Activity / Users / Settings / System). Acceptable, but verify with screenshot during walkthrough. |
| `MEDIA_LIB_LIMIT` env var declaration | Promoted to Task 1.0 as explicit prerequisite sub-step before MediaLibrary handler implementation. |

---

## Rollback strategy

```bash
cd C:/Data/Git/ANXI/nas && git revert -m 1 <merge-sha> && git push origin main
```

Backend endpoint is additive; reverting removes nav entries and tab components. No data loss possible (read-only walk).

---

## 5-pass self-review

| Pass | Concern | Action |
|---|---|---|
| 1. Structure | Phases 0→4 (Phase 5 out-of-scope clearly marked); each Phase has tasks; each task has step + verify | ✅ |
| 2. Spec correctness | `filepath.Walk` is real Go API; `addFileRoute` signature matches router.go:117 line; lucide icons `music`/`film` confirmed in TabBar imports; MediaViewer payload shape confirmed in TabContent | ✅ |
| 3. Risk | Walk perf, OOM, skipping dotfiles, ffmpeg absent, sidebar clutter, env var registration — each mitigated | ✅ |
| 4. Consistency | Reuses `formatBytes`/`formatRelTime` (planned in file-details plan), Tailwind tokens, monospace small text. Click-to-play uses existing `tabs.open` API. | ✅ |
| 5. Completeness | Acceptance: Music tab ✅, Video tab ✅, recursive walk ✅, click-to-play ✅, pagination ✅. ID3/thumbnails explicitly deferred. Rollback concrete. | ✅ |

**Cross-plan note:** `formatBytes` / `formatRelTime` referenced here originate in the file-details plan ([file-details](2026-05-14-file-details.md) Phase 2 Task 2.2). If file-details merges first, reuse. If this plan merges first, hoist the same helpers from the corresponding tasks here instead.
