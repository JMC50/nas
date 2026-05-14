# Mobile UI Support — Implementation Plan

> **For agentic workers:** Use `superpowers:subagent-driven-development` or `superpowers:executing-plans`. Steps use `- [ ]` syntax.

**Goal:** Make the entire NAS UI usable on viewports `< 768px` (Tailwind `md` breakpoint) — sidebar becomes a slide-in drawer, toolbars condense, tab bar scrolls horizontally, side-panels become bottom sheets, touch replaces hover/right-click semantics where appropriate.

**Architecture:**
- **Reuse the existing `ui.isMobile`** $derived flag in [ui.svelte.ts:31](frontend/src/lib/store/ui.svelte.ts:31) (already wired to `viewportWidth` + resize listener at lines 36-41) — DO NOT create a parallel `viewport.svelte.ts` store. Extend `ui` store with `drawerOpen` and helper toggles only.
- Layout grid in `(app)/+layout.svelte` switches from `grid-rows-[48px_1fr_28px] grid-cols-[auto_1fr]` to a stacked layout on mobile: `grid-rows-[48px_1fr_28px] grid-cols-[1fr]` with the sidebar removed from the DOM grid and overlaid as a drawer.
- Tailwind 4 `md:` / `sm:` utility prefixes used throughout — single source of truth for responsive behavior, no JS-driven CSS string concatenation.
- Touch interactions: long-press (500ms) replaces right-click for context menus; double-tap replaces double-click for "open"; native HTML5 drag-and-drop stays mouse-only (long-press could be added in a future plan).
- iOS-specific: use `100dvh` over `100vh` to avoid Safari URL-bar bug; bottom sheets and the StatusBar use `env(safe-area-inset-bottom)` padding; `<html>` gets `overflow: hidden` while `ui.drawerOpen` to prevent body scroll bleed; `visualViewport.height` API tracked when QuickOpen has focus so the input stays above the virtual keyboard.

**Tech Stack:** Svelte 5 runes, Tailwind 4, lucide-svelte. No new dependencies.

---

## Context

- [frontend/src/routes/(app)/+layout.svelte:31](frontend/src/routes/(app)/+layout.svelte:31) — top-level grid layout.
- [frontend/src/lib/components/Shell/VerticalNav.svelte:48-83](frontend/src/lib/components/Shell/VerticalNav.svelte:48) — sidebar with `w-12 / w-[200px]` toggle via `ui.sidebarCollapsed`.
- [frontend/src/lib/components/Shell/Header.svelte:37-79](frontend/src/lib/components/Shell/Header.svelte:37) — top bar with logo, quick-open search, theme toggle, account, logout.
- [frontend/src/lib/components/Shell/StatusBar.svelte](frontend/src/lib/components/Shell/StatusBar.svelte) — footer bar (uploads progress + theme + version).
- [frontend/src/lib/components/Explorer/Toolbar.svelte](frontend/src/lib/components/Explorer/Toolbar.svelte) — Up / Upload / Upload folder / New folder / search / Grid / List.
- [frontend/src/lib/components/Tabs/TabBar.svelte](frontend/src/lib/components/Tabs/TabBar.svelte) — horizontal tab strip with reorder DnD + new-tab button.
- [frontend/src/lib/components/Uploads/UploadPanel.svelte:42-44](frontend/src/lib/components/Uploads/UploadPanel.svelte:42) — `w-[360px]` fixed right panel; same pattern reused by Inspector (file-details plan).
- [frontend/src/lib/store/ui.svelte.ts](frontend/src/lib/store/ui.svelte.ts) — already exposes `sidebarCollapsed` (line 15), `quickOpenVisible` (line 16), `uploadsPanelOpen` (line 17), `viewportWidth`/`viewportHeight` with resize listener (lines 18-19, 36-41), `breakpoint` $derived (lines 21-29), `isMobile` $derived (line 31), `BREAKPOINTS` constant (lines 7-12). API: `openQuickOpen()/closeQuickOpen()`, `toggleSidebar()`, `toggleUploadsPanel()/closeUploadsPanel()`.
- [frontend/src/lib/types.ts:97](frontend/src/lib/types.ts:97) — `Breakpoint = "sm" | "md" | "lg" | "xl"` already declared.
- Tailwind 4 default breakpoints: `sm:640px`, `md:768px`, `lg:1024px`, `xl:1280px`.

