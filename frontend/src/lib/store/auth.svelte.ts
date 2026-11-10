// Runes-based bridge over the legacy `useAuth` writable so new code can
// subscribe via reactivity (`auth.current.token`) without touching the
// legacy `subscribe`/`set` consumers. Phase 6 finishes the migration and
// retires `store.ts`.

import { get } from "svelte/store";
import { useAuth, Logout } from "./store";
import type { AuthUser } from "$lib/types";
import { oidcAddresses } from "$lib/oidc";

const SIGN_IN_AGAIN_KEY = "nas.oidc.signInAgain";
const SIGN_IN_AGAIN_GAP_MS = 60_000;

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
        if (response.status === 401) await this.signInAgain();
        return;
      }
      const data = await response.json();
      this.isAdmin = data.isAdmin === true;
    } catch (cause) {
      console.warn("checkAdmin failed", cause);
      this.isAdmin = false;
    }
  }

  // In single sign-on mode a rejected token is replaced by a new sign-in.
  // A second rejection within a minute stops at the signed-out page, so a
  // server that rejects every new token cannot cause a redirect loop.
  private async signInAgain() {
    const oidc = await oidcAddresses();
    if (!oidc) return;
    this.clear();
    const last = Number(sessionStorage.getItem(SIGN_IN_AGAIN_KEY) ?? 0);
    sessionStorage.setItem(SIGN_IN_AGAIN_KEY, String(Date.now()));
    const looping = Date.now() - last < SIGN_IN_AGAIN_GAP_MS;
    window.location.replace(looping ? "/localLogin?signedout=1" : oidc.loginUrl);
  }
}

export const auth = new AuthStore();
