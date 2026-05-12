<script lang="ts">
  import { tabs } from "$lib/store/tabs.svelte";

  import Explorer from "$lib/components/Explorer.svelte";
  import FileManager from "$lib/components/FileManager.svelte";
  import UserManager from "$lib/components/UserManager.svelte";
  import Setting from "$lib/components/Setting.svelte";
  import Account from "$lib/components/Account.svelte";
  import ActivityLog from "$lib/components/ActivityLog.svelte";
  import SystemInfo from "$lib/components/SystemInfo.svelte";
  import ImageViewer from "$lib/components/Viewers/ImageViewer.svelte";
  import MediaViewer from "$lib/components/Viewers/MediaViewer.svelte";
  import PdfViewer from "$lib/components/Viewers/PdfViewer.svelte";
  import MonacoViewer from "$lib/components/Viewers/MonacoViewer.svelte";
  import type { FileEntry, Intent } from "$lib/types";

  interface UserView {
    userId: string;
    username: string;
    krname: string;
    global_name: string;
    intents: Intent[];
  }

  interface FilePayload {
    loc: string;
    name: string;
  }

  // Legacy bindings preserved for Explorer/FileManager/Setting/UserManager which
  // still use export let. Phase 5/7 finishes their store migration.
  let currentPath: string[] = $state([]);
  let openFiles: string[] = $state([]);
  let opened_file: number = $state(0);
  let fileList: FileEntry[] = $state([]);
  let openUsers: string[] = $state([]);
  let opened_user: number = $state(0);
  let userList: UserView[] = $state([]);
  let selected: "folder" | "information" | "account" | "setting" | "log" = $state("folder");
</script>

<div class="h-full w-full overflow-hidden">
  {#each tabs.list as tab (tab.id)}
    {@const isActive = tab.id === tabs.activeId}
    <div class="h-full w-full {isActive ? 'block' : 'hidden'}">
      {#if tab.kind === "explorer"}
        <div class="grid grid-cols-[1fr_1fr] h-full">
          <div class="min-w-0 overflow-auto border-r border-border-default">
            <Explorer bind:openFiles bind:opened_file bind:fileList bind:currentPath />
          </div>
          <div class="min-w-0 overflow-auto">
            <FileManager bind:openFiles bind:opened_file bind:fileList />
          </div>
        </div>
      {:else if tab.kind === "user-manager"}
        <div class="grid grid-cols-[1fr_1fr] h-full">
          <div class="min-w-0 overflow-auto border-r border-border-default">
            <Setting bind:userList bind:opened_user bind:openUsers />
          </div>
          <div class="min-w-0 overflow-auto">
            <UserManager bind:openUsers bind:opened_user bind:userList />
          </div>
        </div>
      {:else if tab.kind === "settings"}
        <div class="h-full overflow-auto">
          <Setting bind:userList bind:opened_user bind:openUsers />
        </div>
      {:else if tab.kind === "activity"}
        <div class="h-full overflow-auto">
          <ActivityLog bind:currentPath bind:selected />
        </div>
      {:else if tab.kind === "account"}
        <div class="h-full overflow-auto">
          <Account />
        </div>
      {:else if tab.kind === "system"}
        <div class="h-full overflow-auto">
          <SystemInfo />
        </div>
      {:else if tab.kind === "image"}
        {@const payload = tab.payload as FilePayload}
        <ImageViewer loc={payload.loc} name={payload.name} />
      {:else if tab.kind === "video"}
        {@const payload = tab.payload as FilePayload}
        <MediaViewer loc={payload.loc} name={payload.name} kind="video" />
      {:else if tab.kind === "audio"}
        {@const payload = tab.payload as FilePayload}
        <MediaViewer loc={payload.loc} name={payload.name} kind="audio" />
      {:else if tab.kind === "pdf"}
        {@const payload = tab.payload as FilePayload}
        <PdfViewer loc={payload.loc} name={payload.name} />
      {:else if tab.kind === "text"}
        {@const payload = tab.payload as FilePayload}
        <MonacoViewer loc={payload.loc} name={payload.name} tabId={tab.id} />
      {/if}
    </div>
  {/each}
</div>
