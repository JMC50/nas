<script lang="ts">
  import { onMount, onDestroy, tick } from "svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  // Static URL import — Vite resolves to an asset path string. The worker code
  // itself stays out of the main bundle (it's loaded via the URL at runtime).
  import workerUrl from "pdfjs-dist/build/pdf.worker.mjs?url";
  import ChevronLeft from "lucide-svelte/icons/chevron-left";
  import ChevronRight from "lucide-svelte/icons/chevron-right";
  import ZoomIn from "lucide-svelte/icons/zoom-in";
  import ZoomOut from "lucide-svelte/icons/zoom-out";
  import Maximize2 from "lucide-svelte/icons/maximize-2";
  import Download from "lucide-svelte/icons/download";
  import MoreHorizontal from "lucide-svelte/icons/more-horizontal";
  import { ui } from "$lib/store/ui.svelte";
  import { clickOutside } from "$lib/actions/click-outside";

  interface Props {
    loc: string;
    name: string;
    urlOverride?: string;
  }

  let { loc, name, urlOverride }: Props = $props();

  let document_: import("pdfjs-dist").PDFDocumentProxy | null = $state(null);
  let pageCount = $state(0);
  let loading = $state(true);
  let canvases = $state<HTMLCanvasElement[]>([]);
  let rendered = new Set<number>();
  let observer: IntersectionObserver | null = null;
  let currentPage = $state(1);
  let scale = $state(1);
  // Track in-flight render tasks per page so zoom/rerender can cancel them
  // before starting a new render (PDF.js logs warnings otherwise).
  const renderTasks = new Map<number, import("pdfjs-dist").RenderTask>();

  const pdfUrl = $derived(
    urlOverride ??
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
    } catch (error) {
      notifications.error(`Failed to load PDF: ${(error as Error).message}`);
      loading = false;
    }
  }

  async function renderPage(pageNumber: number) {
    if (!document_) return;
    const canvas = canvases[pageNumber - 1];
    if (!canvas) return;
    const context = canvas.getContext("2d");
    if (!context) return;

    // Cancel any in-flight render for this page (e.g. zoom changed mid-render).
    renderTasks.get(pageNumber)?.cancel();

    const page = await document_.getPage(pageNumber);
    const viewport = page.getViewport({ scale });

    // HiDPI: render at devicePixelRatio so retina screens stay crisp.
    const outputScale = window.devicePixelRatio || 1;
    canvas.width = Math.floor(viewport.width * outputScale);
    canvas.height = Math.floor(viewport.height * outputScale);
    canvas.style.width = `${Math.floor(viewport.width)}px`;
    canvas.style.height = `${Math.floor(viewport.height)}px`;
    const transform: [number, number, number, number, number, number] | undefined =
      outputScale !== 1 ? [outputScale, 0, 0, outputScale, 0, 0] : undefined;

    const task = page.render({ canvas, canvasContext: context, viewport, transform });
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

  async function sizeAllPages() {
    if (!document_) return;
    for (let i = 1; i <= pageCount; i++) {
      const page = await document_.getPage(i);
      const viewport = page.getViewport({ scale });
      const canvas = canvases[i - 1];
      if (!canvas) continue;
      const outputScale = window.devicePixelRatio || 1;
      canvas.width = Math.floor(viewport.width * outputScale);
      canvas.height = Math.floor(viewport.height * outputScale);
      canvas.style.width = `${Math.floor(viewport.width)}px`;
      canvas.style.height = `${Math.floor(viewport.height)}px`;
    }
  }

  function setupObs() {
    observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue;
          const pageNumber = Number((entry.target as HTMLElement).dataset.page);
          if (!rendered.has(pageNumber)) {
            rendered.add(pageNumber);
            renderPage(pageNumber);
          }
          if (entry.intersectionRatio > 0.5) currentPage = pageNumber;
        }
      },
      { rootMargin: "200px", threshold: [0, 0.5, 1] },
    );
    for (const canvas of canvases) {
      if (canvas) observer.observe(canvas);
    }
  }

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

  function zoomIn() {
    scale = Math.min(4, scale + 0.25);
  }

  function zoomOut() {
    scale = Math.max(0.25, scale - 0.25);
  }

  async function fitWidth() {
    if (!document_) return;
    const firstPage = await document_.getPage(1);
    const baseViewport = firstPage.getViewport({ scale: 1 });
    const container = canvases[0]?.parentElement;
    if (!container) return;
    const available = container.clientWidth - 32;
    scale = available / baseViewport.width;
  }

  async function applyScale() {
    if (!document_) return;
    for (const pageNumber of [...rendered]) {
      await renderPage(pageNumber);
    }
    for (let i = 1; i <= pageCount; i++) {
      if (rendered.has(i)) continue;
      const page = await document_.getPage(i);
      const viewport = page.getViewport({ scale });
      const canvas = canvases[i - 1];
      if (!canvas) continue;
      const outputScale = window.devicePixelRatio || 1;
      canvas.width = Math.floor(viewport.width * outputScale);
      canvas.height = Math.floor(viewport.height * outputScale);
      canvas.style.width = `${Math.floor(viewport.width)}px`;
      canvas.style.height = `${Math.floor(viewport.height)}px`;
    }
  }

  $effect(() => {
    if (pageCount > 0 && canvases.length === pageCount) {
      sizeAllPages().then(setupObs);
    }
  });

  $effect(() => {
    void scale;
    if (!document_ || rendered.size === 0) return;
    applyScale();
  });

  // Mobile-only: zoom controls + fit-width + download collapse into a
  // dropdown so the toolbar fits on a 375px width with prev / page / next
  // staying visible.
  let zoomMenuOpen = $state(false);

  function pickZoomOut() {
    zoomOut();
    zoomMenuOpen = false;
  }
  function pickZoomIn() {
    zoomIn();
    zoomMenuOpen = false;
  }
  function pickFitWidth() {
    fitWidth();
    zoomMenuOpen = false;
  }

  onMount(loadDocument);

  onDestroy(() => {
    observer?.disconnect();
    observer = null;
    for (const task of renderTasks.values()) task.cancel();
    renderTasks.clear();
    document_?.destroy();
    document_ = null;
  });
