// Sidebar navigation entries — shared between the desktop VerticalNav
// (Shell/VerticalNav.svelte) and the mobile drawer (Shell/MobileDrawer.svelte)
// so both surface the same items with identical activation semantics.
import Folder from "lucide-svelte/icons/folder";
import Music from "lucide-svelte/icons/music";
import Film from "lucide-svelte/icons/film";
import Users from "lucide-svelte/icons/users";
import History from "lucide-svelte/icons/history";
import Settings from "lucide-svelte/icons/settings";
import Cpu from "lucide-svelte/icons/cpu";
import { tabs, EXPLORER_TAB_ID } from "$lib/store/tabs.svelte";
import type { TabKind } from "$lib/types";

export interface NavItem {
  id: string;
  kind: TabKind;
  title: string;
  icon: typeof Folder;
  adminOnly?: boolean;
}

export const NAV_ITEMS: NavItem[] = [
  { id: EXPLORER_TAB_ID, kind: "explorer", title: "Files", icon: Folder },
  { id: "system:music-library", kind: "music-library", title: "Music", icon: Music },
  { id: "system:video-library", kind: "video-library", title: "Videos", icon: Film },
  { id: "system:user-manager", kind: "user-manager", title: "Users", icon: Users, adminOnly: true },
  { id: "system:activity", kind: "activity", title: "Activity", icon: History },
  { id: "system:settings", kind: "settings", title: "Settings", icon: Settings, adminOnly: true },
  { id: "system:system", kind: "system", title: "System", icon: Cpu, adminOnly: true },
];

// Sidebar entries that map to singleton tabs the user cannot close — the
// sidebar button is the only way back in.
const NON_CLOSABLE: TabKind[] = ["explorer", "music-library", "video-library"];

export function activate(item: NavItem): void {
  const existing = tabs.list.find((tab) => tab.id === item.id);
  if (existing) {
    tabs.setActive(item.id);
    return;
  }
  tabs.open({
    id: item.id,
    kind: item.kind,
    title: item.title,
    icon: item.kind,
    payload: null,
    closable: !NON_CLOSABLE.includes(item.kind),
  });
}
