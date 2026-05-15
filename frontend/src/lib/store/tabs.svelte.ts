import type { ExplorerPayload, Tab } from "$lib/types";
import { randomId } from "$lib/util/random-id";

const EXPLORER_TAB_ID = "explorer";

function locToTitle(loc: string[]): string {
  return loc.length === 0 ? "Files" : loc[loc.length - 1];
}

function locsEqual(a: string[], b: string[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
  return true;
}

function rootTab(): Tab {
  return {
    id: EXPLORER_TAB_ID,
    kind: "explorer",
    title: "Files",
    icon: "folder",
    payload: { loc: [] } as ExplorerPayload,
    closable: false,
  };
}

class TabStore {
  list = $state<Tab[]>([rootTab()]);
  activeId = $state<string>(EXPLORER_TAB_ID);

  active = $derived<Tab>(
    this.list.find((tab) => tab.id === this.activeId) ?? this.list[0],
  );

  open(input: Omit<Tab, "id" | "closable"> & { id?: string; closable?: boolean }) {
    const id = input.id ?? randomId();
    const existing = this.list.find((tab) => tab.id === id);
    if (existing) {
      this.activeId = existing.id;
      return existing;
    }
    const tab: Tab = {
      id,
      kind: input.kind,
      title: input.title,
      icon: input.icon,
      payload: input.payload,
      dirty: input.dirty,
      closable: input.closable ?? true,
    };
    this.list = [...this.list, tab];
    this.activeId = id;
    return tab;
  }

  openExplorer(loc: string[], options: { focus?: boolean } = {}): Tab {
    const existing = this.list.find((tab) => {
      if (tab.kind !== "explorer") return false;
      const payloadLoc = (tab.payload as ExplorerPayload | null)?.loc ?? [];
      return locsEqual(payloadLoc, loc);
    });
    if (existing) {
      if (options.focus !== false) this.activeId = existing.id;
      return existing;
    }
    return this.cloneExplorer(loc);
  }

  cloneExplorer(loc: string[]): Tab {
    return this.open({
      kind: "explorer",
      title: locToTitle(loc),
      icon: "folder",
      payload: { loc } as ExplorerPayload,
    });
  }

  update(id: string, partial: Partial<Pick<Tab, "title" | "payload" | "dirty">>) {
    this.list = this.list.map((tab) =>
      tab.id === id ? { ...tab, ...partial } : tab,
    );
  }

  close(id: string) {
    const target = this.list.find((tab) => tab.id === id);
    if (!target || !target.closable) return;
    const wasActive = this.activeId === id;
    const filtered = this.list.filter((tab) => tab.id !== id);
    this.list = filtered;
    if (wasActive) {
      this.activeId = filtered[filtered.length - 1]?.id ?? EXPLORER_TAB_ID;
    }
  }

  setActive(id: string) {
    if (this.list.some((tab) => tab.id === id)) {
      this.activeId = id;
    }
  }

  markDirty(id: string, dirty: boolean) {
    this.list = this.list.map((tab) =>
      tab.id === id ? { ...tab, dirty } : tab,
    );
  }

  rename(id: string, title: string) {
    this.list = this.list.map((tab) =>
      tab.id === id ? { ...tab, title } : tab,
    );
  }

  next() {
    const index = this.list.findIndex((tab) => tab.id === this.activeId);
    if (index === -1) return;
    const nextTab = this.list[(index + 1) % this.list.length];
    this.activeId = nextTab.id;
  }

  prev() {
    const index = this.list.findIndex((tab) => tab.id === this.activeId);
    if (index === -1) return;
    const prevTab = this.list[(index - 1 + this.list.length) % this.list.length];
    this.activeId = prevTab.id;
  }

  reorder(sourceId: string, targetId: string) {
    if (sourceId === targetId) return;
    const source = this.list.find((tab) => tab.id === sourceId);
    const target = this.list.find((tab) => tab.id === targetId);
    if (!source || !target) return;
    const filtered = this.list.filter((tab) => tab.id !== sourceId);
    const targetIndex = filtered.findIndex((tab) => tab.id === targetId);
    if (targetIndex === -1) return;
    filtered.splice(targetIndex, 0, source);
    this.list = filtered;
  }
}

export const tabs = new TabStore();
export { EXPLORER_TAB_ID };
