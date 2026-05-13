# Frontend Refactor Plan — VSCode-style Shell + Persistent Upload State

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the Svelte frontend from its current "AI slop" appearance + broken navigation state into a cohesive VSCode-like IDE pattern: persistent left nav, tabbed main area, global upload toast, dark-first design system. Fix the long-standing bug where upload progress disappears when navigating between sections. Migrate Kakao OAuth out, Google OAuth in.

**Architecture decision:** Single-page tab system where the file explorer and individual opened files are all tabs in one persistent surface. Routing handles OAuth callbacks only; in-app navigation is state + tab driven. Global Svelte stores hold uploads, tabs, files, and UI state so that section switching is purely view changes — nothing unmounts mid-operation.

**Tech stack additions:**
- **SvelteKit** (`@sveltejs/kit` + `@sveltejs/adapter-static`) — file-based routing, layouts, type-safe load functions. Static adapter outputs to `frontend/build/`, served by Go backend's `web.MountSPA`
- **TailwindCSS 4** (or 3.x with Vite plugin) — design tokens via `tailwind.config.js`
- **shadcn-svelte** (Svelte 5 + SvelteKit compatible) — unstyled accessible primitives (Dialog, Tabs, Toast, ContextMenu, Sheet, Tooltip, DropdownMenu, ScrollArea)
- **bits-ui** (shadcn-svelte's underlying primitive lib)
- **lucide-svelte** — icon set
- **mode-watcher** — system color-scheme integration
- Keep: **monaco-editor** + **@monaco-editor/loader** (already in deps, central to text editing story)

**Tech stack retired:**
- `*_mobile.svelte` duplicate components — replaced with responsive Tailwind breakpoints
- Raw SCSS in `<style>` blocks — replaced with Tailwind utility classes + a few global CSS variables for theme
- `location.pathname == "/login"` path checks — replaced with SvelteKit file-based routes (`src/routes/login/+page.svelte`)
- Plain Svelte+Vite scaffold — replaced with SvelteKit (file-based routes, layouts, typed loaders)

---

## Design System Foundation

### Color tokens (gruvbox-dark, mandatory)

Source: Pavel Pertsev's gruvbox (https://github.com/morhetz/gruvbox). Warm retro-groove palette — never substitute with cool dev-tool blues.

```css
/* === Dark mode (default) === */

/* Background scale (darkest → lightest) */
--bg-base:      #1d2021;  /* hard bg — page background */
--bg-surface:   #282828;  /* medium bg — sidebar, panels */
--bg-elevated:  #32302f;  /* soft bg — cards, tab strip */
--bg-overlay:   #3c3836;  /* bg1 — modals, popovers */
--bg-hover:     #504945;  /* bg2 — interactive hover */
--bg-selected:  #665c54;  /* bg3 — selected row */

/* Border / divider */
--border-default: #504945;  /* bg2 */
--border-strong:  #665c54;  /* bg3 */
--border-focus:   #fabd2f;  /* yellow — keyboard focus ring */

/* Text */
--fg-primary:   #ebdbb2;  /* fg1 — main text */
--fg-secondary: #d5c4a1;  /* fg2 — secondary */
--fg-muted:     #a89984;  /* fg4 — labels, hints */
--fg-disabled:  #928374;  /* gray */

/* Semantic colors (gruvbox bright variants for legibility on dark bg) */
--fg-accent:    #fabd2f;  /* yellow — primary accent */
--fg-success:   #b8bb26;  /* green */
--fg-warning:   #fe8019;  /* orange */
--fg-danger:    #fb4934;  /* red */
--fg-info:      #83a598;  /* faded blue — secondary accent */
--fg-link:      #8ec07c;  /* aqua — links, breadcrumb */
--fg-purple:    #d3869b;  /* purple — special tags */

/* Brand accent (yellow, used for primary buttons + selected state) */
--accent:       #fabd2f;
--accent-hover: #fbe69b;  /* lighter on hover */
--accent-fg:    #282828;  /* dark fg ON yellow accent (high contrast) */
--accent-muted: #d79921;  /* faded for subtle uses */
```

### Light mode (gruvbox-light)

```css
[data-theme="light"] {
  --bg-base:      #f9f5d7;  /* hard bg — cream */
  --bg-surface:   #fbf1c7;  /* fg0 — paper */
  --bg-elevated:  #f2e5bc;  /* soft bg */
  --bg-overlay:   #ebdbb2;  /* bg1 */
  --bg-hover:     #d5c4a1;  /* bg2 */
  --bg-selected:  #bdae93;  /* bg3 */

  --border-default: #d5c4a1;
  --border-strong:  #bdae93;
  --border-focus:   #b57614;  /* yellow faded */

  --fg-primary:   #3c3836;  /* fg1 */
  --fg-secondary: #504945;  /* fg2 */
  --fg-muted:     #7c6f64;  /* fg4 */
  --fg-disabled:  #928374;  /* gray */

  /* Semantic — gruvbox faded variants on light bg */
  --fg-accent:    #b57614;  /* yellow faded */
  --fg-success:   #79740e;  /* green faded */
  --fg-warning:   #af3a03;  /* orange faded */
  --fg-danger:    #9d0006;  /* red faded */
  --fg-info:      #076678;  /* blue faded */
  --fg-link:      #427b58;  /* aqua faded */
  --fg-purple:    #8f3f71;  /* purple faded */

  --accent:       #b57614;
  --accent-hover: #d79921;
  --accent-fg:    #fbf1c7;  /* cream on yellow */
  --accent-muted: #fabd2f;
}
```

Default is dark. Toggle via header button or follows system `prefers-color-scheme`.

### Monaco editor theme

Use `gruvbox-dark` (and `gruvbox-light` for light mode). Either:
- Install via `monaco-themes` package: `import GruvboxDark from "monaco-themes/themes/Gruvbox Dark.json"`, then `monaco.editor.defineTheme("gruvbox-dark", GruvboxDark)`
- Or define inline matching the CSS tokens above for guaranteed consistency with shell chrome

Set on Monaco instance creation: `monaco.editor.create(el, { theme: "gruvbox-dark", ... })`. Switching mode → call `monaco.editor.setTheme()`.

### Typography

- Sans: `Inter` (or system stack fallback) — UI text
- Mono: `JetBrains Mono` (or `ui-monospace`) — Monaco editor, code blocks, file names of code files
- Size scale: 12 / 13 / 14 / 16 / 18 / 24 / 32 (px). Default UI 13px (dense), body 14px.
- Line heights: tight 1.25 (headings), normal 1.5 (body), loose 1.6 (long reading).

### Spacing scale

Tailwind default (`0.25rem` increments). Project conventions:
- Section padding: `p-4` (16px)
- Card/tab padding: `px-3 py-2` (12/8)
- Dense list item: `px-2 py-1.5`

### Radius & shadow

- Radius scale: `rounded` (4px), `rounded-md` (6px), `rounded-lg` (8px)
- Dark-mode shadow: very subtle — `shadow-[0_1px_2px_rgba(0,0,0,0.6)]` for elevated surfaces. Avoid heavy drop shadows.

### Iconography

- `lucide-svelte` for all icons. Strict 16/20/24px sizes. Stroke-width 1.5 default.

---

## Layout Anatomy (VSCode pattern)

```
┌────────────────────────────────────────────────────────────────────┐
│ ⌘ NAS    🔍 Quick open (Ctrl+P)            ⚙️ 👤 ☀️/🌙   │ 32px Header
├──┬─────────────────────────────────────────────────────────────────┤
│📁│ 📑 Explorer  📝 report.md •  🎬 movie.mp4   [+]              │ 36px Tab bar
│👥├─────────────────────────────────────────────────────────────────┤
│📊│                                                                 │
│⚙️│                       Main content area                         │
│  │                  (active tab content renders here)              │
│  │                                                                 │
│  │                                                                 │
├──┴─────────────────────────────────────────────────────────────────┤
│ 📤 photo.jpg ▓▓▓░░ 60% · 2 more uploading ⏯️                 │ 28px Status bar
└────────────────────────────────────────────────────────────────────┘
 48px                                                                 
```

Components map:
- **Header**: `Shell/Header.svelte` (logo + Cmd+P search trigger + theme toggle + user menu)
- **VerticalNav**: `Shell/VerticalNav.svelte` (icon-only nav, optional labels on hover)
- **TabBar**: `Tabs/TabBar.svelte` (Chrome-style tabs with close buttons, drag-reorder, +)
- **Main**: `Tabs/TabContent.svelte` (renders active tab's component)
- **StatusBar**: `Shell/StatusBar.svelte` (upload toast collapsed view + connection status)

---

## Target Directory Structure (post-Phase 0)

```
frontend/
├── svelte.config.js           # adapter-static + Kit config
├── vite.config.ts             # SvelteKit Vite plugin + Tailwind plugin
├── tailwind.config.js
├── postcss.config.js
├── tsconfig.json              # extends .svelte-kit/tsconfig
├── components.json            # shadcn-svelte config
├── package.json
├── src/
│   ├── app.html               # root HTML template
│   ├── app.css                # Tailwind imports + theme tokens
│   ├── app.d.ts               # SvelteKit types
│   ├── routes/
│   │   ├── +layout.svelte     # Shell wrapper (Header + VerticalNav + StatusBar)
│   │   ├── +layout.ts         # auth guard + initial data
│   │   ├── +page.svelte       # / → main app surface (tabs + Explorer)
│   │   ├── login/
│   │   │   └── +page.svelte   # Discord OAuth callback
│   │   ├── googleLogin/
│   │   │   └── +page.svelte   # Google OAuth callback
│   │   └── localLogin/
│   │       └── +page.svelte   # Local login form
│   └── lib/
│       ├── components/
│       │   ├── ui/            # shadcn-svelte primitives
│       │   ├── Shell/         # Header, VerticalNav, StatusBar
│       │   ├── Tabs/          # TabBar, TabContent
│       │   ├── Explorer/      # Breadcrumb, FileGrid, FileList, ContextMenu
│       │   ├── Viewers/       # MonacoViewer, ImageViewer, MediaViewer, PdfViewer
│       │   └── Uploads/       # UploadToast, UploadPanel
│       ├── store/
│       │   ├── auth.ts
│       │   ├── tabs.ts
│       │   ├── uploads.ts
│       │   ├── files.ts
│       │   ├── ui.ts
│       │   └── notifications.ts
│       ├── api/
│       │   └── client.ts      # typed fetch wrapper around /server/* endpoints
│       ├── tus-client.ts      # tus 1.0.0 client for uploads
│       └── keyboard.ts        # global shortcut registry
```

Build output goes to `frontend/build/` (set in `svelte.config.js` adapter-static config). Go's `web.MountSPA(cfg.FrontendDir)` points at this directory.

## State Architecture

### New stores (all in `src/store/`)

```
store/
├── auth.ts        # Existing useAuth, refactored to Svelte 5 runes + strict types
├── tabs.ts        # Open tabs (explorer + opened files/users); active tab
├── uploads.ts     # Upload tracker — survives navigation, tus integration
├── files.ts       # Current folder path, file list, selection, view mode (grid/list)
├── ui.ts          # Theme, sidebar collapsed, viewport size, quick-open palette
└── notifications.ts  # Toast/alert queue
```

### Tab store shape

```typescript
type TabKind = "explorer" | "text" | "image" | "video" | "audio" | "pdf" | "user-manager";

interface Tab {
  id: string;            // uuid
  kind: TabKind;
  title: string;         // displayed in tab
  icon: string;          // lucide icon name
  payload: unknown;      // tab-specific (e.g. {loc, name} for files)
  dirty?: boolean;       // text tabs: unsaved changes
  closable: boolean;     // explorer tab cannot be closed
}

const tabs = writable<Tab[]>([{ id: "explorer", kind: "explorer", title: "Files", icon: "folder", payload: null, closable: false }]);
const activeTabId = writable<string>("explorer");
```

### Upload store shape

```typescript
type UploadStatus = "queued" | "uploading" | "paused" | "complete" | "error" | "cancelled";

interface Upload {
  id: string;            // tus upload URL or uuid
  filename: string;
  loc: string;
  totalBytes: number;
  uploadedBytes: number;
  status: UploadStatus;
  startedAt: number;
  completedAt?: number;
  errorMessage?: string;
  tusUrl?: string;       // for resume
  file: File | Blob;     // source (held in memory while uploading)
}

const uploads = writable<Upload[]>([]);
// Derived:
const activeUploads = derived(uploads, $u => $u.filter(u => u.status === "uploading" || u.status === "paused" || u.status === "queued"));
const overallProgress = derived(activeUploads, ...);
```

**Persistence:** uploads store is in-memory only (File objects can't survive page reload). On reload, paused/in-flight uploads show as "expired" — user must re-add the file. The tus protocol allows resume from server offset, but we'd need to ask the user to re-select the same file.

For uploads completed within the session, we keep them in store for ~60s after completion (so they show in the "Recently uploaded" panel) then auto-prune.

---

## Per-Phase Quality Gate

**MANDATORY after every Task that writes/edits code, in EVERY phase:**
1. `npm run build` (or `npx vite build`) succeeds with zero TS errors
2. `npx svelte-check --fail-on-warnings` clean
3. Visual smoke: build, run Go backend, navigate via Playwright MCP, confirm no console errors and expected UI state
4. **Invoke `code-review` skill on changed files** (per global CLAUDE.md rule — NOT optional)
5. Fix every ❌ Critical from review before next Task; ⚠️ Warnings get surfaced to user

Phase 1 spells the gate out explicitly so the pattern is clear; subsequent phases assume it.

---

# PHASE 0: SvelteKit migration (Day 1)

**Goal:** Replace the plain Svelte + Vite scaffold with SvelteKit (static adapter). Existing components keep working; only the surrounding shell changes. After this phase, file-based routing is the canonical way to add new pages.

**Done when:**
- `npm run dev` boots SvelteKit, serves at port 8086 with HMR working
- `npm run build` produces `frontend/build/` (SPA assets, fallback `index.html`)
- Visiting `/`, `/login`, `/googleLogin` (placeholder), `/localLogin` resolves to a SvelteKit route
- Go backend serves `frontend/build/` correctly (verified by `curl http://localhost:7777/` returning the SPA HTML)

### Task 0.1: Install SvelteKit + adapter

- [ ] **Step 1: Add deps**

```bash
cd frontend
npm i -D @sveltejs/kit @sveltejs/adapter-static @sveltejs/vite-plugin-svelte
# @sveltejs/vite-plugin-svelte should already be present — verify and update if older than 4.x
```

- [ ] **Step 2: Create `svelte.config.js`**

```js
import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

export default {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({
      pages: "build",
      assets: "build",
      fallback: "index.html",   // SPA mode — all unknown routes fall back to client-side router
      precompress: false,
      strict: true,
    }),
    alias: {
      "$lib": "./src/lib",
      "$lib/*": "./src/lib/*",
    },
  },
};
```

- [ ] **Step 3: Replace `vite.config.mts` → `vite.config.ts` with SvelteKit plugin**

```ts
import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig, loadEnv } from "vite";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const serverURL = env.SERVER_URL || "http://localhost:7777";

  return {
    plugins: [sveltekit()],
    server: {
      port: 8086,
      host: "0.0.0.0",
      proxy: {
        "/server": {
          target: serverURL,
          changeOrigin: true,
          secure: false,
          rewrite: (p) => p.replace(/^\/server/, ""),
        },
      },
    },
    define: {
      "process.env.SERVER_URL": JSON.stringify(serverURL),
      "process.env.LOGIN_URL": JSON.stringify(env.DISCORD_LOGIN_URL ?? ""),
      "process.env.GOOGLE_CLIENT_ID": JSON.stringify(env.GOOGLE_CLIENT_ID ?? ""),
      "process.env.GOOGLE_REDIRECT_URI": JSON.stringify(env.GOOGLE_REDIRECT_URI ?? ""),
    },
  };
});
```

- [ ] **Step 4: Delete `vite.config.mts`**

### Task 0.2: Restructure source tree

- [ ] **Step 1: Create SvelteKit skeleton files**

`src/app.html`:
```html
<!doctype html>
<html lang="ko" data-theme="dark">
  <head>
    <meta charset="utf-8" />
    <link rel="icon" href="%sveltekit.assets%/favicon.ico" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>NAS</title>
    %sveltekit.head%
  </head>
  <body data-sveltekit-preload-data="hover">
    <div style="display: contents">%sveltekit.body%</div>
  </body>
</html>
```

`src/app.d.ts`:
```ts
declare global {
  namespace App {
    // interface Error {}
    // interface Locals {}
    // interface PageData {}
    // interface Platform {}
  }
}
export {};
```

- [ ] **Step 2: Migrate routing — App.svelte path checks → SvelteKit routes**

Create the following stub routes (real content comes later):

`src/routes/+layout.svelte` — temporarily just `<slot />` (Shell wraps it in Phase 3)

`src/routes/+page.svelte` — move the non-OAuth `<main>` content here (the section switcher + Explorer/Account/etc.)

`src/routes/login/+page.svelte` — move `LoginRedirect.svelte` content

`src/routes/localLogin/+page.svelte` — move `LocalLogin.svelte` content

`src/routes/googleLogin/+page.svelte` — placeholder for Phase 6

- [ ] **Step 3: Move components to `$lib/components/`**

```bash
# Inside frontend/src/
mkdir -p lib/components
# Move existing components — they become reusable building blocks rather than route components
git mv lib/Account.svelte lib/components/Account.svelte
git mv lib/AccountViewer.svelte lib/components/AccountViewer.svelte
git mv lib/ActivityLog.svelte lib/components/ActivityLog.svelte
git mv lib/BottomMenu.svelte lib/components/BottomMenu.svelte
git mv lib/Explorer.svelte lib/components/Explorer.svelte
git mv lib/Explorer_mobile.svelte lib/components/Explorer_mobile.svelte
git mv lib/FileManager.svelte lib/components/FileManager.svelte
git mv lib/FileViewer.svelte lib/components/FileViewer.svelte
git mv lib/FileViewer_mobile.svelte lib/components/FileViewer_mobile.svelte
git mv lib/Setting.svelte lib/components/Setting.svelte
git mv lib/SideMenu.svelte lib/components/SideMenu.svelte
git mv lib/SystemInfo.svelte lib/components/SystemInfo.svelte
git mv lib/SystemInfo_mobile.svelte lib/components/SystemInfo_mobile.svelte
git mv lib/UserManager.svelte lib/components/UserManager.svelte
# Leftovers like LocalLogin/LoginRedirect/LoginRedirectKakao moved into routes/ above —
# delete originals after route migration is confirmed working
```

- [ ] **Step 4: Update all imports throughout the codebase**

Run search-and-replace: `from "./lib/Xxx.svelte"` → `from "$lib/components/Xxx.svelte"`. SvelteKit's `$lib` alias resolves to `src/lib/`.

- [ ] **Step 5: Delete old files**

```bash
rm src/App.svelte       # replaced by src/routes/+page.svelte + +layout.svelte
rm src/main.ts          # SvelteKit handles entry
```

`index.html` at the project root also goes — SvelteKit uses `src/app.html`.

- [ ] **Step 6: Update `package.json` scripts**

```json
"scripts": {
  "dev": "vite dev",
  "build": "vite build",
  "preview": "vite preview",
  "check": "svelte-kit sync && svelte-check --tsconfig ./tsconfig.json",
  "check:watch": "svelte-kit sync && svelte-check --tsconfig ./tsconfig.json --watch"
}
```

- [ ] **Step 7: Verify dev mode + build**

```bash
npm run dev
# Open http://localhost:8086/ → should show same content as before (just routed through Kit)
# Open http://localhost:8086/localLogin → should show local login UI
```

```bash
npm run build
ls build/
# Expect: index.html, _app/*, favicon.ico
```

- [ ] **Step 8: Update Go backend's SPA path**

`frontend/dist/` → `frontend/build/` everywhere:
- `backend/Dockerfile`: `COPY --from=frontend-build /work/frontend/build /app/frontend/build` (and `ENV FRONTEND_DIR=/app/frontend/build`)
- Any dev scripts referencing `frontend/dist`

- [ ] **Step 9: End-to-end smoke**

```bash
# Build frontend
cd frontend && npm run build && cd ..

# Run Go backend pointing at new build dir
cd backend && go build -o ./bin/server.exe ./cmd/server
$env:FRONTEND_DIR = "C:\Data\Git\ANXI\nas\frontend\build"; ./bin/server.exe &

# Verify
curl http://localhost:7778/        # SPA HTML
curl http://localhost:7778/server/healthz  # JSON
```

- [ ] **Step 10: Quality gate + commit**

```bash
git add frontend/ backend/Dockerfile
git commit -m "[refactor] migrate frontend to SvelteKit with adapter-static

[Body]
- @sveltejs/kit + @sveltejs/adapter-static replace plain Svelte+Vite scaffold
- src/routes/ for file-based routing (replaces location.pathname checks in old App.svelte)
- src/lib/components/ holds existing components, reachable via \$lib alias
- OAuth callbacks become proper routes: /login (Discord), /localLogin, /googleLogin (placeholder)
- Build output frontend/build/ (was frontend/dist/) — Dockerfile + FRONTEND_DIR updated
- Vite proxy unchanged: /server/* still hits backend at :7777"
```

### Task 0.3: Phase 0 closeout

- [ ] **Step 1: Manual visual audit** — visit all 4 routes (`/`, `/login`, `/localLogin`, `/googleLogin`) and confirm each renders without console errors. UI looks identical to pre-migration (no design changes in Phase 0).

- [ ] **Step 2: Update plan retrospective with anything that surprised**

---

# PHASE 1: Foundation — Tailwind, design tokens, shadcn-svelte (Day 2)

**Goal:** Tailwind installed, theme tokens defined, shadcn-svelte ready, dark mode default, no visual changes yet to existing pages (so we don't break things mid-migration).

**Done when:**
- `npm run dev` shows the existing app unchanged
- A single `<button class="bg-accent text-accent-fg px-3 py-1.5 rounded-md">Test</button>` rendered in `src/routes/+page.svelte` previews correctly with theme tokens
- Dark mode active by default; light mode toggleable via system preference

### Task 1.1: Install Tailwind + tooling

- [ ] **Step 1: Add dependencies**

```bash
cd frontend
npm i -D tailwindcss postcss autoprefixer @tailwindcss/vite
npm i -D bits-ui lucide-svelte mode-watcher
npm i -D shadcn-svelte # CLI tool for component installs (devDep)
```

- [ ] **Step 2: Initialize Tailwind config**

`frontend/tailwind.config.js`:
```js
import { fontFamily } from "tailwindcss/defaultTheme";

export default {
  darkMode: ["class", '[data-theme="dark"]'],
  content: ["./index.html", "./src/**/*.{svelte,ts,js}"],
  theme: {
    extend: {
      colors: {
        bg: {
          base: "var(--bg-base)",
          surface: "var(--bg-surface)",
          elevated: "var(--bg-elevated)",
          overlay: "var(--bg-overlay)",
          hover: "var(--bg-hover)",
          selected: "var(--bg-selected)",
        },
        fg: {
          DEFAULT: "var(--fg-primary)",
          secondary: "var(--fg-secondary)",
          muted: "var(--fg-muted)",
          disabled: "var(--fg-disabled)",
          accent: "var(--fg-accent)",
          success: "var(--fg-success)",
          warning: "var(--fg-warning)",
          danger: "var(--fg-danger)",
          info: "var(--fg-info)",
          link: "var(--fg-link)",
          purple: "var(--fg-purple)",
        },
        border: {
          DEFAULT: "var(--border-default)",
          strong: "var(--border-strong)",
          focus: "var(--border-focus)",
        },
        accent: {
          DEFAULT: "var(--accent)",
          hover: "var(--accent-hover)",
          fg: "var(--accent-fg)",
          muted: "var(--accent-muted)",
        },
      },
      fontFamily: {
        sans: ["Inter", ...fontFamily.sans],
        mono: ["JetBrains Mono", ...fontFamily.mono],
      },
      fontSize: {
        xs: "12px",
        sm: "13px",
        base: "14px",
        md: "16px",
        lg: "18px",
        xl: "24px",
        "2xl": "32px",
      },
      transitionTimingFunction: {
        smooth: "cubic-bezier(0.16, 1, 0.3, 1)",
      },
    },
  },
};
```

- [ ] **Step 3: Add Tailwind directives to main CSS**

Create `frontend/src/app.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    /* gruvbox-dark (default) — paste full set from "Color tokens (gruvbox-dark, mandatory)" */
    --bg-base: #1d2021;
    --bg-surface: #282828;
    --bg-elevated: #32302f;
    --bg-overlay: #3c3836;
    --bg-hover: #504945;
    --bg-selected: #665c54;
    --border-default: #504945;
    --border-strong: #665c54;
    --border-focus: #fabd2f;
    --fg-primary: #ebdbb2;
    --fg-secondary: #d5c4a1;
    --fg-muted: #a89984;
    --fg-disabled: #928374;
    --fg-accent: #fabd2f;
    --fg-success: #b8bb26;
    --fg-warning: #fe8019;
    --fg-danger: #fb4934;
    --fg-info: #83a598;
    --fg-link: #8ec07c;
    --fg-purple: #d3869b;
    --accent: #fabd2f;
    --accent-hover: #fbe69b;
    --accent-fg: #282828;
    --accent-muted: #d79921;
  }

  [data-theme="light"] {
    /* gruvbox-light overrides — see "Light mode (gruvbox-light)" section */
    --bg-base: #f9f5d7;
    --bg-surface: #fbf1c7;
    --bg-elevated: #f2e5bc;
    --bg-overlay: #ebdbb2;
    --bg-hover: #d5c4a1;
    --bg-selected: #bdae93;
    --border-default: #d5c4a1;
    --border-strong: #bdae93;
    --border-focus: #b57614;
    --fg-primary: #3c3836;
    --fg-secondary: #504945;
    --fg-muted: #7c6f64;
    --fg-disabled: #928374;
    --fg-accent: #b57614;
    --fg-success: #79740e;
    --fg-warning: #af3a03;
    --fg-danger: #9d0006;
    --fg-info: #076678;
    --fg-link: #427b58;
    --fg-purple: #8f3f71;
    --accent: #b57614;
    --accent-hover: #d79921;
    --accent-fg: #fbf1c7;
    --accent-muted: #fabd2f;
  }

  html {
    @apply bg-bg-base text-fg;
    font-family: theme("fontFamily.sans");
    font-size: 13px;
  }

  body {
    @apply h-screen w-screen overflow-hidden;
  }
}
```

- [ ] **Step 4: Wire Tailwind into Vite**

Update `frontend/vite.config.ts` (renamed in Phase 0):
```ts
import tailwindcss from "@tailwindcss/vite";
// inside defineConfig.plugins
plugins: [sveltekit(), tailwindcss()],
```

- [ ] **Step 5: Import app.css from root layout**

`frontend/src/routes/+layout.svelte` (created in Phase 0) — add at top:
```svelte
<script lang="ts">
  import "../app.css";
</script>

<slot />
```

- [ ] **Step 6: Verify build**

```bash
cd frontend && npx vite build
```

Expected: build succeeds. Bundle size may increase ~20-50KB for Tailwind (PurgeCSS removes unused).

- [ ] **Step 7: Quality gate + commit**

```bash
git add frontend/{package.json,package-lock.json,tailwind.config.js,vite.config.ts,src/app.css,src/routes/+layout.svelte}
git commit -m "[chore] add tailwindcss with dark-first design tokens

[Body]
- Tailwind 3.x + @tailwindcss/vite plugin
- CSS variables for theme colors (bg/fg/border/accent scales)
- Dark mode default, light mode via [data-theme=light]
- Inter + JetBrains Mono typeface families
- bits-ui, lucide-svelte, mode-watcher installed for upcoming phases"
```

### Task 1.2: shadcn-svelte init

- [ ] **Step 1: Run shadcn-svelte init**

```bash
cd frontend && npx shadcn-svelte@latest init
```

Choose: TypeScript yes, Tailwind config detected, base color slate, components path `src/lib/components/ui`.

- [ ] **Step 2: Install essential primitives**

```bash
npx shadcn-svelte@latest add button dialog dropdown-menu tabs tooltip toast sheet context-menu scroll-area separator badge progress
```

- [ ] **Step 3: Verify imports work**

In `src/routes/+page.svelte` temporarily add:
```svelte
<script>
  import { Button } from "$lib/components/ui/button";
</script>
<Button variant="default">Test</Button>
```

Rebuild and visually confirm the button renders with theme tokens. Revert the test addition.

- [ ] **Step 4: Quality gate (build + svelte-check)**

- [ ] **Step 5: Invoke code-review skill on changed files (especially Tailwind config + app.css)**

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/components/ui/ frontend/components.json
git commit -m "[chore] add shadcn-svelte primitives (Button, Dialog, Tabs, Toast, etc.)"
```

### Task 1.3: Theme toggle + mode-watcher

- [ ] **Step 1: Wire mode-watcher**

Create `frontend/src/lib/components/ThemeToggle.svelte`:
```svelte
<script lang="ts">
  import { setMode, mode } from "mode-watcher";
  import Sun from "lucide-svelte/icons/sun";
  import Moon from "lucide-svelte/icons/moon";
</script>

<button
  class="p-1.5 rounded hover:bg-bg-hover transition-colors"
  onclick={() => setMode($mode === "dark" ? "light" : "dark")}
  aria-label="Toggle theme"
>
  {#if $mode === "dark"}
    <Sun size="16" />
  {:else}
    <Moon size="16" />
  {/if}
</button>
```

- [ ] **Step 2: Add `<ModeWatcher />` to `src/routes/+layout.svelte`** (initializes from system + persists to localStorage automatically)

- [ ] **Step 3: Quality gate + code-review + commit**

---

# PHASE 2: Global state stores + Svelte 5 runes migration (Day 3)

**Goal:** All shared state lives in `src/lib/store/`. `+page.svelte` and `+layout.svelte` stop holding `let openFiles`, `let opened_file`, etc. Components subscribe to stores instead of receiving via `bind:`.

**Done when:**
- `src/store/{tabs,uploads,files,ui,notifications}.ts` exist with typed APIs
- `auth.ts` migrated from old `store.ts` with Svelte 5 runes
- A new throwaway debug page renders state from each store correctly
- Tests still pass (Go backend tests untouched)

### Tasks
- **2.1** Refactor `store/store.ts` → `store/auth.ts` with Svelte 5 runes (`$state`, `$derived`) and explicit TS interfaces. Add `clearAuth()` helper.
- **2.2** Create `store/ui.ts` — theme (proxied to mode-watcher), sidebar collapsed (persist to localStorage), viewport breakpoint (derived from window.innerWidth).
- **2.3** Create `store/tabs.ts` — tab list, active tab id, helpers `openTab(tab)`, `closeTab(id)`, `setActive(id)`. Explorer tab is permanent.
- **2.4** Create `store/files.ts` — currentLoc (path[]), fileList, selection (Set<string>), viewMode ("grid" | "list"), sortBy, sortDir. Action: `setCurrentLoc(path)` triggers refetch via API.
- **2.5** Create `store/uploads.ts` — Upload list, helpers `enqueue(file, loc, name)`, `pause(id)`, `resume(id)`, `cancel(id)`. Hook to tus client (defined in Phase 5).
- **2.6** Create `store/notifications.ts` — toast queue, helpers `notify({ type, message, duration })`. Used by all error paths.
- **2.7** Migration smoke: replace the old `bind:` props in `+page.svelte` with store subscriptions; verify Explorer still works.

---

# PHASE 3: App shell — Header, VerticalNav, StatusBar (Day 4)

**Goal:** New persistent shell renders around all content. Section switching changes the main panel content but never unmounts header/nav/statusbar.

**Done when:**
- Shell components (Header / VerticalNav / StatusBar) compose inside `src/routes/+layout.svelte` so they persist across all routes (`+page.svelte`, nested routes)
- Sidebar shows 5 icons (Files / Users / Activity / Settings / System) — clicking switches main content
- Status bar shows mock upload toast (real wiring in Phase 5)
- Header has logo + theme toggle + user dropdown (avatar circle)
- Old `SideMenu.svelte`, `BottomMenu.svelte` deprecated and deleted in cleanup

### Tasks
- **3.1** `src/lib/components/Shell/Header.svelte` — logo + Ctrl+P quick open placeholder + ThemeToggle + user menu (DropdownMenu from shadcn)
- **3.2** `src/lib/components/Shell/VerticalNav.svelte` — icon-only nav, lucide icons, active indicator. Width 48px collapsed, 200px expanded (toggle via ui store)
- **3.3** `src/lib/components/Shell/StatusBar.svelte` — 28px height. Shows: active upload count + overall %, theme indicator, version
- **3.4** `src/routes/+layout.svelte` — CSS grid combining Header + VerticalNav + StatusBar + `<slot />` for `+page.svelte`. Auth guard runs in `+layout.ts` (redirect to /localLogin if no token)
- **3.5** Move shared types/interfaces into `$lib/types.ts` (User, Tab, Upload, FileEntry)
- **3.6** Delete `BottomMenu.svelte`, `SideMenu.svelte` (refactored into VerticalNav)
- **3.7** Phase 3 closeout: visual smoke via Playwright MCP — switch each nav item, confirm UI updates without flicker. Layout persists across section changes (only `<slot />` content swaps)

---

# PHASE 4: Tab system + viewer registry (Day 5-6)

**Goal:** Main content area is a tab strip. "Files" tab is permanent (Explorer). Opening a file from Explorer adds a new tab with the right viewer. Multiple files can be open simultaneously, switched without losing scroll position or upload state.

**Done when:**
- Tabs render with title, icon, close button, dirty indicator
- Click tab → switches main content
- Cmd/Ctrl+W closes active tab (unless permanent)
- Cmd/Ctrl+Tab cycles tabs
- Opening same file twice focuses existing tab (no duplicates)
- Text files open in Monaco with correct syntax (md, json, yaml, js, ts, py, go, sh, etc.)
- Image files open in zoomable viewer
- Video/audio open in <video>/<audio> with controls
- PDF files open in browser's native PDF viewer or [pdf.js](https://github.com/mozilla/pdf.js) embed
- Other extensions show a "Download" placeholder

### Tasks
- **4.1** `src/lib/Tabs/TabBar.svelte` — horizontal tab strip, scrollable, drag-to-reorder optional (defer)
- **4.2** `src/lib/Tabs/TabContent.svelte` — switch statement / component map by tab.kind
- **4.3** Viewers:
  - `src/lib/Viewers/ExplorerViewer.svelte` (wraps existing Explorer.svelte refactored)
  - `src/lib/Viewers/MonacoViewer.svelte` — lazy-loads monaco, applies dark+ theme, calls `/saveTextFile` on Ctrl+S, tracks dirty state
  - `src/lib/Viewers/ImageViewer.svelte` — fetches via `/getImageData?token=…&loc=&name=`, zoom controls
  - `src/lib/Viewers/MediaViewer.svelte` — video/audio with HTML5 controls, sources from `/getVideoData` / `/getAudioData`
  - `src/lib/Viewers/PdfViewer.svelte` — `<iframe>` or pdf.js
  - `src/lib/Viewers/UserManager.svelte` — refactored from existing
- **4.4** Viewer registry: `src/lib/Viewers/registry.ts` — `pickViewer(extension): TabKind`
- **4.5** Keyboard shortcuts: `src/lib/keyboard.ts` — global listener registers Cmd+P, Cmd+W, Cmd+Tab, Cmd+S
- **4.6** Phase 4 closeout: open 3 different file types simultaneously, switch between, confirm no state loss

---

# PHASE 5: Explorer redesign + Upload manager (Day 7-8)

**Goal:** File explorer is modern, accessible, and uploads persist globally.

**Done when:**
- Explorer shows grid + list view toggle
- Breadcrumb is clickable; each segment navigates
- Selection: single-click selects, Cmd-click adds to selection, Shift-click range
- Right-click (or long-press on touch) shows context menu (Open / Download / Rename / Copy / Move / Delete)
- Drag-and-drop files onto Explorer triggers upload
- Multiple uploads run concurrently (up to 3 by default)
- Pause/Resume/Cancel works per upload (tus protocol supports pause/resume server-side via offset)
- Upload toast at bottom shows aggregate progress; click toast expands into side panel listing each upload
- Navigating to another nav section (Users, Activity) does NOT stop uploads — they persist in `uploads` store
- Returning to Explorer shows uploads still in progress

### Tasks
- **5.1** `src/lib/Explorer/Breadcrumb.svelte` — chevron-separated path segments, last is non-clickable
- **5.2** `src/lib/Explorer/Toolbar.svelte` — upload button, new folder button, view toggle, search input, sort menu
- **5.3** `src/lib/Explorer/FileGrid.svelte` — responsive grid of file cards (icon, name, modified date)
- **5.4** `src/lib/Explorer/FileList.svelte` — table view with columns (icon, name, type, size, modified)
- **5.5** `src/lib/Explorer/ContextMenu.svelte` — uses shadcn-svelte ContextMenu, items dispatched to store actions
- **5.6** `src/lib/Explorer/DragDropOverlay.svelte` — fullscreen drop zone activates on dragenter, dispatches to upload store
- **5.7** `src/lib/tus-client.ts` — minimal tus 1.0.0 client (mirror of `scripts/tus-client.ps1` logic in TS): POST create, PATCH chunks, HEAD for resume. Uses fetch + AbortController for cancel. Yields progress events to store
- **5.8** `src/lib/Uploads/UploadToast.svelte` — collapsed bar in StatusBar showing aggregate
- **5.9** `src/lib/Uploads/UploadPanel.svelte` — right-side Sheet (shadcn) with per-upload row: filename, progress, pause/resume/cancel buttons
- **5.10** Wire `uploads` store events to tus-client; failure path notifies via `notifications` store
- **5.11** Phase 5 closeout: upload 3 files in parallel (one large), navigate away to Settings, return, confirm all 3 still progressing

---

# PHASE 6: Auth refactor + Kakao → Google OAuth (Day 9)

**Goal:** Login pages match new design system. Replace Kakao OAuth with Google.

**Done when:**
- Login screen shows Local Login form + (if enabled) Discord button + Google button
- New user registration flow polished (multi-step or single form, designer choice)
- Backend: `internal/auth/kakao.go` deleted, `internal/auth/google.go` added
- Frontend: `LoginRedirectKakao.svelte` deleted, `LoginRedirectGoogle.svelte` added
- `.env.example`, `docker-compose.yml`, `vite.config.ts` updated (KAKAO_* → GOOGLE_*)
- All existing tus + file features still work after auth refactor

### Tasks
- **6.1** Backend Google OAuth: copy `kakao.go` → `google.go`, swap URLs (`oauth2.googleapis.com/token`, `googleapis.com/oauth2/v2/userinfo`), update Config fields, register routes (`GET /googleLogin`, `POST /registerGoogle`)
- **6.2** Backend cleanup: delete `kakao.go`, remove routes, remove config fields
- **6.3** Frontend `src/routes/googleLogin/+page.svelte` — code exchange callback, store token, redirect to `/`. Uses `$page.url.searchParams` for the OAuth code
- **6.4** Refactor `src/routes/localLogin/+page.svelte` with new design system (centered card, accent button, error toast)
- **6.5** Refactor `src/routes/login/+page.svelte` (Discord) with new design system
- **6.6** Verify `LoginRedirectKakao.svelte` is already gone (deleted in Phase 0 migration)
- **6.7** Update `frontend/vite.config.ts` — KAKAO_* → GOOGLE_* environment variables
- **6.8** Update `docker-compose.yml`, `.env.example` — KAKAO_* → GOOGLE_*
- **6.9** Update Go integration tests if any Kakao references exist (probably none)
- **6.10** Phase 6 closeout: complete local register/login round-trip, Discord OAuth round-trip (if creds available), Google OAuth round-trip (manual with real Google Cloud project)

---

# PHASE 7: Admin / Activity / Settings / System refactor (Day 10)

**Goal:** Remaining sections adopt new design system. No more SCSS, all Tailwind.

### Tasks
- **7.1** `Account.svelte` → migrate to new design (user profile card, change password section, intent list)
- **7.2** `AccountViewer.svelte` → user list (admin only), redesigned with Table component
- **7.3** `UserManager.svelte` → user editor (intent toggles via checkboxes, save button)
- **7.4** `ActivityLog.svelte` → activity timeline (filter by user, type, date), virtualized list if rows > 100
- **7.5** `Setting.svelte` → settings page with sections (Account, Appearance, About). System info embedded
- **7.6** `SystemInfo.svelte` → 4 stat cards (CPU, Memory, Disk, Uptime). Polled every 5s via `/getSystemInfo`
- **7.7** Delete `SystemInfo_mobile.svelte`, `Explorer_mobile.svelte`, `FileViewer_mobile.svelte` (responsive achieved via Tailwind in single components)
- **7.8** Phase 7 closeout: visit each section, confirm rendering + interactivity

---

# PHASE 8: Mobile responsive + polish (Day 11)

**Goal:** Single set of components handles mobile + desktop. Animations smooth. Loading and error states present everywhere.

**Done when:**
- Sidebar collapses to hamburger on viewport < 768px
- Tab bar scrolls horizontally on narrow screens
- Touch: long-press triggers context menu
- All async actions show skeleton / spinner during fetch
- Error states show inline message with retry button
- Empty states have illustration or muted message

### Tasks
- **8.1** Audit each viewer + section for breakpoints (sm/md/lg)
- **8.2** Add `Sheet` (drawer) for mobile sidebar
- **8.3** Add skeleton components for FileGrid, ActivityLog, SystemInfo
- **8.4** Add transitions: tab open/close (fly), upload toast (slide), section change (fade)
- **8.5** Loading + error states for all `/server/*` calls
- **8.6** Phase 8 closeout: Playwright MCP visit at 375px, 768px, 1280px, 1920px — confirm no overflow / broken layout

---

# PHASE 9: Migration cutover + cleanup (Day 12)

**Goal:** Old code gone, new design lives, all features verified.

### Tasks
- **9.1** Delete any remaining `*.scss` in `src/lib/` (should be empty by now)
- **9.2** Remove unused npm packages: `sass`, anything else that became obsolete
- **9.3** Run `npx svelte-check` with no warnings allowed; fix any remaining
- **9.4** Run `npm run build`, check final bundle size — should be reasonable (Monaco lazy-loaded, total app bundle < 500KB excluding Monaco)
- **9.5** Playwright MCP E2E: register → login → upload (sm+lg) → preview → rename → delete → log out — full happy path
- **9.6** Tag `frontend-refactor-complete`
- **9.7** Update README screenshots
- **9.8** Phase 9 closeout: announce in commit message

---

## Out of Scope (Future Plans)

1. **Real-time collaboration** (multi-user concurrent edit on the same Monaco doc) — needs CRDT/OT
2. **Mobile native apps** — current responsive web is sufficient
3. **Custom Monaco themes / settings persistence** — deferred until users ask
4. **File comments / annotations** — Phase 2 plan
5. **Sharing UI** (FileBrowser Quantum-style share links) — Phase 2 plan, needs backend share table
6. **Trash bin** — Phase 2 plan, needs backend soft-delete

---

## Plan Retrospective

(Fill in after each phase)

### Phase 1 — TBD
- What surprised me:
- Adjustments for next phase:

### Phase 2 — TBD
- …

---

## Execution Handoff

After plan sign-off, two execution modes available:

**Subagent-Driven (recommended for radical pace)** — fresh subagent per task, two-stage review, parallel where possible. Many phases (3 and beyond) have independent components that parallelize well.

**Inline Execution** — execute in this session with checkpoints between Phases.
