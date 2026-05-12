<script lang="ts">
  import { onMount } from "svelte";
  import { useAuth } from "$lib/store/store";
  import HardDrive from "lucide-svelte/icons/hard-drive";
  import LogIn from "lucide-svelte/icons/log-in";
  import UserPlus from "lucide-svelte/icons/user-plus";

  interface PasswordRequirements {
    minLength: number;
    requireUppercase: boolean;
    requireLowercase: boolean;
    requireNumber: boolean;
    requireSpecial: boolean;
  }

  interface AuthConfig {
    authType: "oauth" | "local" | "both";
    localAuthEnabled: boolean;
    oauthEnabled: boolean;
    passwordRequirements: PasswordRequirements;
  }

  let authConfig: AuthConfig | null = $state(null);
  let mode: "login" | "register" = $state("login");
  let loading = $state(false);
  let error = $state("");

  let userId = $state("");
  let username = $state("");
  let password = $state("");
  let confirmPassword = $state("");
  let koreanName = $state("");

  onMount(async () => {
    try {
      const response = await fetch("/server/auth/config");
      authConfig = await response.json();
    } catch (err) {
      error = `Failed to load config: ${(err as Error).message}`;
    }
  });

  function validatePassword(pwd: string): string | null {
    if (!authConfig) return null;
    const reqs = authConfig.passwordRequirements;
    if (pwd.length < reqs.minLength) {
      return `Password must be at least ${reqs.minLength} characters.`;
    }
    if (reqs.requireUppercase && !/[A-Z]/.test(pwd)) {
      return "Password needs at least one uppercase letter.";
    }
    if (reqs.requireLowercase && !/[a-z]/.test(pwd)) {
      return "Password needs at least one lowercase letter.";
    }
    if (reqs.requireNumber && !/[0-9]/.test(pwd)) {
      return "Password needs at least one number.";
    }
    if (reqs.requireSpecial && !/[!@#$%^&*(),.?":{}|<>]/.test(pwd)) {
      return "Password needs at least one special character.";
    }
    return null;
  }

  async function handleLogin() {
    error = "";
    loading = true;
    try {
      const response = await fetch("/server/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ userId, password }),
      });
      const data = await response.json();
      if (data.success) {
        useAuth.set({
          userId: data.user.userId,
          username: data.user.username,
          krname: data.user.krname ?? "",
          global_name: data.user.global_name,
          token: data.token,
        });
        const baseUrl = `${window.location.protocol}//${window.location.host}/`;
        window.location.replace(baseUrl);
      } else {
        error = data.message ?? "Login failed.";
      }
    } catch {
      error = "Network error. Please try again.";
    } finally {
      loading = false;
    }
  }

  async function handleRegister() {
    error = "";
    if (!userId || !username || !password) {
      error = "Fill in all required fields.";
      return;
    }
    if (password !== confirmPassword) {
      error = "Passwords do not match.";
      return;
    }
    const violation = validatePassword(password);
    if (violation) {
      error = violation;
      return;
    }
    loading = true;
    try {
      const response = await fetch("/server/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ userId, username, password, krname: koreanName }),
      });
      const data = await response.json();
      if (data.success) {
        useAuth.set({
          userId: data.user.userId,
          username: data.user.username,
          krname: koreanName,
          global_name: data.user.global_name,
          token: data.token,
        });
        const baseUrl = `${window.location.protocol}//${window.location.host}/`;
        window.location.replace(baseUrl);
      } else {
        error = data.message ?? "Registration failed.";
      }
    } catch {
      error = "Network error. Please try again.";
    } finally {
      loading = false;
    }
  }

  function switchMode() {
    mode = mode === "login" ? "register" : "login";
    error = "";
    userId = "";
    username = "";
    password = "";
    confirmPassword = "";
    koreanName = "";
  }

  function submit(event: SubmitEvent) {
    event.preventDefault();
    if (mode === "login") handleLogin();
    else handleRegister();
  }
</script>

