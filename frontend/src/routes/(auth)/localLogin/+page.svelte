<script lang="ts">
  import { onMount } from "svelte";
  import HardDrive from "lucide-svelte/icons/hard-drive";
  import LogIn from "lucide-svelte/icons/log-in";
  import UserPlus from "lucide-svelte/icons/user-plus";
  import TextField from "$lib/components/Auth/TextField.svelte";
  import PasswordRules from "$lib/components/Auth/PasswordRules.svelte";
  import SignInOptions from "$lib/components/Auth/SignInOptions.svelte";
  import {
    signInLocal,
    signUpLocal,
    discordUrl,
    googleUrl,
  } from "$lib/auth-actions";

  interface Rules {
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
    passwordRequirements: Rules;
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
    } catch (cause) {
      error = `Failed to load config: ${(cause as Error).message}`;
    }
  });

  function validate(value: string): string | null {
    if (!authConfig) return null;
    const rules = authConfig.passwordRequirements;
    if (value.length < rules.minLength) return `Password must be at least ${rules.minLength} characters.`;
    if (rules.requireUppercase && !/[A-Z]/.test(value)) return "Password needs at least one uppercase letter.";
    if (rules.requireLowercase && !/[a-z]/.test(value)) return "Password needs at least one lowercase letter.";
    if (rules.requireNumber && !/[0-9]/.test(value)) return "Password needs at least one number.";
    if (rules.requireSpecial && !/[!@#$%^&*(),.?":{}|<>]/.test(value)) return "Password needs at least one special character.";
    return null;
  }

  async function onSignIn() {
    error = "";
    loading = true;
    const result = await signInLocal({ userId, password });
    if (!result.success) error = result.message ?? "";
    loading = false;
  }

  async function onRegister() {
    error = "";
    if (!userId || !username || !password) { error = "Fill in all required fields."; return; }
    if (password !== confirmPassword) { error = "Passwords do not match."; return; }
    const violation = validate(password);
    if (violation) { error = violation; return; }
    loading = true;
    const result = await signUpLocal({ userId, username, password, krname: koreanName });
    if (!result.success) error = result.message ?? "";
    loading = false;
  }

  function switchMode() {
    mode = mode === "login" ? "register" : "login";
    error = "";
    userId = ""; username = ""; password = ""; confirmPassword = ""; koreanName = "";
  }

  function submit(event: SubmitEvent) {
    event.preventDefault();
    if (mode === "login") onSignIn();
    else onRegister();
  }

  function goDiscord() {
    const target = discordUrl();
    if (target) location.href = target;
  }

  function goGoogle() {
    const target = googleUrl();
    if (target) location.href = target;
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
      {mode === "login" ? "Use your ID and password to continue." : "Pick an ID and password to register."}
    </p>

    {#if !authConfig}
      <div class="text-xs text-fg-muted">Loading configuration…</div>
    {:else if !authConfig.oauthEnabled && !authConfig.localAuthEnabled}
      <div class="p-3 rounded-md bg-fg-warning/10 border border-fg-warning/30 text-fg-warning text-xs">
        No authentication methods enabled.
      </div>
    {:else}
      {#if authConfig.oauthEnabled}
        <SignInOptions oauthEnabled={true} localEnabled={false} onDiscord={goDiscord} onGoogle={goGoogle} onLocal={() => {}} />
        {#if authConfig.localAuthEnabled}
          <div class="flex items-center gap-3 my-4 text-xs text-fg-muted">
            <div class="flex-1 h-px bg-border-default"></div>
            <span>or</span>
            <div class="flex-1 h-px bg-border-default"></div>
          </div>
        {/if}
      {/if}

      {#if authConfig.localAuthEnabled}
        {#if error}
          <div class="mb-4 p-3 rounded-md bg-fg-danger/10 border border-fg-danger/30 text-fg-danger text-xs">{error}</div>
        {/if}

        <form class="space-y-3" onsubmit={submit}>
          <TextField id="userId" label="User ID" value={userId} placeholder="your-id" disabled={loading} required onInput={(v) => (userId = v)} />

          {#if mode === "register"}
            <TextField id="username" label="Username" value={username} placeholder="Display name" disabled={loading} required onInput={(v) => (username = v)} />
            <TextField id="krname-reg" label="Korean name (optional)" value={koreanName} placeholder="홍길동" disabled={loading} onInput={(v) => (koreanName = v)} />
          {/if}

          <TextField id="password" label="Password" type="password" value={password} disabled={loading} required onInput={(v) => (password = v)} />

          {#if mode === "register"}
            <TextField id="confirmPassword" label="Confirm password" type="password" value={confirmPassword} disabled={loading} required onInput={(v) => (confirmPassword = v)} />
            <PasswordRules rules={authConfig.passwordRequirements} />
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

          <button type="button" class="w-full text-xs text-fg-muted hover:text-fg-link mt-2 transition-colors" onclick={switchMode}>
            {mode === "login" ? "Need an account? Register" : "Already registered? Sign in"}
          </button>
        </form>
      {/if}
    {/if}
  </div>
</main>
