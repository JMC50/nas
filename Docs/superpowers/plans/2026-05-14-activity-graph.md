# Activity Graph View — Implementation Plan

> **For agentic workers:** Use `superpowers:subagent-driven-development` or `superpowers:executing-plans`. Steps use `- [ ]` syntax.

**Goal:** Add an Activity-page graph header showing events-per-day (bar/area chart) and activity-type distribution (compact horizontal stack), inline with the existing timeline list.

**Architecture:** Reuse [Sparkline.svelte](frontend/src/lib/components/System/Sparkline.svelte) as-is. Confirmed prop signature: `{ points: number[]; color: string; height?: number; max?: number }` where `color` is a CSS color string (used as SVG `stroke` and gradient stop, e.g. `"#fabd2f"`) and `max` is the y-axis ceiling (default 100 — must be overridden for event counts to avoid clipping). No generalization required. Add a small inline activity-graph banner above the existing timeline in [ActivityLog.svelte](frontend/src/lib/components/ActivityLog.svelte). Aggregation runs client-side on the already-loaded `entries` array — no backend change.

**Tech Stack:** Svelte 5 runes, Tailwind 4, lucide-svelte. No charting library.

---

## Context

- [frontend/src/lib/components/ActivityLog.svelte](frontend/src/lib/components/ActivityLog.svelte) — current timeline; loads `/server/getActivityLog`, sorts by `time`, renders list. **No chart yet.**
- [frontend/src/lib/components/System/Sparkline.svelte](frontend/src/lib/components/System/Sparkline.svelte) — existing sparkline, used by [MetricCard.svelte](frontend/src/lib/components/System/MetricCard.svelte).
- ActivityEntry fields available: `id?, activity, description, userId?, krname?, username?, time, loc?`.
- Activity types (from ACTIVITY_DOT map): UPLOAD, DELETE, DOWNLOAD, RENAME, COPY, MOVE, VIEW, OPEN. Color tokens: `bg-fg-success/danger/info/warning/link/muted`.

---

## Phase 1 — (resolved during planning; Sparkline already generic)

Sparkline.svelte is already prop-driven (`points`, `color`, `height`, `max`). MetricCard.svelte passes hex colors. No code change needed. Skip to Phase 2.

---

## Phase 2 — Aggregation helper

### Task 2.1: Create aggregation module

**Files:**
- Create: `frontend/src/lib/components/Activity/aggregate.ts`

- [ ] **Step 1** — Export `interface DayBucket { day: string; count: number }` (`day` = local YYYY-MM-DD).
- [ ] **Step 2** — Export `function dailyCounts(entries, days = 30): DayBucket[]` returning N most recent days (oldest → newest), zero-filled for days with no events.
- [ ] **Step 3** — Export `interface TypeShare { type: string; count: number; ratio: number }`.
- [ ] **Step 4** — Export `function typeDistribution(entries): TypeShare[]` sorted by count descending. Folds "OPEN"+"VIEW" into "VIEW" for a cleaner display, documented in inline note.
- [ ] **Verify:** Walk through table: entries with times spanning 5 days, 3 events on day 1, 0 on day 2, etc. → bucket counts match.

### Task 2.2: Commit Phase 2

- [ ] `git commit -m "[feat] activity log aggregation helpers (daily count, type share)"`

---

## Phase 3 — Render graph banner

### Task 3.1: Activity graph component

**Files:**
- Create: `frontend/src/lib/components/Activity/ActivityGraph.svelte`

- [ ] **Step 1** — Props `{ entries: ActivityEntry[] }`.
- [ ] **Step 2** — Compute `const daily = $derived(dailyCounts(entries, 30));` and `const dist = $derived(typeDistribution(entries));`.
- [ ] **Step 3** — Top row, two-column flex layout (gap-4, h-20):
  - Left (flex-1): "Last 30 days" small muted label, then:
    ```svelte
    {@const counts = daily.map(d => d.count)}
    {@const ceiling = Math.max(1, ...counts)}
    <Sparkline points={counts} color="#fabd2f" height={48} max={ceiling} />
    ```
    Below: tiny x-axis with first-day and today date labels (`text-[10px] font-mono text-fg-muted`). Gruvbox accent yellow `#fabd2f` (matches CSS token `--color-accent`).
  - Right (w-72): "Distribution" label + a horizontal stacked bar (h-2 rounded-full) of segments colored by `ACTIVITY_DOT` lookup, widths proportional to `ratio`. Legend below as inline pills `<dot> <type> <count>` per type.