---

## Phase 1 — Extend ui store with `drawerOpen` (no parallel viewport store)

`ui.isMobile`, `ui.breakpoint`, `ui.viewportWidth/Height`, and resize listener already exist in `ui.svelte.ts`. Reuse them directly.

### Task 1.1: Add `drawerOpen` state to UIStore

**Files:**
- Modify: `frontend/src/lib/store/ui.svelte.ts`

- [ ] **Step 1** — Add field `drawerOpen = $state<boolean>(false);` near other state.
- [ ] **Step 2** — Add `openDrawer()`, `closeDrawer()`, `toggleDrawer()` methods following the existing `openQuickOpen`/`closeQuickOpen` pattern (lines 52-58).
- [ ] **Step 3** — Add an `$effect`-like browser side effect in the constructor: when `drawerOpen` becomes true, set `document.documentElement.style.overflow = "hidden"`; when false, restore to `""`. (Implement as a method `setDrawerOpen(value: boolean)` that does both store update and DOM mutation atomically; use it from openDrawer/closeDrawer/toggleDrawer.)
- [ ] **Step 4** — Auto-close on viewport upsize: if `viewportWidth >= BREAKPOINTS.md`, force `drawerOpen = false` and restore body scroll. Hook into the existing resize listener at lines 38-41.
- [ ] **Verify:** `npm run check` → 0/0; Playwright: opening drawer locks body scroll; resize past md threshold auto-closes.

### Task 1.2: Add 100dvh + safe-area helpers (CSS only)

**Files:**
- Modify: `frontend/src/app.css`

