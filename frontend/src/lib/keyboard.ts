import { tabs } from "$lib/store/tabs.svelte";
import { ui } from "$lib/store/ui.svelte";

let installed = false;

function isModifier(event: KeyboardEvent) {
  return event.metaKey || event.ctrlKey;
}

function handler(event: KeyboardEvent) {
  if (!isModifier(event)) return;
  switch (event.key.toLowerCase()) {
    case "p":
      event.preventDefault();
      ui.openQuickOpen();
      return;
    case "w":
      if (tabs.active.closable) {
        event.preventDefault();
        tabs.close(tabs.activeId);
      }
      return;
    case "tab":
      event.preventDefault();
      if (event.shiftKey) tabs.prev();
      else tabs.next();
      return;
  }
}

export function installShortcuts() {
  if (installed || typeof window === "undefined") return;
  window.addEventListener("keydown", handler);
  installed = true;
}

export function uninstallShortcuts() {
  if (!installed || typeof window === "undefined") return;
  window.removeEventListener("keydown", handler);
  installed = false;
}
