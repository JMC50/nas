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
  let rendered = new Set<number>();
  let observer: IntersectionObserver | null = null;
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
    const viewport = page.getViewport({ scale: 1 });

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
      const viewport = page.getViewport({ scale: 1 });
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
          if (rendered.has(pageNumber)) continue;
          rendered.add(pageNumber);
          renderPage(pageNumber);
        }
      },
      { rootMargin: "200px" },
    );
    for (const canvas of canvases) {
      if (canvas) observer.observe(canvas);
    }
  }

  $effect(() => {
    if (pageCount > 0 && canvases.length === pageCount) {
      sizeAllPages().then(setupObs);
    }
  });

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
        data-page={i + 1}
        class="bg-bg-elevated shadow-lg"
      ></canvas>
    {/each}
  </div>
</div>
