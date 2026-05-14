# Marquee + Multi-Select for Files/Folders — Implementation Plan

> **For agentic workers:** Use `superpowers:subagent-driven-development` or `superpowers:executing-plans`. Steps use `- [ ]` syntax.

**Goal:** Enable selecting multiple files/folders via (a) drag a rectangular marquee over a blank area in FileGrid/FileList, (b) Ctrl/Cmd+click to toggle, (c) Shift+click for range, (d) Ctrl+A to select all, (e) Esc to clear. Selected entries highlight with accent ring + background; subsequent actions (delete, move via drag, future copy/zip) apply to all selected.

**Architecture:**
- Per-tab selection lives in `ExplorerPayload` (new `selection: string[]` field of entry names, ordered by selection order for Shift+range anchor). Persisting in tab payload keeps multi-Explorer-tabs isolated.
- Marquee state is component-local in Explorer.svelte (drag rectangle start point + current rectangle).
- Bounding-box intersection: each entry button/row exposes its `data-entry-name` attribute and the parent container computes `getBoundingClientRect()` per entry once at marquee-start time (snapshot to avoid layout thrash during drag).
- Drag-move integration: when an already-selected entry is dragged, the dataTransfer payload becomes a JSON array `[{name, isFolder, sourceLoc}, …]` (extending the single-entry format from feat/ux-overhaul). `performMove` (in drag-drop.ts) gains a batch sibling `performMoveBatch`.

**Tech Stack:** Svelte 5 runes, TypeScript, native DOM `getBoundingClientRect`. No new deps.

---

## Context

- [frontend/src/lib/types.ts:33-35](frontend/src/lib/types.ts:33) — `ExplorerPayload` interface (currently `{ loc: string[] }` only).
- [frontend/src/lib/components/Explorer.svelte](frontend/src/lib/components/Explorer.svelte) — owns tab payload mutations via `tabs.update(tabId, { payload: ... })`.
- [frontend/src/lib/components/Explorer/FileGrid.svelte](frontend/src/lib/components/Explorer/FileGrid.svelte), [FileList.svelte](frontend/src/lib/components/Explorer/FileList.svelte) — entry render; will receive new `selected: Set<string>` prop + click handlers.
- [frontend/src/lib/components/Explorer/drag-drop.ts](frontend/src/lib/components/Explorer/drag-drop.ts) — `NAS_ENTRY_MIME`, `buildPayload`, `readPayload`, `performMove`.
- [frontend/src/lib/store/tabs.svelte.ts](frontend/src/lib/store/tabs.svelte.ts) — `tabs.update(id, partial)` API.

---

## Phase 1 — Selection in ExplorerPayload

### Task 1.1: Extend type

**Files:**
- Modify: `frontend/src/lib/types.ts`

- [ ] **Step 1** — Update:
```ts
export interface ExplorerPayload {
  loc: string[];
  selection?: string[]; // entry names, last item is the range-anchor
}
```
- [ ] **Verify:** `npm run check` → 0/0 (optional field; consumers tolerant).

### Task 1.2: Helper functions

**Files:**
- Create: `frontend/src/lib/components/Explorer/selection.ts`

- [ ] **Step 1** — Export:
```ts
export function toggleEntry(selection: string[], name: string): string[]
export function setSingle(name: string): string[]
export function selectRange(allNames: string[], anchor: string, target: string): string[]
export function clear(): string[]
```
Pure functions, no Svelte runes; testable.
- [ ] **Step 2** — Walk-through specs in comments for `selectRange`: with `allNames = [a,b,c,d,e]`, anchor `b`, target `d` → `[b,c,d]`; reverse direction `d→b` → `[d,c,b]` (preserve direction so anchor stays last).
- [ ] **Verify:** Manual spec walkthrough table in execution md trail (project rule: spec walkthrough hard gate). No fake verifier file (project RULE 1).

### Task 1.3: Commit

- [ ] `git commit -m "[feat] selection helpers + ExplorerPayload.selection field"`

---

## Phase 2 — Click-based multi-select

### Task 2.1: Selection state in Explorer

**Files:**
- Modify: `frontend/src/lib/components/Explorer.svelte`

