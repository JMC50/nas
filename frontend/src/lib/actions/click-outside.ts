// `clickOutside` — Svelte 5 action that fires a callback when a click lands
// outside the node it's attached to. Used by ad-hoc popovers (overflow menus,
// dropdowns) that need to close on dismiss without owning an explicit
// document-level listener each time. The action attaches the listener on
// mount and removes it on destroy.
import type { Action } from "svelte/action";

export interface ClickOutsideParams {
  onOutside: () => void;
  // When false, the listener is detached — useful when the popover is closed
  // so the cost is zero while not open.
  enabled?: boolean;
}

export const clickOutside: Action<HTMLElement, ClickOutsideParams> = (node, initial) => {
  let params = initial;
  let attached = false;

  function onClick(event: MouseEvent) {
    if (!(event.target instanceof Node)) return;
    if (node.contains(event.target)) return;
    params.onOutside();
  }

  function sync() {
    const want = params.enabled !== false;
    if (want && !attached) {
      document.addEventListener("click", onClick);
      attached = true;
    } else if (!want && attached) {
      document.removeEventListener("click", onClick);
      attached = false;
    }
  }

  sync();

  return {
    update(next: ClickOutsideParams) {
      params = next;
      sync();
    },
    destroy() {
      if (attached) document.removeEventListener("click", onClick);
    },
  };
};