</script>

<div class="flex flex-col h-full w-full bg-bg-base">
  <div class="sg-viewer-bar">
    <span class="truncate min-w-0 max-w-[40%] text-fg-primary font-semibold" title={name}>{name}</span>

    <span class="sg-viewer-sep" aria-hidden="true"></span>

    <button
      type="button"
      onclick={() => scrollToPage(currentPage - 1)}
      disabled={currentPage <= 1}
      aria-label="Previous page"
      class="sg-btn sg-btn-ghost sg-btn-icon"
    >
      <ChevronLeft size="16" />
    </button>

    <div class="flex items-center gap-1.5 tabular-nums">
      <input
        type="number"
        min="1"
        max={pageCount}
        value={currentPage}
        onkeydown={onPageInput}
        aria-label="Page number"
        class="sg-input w-16 text-center tabular-nums"
      />
      <span class="text-fg-muted">/ {pageCount}</span>
    </div>

    <button
      type="button"
      onclick={() => scrollToPage(currentPage + 1)}
      disabled={currentPage >= pageCount}
      aria-label="Next page"
      class="sg-btn sg-btn-ghost sg-btn-icon"
    >
      <ChevronRight size="16" />
    </button>

    <span class="hidden md:block sg-viewer-sep" aria-hidden="true"></span>

    <button
      type="button"
      onclick={zoomOut}
      aria-label="Zoom out"
      class="hidden md:inline-flex sg-btn sg-btn-ghost sg-btn-icon"
    >
      <ZoomOut size="16" />
    </button>
    <span class="hidden md:inline tabular-nums w-12 text-center">{Math.round(scale * 100)}%</span>
    <button
      type="button"
      onclick={zoomIn}
      aria-label="Zoom in"
      class="hidden md:inline-flex sg-btn sg-btn-ghost sg-btn-icon"
    >
      <ZoomIn size="16" />
    </button>
    <button
      type="button"
      onclick={fitWidth}
      aria-label="Fit width"
      class="hidden md:inline-flex sg-btn sg-btn-ghost sg-btn-icon"
    >
      <Maximize2 size="16" />
    </button>

    <span class="hidden md:block sg-viewer-sep" aria-hidden="true"></span>
    <a
      href={pdfUrl}
      download={name}
      target="_blank"
      rel="noopener"
      aria-label="Download PDF"
      class="hidden md:inline-flex sg-btn sg-btn-ghost sg-btn-icon"
    >
      <Download size="16" />
    </a>

    <!-- Mobile: zoom / fit / download collapse into an overflow menu -->
    {#if ui.isMobile}
      <div class="relative" use:clickOutside={{ onOutside: () => (zoomMenuOpen = false), enabled: zoomMenuOpen }}>
        <button
          type="button"
          onclick={(event) => {
            event.stopPropagation();
            zoomMenuOpen = !zoomMenuOpen;
          }}
          aria-label="More PDF actions"
          aria-expanded={zoomMenuOpen}
          class="sg-btn sg-btn-ghost sg-btn-icon"
        >
          <MoreHorizontal size="16" />
        </button>
        {#if zoomMenuOpen}
          <div
            class="absolute right-0 top-full mt-1 z-30 sg-menu animate-pop origin-top-right"
            role="menu"
          >
            <button
              type="button"
              onclick={pickZoomOut}
              class="sg-menu-item"
              role="menuitem"
            >
              <ZoomOut size="16" />
              <span>Zoom out</span>
              <span class="ml-auto tabular-nums text-fg-muted">{Math.round(scale * 100)}%</span>
            </button>
            <button
              type="button"
              onclick={pickZoomIn}
              class="sg-menu-item"
              role="menuitem"
            >
              <ZoomIn size="16" />
              <span>Zoom in</span>
            </button>
            <button
              type="button"
              onclick={pickFitWidth}
              class="sg-menu-item"
              role="menuitem"
            >
              <Maximize2 size="16" />
              <span>Fit width</span>
            </button>
            <div class="sg-menu-sep" role="separator"></div>
            <a
              href={pdfUrl}
              download={name}
              target="_blank"
              rel="noopener"
              class="sg-menu-item"
              role="menuitem"
            >
              <Download size="16" />
              <span>Download</span>
            </a>
          </div>
        {/if}
      </div>
    {/if}

    <span class="ml-auto text-xs text-fg-muted" role="status">{loading ? "Loading…" : ""}</span>
  </div>
  <div class="flex-1 overflow-auto p-4 flex flex-col items-center gap-4">
    {#each Array(pageCount) as _, i (i)}
      <canvas
        bind:this={canvases[i]}
        data-page={i + 1}
        class="max-w-full bg-bg-elevated rounded-chip shadow-overlay"
      ></canvas>
    {/each}
  </div>
</div>
