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

  const filtered = $derived(
    query
      ? tabs.list.filter((tab) =>
          tab.title.toLowerCase().includes(query.toLowerCase()),
        )
      : tabs.list,
  );

  $effect(() => {
    if (highlighted >= filtered.length) highlighted = 0;
  });

  async function focusInput() {
    await tick();
    inputEl?.focus();
    inputEl?.select();
  }

  $effect(() => {
    if (ui.quickOpenVisible) {
      query = "";
      highlighted = 0;
      focusInput();
    }
  });

  function pick(tab: Tab) {
    tabs.setActive(tab.id);
    ui.closeQuickOpen();
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
    class="fixed inset-0 z-50 bg-bg-base/70 backdrop-blur-sm flex items-start justify-center pt-[20vh]"
    onclick={() => ui.closeQuickOpen()}
  >
    <div
      role="dialog"
      tabindex="-1"
      aria-modal="true"
      aria-label="Quick open"
      class="w-full max-w-lg mx-4 rounded-lg bg-bg-surface border border-border-strong shadow-[0_8px_32px_rgba(0,0,0,0.6)] overflow-hidden cursor-default"
      onclick={(event) => event.stopPropagation()}
    >
      <div class="flex items-center gap-2 px-3 h-11 border-b border-border-default">
        <Search size="14" class="text-fg-muted shrink-0" />
        <input
          bind:this={inputEl}
          bind:value={query}
          type="text"
          placeholder="Search open tabs…"
          class="flex-1 bg-transparent text-sm text-fg-primary placeholder:text-fg-muted outline-none"
        />
        <span class="text-[10px] text-fg-muted font-mono">ESC</span>
      </div>

      <div class="max-h-80 overflow-y-auto py-1">
        {#if filtered.length === 0}
          <div class="px-3 py-6 text-center text-xs text-fg-muted">No matching tabs.</div>
        {/if}
        {#each filtered as tab, index (tab.id)}
          {@const Icon = KIND_TO_ICON[tab.kind] ?? Folder}
          {@const active = index === highlighted}
          <button
            type="button"
            class="w-full flex items-center gap-2.5 px-3 h-9 text-left transition-colors {active ? 'bg-bg-hover text-fg-primary' : 'text-fg-secondary hover:bg-bg-hover/60'}"
            onclick={() => pick(tab)}
            onmouseenter={() => (highlighted = index)}
          >
            <Icon size="14" class="shrink-0 text-fg-muted" />
            <span class="text-sm flex-1 truncate">{tab.title}</span>
            <span class="text-[10px] text-fg-muted font-mono uppercase">{tab.kind}</span>
          </button>
        {/each}
      </div>
    </div>
  </div>
{/if}
