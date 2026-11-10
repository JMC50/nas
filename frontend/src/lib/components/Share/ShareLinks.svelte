<script lang="ts">
  import { onMount } from "svelte";
  import Link2 from "lucide-svelte/icons/link-2";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import FolderIcon from "lucide-svelte/icons/folder";
  import FileIcon from "lucide-svelte/icons/file";
  import { auth } from "$lib/store/auth.svelte";
  import { notifications } from "$lib/store/notifications.svelte";

  type LinkStatus = "active" | "expired" | "revoked";

  interface ShareLink {
    id: number;
    loc: string;
    name: string;
    isFolder: boolean;
    creatorId: string;
    creatorName: string;
    createdAt: number;
    expiresAt: number;
    revokedAt: number | null;
    downloadCount: number;
    status: LinkStatus;
  }

  const STATUS_LABEL: Record<LinkStatus, string> = {
    active: "Active",
    expired: "Expired",
    revoked: "Revoked",
  };

  const STATUS_TEXT: Record<LinkStatus, string> = {
    active: "text-fg-success",
    expired: "text-fg-muted",
    revoked: "text-fg-danger",
  };

  let links = $state<ShareLink[]>([]);
  let showsAll = $state(false);
  let loading = $state(true);
  let loadError = $state("");
  let revoking = $state<number | null>(null);

  async function load() {
    loading = true;
    loadError = "";
    try {
      const response = await fetch("/server/share-links", {
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = (await response.json()) as { links: ShareLink[]; all: boolean };
      links = data.links;
      showsAll = data.all;
    } catch (cause) {
      loadError = (cause as Error).message;
    } finally {
      loading = false;
    }
  }

  async function revoke(link: ShareLink) {
    if (!confirm(`Revoke the link to "${link.name}"? Anyone who opens it will get "not found".`)) return;
    revoking = link.id;
    try {
      const response = await fetch(`/server/share-links/${link.id}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      notifications.success("Link revoked.");
      await load();
    } catch (cause) {
      notifications.error(`Could not revoke the link: ${(cause as Error).message}`);
    } finally {
      revoking = null;
    }
  }

  function shortTime(time: number): string {
    return new Date(time).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
  }

  function fullPath(link: ShareLink): string {
    return `${link.loc.replace(/\/+$/, "")}/${link.name}`;
  }

  function downloadsText(count: number): string {
    return `${count} ${count === 1 ? "download" : "downloads"}`;
  }

  onMount(load);
</script>

<section class="flex flex-col h-full bg-bg-base overflow-hidden">
  <header class="sg-page-head">
    <Link2 size="18" class="text-fg-secondary" />
    <h1 class="sg-page-title">Download links</h1>
    {#if !loading || links.length > 0}
      <span class="sg-count">{links.length}</span>
    {/if}
    <button type="button" class="ml-auto sg-btn sg-btn-ghost" onclick={load} aria-label="Refresh" disabled={loading}>
      <RefreshCw size="16" />
      <span class="hidden sm:inline">Refresh</span>
    </button>
  </header>

  <div class="flex-1 overflow-auto">
    {#if loadError}
      <div class="m-gutter p-3 rounded-panel border border-fg-danger/45 text-sm text-fg-danger" role="alert">
        Failed to load links: {loadError}
      </div>
    {:else if loading && links.length === 0}
      <div class="px-gutter py-6 text-sm text-fg-muted" role="status">Loading…</div>
    {:else if links.length === 0}
      <div class="flex flex-col items-center gap-2 px-gutter py-16 text-center">
        <Link2 size="28" class="text-fg-muted" />
        <p class="text-sm font-semibold text-fg-primary">No download links yet.</p>
        <p class="text-xs text-fg-muted">Right-click a file or folder and choose Create download link.</p>
      </div>
    {:else}
      <p class="px-gutter py-3 text-xs text-fg-muted">
        {showsAll ? "Links created by every user." : "Links you created."} The full link is shown only once, when it is created.
      </p>
      <table class="w-full text-sm table-fixed border-t border-border-default">
        <thead>
          <tr class="text-xs text-fg-muted border-b border-border-default">
            <th scope="col" class="text-left font-semibold pl-gutter pr-3 h-8">File</th>
            {#if showsAll}
              <th scope="col" class="hidden lg:table-cell text-left font-semibold px-3 w-40">Created by</th>
            {/if}
            <th scope="col" class="hidden md:table-cell text-left font-semibold px-3 w-44">Expires</th>
            <th scope="col" class="hidden sm:table-cell text-right font-semibold px-3 w-24">Downloads</th>
            <th scope="col" class="hidden sm:table-cell text-left font-semibold px-3 w-28">Status</th>
            <th scope="col" class="w-28 pr-gutter"><span class="sr-only">Actions</span></th>
          </tr>
        </thead>
        <tbody>
          {#each links as link (link.id)}
            <tr class="border-b border-border-default align-middle">
              <td class="pl-gutter pr-3 py-2">
                <div class="flex items-center gap-3 min-w-0">
                  {#if link.isFolder}
                    <FolderIcon size="16" class="shrink-0 text-fg-secondary" />
                  {:else}
                    <FileIcon size="16" class="shrink-0 text-fg-muted" />
                  {/if}
                  <div class="min-w-0">
                    <p class="truncate text-fg-primary" title={link.name}>{link.name}</p>
                    <p class="truncate text-xs text-fg-muted font-mono" title={fullPath(link)}>{link.loc}</p>
                    <p class="sm:hidden mt-1">
                      <span class="sg-chip {STATUS_TEXT[link.status]}">{STATUS_LABEL[link.status]}</span>
                    </p>
                    <p class="md:hidden text-xs text-fg-muted">
                      Expires {shortTime(link.expiresAt)}<span class="sm:hidden"> · {downloadsText(link.downloadCount)}</span>
                    </p>
                    {#if showsAll}
                      <p class="lg:hidden truncate text-xs text-fg-muted">By {link.creatorName || link.creatorId}</p>
                    {/if}
                  </div>
                </div>
              </td>
              {#if showsAll}
                <td class="hidden lg:table-cell px-3 text-xs text-fg-secondary truncate" title={link.creatorId}>
                  {link.creatorName || link.creatorId}
                </td>
              {/if}
              <td class="hidden md:table-cell px-3 text-xs text-fg-muted whitespace-nowrap">
                {shortTime(link.expiresAt)}
              </td>
              <td class="hidden sm:table-cell px-3 text-xs text-fg-muted text-right tabular-nums">
                {link.downloadCount}
              </td>
              <td class="hidden sm:table-cell px-3">
                <span class="sg-chip {STATUS_TEXT[link.status]}">{STATUS_LABEL[link.status]}</span>
              </td>
              <td class="pr-gutter text-right">
                {#if link.status === "active"}
                  <button
                    type="button"
                    class="sg-btn sg-btn-danger"
                    onclick={() => revoke(link)}
                    disabled={revoking !== null}
                    aria-label="Revoke link to {link.name}"
                  >
                    {revoking === link.id ? "Revoking…" : "Revoke"}
                  </button>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</section>
