# Folder Drag-and-Drop Upload — Implementation Plan

> **For agentic workers:** Use `superpowers:subagent-driven-development` or `superpowers:executing-plans`. Steps use `- [ ]` syntax for tracking.

**Goal:** Make folder drag-and-drop and `Upload folder` picker reliably upload an entire nested folder tree to the corresponding NAS sub-path, with verifiable Playwright e2e coverage.

**Architecture:** Frontend already collects all files with sub-paths via `webkitGetAsEntry()` recursion (DragDropOverlay.svelte) and `webkitdirectory` input (Explorer.svelte's `onPickFolder`). Backend TUS `finalizeUpload` already runs `os.MkdirAll(filepath.Dir(target), 0o755)` ([tus.go:112](backend/internal/upload/tus.go:112)) so per-file sub-paths auto-create intermediate folders. This plan hardens the collection path (depth/symlink/error handling), adds an integration test for nested upload, and defines a deterministic Playwright proxy for verification.

**Tech Stack:** Svelte 5 (runes), SvelteKit, TypeScript, Tailwind 4, lucide-svelte. Backend: Go, tusd v2, chi router. Test harness: Playwright MCP, Go integration tests (`backend/tests/integration/`).

---

## Context

- [DragDropOverlay.svelte](frontend/src/lib/components/Uploads/DragDropOverlay.svelte) — `walkEntry()` recurses via `FileSystemDirectoryEntry.createReader()`; collects `{file, loc}` then calls `uploads.enqueue` per file.
- [Explorer.svelte:128-148](frontend/src/lib/components/Explorer.svelte) — `onPickFolder` reads `webkitRelativePath` from each `File`, appends dir parts to `here` (current loc string), enqueues.
- [tus-client.ts](frontend/src/lib/tus-client.ts) — `metaHeader(loc, filename)` encodes `loc` and `name` as base64 in `Upload-Metadata`.
- [backend/internal/upload/tus.go:100-132](backend/internal/upload/tus.go) — `finalizeUpload` parses `loc` + `filename`, `SafeJoin`s into NASDataDir, `os.MkdirAll` of `filepath.Dir(target)` then `os.Rename` from staging. **Sub-paths already work.**
- [backend/tests/integration/tus_e2e_test.go](backend/tests/integration/tus_e2e_test.go) — existing TUS e2e harness for adding the nested-folder integration test.

---

## Phase 1 — Backend integration test for nested folder upload

### Task 1.1: Add integration test exercising sub-path creation

**Files:**
- Modify: `backend/tests/integration/tus_e2e_test.go`

- [ ] **Step 1** — Add `TestTusUploadCreatesNestedFolders` that performs three tus POST→PATCH sequences with `loc` metadata pointing to `/A/B/C` (three depth levels not pre-existing on disk).
- [ ] **Step 2** — Assert that after each upload, the file appears at `<NASDataDir>/A/B/C/<filename>` and that intermediate directories `<NASDataDir>/A`, `<NASDataDir>/A/B`, `<NASDataDir>/A/B/C` all exist with mode 0755.
- [ ] **Step 3** — Add a second sub-test uploading a file with `loc = /one/two/three/four/five` (5 levels) and a file at root level in the same test run, confirming no interference.
- [ ] **Verify:** `cd backend && go test ./tests/integration/ -run TestTusUploadCreatesNestedFolders -v` → PASS

### Task 1.2: Commit Phase 1

- [ ] `git add backend/tests/integration/tus_e2e_test.go`
- [ ] `git commit -m "[test] tus upload creates intermediate folders for nested loc"`
- [ ] **Verify:** `git log -1 --oneline` shows the commit

---

## Phase 2 — Frontend collection-path robustness

### Task 2.1: Cap recursion depth and skip dot-folders in `walkEntry`

**Files:**
- Modify: `frontend/src/lib/components/Uploads/DragDropOverlay.svelte` (the `walkEntry` function)

- [ ] **Step 1** — Add `depth` parameter to `walkEntry`, default 0. If `depth > 32`, push a notification "Folder depth exceeded (32) — partial upload" and return.
- [ ] **Step 2** — Skip entries whose `name` starts with `.` (dot-files / dot-folders) silently. Rationale: macOS `.DS_Store`, `.git`, `__MACOSX` pollution from zip-extracted folders.
- [ ] **Step 3** — Wrap each `reader.readEntries` call in `try/catch`; on error, surface one notification "Failed to read part of dropped folder — uploads may be incomplete" but continue with collected files.
- [ ] **Verify:** `cd frontend && npm run check` → 0/0; manual unit reasoning that `walkEntry` short-circuits at depth 33.

### Task 2.2: Apply equivalent skip-dotfiles guard to `onPickFolder`

**Files:**
- Modify: `frontend/src/lib/components/Explorer.svelte` — `onPickFolder` function (locate by name; line numbers shift as features land)

- [ ] **Step 1** — Before enqueuing, skip `File`s whose `webkitRelativePath` segment starts with `.` (any segment, e.g. `myfolder/.git/HEAD`).
- [ ] **Verify:** `npm run check` → 0/0.

### Task 2.3: Surface upload-set summary toast

**Files:**
- Modify: `frontend/src/lib/components/Uploads/DragDropOverlay.svelte`, `frontend/src/lib/components/Explorer.svelte`

- [ ] **Step 1** — After `walkEntry` / `onPickFolder` enqueue loop completes, call `notifications.info(\`Queued <count> file(s) for upload\`)` with the actual count.
- [ ] **Step 2** — If count is 0 after dropping a folder, call `notifications.warning(\`No uploadable files in dropped folder\`)`.
- [ ] **Verify:** `npm run check` → 0/0.

### Task 2.4: Commit Phase 2

- [ ] `git add frontend/src/lib/components/Uploads/DragDropOverlay.svelte frontend/src/lib/components/Explorer.svelte`
- [ ] `git commit -m "[feat] folder upload: depth cap, dotfile skip, summary toast"`

---

## Phase 3 — Playwright e2e verification (RULE 1)

Playwright cannot scriptably simulate OS-level folder drag with `webkitGetAsEntry`. Use this proxy approach (per project policy: "Acceptable verify forms: spec table walked through against actual call sites"):

### Task 3.1: Browser-evaluate harness for folder upload

**Files:**
- Modify: `frontend/src/lib/components/Uploads/DragDropOverlay.svelte` — expose a dev-only `__nasTestEnqueueFolder` on `window` behind `import.meta.env.DEV`

- [ ] **Step 1** — Inside `onMount`, if `import.meta.env.DEV`, attach a test helper to `window` that takes an array of `{ name, webkitRelativePath, content }` plain objects, constructs `File` objects, and runs them through the same enqueue+sub-loc-derivation code path that `onPickFolder` uses.
- [ ] **Step 2** — Add a comment "test-only DEV harness; treeshaken from prod build" and verify Vite drops it (check production bundle does not contain the symbol).
- [ ] **Verify:** `cd frontend && npm run build` then `grep -r "__nasTestEnqueueFolder" frontend/build/` → no matches.

### Task 3.2: Playwright walkthrough

Performed during execution, not committed:

| # | Action | Expected |
|---|---|---|
| 1 | Login as `uxtest`, claim admin | Files explorer shows |
| 2 | Click "Upload folder" button, pick a real OS folder containing `a/b/c.txt` (manual once) OR via `__nasTestEnqueueFolder([{webkitRelativePath:"deepA/sub/c.txt", content:"x"}])` | Toast `Queued 1 file(s) for upload`; uploads panel shows progress |
| 3 | After upload completes, navigate `/deepA/sub` → `c.txt` listed | OK |
| 4 | Drag empty folder (Use harness with `[]`) | Toast `No uploadable files in dropped folder` |
| 5 | Harness with `webkitRelativePath:".git/HEAD"` | Skipped, no upload enqueued |
| 6 | Harness with depth 35 path | Toast `Folder depth exceeded (32) — partial upload`; files up to depth 32 uploaded |
| 7 | Visit `/deepA/sub`, drag a file out of the folder (existing drag-drop move) | Move works regardless |

Spec walkthrough table is the **hard gate per project RULE 2**: every row must record ✅/❌ with screenshot reference in the execution md trail.

### Task 3.3: Commit Phase 3

- [ ] `git add frontend/src/lib/components/Uploads/DragDropOverlay.svelte`
- [ ] `git commit -m "[test] DEV-only enqueue-folder harness for Playwright verification"`

---

## Phase 4 — `/code-review` + merge

- [ ] **Step 1** — Invoke `code-review` skill on every changed file. Loop until 0 ❌ Critical.
- [ ] **Step 2** — Confirm pre-merge checklist (project RULE 6): npm check/build pass, Playwright walkthrough table fully ticked, md trail updated.
- [ ] **Step 3** — Merge `feat/folder-upload-hardening` → main, push origin.
- [ ] **Verify:** `git log origin/main -1` shows merge commit.

---

## Risks & open questions

| Risk | Mitigation |
|---|---|
| **TUS staging dir fills up on huge folder dump** | Existing tusd cleanup runs after `finalizeUpload`; verify no orphan staging entries after Task 1.1 integration test. If found, add cleanup. |
| **Browser blocks `webkitGetAsEntry` for sandboxed iframes** | NAS frontend is top-level; not applicable today. Note in plan. |
| **Symlink loops inside dropped folder** | `webkitGetAsEntry` is a snapshot DOM API and does not follow OS symlinks in practice. Depth cap (32) is final guard. |
| **Files with path-traversal names (`../escape.txt`)** | Backend `SafeJoin` ([files/safepath.go](backend/internal/files/safepath.go)) rejects `..` segments. Verified in `safepath_test.go`. |
| **Hangul / non-ASCII filenames** | tus metadata is base64-UTF-8 already; verified earlier in download-filename fix. |
| **No backend mkdir-p needed** | Already confirmed at tus.go:112 — `MkdirAll` runs per file. Plan therefore touches no backend production code. |
| **DEV harness leaking into prod build** | Vite tree-shaking under `import.meta.env.DEV` guard verified in Task 3.1 step 2. |

---

## Rollback strategy

If post-merge regression appears:
```bash
cd C:/Data/Git/ANXI/nas && git revert -m 1 <merge-sha> && git push origin main
```
Each phase is committed independently so partial revert is also possible:
```bash
git log --oneline feat/folder-upload-hardening..HEAD  # find which commits to revert
```

---

## 5-pass self-review

| Pass | Concern | Action |
|---|---|---|
| 1. Structure | All phases have steps; every step has a verify; ordering: backend test first (proves invariant), then frontend hardening, then Playwright | ✅ |
| 2. Spec correctness | `tus.go:112` MkdirAll verified by reading the file; `tus_e2e_test.go` exists; `walkEntry` exists in DragDropOverlay; `onPickFolder` exists in Explorer.svelte | ✅ |
| 3. Risk | Depth, dotfiles, symlinks, traversal, staging cleanup, DEV harness leak — each mitigated above | ✅ |
| 4. Consistency | Toast colors via existing `notifications.info/warning`; Tailwind classes unchanged; no chart libs needed; commit format `[type] subject` | ✅ |
| 5. Completeness | Acceptance criteria from spec (folder drag-drop reliable, nested sub-paths, dotfile skip, Playwright-coverable) all map to tasks; rollback is concrete `git revert` | ✅ |
