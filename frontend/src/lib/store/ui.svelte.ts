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
  drawerOpen = $state<boolean>(false);
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
        // Auto-close drawer when upsizing past md — the drawer is mobile-only
        // and the sidebar grid column reappears at md+.
        if (this.viewportWidth >= BREAKPOINTS.md && this.drawerOpen) {
          this.closeDrawer();
        }
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

  // Mobile drawer: open/close mutates state AND locks <html> scroll to
  // prevent body-scroll bleed under the overlay.
  setDrawerOpen(value: boolean) {
    this.drawerOpen = value;
    if (browser) {
      document.documentElement.style.overflow = value ? "hidden" : "";
    }
  }

  openDrawer() {
    this.setDrawerOpen(true);
  }

  closeDrawer() {
    this.setDrawerOpen(false);
  }

  toggleDrawer() {
    this.setDrawerOpen(!this.drawerOpen);
  }
}

export const ui = new UIStore();
