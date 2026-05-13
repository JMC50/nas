# Professional Media Player UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the bare HTML5 `<video controls>` / `<audio controls>` in NAS file viewer with a polished, professional custom media player UI that matches the app's Gruvbox-dark design language.

**Architecture:** Two dedicated player components — `VideoPlayer.svelte` for video, `AudioPlayer.svelte` for audio — sharing a tiny pure-function helpers module (`media-utils.ts`). `MediaViewer.svelte` becomes a thin dispatcher choosing the right player by `kind` prop. Each player owns its `HTMLMediaElement` directly via `bind:this` and exposes a custom controls bar plus keyboard shortcuts. No global state — everything lives in the player component instance using Svelte 5 runes.

**Tech Stack:** Svelte 5 runes (`$state` / `$derived` / `$effect` / `bind:this`), Tailwind 4 with project's existing Gruvbox-dark theme tokens (`bg-bg-overlay`, `bg-accent`, `text-fg-accent` etc.), lucide-svelte icons (already used elsewhere in app), FiraD2 font (already loaded via `@font-face` in `app.css`), HTMLMediaElement browser API, native Fullscreen + Picture-in-Picture APIs. **No new npm dependencies.**

**Visual Design Spec:**
- Gruvbox-dark base, yellow `#fabd2f` (`text-fg-accent` / `bg-accent`) as primary accent
- Bottom controls bar: 56px height, `bg-bg-overlay/80` with `backdrop-blur-sm`, auto-fade after 2.5s idle, instant reveal on mousemove/focus
- Scrubber: 4px tall normally, expands to 6px on hover with thumb circle (10px, `bg-accent`)
- Buffered ranges: `bg-bg-selected` overlaid below played portion
- Progress (played): `bg-accent` yellow
- Buttons: 36px square icon-only, hover: `bg-bg-hover`, active: `text-fg-accent`, focus-visible: 2px `border-border-focus` outline
- Time displays: `font-mono` (FiraD2), `text-xs` (12px), `text-fg-secondary`
- Audio layout: centered card, 360px max width, large 96px circular play button with `bg-accent text-accent-fg`, animated 5-bar equalizer (pure CSS keyframes, no Web Audio API) — bars only animate while playing
- Transitions: 200ms with `ease-smooth` (`cubic-bezier(0.16, 1, 0.3, 1)`) custom token

**Testing Approach:** Frontend has **no test framework** — `package.json` scripts are dev/build/check only. Per project rule #1 (no fake verifier boilerplate), each task uses a **spec table + manual verification** in the dev server. Per-task gates: `npm run check` (svelte-check 0 errors). Per-implementation gate: `/code-review` (0 ❌ Critical). Final gate: `npm run build` + manual visual walkthrough of every control + keyboard shortcut against the spec table in Task 15.

**Out of scope (separate plans):**
- Markdown render toggle (view/edit) — separate plan
- PDF.js integration — separate plan
- Office documents via LibreOffice backend — separate plan
- Server-side video thumbnail generation for scrubber hover preview — YAGNI v1
- Web Audio API waveform parsing (decodeAudioData on full streamed file) — YAGNI v1 (the CSS equalizer is the "professional" placeholder; real waveform is a v2 enhancement)
- Subtitle / caption support — not requested
- Mobile gesture controls — desktop-first; mobile gets sane defaults but no swipe-to-seek
- Playlist / queue — files open as individual tabs; no playlist concept

---

## File Structure

| File | Status | Responsibility | Target Lines |
|---|---|---|---|
| `frontend/src/lib/components/Viewers/media-utils.ts` | Create | Pure helpers: `formatTime`, `clampVolume`, `pickMediaKey`, `MediaKey` type | ≤80 |
| `frontend/src/lib/components/Viewers/VideoPlayer.svelte` | Create | Video viewport + custom controls bar + keyboard + auto-hide overlay | ≤300 |
| `frontend/src/lib/components/Viewers/AudioPlayer.svelte` | Create | Centered audio UI: big play button, equalizer, scrubber, volume | ≤220 |
| `frontend/src/lib/components/Viewers/MediaViewer.svelte` | Modify (replace body) | Dispatcher: pick VideoPlayer or AudioPlayer by `kind` prop | ≤30 |

**Why this split:**
- `media-utils.ts` is pure logic → testable by spec table without DOM.
- Video and audio have fundamentally different layouts (viewport vs centered card) and different controls (PiP/fullscreen only apply to video). Forcing one component to handle both would be a ⚠️ SOLID/SRP violation. Two siblings is the cleaner split.
- Both players use the same media URL pattern + share the same key-handling logic via `media-utils.ts`. DRY satisfied without forcing premature abstraction of UI.
- `MediaViewer.svelte` stays thin (~30 lines) as a discriminating dispatcher — matches the existing pattern in `TabContent.svelte`.

**Threshold:** If `VideoPlayer.svelte` exceeds 300 lines during implementation, extract `VideoControls.svelte` (the bottom bar) as a sub-component. Plan currently assumes single file; revisit after Task 9.

**No backend changes.** Existing `/server/getVideoData` and `/server/getAudioData` already serve via `http.ServeContent` (HTTP 206 Range supported) — perfect for `<video>` / `<audio>` streaming.

---

## Task 0: Branch setup

**Files:** None (git only).

- [ ] **Step 1: Create and check out the feature branch**

```bash
git checkout main
git pull
git checkout -b feat/media-viewers
```

Expected: `Switched to a new branch 'feat/media-viewers'`. All subsequent tasks commit to this branch.

---

## Task 1: `media-utils.ts` — pure helpers

**Files:**
- Create: `frontend/src/lib/components/Viewers/media-utils.ts`

**Spec table:**

| Function | Input | Output |
|---|---|---|
| `formatTime` | `0` | `"0:00"` |
| `formatTime` | `5` | `"0:05"` |
| `formatTime` | `65` | `"1:05"` |
| `formatTime` | `3725` | `"1:02:05"` |
| `formatTime` | `NaN` / `Infinity` / `-1` | `"0:00"` |
| `clampVolume` | `-0.5` | `0` |
| `clampVolume` | `1.5` | `1` |
| `clampVolume` | `0.7` | `0.7` |
| `clampVolume` | `NaN` | `0` |
| `pickMediaKey` | `KeyboardEvent({code:"Space"})` | `"toggle"` |
| `pickMediaKey` | `KeyboardEvent({code:"KeyK"})` | `"toggle"` |
| `pickMediaKey` | `KeyboardEvent({code:"KeyM"})` | `"mute"` |
| `pickMediaKey` | `KeyboardEvent({code:"KeyF"})` | `"fullscreen"` |
| `pickMediaKey` | `KeyboardEvent({code:"ArrowLeft"})` | `"seekBack"` |
| `pickMediaKey` | `KeyboardEvent({code:"ArrowRight"})` | `"seekForward"` |
| `pickMediaKey` | `KeyboardEvent({code:"ArrowUp"})` | `"volumeUp"` |
| `pickMediaKey` | `KeyboardEvent({code:"ArrowDown"})` | `"volumeDown"` |
| `pickMediaKey` | `KeyboardEvent({code:"KeyZ"})` | `null` |

