<script lang="ts">
  import logoMark from "$lib/assets/logo-mark.svg";

  let { size = 20 }: { size?: number } = $props();
</script>

<img src={logoMark} width={size} height={size} alt="" class="shrink-0" draggable="false" />
