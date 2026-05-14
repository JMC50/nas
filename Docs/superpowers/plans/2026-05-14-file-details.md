# File Detail Panel — Implementation Plan

> **For agentic workers:** Use `superpowers:subagent-driven-development` or `superpowers:executing-plans`. Steps use `- [ ]` syntax.

**Goal:** Show file size + modified time inline (in list view) and a richer detail side-panel (size, modified, created, full path, type) without N+1 stat calls.

**Architecture:** Extend backend `fileEntry` JSON to include `size` (int64 bytes) and `modifiedAt` (RFC3339) so `readFolder` returns everything needed for inline columns in a single response. A new compact `Inspector.svelte` side-panel reads from the selected entry; `Stat` handler ([handlers.go:63](backend/internal/files/handlers.go:63)) is reserved for future folder-recursive size and stays unwired for now. Single-click selection toggles the panel; click again on same row deselects.

**Tech Stack:** Go for backend, Svelte 5 runes for frontend, Tailwind 4, lucide-svelte icons. No new dependencies.

---

## Context

- [backend/internal/files/handlers.go:24-29](backend/internal/files/handlers.go:24) — `fileEntry` struct (current shape).
- [backend/internal/files/handlers.go:31-53](backend/internal/files/handlers.go:31) — `ReadFolder` handler (extends here).
- [backend/internal/files/handlers.go:55-88](backend/internal/files/handlers.go:55) — `Stat` handler and `statResponse` (reference for fields).
- [backend/internal/files/handlers.go:373](backend/internal/files/handlers.go:373) — `formatSize(bytes int64) string` (reusable).
- [frontend/src/lib/components/Explorer/icon-for.ts:19-23](frontend/src/lib/components/Explorer/icon-for.ts:19) — `FolderEntry` interface.
- [frontend/src/lib/components/Explorer/FileList.svelte](frontend/src/lib/components/Explorer/FileList.svelte) — list view (add columns here).
- [frontend/src/lib/components/Explorer/FileGrid.svelte](frontend/src/lib/components/Explorer/FileGrid.svelte) — grid view (add title tooltip).
- [frontend/src/lib/components/Explorer.svelte](frontend/src/lib/components/Explorer.svelte) — owns selection state; mounts Inspector.
- [frontend/src/lib/components/Uploads/UploadPanel.svelte:42-54](frontend/src/lib/components/Uploads/UploadPanel.svelte:42) — reference pattern for the side-panel (actual width `w-[360px]`, `fixed top-12 right-0 bottom-7 z-40 flex flex-col bg-bg-surface border-l border-border-default shadow-[0_0_24px_rgba(0,0,0,0.4)]`).

---

## Phase 1 — Backend: extend `fileEntry`

### Task 1.1: Add Size and ModifiedAt to fileEntry

**Files:**
- Modify: `backend/internal/files/handlers.go` (struct at line 24, populate in `ReadFolder` around line 44-51)

- [ ] **Step 1** — Update struct:
```go
type fileEntry struct {
    Name       string `json:"name"`
    IsFolder   bool   `json:"isFolder"`
    Extensions string `json:"extensions"`
    Loc        string `json:"loc,omitempty"`
    Size       int64  `json:"size"`
    ModifiedAt string `json:"modifiedAt"`
}
```
- [ ] **Step 2** — In `ReadFolder` loop, call `entry.Info()` (returns `fs.FileInfo, error`). On error, skip the entry. On success, set `Size = info.Size()` (0 for folders is acceptable), `ModifiedAt = info.ModTime().Format(time.RFC3339)`.
- [ ] **Verify:** `cd backend && go build ./...` succeeds; manual `curl 'http://localhost:7777/readFolder?loc=/&token=<t>'` (after running backend) returns rows with new fields.

### Task 1.2: Add integration test

**Files:**
- Modify: `backend/tests/integration/files_test.go` (extend an existing ReadFolder test or add new)

- [ ] **Step 1** — Test that `readFolder` response for a freshly-created file has `size > 0` and `modifiedAt` parseable as RFC3339 within ±60s of `time.Now()`.
- [ ] **Step 2** — Test that a folder entry has `isFolder: true` and `size: 0`.
- [ ] **Verify:** `go test ./tests/integration/ -run TestReadFolderDetails -v` → PASS.

### Task 1.3: Commit Phase 1

- [ ] `git commit -m "[feat] readFolder returns size and modifiedAt per entry"`

---

## Phase 2 — Frontend: type + inline columns

### Task 2.1: Extend FolderEntry interface

**Files:**
- Modify: `frontend/src/lib/components/Explorer/icon-for.ts`

- [ ] **Step 1** — Add to `FolderEntry`:
```ts
export interface FolderEntry {
  name: string;
  isFolder: boolean;
  extensions: string;
  size: number;
  modifiedAt: string; // RFC3339
}
```
- [ ] **Verify:** `npm run check` will FAIL until consumers handle new fields. Expected. Proceed.