<main class="min-h-screen flex items-center justify-center bg-bg-base p-6">
  <div class="w-full max-w-md bg-bg-surface border border-border-default rounded-lg p-6 shadow-[0_4px_12px_rgba(0,0,0,0.4)]">
    <div class="flex items-center gap-2 mb-6">
      <HardDrive size="20" class="text-accent" />
      <span class="text-base font-semibold text-fg-primary tracking-tight">NAS</span>
    </div>

    <h1 class="text-base font-semibold text-fg-primary mb-1">
      {mode === "login" ? "Sign in" : "Create account"}
    </h1>
    <p class="text-xs text-fg-muted mb-5">
      {mode === "login"
        ? "Use your ID and password to continue."
        : "Pick an ID and password to register."}
    </p>

    {#if !authConfig}
      <div class="text-xs text-fg-muted">Loading configuration…</div>
    {:else if !authConfig.localAuthEnabled}
      <div class="p-3 rounded-md bg-fg-warning/10 border border-fg-warning/30 text-fg-warning text-xs">
        Local authentication is disabled. Use an OAuth provider instead.
      </div>
    {:else}
      {#if error}
        <div class="mb-4 p-3 rounded-md bg-fg-danger/10 border border-fg-danger/30 text-fg-danger text-xs">
          {error}
        </div>
      {/if}

      <form class="space-y-3" onsubmit={submit}>
        <div>
          <label for="userId" class="block text-[11px] text-fg-muted mb-1">User ID</label>
          <input
            id="userId"
            type="text"
            bind:value={userId}
            placeholder="your-id"
            disabled={loading}
            required
            class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-primary text-sm focus:border-border-focus outline-none disabled:opacity-60"
          />
        </div>

        {#if mode === "register"}
          <div>
            <label for="username" class="block text-[11px] text-fg-muted mb-1">Username</label>
            <input
              id="username"
              type="text"
              bind:value={username}
              placeholder="Display name"
              disabled={loading}
              required
              class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-primary text-sm focus:border-border-focus outline-none disabled:opacity-60"
            />
          </div>
          <div>
            <label for="krname-reg" class="block text-[11px] text-fg-muted mb-1">Korean name (optional)</label>
            <input
              id="krname-reg"
              type="text"
              bind:value={koreanName}
              placeholder="홍길동"
              disabled={loading}
              class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-primary text-sm focus:border-border-focus outline-none disabled:opacity-60"
            />
          </div>
        {/if}

        <div>
          <label for="password" class="block text-[11px] text-fg-muted mb-1">Password</label>
          <input
            id="password"
            type="password"
            bind:value={password}
            disabled={loading}
            required
            class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-primary text-sm focus:border-border-focus outline-none disabled:opacity-60"
          />
        </div>

        {#if mode === "register"}
          <div>
            <label for="confirmPassword" class="block text-[11px] text-fg-muted mb-1">Confirm password</label>
            <input
              id="confirmPassword"
              type="password"
              bind:value={confirmPassword}
              disabled={loading}
              required
              class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-primary text-sm focus:border-border-focus outline-none disabled:opacity-60"
            />
          </div>

          <div class="p-3 rounded-md bg-bg-elevated border border-border-default">
            <div class="text-[11px] text-fg-muted mb-1.5">Password requirements</div>
            <ul class="text-xs text-fg-secondary space-y-0.5">
              <li>• At least {authConfig.passwordRequirements.minLength} characters</li>
              {#if authConfig.passwordRequirements.requireUppercase}<li>• Uppercase letter</li>{/if}
              {#if authConfig.passwordRequirements.requireLowercase}<li>• Lowercase letter</li>{/if}
              {#if authConfig.passwordRequirements.requireNumber}<li>• Number</li>{/if}
              {#if authConfig.passwordRequirements.requireSpecial}<li>• Special character</li>{/if}
            </ul>
          </div>
        {/if}

        <button
          type="submit"
          class="w-full h-10 inline-flex items-center justify-center gap-2 rounded-md bg-accent text-accent-fg text-sm font-semibold hover:bg-accent-hover transition-colors disabled:opacity-60"
          disabled={loading}
        >
          {#if mode === "login"}
            <LogIn size="14" />
            {loading ? "Signing in…" : "Sign in"}
          {:else}
            <UserPlus size="14" />
            {loading ? "Creating…" : "Create account"}
          {/if}
        </button>

        <button
          type="button"
          class="w-full text-xs text-fg-muted hover:text-fg-link mt-2 transition-colors"
          onclick={switchMode}
        >
          {mode === "login" ? "Need an account? Register" : "Already registered? Sign in"}
        </button>
      </form>
    {/if}
  </div>
</main>