- [ ] **Step 1: Write the module**

```ts
// frontend/src/lib/components/Viewers/media-utils.ts

export type MediaKey =
  | "toggle"
  | "mute"
  | "fullscreen"
  | "seekBack"
  | "seekForward"
  | "volumeUp"
  | "volumeDown";

const SECONDS_PER_MINUTE = 60;
const SECONDS_PER_HOUR = 3600;

export function formatTime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return "0:00";
  const total = Math.floor(seconds);
  const h = Math.floor(total / SECONDS_PER_HOUR);
  const m = Math.floor((total % SECONDS_PER_HOUR) / SECONDS_PER_MINUTE);
  const s = total % SECONDS_PER_MINUTE;
  const pad = (n: number) => n.toString().padStart(2, "0");
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${m}:${pad(s)}`;
}

export function clampVolume(value: number): number {
  if (!Number.isFinite(value)) return 0;
  return Math.min(1, Math.max(0, value));
}

const KEY_MAP: Record<string, MediaKey> = {
  Space: "toggle",
  KeyK: "toggle",
  KeyM: "mute",
  KeyF: "fullscreen",
  ArrowLeft: "seekBack",
  ArrowRight: "seekForward",
  ArrowUp: "volumeUp",
  ArrowDown: "volumeDown",
};

export function pickMediaKey(event: KeyboardEvent): MediaKey | null {
  return KEY_MAP[event.code] ?? null;
}
```

- [ ] **Step 2: Verify the spec table** — walk through each row mentally against the implementation. Pure functions with deterministic output; the spec table itself IS the verification artifact. No mock DOM needed.

- [ ] **Step 3: Type check**

```bash
cd frontend && npm run check
```

Expected: `svelte-check found 0 errors and 0 warnings`.

- [ ] **Step 4: `/code-review` on `media-utils.ts`**

Invoke `code-review` skill. Expected: 14/14 criteria pass, 0 ❌ Critical.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/components/Viewers/media-utils.ts
git commit -m "[feat] add media-utils helpers (formatTime, clampVolume, pickMediaKey)"
```

---

## Task 2: `VideoPlayer.svelte` — skeleton, viewport, file label

**Files:**
- Create: `frontend/src/lib/components/Viewers/VideoPlayer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Component renders with `loc`/`name` props | Filename shown in top-right corner of viewport |
| `<video>` element bound to media URL | Source URL = `/server/getVideoData?token=...&loc=...&name=...` |
| Native browser controls | Hidden (`controls` attribute NOT set) |
| Container | Fills parent (`h-full w-full`), `bg-black` |
| Viewport | Centered video, `object-contain`, max width/height of container |
| Top bar | 32px tall, semi-transparent gradient (`bg-gradient-to-b from-bg-overlay/80 to-transparent`), shows filename, fades on `controls-hidden` state (controlled in Task 9) |

- [ ] **Step 1: Write the component skeleton**

```svelte
<!-- frontend/src/lib/components/Viewers/VideoPlayer.svelte -->
<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  let video: HTMLVideoElement | null = $state(null);

  const mediaUrl = $derived(
    `/server/getVideoData?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );
</script>

<div class="relative h-full w-full bg-black overflow-hidden group">
  <!-- top bar (filename) -->
  <div
    class="absolute top-0 left-0 right-0 h-9 px-4 flex items-center
           bg-gradient-to-b from-bg-overlay/80 to-transparent
           text-fg-secondary text-xs font-mono z-10 pointer-events-none"
  >
    <span class="truncate">{name}</span>
  </div>

  <!-- video element -->
  <!-- svelte-ignore a11y_media_has_caption -->
  <video
    bind:this={video}
    src={mediaUrl}
    class="w-full h-full object-contain"
    preload="metadata"
  ></video>
</div>
```

- [ ] **Step 2: Wire into `MediaViewer.svelte` temporarily for visual check**

Edit `frontend/src/lib/components/Viewers/MediaViewer.svelte` to route `kind === "video"` to the new `VideoPlayer`. Audio path stays unchanged for now.

```svelte
<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";
  import VideoPlayer from "./VideoPlayer.svelte";

  interface Props {
    loc: string;
    name: string;
    kind: "video" | "audio";
  }

  let { loc, name, kind }: Props = $props();

  // audio path unchanged for now
  const audioUrl = $derived(
    `/server/getAudioData?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );
</script>

