<script lang="ts">
    import { useAuth } from "../store/store";

    interface log {
        activity: string;
        description: string;
        userId: string;
        username: string;
        krname: string;
        time: number;
        loc: string;
    }

    type options = "folder"|"information"|"setting"|"account"|"log";

    let loading = new Promise(res => {});
    let checkingAdmin = new Promise(res => {});
    let logs:log[] = [];
    let select:number;
    let lastClicked:number = 0;
    export let currentPath:string[];
    export let selected:options;

    async function loadData() {
        const res = await fetch(`/server/getActivityLog`);
        const data = await res.json();

        logs = data.data;

        return;
    }

    function click(index:number) {
        if(select == index){
            if(Date.now() - lastClicked <= 1000){
                if(logs[index].loc == ""){
                    currentPath = [];
                }else{
                    currentPath = logs[index].loc.split("/");
                }
                selected = "folder"
                return;
            }
        }else{
            select = index;
        }

        lastClicked = Date.now();
    }

    function checkClick(e:MouseEvent) {
        const element = e.target as HTMLElement;
        if(!element.classList.contains("itemC")){
            select = -1;
        }
    }

    function formatTimestamp(timestamp: number): string {
        const date = new Date(timestamp);
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        const seconds = String(date.getSeconds()).padStart(2, '0');

        return `${year}-${month}-${day} ${hours}시 ${minutes}분 ${seconds}초`;
    }

    async function checkAdmin() {
        if($useAuth.token == "") return;

        const res = await fetch(`/server/checkAdmin?token=${$useAuth.token}`);
        const data = await res.json();

        if(data.isAdmin){
            return true;
        }else{
            return false;
        }
    }

    loading = loadData();
    checkingAdmin = checkAdmin();
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<main id="main" on:click={checkClick}>
    <main id="padding">
        <div class="title">📜 Activity Log</div>
        {#await loading}
            <div class="subtitle">Fetching logs...</div>
        {:then value}
            {#await checkingAdmin}
                <div class="subtitle">Checking ADMIN intent...</div>
            {:then value} 
                {#if value}
                    {#if logs.length == 0}
                        <div class="subtitle">There are no activity records yet. 😅</div>
                    {/if}
                    <div id="itemCon">
                        {#each logs as log, index}
                            <div class="item {select == index ? "selected" : ""} itemC" on:click={() => {click(index)}}>
                                <div class="left itemC">
                                    <div class="item_title itemC">[{log.activity}]</div>
                                    <div class="mid itemC">
                                        <div class="item_description itemC">{log.description}</div>
                                        <div class="item_time itemC">{formatTimestamp(log.time)}</div>
                                    </div>
                                </div>
                                <div class="right itemC">
                                    <div class="item_username itemC">{log.krname}</div>
                                </div>
                            </div>
                        {/each}
                    </div>
                {:else}
                    <div class="subtitle">This page is only for admin.</div>
                    <div class="subtitle">Please return.</div>
                {/if}
            {/await}
        {/await}
    </main>
</main>

<style lang="scss">
    #main{
        padding-left: 40px;
        padding-right: 40px;
        overflow-y: scroll;
        height: 100vh;
        user-select: none;
    }

    #main::-webkit-scrollbar {
        width: 5px;
        position: absolute;
        top: 0;
        right: 0;
    }

    #main::-webkit-scrollbar-track {
        background-color: #121212;
        border-radius: 5px;
    }

    #main::-webkit-scrollbar-thumb { 
        background-color: #565656;
        border-radius: 5px;
    }

    #main::-webkit-scrollbar-button {
        display: none;
    }

    #padding{
        padding-top: 40px;
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

    #itemCon{
        display: flex;
        flex-direction: column;
        margin-top: 30px;
        margin-bottom: 100px;
        gap: 5px;
    }

    .item{
        display: flex;
        flex-direction: row;
        align-items: center;
        border: 1px solid #121212;
        color: white;
        // gap: 20px;
        padding: 10px 20px 10px 20px;
        justify-content: space-between;
    }

    .item:hover {
        cursor: default;
        background-color: #383838;
    }

    .selected{
        border: 1px solid #878787;
    }

    .left{
        display: flex;
        align-items: center;
        gap: 20px;
        width: 92%;
    }

    .item_title{
        font-weight: bolder;
        font-size: larger;
    }

    .item_description{
        font-size: medium;
    }

    .item_time{
        font-size: small;
        color: lightgray;
    }

    .item_username{
        font-size: large;
        font-weight: bolder;
    }
</style>