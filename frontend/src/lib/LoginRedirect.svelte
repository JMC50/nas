<script lang="ts">
    import { onMount } from "svelte";
    import { useAuth } from "../store/store";

    interface userdata{
        userId: string;
        username: string;
        global_name: string;
    }

    let register:boolean = false;
    let saveData:userdata = {
        userId: '',
        username: '',
        global_name: ''
    }
    let koreanName:string = "";
    let access_token:string;

    onMount(async () => {
        access_token = location.href.split("access_token=")[1].split("&expires_in=")[0];

        const loginRES = await fetch(`/server/login?access_token=${access_token}`);
        const res = await loginRES.json();

        if(res.status == "new"){
            register = true;
            saveData.userId = res.userId;
            saveData.username = res.username;
            saveData.global_name = res.global_name;
        }else{
            useAuth.set({
                userId: res.userId,
                username: res.username,
                krname: res.krname,
                global_name: res.global_name,
                token: res.token
            })
            const baseUrl = `${window.location.protocol}//${window.location.host}/`;
            window.location.replace(baseUrl);
        }
    })

    async function goregister() {
        if(koreanName == ""){
            alert("한국 이름을 입력해주세요.");
            return;
        }

        const res = await fetch(`/server/register`, {
            method:"POST",
            headers:{
                'Content-Type':'application/json'
            },
            body: JSON.stringify({
                access_token,
                krname: koreanName
            })
        });

        const data = await res.json();

        if(data.status == "complete"){
            useAuth.set({
                userId: data.userId,
                username: data.username,
                global_name: data.global_name,
                token: data.token,
                krname: koreanName
            })
            const baseUrl = `${window.location.protocol}//${window.location.host}/`;
            window.location.replace(baseUrl);
        }else{
            alert("회원가입을 하는 도중 문제가 발생했습니다.")
            return;
        }
    }
</script>

<main>
    {#if !register}
        <div class="title">
            Redirecting... please wait...
        </div>
    {:else}
        <div id="registerCon">
            <div class="title">👋 You're new here!</div>
            <div class="subtitle">Before logging in, please let us know your Korean name</div>
            <div class="inputCon">
                <div class="inputTitle">Your discord id</div>
                <input type="text" value="{saveData.userId}" readonly class="noEdit">
            </div>
            <div class="inputCon">
                <div class="inputTitle">Your discord username</div>
                <input type="text" value="{saveData.username}" readonly class="noEdit">
            </div>
            <div class="inputCon">
                <div class="inputTitle">Your discord global name</div>
                <input type="text" value="{saveData.global_name}" readonly class="noEdit">
            </div>
            <div class="inputCon">
                <div class="inputTitle">Your Korean name</div>
                <input type="text" class="edit" bind:value={koreanName} placeholder="type plz">
            </div>
            <div id="sort">
                <button class="button" on:click={goregister}><span class="text">COMEPLETE</span></button>
            </div>
        </div>
    {/if}
</main>

<style lang="scss">
    $black: #000000;
    $white: #FFFFFF;
    $semibold: 500;
    $gray: rgba($black, 0.6);
    $dark-gray: rgba($black, 0.8);
    $light-gray: rgba($white, 0.8);

    main{
        display: flex;
        justify-content: center;
        align-items: center;
        width: 100vw;
        height: 100vh;
        overflow: none;
    }

    #registerCon{
        background-color: #383838;
        box-shadow: 4px 4px 4px #000000;
        padding: 20px;
        border-radius: 20px;
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
        margin-bottom: 20px;
    }

    .inputCon{
        display: flex;
        flex-direction: column;
        gap: 10px;
        margin-bottom: 10px;
    }

    .inputTitle{
        color: white;
        font-weight: bolder;
        font-size: large;
    }

    input{
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

    .edit{
        &:focus {
            color: $white;
        }
    }

    .noEdit:hover{
        cursor: default;
    }

    #sort{
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        margin-top: 20px;
    }

    .button {
        align-items: center;
        background-image: linear-gradient(144deg,#AF40FF, #5B42F3 50%,#00DDEB);
        border: 0;
        border-radius: 8px;
        box-shadow: rgba(151, 65, 252, 0.2) 0 15px 30px -5px;
        box-sizing: border-box;
        color: #FFFFFF;
        display: flex;
        font-size: 20px;
        justify-content: center;
        line-height: 1em;
        max-width: 100%;
        min-width: 140px;
        padding: 3px;
        text-decoration: none;
        user-select: none;
        -webkit-user-select: none;
        touch-action: manipulation;
        white-space: nowrap;
        cursor: pointer;
    }

    .button:active,
    .button:hover {
        outline: 0;
    }

    .button span {
        background-color: rgb(5, 6, 45);
        padding: 16px 24px;
        border-radius: 6px;
        width: 100%;
        height: 100%;
        transition: 300ms;
    }

    .button:hover span {
        background: none;
    }

    @media (min-width: 768px) {
        .button {
            font-size: 24px;
            min-width: 196px;
        }
    }
    
</style>