{#if kind === "video"}
  <VideoPlayer {loc} {name} />
{:else}
  <div class="flex flex-col h-full w-full bg-bg-base">
    <div class="flex items-center px-3 h-9 border-b border-border-default text-xs text-fg-secondary">
      <span class="truncate">{name}</span>
    </div>
    <div class="flex-1 flex items-center justify-center p-4">
      <audio src={audioUrl} controls class="w-full max-w-xl"></audio>
    </div>
  </div>
{/if}
```

- [ ] **Step 3: Type check**

```bash
cd frontend && npm run check
```

Expected: 0 errors.

- [ ] **Step 4: Manual visual verification**

```bash
cd frontend && npm run dev
```

Open the app, navigate to a video file (`.mp4`/`.webm`), double-click. Expected:
- Video shows in viewport, fits container
- Filename visible in top-left
- **No native controls** (no play button at bottom yet — that's the next task)
- Top bar gradient looks like a subtle fade

(Video won't be playable yet because there's no play button. That's expected.)

- [ ] **Step 5: `/code-review` on `VideoPlayer.svelte` + the MediaViewer modification**

Expected: 0 ❌ Critical. ⚠️ "no play button yet" is expected — it's a deliberate WIP state, not a violation.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/components/Viewers/VideoPlayer.svelte frontend/src/lib/components/Viewers/MediaViewer.svelte
git commit -m "[feat] add VideoPlayer skeleton with custom viewport and filename header"
```

---

## Task 3: Play/Pause — button + viewport tap + state binding

**Files:**
- Modify: `frontend/src/lib/components/Viewers/VideoPlayer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Click viewport | Toggles play/pause |
| Click controls bar play button | Toggles play/pause |
| Button icon | `Play` when paused, `Pause` when playing (lucide-svelte) |
| Center overlay | When paused: large translucent play button (96px) in center; click also toggles |
| When playing | Center overlay hidden |

- [ ] **Step 1: Add state + control logic**

Inside `<script>` of `VideoPlayer.svelte`, below the existing declarations:

```ts
let playing = $state(false);

function toggle() {
  if (!video) return;
  if (video.paused) video.play().catch(() => {});
  else video.pause();
}
```

Bind `playing` to video element's playing state by adding `onplay` / `onpause` handlers:

```ts
function onPlay() { playing = true; }
function onPause() { playing = false; }
```

- [ ] **Step 2: Update markup — viewport tap + center overlay + controls bar with play button**

Replace the existing `<video>` block and add the controls bar:

```svelte
<!-- imports at top of script -->
<script lang="ts">
  import Play from "lucide-svelte/icons/play";
  import Pause from "lucide-svelte/icons/pause";
  // ... existing imports
</script>

<!-- replace video element with: -->
<!-- svelte-ignore a11y_media_has_caption -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<video
  bind:this={video}
  src={mediaUrl}
  class="w-full h-full object-contain cursor-pointer"
  preload="metadata"
  onclick={toggle}
  onplay={onPlay}
  onpause={onPause}
></video>

<!-- center overlay (visible when paused) -->
{#if !playing}
  <button
    type="button"
    onclick={toggle}
    aria-label="Play"
    class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2
           w-24 h-24 rounded-full bg-bg-overlay/70 backdrop-blur-sm
           text-fg-primary flex items-center justify-center
           hover:bg-bg-overlay/90 hover:text-fg-accent transition-colors z-20"
  >
    <Play size="40" fill="currentColor" />
  </button>
{/if}

<!-- bottom controls bar (placeholder — extends in later tasks) -->
<div
  class="absolute bottom-0 left-0 right-0 h-14 px-4 flex items-center gap-2
         bg-bg-overlay/80 backdrop-blur-sm
         text-fg-primary z-10"
>
  <button
    type="button"
    onclick={toggle}
    aria-label={playing ? "Pause" : "Play"}
    class="inline-flex items-center justify-center w-9 h-9 rounded
           hover:bg-bg-hover hover:text-fg-accent transition-colors"
  >
    {#if playing}
      <Pause size="18" />
    {:else}
      <Play size="18" />
    {/if}
  </button>
</div>
```

- [ ] **Step 3: Type check**

```bash
cd frontend && npm run check
```

Expected: 0 errors.

- [ ] **Step 4: Manual visual verification**

In dev server, open a video. Expected:
- Big play button in center when first loaded
- Click center → video plays, center button disappears
- Click viewport again → pauses, center button reappears
- Bottom-left has a smaller play/pause icon button that mirrors the state
- Hover bottom button → `bg-bg-hover` and icon color becomes yellow (`text-fg-accent`)

- [ ] **Step 5: `/code-review`**

Expected: 0 ❌ Critical.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/components/Viewers/VideoPlayer.svelte
git commit -m "[feat] add play/pause control with center overlay and tap-to-toggle"
```

---

## Task 4: Scrubber + time display + buffered ranges

**Files:**
- Modify: `frontend/src/lib/components/Viewers/VideoPlayer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Scrubber position | Reflects `video.currentTime / video.duration` |
| Click scrubber | Seeks to clicked position |
| Drag scrubber | Live preview seek while dragging |
| Buffered range | Shows lighter overlay from start to most-recent buffered TimeRange.end |
| Time display | `currentTime / duration` formatted via `formatTime`, monospace |
| Scrubber height | 4px default, 6px on hover (CSS only via `group-hover` or local hover state) |
| Thumb circle | 10px, visible on hover, follows progress |

- [ ] **Step 1: Add state + handlers**

In `<script>`:

```ts
import { formatTime } from "./media-utils";

let currentTime = $state(0);
let duration = $state(0);
let buffered = $state(0); // most recent buffered end

function onTimeUpdate() {
  if (!video) return;
  currentTime = video.currentTime;
}
function onLoadedMetadata() {
  if (!video) return;
  duration = video.duration;
}
function onProgress() {
  if (!video) return;
  const ranges = video.buffered;
  buffered = ranges.length > 0 ? ranges.end(ranges.length - 1) : 0;
}

function seek(event: MouseEvent) {
  if (!video || duration === 0) return;
  const bar = event.currentTarget as HTMLElement;
  const rect = bar.getBoundingClientRect();
  const ratio = Math.max(0, Math.min(1, (event.clientX - rect.left) / rect.width));
  video.currentTime = ratio * duration;
  currentTime = video.currentTime;
}

const playedRatio = $derived(duration > 0 ? currentTime / duration : 0);
const bufferedRatio = $derived(duration > 0 ? buffered / duration : 0);
```

- [ ] **Step 2: Add event listeners to `<video>` element**

```svelte
<video
  bind:this={video}
  src={mediaUrl}
  class="..."
  preload="metadata"
  onclick={toggle}
  onplay={onPlay}
  onpause={onPause}
  ontimeupdate={onTimeUpdate}
  onloadedmetadata={onLoadedMetadata}
  onprogress={onProgress}
></video>
```

- [ ] **Step 3: Add scrubber + time display to controls bar**

Replace the controls bar block with the play button + scrubber + time:

```svelte
<div
  class="absolute bottom-0 left-0 right-0 px-4 pb-2 pt-3 flex flex-col gap-1
         bg-bg-overlay/80 backdrop-blur-sm
         text-fg-primary z-10"
>
  <!-- scrubber -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    onclick={seek}
    class="relative h-1.5 rounded-full bg-bg-hover cursor-pointer group/scrub
           hover:h-2 transition-all"
  >
    <!-- buffered -->
    <div
      class="absolute top-0 left-0 h-full rounded-full bg-bg-selected"
      style="width: {bufferedRatio * 100}%"
    ></div>
    <!-- played -->
    <div
      class="absolute top-0 left-0 h-full rounded-full bg-accent"
      style="width: {playedRatio * 100}%"
    ></div>
    <!-- thumb -->
    <div
      class="absolute top-1/2 -translate-x-1/2 -translate-y-1/2 w-2.5 h-2.5 rounded-full bg-accent
             opacity-0 group-hover/scrub:opacity-100 transition-opacity"
      style="left: {playedRatio * 100}%"
    ></div>
  </div>

  <!-- controls row -->
  <div class="flex items-center gap-2 h-9">
    <button
      type="button"
      onclick={toggle}
      aria-label={playing ? "Pause" : "Play"}
      class="inline-flex items-center justify-center w-9 h-9 rounded
             hover:bg-bg-hover hover:text-fg-accent transition-colors"
    >
      {#if playing}
        <Pause size="18" />
      {:else}
        <Play size="18" />
      {/if}
    </button>

    <span class="font-mono text-xs text-fg-secondary tabular-nums">
      {formatTime(currentTime)} / {formatTime(duration)}
    </span>

    <!-- spacer; later tasks fill the right side -->
    <div class="ml-auto"></div>
  </div>
</div>
```

- [ ] **Step 4: Type check**

```bash
cd frontend && npm run check
```

Expected: 0 errors.

- [ ] **Step 5: Manual visual verification**

In dev server, play a video. Expected:
- Scrubber fills with yellow `#fabd2f` as video plays
- Lighter shade ahead of the yellow shows buffered region
- Time display updates every ~100ms with `0:00 / 1:23` format
- Hover scrubber → grows from 6px to 8px, thumb circle appears at progress position
- Click scrubber at 50% → video seeks to middle, current time jumps

- [ ] **Step 6: `/code-review`**

Expected: 0 ❌ Critical.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/lib/components/Viewers/VideoPlayer.svelte
git commit -m "[feat] add video scrubber with buffered ranges and time display"
```

---

## Task 5: Volume + mute

**Files:**
- Modify: `frontend/src/lib/components/Viewers/VideoPlayer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Volume button | Icon: `Volume2` (full), `Volume1` (low), `VolumeX` (muted/0) |
| Click volume button | Toggles mute |
| Hover volume button | Slider slides out horizontally (80px wide, animates in) |
| Volume slider | Click sets volume 0..1; visual is a horizontal bar with thumb |
| Volume persists | Reflects `video.muted` and `video.volume` state |
| Mute restores | Unmuting returns to pre-mute volume |

- [ ] **Step 1: Add state + handlers**

In `<script>`:

```ts
import Volume2 from "lucide-svelte/icons/volume-2";
import Volume1 from "lucide-svelte/icons/volume-1";
import VolumeX from "lucide-svelte/icons/volume-x";
import { clampVolume } from "./media-utils";

let volume = $state(1);
let muted = $state(false);

function onVolumeChange() {
  if (!video) return;
  volume = video.volume;
  muted = video.muted;
}

function toggleMute() {
  if (!video) return;
  video.muted = !video.muted;
}

function setVolume(event: MouseEvent) {
  if (!video) return;
  const bar = event.currentTarget as HTMLElement;
  const rect = bar.getBoundingClientRect();
  const ratio = clampVolume((event.clientX - rect.left) / rect.width);
  video.volume = ratio;
  if (ratio > 0) video.muted = false;
}

// Svelte 5: capitalize the variable name to render it directly as <VolumeIcon />.
// Avoids the deprecated <svelte:component> syntax (which warns under svelte-check).
const VolumeIcon = $derived(
  muted || volume === 0 ? VolumeX : volume < 0.5 ? Volume1 : Volume2,
);
```

Add `onvolumechange={onVolumeChange}` to the `<video>` element.

- [ ] **Step 2: Add volume control to the controls row** (after time display, before the spacer):

```svelte
<!-- volume control (button + slide-out slider) -->
<div class="group/vol flex items-center">
  <button
    type="button"
    onclick={toggleMute}
    aria-label={muted ? "Unmute" : "Mute"}
    class="inline-flex items-center justify-center w-9 h-9 rounded
           hover:bg-bg-hover hover:text-fg-accent transition-colors"
  >
    <VolumeIcon size="18" />
  </button>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    onclick={setVolume}
    class="relative h-1.5 rounded-full bg-bg-hover cursor-pointer
           w-0 group-hover/vol:w-20 ml-0 group-hover/vol:ml-2
           transition-all overflow-hidden"
  >
    <div
      class="absolute top-0 left-0 h-full rounded-full bg-fg-secondary"
      style="width: {muted ? 0 : volume * 100}%"
    ></div>
  </div>
</div>
```

Replace `<!-- spacer; later tasks fill the right side -->` with the new structure (volume sits left of the spacer).

- [ ] **Step 3: Type check**

```bash
cd frontend && npm run check
```

Expected: 0 errors.

- [ ] **Step 4: Manual visual verification**

In dev server, play a video. Expected:
- Volume icon visible (Volume2 when loaded fresh, volume=1)
- Hover the volume icon → slider slides out to the right, 80px wide
- Click 50% on the slider → audio volume halves, icon changes to Volume1
- Click 0% on the slider → muted, icon becomes VolumeX
- Click volume button (not slider) → toggles mute; icon swaps; previous volume restored on unmute

- [ ] **Step 5: `/code-review`**

Expected: 0 ❌ Critical.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/components/Viewers/VideoPlayer.svelte
git commit -m "[feat] add volume control with slide-out slider and mute toggle"
```

---

## Task 6: Fullscreen + Picture-in-Picture

**Files:**
- Modify: `frontend/src/lib/components/Viewers/VideoPlayer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Fullscreen button | Toggles fullscreen on the container element |
| Fullscreen state | Reflected by `Maximize`/`Minimize` icon swap |
| PiP button | Toggles `video.requestPictureInPicture()` |
| PiP button hidden | If `document.pictureInPictureEnabled === false` |
| Both buttons | On the right side of the controls bar |

- [ ] **Step 1: Add state + handlers**

In `<script>`:

```ts
import Maximize from "lucide-svelte/icons/maximize";
import Minimize from "lucide-svelte/icons/minimize";
import PictureInPicture2 from "lucide-svelte/icons/picture-in-picture-2";

let container: HTMLDivElement | null = $state(null);
let isFullscreen = $state(false);

function toggleFullscreen() {
  if (!container) return;
  if (document.fullscreenElement === container) {
    document.exitFullscreen().catch(() => {});
  } else {
    container.requestFullscreen().catch(() => {});
  }
}

function onFullscreenChange() {
  isFullscreen = document.fullscreenElement === container;
}

async function togglePip() {
  if (!video) return;
  try {
    if (document.pictureInPictureElement === video) {
      await document.exitPictureInPicture();
    } else {
      await video.requestPictureInPicture();
    }
  } catch {
    /* user dismissed or browser blocked — no-op */
  }
}

const pipSupported = $derived(
  typeof document !== "undefined" && document.pictureInPictureEnabled,
);

$effect(() => {
  document.addEventListener("fullscreenchange", onFullscreenChange);
  return () => document.removeEventListener("fullscreenchange", onFullscreenChange);
});
```

- [ ] **Step 2: Bind `container` to the outer wrapper div + add the buttons to controls row**

Modify the outer wrapper:

```svelte
<div bind:this={container} class="relative h-full w-full bg-black overflow-hidden group">
```

Add right-aligned buttons (after the spacer):

```svelte
{#if pipSupported}
  <button
    type="button"
    onclick={togglePip}
    aria-label="Picture in Picture"
    class="inline-flex items-center justify-center w-9 h-9 rounded
           hover:bg-bg-hover hover:text-fg-accent transition-colors"
  >
    <PictureInPicture2 size="18" />
  </button>
{/if}
<button
  type="button"
  onclick={toggleFullscreen}
  aria-label={isFullscreen ? "Exit fullscreen" : "Fullscreen"}
  class="inline-flex items-center justify-center w-9 h-9 rounded
         hover:bg-bg-hover hover:text-fg-accent transition-colors"
>
  {#if isFullscreen}
    <Minimize size="18" />
  {:else}
    <Maximize size="18" />
  {/if}
</button>
```

- [ ] **Step 3: Type check**

```bash
cd frontend && npm run check
```

Expected: 0 errors.

- [ ] **Step 4: Manual visual verification**

In dev server (Chromium-based browser for full PiP support):
- Click fullscreen button → video fills screen, icon becomes Minimize
- Press Esc OR click again → exits fullscreen
- Click PiP button → mini floating video appears (browser-native), icon stays the same but underlying state toggles
- Click PiP button again OR close the mini → exits PiP

Firefox/Safari: PiP icon may be missing (not supported) — that's correct per spec.

- [ ] **Step 5: `/code-review`**

Expected: 0 ❌ Critical.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/components/Viewers/VideoPlayer.svelte
git commit -m "[feat] add fullscreen and picture-in-picture controls"
```

---

## Task 7: Playback speed selector

**Files:**
- Modify: `frontend/src/lib/components/Viewers/VideoPlayer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Speed button | Shows current rate as text (`1×`, `1.5×`, etc.), 36px tall |
| Click button | Opens vertical menu above with rates: 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2 |
| Click menu item | Sets `video.playbackRate`, closes menu |
| Active rate | Highlighted in menu (`text-fg-accent`) |
| Click outside | Closes menu |

- [ ] **Step 1: Add state + handlers**

```ts
const SPEEDS = [0.5, 0.75, 1, 1.25, 1.5, 1.75, 2];

let speed = $state(1);
let speedMenuOpen = $state(false);

function setSpeed(rate: number) {
  if (!video) return;
  video.playbackRate = rate;
  speed = rate;
  speedMenuOpen = false;
}

function onRateChange() {
  if (!video) return;
  speed = video.playbackRate;
}

function closeSpeedMenu(event: MouseEvent) {
  const target = event.target as HTMLElement;
  if (!target.closest("[data-speed-menu]")) speedMenuOpen = false;
}

$effect(() => {
  if (speedMenuOpen) {
    document.addEventListener("click", closeSpeedMenu);
    return () => document.removeEventListener("click", closeSpeedMenu);
  }
});
```

Add `onratechange={onRateChange}` to the `<video>` element.

- [ ] **Step 2: Add speed button + menu** (right side of controls, before PiP/Fullscreen):

```svelte
<div class="relative" data-speed-menu>
  <button
    type="button"
    onclick={(e) => { e.stopPropagation(); speedMenuOpen = !speedMenuOpen; }}
    aria-label="Playback speed"
    class="inline-flex items-center justify-center min-w-9 h-9 px-2 rounded
           font-mono text-xs tabular-nums
           hover:bg-bg-hover hover:text-fg-accent transition-colors"
  >
    {speed}×
  </button>
  {#if speedMenuOpen}
    <div
      class="absolute bottom-full right-0 mb-2 py-1 rounded-md
             bg-bg-elevated border border-border-default
             shadow-lg z-30 min-w-[72px]"
    >
      {#each SPEEDS as rate (rate)}
        <button
          type="button"
          onclick={() => setSpeed(rate)}
          class="w-full px-3 py-1.5 text-xs font-mono tabular-nums text-left
                 hover:bg-bg-hover transition-colors
                 {speed === rate ? 'text-fg-accent' : 'text-fg-primary'}"
        >
          {rate}×
        </button>
      {/each}
    </div>
  {/if}
</div>
```

- [ ] **Step 3: Type check** → expected 0 errors.

- [ ] **Step 4: Manual visual verification**
- Click speed button → vertical menu appears upward, list of 7 rates
- Click `1.5` → playback speeds up, menu closes, button shows `1.5×`
- Click button again, click outside the menu → menu closes
- Current rate highlighted in yellow in the menu

- [ ] **Step 5: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/components/Viewers/VideoPlayer.svelte
git commit -m "[feat] add playback speed selector menu"
```

---

## Task 8: Keyboard shortcuts

**Files:**
- Modify: `frontend/src/lib/components/Viewers/VideoPlayer.svelte`

**Spec:**

| Key | Action |
|---|---|
| `Space` / `K` | Toggle play/pause |
| `M` | Toggle mute |
| `F` | Toggle fullscreen |
| `←` | Seek -5 seconds |
| `→` | Seek +5 seconds |
| `↑` | Volume +10% |
| `↓` | Volume -10% |

Shortcuts only fire when the player is mounted and the user isn't typing in another input. Use `pickMediaKey` from `media-utils.ts` to map. Listener attaches on mount, detaches on destroy.

- [ ] **Step 1: Add the keyboard handler**

```ts
import { pickMediaKey, clampVolume } from "./media-utils";

const SEEK_STEP = 5;
const VOLUME_STEP = 0.1;

function onKeyDown(event: KeyboardEvent) {
  if (!video) return;
  // Skip if user is typing in an input/textarea/contenteditable
  const target = event.target as HTMLElement;
  if (target.matches("input, textarea, [contenteditable='true']")) return;

  const key = pickMediaKey(event);
  if (!key) return;
  event.preventDefault();

  switch (key) {
    case "toggle": toggle(); break;
    case "mute": toggleMute(); break;
    case "fullscreen": toggleFullscreen(); break;
    case "seekBack": video.currentTime = Math.max(0, video.currentTime - SEEK_STEP); break;
    case "seekForward": video.currentTime = Math.min(duration, video.currentTime + SEEK_STEP); break;
    case "volumeUp": video.volume = clampVolume(video.volume + VOLUME_STEP); video.muted = false; break;
    case "volumeDown": video.volume = clampVolume(video.volume - VOLUME_STEP); break;
  }
}

$effect(() => {
  window.addEventListener("keydown", onKeyDown);
  return () => window.removeEventListener("keydown", onKeyDown);
});
```

- [ ] **Step 2: Type check** → expected 0 errors.

- [ ] **Step 3: Manual verification** — walk through every row of the spec table while a video is open and confirm the action fires.

- [ ] **Step 4: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/components/Viewers/VideoPlayer.svelte
git commit -m "[feat] add keyboard shortcuts for video player"
```

---

## Task 9: Auto-hide controls overlay

**Files:**
- Modify: `frontend/src/lib/components/Viewers/VideoPlayer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Idle 2.5s while playing | Top bar + bottom controls fade out (`opacity-0`) + cursor hidden |
| Mousemove in viewport | Both fade back in immediately, idle timer resets |
| Paused | Always visible (no auto-hide while paused) |
| Hovering controls | Stays visible (cursor over controls = active) |
| Animation | `transition-opacity duration-200 ease-smooth` |

- [ ] **Step 1: Add state + idle logic**

```ts
const IDLE_HIDE_MS = 2500;

let controlsVisible = $state(true);
let idleTimer: number | null = null;

function showControls() {
  controlsVisible = true;
  if (idleTimer !== null) window.clearTimeout(idleTimer);
  if (playing) {
    idleTimer = window.setTimeout(() => {
      controlsVisible = false;
    }, IDLE_HIDE_MS);
  }
}

function onMouseMove() { showControls(); }

$effect(() => {
  // Re-run when `playing` flips: paused → always show, playing → start idle countdown
  if (!playing) {
    controlsVisible = true;
    if (idleTimer !== null) window.clearTimeout(idleTimer);
  } else {
    showControls();
  }
  return () => {
    if (idleTimer !== null) window.clearTimeout(idleTimer);
  };
});
```

- [ ] **Step 2: Wire `onmousemove` on the container + apply opacity classes to top bar + controls bar**

Add `onmousemove={onMouseMove}` to the outer `<div bind:this={container}>`.

Modify the top bar and controls bar wrappers to include the visibility class:

```svelte
<!-- top bar -->
<div
  class="absolute top-0 left-0 right-0 h-9 px-4 flex items-center
         bg-gradient-to-b from-bg-overlay/80 to-transparent
         text-fg-secondary text-xs font-mono z-10 pointer-events-none
         transition-opacity duration-200 ease-smooth
         {controlsVisible ? 'opacity-100' : 'opacity-0'}"
>
  ...
</div>

<!-- bottom controls bar -->
<div
  class="absolute bottom-0 left-0 right-0 px-4 pb-2 pt-3 flex flex-col gap-1
         bg-bg-overlay/80 backdrop-blur-sm
         text-fg-primary z-10
         transition-opacity duration-200 ease-smooth
         {controlsVisible ? 'opacity-100' : 'opacity-0 pointer-events-none'}"
>
  ...
</div>
```

Also conditionally hide the cursor when controls hidden:

```svelte
<div
  bind:this={container}
  onmousemove={onMouseMove}
  class="relative h-full w-full bg-black overflow-hidden group
         {controlsVisible ? '' : 'cursor-none'}"
>
```

- [ ] **Step 3: Type check** → 0 errors.

- [ ] **Step 4: Manual verification**
- Start playing a video, stop moving the mouse for 2.5s → controls + top bar fade out, cursor disappears
- Move mouse → controls reappear instantly, cursor returns
- Pause the video → controls stay visible indefinitely
- Resume → idle countdown restarts

- [ ] **Step 5: Check VideoPlayer.svelte line count**

```bash
wc -l frontend/src/lib/components/Viewers/VideoPlayer.svelte
```

Expected: ≤ 300 lines. If 250-300, proceed with ⚠️ warning. If > 300, **STOP and extract** `VideoControls.svelte` as a sub-component before continuing — that's a separate mini-task: pull the controls bar markup + relevant props into a new file, update VideoPlayer to use it.

- [ ] **Step 6: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/lib/components/Viewers/VideoPlayer.svelte
git commit -m "[feat] add auto-hide controls overlay with idle timer"
```

---

## Task 10: `AudioPlayer.svelte` — skeleton, centered layout

**Files:**
- Create: `frontend/src/lib/components/Viewers/AudioPlayer.svelte`

**Spec:**

| Behavior | Expected |
|---|---|
| Layout | Centered card, max 360px wide, vertically centered in container |
| Filename | Large (`text-lg`) at top of card, truncated if too long |
| `<audio>` element | Hidden (native controls not used; element is just for state) |
| Big play button | 96px circular, `bg-accent text-accent-fg`, centered below filename |
| Click big button | Toggles play/pause |
| Container | Fills parent, `bg-bg-base` (default app background) |

- [ ] **Step 1: Write the component**

```svelte
<!-- frontend/src/lib/components/Viewers/AudioPlayer.svelte -->
<script lang="ts">
  import Play from "lucide-svelte/icons/play";
  import Pause from "lucide-svelte/icons/pause";
  import { auth } from "$lib/store/auth.svelte";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  let audio: HTMLAudioElement | null = $state(null);
  let playing = $state(false);

  const mediaUrl = $derived(
    `/server/getAudioData?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );

  function toggle() {
    if (!audio) return;
    if (audio.paused) audio.play().catch(() => {});
    else audio.pause();
  }

  function onPlay() { playing = true; }
  function onPause() { playing = false; }
</script>

<div class="flex flex-col h-full w-full bg-bg-base items-center justify-center p-8">
  <div class="w-full max-w-[360px] flex flex-col items-center gap-6">
    <h2 class="text-lg font-sans text-fg-primary truncate w-full text-center" title={name}>
      {name}
    </h2>

    <button
      type="button"
      onclick={toggle}
      aria-label={playing ? "Pause" : "Play"}
      class="w-24 h-24 rounded-full bg-accent text-accent-fg
             flex items-center justify-center
             hover:bg-accent-hover transition-colors
             shadow-lg"
    >
      {#if playing}
        <Pause size="40" fill="currentColor" />
      {:else}
        <Play size="40" fill="currentColor" class="ml-1" />
      {/if}
    </button>

    <!-- placeholder: equalizer goes here in Task 12 -->
    <!-- placeholder: scrubber + time + volume go here in Task 11 -->
  </div>

  <audio
    bind:this={audio}
    src={mediaUrl}
    preload="metadata"
    onplay={onPlay}
    onpause={onPause}
    class="hidden"
  ></audio>
</div>
```

- [ ] **Step 2: Wire into `MediaViewer.svelte`**

```svelte
<script lang="ts">
  import VideoPlayer from "./VideoPlayer.svelte";
  import AudioPlayer from "./AudioPlayer.svelte";

  interface Props {
    loc: string;
    name: string;
    kind: "video" | "audio";
  }

  let { loc, name, kind }: Props = $props();
</script>

{#if kind === "video"}
  <VideoPlayer {loc} {name} />
{:else}
  <AudioPlayer {loc} {name} />
{/if}
```

- [ ] **Step 3: Type check** → 0 errors.

- [ ] **Step 4: Manual verification**
- Open an mp3/wav file. Expected: centered card, filename at top, big yellow circular play button below
- Click button → audio plays, icon changes to Pause
- Click again → pauses

- [ ] **Step 5: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/components/Viewers/AudioPlayer.svelte frontend/src/lib/components/Viewers/MediaViewer.svelte
git commit -m "[feat] add AudioPlayer skeleton with centered card and big play button"
```

---

## Task 11: AudioPlayer scrubber + time + volume

**Files:**
- Modify: `frontend/src/lib/components/Viewers/AudioPlayer.svelte`

**Spec:** Same as VideoPlayer's Task 4 + Task 5 (scrubber + buffered + time + volume slider with mute), but laid out vertically in the centered card. Volume slider stays always-visible (no hover slide-out) since there's room.

- [ ] **Step 1: Add state + handlers**

Add the same state (`currentTime`, `duration`, `buffered`, `volume`, `muted`) and handlers (`onTimeUpdate`, `onLoadedMetadata`, `onProgress`, `seek`, `setVolume`, `toggleMute`, `onVolumeChange`) as VideoPlayer Task 4 + 5, adapted to `audio` instead of `video`.

Add imports + derived icon (Svelte 5 capitalized name to avoid deprecated `<svelte:component>`):

```ts
import Volume2 from "lucide-svelte/icons/volume-2";
import Volume1 from "lucide-svelte/icons/volume-1";
import VolumeX from "lucide-svelte/icons/volume-x";
import { formatTime, clampVolume } from "./media-utils";

const VolumeIcon = $derived(
  muted || volume === 0 ? VolumeX : volume < 0.5 ? Volume1 : Volume2,
);
```

- [ ] **Step 2: Add scrubber + time + volume markup below the equalizer placeholder**

```svelte
<!-- replace the placeholder comments with: -->

<!-- scrubber -->
<div class="w-full flex flex-col gap-1">
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    onclick={seek}
    class="relative h-1.5 rounded-full bg-bg-hover cursor-pointer
           hover:h-2 transition-all"
  >
    <div
      class="absolute top-0 left-0 h-full rounded-full bg-bg-selected"
      style="width: {bufferedRatio * 100}%"
    ></div>
    <div
      class="absolute top-0 left-0 h-full rounded-full bg-accent"
      style="width: {playedRatio * 100}%"
    ></div>
  </div>
  <div class="flex justify-between font-mono text-xs text-fg-secondary tabular-nums">
    <span>{formatTime(currentTime)}</span>
    <span>{formatTime(duration)}</span>
  </div>
</div>

<!-- volume row -->
<div class="w-full flex items-center gap-2">
  <button
    type="button"
    onclick={toggleMute}
    aria-label={muted ? "Unmute" : "Mute"}
    class="inline-flex items-center justify-center w-8 h-8 rounded
           text-fg-muted hover:bg-bg-hover hover:text-fg-accent transition-colors"
  >
    <VolumeIcon size="16" />
  </button>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    onclick={setVolume}
    class="flex-1 relative h-1.5 rounded-full bg-bg-hover cursor-pointer"
  >
    <div
      class="absolute top-0 left-0 h-full rounded-full bg-fg-secondary"
      style="width: {muted ? 0 : volume * 100}%"
    ></div>
  </div>
</div>
```

Add `ontimeupdate`, `onloadedmetadata`, `onprogress`, `onvolumechange` to the `<audio>` element.

- [ ] **Step 3: Type check** → 0 errors.

- [ ] **Step 4: Manual verification**
- Play audio → scrubber fills yellow, time updates `0:00 / 3:42`
- Click scrubber midway → audio seeks
- Click volume slider → audio volume changes
- Click volume icon → mute toggles, icon swaps

- [ ] **Step 5: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/components/Viewers/AudioPlayer.svelte
git commit -m "[feat] add audio scrubber, time display, and volume control"
```

---

## Task 12: AudioPlayer animated equalizer

**Files:**
- Modify: `frontend/src/lib/components/Viewers/AudioPlayer.svelte`
- Modify: `frontend/src/app.css` (add keyframes — `@keyframes equalize-N`)

**Spec:**

| Behavior | Expected |
|---|---|
| 5 vertical bars | 6px wide, ~32px max height, gap 4px, centered horizontally |
| Color | `bg-accent` (yellow) |
| While playing | Each bar animates with a different period (0.6s, 0.8s, 1.0s, 0.7s, 0.9s) and offset; height oscillates 20-100% via CSS keyframes |
| While paused | All bars at 30% height, no animation (set via conditional class) |

This is **pure CSS** — no Web Audio API.

- [ ] **Step 1: Add keyframes to `app.css`**

Append to the bottom of `frontend/src/app.css`:

```css
@keyframes equalize-1 { 0%, 100% { transform: scaleY(0.3); } 50% { transform: scaleY(1); } }
@keyframes equalize-2 { 0%, 100% { transform: scaleY(0.5); } 50% { transform: scaleY(0.9); } }
@keyframes equalize-3 { 0%, 100% { transform: scaleY(0.7); } 50% { transform: scaleY(0.4); } }
@keyframes equalize-4 { 0%, 100% { transform: scaleY(0.4); } 50% { transform: scaleY(0.8); } }
@keyframes equalize-5 { 0%, 100% { transform: scaleY(0.6); } 50% { transform: scaleY(1); } }
```

- [ ] **Step 2: Add equalizer markup** in `AudioPlayer.svelte` between the big play button and the scrubber:

```svelte
<div class="flex items-end justify-center gap-1 h-8">
  {#each [1, 2, 3, 4, 5] as i (i)}
    <span
      class="block w-1.5 h-full rounded-sm bg-accent origin-bottom"
      style="animation: equalize-{i} {0.5 + i * 0.1}s ease-in-out infinite;
             animation-play-state: {playing ? 'running' : 'paused'};
             transform: scaleY({playing ? 1 : 0.3});"
    ></span>
  {/each}
</div>
```

- [ ] **Step 3: Type check** → 0 errors.

- [ ] **Step 4: Manual verification**
- Pause state: 5 short bars (~30% height), still
- Play state: bars dance with offset rhythms
- Pause again: bars settle back to 30%

If paused bars don't visually settle to 30% (CSS keyframes may freeze mid-cycle), add `animation-fill-mode: both` to the inline style, or fall back to a class toggle:

```svelte
<span
  class="block w-1.5 h-full rounded-sm bg-accent origin-bottom {playing ? 'animate-eq-' + i : ''}"
  style={playing ? `animation: equalize-${i} ${0.5 + i * 0.1}s ease-in-out infinite;` : `transform: scaleY(0.3);`}
></span>
```
(Decide during this step — keep the simpler inline form if it works.)

- [ ] **Step 5: `/code-review`** on both files → 0 ❌ Critical.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/components/Viewers/AudioPlayer.svelte frontend/src/app.css
git commit -m "[feat] add animated equalizer bars to audio player"
```

---

## Task 13: AudioPlayer keyboard shortcuts

**Files:**
- Modify: `frontend/src/lib/components/Viewers/AudioPlayer.svelte`

**Spec:** Same keys as VideoPlayer Task 8 **minus fullscreen** (no fullscreen on audio).

| Key | Action |
|---|---|
| `Space` / `K` | Toggle |
| `M` | Mute |
| `←` | Seek -5s |
| `→` | Seek +5s |
| `↑` | Volume +10% |
| `↓` | Volume -10% |
| `F` | (no-op) |

- [ ] **Step 1: Add the handler** (analogous to VideoPlayer Task 8, but ignore `fullscreen` case):

```ts
import { pickMediaKey } from "./media-utils";

const SEEK_STEP = 5;
const VOLUME_STEP = 0.1;

function onKeyDown(event: KeyboardEvent) {
  if (!audio) return;
  const target = event.target as HTMLElement;
  if (target.matches("input, textarea, [contenteditable='true']")) return;

  const key = pickMediaKey(event);
  if (!key || key === "fullscreen") return;
  event.preventDefault();

  switch (key) {
    case "toggle": toggle(); break;
    case "mute": toggleMute(); break;
    case "seekBack": audio.currentTime = Math.max(0, audio.currentTime - SEEK_STEP); break;
    case "seekForward": audio.currentTime = Math.min(duration, audio.currentTime + SEEK_STEP); break;
    case "volumeUp": audio.volume = clampVolume(audio.volume + VOLUME_STEP); audio.muted = false; break;
    case "volumeDown": audio.volume = clampVolume(audio.volume - VOLUME_STEP); break;
  }
}

$effect(() => {
  window.addEventListener("keydown", onKeyDown);
  return () => window.removeEventListener("keydown", onKeyDown);
});
```

- [ ] **Step 2: Type check** → 0 errors.

- [ ] **Step 3: Manual verification** — walk the spec table.

- [ ] **Step 4: `/code-review`** → 0 ❌ Critical.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/components/Viewers/AudioPlayer.svelte
git commit -m "[feat] add keyboard shortcuts for audio player"
```

---

## Task 14: `MediaViewer.svelte` final dispatcher cleanup

**Files:**
- Modify: `frontend/src/lib/components/Viewers/MediaViewer.svelte`

**Spec:** Should already be the clean ~12-line dispatcher from Task 10. Verify and tidy.

Final shape:

```svelte
<script lang="ts">
  import VideoPlayer from "./VideoPlayer.svelte";
  import AudioPlayer from "./AudioPlayer.svelte";

  interface Props {
    loc: string;
    name: string;
    kind: "video" | "audio";
  }

  let { loc, name, kind }: Props = $props();
</script>

{#if kind === "video"}
  <VideoPlayer {loc} {name} />
{:else}
  <AudioPlayer {loc} {name} />
{/if}
```

- [ ] **Step 1: Verify file contents match the above.** If yes, no commit needed (already done in Task 10).

- [ ] **Step 2: Line count check**

```bash
wc -l frontend/src/lib/components/Viewers/MediaViewer.svelte
```

Expected: ≤ 20 lines.

- [ ] **Step 3: `/code-review`** → 0 ❌ Critical.

(No commit step if no changes since Task 10.)

---

## Task 15: Final integration verification + build

**Files:** None modified.

- [ ] **Step 1: Type check entire frontend**

```bash
cd frontend && npm run check
```

Expected: 0 errors, 0 warnings.

- [ ] **Step 2: Production build**

```bash
cd frontend && npm run build
```

Expected: build succeeds, no errors.

- [ ] **Step 3: Backend integration test sanity**

```bash
cd backend && go test ./tests/integration/ -run TestStream 2>&1 | tail -20
```

Expected: existing stream tests still pass (Range support unchanged).

- [ ] **Step 4: Full manual walkthrough**

Start dev server, log in, then test the matrix:

| File type | Expected behavior |
|---|---|
| `.mp4` | VideoPlayer loads, all controls work, keyboard works, fullscreen + PiP work, auto-hide works |
| `.webm` | Same as above |
| `.mp3` | AudioPlayer loads, centered card, big play button, scrubber works, equalizer animates, keyboard works |
| `.wav` | Same |
| `.flac` | Same |

For each, run the full **Spec Walkthrough Table**:

| # | Action | Expected | Status |
|---|---|---|---|
| 1 | Open file | Player loads, filename visible | |
| 2 | Click center/big play | Plays | |
| 3 | Click again | Pauses | |
| 4 | Click scrubber 50% | Seeks to middle | |
| 5 | Click scrubber 10% | Seeks back | |
| 6 | Hover volume icon (video only) | Slider slides out | |
| 7 | Click volume slider 50% | Volume halves | |
| 8 | Click volume icon | Mutes, icon → VolumeX | |
| 9 | Click again | Unmutes, restores volume |  |
| 10 | Click PiP (Chromium, video only) | Floating mini player appears | |
| 11 | Click Fullscreen (video only) | Container fills screen | |
| 12 | Esc / click Fullscreen again | Exits fullscreen | |
| 13 | Click speed button (video only) | Menu opens above | |
| 14 | Click 1.5× | Plays at 1.5×, button label updates | |
| 15 | Click outside menu | Menu closes | |
| 16 | Press Space | Toggles play | |
| 17 | Press M | Toggles mute | |
| 18 | Press F (video only) | Toggles fullscreen | |
| 19 | Press ← | -5s seek | |
| 20 | Press → | +5s seek | |
| 21 | Press ↑ | +10% volume | |
| 22 | Press ↓ | -10% volume | |
| 23 | Play video, idle 3s (video only) | Controls fade out, cursor hides | |
| 24 | Mousemove (video only) | Controls return | |
| 25 | Pause (video only) | Controls stay forever | |
| 26 | Equalizer (audio only, playing) | Bars dance | |
| 27 | Equalizer (audio only, paused) | Bars at 30%, still | |

- [ ] **Step 5: If all 27 rows pass, run final `/code-review`** on:
- `frontend/src/lib/components/Viewers/media-utils.ts`
- `frontend/src/lib/components/Viewers/VideoPlayer.svelte`
- `frontend/src/lib/components/Viewers/AudioPlayer.svelte`
- `frontend/src/lib/components/Viewers/MediaViewer.svelte`
- `frontend/src/app.css` (only the keyframes block added in Task 12)

Expected: 14/14 pass on each file, 0 ❌ Critical, 0 in-scope ⚠️ Warning.

- [ ] **Step 6: Final commit** (only if anything was tweaked in Step 5; otherwise skip):

```bash
git commit -m "[feat] complete professional media player UI for video and audio"
```

- [ ] **Step 7: Push the feature branch**

```bash
git push origin feat/media-viewers
```

- [ ] **Step 8: Open PR** via GitHub web UI:

`https://github.com/JMC50/nas/pull/new/feat/media-viewers`

PR title: `[feat] professional video & audio player UI`

PR body should reference this plan: `Implements docs/superpowers/plans/2026-05-13-pro-media-player.md`. Include the 27-row Spec Walkthrough Table as the test plan, marked ✅ for each verified row.

---

## Completion Criteria

- All 15 tasks committed.
- `npm run check`: 0 errors / 0 warnings.
- `npm run build`: succeeds.
- `/code-review` on every new/modified file: 0 ❌ Critical, 0 in-scope ⚠️ Warning.
- All 27 rows of the Spec Walkthrough Table in Task 15 pass.
- PR opened against `main`.

## Risk Register

| Risk | Mitigation |
|---|---|
| `VideoPlayer.svelte` exceeds 300 lines | Task 9 Step 5 explicit checkpoint; extract `VideoControls.svelte` before continuing |
| Browser doesn't support PiP (Firefox, Safari) | `pipSupported` derived, button hides automatically |
| Browser blocks fullscreen (user dismissed) | `.catch(() => {})` swallows promise rejection — user just sees nothing happen, no error toast |
| Range requests fail on backend | Out of scope — existing `http.ServeContent` already serves Range correctly per `backend/tests/integration/stream_test.go` |
| User keyboard shortcuts conflict with global app shortcuts | Listener checks `target.matches("input, textarea, [contenteditable]")` to skip when typing; global app uses Ctrl+S etc., no plain key conflicts |
| Auto-hide controls interferes with accessibility | `pointer-events-none` only when hidden; keyboard focus on a button keeps controls visible via mousemove-equivalent (focus listener could be added if needed — flag as v2) |

## Related Skills (for implementing this plan)

- `superpowers:subagent-driven-development` — dispatching fresh sub-agents per task with two-stage review (recommended)
- `superpowers:executing-plans` — inline execution with checkpoints
- `code-review` — invoked after every implementation step
- Project-specific: ralph-mode workflow (RIPER-5) per `~/.claude/CLAUDE.md` and `~/.claude/RULES.md`
