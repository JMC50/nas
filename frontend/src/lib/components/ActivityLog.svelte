<script lang="ts">
  import { onMount } from "svelte";
  import History from "lucide-svelte/icons/history";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import ArrowRight from "lucide-svelte/icons/arrow-right";
  import { notifications } from "$lib/store/notifications.svelte";
  import ActivityGraph from "./Activity/ActivityGraph.svelte";

  interface ActivityEntry {
    id?: number;
    activity: string;
    description: string;
    userId?: string;
    krname?: string;
    username?: string;
    time: number;
    loc?: string;
  }

  let entries: ActivityEntry[] = $state([]);
  let loading = $state(true);

  async function load() {
    loading = true;
    try {
      const response = await fetch("/server/getActivityLog");
      const data = await response.json();
      const raw = (data.logs ?? data.data ?? (Array.isArray(data) ? data : [])) as ActivityEntry[];
      entries = raw.sort((a, b) => b.time - a.time);
    } catch (cause) {
      notifications.error(`Failed: ${(cause as Error).message}`);
    } finally {
      loading = false;
    }
  }

  const ACTIVITY_DOT: Record<string, string> = {
    UPLOAD: "bg-fg-success",
    DELETE: "bg-fg-danger",
    DOWNLOAD: "bg-fg-info",
    RENAME: "bg-fg-warning",
    COPY: "bg-fg-link",
    MOVE: "bg-fg-link",
    VIEW: "bg-fg-muted",
    OPEN: "bg-fg-muted",
  };

  const ACTIVITY_TEXT: Record<string, string> = {
    UPLOAD: "text-fg-success",
    DELETE: "text-fg-danger",
    DOWNLOAD: "text-fg-info",
    RENAME: "text-fg-warning",
    COPY: "text-fg-link",
    MOVE: "text-fg-link",
    VIEW: "text-fg-muted",
    OPEN: "text-fg-muted",
  };

  function dotClass(activity: string): string {
    return ACTIVITY_DOT[activity.toUpperCase()] ?? "bg-fg-secondary";
  }

  function textClass(activity: string): string {
    return ACTIVITY_TEXT[activity.toUpperCase()] ?? "text-fg-secondary";
  }

  function timeOfDay(time: number): string {
    const d = new Date(time);
    const h = String(d.getHours()).padStart(2, "0");
    const m = String(d.getMinutes()).padStart(2, "0");
    return `${h}:${m}`;
  }

  function fullTime(time: number): string {
    return new Date(time).toLocaleString();
  }

  function localDayKey(time: number): string {
    const d = new Date(time);
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
  }

  function dayLabel(time: number): string {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const day = new Date(time);
    day.setHours(0, 0, 0, 0);
    const deltaDays = Math.round((today.getTime() - day.getTime()) / (1000 * 60 * 60 * 24));
    if (deltaDays === 0) return "Today";
    if (deltaDays === 1) return "Yesterday";
    return day.toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
  }

  // Group consecutive entries by local YYYY-MM-DD; preserve sort order.
  interface DayGroup {
    key: string;
    label: string;
    iso: string;
    entries: ActivityEntry[];
  }
  const dayGroups = $derived.by<DayGroup[]>(() => {
    const out: DayGroup[] = [];
    for (const entry of entries) {
      const key = localDayKey(entry.time);
      const last = out[out.length - 1];
      if (!last || last.key !== key) {
        out.push({ key, label: dayLabel(entry.time), iso: key, entries: [entry] });
      } else {
        last.entries.push(entry);
      }
    }
    return out;
  });

  // Parse legacy description strings into structured pieces:
  //   "MOVE [FILE] FROM /a/b TO /c/d"       → { kind: "move", from, to, target: "FILE" }
  //   "UPLOAD [FILE] AT /a/b"               → { kind: "simple", path, target: "FILE" }
  //   "CREATE [FOLDER] AT /a/b"             → { kind: "simple", path, target: "FOLDER" }
  //   anything else                          → { kind: "raw", text }
  interface Parsed {
    kind: "move" | "simple" | "raw";
    target?: string;
    from?: string;
    to?: string;
    path?: string;
    text?: string;
  }
  function parse(description: string): Parsed {
    const moveMatch = description.match(/^[A-Z]+ \[([A-Z]+)\] FROM (.+) TO (.+)$/);
    if (moveMatch) {
      return { kind: "move", target: moveMatch[1], from: moveMatch[2], to: moveMatch[3] };
    }
    const simpleMatch = description.match(/^[A-Z]+ \[([A-Z]+)\] AT (.+)$/);
    if (simpleMatch) {
      return { kind: "simple", target: simpleMatch[1], path: simpleMatch[2] };
    }
    return { kind: "raw", text: description };
  }

  function authorInitial(entry: ActivityEntry): string {
    const name = entry.krname || entry.username || entry.userId || "?";
    return name.slice(0, 1).toUpperCase();
  }

  function authorName(entry: ActivityEntry): string {
    return entry.krname || entry.username || entry.userId || "Unknown";
  }

  onMount(load);