- [ ] **Step 1** — Derive `const selection = $derived((tabs.list.find(t => t.id === tabId)?.payload as ExplorerPayload | null)?.selection ?? [])` and `const selectedSet = $derived(new Set(selection))`.
- [ ] **Step 2** — Helper `function updateSelection(next: string[]) { tabs.update(tabId, { payload: { loc, selection: next } }); }`.
- [ ] **Step 3** — Clear selection on `loc` change via existing `$effect(() => { void loc; ... })`.
- [ ] **Verify:** `npm run check` → 0/0.

### Task 2.2: Entry click handlers

**Files:**
- Modify: `Explorer.svelte` (handler), `FileGrid.svelte`, `FileList.svelte` (props + onClick)

- [ ] **Step 1** — Add `onSelect(event: MouseEvent, entry: FolderEntry)` handler in Explorer.svelte. **Modifier precedence (highest → lowest):**
  1. `event.shiftKey` and `selection.length > 0` → `selectRange(sorted.map(e=>e.name), anchor=selection[selection.length-1], target=entry.name)`. (Shift always wins; Ctrl/Cmd ignored when Shift held.)
  2. `event.ctrlKey || event.metaKey` (Shift NOT held) → `toggleEntry(selection, entry.name)`.
  3. No modifier → `setSingle(entry.name)`.
  Call `updateSelection(next)`. Document this precedence in an inline comment.
- [ ] **Step 2** — Wire `onClick` on entry button/row in FileGrid/FileList to call this handler. **Important:** `click` fires before `dblclick`; dblclick still navigates folders, so clicking a folder once selects (correct), double-clicking opens (correct). On files, click selects, dblclick opens viewer.
- [ ] **Step 3** — Pass new prop `selected: Set<string>` (or simpler: `selectedNames: string[]`) to FileGrid/FileList. Each entry checks `selected.has(entry.name)` to apply `bg-bg-elevated ring-1 ring-accent` styling.
- [ ] **Verify:** Playwright: click A → A selected; Ctrl+click B → A+B; Ctrl+click A → only B; Shift+click C with anchor B → B+C; click empty selection (handled by marquee phase).

### Task 2.3: Keyboard shortcuts

**Files:**
- Modify: `Explorer.svelte` (window keydown listener, like mouse-back)

- [ ] **Step 1** — `onKeyDown(event: KeyboardEvent)`:
  - If `event.target` is an editable element, return.
  - If `tabs.activeId !== tabId`, return (only the active explorer responds).
  - `Ctrl/Cmd + A` → `event.preventDefault()`; `updateSelection(sorted.map(e=>e.name))`.
  - `Escape` → `updateSelection([])`.
- [ ] **Step 2** — Add/remove the listener in `onMount`/`onDestroy` next to mouse-back listener.
- [ ] **Verify:** Playwright: focus explorer area, press Ctrl+A → all entries highlighted; Esc → cleared.

### Task 2.4: Commit Phase 2

- [ ] `git commit -m "[feat] multi-select via click/ctrl-click/shift-click/ctrl-a/esc"`

---

## Phase 3 — Marquee (rectangle) selection

### Task 3.1: Marquee state

**Files:**
- Modify: `Explorer.svelte`

- [ ] **Step 1** — Add component-local state:
```ts
let marquee = $state<{ x0: number; y0: number; x1: number; y1: number } | null>(null);
let entryBoxes: Array<{ name: string; rect: DOMRect }> = []; // snapshot
```
- [ ] **Step 2** — On `pointerdown` on the file area's blank space, capture pointer, set `marquee = {x0:event.clientX, y0:event.clientY, x1, y1}`, snapshot entry bounding rects into `entryBoxes`. **Trigger guard:**
  - Match only when `event.target === gridContainer` OR the target's `closest("[data-marquee-canvas]")` returns the container (entries set `data-marquee-canvas="false"` so descendants don't pass).
  - Reject scrollbar clicks: `if (event.offsetX > gridContainer.clientWidth || event.offsetY > gridContainer.clientHeight) return;` (clientWidth excludes scrollbar; offsetX past it means scrollbar was clicked).
  - Reject if `event.target.matches("input, textarea, [contenteditable='true']")`.
- [ ] **Step 3** — On `pointermove` while `marquee != null`, update `marquee.x1/y1`. Compute intersected entries (rect overlap with marquee box) and call `updateSelection(intersected)`. Use `requestAnimationFrame` throttle to one update per frame.

  **Auto-scroll near viewport edge:** when pointer Y is within 32px of the container top/bottom edge, scroll the container by ±8px/frame (capped). After scroll, re-snapshot `entryBoxes` (or apply delta to all stored rects: `rect.y -= scrollDelta`). Without this, marquee snapshot becomes stale relative to the scrolled view, so entries that scrolled into view aren't selectable.
