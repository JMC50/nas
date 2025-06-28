<script lang="ts">
    import { onMount } from "svelte";
    import Explorer from "./lib/Explorer.svelte";
    import FileManager from "./lib/FileManager.svelte";
    import SideMenu from "./lib/SideMenu.svelte";
    import SystemInfo from "./lib/SystemInfo.svelte";
    import ExplorerMobile from "./lib/Explorer_mobile.svelte";
    import BottomMenu from "./lib/BottomMenu.svelte";
    import SystemInfoMobile from "./lib/SystemInfo_mobile.svelte";
    import FileViewerMobile from "./lib/FileViewer_mobile.svelte";
    import Account from "./lib/Account.svelte";
    import LoginRedirect from "./lib/LoginRedirect.svelte";
    import ActivityLog from "./lib/ActivityLog.svelte";
    import Setting from "./lib/Setting.svelte";
    import UserManager from "./lib/UserManager.svelte";
    import LoginRedirectKakao from "./lib/LoginRedirectKakao.svelte";
    type options = "folder"|"information"|"account"|"setting"|"log";

    interface file{
        name: string;
        loc: string;
        extensions: string;
        modified: boolean;
    }

    type intentList = "ADMIN"|"VIEW"|"OPEN"|"DOWNLOAD"|"UPLOAD"|"COPY"|"DELETE"|"RENAME";

    interface user {
        userId: string;
        username: string;
        krname: string;
        global_name: string;
        intents: intentList[];
    }

    let selected:options = "folder";
    let currentPath:string[] = [];
    
    let openFiles:string[] = [];
    let opened_file:number;
    let fileList:file[] = [];

    let openUsers:string[] = [];
    let opened_user:number;
    let userList:user[] = [];

    let screenType:"mobile"|"pc" = "mobile";
    let screenWidth = 0;
    let screenHeight = 0;

    $:{
        screenWidth = screenWidth;
        screenHeight = screenHeight;
    }

    onMount(() => {
        screenWidth = window.screen.width;
        screenHeight = window.screen.height;
        if(screenWidth < 600 && screenHeight > screenWidth){
            screenType = "mobile";
        }else if(screenWidth > 1000 && screenHeight < screenWidth){
            screenType = "pc";
        }

        window.addEventListener("resize", e => {
            screenWidth = window.screen.width;
            screenHeight = window.screen.height;
            if(screenWidth < 600 && screenHeight > screenWidth){
                screenType = "mobile";
            }else if(screenWidth > 1000 && screenHeight < screenWidth){
                screenType = "pc";
            }
        })
    })
</script>

<main>
    {#if location.pathname == "/FileViewer"}
        <FileViewerMobile />
    {:else if location.pathname == "/login"}
        <LoginRedirect />
    {:else if location.pathname == "/kakaoLogin"}
        <LoginRedirectKakao />
    {:else}
        {#if screenType == "pc"}
            <div id="sideMenu">
                <SideMenu bind:selected/>
            </div>
            <div id="main_pc">
                {#if selected == "folder"}
                    <Explorer bind:openFiles bind:opened_file bind:fileList bind:currentPath/>
                {:else if selected == "information"}
                    <SystemInfo />
                {:else if selected == "account"}
                    <Account />
                {:else if selected == "log"}
                    <ActivityLog bind:currentPath bind:selected/>
                {:else if selected == "setting"}
                    <Setting bind:userList bind:opened_user bind:openUsers/>
                {/if}
            </div>
            <div id="line"></div>
            <div id="fileManager">
                {#if selected == "setting"}
                    <UserManager bind:openUsers bind:opened_user bind:userList/>
                {:else}
                    <FileManager bind:openFiles bind:opened_file bind:fileList/>
                {/if}
            </div>
        {:else}
            <div id="bottomMenu">
                <BottomMenu bind:selected />
            </div>
            <div id="main_mobile">
                {#if selected == "folder"}
                    <ExplorerMobile/>
                {:else if selected == "information"}
                    <SystemInfoMobile />
                {:else if selected == "account"}
                    <Account />
                {:else if selected == "log"}
                    <ActivityLog bind:currentPath bind:selected/>
                {:else if selected == "setting"}
                    <Setting bind:userList bind:opened_user bind:openUsers/>
                {/if}
            </div>
        {/if}
    {/if}
</main>

<style lang="scss">
    main{
        display: flex;
        flex-direction: row;
        justify-content: left;
        height: 100vh;
    }

    #sideMenu{
        flex-shrink: 0;
    }

    #bottomMenu{
        position: fixed;
        left: 0;
        bottom: 0;
    }

    #main_pc{
        max-width: 50vw;
        flex-grow: 1;
        // overflow-y: auto;
        max-height: 100vh;
    }

    #main_mobile{
        flex-grow: 1;
        overflow-y: auto;
        max-height: 100vh;
    }

    #fileManager{
        border-left: 1px solid #878787;
    }
</style>