- [ ] **Step 4** — When `entries.length === 0`, render skeleton placeholders, not real chart.
- [ ] **Verify:** `npm run check` → 0/0; visual inspection via Playwright.

### Task 3.2: Mount in ActivityLog

**Files:**
- Modify: `frontend/src/lib/components/ActivityLog.svelte`

- [ ] **Step 1** — Below the `<header>`, above the timeline `<ol>`, render `<ActivityGraph {entries} />`.
- [ ] **Step 2** — When `loading` is true, do not render graph (it expects loaded entries).
- [ ] **Verify:** `npm run check` → 0/0.

### Task 3.3: Commit Phase 3

- [ ] `git commit -m "[feat] activity page graph banner (daily count sparkline + type distribution)"`

---

## Phase 4 — /code-review + Playwright + merge

### Task 4.1: code-review loop

- [ ] Invoke `code-review` skill on Activity*.svelte + aggregate.ts. Auto-fix all ❌ Critical.

### Task 4.2: Playwright walkthrough

| # | Action | Expected |
|---|---|---|
| 1 | Open Activity tab with non-empty log | Banner shows sparkline + distribution bar + legend |
| 2 | Hover/inspect: distribution segments | Widths sum to ≈100% |
| 3 | Filter scenario: artificially seed entries spanning 5 days | Sparkline shows 5 non-zero days, rest zero; max-bar peak reaches top of svg (ceiling = local max) |
| 4 | Reload empty Activity (no entries) | Skeleton placeholders; no NaN/Infinity in DOM |
| 5 | Theme toggle dark/light | Banner colors swap consistently with rest of UI |
| 6 | Tab switch to Files and back | Graph re-renders without flicker (entries cached or refetched cleanly) |
| 7 | Seed entries with OPEN and VIEW activities | Distribution legend shows merged "VIEW" pill with combined count (documented fold) |
| 8 | Hover legend pill | Title attr shows full count and percentage |

### Task 4.3: Pre-merge gate + merge

- [ ] RULE 6 checklist; merge `feat/activity-graph` → main; push origin.

---

## Risks & open questions

| Risk | Mitigation |
|---|---|
| Sparkline.svelte not actually generic | Phase 1 audits and generalizes; small touch, verify MetricCard unchanged |
| Activity log is enormous (10k+ entries) → aggregation O(N) | Acceptable for in-memory; if measured slow, plan adds memo of last-aggregate-input-hash. Out of scope until measured. |
| Type distribution color contrast in Gruvbox light theme | All colors use existing `fg-success/danger/...` tokens that already pass contrast check elsewhere |
| OPEN vs VIEW conflation | Documented inline; reviewer may push back. Easy revert: remove fold. |
| Reactivity loops if `$derived` is read inside `$effect` carelessly | Components only `$derive` from props; no `$effect` writes to derivations |

---

## Rollback strategy

```bash
cd C:/Data/Git/ANXI/nas && git revert -m 1 <merge-sha> && git push origin main
```

Sparkline.svelte is untouched, so revert is purely additive — only ActivityLog import and new components disappear.

---

## 5-pass self-review

| Pass | Concern | Action |
|---|---|---|
| 1. Structure | Audit → generalize (conditional) → aggregator → component → mount. Each step has verify. | ✅ |
| 2. Spec correctness | `ActivityLog.svelte` import already includes `History`, `RefreshCw` from lucide; `ACTIVITY_DOT` confirmed at lines 35-44; `Sparkline.svelte` path confirmed via grep result | ✅ |
| 3. Risk | Big logs, theme contrast, OPEN/VIEW fold, reactivity — each mitigated | ✅ |
| 4. Consistency | Reuses existing tokens (`bg-fg-success` etc.) and Sparkline; compact monospace small text; matches NAS VSCode tone | ✅ |
| 5. Completeness | Acceptance: graph exists and works ✅, events-per-day ✅, type distribution ✅. Rollback concrete. | ✅ |