</script>

<section class="flex flex-col h-full bg-bg-base overflow-hidden">
  <header class="flex items-center justify-between gap-2 px-6 h-12 border-b border-border-default bg-bg-surface">
    <div class="flex items-center gap-2">
      <History size="18" class="text-accent" />
      <h1 class="text-sm font-semibold text-fg-primary">Activity</h1>
      <span class="text-xs text-fg-muted font-mono ml-2">{entries.length}</span>
    </div>
    <button
      type="button"
      class="inline-flex items-center gap-1.5 h-8 px-2.5 rounded-md text-fg-muted hover:text-fg-primary hover:bg-bg-hover text-xs transition-colors"
      onclick={load}
      aria-label="Refresh"
    >
      <RefreshCw size="13" />
      <span>Refresh</span>
    </button>
  </header>

  {#if !loading}
    <ActivityGraph {entries} />
  {/if}

  <div class="flex-1 overflow-auto">
    {#if loading}
      <div class="p-6 text-sm text-fg-muted">Loading…</div>
    {:else if entries.length === 0}
      <div class="p-12 text-center text-sm text-fg-muted">No activity yet.</div>
    {:else}
      <div class="py-2">
        {#each dayGroups as group (group.key)}
          <div class="sticky top-0 z-10 flex items-baseline gap-3 px-6 h-7 bg-bg-base border-y border-border-default/60">
            <span class="text-[11px] font-semibold text-fg-primary uppercase tracking-wide">
              {group.label}
            </span>
            <span class="text-[10px] font-mono text-fg-muted">{group.iso}</span>
            <span class="ml-auto text-[10px] font-mono text-fg-muted">
              {group.entries.length} {group.entries.length === 1 ? "event" : "events"}
            </span>
          </div>

          <ol>
            {#each group.entries as entry (entry.id ?? `${entry.time}-${entry.activity}`)}
              {@const parsed = parse(entry.description)}
              <li
                class="grid grid-cols-[58px_22px_1fr] items-start gap-2 px-6 py-1.5 group hover:bg-bg-hover/40 transition-colors"
              >
                <!-- Time column -->
                <div
                  class="font-mono text-[11px] text-fg-muted pt-1"
                  title={fullTime(entry.time)}
                >
                  {timeOfDay(entry.time)}
                </div>

                <!-- Dot + author chip column -->
                <div class="flex flex-col items-center pt-1.5 gap-1">
                  <span
                    class="w-2.5 h-2.5 rounded-full {dotClass(entry.activity)} ring-2 ring-bg-base"
                    aria-hidden="true"
                  ></span>
                  <span
                    class="w-5 h-5 rounded-full bg-bg-elevated text-fg-primary text-[10px] font-semibold flex items-center justify-center"
                    title={authorName(entry)}
                  >
                    {authorInitial(entry)}
                  </span>
                </div>

                <!-- Content column -->
                <div class="min-w-0">
                  <div class="flex items-baseline gap-2 flex-wrap">
                    <span
                      class="font-mono text-[10px] uppercase tracking-wide {textClass(entry.activity)}"
                    >
                      {entry.activity}
                    </span>
                    {#if parsed.target}
                      <span class="text-[10px] font-mono text-fg-muted">[{parsed.target}]</span>
                    {/if}
                    <span class="text-xs text-fg-secondary">{authorName(entry)}</span>
                  </div>

                  {#if parsed.kind === "move"}
                    <div class="text-sm flex items-center gap-1.5 flex-wrap mt-0.5">
                      <span class="font-mono text-fg-muted">{parsed.from}</span>
                      <ArrowRight size="12" class="text-fg-muted shrink-0" />
                      <span class="font-mono text-fg-primary">{parsed.to}</span>
                    </div>
                  {:else if parsed.kind === "simple"}
                    <div class="text-sm font-mono text-fg-primary mt-0.5 break-all">{parsed.path}</div>
                  {:else}
                    <div class="text-sm text-fg-primary mt-0.5 break-words">{parsed.text}</div>
                  {/if}
                </div>
              </li>
            {/each}
          </ol>
        {/each}
      </div>
    {/if}
  </div>
</section>
