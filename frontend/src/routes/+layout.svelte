<script lang="ts">
  import "../app.css";
  import { ModeWatcher } from "mode-watcher";
</script>

<ModeWatcher defaultMode="dark" darkClassNames={["dark"]} lightClassNames={["light"]} />

<slot />