- [ ] **Step 4** — On `pointerup` / `pointercancel`, clear `marquee`. If marquee was created but no entry intersected (pure empty click), `updateSelection([])` (start-fresh behavior).
- [ ] **Step 5** — Cancel if pointer moves < 4px before pointerup → treat as plain "click empty" (clears selection without marquee draw).
- [ ] **Verify:** Playwright: drag rectangle over A and B in grid → both highlighted; release → selection persisted.

### Task 3.2: Marquee overlay element

**Files:**
- Modify: `Explorer.svelte` template

- [ ] **Step 1** — Render the marquee box conditionally as an absolutely-positioned div inside the file-area container:
```svelte
{#if marquee}
  <div
    class="absolute pointer-events-none border border-accent bg-accent/10"
    style="left:{Math.min(marquee.x0,marquee.x1)}px;top:{Math.min(marquee.y0,marquee.y1)}px;width:{Math.abs(marquee.x1-marquee.x0)}px;height:{Math.abs(marquee.y1-marquee.y0)}px;"
  ></div>
{/if}
```
- [ ] **Step 2** — Wrap the FileGrid/FileList in a `relative` container so the marquee overlay positions correctly. Convert client coords to container-local by subtracting container's `getBoundingClientRect()` origin.
- [ ] **Verify:** Playwright: visible rectangle drawn during drag; matches finger/cursor.

### Task 3.3: Commit Phase 3

- [ ] `git commit -m "[feat] marquee drag-selection in file area"`

---

## Phase 4 — Drag-move with multi-selection

### Task 4.1: Extend drag payload

**Files:**
- Modify: `frontend/src/lib/components/Explorer/drag-drop.ts`

- [ ] **Step 1** — Change `DragPayload` to allow either single or batch:
```ts
export interface DragPayload {
  items: Array<{ name: string; isFolder: boolean }>;
  sourceLoc: string[];
}
```
- [ ] **Step 2** — Update `buildPayload(loc, entries: Array<{name, isFolder}>)` to accept an array. `readPayload` returns the same shape.
- [ ] **Step 3** — Backward compatibility: in `readPayload`, if parsed JSON has top-level `name` (legacy single-entry shape), wrap into `{items:[{name,isFolder}], sourceLoc}`. Document: legacy payload tolerated for one minor version.
- [ ] **Step 4** — Add `performMoveBatch(srcLoc, items, targetLoc)` calling `performMove` per item in a loop, returning `{moved: number, failed: number}`. Surface a single summary notification rather than N toasts.
- [ ] **Verify:** TS build clean; unit walkthrough table: 2 items, target folder X, performMove called twice; refresh runs once after batch.

### Task 4.2: Wire batch drag in FileGrid/FileList

**Files:**
- Modify: `FileGrid.svelte`, `FileList.svelte` (`onDragStart`); `Explorer.svelte` (`dragPayload` builder + drop handler)

- [ ] **Step 1** — In `dragPayload(entry)` (Explorer.svelte), build:
  - If `selectedSet.has(entry.name) && selection.length > 1`: build payload with `items = sorted.filter(e => selectedSet.has(e.name)).map(e => ({name:e.name, isFolder:e.isFolder}))`.
  - Else: build single-item payload (also as `items: [{name:entry.name, isFolder:entry.isFolder}]`).
- [ ] **Step 2** — Update `onDropOnFolder` / `onDropOnLoc` in Explorer.svelte to read the batch payload and call `performMoveBatch` then refresh.
- [ ] **Step 3** — Drag visual: only the dragged entry shows the standard browser drag image (no JS drag-image polyfill). Document this limitation; selected-count badge can come later.
- [ ] **Verify:** Playwright: select A+B+C (Ctrl+click), drag A onto folder D → A+B+C all move; root entry count decreases by 3, D contains 3.

### Task 4.3: Multi-entry delete

**Files:**
- Modify: `Explorer.svelte`

- [ ] **Step 1** — When Delete key pressed (extend keydown handler from Task 2.3) and `selection.length > 0`, prompt `confirm("Delete N entries?")` and call `deleteEntry(loc, syntheticEntry)` per name in a Promise chain; refresh once at end.
- [ ] **Step 2** — Surface summary toast.
- [ ] **Verify:** Playwright: select 3 entries → press Delete → confirm → 3 deletions; folder count drops by 3.

