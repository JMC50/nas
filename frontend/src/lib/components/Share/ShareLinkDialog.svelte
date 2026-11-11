<script lang="ts">
  import { onMount, tick } from "svelte";
  import Link2 from "lucide-svelte/icons/link-2";
  import X from "lucide-svelte/icons/x";
  import Copy from "lucide-svelte/icons/copy";
  import Check from "lucide-svelte/icons/check";
  import { auth } from "$lib/store/auth.svelte";
  import { ui } from "$lib/store/ui.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  import { iconFor, type FolderEntry } from "$lib/components/Explorer/icon-for";

  interface Props {
    target: FolderEntry;
    loc: string[];
    onClose: () => void;
  }

  let { target, loc, onClose }: Props = $props();

  const EXPIRY_CHOICES = [
    { value: "1h", label: "1 hour" },
    { value: "1d", label: "1 day" },
    { value: "7d", label: "7 days" },
    { value: "30d", label: "30 days" },
  ];

  let expiresIn = $state("7d");
  let busy = $state(false);
  let errorMessage = $state("");
  let created = $state<{ url: string; expiresAt: number } | null>(null);
  let copied = $state(false);
  let copiedTimer: ReturnType<typeof setTimeout> | undefined;

  let dialogEl: HTMLDivElement | undefined = $state();
  let urlEl: HTMLInputElement | undefined = $state();
  let copyEl: HTMLButtonElement | undefined = $state();

  const Icon = $derived(iconFor(target));
  const locPath = $derived("/" + loc.join("/"));
  const kindWord = $derived(target.isFolder ? "folder" : "file");

  function expiryText(at: number): string {
    return new Date(at).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
  }

  function failureText(status: number): string {
    if (status === 403) return `You don't have permission to share this ${kindWord}.`;
    if (status === 404) return `This ${kindWord} no longer exists.`;
    return `Could not create the link (HTTP ${status}).`;
  }

  async function create() {
    busy = true;
    errorMessage = "";
    try {
      const response = await fetch("/server/share-links", {
        method: "POST",
        headers: { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" },
        body: JSON.stringify({ loc: locPath, name: target.name, expiresIn }),
      });
      if (!response.ok) {
        errorMessage = failureText(response.status);
        return;
      }
      const data = (await response.json()) as { path: string; expiresAt: number };
      created = { url: `${location.origin}${data.path}`, expiresAt: data.expiresAt };
      await tick();
      copyEl?.focus();
    } catch {
      errorMessage = "Could not reach the server.";
    } finally {
      busy = false;
    }
  }

  async function copy() {
    if (!created) return;
    let ok = false;
    try {
      await navigator.clipboard.writeText(created.url);
      ok = true;
    } catch {
      urlEl?.select();
      ok = document.execCommand("copy");
    }
    if (!ok) {
      notifications.error("Copy failed. Select the link and copy it.");
      return;
    }
    copied = true;
    clearTimeout(copiedTimer);
    copiedTimer = setTimeout(() => (copied = false), 2000);
    notifications.success("Link copied.");
  }

  function focusables(): HTMLElement[] {
    if (!dialogEl) return [];
    return Array.from(
      dialogEl.querySelectorAll<HTMLElement>("button:not(:disabled), input:not(:disabled)"),
    ).filter((element) => !(element instanceof HTMLInputElement && element.type === "radio" && !element.checked));
  }

  function onKey(event: KeyboardEvent) {
    if (event.key === "Escape") {
      event.preventDefault();
      event.stopPropagation();
      onClose();
      return;
    }
    if (event.key !== "Tab") return;
    const items = focusables();
    if (items.length === 0) return;
    const first = items[0];
    const last = items[items.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  onMount(() => {
    dialogEl?.querySelector<HTMLInputElement>("input[type=radio]:checked")?.focus();
    return () => clearTimeout(copiedTimer);
  });
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="fixed inset-0 z-50 bg-scrim flex justify-center {ui.isMobile ? 'items-end' : 'items-center'} animate-fade"
  onclick={onClose}
>
  <div
    bind:this={dialogEl}
    role="dialog"
    tabindex="-1"
    aria-modal="true"
    aria-labelledby="share-title"
    aria-describedby="share-note"
    class="w-full max-h-[calc(100dvh-1rem)] bg-bg-overlay shadow-overlay cursor-default flex flex-col {ui.isMobile
      ? 'rounded-t-panel pb-[env(safe-area-inset-bottom)] animate-sheet'
      : 'max-w-[440px] mx-4 rounded-panel animate-pop'}"
    onclick={(event) => event.stopPropagation()}
    onkeydown={onKey}
  >
    <header class="flex items-center gap-2 min-h-12 pl-4 pr-2 border-b border-border-default">
      <Link2 size="18" class="shrink-0 text-fg-secondary" />
      <h2 id="share-title" class="sg-section-title">Create download link</h2>
      <button type="button" class="ml-auto sg-btn sg-btn-ghost sg-btn-icon" onclick={onClose} aria-label="Close">
        <X size="18" />
      </button>
    </header>

    <div class="flex flex-col gap-4 p-4 min-h-0 overflow-y-auto">
      <div class="flex items-center gap-3 min-w-0">
        <Icon size="20" class="shrink-0 {target.isFolder ? 'text-fg-secondary' : 'text-fg-muted'}" />
        <div class="min-w-0">
          <p class="truncate text-sm text-fg-primary" title={target.name}>{target.name}</p>
          <p class="truncate text-xs text-fg-muted">
            <span class="font-mono">{locPath}</span>
            {#if target.isFolder}<span> · Downloads as a .zip file</span>{/if}
          </p>
        </div>
      </div>

      {#if !created}
        <fieldset disabled={busy}>
          <legend class="sg-label">Expires after</legend>
          <div class="sg-seg grid grid-cols-4 w-full">
            {#each EXPIRY_CHOICES as choice (choice.value)}
              <label class="sg-seg-btn">
                <input type="radio" name="share-expiry" value={choice.value} bind:group={expiresIn} class="sr-only" />
                {choice.label}
              </label>
            {/each}
          </div>
        </fieldset>
        <p id="share-note" class="text-xs text-fg-muted">
          Anyone with the link can download this {kindWord} until it expires. You can revoke it under Links.
        </p>
        {#if errorMessage}
          <p class="text-xs text-fg-danger" role="alert">{errorMessage}</p>
        {/if}
      {:else}
        <div>
          <label for="share-url" class="sg-label">Link</label>
          <div class="flex gap-2">
            <input
              bind:this={urlEl}
              id="share-url"
              type="text"
              readonly
              value={created.url}
              class="sg-input flex-1 min-w-0 font-mono text-xs"
              onfocus={(event) => event.currentTarget.select()}
            />
            <button bind:this={copyEl} type="button" class="sg-btn sg-btn-primary" onclick={copy}>
              {#if copied}<Check size="16" />Copied{:else}<Copy size="16" />Copy{/if}
            </button>
          </div>
        </div>
        <p id="share-note" class="text-xs text-fg-muted" role="status">
          Link created. It expires {expiryText(created.expiresAt)}. This is the only time the full link is shown.
        </p>
      {/if}
    </div>

    <footer class="flex justify-end gap-2 px-4 py-3 border-t border-border-default">
      {#if !created}
        <button type="button" class="sg-btn sg-btn-ghost" onclick={onClose}>Cancel</button>
        <button type="button" class="sg-btn sg-btn-primary" onclick={create} disabled={busy}>
          {busy ? "Creating…" : "Create link"}
        </button>
      {:else}
        <button type="button" class="sg-btn" onclick={onClose}>Done</button>
      {/if}
    </footer>
  </div>
</div>
