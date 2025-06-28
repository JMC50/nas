<script lang="ts">
    import { onDestroy, onMount } from "svelte";
    import FileViewer from "./FileViewer.svelte";

    export let openFiles:string[];
    export let opened_file:number;
    export let fileList:file[];

    interface file{
        name: string;
        loc: string;
        extensions: string;
        modified: boolean;
    }

    const fileIconsAvailable = ["ai", "avi", "css", "csv", "dbf", "doc", "dwg", "exe", "fla", "html", "iso", "jpeg", "jpg", "js", "json", "mov", "mp3", "mp4", "pdf", "png", "ppt", "psd", "rtf", "svg", "txt", "wav", "xls", "xml", "zip", "unitypackage"];
    let openedFile:file;
    let fileListCon:HTMLDivElement;

    $:{
        openFiles = openFiles;
        opened_file = opened_file;
        openedFile = fileList[opened_file];
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
            const targetElement = fileElements[opened_file];
    
            if(targetElement){
                targetElement.scrollIntoView({ behavior: "smooth", block: "nearest", inline: "center" });
            }
        }, 10);
    }

    $: opened_file, scrollToFile();

    function changeOpen(index:number, e:MouseEvent) {
        const target = e.target as HTMLDivElement;
        if(target.classList.contains("close")) return;
        opened_file = index
    }

    function closeFile(index:number) {
        if(fileList[index].modified){
            const check = confirm("해당 파일은 저장되지 않았습니다. 저장하지 않고 종료하시겠습니까?");
            if(!check) return;
        }

        if(opened_file >= index && fileList.length != 1){
            opened_file = opened_file - 1;
        }
        if(opened_file == -1){
            opened_file = 0;
        }
        fileList.splice(index, 1);
        openFiles.splice(index, 1);
        fileList = fileList;
        openFiles = openFiles;
    }

</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<main>
    <div id="fileListCon" bind:this={fileListCon}>
        {#each fileList as file, index}
            <div class="file {opened_file == index ? "opened" : ""}" on:click={(e) => {changeOpen(index, e)}}>
                <img src={fileIconsAvailable.includes(file.extensions) ? `/png/${file.extensions}.png` : "/png/file.png"} alt="" class="icon">
                <div class="fileName">{file.name}</div>
                <div class="close" on:click={() => {closeFile(index)}}>{file.modified ? "●" : "X"}</div>
            </div>
        {/each}
    </div>
    <div id="viewerCon">
        {#each fileList as file, index (file.name)}
            <div class={opened_file == index ? "show" : "hide"}>
                <FileViewer bind:file={file} bind:fileList />
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