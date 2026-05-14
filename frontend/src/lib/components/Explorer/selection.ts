// Pure helpers for multi-select state. Selection is an ordered list of entry
// names: the LAST item is the range-anchor used by Shift+click. Stored in
// `ExplorerPayload.selection` so each Explorer tab keeps its own selection.

/**
 * Toggle `name` in the selection. If absent, append (becomes new anchor).
 * If present, remove (anchor falls back to whatever ends up last, or none).
 */
export function toggleEntry(selection: string[], name: string): string[] {
  return selection.includes(name)
    ? selection.filter((n) => n !== name)
    : [...selection, name];
}

/** Replace selection with a single entry; becomes the new range anchor. */
export function setSingle(name: string): string[] {
  return [name];
}

/**
 * Range select from `anchor` to `target` within `allNames`. The anchor stays
 * at the END of the returned list so it remains the range-anchor for further
 * Shift+click operations.
 *
 * Spec walkthrough (allNames = [a,b,c,d,e]):
 *  - anchor=b, target=d → [b,c,d]   (ai=1, ti=3, forward, slice is [b,c,d], anchor already last)
 *  - anchor=d, target=b → [d,c,b]   (ai=3, ti=1, reverse, slice [b,c,d] reversed → [d,c,b], anchor last)
 *  - anchor=b, target=b → [b]       (degenerate, slice is just [b])
 *  - anchor=x not in list → [target] (recover by single-selecting target)
 */
export function selectRange(allNames: string[], anchor: string, target: string): string[] {
  const ai = allNames.indexOf(anchor);
  const ti = allNames.indexOf(target);
  if (ai === -1 || ti === -1) return [target];
  const [from, to] = ai <= ti ? [ai, ti] : [ti, ai];
  const slice = allNames.slice(from, to + 1);
  return ai <= ti ? slice : slice.reverse();
}

/** Clear selection. */
export function clear(): string[] {
  return [];
}
