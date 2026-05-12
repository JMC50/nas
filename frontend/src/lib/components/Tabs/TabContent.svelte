<script lang="ts">
  import { tabs } from "$lib/store/tabs.svelte";

  import Explorer from "$lib/components/Explorer.svelte";
  import UserManager from "$lib/components/UserManager.svelte";
  import Setting from "$lib/components/Setting.svelte";
  import Account from "$lib/components/Account.svelte";
  import ActivityLog from "$lib/components/ActivityLog.svelte";
  import SystemInfo from "$lib/components/SystemInfo.svelte";
  import ImageViewer from "$lib/components/Viewers/ImageViewer.svelte";
  import MediaViewer from "$lib/components/Viewers/MediaViewer.svelte";
  import PdfViewer from "$lib/components/Viewers/PdfViewer.svelte";
  import MonacoViewer from "$lib/components/Viewers/MonacoViewer.svelte";

  interface FilePayload {
    loc: string;
    name: string;
  }

  interface UserPayload {
    userId: string;
  }
</script>

<div class="h-full w-full overflow-hidden">
  {#each tabs.list as tab (tab.id)}
    {@const isActive = tab.id === tabs.activeId}
    <div class="h-full w-full {isActive ? 'block' : 'hidden'}">
      {#if tab.kind === "explorer"}
        <Explorer />
      {:else if tab.kind === "user-manager"}
        {@const payload = tab.payload as UserPayload | null}
        <UserManager initialUserId={payload?.userId ?? ""} />
      {:else if tab.kind === "settings"}
        <Setting />
      {:else if tab.kind === "activity"}
        <ActivityLog />
      {:else if tab.kind === "account"}
        <Account />
      {:else if tab.kind === "system"}
        <SystemInfo />
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
