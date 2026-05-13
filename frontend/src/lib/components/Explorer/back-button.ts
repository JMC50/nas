import { onMount, onDestroy } from "svelte";
import { tabs } from "$lib/store/tabs.svelte";

const X1_BUTTON = 3;
const EDITABLE_SELECTOR = "input, textarea, [contenteditable='true']";

// Listens on auxclick only (fires once per X1 click on same target).
// Combining mouseup + auxclick would fire twice per click.
export function useBackButton(tabId: string, onBack: () => void) {
  function handle(event: MouseEvent) {
    if (event.button !== X1_BUTTON) return;
    if (tabs.activeId !== tabId) return;
    const target = event.target as HTMLElement | null;
    if (target?.matches?.(EDITABLE_SELECTOR)) return;
    event.preventDefault();
    onBack();
  }

  onMount(() => window.addEventListener("auxclick", handle));
  onDestroy(() => window.removeEventListener("auxclick", handle));
}
