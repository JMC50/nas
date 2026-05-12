<script lang="ts" module>
  const EXT_TO_LANGUAGE: Record<string, string> = {
    md: "markdown",
    ts: "typescript",
    tsx: "typescript",
    js: "javascript",
    jsx: "javascript",
    mjs: "javascript",
    cjs: "javascript",
    json: "json",
    yaml: "yaml",
    yml: "yaml",
    py: "python",
    go: "go",
    rs: "rust",
    java: "java",
    kt: "kotlin",
    swift: "swift",
    c: "c",
    cpp: "cpp",
    cc: "cpp",
    cs: "csharp",
    h: "c",
    hpp: "cpp",
    rb: "ruby",
    php: "php",
    sh: "shell",
    bash: "shell",
    zsh: "shell",
    ps1: "powershell",
    sql: "sql",
    html: "html",
    htm: "html",
    xml: "xml",
    svg: "xml",
    css: "css",
    scss: "scss",
    sass: "scss",
    less: "less",
    vue: "html",
    svelte: "html",
    dockerfile: "dockerfile",
    makefile: "makefile",
  };

  function detectLanguage(filename: string): string {
    const ext = filename.split(".").pop()?.toLowerCase() ?? "";
    return EXT_TO_LANGUAGE[ext] ?? "plaintext";
  }
</script>

<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { tabs } from "$lib/store/tabs.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  import Save from "lucide-svelte/icons/save";

  interface Props {
    loc: string;
    name: string;
    tabId: string;
  }

  let { loc, name, tabId }: Props = $props();

  let container: HTMLDivElement;
  let editor: import("monaco-editor").editor.IStandaloneCodeEditor | null = null;
  let initialContent = "";
  let loading = $state(true);
  let saving = $state(false);

  const language = $derived(detectLanguage(name));

  async function loadEditor() {
    const [{ default: loader }, fileResponse] = await Promise.all([
      import("@monaco-editor/loader"),
      fetch(
        `/server/getTextFile?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
      ),
    ]);

    if (!fileResponse.ok) {
      notifications.error(`Failed to load ${name}: ${fileResponse.status}`);
      loading = false;
      return;
    }

    initialContent = await fileResponse.text();

    const monaco = await loader.init();
    defineGruvboxTheme(monaco);

    const instance = monaco.editor.create(container, {
      value: initialContent,
      language,
      theme: "gruvbox-dark",
      automaticLayout: true,
      fontSize: 13,
      fontFamily: "'JetBrains Mono', ui-monospace, monospace",
      minimap: { enabled: true },
      scrollBeyondLastLine: false,
      tabSize: 2,
    });

    instance.onDidChangeModelContent(() => {
      tabs.markDirty(tabId, instance.getValue() !== initialContent);
    });

    editor = instance;

    loading = false;
  }

  function defineGruvboxTheme(monaco: typeof import("monaco-editor")) {
    monaco.editor.defineTheme("gruvbox-dark", {
      base: "vs-dark",
      inherit: true,
      rules: [
        { token: "comment", foreground: "928374", fontStyle: "italic" },
        { token: "keyword", foreground: "fb4934" },
        { token: "string", foreground: "b8bb26" },
        { token: "number", foreground: "d3869b" },
        { token: "type", foreground: "fabd2f" },
        { token: "function", foreground: "8ec07c" },
      ],
      colors: {
        "editor.background": "#1d2021",
        "editor.foreground": "#ebdbb2",
        "editor.lineHighlightBackground": "#282828",
        "editorLineNumber.foreground": "#665c54",
        "editorLineNumber.activeForeground": "#fabd2f",
        "editorCursor.foreground": "#fabd2f",
        "editor.selectionBackground": "#504945",
        "editor.inactiveSelectionBackground": "#3c3836",
        "editorIndentGuide.background": "#3c3836",
        "editorIndentGuide.activeBackground": "#504945",
        "editorGutter.background": "#1d2021",
      },
    });
  }

  async function save() {
    if (!editor) return;
    const content = editor.getValue();
    saving = true;
    try {
      const response = await fetch(
        `/server/saveTextFile?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
        {
          method: "POST",
          headers: { "Content-Type": "text/plain; charset=utf-8" },
          body: content,
        },
      );
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      initialContent = content;
      tabs.markDirty(tabId, false);
      notifications.success(`Saved ${name}`, 2500);
    } catch (error) {
      notifications.error(`Save failed: ${(error as Error).message}`);
    } finally {
      saving = false;
    }
  }

  function onKeyDown(event: KeyboardEvent) {
    if ((event.metaKey || event.ctrlKey) && event.key === "s") {
      event.preventDefault();
      save();
    }
  }

  onMount(() => {
    loadEditor();
    window.addEventListener("keydown", onKeyDown);
  });

  onDestroy(() => {
    window.removeEventListener("keydown", onKeyDown);
    editor?.dispose();
    editor = null;
  });
</script>

<div class="flex flex-col h-full w-full bg-bg-base">
  <div class="flex items-center gap-3 px-3 h-9 border-b border-border-default text-xs text-fg-secondary">
    <span class="truncate">{name}</span>
    <span class="text-fg-muted">{language}</span>
    <button
      type="button"
      class="ml-auto inline-flex items-center gap-1.5 px-2 h-7 rounded text-fg-primary hover:bg-bg-hover disabled:opacity-50"
      onclick={save}
      disabled={saving}
      aria-label="Save (Ctrl+S)"
    >
      <Save size="12" />
      <span>{saving ? "Saving…" : "Save"}</span>
    </button>
  </div>
  <div class="flex-1 relative">
    {#if loading}
      <div class="absolute inset-0 flex items-center justify-center text-fg-muted text-xs">
        Loading editor…
      </div>
    {/if}
    <div bind:this={container} class="absolute inset-0"></div>
  </div>
</div>