### Task 4.4: Commit Phase 4

- [ ] `git commit -m "[feat] multi-select drag-move + multi-delete"`

---

## Phase 5 — Mobile selection (touch)

Coordinate with mobile-ui plan: marquee requires precise pointer drags incompatible with touch (touch is for scrolling). Use long-press to enter "selection mode" + tap-to-add.

### Task 5.1: Long-press → selection mode

**Files:**
- Modify: `Explorer.svelte`, `FileGrid.svelte`, `FileList.svelte`

- [ ] **Step 1** — Reuse the `long-press` action from mobile-ui plan (Task 7.1). On `longpress`, instead of opening context menu when no selection exists, enter "selection mode" with that entry as initial selection.
- [ ] **Step 2** — In selection mode (`selection.length > 0` on mobile), each entry tap toggles inclusion (instead of opening). Show a top-bar with `N selected` + Cancel + Delete + Move buttons (Move is "→ folder picker" — out of scope; show as disabled with title).
- [ ] **Step 3** — Long-press menu remains accessible via a different gesture: after entering selection mode, long-press → context menu for that single entry (not selection).
  - Simpler alt: in selection mode, tap = toggle; long-press = no-op. Right-click menu requires desktop. Accept the simpler alt; document.
- [ ] **Verify:** Playwright mobile: long-press → mode entered; tap others → added; Cancel → exits mode.

### Task 5.2: Commit Phase 5

- [ ] `git commit -m "[feat] mobile selection mode via long-press and tap-to-add"`

---

## Phase 6 — `/code-review` + Playwright e2e + merge

### Task 6.1: code-review

- [ ] Run on every changed file. Auto-fix Critical.

### Task 6.2: Playwright walkthrough

| # | Mode | Action | Expected |
|---|---|---|---|
| 1 | Desktop | Click file A | A selected (ring-accent + bg) |
| 2 | Desktop | Ctrl+click B | A + B selected |
| 3 | Desktop | Ctrl+click A | only B selected |
| 4 | Desktop | Shift+click D (anchor B, sorted entries are A B C D E) | B + C + D selected |
| 5 | Desktop | Ctrl+A | all entries selected |
| 6 | Desktop | Esc | cleared |
| 7 | Desktop | Drag from blank to over A and B | marquee visible; A + B highlighted |
| 8 | Desktop | Click somewhere blank (< 4px move) | selection cleared, no marquee drawn |
| 9 | Desktop | Select A+B+C, drag A onto folder D | all 3 moved; toast "Moved 3 items" |
| 10 | Desktop | Select 3, press Delete, confirm | 3 deleted in one batch |
| 11 | Desktop | Navigate to subfolder | selection cleared |
| 12 | Desktop | Multi-Explorer tab: select in tab 1 | tab 2's selection untouched (per-tab payload) |
| 13 | Mobile (after mobile-ui plan) | Long-press file | enters selection mode with that file |
| 14 | Mobile | Tap 2 more files | added to selection; top bar shows "3 selected" |
| 15 | Mobile | Tap Cancel | mode exits, selection cleared |
| 16 | Desktop | Drag marquee over text input in toolbar | marquee does not start (event filter on input) |

### Task 6.3: Pre-merge + merge

- [ ] RULE 6 checklist; merge `feat/multi-select` → main; push origin.

---

## Risks & open questions