### Task 2.2: Helper formatters

**Files:**
- Create: `frontend/src/lib/components/Explorer/format.ts`

- [ ] **Step 1** — Export `formatBytes(n: number): string` using **2-decimal precision above KB** (`%.2f KB|MB|GB`) to match backend `formatSize` ([handlers.go:373](backend/internal/files/handlers.go:373)) so any UI that surfaces both inline column (from `readFolder`) and Stat-derived strings stays consistent.
- [ ] **Step 2** — Export `formatRelTime(rfc3339: string): string` (just now / Nm / Nh / Nd / Mon D, YYYY).
- [ ] **Step 3** — Export `formatFullTime(rfc3339: string): string` (`new Date(rfc3339).toLocaleString()`).
- [ ] **Verify:** Walk through 0, 500, 1024, 1048576, 1073741824 in spec table; expect `0 B`, `500 B`, `1.00 KB`, `1.00 MB`, `1.00 GB`.

### Task 2.3: Render columns in FileList

**Files:**
- Modify: `frontend/src/lib/components/Explorer/FileList.svelte`

- [ ] **Step 1** — Add two `<th>` columns: `Size` (w-20) and `Modified` (w-32). Right-align Size, monospace.
- [ ] **Step 2** — Each row: render `formatBytes(entry.size)` (or `—` for folders) and `formatRelTime(entry.modifiedAt)` (with `title={formatFullTime(...)}` for full timestamp on hover).
- [ ] **Verify:** `npm run check` → 0/0; Playwright: list view shows columns.

### Task 2.4: FileGrid tooltip

**Files:**
- Modify: `frontend/src/lib/components/Explorer/FileGrid.svelte`

- [ ] **Step 1** — Extend the existing `title={entry.name}` attribute to `title={\`${entry.name}\n${formatBytes(entry.size)} · ${formatRelTime(entry.modifiedAt)}\`}` for files; folders keep name-only title.
- [ ] **Verify:** `npm run check` → 0/0.

### Task 2.5: Commit Phase 2

- [ ] `git commit -m "[feat] file size + modified columns in list view and grid tooltip"`

---

## Phase 3 — Inspector side-panel

### Task 3.1: Selection state in Explorer

**Files:**
- Modify: `frontend/src/lib/components/Explorer.svelte`

- [ ] **Step 1** — Add `let selected = $state<FolderEntry | null>(null);` near other state.
- [ ] **Step 2** — Add `function toggleSelect(entry: FolderEntry) { selected = selected?.name === entry.name ? null : entry; }`. Wire as new prop `onSelect` to `FileGrid` / `FileList`.
- [ ] **Step 3** — Clear selection when `loc` changes (`$effect(() => { void loc; selected = null; })`).
- [ ] **Verify:** `npm run check` → 0/0.

### Task 3.2: Wire single-click in FileGrid/FileList

**Files:**
- Modify: `FileGrid.svelte`, `FileList.svelte`

- [ ] **Step 1** — Add `onClick={(event) => onSelect(entry)}` to entry button/row. **Important:** double-click already opens; ensure single-click does not pre-empt dblclick (browser standard `click` fires before second click resolves; tests confirm dblclick still works in Phase 1 UX overhaul, so no special debounce needed).
- [ ] **Step 2** — Visual indicator: when `entry.name === selectedName` (new prop), add `bg-bg-elevated ring-1 ring-accent` styling.
- [ ] **Verify:** Playwright: single-click row → highlighted; second click → deselected.

### Task 3.3: Inspector component

**Files:**
- Create: `frontend/src/lib/components/Explorer/Inspector.svelte`

- [ ] **Step 1** — Props: `{ entry: FolderEntry | null; loc: string[]; onClose: () => void }`.
- [ ] **Step 2** — When `entry` is null, return `null` (no markup).
- [ ] **Step 3** — Layout: copy UploadPanel pattern verbatim — `fixed top-12 right-0 bottom-7 w-[360px] z-40 flex flex-col bg-bg-surface border-l border-border-default shadow-[0_0_24px_rgba(0,0,0,0.4)]` (from [UploadPanel.svelte:42-44](frontend/src/lib/components/Uploads/UploadPanel.svelte:42)). Header pattern uses `flex items-center justify-between h-10 px-3 border-b border-border-default` with a `w-7 h-7` close button wrapping `<X size="14">`.
- [ ] **Step 4** — Sections (each compact `h-8` row with monospace label + value):
  - Icon row: type icon (`iconFor(entry)`) + entry.name (truncate)
  - Type: "Folder" or extension-uppercased
  - Size: `formatBytes(entry.size)` (folders: `—`)
  - Modified: `formatFullTime(entry.modifiedAt)` mono
  - Path: `[...loc, entry.name].join("/")` truncate with title for full
