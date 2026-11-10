import { auth } from "$lib/store/auth.svelte";
import { tabs } from "$lib/store/tabs.svelte";
import { signOutAddress } from "$lib/oidc";

export function displayName(): string {
  return auth.current.krname || auth.current.global_name || auth.current.username || "Guest";
}

export function openAccount(): void {
  tabs.open({
    id: "system:account",
    kind: "account",
    title: "Account",
    icon: "account",
    payload: null,
    closable: true,
  });
}

export function logout(): void {
  auth.clear();
  if (typeof window !== "undefined") {
    signOutAddress("/localLogin").then((address) => {
      window.location.href = address;
    });
  }
}
