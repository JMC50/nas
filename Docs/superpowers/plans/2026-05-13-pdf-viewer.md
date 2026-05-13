# PDF.js-Based PDF Viewer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current `<iframe src="/server/download">` PDF viewer with a **PDF.js-based** canvas renderer that gives consistent cross-browser display, explicit page navigation, zoom controls, and a foundation for future search / text-layer features. The iframe approach falls back to the browser's built-in renderer which differs significantly between Chrome, Firefox, Safari, and mobile — PDF.js renders uniformly.

**Architecture:** Lazy-load `pdfjs-dist` on viewer mount (its main bundle is ~1.5MB — we don't want to inflate the app's initial JS chunk). Render each page to its own `<canvas>` element inside a scrollable container. Use `IntersectionObserver` so we only render pages currently in viewport (renders are CPU-expensive). A top toolbar exposes page navigation (`prev / page X of Y / next`), zoom (in / out / fit-width / 100%), and is sticky over the scroll body.

**Tech Stack:** Svelte 5 runes, Tailwind 4 (project theme), `pdfjs-dist` v4+ (Mozilla's PDF.js packaged for browsers), Vite's `?url` import for the PDF.js worker, lucide-svelte icons.

**Visual Design:**
- Toolbar: 40px tall, `bg-bg-overlay` with `border-b border-border-default`, sticky at top
- Toolbar layout (left → right): filename, separator, prev button, page input `[ 3 ] / 27`, next button, separator, zoom out, zoom percent display, zoom in, fit-width toggle, separator, download button
- Pages: each page on a `<canvas>` with `bg-bg-elevated` shadow card, vertical gap `gap-4`, centered with `mx-auto`
- Scroll container: `bg-bg-base`, fills below toolbar, smooth scrolling
- Buttons: 32px, same hover style as other viewers (`hover:bg-bg-hover hover:text-fg-accent`)

**Testing Approach:** No frontend test framework — spec-driven manual verification. Gates: `npm run check` per task, `/code-review` per implementation file, manual visual walkthrough at the end.

**Out of scope (separate / future):**
- Text selection / copy from PDFs (would need PDF.js text layer) — v2
- Search within PDF — v2 (needs text layer + UI)
- PDF annotations / forms — v3
- Print custom rendering (use browser's native print) — N/A
- Mobile gesture pinch-to-zoom — keep the +/- buttons; gesture v2
- Thumbnail sidebar — v2

---

## File Structure

| File | Status | Responsibility | Target Lines |
|---|---|---|---|
| `frontend/src/lib/components/Viewers/PdfViewer.svelte` | Modify (full rewrite) | Toolbar + page canvases + lazy render via IntersectionObserver | ≤280 |
| `frontend/package.json` | Modify | Add `pdfjs-dist` dependency | +1 |

**Why single file:** PDF viewer logic is cohesive — toolbar state, current page, zoom factor, document instance all interact tightly. Splitting into a separate `PdfToolbar.svelte` adds prop-drilling overhead with no real reuse. If the file exceeds 280 lines during implementation, extract `PdfToolbar.svelte` as a sub-component.

**No backend changes.** Existing `/server/download` already serves the raw PDF with HTTP 206 Range support (PDF.js can use ranges for big files via its `rangeChunkSize` option — we'll opt into this).

---

## Task 0: Branch setup

- [ ] **Step 1: Create branch**

```bash
git checkout main && git pull
git checkout -b feat/pdf-viewer
```

---

## Task 1: Install pdfjs-dist

**Files:**
- Modify: `frontend/package.json`

- [ ] **Step 1: Install**

```bash
cd frontend
npm install pdfjs-dist
```

- [ ] **Step 2: Verify**

`package.json` shows `"pdfjs-dist": "^4.0.0"` (or higher) under `dependencies`.

- [ ] **Step 3: Build sanity**

```bash
npm run check && npm run build
```

Expected: passes. Vite should auto-handle the new module.

- [ ] **Step 4: Commit**

```bash
git add frontend/package.json frontend/package-lock.json
git commit -m "[chore] add pdfjs-dist for native PDF rendering"
```

---

## Task 2: Skeleton — load PDF document, render page 1 only

**Files:**
- Modify: `frontend/src/lib/components/Viewers/PdfViewer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Mount | Loads `pdfjs-dist` dynamically; sets worker URL |
| Fetches PDF from `/server/download` with auth token | PDF document instance retrieved |
| Total pages | `document.numPages` exposed as state |
| Page 1 canvas | Rendered at 1.0 scale; visible centered in scroll container |
| Toolbar (placeholder) | Filename only for now |
| Error | If load fails, show "Failed to load PDF" message and use notifications.error |

- [ ] **Step 1: Replace `PdfViewer.svelte` with skeleton**

```svelte
<!-- frontend/src/lib/components/Viewers/PdfViewer.svelte -->
<script lang="ts">
  import { onMount, onDestroy, tick } from "svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  // Static URL import — Vite resolves to an asset path string. The worker code
  // itself stays out of the main bundle (it's loaded via the URL at runtime).
  import workerUrl from "pdfjs-dist/build/pdf.worker.mjs?url";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  let document_: import("pdfjs-dist").PDFDocumentProxy | null = $state(null);
  let pageCount = $state(0);
  let loading = $state(true);
  let canvases: HTMLCanvasElement[] = [];
  // Track in-flight render tasks per page so zoom/rerender can cancel them
  // before starting a new render (PDF.js logs warnings otherwise).
  const renderTasks = new Map<number, import("pdfjs-dist").RenderTask>();

  const pdfUrl = $derived(
    `/server/download?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );

  async function loadDocument() {
    try {
      const pdfjs = await import("pdfjs-dist");
      pdfjs.GlobalWorkerOptions.workerSrc = workerUrl;

      const task = pdfjs.getDocument({ url: pdfUrl, rangeChunkSize: 65536 });
      const doc = await task.promise;
      document_ = doc;
      pageCount = doc.numPages;
      loading = false;
      // tick() lets the {#each} below mount canvases before we touch them.
      await tick();
      await renderPage(1);
    } catch (error) {
      notifications.error(`Failed to load PDF: ${(error as Error).message}`);
      loading = false;
    }
  }

  async function renderPage(pageNumber: number) {
    if (!document_) return;
    const canvas = canvases[pageNumber - 1];
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    // Cancel any in-flight render for this page (e.g. zoom changed mid-render).
    renderTasks.get(pageNumber)?.cancel();

    const page = await document_.getPage(pageNumber);
    const viewport = page.getViewport({ scale: 1 });

    // HiDPI: render at devicePixelRatio so retina screens stay crisp.
    const outputScale = window.devicePixelRatio || 1;
    canvas.width = Math.floor(viewport.width * outputScale);
    canvas.height = Math.floor(viewport.height * outputScale);
    canvas.style.width = `${Math.floor(viewport.width)}px`;
    canvas.style.height = `${Math.floor(viewport.height)}px`;
    const transform: [number, number, number, number, number, number] | undefined =
      outputScale !== 1 ? [outputScale, 0, 0, outputScale, 0, 0] : undefined;

    const task = page.render({ canvasContext: ctx, viewport, transform });
    renderTasks.set(pageNumber, task);
    try {
      await task.promise;
    } catch (error) {
      // RenderingCancelledException is expected when zoom interrupts; ignore.
      if ((error as Error).name !== "RenderingCancelledException") throw error;
    } finally {
      renderTasks.delete(pageNumber);
    }
  }

  onMount(loadDocument);

  onDestroy(() => {
    for (const task of renderTasks.values()) task.cancel();
    renderTasks.clear();
    document_?.destroy();
    document_ = null;
  });
</script>

<div class="flex flex-col h-full w-full bg-bg-base">
  <div class="flex items-center px-3 h-10 border-b border-border-default bg-bg-overlay text-xs text-fg-secondary">
    <span class="truncate">{name}</span>
    <span class="ml-auto text-fg-muted">
      {#if loading}Loading…{:else}{pageCount} pages{/if}
    </span>
  </div>
  <div class="flex-1 overflow-auto p-4 flex flex-col items-center gap-4">
    {#each Array(pageCount) as _, i (i)}
      <canvas
        bind:this={canvases[i]}
        class="bg-bg-elevated shadow-lg"
      ></canvas>
    {/each}
  </div>
</div>
```

- [ ] **Step 2: Type check** → 0 errors. PDF.js types may complain — if so, follow Vite + PDF.js docs and ensure `pdfjs-dist` package has built-in types (it does as of v4).

- [ ] **Step 3: Manual verification**

```bash
npm run dev
```

Open a PDF file. Expected:
- Toolbar shows filename + "N pages"
- Page 1 renders as a canvas with dark background and PDF content visible
- Pages 2..N show empty canvases (will fix in Task 3)
- No console errors related to worker loading

- [ ] **Step 4: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/components/Viewers/PdfViewer.svelte
git commit -m "[feat] add PDF.js-based renderer with single-page rendering"
```

---

## Task 3: Render all pages with IntersectionObserver (lazy)

**Files:**
- Modify: `frontend/src/lib/components/Viewers/PdfViewer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Initial render | Pages 1-3 (or those in viewport) rendered eagerly |
| Scroll | Pages become visible, render on demand |
| Rendered pages cached | Re-scrolling doesn't re-render |
| Memory cap | Pages outside ±5 of current viewport can stay rendered (PDF.js holds page objects lightly) |

- [ ] **Step 1: Add IntersectionObserver-based rendering**

Replace the `renderPage` calls in `loadDocument` with sizing-then-lazy-render. The approach:
1. After getting document, pre-size each canvas to its page viewport (scale 1) so the layout is stable before rendering.
2. Set up an `IntersectionObserver` watching all canvases. When a canvas intersects the viewport (with `rootMargin: "200px"` for prefetch), call `renderPage` on it.
3. Track which pages are already rendered to avoid double-render.

```ts
let renderedPages = new Set<number>();
let observer: IntersectionObserver | null = null;

async function sizeAllPages() {
  if (!document_) return;
  for (let i = 1; i <= pageCount; i++) {
    const page = await document_.getPage(i);
    const viewport = page.getViewport({ scale: 1 });
    const canvas = canvases[i - 1];
    if (canvas) {
      canvas.width = viewport.width;
      canvas.height = viewport.height;
    }
  }
}

function setupObserver() {
  observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (!entry.isIntersecting) continue;
        const pageNumber = Number((entry.target as HTMLElement).dataset.page);
        if (renderedPages.has(pageNumber)) continue;
        renderedPages.add(pageNumber);
        renderPage(pageNumber);
      }
    },
    { rootMargin: "200px" },
  );
  for (const canvas of canvases) {
    if (canvas) observer.observe(canvas);
  }
}

// In loadDocument, after `pageCount = doc.numPages;`:
// (use $effect to set up observer when canvases are bound)
$effect(() => {
  if (pageCount > 0 && canvases.length === pageCount) {
    sizeAllPages().then(setupObserver);
  }
});
```

- [ ] **Step 2: Add `data-page` attribute to canvases**

```svelte
{#each Array(pageCount) as _, i (i)}
  <canvas
    bind:this={canvases[i]}
    data-page={i + 1}
    class="bg-bg-elevated shadow-lg"
  ></canvas>
{/each}
```

- [ ] **Step 3: Cleanup observer in onDestroy**

```ts
onDestroy(() => {
  observer?.disconnect();
  observer = null;
  document_?.destroy();
  document_ = null;
});
```

- [ ] **Step 4: Type check + manual verification**
- Scroll down: lower pages render as they enter viewport
- Scroll back up: already-rendered pages stay visible (no re-render flash)
- No console errors

- [ ] **Step 5: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/components/Viewers/PdfViewer.svelte
git commit -m "[feat] lazy-render PDF pages via IntersectionObserver"
```

---

## Task 4: Page navigation toolbar (prev / current / next)

**Files:**
- Modify: `frontend/src/lib/components/Viewers/PdfViewer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Toolbar has `[<]` `[input 3 ]/ 27 [>]` controls | Visible after filename |
| Click `<` | Scrolls smoothly to previous page canvas |
| Click `>` | Scrolls smoothly to next page canvas |
| Type page number into input + Enter | Scrolls to that page |
| Out-of-range input | Clamps to 1..pageCount |
| Current page indicator | Updates as user scrolls (via IntersectionObserver — track page most in view) |

- [ ] **Step 1: Add state + handlers**

```ts
import ChevronLeft from "lucide-svelte/icons/chevron-left";
import ChevronRight from "lucide-svelte/icons/chevron-right";

let currentPage = $state(1);

function scrollToPage(page: number) {
  const target = Math.max(1, Math.min(pageCount, page));
  canvases[target - 1]?.scrollIntoView({ behavior: "smooth", block: "start" });
  currentPage = target;
}

function onPageInput(event: KeyboardEvent) {
  if (event.key !== "Enter") return;
  const value = parseInt((event.currentTarget as HTMLInputElement).value, 10);
  if (Number.isFinite(value)) scrollToPage(value);
}
```

In `setupObserver`, also update `currentPage` when a canvas crosses center:

```ts
function setupObserver() {
  observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          const pageNumber = Number((entry.target as HTMLElement).dataset.page);
          if (!renderedPages.has(pageNumber)) {
            renderedPages.add(pageNumber);
            renderPage(pageNumber);
          }
          // Update current page if this entry is most visible
          if (entry.intersectionRatio > 0.5) currentPage = pageNumber;
        }
      }
    },
    { rootMargin: "200px", threshold: [0, 0.5, 1] },
  );
  for (const canvas of canvases) {
    if (canvas) observer.observe(canvas);
  }
}
```

- [ ] **Step 2: Add toolbar controls (replace the simple toolbar)**

```svelte
<div class="flex items-center gap-2 px-3 h-10 border-b border-border-default bg-bg-overlay text-xs text-fg-secondary">
  <span class="truncate max-w-[40%]">{name}</span>

  <span class="text-fg-muted">|</span>

  <button
    type="button"
    onclick={() => scrollToPage(currentPage - 1)}
    disabled={currentPage <= 1}
    aria-label="Previous page"
    class="inline-flex items-center justify-center w-8 h-8 rounded
           hover:bg-bg-hover hover:text-fg-accent disabled:opacity-50 disabled:hover:bg-transparent transition-colors"
  >
    <ChevronLeft size="16" />
  </button>

  <div class="flex items-center gap-1 font-mono tabular-nums">
    <input
      type="number"
      min="1"
      max={pageCount}
      value={currentPage}
      onkeydown={onPageInput}
      class="w-12 h-7 px-1 rounded border border-border-default bg-bg-base text-center text-xs"
    />
    <span class="text-fg-muted">/ {pageCount}</span>
  </div>

  <button
    type="button"
    onclick={() => scrollToPage(currentPage + 1)}
    disabled={currentPage >= pageCount}
    aria-label="Next page"
    class="inline-flex items-center justify-center w-8 h-8 rounded
           hover:bg-bg-hover hover:text-fg-accent disabled:opacity-50 disabled:hover:bg-transparent transition-colors"
  >
    <ChevronRight size="16" />
  </button>

  <span class="ml-auto text-fg-muted">{loading ? "Loading…" : ""}</span>
</div>
```

- [ ] **Step 3: Type check + manual verification**
- Toolbar shows page input "1 / N" with arrows
- Click `>` repeatedly → smoothly scrolls down page by page
- Type "10" + Enter → jumps to page 10
- Type "999" → clamps to last page
- Manual scroll → input updates as the visible page changes

- [ ] **Step 4: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/components/Viewers/PdfViewer.svelte
git commit -m "[feat] add page navigation toolbar for PDF viewer"
```

---

## Task 5: Zoom controls

**Files:**
- Modify: `frontend/src/lib/components/Viewers/PdfViewer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Zoom in button (+) | Scale increases by 0.25, max 4 |
| Zoom out button (-) | Scale decreases by 0.25, min 0.25 |
| Zoom percent display | Shows `100%`, `125%`, etc. |
| Fit-width button | Scale = container_width / page_width |
| Change scale | All rendered pages re-render at new scale; un-rendered keep new placeholder size |

- [ ] **Step 1: Add scale state + zoom handlers**

```ts
import ZoomIn from "lucide-svelte/icons/zoom-in";
import ZoomOut from "lucide-svelte/icons/zoom-out";
import Maximize2 from "lucide-svelte/icons/maximize-2";

let scale = $state(1);

function zoomIn() { scale = Math.min(4, scale + 0.25); }
function zoomOut() { scale = Math.max(0.25, scale - 0.25); }
async function fitWidth() {
  if (!document_) return;
  const firstPage = await document_.getPage(1);
  const baseViewport = firstPage.getViewport({ scale: 1 });
  // assume container is the parent of canvases with horizontal padding p-4 (16px each side)
  const container = canvases[0]?.parentElement;
  if (!container) return;
  const available = container.clientWidth - 32; // 16px padding * 2
  scale = available / baseViewport.width;
}

// when scale changes, re-render all rendered pages
$effect(() => {
  // depend on `scale` so this runs whenever it changes
  void scale;
  if (!document_ || renderedPages.size === 0) return;
  reRenderAllPages();
});

async function reRenderAllPages() {
  if (!document_) return;
  // Re-render already-rendered pages at the new scale.
  for (const pageNumber of [...renderedPages]) {
    await renderPage(pageNumber);
  }
  // Resize un-rendered placeholders so the page layout stays correct.
  for (let i = 1; i <= pageCount; i++) {
    if (renderedPages.has(i)) continue;
    const page = await document_.getPage(i);
    const viewport = page.getViewport({ scale });
    const canvas = canvases[i - 1];
    if (canvas) {
      const outputScale = window.devicePixelRatio || 1;
      canvas.width = Math.floor(viewport.width * outputScale);
      canvas.height = Math.floor(viewport.height * outputScale);
      canvas.style.width = `${Math.floor(viewport.width)}px`;
      canvas.style.height = `${Math.floor(viewport.height)}px`;
    }
  }
}

// Update existing renderPage to use `scale` (replace the Task 2 version):
async function renderPage(pageNumber: number) {
  if (!document_) return;
  const canvas = canvases[pageNumber - 1];
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  renderTasks.get(pageNumber)?.cancel();

  const page = await document_.getPage(pageNumber);
  const viewport = page.getViewport({ scale });

  const outputScale = window.devicePixelRatio || 1;
  canvas.width = Math.floor(viewport.width * outputScale);
  canvas.height = Math.floor(viewport.height * outputScale);
  canvas.style.width = `${Math.floor(viewport.width)}px`;
  canvas.style.height = `${Math.floor(viewport.height)}px`;
  const transform: [number, number, number, number, number, number] | undefined =
    outputScale !== 1 ? [outputScale, 0, 0, outputScale, 0, 0] : undefined;

  const task = page.render({ canvasContext: ctx, viewport, transform });
  renderTasks.set(pageNumber, task);
  try {
    await task.promise;
  } catch (error) {
    if ((error as Error).name !== "RenderingCancelledException") throw error;
  } finally {
    renderTasks.delete(pageNumber);
  }
}

// Update sizeAllPages similarly to use `scale` + HiDPI dimensions (mirror the
// canvas.style.width/height pattern above).
```

- [ ] **Step 2: Add zoom controls to toolbar** (after the next-page button):

```svelte
<span class="text-fg-muted">|</span>

<button
  type="button"
  onclick={zoomOut}
  aria-label="Zoom out"
  class="inline-flex items-center justify-center w-8 h-8 rounded
         hover:bg-bg-hover hover:text-fg-accent transition-colors"
>
  <ZoomOut size="16" />
</button>
<span class="font-mono tabular-nums w-12 text-center">{Math.round(scale * 100)}%</span>
<button
  type="button"
  onclick={zoomIn}
  aria-label="Zoom in"
  class="inline-flex items-center justify-center w-8 h-8 rounded
         hover:bg-bg-hover hover:text-fg-accent transition-colors"
>
  <ZoomIn size="16" />
</button>
<button
  type="button"
  onclick={fitWidth}
  aria-label="Fit width"
  class="inline-flex items-center justify-center w-8 h-8 rounded
         hover:bg-bg-hover hover:text-fg-accent transition-colors"
>
  <Maximize2 size="16" />
</button>
```

- [ ] **Step 3: Type check + manual verification**
- Click + → all rendered pages re-render bigger; un-rendered placeholders resize
- Click - → smaller
- Click fit-width → page fills horizontal space minus padding
- Zoom percent updates correctly

- [ ] **Step 4: `/code-review`** → 0 ❌ Critical. Note: re-render performance is acceptable for typical PDFs (<50 pages); for huge PDFs, only-render-on-scale-confirm could be added — flag in Risk Register.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/components/Viewers/PdfViewer.svelte
git commit -m "[feat] add zoom controls (in/out/fit-width) for PDF viewer"
```

---

## Task 6: Download button + final polish

**Files:**
- Modify: `frontend/src/lib/components/Viewers/PdfViewer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Download button on the right side of toolbar | Triggers download of the original PDF |
| Click opens `pdfUrl` in new tab with `download` attr | Browser downloads with the file's name |
| File line count | ≤280 lines |

- [ ] **Step 1: Add download button** (right-aligned, after zoom controls):

```svelte
import Download from "lucide-svelte/icons/download";
```

```svelte
<span class="text-fg-muted">|</span>
<a
  href={pdfUrl}
  download={name}
  target="_blank"
  rel="noopener"
  aria-label="Download PDF"
  class="inline-flex items-center justify-center w-8 h-8 rounded
         hover:bg-bg-hover hover:text-fg-accent transition-colors"
>
  <Download size="16" />
</a>
<span class="ml-auto text-fg-muted">{loading ? "Loading…" : ""}</span>
```

(Remove the duplicate `<span class="ml-auto ...">` from Task 4; only one should exist.)

- [ ] **Step 2: Line count check**

```bash
wc -l frontend/src/lib/components/Viewers/PdfViewer.svelte
```

Expected: ≤280 lines. If 250-280, ⚠️ acceptable. If > 280, extract `PdfToolbar.svelte`.

- [ ] **Step 3: Type check + manual verification**

- [ ] **Step 4: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/components/Viewers/PdfViewer.svelte
git commit -m "[feat] add PDF download button to toolbar"
```

---

## Task 7: Final integration verification + build

**Files:** None modified.

- [ ] **Step 1: Type check + build**

```bash
cd frontend && npm run check && npm run build
```

Expected: both pass; bundle build report shows pdfjs-dist as a dynamically loaded chunk (lazy import worked).

- [ ] **Step 2: Spec walkthrough**

| # | Action | Expected | Status |
|---|---|---|---|
| 1 | Open small PDF (1-2 pages) | Loads in ≤1 second, both pages rendered | |
| 2 | Open medium PDF (10-50 pages) | First page renders fast; scrolling triggers lazy render | |
| 3 | Open large PDF (100+ pages) | Initial load fast; smooth scroll; no obvious memory issues | |
| 4 | Click `<` / `>` | Smooth scroll between pages | |
| 5 | Type page number + Enter | Jumps correctly | |
| 6 | Zoom in repeatedly | Up to 400%, content stays crisp (canvas re-renders at scale) | |
| 7 | Zoom out repeatedly | Down to 25% | |
| 8 | Fit width | Page exactly fills container minus padding | |
| 9 | Download button | Browser downloads the PDF with correct filename | |
| 10 | Tab close | PDF.js document destroyed, no console error | |

- [ ] **Step 3: Final `/code-review`** on `PdfViewer.svelte` → 0 ❌ Critical.

- [ ] **Step 4: Push + PR**

```bash
git push -u origin feat/pdf-viewer
```

PR title: `[feat] PDF.js-based PDF viewer with zoom and page navigation`.
PR body: reference this plan + spec walkthrough table.

---

## Completion Criteria

- All 8 tasks (Task 0 - Task 7) committed.
- `npm run check`: 0 errors / 0 warnings.
- `npm run build`: succeeds; PDF.js chunk lazy-loaded.
- `/code-review`: 0 ❌ Critical.
- All 10 spec walkthrough rows pass.
- PR opened against `main`.

## Risk Register

| Risk | Mitigation |
|---|---|
| PDF.js bundle (~1.5MB) bloats initial app load | Lazy-load via `await import("pdfjs-dist")` inside `onMount` — only loads when user opens a PDF |
| PDF.js worker URL mis-configured under Vite | Vite's `?url` import handles this; verified pattern. If broken, fallback is `pdfjs-dist/legacy/build/pdf.worker.min.js` copied to `static/` and set explicitly |
| Big PDFs (100MB+) memory pressure | Range requests + lazy render mitigate; PDF.js holds page proxies, not full data |
| Re-render storm on zoom | Acceptable for typical sizes; v2 could debounce or render only visible pages on scale change |
| Backend Range support absent | Verified via `backend/tests/integration/stream_test.go`; `http.ServeContent` handles it |
| Cross-origin worker security | Worker loaded from same origin (Vite serves it); no CORS issue |
| Print fidelity | Browser's print dialog uses canvas — may need to enable text layer in v2 for searchable/selectable text in print output |