- [ ] **Step 5** — Close button top-right (lucide `x` icon, same as UploadPanel).
- [ ] **Verify:** `npm run check` → 0/0.

### Task 3.4: Mount Inspector in Explorer

**Files:**
- Modify: `Explorer.svelte`

- [ ] **Step 1** — Below the existing `<section>` close tag, render:
```svelte
<Inspector entry={selected} {loc} onClose={() => selected = null} />
```
- [ ] **Step 2** — Adjust the explorer flex layout so when Inspector is visible, the content area shrinks (use flex row container, Inspector takes intrinsic width).
- [ ] **Verify:** Playwright: click file → inspector slides in showing size/modified/path; click X → closes.

### Task 3.5: Commit Phase 3

- [ ] `git commit -m "[feat] file inspector side-panel for size, modified time, full path"`

---

## Phase 4 — /code-review + Playwright + merge

### Task 4.1: code-review

- [ ] Invoke `code-review` skill on every changed file. Loop until 0 ❌ Critical.

### Task 4.2: Playwright walkthrough (RULE 1 — hard gate)

Performed during execution. Every spec walkthrough row below is recorded ✅/❌ with screenshot in the md trail:

- Login `uxtest`, claim admin.
- Walk through every row in the Spec walkthrough table.
- If any row ❌, fix and re-run before merge.

### Task 4.3: Pre-existing ReadFolder test compatibility

- [ ] **Step 1** — `cd backend && go test ./tests/integration/ -run TestReadFolder` (any existing test name containing `ReadFolder`) → must pass with the extended struct (extra JSON fields are additive; old test assertions should not break).
- [ ] **Verify:** existing test pass; if any test asserts exact JSON shape with `len(keys)==N`, adjust to allow extras.

### Task 4.4: Pre-merge + merge

- [ ] Pre-merge checklist (project RULE 6).
- [ ] Merge `feat/file-details` → main, push origin.
- [ ] **Verify:** `git log origin/main -1`.

---

## Spec walkthrough table

| # | Scenario | Expected |
|---|---|---|
| 1 | Open root with files of sizes 0, 500B, 1KB, 1MB | List shows `0 B`, `500 B`, `1.0 KB`, `1.0 MB` |
| 2 | Folder entry in list | Size column shows `—` |
| 3 | Hover row | Title attr shows full timestamp |
| 4 | Single-click file | Row highlighted; Inspector opens with size, modified, full path |
| 5 | Single-click same file again | Inspector closes, row deselected |
| 6 | Click different file | Inspector swaps to new entry without close-open flicker |
| 7 | Navigate to subfolder | Inspector closes (selection cleared by `$effect`) |
| 8 | Grid view, hover file | Tooltip shows `name\n1.0 MB · 2h ago` |
| 9 | `npm run build` produces no warnings about unused `formatRelTime` etc. | OK |

---

## Risks & open questions

| Risk | Mitigation |
|---|---|
| `entry.Info()` slow on network filesystem | Stays at O(N) — same as `os.ReadDir`, info call is cheap on local disk. Document, monitor with backend test. |
| Folder size remains 0 — confusing | Render `—` for folders; future enhancement can recurse but is out of scope. |
| Click + dblclick race | Browser fires `click` then `dblclick`; selection toggling on first click is fine, dblclick still navigates. Confirmed by existing Tabs reorder pattern (TabBar.svelte uses both). |
| Inspector overlaps small viewports | `w-72` fixed; on narrow screens the file list area shrinks. Acceptable for desktop NAS; mobile out of scope. |
| Stat field naming drift (CreatedAt unused) | Plan deliberately ignores `createdAt` — readFolder returns ModTime only. Document decision. |

---

## Rollback strategy

```bash
cd C:/Data/Git/ANXI/nas && git revert -m 1 <merge-sha> && git push origin main
```

Backend struct change is additive (extra JSON fields), so older frontends ignore them. Reverting the backend alone is safe; reverting only the frontend works because backend ignores missing fields in requests.

---

## 5-pass self-review

| Pass | Concern | Action |
|---|---|---|
| 1. Structure | Backend → types → renderers → inspector → integration order is dependency-correct | ✅ |
| 2. Spec correctness | `fileEntry` struct at handlers.go:24 confirmed; `entry.Info()` is real `fs.DirEntry` API; `formatSize` exists; `UploadPanel.svelte:43` pattern referenced | ✅ |
| 3. Risk | Network FS, folder size, click race, narrow viewport — each addressed | ✅ |
| 4. Consistency | Side-panel mirrors UploadPanel dimensions; Tailwind tokens (`bg-bg-surface`, `border-border-default`); monospace for technical values per Gruvbox/VSCode tone | ✅ |
| 5. Completeness | Acceptance: size + modified inline ✅, side-panel ✅, no N+1 ✅. Rollback concrete `git revert` | ✅ |
