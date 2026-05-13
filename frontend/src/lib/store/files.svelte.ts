import type { ViewMode } from "$lib/types";

const VIEW_MODE_STORAGE_KEY = "nas:files:viewMode";

class FilesStore {
  viewMode = $state<ViewMode>("grid");

  constructor() {
    if (typeof localStorage !== "undefined") {
      const stored = localStorage.getItem(VIEW_MODE_STORAGE_KEY);
      if (stored === "grid" || stored === "list") {
        this.viewMode = stored;
      }
    }
  }

  setViewMode(mode: ViewMode) {
    this.viewMode = mode;
    if (typeof localStorage !== "undefined") {
      localStorage.setItem(VIEW_MODE_STORAGE_KEY, mode);
    }
  }
}

export const files = new FilesStore();
