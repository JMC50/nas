<script lang="ts">
    import { useAuth } from "../store/store";

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

    let loading = new Promise(res => {});
    let userSort:string;
    let users:user[];
    let selected:number = -1;
    let lastClicked:number;
    let sorted:user[];

    $: {
        sorted = sorted;
        users = users;
    }

    function click(index:number) {
        if(selected == index){
            if(Date.now() - lastClicked <= 1000){
                if(openUsers.includes(sorted[index].userId)){
                    opened_user = openUsers.indexOf(sorted[index].userId);
                    opened_user = opened_user;
                }else{
                    openUsers.push(sorted[index].userId);
                    openUsers = openUsers;
                    opened_user = openUsers.indexOf(sorted[index].userId);
                    opened_user = opened_user;
                    userList.push(sorted[index]);
                    userList = userList;
                }
                return;
            }
        }else{
            selected = index;
        }

        lastClicked = Date.now();
    }

    function checkClick(e:MouseEvent) {
        const element = e.target as HTMLElement;
        if(!element.classList.contains("userC")){
            selected = -1;
        }
    }

    async function sort() {
        selected = -1;
        const arr = [];
        for(let user of users){
            if(user.krname.includes(userSort)){
                arr.push(user);
            }
        }
        sorted = arr;
    }

    async function checkAdmin() {
        if($useAuth.token == "") return;

        const res = await fetch(`/server/checkAdmin?token=${$useAuth.token}`);
        const data = await res.json();

        if(data.isAdmin){
            const getAllUsersRES = await fetch(`/server/getAllUsers`);
            const data = await getAllUsersRES.json();
            users = data.users;
            sorted = users;

            return true;
        }else{
            return false;
        }
    }

    loading = checkAdmin();
</script>


<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<main id="main" on:click={checkClick}>
    <main id="padding">
        <div class="title">⚙️ Setting</div>
        {#await loading}
            <div class="subtitle">Checking ADMIN intent...</div>
        {:then value} 
            {#if !value}
                <div class="subtitle">This page is only for admin.</div>
                <div class="subtitle">Please return.</div>
            {:else}
                <div class="subtitle">✏️ User Management</div>
                <div id="searchCon">
                    <img src="/svg/search_black.svg" alt="">
                    <input type="search" placeholder="Search user by korean name" id="searchInput" bind:value={userSort} on:input={sort} autocomplete="off">
                </div>
                <div id="userListCon">
                    {#each sorted as user, index}
                        <div class="user userC {selected == index ? "selected" : ""}" on:click={() => {click(index)}}>
                            <div class="left userC">
                                <div class="user_name userC">{user.krname}</div>
                                <div class="user_name userC">{user.username}</div>
                            </div>
                            <div class="right userC">
                                <div class="user_name userC">{user.userId}</div>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
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

    #searchCon{
        background-color: white;
        flex-grow: 1;
        border-radius: 5px;
        display: flex;
        align-items: center;
        padding: 10px 15px;
        gap: 10px;
        margin-top: 20px;
    }

    #searchInput{
        background: none;
        font-size: large;
        border: 0;
        height: auto;
        width: 100%;
    }

    input:focus {outline:none;}

    #userListCon{
        display: flex;
        flex-direction: column;
        margin-top: 20px;
        gap: 10px;
    }

    .user{
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        padding: 12px 10px 12px 10px;
        border: 1px solid #121212;
    }

    .user:hover{
        cursor: default;
        background-color: #383838;
    }

    .selected{
        border: 1px solid #878787;
    }

    .left{
        display: flex;
        flex-direction: row;
        gap: 20px;
    }

    .user_name{
        color: white;
        font-size: large;
    }
</style>