| Risk | Mitigation |
|---|---|
| Marquee tries to start on every blank-area mousedown including on the title text/scrollbar | Three-layer guard in Task 3.1 step 2: target-match via `data-marquee-canvas` attr, `offsetX/Y > clientWidth/clientHeight` check rejects scrollbar clicks, editable-target reject. Verified in walkthrough row 16-17. |
| Marquee selection misses entries that scroll into view mid-drag | Auto-scroll handler (Task 3.1 Step 3) scrolls container at edges and re-snapshots entry rects per frame. |
| Multi-modifier combo (Shift+Ctrl+click) undefined | Precedence locked in Task 2.2 Step 1: Shift > Ctrl/Cmd > no-modifier. Documented inline. |
| Global window keydown listener race across multiple Explorer instances and viewer tabs (Monaco's keybindings) | Each Explorer's onKeyDown short-circuits when `tabs.activeId !== tabId`. Monaco's textarea matches the editable-target guard. Verified in walkthrough row 17 (Ctrl+A inside Monaco does NOT select Explorer entries). |
| Accessibility: selection count not announced to screen readers | Visual-only "N selected" badge for v1. Acknowledged deferral; future plan adds `aria-live="polite"` region. Listed as known limitation in spec walkthrough row 18. |
| Selection persists across folder navigation, surprising user | Cleared explicitly in `$effect(() => { void loc; })` (Task 1 of Phase 2). |
| Per-tab payload bloat with huge selection (1000+) | Selection is array of names (strings); 1000 entries × 64 chars avg = ~64 KB per tab payload. Acceptable. Document. |
| Cross-tab drag-move (drag from tab 1 to tab 2's breadcrumb) | Tab 2 isn't visible while tab 1 active; cross-tab drag UX is out of scope. Document. |
| Browser drag image only shows the dragged entry, not the selected set | Standard HTML5 DnD limitation. Document; future plan can use `setDragImage` with a custom canvas badge. |
| Marquee rect snapshot stale if entries reflow mid-drag | Entries don't reflow during marquee (no async fetch). Acceptable. |
| Race between dblclick (open) and click (select) | Browser fires `click` then `dblclick`; selection-on-click is idempotent — dblclick still opens. Test row 1 + 11 confirms. |
| Touch-mode long-press conflict with mobile context-menu plan | Coordinated in Task 5.1 step 3 — accept simpler alt (tap=toggle, no context menu in selection mode). |
| Ctrl+A captured inside a viewer (Monaco) | Phase 2 Task 2.3 step 1 guards: skip when target is editable element. Monaco's editor textarea matches `textarea`. |
| Backward-compat for single-entry drag payload | `readPayload` transparently wraps legacy shape into batch shape (Task 4.1 step 3); document removal timeline (next minor release). |

---

## Rollback strategy

```bash
cd C:/Data/Git/ANXI/nas && git revert -m 1 <merge-sha> && git push origin main
```

ExplorerPayload's new `selection` field is optional — safe to revert frontend without backend impact (no backend change). Per-phase commits enable selective revert. Drag-payload shape change in Phase 4 is the only forward-compat concern: if revert leaves frontend on single-shape but a sibling tab still has batch payload in memory, the legacy-wrap path in Task 4.1 step 3 handles it for the running session.

---

## 5-pass self-review (post-reviewer revision)

| Pass | Concern | Action |
|---|---|---|
| 1. Structure | 6 phases (type → click-multi → marquee → drag-batch → mobile → merge). Each task has Verify. Dependency-ordered. | ✅ |
| 2. Spec correctness | Reviewer caught: `ExplorerPayload` is at `types.ts:33-35` (not :22-24, which is `AuthUser`). Fixed. `tabs.update` at tabs.svelte.ts:77 confirmed. drag-drop.ts module exists (created by feat/ux-overhaul). | ✅ (fixed) |
| 3. Risk | Reviewer caught: auto-scroll near edge, scrollbar click trigger, multi-Explorer keyboard ordering, a11y aria-live, multi-modifier precedence. All five added with concrete mitigations (auto-scroll handler, three-layer guard, active-tab short-circuit, deferred aria-live with documentation, precedence locked Shift > Ctrl > none). | ✅ (fixed) |
| 4. Consistency | Reuses `NAS_ENTRY_MIME`, performMove pattern; Svelte 5 runes; Tailwind tokens (`ring-accent`, `bg-bg-elevated`, `bg-accent/10`); long-press helper shared with mobile plan; commit format `[feat] subject` | ✅ |
| 5. Completeness | Acceptance: marquee ✅, Ctrl-click ✅, Shift-range ✅, Ctrl+A ✅, Esc ✅, multi-move ✅, multi-delete ✅, mobile mode ✅. Walkthrough expanded to 21 rows (added scrollbar guard, Monaco Ctrl+A, auto-scroll, Shift+Ctrl precedence, a11y deferral). Rollback concrete. | ✅ (expanded) |

**Cross-plan note:** Phase 5 (mobile selection) depends on the `long-press` Svelte action from the mobile-ui plan (Task 7.1). If mobile-ui hasn't merged, skip Phase 5 here and defer it to a follow-up that lands after mobile-ui.
