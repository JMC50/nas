import { browser } from "$app/environment";
import type { Breakpoint } from "$lib/types";

const QUICKOPEN_DEFAULT = false;

const BREAKPOINTS: Record<Breakpoint, number> = {
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
};

class UIStore {
  quickOpenVisible = $state<boolean>(QUICKOPEN_DEFAULT);
  uploadsPanelOpen = $state<boolean>(false);
  drawerOpen = $state<boolean>(false);
  viewportWidth = $state<number>(0);
  viewportHeight = $state<number>(0);
  coarsePointer = $state<boolean>(false);

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
  // 가로로 든 휴대폰과 태블릿은 폭이 768 이상이어도 손가락으로 다룸
  touchLayout = $derived<boolean>(this.isMobile || this.coarsePointer);

  constructor() {
    if (browser) {
      this.viewportWidth = window.innerWidth;
      this.viewportHeight = window.innerHeight;
      const coarse = window.matchMedia("(pointer: coarse)");
      this.coarsePointer = coarse.matches;
      coarse.addEventListener("change", (event) => (this.coarsePointer = event.matches));
      window.addEventListener("resize", () => {
        this.viewportWidth = window.innerWidth;
        this.viewportHeight = window.innerHeight;
        // 1024px 이상에서는 레일이 서랍을 대신함
        if (this.viewportWidth >= BREAKPOINTS.lg && this.drawerOpen) {
          this.closeDrawer();
        }
      });
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

  private drawerOpener: HTMLElement | null = null;

  openDrawer(opener?: HTMLElement) {
    this.drawerOpener = opener ?? null;
    this.setDrawerOpen(true);
  }

  closeDrawer() {
    this.setDrawerOpen(false);
    this.drawerOpener?.focus();
    this.drawerOpener = null;
  }
}

export const ui = new UIStore();
