<script lang="ts">
    import { Logout, useAuth } from "$lib/store/store";
    import { onMount } from "svelte";

    type intentList = "ADMIN"|"VIEW"|"OPEN"|"DOWNLOAD"|"UPLOAD"|"COPY"|"DELETE"|"RENAME";

    interface AuthConfig {
        authType: 'oauth' | 'local' | 'both';
        localAuthEnabled: boolean;
        oauthEnabled: boolean;
    }

    const loginURL = process.env.LOGIN_URL as string;
    const GOOGLE_CLIENT_ID = process.env.GOOGLE_CLIENT_ID as string;
    const GOOGLE_REDIRECT_URI = process.env.GOOGLE_REDIRECT_URI as string;

    let loading = new Promise(res => {});
    let intents:intentList[];
    let unauthorized:intentList[] = ["ADMIN", "COPY", "DELETE", "DOWNLOAD", "OPEN", "RENAME", "UPLOAD", "VIEW"];
    let authConfig: AuthConfig | null = null;

    onMount(async () => {
        // Get auth configuration
        const res = await fetch('/server/auth/config');
        authConfig = await res.json();
    });

    async function login() {
        location.href = loginURL;
    }

    async function loginWGoogle() {
        const scope = encodeURIComponent("openid email profile");
        const redirect = encodeURIComponent(GOOGLE_REDIRECT_URI);
        location.href = `https://accounts.google.com/o/oauth2/v2/auth?response_type=code&client_id=${GOOGLE_CLIENT_ID}&redirect_uri=${redirect}&scope=${scope}`;
    }

    async function loginLocal() {
        location.href = "/localLogin";
    }

    async function logout() {
        Logout();
        const baseUrl = `${window.location.protocol}//${window.location.host}/`;
        window.location.replace(baseUrl);
    }

    async function getAdmin() {
        const pwd = prompt("password?");

        if(pwd){
            const res = await fetch(`/server/requestAdminIntent?token=${$useAuth.token}`, {
                method:"POST",
                headers:{
                    'Content-Type':'application/json'
                },
                body: JSON.stringify({
                    pwd
                })
            });

            const data = await res.text();

            if(data == "complete"){
                alert("관리자 권한이 지급되었습니다.");
                loading = getintents();
            }else{
                alert("잘못된 비밀번호입니다.")
            }
        }
    }

    async function getintents() {
        if($useAuth.userId == "") return;

        const res = await fetch(`/server/getIntents?userId=${$useAuth.userId}`);
        const data = await res.json();
        intents = data.intents;

        for(let intent of intents){
            const index = unauthorized.indexOf(intent);
            unauthorized.splice(index, 1);
        }
        return;
    }

    loading = getintents();
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<main>
    <div class="title">✏️ Account Manager</div>
    {#if $useAuth.userId == ""}
        <div class="subtitle">Please login to use all functions</div>
        <div class="margin"></div>
        {#if authConfig}
            {#if authConfig.oauthEnabled}
                <div class="button" on:click={login}>LOGIN WITH DISCORD</div>
                <div class="margin"></div>
                <div class="button" on:click={loginWGoogle}>LOGIN WITH GOOGLE</div>
                <div class="margin"></div>
            {/if}
            {#if authConfig.localAuthEnabled}
                <div class="button" on:click={loginLocal}>LOGIN WITH ID/PASSWORD</div>
                <div class="margin"></div>
            {/if}
        {:else}
            <div class="subtitle">Loading authentication options...</div>
        {/if}
    {:else}
        <div class="subtitle">You're logged in !</div>
        <div class="subtitle">Hello {$useAuth.global_name} 👋</div>
        <div class="margin"></div>

        {#await loading}
            <div class="subtitle">Checking permissions...</div>
        {:then value} 
            <div class="subtitle">Authorized intents</div>
            <div class="margin"></div>
            <div id="intentsCon">
                {#each intents as intent}
                    <div class="authorized">{intent}</div>
                {/each}
            </div>
            <div class="margin"></div>
            <div class="subtitle">Unauthorized intents</div>
            <div class="margin"></div>
            <div id="intentsCon">
                {#each unauthorized as intent}
                    <div class="unauthorized">{intent}</div>
                {/each}
            </div>
            <div class="margin"></div>

            {#if !intents.includes("ADMIN")}
                <div class="button" on:click={getAdmin}>REQUEST ADMIN INTENTS</div>
                <div class="margin"></div>
            {/if}
        {/await}


        <div class="button" on:click={logout}>LOGOUT</div>
    {/if}
</main>

<style lang="scss">
    main{
        padding-top: 40px;
        padding-left: 40px;
        padding-right: 40px;
    }

    .title{
        font-size: xx-large;
        color: white;
        font-weight: bolder;
        margin-bottom: 20px;
    }

    .subtitle{
        font-size: x-large;
        color: white;
        font-weight: bolder;
    }

    .margin{
        margin-top: 20px;
    }

    #intentsCon{
        display: flex;
        flex-direction: row;
        gap: 10px;
    }

    .authorized{
        color: white;
        font-weight: bolder;
        padding: 5px 10px 5px 10px;
        border-radius: 10px;
        border: 2px solid rgb(127, 255, 127);
        background-color: rgba(110, 255, 110, 0.5);
    }

    .unauthorized{
        color: white;
        font-weight: bolder;
        padding: 5px 10px 5px 10px;
        border-radius: 10px;
        border: 2px solid rgb(255, 127, 127);
        background-color: rgba(255, 110, 110, 0.5);
    }

    .button{
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
    }

    .button:hover{
        cursor: pointer;
        background-color: rgb(92, 92, 92);
    }
</style>