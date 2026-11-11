<script lang="ts">
  import { onMount } from "svelte";
  import History from "lucide-svelte/icons/history";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import ArrowRight from "lucide-svelte/icons/arrow-right";
  import { notifications } from "$lib/store/notifications.svelte";
  import { auth } from "$lib/store/auth.svelte";
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
      const response = await fetch("/server/getActivityLog", {
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
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
    COPY: "bg-fg-purple",
    MOVE: "bg-fg-purple",
    VIEW: "bg-fg-muted",
    OPEN: "bg-fg-muted",
  };

  const ACTIVITY_TEXT: Record<string, string> = {
    UPLOAD: "text-fg-success",
    DELETE: "text-fg-danger",
    DOWNLOAD: "text-fg-info",
    RENAME: "text-fg-warning",
    COPY: "text-fg-purple",
    MOVE: "text-fg-purple",
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
  <header class="sg-page-head">
    <History size="18" class="text-fg-secondary" />
    <h1 class="sg-page-title">Activity</h1>
    {#if !loading}
      <span class="sg-count">{entries.length}</span>
    {/if}
    <button type="button" class="ml-auto sg-btn sg-btn-ghost" onclick={load} aria-label="Refresh">
      <RefreshCw size="16" />
      <span class="hidden sm:inline">Refresh</span>
    </button>
  </header>

  <ActivityGraph {entries} {loading} />

  <div class="flex-1 overflow-auto">
    {#if loading}
      <div class="px-gutter py-6 text-sm text-fg-muted" role="status">Loading…</div>
    {:else if entries.length === 0}
      <div class="px-gutter py-16 text-center text-sm text-fg-muted">No activity yet.</div>
    {:else}
      <div class="pb-4">
        {#each dayGroups as group (group.key)}
          <div class="sticky top-0 z-10 flex items-baseline gap-3 px-gutter h-9 bg-bg-surface border-y border-border-default">
            <h2 class="self-center text-sm font-semibold text-fg-primary">
              {group.label}
            </h2>
            <span class="self-center text-xs text-fg-muted tabular-nums">{group.iso}</span>
            <span class="self-center ml-auto text-xs text-fg-muted tabular-nums">
              {group.entries.length} {group.entries.length === 1 ? "event" : "events"}
            </span>
          </div>

          <ol>
            <!-- 서버가 id를 주지 않고 같은 ms에 같은 종류 기록이 여러 개 생길 수 있어 순번을 붙임 -->
            {#each group.entries as entry, index (entry.id ?? `${entry.time}-${entry.activity}-${index}`)}
              {@const parsed = parse(entry.description)}
              <li
                class="grid grid-cols-[48px_24px_minmax(0,1fr)] items-start gap-3 px-gutter py-2 border-b border-border-default hover:bg-bg-hover transition-colors"
              >
                <!-- Time column -->
                <div
                  class="text-xs text-fg-muted tabular-nums pt-0.5"
                  title={fullTime(entry.time)}
                >
                  {timeOfDay(entry.time)}
                </div>

                <!-- Dot + author chip column -->
                <div class="flex flex-col items-center gap-1.5 pt-0.5">
                  <span
                    class="size-2 rounded-full {dotClass(entry.activity)}"
                    aria-hidden="true"
                  ></span>
                  <span
                    class="grid place-items-center size-6 rounded-control bg-bg-elevated text-fg-secondary text-xs font-bold"
                    title={authorName(entry)}
                    aria-hidden="true"
                  >
                    {authorInitial(entry)}
                  </span>
                </div>

                <!-- Content column -->
                <div class="min-w-0">
                  <div class="flex items-baseline gap-2 flex-wrap">
                    <span class="sg-chip font-mono {textClass(entry.activity)}">
                      {entry.activity}
                    </span>
                    {#if parsed.target}
                      <span class="text-xs font-mono text-fg-muted">[{parsed.target}]</span>
                    {/if}
                    <span class="text-xs text-fg-secondary">{authorName(entry)}</span>
                  </div>

                  {#if parsed.kind === "move"}
                    <div class="text-sm flex items-center gap-1.5 flex-wrap mt-1">
                      <span class="font-mono text-fg-muted">{parsed.from}</span>
                      <ArrowRight size="14" class="text-fg-muted shrink-0" />
                      <span class="font-mono text-fg-primary">{parsed.to}</span>
                    </div>
                  {:else if parsed.kind === "simple"}
                    <div class="text-sm font-mono text-fg-primary mt-1 break-all">{parsed.path}</div>
                  {:else}
                    <div class="text-sm text-fg-primary mt-1">{parsed.text}</div>
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
