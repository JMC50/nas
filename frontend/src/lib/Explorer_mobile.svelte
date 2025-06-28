<script lang="ts">
    import { onMount } from "svelte";
    import { useAuth } from "../store/store";

    interface folderData {
        name: string;
        isFolder: boolean;
        extensions: string;   
    }

    interface uploadFileData {
        name: string;
        loc: string;
        percent: string;
        extensions: string;
    }

    interface searchData{
        name: string;
        isFolder: boolean;
        extensions: string;
        loc: string;
    }

    interface fileInfo{
        name: string;
        size: string;
        type: string;
        createdAt: string;
        modifiedAt: string;
    }

    interface copyData {
        loc: string;
        name: string;
    }

    type intentList = "ADMIN"|"VIEW"|"OPEN"|"DOWNLOAD"|"UPLOAD"|"COPY"|"DELETE"|"RENAME";
    
    let currentPath:string[] = [];
    let currentFiles:folderData[];
    let sortedFiles:folderData[];
    let fileIconsAvailable = ["ai", "avi", "css", "csv", "dbf", "doc", "dwg", "exe", "fla", "html", "iso", "jpeg", "jpg", "js", "json", "mov", "mp3", "mp4", "pdf", "png", "ppt", "psd", "rtf", "svg", "txt", "wav", "xls", "xml", "zip", "unitypackage"];
    let sortValue:string = "";
    let searchValue:string = "";
    let input:HTMLInputElement;
    let fileInformation:fileInfo;
    let clipboard:copyData[] = [];
    let uploadingFiles: uploadFileData[] = [];
    let itemCon:HTMLElement;
    let searchedData:searchData[] = [];
    let tempSelect:number[] = [];
    let loading = new Promise(res => {});

    $: {
        currentPath = currentPath;
        sortedFiles = sortedFiles;
        fileInformation = fileInformation;
        uploadingFiles = uploadingFiles;

        if(sortedFiles){
            sortedFiles = [...sortedFiles].sort((a, b) => {
                if(a.isFolder !== b.isFolder){
                    return a.isFolder ? -1 : 1;
                }
                return a.name.localeCompare(b.name);
            })
        }
    }

    async function checkLogin(checkIntent:intentList) {
        if($useAuth.userId == ""){
            alert("You need to log in to use this function.");
            return false;
        }

        const res = await fetch(`/server/checkIntent?intent=${checkIntent}&token=${$useAuth.token}`);
        const data = await res.json();
        if(data.status){
            return true;
        }else{
            alert("You don't have permission to use this function.");
            return false;
        }
    }

    onMount(async() => {
        async function uploadFile(files:FileList) {
            let sizes:number[] = [];
            let totalSize = 0;
            for(let file of files){
                if(file.size > 12 * 1024 * 1024 * 1024){
                    alert("업로드하려는 파일의 크기는 최대 12GB입니다다.");
                    return;
                }
                sizes.push(file.size);
                totalSize += file.size;
            }
            
            let uploadedSize = 0;
            const currentLoc = currentPath.join("/");
            
            for(let i = 0; i < files.length; i++){
                const file = files[i];
                
                const match = file.name.match(/\.([^.]+)$/);
                let extensions = match ? match[1] : "file";

                const uploadData: uploadFileData = {
                    name: file.name,
                    loc: currentLoc,
                    percent: "0",
                    extensions
                };
                uploadingFiles.push(uploadData);
                uploadingFiles = uploadingFiles;

                new Promise<void>(async (resolve, reject) => {
                    const xhr = new XMLHttpRequest();
                    xhr.open("POST", `/server/input?name=${encodeURIComponent(file.name)}&loc=/${encodeURIComponent(currentLoc)}&token=${$useAuth.token}`, true);

                    let lastUploaded = 0;

                    xhr.upload.onprogress = (event) => {
                        if(event.lengthComputable){
                            let chunkSize = event.loaded - lastUploaded;
                            uploadedSize += chunkSize;
                            lastUploaded = event.loaded;

                            let progress = ((uploadedSize / totalSize) * 100).toFixed(2);
                            uploadData.percent = progress;
                            uploadingFiles = uploadingFiles;
                        }
                    };
                    
                    xhr.onload = () => {
                        if(xhr.status === 200){
                            uploadingFiles = uploadingFiles.filter(f => f.name !== file.name);
                            loading = getFiles();
                            resolve();
                        }else{
                            reject(`Error ${xhr.status}: ${xhr.statusText}`);
                        }
                    };
                    
                    xhr.onerror = () => reject("네트워크 오류 발생");
                    if(file.type == "application/json"){
                        const text = await file.text();
                        xhr.send(text);
                    }else{
                        xhr.send(file);
                    }
                });
            }
            loading = getFiles();
        }

        input.addEventListener('change', async e => {
            if(input.files){
                uploadFile(input.files);
            }
        });

        window.addEventListener("click", e => {
            const target = e.target as HTMLElement;
            if(!target.classList.contains("menu")){
                showContext = false;
            }
        })
    })

    async function deleteFile() {
        const loginCheck = await checkLogin("DELETE");
        if(!loginCheck) return;

        const check = confirm(`파일/폴더 ${tempSelect.length} 개가 선택되었습니다. 정말로 삭제하시겠습니까?`);
        if(!check) return;
        const promises = [];
        for(let i of tempSelect){
            promises.push(fetch(`/server/forceDelete?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(sortedFiles[i].name)}&token=${$useAuth.token}`));
        }
        await Promise.all(promises);
        tempSelect = [];
        showContext = false;
        loading = getFiles();
    }

    async function renameFile() {
        if(tempSelect.length == 0) return;
        const loginCheck = await checkLogin("RENAME");
        if(!loginCheck) return;

        const selectedFile = sortedFiles[tempSelect[0]];
        const newName = prompt("새로운 파일 이름을 입력하세요. (확장자는 변경할 수 없습니다.)", selectedFile.name);
        if(!newName || newName == "") return;
        const match = selectedFile.name.match(/\.([^.]+)$/);
        if(match){
            const check = newName.match(/\.([^.]+)$/);
            if(check){
                if(match[1] == check[1]){
                    const res = await fetch(`/server/rename?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(selectedFile.name)}&change=${encodeURIComponent(newName)}&token=${$useAuth.token}`);
                    const data = await res.text();
                    if(data !== "complete"){
                        alert("파일의 이름을 바꾸는데에 실패하였습니다.");
                    }
                }else{
                    const res = await fetch(`/server/rename?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(selectedFile.name)}&change=${encodeURIComponent(`${newName}.${match[1]}`)}&token=${$useAuth.token}`);
                    const data = await res.text();
                    if(data !== "complete"){
                        alert("파일의 이름을 바꾸는데에 실패하였습니다.");
                    }
                }
            }else{
                const res = await fetch(`/server/rename?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(selectedFile.name)}&change=${encodeURIComponent(`${newName}.${match[1]}`)}&token=${$useAuth.token}`);
                const data = await res.text();
                if(data !== "complete"){
                    alert("파일의 이름을 바꾸는데에 실패하였습니다.");
                }
            }

            loading = getFiles();
        }
        showContext = false;
    }

    const serverURL = process.env.SERVER_URL as string;

    async function downloadFile() {
        window.open(`${serverURL}/download?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(sortedFiles[tempSelect[0]].name)}&token=${$useAuth.token}`);
        showContext = false;
    }

    async function copyDonwloadLink() {
        navigator.clipboard.writeText(`${serverURL}/download?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(sortedFiles[tempSelect[0]].name)}&token=${$useAuth.token}`).then(() => {
            alert("다운로드 링크가 클립보드에 복사되었습니다.");
        })
        showContext = false;
    }

    let showFileInformation:boolean = false;

    async function getfileInformation() {
        const res = await fetch(`/server/stat?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(sortedFiles[tempSelect[0]].name)}`);
        fileInformation = await res.json() as fileInfo;
        showFileInformation = true;
        showContext = false;
    }

    async function close() {
        showFileInformation = false;
    }

    async function detectClose(e:MouseEvent) {
        const target = e.target as HTMLElement;
        if(target.id == "statBackground"){
            showFileInformation = false;
        }
    }

    async function getFiles() {
        const res = await fetch(`/server/readFolder?loc=/${encodeURIComponent(currentPath.join("/"))}&token=${$useAuth.token}`);
        const data = await res.json() as folderData[];
        currentFiles = data;
        sortedFiles = [];
        const uploads:string[] = [];
        const locs:string[] = [];
        for(let file of uploadingFiles){
            uploads.push(file.name);
            locs.push(`${file.loc}/${file.name}`);
        }
        for(let i = 0; i < currentFiles.length; i++){
            if(!locs.includes(`${currentPath.join("/")}/${currentFiles[i].name}`)){
                sortedFiles.push(currentFiles[i]);
            }
        }
        return;
    }

    async function copyFiles() {
        clipboard = [];
        for(let i of tempSelect){
            clipboard.push({
                loc: currentPath.join("/"),
                name: sortedFiles[i].name
            })
        };
        showContext = false;
    }

    async function pasteFiles() {
        const loginCheck = await checkLogin("COPY");
        if(!loginCheck) return;

        if(clipboard.length == 0) return;
        const promises = [];
        for(let data of clipboard){
            promises.push(fetch(`/server/copy?originLoc=/${encodeURIComponent(data.loc)}&fileName=${encodeURIComponent(data.name)}&targetLoc=/${encodeURIComponent(currentPath.join("/"))}&token=${$useAuth.token}`));
        }
        await Promise.all(promises);
        loading = getFiles();
        showContext = false;
    }

    async function click(target:number) {
        if(target == -1){
            currentPath.pop();
            currentPath = currentPath;
            loading = getFiles();
        }else{
            tempSelect = [target];
            if(searchValue == ""){
                if(sortedFiles[target].isFolder){
                    currentPath.push(sortedFiles[target].name);
                    currentPath = currentPath;
                    loading = getFiles();
                }else{
                    window.open(`${location.origin}/FileViewer?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(sortedFiles[target].name)}&token=${$useAuth.token}`);
                }
                tempSelect = [];
            }else{
                if(searchedData[target].isFolder){
                    currentPath = searchedData[target].loc.split("/");
                    currentPath = currentPath;
                    searchValue = "";
                    loading = getFiles();
                }else{
                    window.open(`${location.origin}/FileViewer?loc=/${encodeURIComponent(searchedData[target].loc)}&name=${encodeURIComponent(searchedData[target].name)}&token=${$useAuth.token}`);
                }
                searchValue = "";
                tempSelect = [];
            }
        }
    }

    async function upload() {
        const loginCheck = await checkLogin("UPLOAD");
        if(!loginCheck) return;

        input.click();
    }

    async function makedir() {
        const loginCheck = await checkLogin("UPLOAD");
        if(!loginCheck) return;
        
        const dirname = prompt("생성할 폴더의 이름을 입력하세요.")
        if(!dirname || dirname == "") return;
        const res = await fetch(`/server/makedir?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(dirname)}&token=${$useAuth.token}`);
        const data = await res.text();
        if(data == "complete"){
            loading = getFiles();
        }else{
            alert("동일한 이름의 폴더가 이미 존재합니다.");
        }
        showContext = false;
    }

    async function sort() {
        tempSelect = [];
        const arr = [];
        for(let file of currentFiles){
            if(file.name.toLowerCase().includes(sortValue.toLowerCase())){
                arr.push(file);
            }
        }
        sortedFiles = arr;
    }

    async function search() {
        currentPath = [];
        const res = await fetch(`/server/searchInAllFiles?query=${encodeURIComponent(searchValue)}&token=${$useAuth.token}`);
        const data = await res.json() as searchData[];

        searchedData = data;
    }

    let showContext:boolean = false;
    let contextCon:HTMLDivElement;
    let contextType:"file"|"none" = "file";

    async function context(e:MouseEvent) {
        e.preventDefault();
        const target = e.target as HTMLElement;
        let selected:number = -1;
        if(target.classList.contains("back")) return;
        if(target.classList.contains("item")){
            selected = Number(target.dataset.index);
        }else{
            if(target.parentElement?.classList.contains("item")){
                const temp = target.parentElement as HTMLElement;
                selected = Number(temp.dataset.index);
            }
        }

        if(selected != -1){
            if(!tempSelect.includes(selected)){
                tempSelect = [selected];
            }
            contextType = "file";
        }else{
            contextType = "none";
        }

        contextCon.style.top = `${e.pageY}px`;
        contextCon.style.left = `${e.pageX}px`;
        showContext = true;
    }

    loading = getFiles();
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<main on:contextmenu={context} id="drag" bind:this={itemCon}>
    <div class="title">📂 File Explorer</div>
    <div class="subtitle">
        <div>Current Location : </div>
        <div>/{currentPath.join("/")}</div>
    </div>
    <div id="itemCon">
        {#await loading}
        <div class="subtitle">Fetching Data...</div>
        {:then value}
            <div id="functionsCon">
                <div id="uploadbtn" on:click={upload}>File Upload</div>
                <div id="uploadbtn" on:click={makedir}>New Dir</div>
            </div>
            <div id="additional_input">
                <div id="searchCon">
                    <img src="/svg/search_black.svg" alt="">
                    <input type="search" placeholder="Search in current dir" id="searchInput" bind:value={sortValue} on:input={sort} autocomplete="off">
                </div>
                <div id="searchCon">
                    <img src="/svg/search_black.svg" alt="">
                    <input type="search" placeholder="Search In All Files" id="searchInput" bind:value={searchValue} on:input={search} autocomplete="off">
                </div>
            </div>
            {#if searchValue == ""}
                {#if currentPath.length !== 0}
                    <div class="back {tempSelect.length == 1 && tempSelect[0] == -1 ? "tempSelect" : ""}" on:click={() => {click(-1)}} >
                        <img src='/png/back.png' alt="" class="item_icon">
                        <div class="item_name">..</div>
                    </div>
                {/if}
                {#each uploadingFiles as file}
                    <div class="uploading">
                        <div class="left">
                            <img src={fileIconsAvailable.includes(file.extensions) ? `/png/${file.extensions}.png` : "/png/file.png"} alt="" class="item_icon">
                            <div class="item_textCon">
                                <div class="item_name">{file.name}</div>
                                <div class="item_loc">/{file.loc}</div>
                            </div>
                        </div>
                        <div class="right">
                            <div class="percent">{file.percent}%</div>
                        </div>
                    </div>
                {/each}
                {#each sortedFiles as data, index}
                    <div class="item {data.isFolder ? "folder" : "file"} {tempSelect.includes(index) ? "tempSelect" : ""}" on:click={() => {click(index)}} data-index={index}>
                        <img src={data.isFolder ? "/png/folder.png" : `${fileIconsAvailable.includes(data.extensions) ? `/png/${data.extensions}.png` : "/png/file.png"}`} alt="" class="item_icon">
                        <div class="item_name">{data.name}</div>
                    </div>
                {/each}
            {:else}
                {#each searchedData as data, index}
                    <div class="item {data.isFolder ? "folder" : "file"} {tempSelect.includes(index) ? "tempSelect" : ""}" on:click={() => {click(index)}} data-index={index}>
                        <img src={data.isFolder ? "/png/folder.png" : `${fileIconsAvailable.includes(data.extensions) ? `/png/${data.extensions}.png` : "/png/file.png"}`} alt="" class="item_icon">
                        <div class="item_textCon">
                            <div class="item_name">{data.name}</div>
                            <div class="item_loc">/{data.loc}</div>
                        </div>
                    </div>
                {/each}
            {/if}
        {/await}
    </div>
    <input type="file" multiple bind:this={input} class="none">
    <div id="contextCon" class={showContext ? "" : "none"} bind:this={contextCon}>
        {#if contextType == "file"}
            <div class="menu topRadius" on:click={deleteFile}>Delete</div>
            {#if tempSelect.length == 1}
                <div class="menu" on:click={renameFile}>Rename</div>
                {#if !sortedFiles[tempSelect[0]].isFolder}
                    <div class="menu" on:click={downloadFile}>Download</div>
                    <div class="menu" on:click={copyDonwloadLink}>Copy Download Link</div>
                {/if}
                <div class="menu" on:click={getfileInformation}>Information</div>
            {/if}
            <div class="menu bottomRadius" on:click={copyFiles}>Copy File(s)</div>
        {:else}
            <div class="menu topRadius" on:click={makedir}>New Dir</div>
            <div class="menu bottomRadius" on:click={pasteFiles}>Paste</div>
        {/if}
    </div>
</main>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
{#if fileInformation && showFileInformation}
<div id="statBackground" class={showFileInformation ? "" : "none"} on:click={detectClose}>
    <div id="statCon">
        <div class="top">
            <div></div>
            <div class="information_title">FILE INFO</div>
            <div class="information_title close" on:click={close}>X</div>
        </div>
        <div id="contents">
            <table>
                <tbody>
                    <tr>
                        <td class="content_title">NAME : </td>
                        <td class="left" id="statName">{fileInformation.name}</td>
                    </tr>
                    <tr>
                        <td class="content_title">SIZE : </td>
                        <td class="left" id="statSize">{fileInformation.size}</td>
                    </tr>
                    <tr>
                        <td class="content_title">TYPE : </td>
                        <td class="left" id="statType">{fileInformation.type}</td>
                    </tr>
                    <tr>
                        <td class="content_title">CREATED AT : </td>
                        <td class="left" id="statCreatedAt">{fileInformation.createdAt}</td>
                    </tr>
                    <tr>
                        <td class="content_title">MODIFIED AT : </td>
                        <td class="left" id="statModifiedAt">{fileInformation.modifiedAt}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</div>
{:else}
<div></div>
{/if}


<style lang="scss">
    main{
        // height: 100vh;
        overflow-y: auto;
        overflow-x: hidden;
        padding: 30px 20px 200px 20px;
        user-select: none;
        border: 2px solid #121212;
        border-radius: 10px;
        transition: all 0.2s ease;
    }

    .title{
        font-size: xx-large;
        color: white;
        font-weight: bolder;
        margin-bottom: 20px;
    }

    .subtitle{
        display: flex;
        flex-direction: column;
        gap: 10px;
        font-size: x-large;
        color: white;
        font-weight: bolder;
    }

    #itemCon{
        display: flex;
        flex-direction: column;
        margin-top: 30px;
        gap: 5px;
    }

    #functionsCon{
        display: flex;
        flex-direction: row;
        align-items: center;
        gap: 10px;
        width: 100%;
        margin-bottom: 10px;
    }

    #additional_input{
        display: flex;
        flex-direction: column;
        gap: 10px;
        margin-bottom: 30px;
    }

    #uploadbtn{
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
        width: 50%;
        text-align: center;
    }

    #uploadbtn:hover{
        cursor: pointer;
        background-color: rgb(92, 92, 92);
    }

    #searchCon{
        background-color: white;
        flex-grow: 1;
        border-radius: 5px;
        display: flex;
        align-items: center;
        padding: 10px 15px;
        gap: 10px;
    }

    #searchInput{
        background: none;
        font-size: large;
        border: 0;
        height: auto;
        width: 100%;
    }

    input:focus {outline:none;}

    .back{
        display: flex;
        flex-direction: row;
        align-items: center;
        border: 1px solid #121212;
        color: white;
        gap: 10px;
        padding: 10px;
    }

    .uploading{
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        align-items: center;
        border: 1px solid #121212;
        background-color: #4e86ff;
        color: white;
        padding: 10px;
    }

    .left{
        display: flex;
        flex-direction: row;
        gap: 10px;
        align-items: center;
    }

    .item{
        display: flex;
        flex-direction: row;
        align-items: center;
        border: 1px solid #121212;
        color: white;
        gap: 10px;
        padding: 10px;
    }

    .tempSelect{
        border: 1px solid #878787;
    }

    .item_icon{
        height: 30px;
    }

    .item_textCon{
        display: flex;
        flex-direction: column;
    }

    .item_name {
        color: white;
        font-size: larger;
    }

    .item_loc {
        color: #c9c9c9;
        font-size: small;
    }

    .none{
        display: none;
    }

    #contextCon{
        position: absolute;
        background-color: #383838;
        box-shadow: 4px 4px 4px #000000;
        border-radius: 10px;
    }

    .menu{
        color: white;
        font-weight: bolder;
        font-size: medium;
        padding: 10px;
    }

    .menu:hover{
        cursor: pointer;
        background-color: rgb(92, 92, 92);
    }

    .topRadius{
        border-top-left-radius: 10px;
        border-top-right-radius: 10px;
    }
    
    .bottomRadius{
        border-bottom-left-radius: 10px;
        border-bottom-right-radius: 10px;
    }

    #statBackground{
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background-color: rgba(0, 0, 0, 0.5);
        display: flex;
        justify-content: center;
        align-items: center;
    }

    #statCon{
        display: flex;
        flex-direction: column;
        gap: 20px;
        padding: 20px 20px 20px 20px;
        border-radius: 20px;
        background-color: #383838;
        box-shadow: 4px 4px 4px #000000;
        color: white;
        font-weight: bolder;
    }

    .left{
        text-align: left;
    }

    .top{
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        font-size: x-large;
    }

    .close:hover{
        cursor: pointer;
    }

    #contents{
        display: flex;
        flex-direction: column;
        gap: 10px;
    }
</style>