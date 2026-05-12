<script lang="ts">
  import Explorer from "$lib/components/Explorer.svelte";
  import FileManager from "$lib/components/FileManager.svelte";
  import UserManager from "$lib/components/UserManager.svelte";
  import Setting from "$lib/components/Setting.svelte";
  import Account from "$lib/components/Account.svelte";
  import ActivityLog from "$lib/components/ActivityLog.svelte";
  import SystemInfo from "$lib/components/SystemInfo.svelte";
  import { tabs, EXPLORER_TAB_ID } from "$lib/store/tabs.svelte";
  import type { FileEntry, Intent, TabKind } from "$lib/types";

  interface UserView {
    userId: string;
    username: string;
    krname: string;
    global_name: string;
    intents: Intent[];
  }

  type Section = "folder" | "information" | "account" | "setting" | "log";

  // Legacy components (Explorer/FileManager/Setting/etc.) accept these as bind: props.
  // Phase 4 routes tabs.active directly to viewer components and retires the bridge.
  let currentPath: string[] = $state([]);
  let openFiles: string[] = $state([]);
  let opened_file: number = $state(0);
  let fileList: FileEntry[] = $state([]);
  let openUsers: string[] = $state([]);
  let opened_user: number = $state(0);
  let userList: UserView[] = $state([]);

  let selected: Section = $state("folder");

  const KIND_TO_SECTION: Partial<Record<TabKind, Section>> = {
    explorer: "folder",
    system: "information",
    account: "account",
    settings: "setting",
    activity: "log",
  };

  // tabs.active.kind → selected (one-way sync for legacy components).
  $effect(() => {
    selected = KIND_TO_SECTION[tabs.active.kind] ?? "folder";
  });

  // selected → tabs (only for ActivityLog's `selected = "folder"` back-nav).
  $effect(() => {
    if (selected === "folder" && tabs.active.kind !== "explorer") {
      tabs.setActive(EXPLORER_TAB_ID);
    }
  });
</script>

{#if tabs.active.kind === "explorer"}
  <div class="grid grid-cols-[1fr_1fr] h-full">
    <div class="min-w-0 overflow-auto border-r border-border-default">
      <Explorer bind:openFiles bind:opened_file bind:fileList bind:currentPath />
    </div>
    <div class="min-w-0 overflow-auto">
      <FileManager bind:openFiles bind:opened_file bind:fileList />
    </div>
  </div>
{:else if tabs.active.kind === "user-manager"}
  <div class="grid grid-cols-[1fr_1fr] h-full">
    <div class="min-w-0 overflow-auto border-r border-border-default">
      <Setting bind:userList bind:opened_user bind:openUsers />
    </div>
    <div class="min-w-0 overflow-auto">
      <UserManager bind:openUsers bind:opened_user bind:userList />
    </div>
  </div>
{:else if tabs.active.kind === "settings"}
  <div class="h-full overflow-auto">
    <Setting bind:userList bind:opened_user bind:openUsers />
  </div>
{:else if tabs.active.kind === "activity"}
  <div class="h-full overflow-auto">
    <ActivityLog bind:currentPath bind:selected />
  </div>
{:else if tabs.active.kind === "account"}
  <div class="h-full overflow-auto">
    <Account />
  </div>
{:else if tabs.active.kind === "system"}
  <div class="h-full overflow-auto">
    <SystemInfo />
  </div>
{/if}
