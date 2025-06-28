<script lang="ts">
    type options = "folder"|"information"|"account"|"setting"|"log";
    export let selected:options;
    
    let hovering:options|null = null;

    function hover(target:options) {
        hovering = target;
    }

    function hoverout() {
        hovering = null;
    }

    function select(target:options) {
        selected = target;
    }

    let white = "#FFFFFF";
    let gray = "#878787";
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<main>
    <div id="innerCon">
        <div class="button {selected == "folder" ? "selected" : ""}" on:click={() => {select("folder")}} on:mouseenter={() => hover("folder")} on:mouseleave={hoverout}> <!-- folder -->
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M3 8.2C3 7.07989 3 6.51984 3.21799 6.09202C3.40973 5.71569 3.71569 5.40973 4.09202 5.21799C4.51984 5 5.0799 5 6.2 5H9.67452C10.1637 5 10.4083 5 10.6385 5.05526C10.8425 5.10425 11.0376 5.18506 11.2166 5.29472C11.4184 5.4184 11.5914 5.59135 11.9373 5.93726L12.0627 6.06274C12.4086 6.40865 12.5816 6.5816 12.7834 6.70528C12.9624 6.81494 13.1575 6.89575 13.3615 6.94474C13.5917 7 13.8363 7 14.3255 7H17.8C18.9201 7 19.4802 7 19.908 7.21799C20.2843 7.40973 20.5903 7.71569 20.782 8.09202C21 8.51984 21 9.0799 21 10.2V15.8C21 16.9201 21 17.4802 20.782 17.908C20.5903 18.2843 20.2843 18.5903 19.908 18.782C19.4802 19 18.9201 19 17.8 19H6.2C5.07989 19 4.51984 19 4.09202 18.782C3.71569 18.5903 3.40973 18.2843 3.21799 17.908C3 17.4802 3 16.9201 3 15.8V8.2Z" stroke={hovering == "folder" || selected == "folder" ? white : gray} stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
        </div>

        <div class="button {selected == "information" ? "selected" : ""}" on:click={() => {select("information")}} on:mouseenter={() => hover("information")} on:mouseleave={hoverout}> <!-- systemInfo -->
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <g clip-path="url(#clip0_429_11160)">
                <circle cx="12" cy="11.9999" r="9" stroke={hovering == "information" || selected == "information" ? white : gray} stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
                <rect x="12" y="8" width="0.01" height="0.01" stroke={hovering == "information" || selected == "information" ? white : gray} stroke-width="3.75" stroke-linejoin="round"/>
                <path d="M12 12V16" stroke={hovering == "information" || selected == "information" ? white : gray} stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
                </g>
                <defs>
                <clipPath id="clip0_429_11160">
                <rect width="24" height="24" fill="white"/>
                </clipPath>
                </defs>
            </svg>
        </div>

        <div class="button {selected == "account" ? "selected" : ""}" on:click={() => {select("account")}} on:mouseenter={() => hover("account")} on:mouseleave={hoverout}> <!-- account -->
            <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
                <path d="M0 0h48v48H0z" fill="none"/>
                <g id="Shopicon">
                    <path d="M31.278,25.525C34.144,23.332,36,19.887,36,16c0-6.627-5.373-12-12-12c-6.627,0-12,5.373-12,12
                        c0,3.887,1.856,7.332,4.722,9.525C9.84,28.531,5,35.665,5,44h38C43,35.665,38.16,28.531,31.278,25.525z M16,16c0-4.411,3.589-8,8-8
                        s8,3.589,8,8c0,4.411-3.589,8-8,8S16,20.411,16,16z M24,28c6.977,0,12.856,5.107,14.525,12H9.475C11.144,33.107,17.023,28,24,28z"
                    fill={hovering == "account" || selected == "account" ? white : gray}/>
                </g>
            </svg>
        </div>

        <div class="button {selected == "log" ? "selected" : ""}" on:click={() => {select("log")}} on:mouseenter={() => hover("log")} on:mouseleave={hoverout}> <!-- log -->
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <g id="SVGRepo_bgCarrier" stroke-width="0"></g>
                <g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g>
                <g id="SVGRepo_iconCarrier"> 
                    <path d="M12 8V12L14.5 14.5" stroke={hovering == "log" || selected == "log" ? white : gray} stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
                    <path d="M5.60423 5.60423L5.0739 5.0739V5.0739L5.60423 5.60423ZM4.33785 6.87061L3.58786 6.87438C3.58992 7.28564 3.92281 7.61853 4.33408 7.6206L4.33785 6.87061ZM6.87963 7.63339C7.29384 7.63547 7.63131 7.30138 7.63339 6.88717C7.63547 6.47296 7.30138 6.13549 6.88717 6.13341L6.87963 7.63339ZM5.07505 4.32129C5.07296 3.90708 4.7355 3.57298 4.32129 3.57506C3.90708 3.57715 3.57298 3.91462 3.57507 4.32882L5.07505 4.32129ZM3.75 12C3.75 11.5858 3.41421 11.25 3 11.25C2.58579 11.25 2.25 11.5858 2.25 12H3.75ZM16.8755 20.4452C17.2341 20.2378 17.3566 19.779 17.1492 19.4204C16.9418 19.0619 16.483 18.9393 16.1245 19.1468L16.8755 20.4452ZM19.1468 16.1245C18.9393 16.483 19.0619 16.9418 19.4204 17.1492C19.779 17.3566 20.2378 17.2341 20.4452 16.8755L19.1468 16.1245ZM5.14033 5.07126C4.84598 5.36269 4.84361 5.83756 5.13505 6.13191C5.42648 6.42626 5.90134 6.42862 6.19569 6.13719L5.14033 5.07126ZM18.8623 5.13786C15.0421 1.31766 8.86882 1.27898 5.0739 5.0739L6.13456 6.13456C9.33366 2.93545 14.5572 2.95404 17.8017 6.19852L18.8623 5.13786ZM5.0739 5.0739L3.80752 6.34028L4.86818 7.40094L6.13456 6.13456L5.0739 5.0739ZM4.33408 7.6206L6.87963 7.63339L6.88717 6.13341L4.34162 6.12062L4.33408 7.6206ZM5.08784 6.86684L5.07505 4.32129L3.57507 4.32882L3.58786 6.87438L5.08784 6.86684ZM12 3.75C16.5563 3.75 20.25 7.44365 20.25 12H21.75C21.75 6.61522 17.3848 2.25 12 2.25V3.75ZM12 20.25C7.44365 20.25 3.75 16.5563 3.75 12H2.25C2.25 17.3848 6.61522 21.75 12 21.75V20.25ZM16.1245 19.1468C14.9118 19.8483 13.5039 20.25 12 20.25V21.75C13.7747 21.75 15.4407 21.2752 16.8755 20.4452L16.1245 19.1468ZM20.25 12C20.25 13.5039 19.8483 14.9118 19.1468 16.1245L20.4452 16.8755C21.2752 15.4407 21.75 13.7747 21.75 12H20.25ZM6.19569 6.13719C7.68707 4.66059 9.73646 3.75 12 3.75V2.25C9.32542 2.25 6.90113 3.32791 5.14033 5.07126L6.19569 6.13719Z" fill={hovering == "log" || selected == "log" ? white : gray}></path>
                </g>
            </svg>
        </div>

        <div class="button {selected == "setting" ? "selected" : ""}" on:click={() => {select("setting")}} on:mouseenter={() => hover("setting")} on:mouseleave={hoverout}> <!-- setting -->
            <svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg">
                <path fill={hovering == "setting" || selected == "setting" ? white : gray} d="M600.704 64a32 32 0 0 1 30.464 22.208l35.2 109.376c14.784 7.232 28.928 15.36 42.432 24.512l112.384-24.192a32 32 0 0 1 34.432 15.36L944.32 364.8a32 32 0 0 1-4.032 37.504l-77.12 85.12a357.12 357.12 0 0 1 0 49.024l77.12 85.248a32 32 0 0 1 4.032 37.504l-88.704 153.6a32 32 0 0 1-34.432 15.296L708.8 803.904c-13.44 9.088-27.648 17.28-42.368 24.512l-35.264 109.376A32 32 0 0 1 600.704 960H423.296a32 32 0 0 1-30.464-22.208L357.696 828.48a351.616 351.616 0 0 1-42.56-24.64l-112.32 24.256a32 32 0 0 1-34.432-15.36L79.68 659.2a32 32 0 0 1 4.032-37.504l77.12-85.248a357.12 357.12 0 0 1 0-48.896l-77.12-85.248A32 32 0 0 1 79.68 364.8l88.704-153.6a32 32 0 0 1 34.432-15.296l112.32 24.256c13.568-9.152 27.776-17.408 42.56-24.64l35.2-109.312A32 32 0 0 1 423.232 64H600.64zm-23.424 64H446.72l-36.352 113.088-24.512 11.968a294.113 294.113 0 0 0-34.816 20.096l-22.656 15.36-116.224-25.088-65.28 113.152 79.68 88.192-1.92 27.136a293.12 293.12 0 0 0 0 40.192l1.92 27.136-79.808 88.192 65.344 113.152 116.224-25.024 22.656 15.296a294.113 294.113 0 0 0 34.816 20.096l24.512 11.968L446.72 896h130.688l36.48-113.152 24.448-11.904a288.282 288.282 0 0 0 34.752-20.096l22.592-15.296 116.288 25.024 65.28-113.152-79.744-88.192 1.92-27.136a293.12 293.12 0 0 0 0-40.256l-1.92-27.136 79.808-88.128-65.344-113.152-116.288 24.96-22.592-15.232a287.616 287.616 0 0 0-34.752-20.096l-24.448-11.904L577.344 128zM512 320a192 192 0 1 1 0 384 192 192 0 0 1 0-384zm0 64a128 128 0 1 0 0 256 128 128 0 0 0 0-256z"/>
            </svg>
        </div>
    </div>
</main>

<style lang="scss">
    main{
        display: inline-block;
        width: 100vw;
        flex-direction: column;
        background: linear-gradient(to bottom, rgba(18, 18, 18, 0) 0%, rgba(18, 18, 18, 0.7) 50%, rgba(18, 18, 18, 1) 100%);
    }

    #innerCon{
        display: flex;
        flex-direction: row;
        justify-content: space-around;
        // gap: 20px;
        // padding-top: 10px;
    }

    .button{
        width: 20%;
        text-align: center;
        border-bottom: 3px solid #121212;
    }

    .button:hover{
        cursor: pointer;
    }

    .button.selected{
        border-bottom: 3px solid #2f5fff;
    }

    svg{
        width: 40px;
        height: 40px;
        padding: 10px;
    }
</style>