- [ ] **Step 1** — Add a global rule overriding `h-screen` for the app root: use `100dvh` with `100vh` fallback. Simplest: a utility class `.h-app { height: 100vh; height: 100dvh; }` applied to the layout grid root. (Or override Tailwind's `h-screen` only inside `(app)` route via a route-scoped class.)
- [ ] **Step 2** — Document safe-area utility: components that pin to bottom (StatusBar, bottom sheets) add `pb-[env(safe-area-inset-bottom)]`. No global CSS change needed — applied per component in later phases.
- [ ] **Verify:** Manual: open on emulated iPhone 12 (390×844 with home indicator) → no content clipped behind bar.

### Task 1.3: Commit Phase 1

- [ ] `git commit -m "[feat] ui store drawer state + body-scroll-lock + 100dvh"`

---

## Phase 2 — Layout: drawer sidebar on mobile

### Task 2.1: Layout grid responsive

**Files:**
- Modify: `frontend/src/routes/(app)/+layout.svelte`

- [ ] **Step 1** — Replace `h-screen` with `.h-app` utility from Task 1.2 (100dvh). Replace grid template with Tailwind responsive:
  - Base (mobile): `grid grid-rows-[48px_1fr_28px] grid-cols-[1fr]`
  - `md:`: `grid-cols-[auto_1fr]`
- [ ] **Step 2** — Wrap `<VerticalNav />` in `<div class="hidden md:contents">...</div>` so it disappears from grid flow on mobile. The mobile drawer is rendered separately (Task 2.2).
- [ ] **Step 3** — `<main>` element: keep `row-start-2 col-start-1 md:col-start-2` — on mobile spans full width.
- [ ] **Step 4** — StatusBar (h-7) gets `pb-[env(safe-area-inset-bottom)]` on mobile only (`md:pb-0`) to clear iOS home-indicator.
- [ ] **Verify:** Playwright at 375×667 (iPhone SE viewport): sidebar invisible; main fills width. At 1280×800: sidebar visible. iPhone 12 (390×844): StatusBar bottom padded.

### Task 2.2: Mobile drawer

**Files:**
- Create: `frontend/src/lib/components/Shell/MobileDrawer.svelte`
- Modify: `frontend/src/lib/components/Shell/VerticalNav.svelte` (extract NAV_ITEMS)

(Drawer state already added to ui store in Task 1.1.)

- [ ] **Step 1** — Extract `NAV_ITEMS` constant + `activate` helper to `frontend/src/lib/components/Shell/nav-items.ts`; both `VerticalNav.svelte` and `MobileDrawer.svelte` import it.
- [ ] **Step 2** — `MobileDrawer.svelte`: when `ui.drawerOpen`, render a `fixed inset-0 z-50` overlay (semi-transparent backdrop, click → `ui.closeDrawer()`) + a left-anchored `w-64 h-full bg-bg-surface border-r border-border-default` panel containing the NAV_ITEMS list. Use `pt-[env(safe-area-inset-top)]` so notches don't overlap the topmost item.
- [ ] **Step 3** — Drawer entry click calls `activate(item)` then `ui.closeDrawer()`.
- [ ] **Step 4** — Esc keydown closes drawer (add handler in MobileDrawer onMount).
- [ ] **Step 5** — Mount `<MobileDrawer />` once in `(app)/+layout.svelte` near `<DragDropOverlay />`.
- [ ] **Verify:** Tap hamburger → drawer slides in; tap backdrop → closes (body scroll restored per Task 1.1 Step 3); Esc → closes; resize past md → auto-closes (per Task 1.1 Step 4).

### Task 2.3: Hamburger button in Header

**Files:**
- Modify: `frontend/src/lib/components/Shell/Header.svelte`

- [ ] **Step 1** — Add hamburger button (lucide `menu` icon) as the leftmost element, visible only on mobile (`md:hidden`). Click → `ui.toggleDrawer()`.
- [ ] **Step 2** — Logo + NAS text label: hide on mobile (`hidden md:flex`) to save space, OR move next to hamburger and shrink. Decision: hide on mobile.
- [ ] **Verify:** Playwright mobile: hamburger visible, NAS text hidden; desktop: logo visible, no hamburger.

### Task 2.4: Commit Phase 2

- [ ] `git commit -m "[feat] mobile drawer sidebar + hamburger menu in header"`

---

## Phase 3 — Mobile-friendly Header

### Task 3.1: Quick-open button condense

**Files:**
- Modify: `frontend/src/lib/components/Shell/Header.svelte`

- [ ] **Step 1** — On mobile, hide the "Search files… (Ctrl+P)" text and show only the search icon `w-8 h-8` button.
- [ ] **Step 2** — Account button: hide username text on mobile; show only user icon.
- [ ] **Verify:** Playwright mobile: header has hamburger · search-icon-only · theme · user-icon · logout — fits in 375px width.

### Task 3.2: Commit

- [ ] `git commit -m "[feat] header condense on mobile (icon-only buttons)"`

---

## Phase 4 — Mobile-friendly Toolbar

### Task 4.1: Toolbar condense + overflow menu

**Files:**
- Modify: `frontend/src/lib/components/Explorer/Toolbar.svelte`
- Create: `frontend/src/lib/components/Explorer/ToolbarOverflow.svelte`

- [ ] **Step 1** — On mobile, primary visible buttons: Up arrow, Upload (icon only), View toggle (Grid/List). The rest (Upload folder, New folder, search) collapse into an "..." overflow menu (lucide `more-horizontal`).
- [ ] **Step 2** — Search input: on mobile, replace inline input with a search icon button that toggles a full-width search bar below the toolbar.
- [ ] **Step 3** — `ToolbarOverflow.svelte` is a dropdown menu (use existing ContextMenu-like positioning) with the collapsed actions.
- [ ] **Verify:** Playwright mobile: Toolbar fits in 375px without horizontal scroll; tap "..." → overflow menu opens with Upload folder + New folder + Search.

### Task 4.2: Commit

- [ ] `git commit -m "[feat] toolbar overflow menu on mobile"`

---

## Phase 5 — Tab bar horizontal scroll + close

### Task 5.1: TabBar mobile scroll

**Files:**
- Modify: `frontend/src/lib/components/Tabs/TabBar.svelte`

- [ ] **Step 1** — On mobile, tab width: shrink `min-w-[120px] max-w-[200px]` to `min-w-[100px] max-w-[140px]`. `overflow-x-auto` already present — verify scrolls smoothly on touch (`-webkit-overflow-scrolling: touch` if needed but Tailwind handles via native scroll).
- [ ] **Step 2** — Close button visibility on mobile: always visible (no `opacity-0 group-hover:opacity-100` because hover doesn't exist on touch). Wrap visibility class in `md:opacity-0 md:group-hover:opacity-100` to keep desktop behavior.
- [ ] **Step 3** — Drag-to-reorder on TabBar: disable on mobile (`draggable={!ui.isMobile}`). Alt mobile reorder is out of scope.
- [ ] **Verify:** Playwright mobile: tab close X always visible; horizontal scroll past 4 tabs works.

### Task 5.2: Commit

- [ ] `git commit -m "[feat] tab bar mobile: shrink width, always-visible close, no drag reorder"`

---

## Phase 6 — Bottom sheets for side panels

### Task 6.1: UploadPanel as bottom sheet on mobile

**Files:**
- Modify: `frontend/src/lib/components/Uploads/UploadPanel.svelte`

- [ ] **Step 1** — Container class becomes conditional:
  - Mobile: `fixed inset-x-0 bottom-0 top-auto max-h-[70dvh] w-full rounded-t-lg border-t border-border-default pb-[env(safe-area-inset-bottom)]`
  - Desktop (`md:`): preserve `top-12 right-0 bottom-7 w-[360px] border-l`
- [ ] **Step 2** — Add a drag-handle bar (`<div class="md:hidden w-12 h-1 bg-fg-muted/40 rounded-full mx-auto mt-2 mb-1">`) to signal swipe-to-close affordance (visual only; no JS swipe handler yet — out of scope).
- [ ] **Verify:** Playwright mobile: uploads panel opens as bottom sheet, height capped at 70dvh; no clipping under iOS home-indicator; desktop: still right side panel.

### Task 6.2: Inspector as bottom sheet (if file-details plan merged)

**Files:**
- Modify: `frontend/src/lib/components/Explorer/Inspector.svelte` (after file-details plan)

- [ ] **Step 1** — Same responsive container swap as Task 6.1.
- [ ] **Verify:** Playwright mobile: tap file → inspector slides up from bottom.

(Skip Task 6.2 entirely if file-details plan hasn't merged; surface a TODO row in the spec table.)

### Task 6.3: QuickOpen as fullscreen on mobile

**Files:**
- Modify: `frontend/src/lib/components/QuickOpen.svelte`

- [ ] **Step 1** — On mobile (`ui.isMobile`), the modal becomes `fixed inset-0 z-50 bg-bg-base` (full screen) instead of centered card. Trigger via the existing `ui.openQuickOpen()` / `ui.closeQuickOpen()` API; the visibility flag is `ui.quickOpenVisible` (not `quickOpenOpen`).
- [ ] **Step 2** — Track virtual-keyboard via `visualViewport.height`: when keyboard opens, set `style="height: ${window.visualViewport.height}px"` on the modal container so the search input stays above the keyboard. Attach `visualViewport.addEventListener("resize", ...)` on mount; clean up on destroy.
- [ ] **Verify:** Playwright mobile: tap search icon → fullscreen quick open with input focused; keyboard up → modal shrinks to visible area, input stays visible.

### Task 6.4: Commit Phase 6

- [ ] `git commit -m "[feat] side panels become bottom sheets / fullscreen on mobile"`

---

## Phase 7 — Touch interactions

### Task 7.1: Long-press → context menu

**Files:**
- Modify: `frontend/src/lib/components/Explorer/FileGrid.svelte`, `FileList.svelte`

- [ ] **Step 1** — Create a `useLongPress` helper (small Svelte action `frontend/src/lib/actions/long-press.ts`):
  - On `pointerdown`, start a 500ms timer with the pointer position.
  - On `pointerup` / `pointercancel` / `pointermove` (move > 8px), cancel.
  - On timer fire, dispatch `longpress` custom event with `{clientX, clientY}`.
- [ ] **Step 2** — Apply `use:longPress` to entry button/row; on event, call existing `onMenu(syntheticMouseEvent, entry)`.
- [ ] **Step 3** — Keep `oncontextmenu` on desktop. Both paths converge on `onMenu`.
- [ ] **Verify:** Playwright mobile: long-press file → context menu opens at finger position.

### Task 7.2: Double-tap → open

Touch already fires `dblclick` after two quick taps (browser behavior). No code change needed. Verify in walkthrough.

### Task 7.3: Disable HTML5 drag on mobile

**Files:**
- Modify: `frontend/src/lib/components/Explorer/FileGrid.svelte`, `FileList.svelte`, `Breadcrumb.svelte`, `Toolbar.svelte`

- [ ] **Step 1** — On entry elements, `draggable={!ui.isMobile}`. Mobile users cannot drag-move; they use context menu → Move (future plan).
- [ ] **Verify:** Playwright mobile: dragging a file does nothing (page may scroll); desktop drag still works.

### Task 7.4: Commit Phase 7

- [ ] `git commit -m "[feat] long-press context menu + drag-disable on mobile"`

---

## Phase 8 — Viewer responsiveness

### Task 8.1: MediaViewer / PdfViewer / Monaco fit mobile

**Files:**
- Modify: `frontend/src/lib/components/Viewers/MediaViewer.svelte`, `VideoControls.svelte`, `VideoPlayer.svelte`, `AudioPlayer.svelte`, `PdfViewer.svelte`, `MonacoViewer.svelte`, `OfficeViewer.svelte`, `ImageViewer.svelte`

- [ ] **Step 1** — Each viewer: verify root container uses `w-full h-full` (no fixed widths). Audit and fix any hardcoded `w-[Npx]` larger than `375px`.
- [ ] **Step 2** — VideoControls: hide secondary controls (volume slider, settings) behind a "..." button on mobile; primary controls (play, scrubber, fullscreen) always visible.
- [ ] **Step 3** — PdfViewer pager toolbar: condense same way (Prev / Page input / Next stays; zoom controls overflow).
- [ ] **Step 4** — Monaco: ensure word-wrap on for narrow viewports (`wordWrap: ui.isMobile ? "on" : "off"` in editor options).
- [ ] **Verify:** Playwright mobile: each viewer fills viewport without horizontal scroll; controls usable with thumb.

### Task 8.2: Commit Phase 8

- [ ] `git commit -m "[feat] viewers responsive on mobile (controls condense, word-wrap)"`

---

## Phase 9 — `/code-review` + Playwright e2e + merge

### Task 9.1: code-review

- [ ] Run on every changed file. Auto-fix Critical.

### Task 9.2: Playwright walkthrough (RULE 1 hard gate)

Emulate iPhone SE (375×667), iPad (768×1024 — `isTablet`), desktop (1280×800):

| # | Viewport | Action | Expected |
|---|---|---|---|
| 1 | Mobile | Open app | Hamburger visible; sidebar hidden; NAS text hidden |
| 2 | Mobile | Tap hamburger | Drawer slides in from left; nav items shown |
| 3 | Mobile | Tap nav "Activity" | Drawer closes, Activity tab active |
| 4 | Mobile | Tap "..." in Toolbar | Overflow menu shows Upload folder + New folder + Search |
| 5 | Mobile | Open Uploads panel | Bottom sheet appears, max-h-[70vh] |
| 6 | Mobile | Tap Ctrl+P icon | QuickOpen fills screen |
| 7 | Mobile | Long-press file (500ms) | Context menu opens at finger position |
| 8 | Mobile | Try to drag file | No drag initiates; page may scroll |
| 9 | Mobile | Open Video viewer | Controls fit; secondary controls behind "..." |
| 10 | Mobile | Open Monaco viewer | Code word-wraps |
| 11 | Tablet | Open app | Sidebar visible (≥md); same as desktop |
| 12 | Desktop | Open app | Original behavior; right-click context menu, drag, hover-revealed close |
| 13 | All | Toggle theme dark/light | Both themes render correctly across viewports |
| 14 | Mobile | Drawer open + tap backdrop | Drawer closes; body scroll restored |
| 15 | Mobile (iPhone 12 emulation, 390×844) | StatusBar visible at bottom | Not clipped by home indicator (safe-area padding present) |
| 16 | Mobile | Open QuickOpen, focus search → soft keyboard appears | Modal height adjusts via visualViewport; search input visible |
| 17 | Mobile | Resize browser past 768px while drawer open | Drawer auto-closes; sidebar appears in DOM grid |
| 18 | Mobile | Drawer open + press Esc | Drawer closes |

### Task 9.3: Pre-merge + merge

- [ ] RULE 6 checklist; merge `feat/mobile-ui` → main; push origin.

---

## Risks & open questions

| Risk | Mitigation |
|---|---|
| iOS Safari 100vh URL-bar bug → bottom 28px StatusBar can be hidden | Phase 1 Task 1.2 replaces `h-screen` with `.h-app { height: 100vh; height: 100dvh; }`. Verified at iPhone emulation in walkthrough row 14. |
| iOS home-indicator clip on bottom sheets / StatusBar | `pb-[env(safe-area-inset-bottom)]` applied per component (Task 2.1 Step 4, Task 6.1 Step 1). |
| Body scroll bleed under open drawer | `ui.openDrawer()` sets `documentElement.style.overflow = "hidden"`; closeDrawer restores. (Task 1.1 Step 3.) |
| Virtual keyboard covering QuickOpen input | `visualViewport.height` tracking in Task 6.3 Step 2 sets explicit modal height so input stays visible. |
| SSR mismatch (server thinks `isMobile=false`) | `ui.viewportWidth` defaults to 0 (line 18) so `isMobile=true` on server. First client paint runs resize listener (lines 36-41) and re-derives. Layout is CSS-utility-driven (Tailwind `md:`) so the only JS-driven branches are event handlers, not initial markup. |
| Long-press 500ms triggers on accidental scroll | Cancel on `pointermove` > 8px (Task 7.1 step 1). Standard mobile UX threshold. |
| Tailwind 4 `md:` arbitrary value (`md:w-[360px]`) generation | Already used in UploadPanel; Tailwind 4 JIT handles it. |
| Drag-reorder loss on mobile TabBar | Documented; out of scope. Future: swipe-to-reorder. |
| ContextMenu position off-screen near right/bottom edge on small viewport | ContextMenu uses fixed `left:Xpx top:Ypx`; verify edge-clip in walkthrough row 7. If clipped, follow-up plan adjusts position. |
| Drawer focus trap / focus loss | Drawer adds Esc key handler (Task 2.2 Step 4); focus trap is out of scope and accepted as v1 limitation, surfaced as a row in spec walkthrough. |
| StatusBar `h-7` overlapping bottom sheet `max-h-[70dvh]` | 70dvh allows 30dvh of UI above; StatusBar at the bottom of the grid stays below the sheet because sheet uses `fixed bottom-0` (z-40) over the grid; verify in walkthrough row 5. |
| Viewer libraries (Monaco) viewport reaction | Monaco respects parent size; resize observer already wires re-layout. Word-wrap option pass-through verified manually. |

---

## Rollback strategy

```bash
cd C:/Data/Git/ANXI/nas && git revert -m 1 <merge-sha> && git push origin main
```

Changes are layered: viewport store + Tailwind utilities + responsive component branches. Reverting the merge removes all responsive behavior; desktop UX is unchanged at all times (mobile branches are additive `md:` overrides, not replacements). Per-phase commits enable selective revert (`git revert <commit-sha>` per phase commit).

---

## 5-pass self-review (post-reviewer revision)

| Pass | Concern | Action |
|---|---|---|
| 1. Structure | 9 phases dependency-ordered (ui-store-extend → layout → drawer → header → toolbar → tabs → panels → touch → viewers → merge). Every task has Verify. | ✅ |
| 2. Spec correctness | Reviewer caught: parallel `viewport.svelte.ts` collision with existing `ui.isMobile` (resolved — Phase 1 now extends ui store), `quickOpenVisible` (not `quickOpenOpen`), line cites `types.ts:97` (not :93) and `+layout.svelte:31`. All paths re-verified against actual files. | ✅ (fixed) |
| 3. Risk | Reviewer caught: iOS 100vh, safe-area-inset, body scroll lock, visualViewport keyboard. All added to risks table with concrete mitigations (100dvh, env() padding, overflow:hidden, visualViewport listener). | ✅ (fixed) |
| 4. Consistency | Reuses existing Gruvbox tokens, lucide icons, Svelte 5 runes, monospace small text. No new deps. Commit format `[type] subject`. | ✅ |
| 5. Completeness | Acceptance: entire UI usable on mobile ✅. 18-row walkthrough covers all phases including iOS-specific rows 15-18 (safe-area, visualViewport, auto-close, Esc). Rollback concrete. | ✅ (expanded) |
