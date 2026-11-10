<script lang="ts">
  import Logo from "$lib/components/Logo.svelte";
  import { onMount } from "svelte";
  import LogIn from "lucide-svelte/icons/log-in";
  import UserPlus from "lucide-svelte/icons/user-plus";
  import TextField from "$lib/components/Auth/TextField.svelte";
  import PasswordRules from "$lib/components/Auth/PasswordRules.svelte";
  import SignInOptions from "$lib/components/Auth/SignInOptions.svelte";
  import { signInLocal, signUpLocal } from "$lib/auth-actions";

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
    discordEnabled: boolean;
    discordLoginUrl: string;
    googleEnabled: boolean;
    googleLoginUrl: string;
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
</script>

<main class="min-h-full flex flex-col items-center justify-center px-gutter py-10">
  <div class="w-full max-w-[400px]">
    <div class="flex items-center gap-2 mb-5">
      <Logo size={26} />
      <span class="text-lg font-bold text-fg-primary">NAS</span>
    </div>

    <div class="sg-panel p-5 sm:p-6">
      <h1 class="text-lg font-semibold text-fg-primary">
        {mode === "login" ? "Sign in" : "Create account"}
      </h1>
      <p class="text-sm text-fg-secondary mt-1 mb-5">
        {mode === "login" ? "Use your ID and password to continue." : "Pick an ID and password to register."}
      </p>

      {#if !authConfig}
        {#if error}
          <div class="p-3 rounded-control border border-fg-danger/45 text-fg-danger text-sm" role="alert">{error}</div>
        {:else}
          <div class="text-sm text-fg-muted" role="status">Loading configuration…</div>
        {/if}
      {:else if !authConfig.oauthEnabled && !authConfig.localAuthEnabled}
        <div class="p-3 rounded-control border border-fg-warning/45 text-fg-warning text-sm" role="alert">
          No authentication methods enabled.
        </div>
      {:else}
        {#if authConfig.oauthEnabled}
          <SignInOptions
            discordEnabled={authConfig.discordEnabled}
            discordUrl={authConfig.discordLoginUrl}
            googleEnabled={authConfig.googleEnabled}
            googleUrl={authConfig.googleLoginUrl}
            localEnabled={false}
            onLocal={() => {}}
          />
          {#if authConfig.localAuthEnabled}
            <div class="flex items-center gap-3 my-5 text-xs text-fg-muted">
              <div class="flex-1 h-px bg-border-default"></div>
              <span>or</span>
              <div class="flex-1 h-px bg-border-default"></div>
            </div>
          {/if}
        {/if}

        {#if authConfig.localAuthEnabled}
          {#if error}
            <div class="mb-4 p-3 rounded-control border border-fg-danger/45 text-fg-danger text-sm" role="alert">{error}</div>
          {/if}

          <form class="flex flex-col gap-3" onsubmit={submit}>
            <TextField id="userId" label="User ID" value={userId} placeholder="your-id" disabled={loading} required autocomplete="username" invalid={!!error} onInput={(v) => (userId = v)} />

            {#if mode === "register"}
              <TextField id="username" label="Username" value={username} placeholder="Display name" disabled={loading} required autocomplete="off" onInput={(v) => (username = v)} />
              <TextField id="krname-reg" label="Korean name (optional)" value={koreanName} placeholder="홍길동" disabled={loading} autocomplete="name" onInput={(v) => (koreanName = v)} />
            {/if}

            <TextField id="password" label="Password" type="password" value={password} disabled={loading} required autocomplete={mode === "login" ? "current-password" : "new-password"} invalid={!!error} onInput={(v) => (password = v)} />

            {#if mode === "register"}
              <TextField id="confirmPassword" label="Confirm password" type="password" value={confirmPassword} disabled={loading} required autocomplete="new-password" onInput={(v) => (confirmPassword = v)} />
              <PasswordRules rules={authConfig.passwordRequirements} />
            {/if}

            <button type="submit" class="mt-1 sg-btn sg-btn-primary sg-btn-block" disabled={loading}>
              {#if mode === "login"}
                <LogIn size="16" />
                {loading ? "Signing in…" : "Sign in"}
              {:else}
                <UserPlus size="16" />
                {loading ? "Creating…" : "Create account"}
              {/if}
            </button>

            <button type="button" class="self-center sg-btn sg-btn-ghost" onclick={switchMode}>
              {mode === "login" ? "Need an account? Register" : "Already registered? Sign in"}
            </button>
          </form>
        {/if}
      {/if}
    </div>
  </div>
</main>
