// Runes-based bridge over the legacy `useAuth` writable so new code can
// subscribe via reactivity (`auth.current.token`) without touching the
// legacy `subscribe`/`set` consumers. Phase 6 finishes the migration and
// retires `store.ts`.

import { get } from "svelte/store";
import { useAuth, Logout } from "./store";
import type { AuthUser } from "$lib/types";

const EMPTY: AuthUser = {
  userId: "",
  username: "",
  krname: "",
  global_name: "",
  token: "",
};

class AuthStore {
  current = $state<AuthUser>(EMPTY);
  isAdmin = $state<boolean>(false);

  isAuthenticated = $derived<boolean>(this.current.token.length > 0);
  token = $derived<string>(this.current.token);

  constructor() {
    this.current = get(useAuth);
    useAuth.subscribe((value: AuthUser) => {
      this.current = value;
    });
  }

  set(value: AuthUser) {
    useAuth.set(value);
  }

  clear() {
    this.isAdmin = false;
    Logout();
  }

  async refreshAdmin() {
    if (!this.token) {
      this.isAdmin = false;
      return;
    }
    try {
      const response = await fetch(`/server/checkAdmin?token=${encodeURIComponent(this.token)}`);
      if (!response.ok) {
        this.isAdmin = false;
        return;
      }
      const data = await response.json();
      this.isAdmin = data.isAdmin === true;
    } catch (cause) {
      console.warn("checkAdmin failed", cause);
      this.isAdmin = false;
    }
  }
}

export const auth = new AuthStore();
