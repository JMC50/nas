<script lang="ts">
    import { onMount } from "svelte";
    import { useAuth } from "../store/store";

    interface AuthConfig {
        authType: 'oauth' | 'local' | 'both';
        localAuthEnabled: boolean;
        oauthEnabled: boolean;
        passwordRequirements: {
            minLength: number;
            requireUppercase: boolean;
            requireLowercase: boolean;
            requireNumber: boolean;
            requireSpecial: boolean;
        };
    }

    let authConfig: AuthConfig | null = null;
    let mode: 'login' | 'register' = 'login';
    let loading = false;
    let error = "";

    // Form fields
    let userId = "";
    let username = "";
    let password = "";
    let confirmPassword = "";
    let koreanName = "";

    onMount(async () => {
        // Get auth configuration
        const res = await fetch('/server/auth/config');
        authConfig = await res.json();
    });

    function validatePassword(pwd: string): string | null {
        if (!authConfig) return null;
        
        const reqs = authConfig.passwordRequirements;
        
        if (pwd.length < reqs.minLength) {
            return `Password must be at least ${reqs.minLength} characters`;
        }
        if (reqs.requireUppercase && !/[A-Z]/.test(pwd)) {
            return "Password must contain at least one uppercase letter";
        }
        if (reqs.requireLowercase && !/[a-z]/.test(pwd)) {
            return "Password must contain at least one lowercase letter";
        }
        if (reqs.requireNumber && !/[0-9]/.test(pwd)) {
            return "Password must contain at least one number";
        }
        if (reqs.requireSpecial && !/[!@#$%^&*(),.?":{}|<>]/.test(pwd)) {
            return "Password must contain at least one special character";
        }
        
        return null;
    }

    async function handleLogin() {
        error = "";
        loading = true;

        try {
            const res = await fetch('/server/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    userId,
                    password
                })
            });

            const data = await res.json();

            if (data.success) {
                useAuth.set({
                    userId: data.user.userId,
                    username: data.user.username,
                    krname: data.user.krname || '',
                    global_name: data.user.global_name,
                    token: data.token
                });
                
                const baseUrl = `${window.location.protocol}//${window.location.host}/`;
                window.location.replace(baseUrl);
            } else {
                error = data.message || "Login failed";
            }
        } catch (err) {
            error = "Network error. Please try again.";
        } finally {
            loading = false;
        }
    }

    async function handleRegister() {
        error = "";
        
        // Validate form
        if (!userId || !username || !password) {
            error = "Please fill in all required fields";
            return;
        }

        if (password !== confirmPassword) {
            error = "Passwords do not match";
            return;
        }

        const passwordError = validatePassword(password);
        if (passwordError) {
            error = passwordError;
            return;
        }

        loading = true;

        try {
            const res = await fetch('/server/auth/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    userId,
                    username,
                    password,
                    krname: koreanName
                })
            });

            const data = await res.json();

            if (data.success) {
                useAuth.set({
                    userId: data.user.userId,
                    username: data.user.username,
                    krname: koreanName || '',
                    global_name: data.user.global_name,
                    token: data.token
                });
                
                const baseUrl = `${window.location.protocol}//${window.location.host}/`;
                window.location.replace(baseUrl);
            } else {
                error = data.message || "Registration failed";
            }
        } catch (err) {
            error = "Network error. Please try again.";
        } finally {
            loading = false;
        }
    }

    function switchMode() {
        mode = mode === 'login' ? 'register' : 'login';
        error = "";
        // Clear form
        userId = "";
        username = "";
        password = "";
        confirmPassword = "";
        koreanName = "";
    }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<main>
    <div class="title">🔐 {mode === 'login' ? 'Login' : 'Register'}</div>
    
    {#if authConfig}
        {#if authConfig.localAuthEnabled}
            {#if error}
                <div class="error">{error}</div>
                <div class="margin"></div>
            {/if}

            <form on:submit|preventDefault={mode === 'login' ? handleLogin : handleRegister}>
                <div class="inputCon">
                    <div class="inputTitle">User ID</div>
                    <input 
                        type="text" 
                        class="edit" 
                        bind:value={userId} 
                        placeholder="Enter your user ID"
                        disabled={loading}
                        required
                    />
                </div>

                {#if mode === 'register'}
                    <div class="inputCon">
                        <div class="inputTitle">Username</div>
                        <input 
                            type="text" 
                            class="edit" 
                            bind:value={username} 
                            placeholder="Enter your username"
                            disabled={loading}
                            required
                        />
                    </div>

                    <div class="inputCon">
                        <div class="inputTitle">Korean Name (Optional)</div>
                        <input 
                            type="text" 
                            class="edit" 
                            bind:value={koreanName} 
                            placeholder="Enter your Korean name"
                            disabled={loading}
                        />
                    </div>
                {/if}

                <div class="inputCon">
                    <div class="inputTitle">Password</div>
                    <input 
                        type="password" 
                        class="edit" 
                        bind:value={password} 
                        placeholder="Enter your password"
                        disabled={loading}
                        required
                    />
                </div>

                {#if mode === 'register'}
                    <div class="inputCon">
                        <div class="inputTitle">Confirm Password</div>
                        <input 
                            type="password" 
                            class="edit" 
                            bind:value={confirmPassword} 
                            placeholder="Confirm your password"
                            disabled={loading}
                            required
                        />
                    </div>

                    {#if authConfig.passwordRequirements}
                        <div class="passwordReqs">
                            <div class="subtitle">Password Requirements:</div>
                            <ul>
                                <li>At least {authConfig.passwordRequirements.minLength} characters</li>
                                {#if authConfig.passwordRequirements.requireUppercase}
                                    <li>At least one uppercase letter</li>
                                {/if}
                                {#if authConfig.passwordRequirements.requireLowercase}
                                    <li>At least one lowercase letter</li>
                                {/if}
                                {#if authConfig.passwordRequirements.requireNumber}
                                    <li>At least one number</li>
                                {/if}
                                {#if authConfig.passwordRequirements.requireSpecial}
                                    <li>At least one special character</li>
                                {/if}
                            </ul>
                        </div>
                        <div class="margin"></div>
                    {/if}
                {/if}

                <div class="button" on:click={mode === 'login' ? handleLogin : handleRegister}>
                    {#if loading}
                        PROCESSING...
                    {:else}
                        {mode === 'login' ? 'LOGIN' : 'REGISTER'}
                    {/if}
                </div>
                
                <div class="margin"></div>
                
                <div class="switchMode" on:click={switchMode}>
                    {mode === 'login' ? "Don't have an account? Register here" : "Already have an account? Login here"}
                </div>
            </form>
        {:else}
            <div class="subtitle">Local authentication is disabled. Please use OAuth.</div>
        {/if}
    {:else}
        <div class="subtitle">Loading configuration...</div>
    {/if}
</main>

<style lang="scss">
    $black: #000000;
    $white: #FFFFFF;
    $semibold: 500;
    $gray: rgba($black, 0.6);
    $dark-gray: rgba($black, 0.8);
    $light-gray: rgba($white, 0.8);

    main {
        padding-top: 40px;
        padding-left: 40px;
        padding-right: 40px;
    }

    .title {
        font-size: xx-large;
        color: white;
        font-weight: bolder;
        margin-bottom: 20px;
    }

    .subtitle {
        font-size: x-large;
        color: white;
        font-weight: bolder;
    }

    .margin {
        margin-top: 20px;
    }

    .inputCon {
        display: flex;
        flex-direction: column;
        gap: 10px;
        margin-bottom: 15px;
        width: 100%;
        max-width: 400px;
    }

    .inputTitle {
        color: white;
        font-weight: bolder;
        font-size: large;
    }

    input {
        outline: none;
        display: block;
        background: rgba($black, 0.1);
        width: 100%;
        border: 0;
        border-radius: 10px;
        box-sizing: border-box;
        padding: 12px 20px;
        color: $light-gray;
        font-family: inherit;
        font-size: inherit;
        font-weight: $semibold;
        line-height: inherit;
        transition: 0.3s ease;
    }

    .edit {
        &:focus {
            color: $white;
        }
    }

    .passwordReqs {
        background: rgba($black, 0.3);
        border-radius: 10px;
        padding: 15px;
        margin-bottom: 15px;
        max-width: 400px;

        .subtitle {
            font-size: medium;
            margin-bottom: 10px;
        }

        ul {
            margin: 0;
            padding-left: 20px;
            
            li {
                color: $light-gray;
                font-size: small;
                margin: 5px 0;
            }
        }
    }

    .error {
        background: rgba(255, 0, 0, 0.3);
        border: 1px solid rgba(255, 0, 0, 0.5);
        color: #ff6b6b;
        padding: 15px;
        border-radius: 10px;
        text-align: center;
        max-width: 400px;
        font-weight: bolder;
    }

    .button {
        display: inline-block;
        width: fit-content;
        background-color: #383838;
        box-shadow: 4px 4px 4px #000000;
        color: white;
        padding: 10px 20px 10px 20px;
        border-radius: 5px;
        font-weight: bolder;
        font-size: large;
        user-select: none;
        cursor: pointer;
    }

    .button:hover {
        background-color: rgb(92, 92, 92);
    }

    .switchMode {
        color: #00DDEB;
        text-decoration: underline;
        cursor: pointer;
        text-align: left;
        font-size: medium;
        font-weight: bolder;

        &:hover {
            color: #AF40FF;
        }
    }
</style>