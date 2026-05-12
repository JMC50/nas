<script lang="ts">
    import { onMount } from "svelte";
    import { useAuth } from "../store/store";

    type intentList = "ADMIN"|"VIEW"|"OPEN"|"DOWNLOAD"|"UPLOAD"|"COPY"|"DELETE"|"RENAME";
    const intents:intentList[] = ["ADMIN", "VIEW", "OPEN", "DOWNLOAD", "UPLOAD", "COPY", "DELETE", "RENAME"];

    interface Iuser {
        userId: string;
        username: string;
        krname: string;
        global_name: string;
        intents: intentList[];
    }

    export let user:Iuser;

    let authorized:intentList[] = [];

    $:{
        authorized = authorized;
    }

    onMount(() => {
        refresh();
    })
    
    function refresh() {
        authorized = [];
        for(let intent of user.intents){
            authorized.push(intent);
        }
        authorized = authorized;
    }

    async function authorize(intent:intentList) {
        authorized.push(intent);
        authorized = authorized;

        const res = await fetch(`/server/authorize?userId=${user.userId}&intent=${intent}&token=${$useAuth.token}`);
        const data = await res.text();
        if(data != "complete"){
            alert("문제발생");
            refresh();
            return;
        }

        user.intents.push(intent);
    }

    async function unauthorize(intent:intentList) {
        const index2 = authorized.indexOf(intent);
        authorized.splice(index2, 1);
        authorized = authorized;

        const res = await fetch(`/server/unauthorize?userId=${user.userId}&intent=${intent}&token=${$useAuth.token}`);
        const data = await res.text();
        if(data != "complete"){
            alert("문제발생");
            refresh();
            return;
        }

        const index = user.intents.indexOf(intent);
        user.intents.splice(index, 1);
    }

</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_media_has_caption -->
<main id="container">
    <main id="padding">
        <div class="title">User Management : {user.krname}</div>
        <table>
            <tbody>
                <tr>
                    <td>ID</td>
                    <td>:</td>
                    <td>{user.userId}</td>
                </tr>
                <tr>
                    <td>USERNAME</td>
                    <td>:</td>
                    <td>{user.username}</td>
                </tr>
                <tr>
                    <td>GLOBAL_NAME</td>
                    <td>:</td>
                    <td>{user.global_name}</td>
                </tr>
                <tr>
                    <td>KOREAN_NAME</td>
                    <td>:</td>
                    <td>{user.krname}</td>
                </tr>
            </tbody>
        </table>
        <div class="title">❗ INTENTS SETTING ❗</div>
        <div id="intentsCon">
            {#each intents as intent}
                <div class={authorized.includes(intent) ? "authorized" : "unauthorized"} on:click={() => {authorized.includes(intent) ? unauthorize(intent) : authorize(intent)}}>{intent}</div>
            {/each}
        </div>
    </main>
</main>

<style lang="scss">
    #container{
        width: 100%;
        height: 95.9vh;
        display: flex;
        overflow-y: hidden;
        user-select: none;
    }

    #padding{
        padding-left: 30px;
        padding-top: 30px;
        padding-right: 30px;
    }

    .title{
        color: white;
        font-weight: bolder;
        font-size: x-large;
    }

    table{
        margin-top: 20px;
        margin-bottom: 20px;
    }

    td{
        color: white;
        font-size: large;
        padding-left: 5px;
        padding-right: 5px;
    }

    #intentsCon{
        margin-top: 20px;
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
        transition: all 0.1s ease;
    }

    .authorized:hover{
        cursor: pointer;
        // border: 2px solid rgb(255, 127, 127);
        // background-color: rgba(255, 110, 110, 0.5);
        border: 2px solid rgb(99, 99, 99);
        background-color: rgba(100, 100, 100, 0.5);
    }

    .unauthorized{
        color: white;
        font-weight: bolder;
        padding: 5px 10px 5px 10px;
        border-radius: 10px;
        border: 2px solid rgb(255, 127, 127);
        background-color: rgba(255, 110, 110, 0.5);
        transition: all 0.1s ease;
    }

    .unauthorized:hover{
        cursor: pointer;
        // border: 2px solid rgb(127, 255, 127);
        // background-color: rgba(110, 255, 110, 0.5);
        border: 2px solid rgb(99, 99, 99);
        background-color: rgba(100, 100, 100, 0.5);
    }
</style>