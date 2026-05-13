<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";
  import PdfViewer from "./PdfViewer.svelte";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  const officeUrl = $derived(
    `/server/getOfficePdf?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );
</script>

<PdfViewer {loc} {name} urlOverride={officeUrl} />
