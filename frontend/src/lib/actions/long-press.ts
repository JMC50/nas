// `longPress` — Svelte 5 action that fires a callback after the user holds
// down on the node for `delayMs` (default 500ms) without moving the pointer
// more than `moveTolerance` pixels (default 8px). Used to replace right-click
// for context menus on touch devices where the browser has no `contextmenu`
// gesture by default. Callback receives the pointer position so consumers can
// open a menu anchored under the finger.
import type { Action } from "svelte/action";

export interface LongPressParams {
  onLongPress: (clientX: number, clientY: number) => void;
  delayMs?: number;
  moveTolerance?: number;
}

const DEFAULT_DELAY_MS = 500;
const DEFAULT_MOVE_TOLERANCE_PX = 8;

export const longPress: Action<HTMLElement, LongPressParams> = (node, initial) => {
  let params = initial;
  let timer: number | null = null;
  let startX = 0;
  let startY = 0;

  function cancel() {
    if (timer !== null) {
      window.clearTimeout(timer);
      timer = null;
    }
  }

  function onPointerDown(event: PointerEvent) {
    // Ignore non-primary buttons (right click etc) — those have their own
    // semantics via contextmenu.
    if (event.button !== 0) return;
    cancel();
    startX = event.clientX;
    startY = event.clientY;
    const delay = params.delayMs ?? DEFAULT_DELAY_MS;
    timer = window.setTimeout(() => {
      timer = null;
      params.onLongPress(startX, startY);
    }, delay);
  }

  function onPointerMove(event: PointerEvent) {
    if (timer === null) return;
    const tolerance = params.moveTolerance ?? DEFAULT_MOVE_TOLERANCE_PX;
    const dx = event.clientX - startX;
    const dy = event.clientY - startY;
    if (dx * dx + dy * dy > tolerance * tolerance) cancel();
  }

  node.addEventListener("pointerdown", onPointerDown);
  node.addEventListener("pointermove", onPointerMove);
  node.addEventListener("pointerup", cancel);
  node.addEventListener("pointercancel", cancel);
  node.addEventListener("pointerleave", cancel);

  return {
    update(next: LongPressParams) {
      params = next;
    },
    destroy() {
      cancel();
      node.removeEventListener("pointerdown", onPointerDown);
      node.removeEventListener("pointermove", onPointerMove);
      node.removeEventListener("pointerup", cancel);
      node.removeEventListener("pointercancel", cancel);
      node.removeEventListener("pointerleave", cancel);
    },
  };
};
