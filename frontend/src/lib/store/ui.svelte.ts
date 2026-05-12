import { browser } from "$app/environment";
import type { Breakpoint } from "$lib/types";

const SIDEBAR_STORAGE_KEY = "nas:ui:sidebarCollapsed";
const QUICKOPEN_DEFAULT = false;

const BREAKPOINTS: Record<Breakpoint, number> = {
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
};

class UIStore {
  sidebarCollapsed = $state<boolean>(false);
  quickOpenVisible = $state<boolean>(QUICKOPEN_DEFAULT);
  uploadsPanelOpen = $state<boolean>(false);
  viewportWidth = $state<number>(0);
  viewportHeight = $state<number>(0);

  breakpoint = $derived<Breakpoint>(
    this.viewportWidth >= BREAKPOINTS.xl
      ? "xl"
      : this.viewportWidth >= BREAKPOINTS.lg
        ? "lg"
        : this.viewportWidth >= BREAKPOINTS.md
          ? "md"
          : "sm",
  );

  isMobile = $derived<boolean>(this.viewportWidth < BREAKPOINTS.md);

  constructor() {
    if (browser) {
      this.sidebarCollapsed = localStorage.getItem(SIDEBAR_STORAGE_KEY) === "1";
      this.viewportWidth = window.innerWidth;
      this.viewportHeight = window.innerHeight;
      window.addEventListener("resize", () => {
        this.viewportWidth = window.innerWidth;
        this.viewportHeight = window.innerHeight;
      });
    }
  }

  toggleSidebar() {
    this.sidebarCollapsed = !this.sidebarCollapsed;
    if (browser) {
      localStorage.setItem(SIDEBAR_STORAGE_KEY, this.sidebarCollapsed ? "1" : "0");
    }
  }

  openQuickOpen() {
    this.quickOpenVisible = true;
  }

  closeQuickOpen() {
    this.quickOpenVisible = false;
  }

  toggleUploadsPanel() {
    this.uploadsPanelOpen = !this.uploadsPanelOpen;
  }

  closeUploadsPanel() {
    this.uploadsPanelOpen = false;
  }
}

export const ui = new UIStore();
