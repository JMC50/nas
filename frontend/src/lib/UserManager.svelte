<script lang="ts">
    import { onDestroy, onMount } from "svelte";
    import FileViewer from "./FileViewer.svelte";
    import AccountViewer from "./AccountViewer.svelte";

    type intentList = "ADMIN"|"VIEW"|"OPEN"|"DOWNLOAD"|"UPLOAD"|"COPY"|"DELETE"|"RENAME";

    interface user {
        userId: string;
        username: string;
        krname: string;
        global_name: string;
        intents: intentList[];
    }

    export let openUsers:string[];
    export let opened_user:number;
    export let userList:user[];

    let openedUser:user;
    let fileListCon:HTMLDivElement;

    $:{
        openUsers = openUsers;
        opened_user = opened_user;
        openedUser = userList[opened_user];
    }

    onMount(() => {
        function handleWheel(event: WheelEvent) {
            event.preventDefault();
            fileListCon.scrollLeft += event.deltaY;
            fileListCon.scrollLeft += event.deltaX;
        }

        fileListCon.addEventListener("wheel", handleWheel, { passive: false });

        onDestroy(() => {
            fileListCon.removeEventListener("wheel", handleWheel);
        });
    })

    function scrollToFile() {
        if(!fileListCon) return;
        setTimeout(() => {
            const fileElements = fileListCon.querySelectorAll(".file");
            const targetElement = fileElements[opened_user];
    
            if(targetElement){
                targetElement.scrollIntoView({ behavior: "smooth", block: "nearest", inline: "center" });
            }
        }, 10);
    }

    $: opened_user, scrollToFile();

    function changeOpen(index:number, e:MouseEvent) {
        const target = e.target as HTMLDivElement;
        if(target.classList.contains("close")) return;
        opened_user = index
    }

    function closeFile(index:number) {
        if(opened_user >= index && userList.length != 1){
            opened_user = opened_user - 1;
        }
        if(opened_user == -1){
            opened_user = 0;
        }
        userList.splice(index, 1);
        openUsers.splice(index, 1);
        userList = userList;
        openUsers = openUsers;
    }

</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<main>
    <div id="fileListCon" bind:this={fileListCon}>
        {#each userList as user, index}
            <div class="file {opened_user == index ? "opened" : ""}" on:click={(e) => {changeOpen(index, e)}}>
                <img src="/svg/account.svg" alt="" class="icon">
                <div class="fileName">{user.krname}</div>
                <div class="close" on:click={() => {closeFile(index)}}>X</div>
            </div>
        {/each}
    </div>
    <div id="viewerCon">
        {#each userList as user, index (user.krname)}
            <div class={opened_user == index ? "show" : "hide"}>
                <AccountViewer bind:user={user} />
            </div>
        {/each}
    </div>
</main>

<style lang="scss">
    main{
        display: flex;
        flex-direction: column;
        width: 47vw;
    }

    #fileListCon {
        width: 100%;
        height: 4vh;
        // max-width: 40vw;
        display: flex;
        flex-direction: row;
        overflow-x: auto;
        user-select: none;
        white-space: nowrap;
        border-bottom: 1px solid #878787;
    }

    #fileListCon::-webkit-scrollbar {
        height: 1px;
        position: absolute;
        top: 0;
        right: 0;
    }

    #fileListCon::-webkit-scrollbar-track {
        background-color: #878787;
        border-radius: 5px;
    }

    #fileListCon::-webkit-scrollbar-thumb { 
        background-color: #565656;
        border-radius: 5px;
    }

    #fileListCon::-webkit-scrollbar-button {
        display: none;
    }

    .file{
        display: flex;
        flex-direction: row;
        background-color: #0f0f0f;
        align-items: center;
        color: #878787;
        padding: 10px;
        gap: 10px;
        border-top: 3px solid #121212;
        border-right: 1px solid #878787;
        border-bottom: 1px solid #878787;
    }

    .file.opened{
        color: white;
        background-color: #121212;
        border-top: 3px solid #2f5fff;
        border-bottom: 1px solid #121212;
    }

    .file:hover{
        cursor: pointer;
    }

    .close{
        padding: 3px 5px 3px 5px;
        border-radius: 5px;
    }

    .close:hover{
        background-color: #565656;
    }

    .icon{
        height: 20px;
    }

    .show{
        display: block;
    }

    .hide{
        display: none;
    }
</style>