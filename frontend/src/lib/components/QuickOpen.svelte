<script lang="ts">
  import { onMount, onDestroy, tick } from "svelte";
  import Search from "lucide-svelte/icons/search";
  import Folder from "lucide-svelte/icons/folder";
  import FileText from "lucide-svelte/icons/file-text";
  import ImageIcon from "lucide-svelte/icons/image";
  import Film from "lucide-svelte/icons/film";
  import Music from "lucide-svelte/icons/music";
  import FileType from "lucide-svelte/icons/file-type";
  import Users from "lucide-svelte/icons/users";
  import Settings from "lucide-svelte/icons/settings";
  import History from "lucide-svelte/icons/history";
  import User from "lucide-svelte/icons/user";
  import Cpu from "lucide-svelte/icons/cpu";
  import { ui } from "$lib/store/ui.svelte";
  import { tabs } from "$lib/store/tabs.svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { pickViewer } from "$lib/components/Viewers/registry";
  import type { TabKind, Tab } from "$lib/types";

  const KIND_TO_ICON: Record<TabKind, typeof Folder> = {
    explorer: Folder,
    text: FileText,
    image: ImageIcon,
    video: Film,
    audio: Music,
    pdf: FileType,
    office: FileText,
    "user-manager": Users,
    settings: Settings,
    activity: History,
    account: User,
    system: Cpu,
    "music-library": Music,
    "video-library": Film,
  };

  let query = $state("");
  let highlighted = $state(0);
  let inputEl: HTMLInputElement | undefined = $state();
  // visualViewport.height tracks the available area shrunk by the on-screen
  // keyboard on mobile. Fallback to `100dvh` when null (desktop / unsupported).
  let modalHeight = $state<number | null>(null);

  interface FileHit {
    name: string;
    isFolder: boolean;
    extensions: string;
    loc: string;
  }
  let fileHits = $state<FileHit[]>([]);
  let fileSearching = $state(false);
  let searchToken = 0;

  const tabHits = $derived(
    query
      ? tabs.list.filter((tab) =>
          tab.title.toLowerCase().includes(query.toLowerCase()),
        )
      : tabs.list,
  );

  // Debounced backend search. When the modal is open and query is non-empty,
  // fire `/server/searchInAllFiles?query=`. Token-based cancellation prevents
  // stale responses from overwriting newer results.
  $effect(() => {
    if (!ui.quickOpenVisible) return;
    if (!query) {
      fileHits = [];
      fileSearching = false;
      return;
    }
    const myToken = ++searchToken;
    fileSearching = true;
    const timer = window.setTimeout(async () => {
      try {
        const r = await fetch(
          `/server/searchInAllFiles?query=${encodeURIComponent(query)}&token=${encodeURIComponent(auth.token ?? "")}`,
        );
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        const data = await r.json();
        if (myToken !== searchToken) return; // stale — newer query in flight
        const rows = (Array.isArray(data) ? data : (data?.data ?? data?.entries ?? [])) as FileHit[];
        fileHits = rows.slice(0, 50);
      } catch {
        if (myToken === searchToken) fileHits = [];
      } finally {
        if (myToken === searchToken) fileSearching = false;
      }
    }, 180);
    return () => window.clearTimeout(timer);
  });

  type Hit =
    | { kind: "tab"; tab: Tab }
    | { kind: "file"; hit: FileHit };
  const filtered = $derived<Hit[]>([
    ...tabHits.map<Hit>((tab) => ({ kind: "tab", tab })),
    ...fileHits.map<Hit>((hit) => ({ kind: "file", hit })),
  ]);

  $effect(() => {
    if (highlighted >= filtered.length) highlighted = 0;
  });

  async function focusInput() {
    await tick();
    inputEl?.focus();
    inputEl?.select();
  }

  function syncViewportHeight() {
    if (typeof window === "undefined") return;
    const vv = window.visualViewport;
    modalHeight = vv ? vv.height : null;
  }

  $effect(() => {
    if (!ui.quickOpenVisible) return;
    query = "";
    highlighted = 0;
    focusInput();
  });

  // Track virtual-keyboard height while QuickOpen is open AND the viewport
  // is mobile-sized; desktop uses a centered card with fixed intrinsic
  // height. The on-screen keyboard shrinks `visualViewport.height` below
  // `innerHeight`, so we mirror it onto the modal's height inline style.
  $effect(() => {
    if (!ui.quickOpenVisible || !ui.isMobile) return;
    if (typeof window === "undefined") return;
    const vv = window.visualViewport;
    if (!vv) return;
    syncViewportHeight();
    vv.addEventListener("resize", syncViewportHeight);
    return () => {
      vv.removeEventListener("resize", syncViewportHeight);
      modalHeight = null;
    };
  });

  function pickTab(tab: Tab) {
    tabs.setActive(tab.id);
    ui.closeQuickOpen();
  }

  function pickFile(hit: FileHit) {
    // Strip trailing slash from loc; normalize to string-array path
    const locPath = (hit.loc ?? "/").replace(/\/+$/, "") || "/";
    const segments = locPath === "/" ? [] : locPath.replace(/^\//, "").split("/");
    if (hit.isFolder) {
      // Open as Explorer tab navigated into the folder
      tabs.openExplorer([...segments, hit.name]);
    } else {
      const kind = pickViewer(hit.extensions);
      if (kind === null) {
        // Unknown viewer — open Explorer at the parent loc instead
        tabs.openExplorer(segments);
      } else {
        tabs.open({
          kind,
          title: hit.name,
          icon: kind,
          payload: { loc: locPath, name: hit.name },
          closable: true,
        });
      }
    }
    ui.closeQuickOpen();
  }

  function pick(item: Hit) {
    if (item.kind === "tab") pickTab(item.tab);
    else pickFile(item.hit);
  }

  function onKeyDown(event: KeyboardEvent) {
    if (!ui.quickOpenVisible) return;
    if (event.key === "Escape") {
      event.preventDefault();
      ui.closeQuickOpen();
      return;
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      highlighted = Math.min(filtered.length - 1, highlighted + 1);
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      highlighted = Math.max(0, highlighted - 1);
      return;
    }
    if (event.key === "Enter") {
      event.preventDefault();
      const target = filtered[highlighted];
      if (target) pick(target);
    }
  }

  onMount(() => window.addEventListener("keydown", onKeyDown));
  onDestroy(() => window.removeEventListener("keydown", onKeyDown));
</script>

{#if ui.quickOpenVisible}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 bg-bg-base md:bg-bg-base/70 md:backdrop-blur-sm flex items-start justify-center md:pt-[20vh]"
    onclick={() => ui.closeQuickOpen()}
  >
    <div
      role="dialog"
      tabindex="-1"
      aria-modal="true"
      aria-label="Quick open"
      class="w-full h-full md:h-auto md:max-w-lg md:mx-4 md:rounded-lg bg-bg-surface md:border md:border-border-strong md:shadow-[0_8px_32px_rgba(0,0,0,0.6)] overflow-hidden cursor-default flex flex-col"
      style={ui.isMobile && modalHeight !== null ? `height: ${modalHeight}px` : ""}
      onclick={(event) => event.stopPropagation()}
    >
      <div class="flex items-center gap-2 px-3 h-11 border-b border-border-default shrink-0">
        <Search size="14" class="text-fg-muted shrink-0" />
        <input
          bind:this={inputEl}
          bind:value={query}
          type="search"
          name="quickopen-q"
          autocomplete="off"
          autocorrect="off"
          spellcheck="false"
          placeholder="Search files, folders, or tabs…"
          class="flex-1 bg-transparent text-sm text-fg-primary placeholder:text-fg-muted outline-none"
        />
        <span class="text-[10px] text-fg-muted font-mono">ESC</span>
      </div>

      <div class="flex-1 md:max-h-80 overflow-y-auto py-1">
        {#if filtered.length === 0 && !fileSearching}
          <div class="px-3 py-6 text-center text-xs text-fg-muted">
            {query ? "No matches." : "Type to search files, folders, or open tabs."}
          </div>
        {/if}
        {#if fileSearching}
          <div class="px-3 py-1.5 text-[10px] text-fg-muted font-mono">Searching files…</div>
        {/if}
        {#each filtered as item, index (item.kind === "tab" ? `tab:${item.tab.id}` : `file:${item.hit.loc}/${item.hit.name}`)}
          {@const active = index === highlighted}
          {#if item.kind === "tab"}
            {@const tab = item.tab}
            {@const Icon = KIND_TO_ICON[tab.kind] ?? Folder}
            <button
              type="button"
              class="w-full flex items-center gap-2.5 px-3 h-9 text-left transition-colors {active ? 'bg-bg-hover text-fg-primary' : 'text-fg-secondary hover:bg-bg-hover/60'}"
              onclick={() => pick(item)}
              onmouseenter={() => (highlighted = index)}
            >
              <Icon size="14" class="shrink-0 text-fg-muted" />
              <span class="text-sm flex-1 truncate">{tab.title}</span>
              <span class="text-[10px] text-fg-muted font-mono uppercase">tab · {tab.kind}</span>
            </button>
          {:else}
            {@const hit = item.hit}
            {@const Icon = hit.isFolder ? Folder : KIND_TO_ICON[pickViewer(hit.extensions) ?? "text"] ?? FileText}
            <button
              type="button"
              class="w-full flex items-center gap-2.5 px-3 h-9 text-left transition-colors {active ? 'bg-bg-hover text-fg-primary' : 'text-fg-secondary hover:bg-bg-hover/60'}"
              onclick={() => pick(item)}
              onmouseenter={() => (highlighted = index)}
            >
              <Icon size="14" class="shrink-0 text-fg-muted" />
              <span class="text-sm truncate">{hit.name}</span>
              <span class="text-[10px] text-fg-muted font-mono truncate flex-1 text-right">{hit.loc || "/"}</span>
            </button>
          {/if}
        {/each}
      </div>
    </div>
  </div>
{/if}
