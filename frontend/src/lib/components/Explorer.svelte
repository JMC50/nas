<script lang="ts">
    import { onMount } from "svelte";
    import { BlobWriter, ZipWriter, BlobReader } from "@zip.js/zip.js";
    import { useAuth } from "../store/store";
    
    type intentList = "ADMIN"|"VIEW"|"OPEN"|"DOWNLOAD"|"UPLOAD"|"COPY"|"DELETE"|"RENAME";

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
        action: "uploading"|"zipping"|"unzipping"|"error";
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

    interface file{
        name: string;
        loc: string;
        extensions: string;
        modified: boolean;
    }

    interface copyData {
        loc: string;
        name: string;
    }

    export let openFiles:string[];
    export let opened_file:number;
    export let fileList:file[];
    export let currentPath:string[];
    let currentFiles:folderData[];
    let sortedFiles:folderData[];
    let tempSelect:number[] = [];
    let lastClicked:number = NaN;
    let multiSelect:boolean = false;
    let shiftSelect:boolean = false;
    let fileIconsAvailable = ["ai", "avi", "css", "csv", "dbf", "doc", "dwg", "exe", "fla", "html", "iso", "jpeg", "jpg", "js", "json", "mov", "mp3", "mp4", "pdf", "png", "ppt", "psd", "rtf", "svg", "txt", "wav", "xls", "xml", "zip", "folder", "unitypackage"];
    let sortValue:string = "";
    let searchValue:string = "";
    let input:HTMLInputElement;
    let fileInformation:fileInfo;
    let clipboard:copyData[] = [];
    let uploadingFiles: uploadFileData[] = [];
    let itemCon:HTMLElement;
    let searchedData:searchData[] = [];
    let loading = new Promise(res => {});
    const uploadQueue: { file: File; uploadLoc: string }[] = [];
    let activeUploads:number = 0;
    const MAX_CONCURRENT_UPLOADS:number = 10;
    let isProcessingQueue:boolean = false;
    let preparingUpload:boolean = false;
    let dragOverIndex: number | null = null;
    let dragOverBack: boolean = false;

    $: {
        currentPath = currentPath;
        sortedFiles = sortedFiles;
        fileInformation = fileInformation;
        uploadingFiles = uploadingFiles;
        preparingUpload = preparingUpload;

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


    async function uploadFile(files: FileList | File[], uploadLoc: string) {
        const loginCheck = await checkLogin("UPLOAD");
        if(!loginCheck) return;

        for(let file of files){
            if(file.size > 50 * 1024 * 1024 * 1024){
                alert("업로드하려는 파일의 크기는 최대 50GB입니다.");
                preparingUpload = false;
                return;
            }
            const match = file.name.match(/\.([^.]+)$/);
            let extensions = match ? match[1] : "file";
            
            const uploadData: uploadFileData = {
                name: file.name,
                loc: uploadLoc,
                percent: "0",
                extensions,
                action: "uploading"
            };
            preparingUpload = false;
            uploadingFiles.push(uploadData);
            uploadingFiles = uploadingFiles;

            uploadQueue.push({ file, uploadLoc });
        }
        
        startUploadQueue();
    }

    async function startUploadQueue() {
        if(isProcessingQueue) return;
        isProcessingQueue = true;

        while(uploadQueue.length > 0 || activeUploads > 0){
            if(uploadQueue.length === 0 || activeUploads >= MAX_CONCURRENT_UPLOADS){
                await new Promise(resolve => setTimeout(resolve, 200));
                continue;
            }

            const { file, uploadLoc } = uploadQueue.shift()!;
            activeUploads++;

            new Promise<void>(async (resolve, reject) => {
                const xhr = new XMLHttpRequest();
                xhr.open("POST", `/server/input?name=${encodeURIComponent(file.name)}&loc=/${encodeURIComponent(uploadLoc)}&token=${$useAuth.token}`, true);

                xhr.upload.onprogress = (event) => {
                    if(event.lengthComputable){
                        const progress = ((event.loaded / file.size) * 100).toFixed(2);
                        const uploadData = uploadingFiles.find(f => {
                            if(file.webkitRelativePath){
                                return `${f.loc}/${f.name}` === file.webkitRelativePath;
                            }else{
                                return f.name === file.name && f.loc === uploadLoc;
                            }
                        });
                        if(uploadData){
                            uploadData.percent = progress;
                            uploadingFiles = uploadingFiles;
                        }
                    }
                };

                xhr.onload = () => {
                    activeUploads--;
                    if(xhr.status === 200){
                        const removedFiles = uploadingFiles.filter(f => f.name === file.name);
                        uploadingFiles = uploadingFiles.filter(f => f.name !== file.name);

                        for(let file of removedFiles){
                            if(file.loc == currentPath.join("/")){
                                sortedFiles.push({
                                    name: file.name,
                                    extensions: file.extensions,
                                    isFolder: false
                                });
                            }
                        }
                        sortedFiles = sortedFiles;

                        if(uploadingFiles.length == 0){
                            loading = getFiles();
                        }
                        resolve();
                    }else{
                        reject(`Error ${xhr.status}: ${xhr.statusText}`);
                    }
                };

                xhr.onerror = () => {
                    activeUploads--;
                    reject("네트워크 오류 발생");
                };

                if(file.type == "application/json"){
                    const text = await file.text();
                    xhr.send(text);
                }else{
                    xhr.send(file);
                }
            }).catch(error => console.error(error));
        }

        isProcessingQueue = false;
    }

    interface FileWithPath {
        file: File;
        relativePath: string;
    }

    async function uploadFilesAsZip(fileEntries: FileWithPath[], uploadLoc: string): Promise<void> {
        const loginCheck = await checkLogin("UPLOAD");
        if(!loginCheck) return;

        const total = fileEntries.reduce((sum, entry) => sum + entry.file.size, 0);
        if(total > 50 * 1024 * 1024 * 1024){
            alert("업로드하려는 파일의 크기는 최대 50GB입니다.");
            preparingUpload = false;
            return;
        }

        const foldername = fileEntries[0].relativePath.split("/")[0];
        const batchUploadData: uploadFileData = {
            name: foldername,
            loc: uploadLoc,
            percent: "0",
            extensions: "folder",
            action: "zipping"
        };
        preparingUpload = false;
        uploadingFiles.push(batchUploadData);
        uploadingFiles = uploadingFiles;

        const blobWriter = new BlobWriter("application/zip");
        const zipWriter = new ZipWriter(blobWriter);

        let totalAdded = 0;

        for(const entry of fileEntries){
            await zipWriter.add(entry.relativePath, new BlobReader(entry.file), { level: 0 });
            totalAdded += entry.file.size;
            batchUploadData.percent = ((totalAdded / total) * 100).toFixed(2);
            uploadingFiles = uploadingFiles;
        }
        const zipBlob = await zipWriter.close();

        const xhr = new XMLHttpRequest();
        batchUploadData.action = "uploading";
        xhr.open("POST", `/server/inputZip?name=${encodeURIComponent(foldername)}.zip&loc=/${encodeURIComponent(uploadLoc)}&token=${$useAuth.token}`, true);

        xhr.upload.onprogress = (event) => {
            if(event.lengthComputable){
                batchUploadData.percent = ((event.loaded / event.total) * 100).toFixed(2);
                if(Math.floor(Number(batchUploadData.percent)) == 100){
                    batchUploadData.action = "unzipping";
                }
                uploadingFiles = uploadingFiles;
            }
        };

        xhr.onload = () => {
            if(xhr.status === 200){
                uploadingFiles = uploadingFiles.filter((f) => f !== batchUploadData);
                sortedFiles.push({ name: foldername, extensions: "folder", isFolder: true });
                sortedFiles = sortedFiles;
                if(uploadingFiles.length === 0){
                    loading = getFiles();
                }
            }else{
                console.error(`Error ${xhr.status}: ${xhr.statusText}`);
            }
        };

        xhr.onerror = () => {
            console.error("네트워크 오류 발생");
        };

        xhr.send(zipBlob);
    }

    onMount(async() => {
        function traverseFileTree(item: any, path: string = ""): Promise<FileWithPath[]> {
            return new Promise((resolve) => {
                if(item.isFile){
                    item.file(
                        (file: File) => {
                            resolve([{ file, relativePath: path + file.name }]);
                        },
                        (error: any) => {
                            console.error(`파일 읽기 실패 (${path}):`, error);
                            resolve([]);
                        }
                    );
                }else if(item.isDirectory){
                    const dirReader = item.createReader();
                    let entries: any[] = [];

                    const readEntries = () => {
                        dirReader.readEntries((results: any[]) => {
                        if(!results.length){
                            Promise.all(
                                entries.map((entry) =>
                                    traverseFileTree(entry, path + item.name + "/")
                                )
                            ).then((fileArrays) => {
                                resolve(fileArrays.flat());
                            });
                        }else{
                            entries = entries.concat(Array.from(results));
                            readEntries();
                        }
                        });
                    };
                    readEntries();
                }
            });
        }


        input.addEventListener('change', async e => {
            if(input.files){
                uploadFile(input.files, currentPath.join("/"));
            }
        });

        function handleDragOver(event: DragEvent) {
            event.preventDefault();
            if(itemCon){
                itemCon.style.borderLeft = "2px solid #FFFFFF";
                itemCon.style.borderRight = "2px solid #FFFFFF";
            }
        }

        function handleDragLeave() {
            if(itemCon){
                itemCon.style.borderLeft = "2px solid #121212";
                itemCon.style.borderRight = "2px solid #121212";
            }
        }

        async function handleDrop(event: DragEvent) {
            event.preventDefault();
            if(itemCon){
                itemCon.style.borderLeft = "2px solid #121212";
                itemCon.style.borderRight = "2px solid #121212";
            }

            const dt = event.dataTransfer;
            if(!dt) return;

            // const loginCheck = await checkLogin("UPLOAD");
            // if(!loginCheck) return;

            const items = dt.items;
            if(!items || items.length === 0) return;

            let isFolderDrop = false;
            for(let i = 0; i < items.length; i++){
                const entry = items[i].webkitGetAsEntry();
                if(entry && entry.isDirectory){
                    isFolderDrop = true;
                    break;
                }
            }

            if(isFolderDrop){
                preparingUpload = true;
                let fileWithPaths: FileWithPath[] = [];
                const promises: Promise<FileWithPath[]>[] = [];
                for(let i = 0; i < items.length; i++){
                    const entry = items[i].webkitGetAsEntry();
                    if(entry){
                        promises.push(traverseFileTree(entry));
                    }
                }
                const results = await Promise.all(promises);
                fileWithPaths = results.flat();

                const groups: { [folder: string]: FileWithPath[] } = {};
                fileWithPaths.forEach(f => {
                    const folder = f.relativePath.split("/")[0];
                    if (!groups[folder]) {
                        groups[folder] = [];
                    }
                    groups[folder].push(f);
                });
                for(const folder in groups){
                    uploadFilesAsZip(groups[folder], currentPath.join("/"));
                }
            }else{
                uploadFile(dt.files, currentPath.join("/"));
            }
        }

        window.addEventListener("click", e => {
            const target = e.target as HTMLElement;
            if(!target.classList.contains("menu")){
                showContext = false;
            }
        })

        window.addEventListener("keydown", e => {
            if(e.key == "Delete"){
                if(tempSelect.length > 0){
                    if(tempSelect[0] != -1){
                        deleteFile();
                    }
                }
            }

            if(e.ctrlKey && e.key === "c"){
                e.preventDefault();
                copyFiles();
            }

            if(e.ctrlKey && e.key === "v"){
                e.preventDefault();
                pasteFiles();
            }

            if(e.key == "Control"){
                multiSelect = true;
            }

            if(e.key == "Shift"){
                shiftSelect = true;
            }
        })

        window.addEventListener("keyup", e => {
            if(e.key == "Control"){
                multiSelect = false;
            }

            if(e.key == "Shift"){
                shiftSelect = false;
            }
        })

        itemCon.addEventListener("dragover", handleDragOver);
        itemCon.addEventListener("dragleave", handleDragLeave);
        itemCon.addEventListener("drop", handleDrop);
    })

    let isDeleting = false;
    
    async function deleteFile() {
        if(isDeleting) return;
        isDeleting = true;
        
        const loginCheck = await checkLogin("DELETE");
        if(!loginCheck) {
            isDeleting = false;
            return;
        }

        const check = confirm(`파일/폴더 ${tempSelect.length} 개가 선택되었습니다. 정말로 삭제하시겠습니까?`);
        if(!check) {
            isDeleting = false;
            return;
        }
        const promises = [];
        for(let i of tempSelect){
            promises.push(fetch(`/server/forceDelete?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(sortedFiles[i].name)}&token=${$useAuth.token}`));
        }
        await Promise.all(promises);
        tempSelect = [];
        showContext = false;
        loading = getFiles();
        isDeleting = false;
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
        if (tempSelect.length > 1 || (tempSelect.length === 1 && sortedFiles[tempSelect[0]].isFolder)) {
            const files = tempSelect.map(i => ({
                ...sortedFiles[i],
                loc: `${currentPath.join("/")}`
            }));
            // zipName은 서버에서 반환된 zipPath의 파일명으로 대체
            const uploadData: uploadFileData = {
                name: "압축 준비중...",
                loc: currentPath.join("/"),
                percent: "0",
                extensions: "zip",
                action: "zipping"
            };
            uploadingFiles.push(uploadData);
            uploadingFiles = uploadingFiles;
            // 서버에 zip 요청
            const res = await fetch(`/server/zipFiles?token=${$useAuth.token}`, {
                method: "POST",
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(files)
            });
            if (res.status === 200) {
                const { zipPath } = await res.json();
                // zipPath에서 파일명 추출
                const zipName = zipPath ? zipPath.split("/").pop() : `download_${Date.now()}.zip`;
                // zip 파일 다운로드 (GET)
                const downloadUrl = `/server/downloadZip?token=${$useAuth.token}&zipPath=${encodeURIComponent(zipPath)}`;
                const a = document.createElement('a');
                a.href = downloadUrl;
                a.download = zipName;
                document.body.appendChild(a);
                a.click();
                setTimeout(() => {
                    document.body.removeChild(a);
                    uploadingFiles = uploadingFiles.filter(f => f !== uploadData);
                }, 100);
                // 다운로드 후 서버에 zip 파일 삭제 요청
                if (zipPath) {
                    fetch(`/server/deleteTempZip?token=${$useAuth.token}&path=${encodeURIComponent(zipPath)}`);
                }
            } else {
                uploadData.action = "error";
            }
        } else {
            window.open(`${serverURL}/download?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(sortedFiles[tempSelect[0]].name)}&token=${$useAuth.token}`);
        }
        showContext = false;
    }

    async function copyDonwloadLink() {
        navigator.clipboard.writeText(`${serverURL}/download?loc=/${encodeURIComponent(currentPath.join("/"))}&name=${encodeURIComponent(sortedFiles[tempSelect[0]].name)}`).then(() => {
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

    let status:"no_intents"|"need_login"|"need_reLogin"|"clear" = "clear";

    async function getFiles() {
        const res = await fetch(`/server/readFolder?loc=/${encodeURIComponent(currentPath.join("/"))}&token=${$useAuth.token}`);

        if(res.status == 500){
            const data = await res.text();
            if(data == "no intents"){
                status = "no_intents";
            }else if(data == "wtf is this token"){
                status = "need_reLogin";
            }else if(data == "token required"){
                status = "need_login";
            }
            return;
        }

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
        const loginCheck = await checkLogin("COPY");
        if(!loginCheck) return;

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

    async function zipFiles() {
        const files = [];
        for(let i of tempSelect){
            files.push({
                ...sortedFiles[i],
                loc: `${currentPath.join("/")}`
            });
        }

        // UI에 표시할 데이터 추가
        const zipName = files.length === 1 ? files[0].name + ".zip" : "archive.zip";
        const uploadData:uploadFileData & { progressId?: string } = {
            name: zipName,
            loc: currentPath.join("/"),
            percent: "0",
            extensions: "zip",
            action: "zipping"
        };
        uploadingFiles.push(uploadData);
        uploadingFiles = uploadingFiles;

        const res = await fetch(`/server/zipFiles?token=${$useAuth.token}`, {
            method:"POST",
            headers:{ 'Content-Type':'application/json' },
            body: JSON.stringify(files)
        });

        if(res.status === 200){
            const { progressId } = await res.json();
            uploadData.progressId = progressId;
            pollProgress(uploadData, progressId);
        } else {
            uploadData.action = "error";
        }
    }

    async function unzip() {
        const file = {
            ...sortedFiles[tempSelect[0]],
            loc: `${currentPath.join("/")}`
        };

        // UI에 표시할 데이터 추가
        const uploadData:uploadFileData & { progressId?: string } = {
            name: file.name,
            loc: currentPath.join("/"),
            percent: "0",
            extensions: "zip",
            action: "unzipping"
        };
        uploadingFiles.push(uploadData);
        uploadingFiles = uploadingFiles;

        const res = await fetch(`/server/unzipFile?token=${$useAuth.token}`, {
            method:"POST",
            headers:{ 'Content-Type':'application/json' },
            body: JSON.stringify(file)
        });

        if(res.status === 200){
            const { progressId } = await res.json();
            uploadData.progressId = progressId;
            pollProgress(uploadData, progressId);
        } else {
            uploadData.action = "error";
        }
    }

    function pollProgress(uploadData: any, progressId: string) {
        let interval = setInterval(async () => {
            const res = await fetch(`/server/progress?progressId=${progressId}`);
            if(res.status === 200){
                const data = await res.json();
                uploadData.percent = String(data.percent);
                if(data.status === "done"){
                    uploadData.action = "uploading";
                    setTimeout(() => {
                        uploadingFiles = uploadingFiles.filter(f => f !== uploadData);
                        loading = getFiles();
                    }, 500);
                    clearInterval(interval);
                } else if(data.status === "error") {
                    uploadData.action = "error";
                    clearInterval(interval);
                } else if(data.status === "zipping") {
                    uploadData.action = "zipping";
                } else if(data.status === "unzipping") {
                    uploadData.action = "unzipping";
                }
            } else {
                uploadData.action = "error";
                clearInterval(interval);
            }
            uploadingFiles = uploadingFiles;
        }, 500);
    }

    async function click(target:number) {
        if(tempSelect.includes(target)){
            if(Date.now() - lastClicked <= 1000){
                if(tempSelect.length == 1 && tempSelect[0] == -1){
                    currentPath.pop();
                    currentPath = currentPath;
                    tempSelect = [];
                    loading = getFiles();
                }else{
                    if(searchValue == ""){
                        if(sortedFiles[target].isFolder){
                            currentPath.push(sortedFiles[target].name);
                            currentPath = currentPath;
                            loading = getFiles();
                        }else{
                            const file = `${currentPath.join("/")}/${sortedFiles[target].name}`;
                            if(openFiles.includes(file)){
                                opened_file = openFiles.indexOf(file);
                            }else{
                                openFiles.push(file);
                                fileList.push({
                                    name: sortedFiles[target].name,
                                    loc: `/${currentPath.join("/")}`,
                                    extensions: sortedFiles[target].extensions,
                                    modified: false
                                })
                                opened_file = openFiles.length - 1;
                            }
                            fileList = fileList;
                            openFiles = openFiles;
                        }
                        tempSelect = [];
                    }else{
                        if(searchedData[target].isFolder){
                            currentPath = searchedData[target].loc.split("/");
                            currentPath = currentPath;
                            searchValue = "";
                            loading = getFiles();
                        }else{
                            const file = `${searchedData[target].loc}/${searchedData[target].name}`;
                            if(openFiles.includes(file)){
                                opened_file = openFiles.indexOf(file);
                            }else{
                                openFiles.push(file);
                                fileList.push({
                                    name: sortedFiles[target].name,
                                    loc: `/${currentPath.join("/")}`,
                                    extensions: sortedFiles[target].extensions,
                                    modified: false
                                })
                                opened_file = openFiles.length - 1;
                            }
                            fileList = fileList;
                            openFiles = openFiles;
                        }
                        searchValue = "";
                        tempSelect = [];
                    }
                }
            }else{
                lastClicked = Date.now();
            }
        }else{
            if(shiftSelect){
                const firstOne = tempSelect[0];
                const selectOne = target;

                if(firstOne - selectOne > 0){
                    tempSelect = [tempSelect[0]];
                    for(let i = selectOne; i < firstOne; i++){
                        tempSelect.push(i);
                    }
                }else if(firstOne - selectOne < 0){
                    tempSelect = [tempSelect[0]];
                    for(let i = firstOne; i <= selectOne; i++){
                        tempSelect.push(i);
                    }
                }
            }else if(multiSelect){
                tempSelect.push(target);
                tempSelect = tempSelect;
            }else{
                tempSelect = [target];
                lastClicked = Date.now();
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
        const res = await fetch(`/server/searchInAllFiles?query=${encodeURIComponent(searchValue)}`);
        const data = await res.json() as searchData[];

        searchedData = data;
    }

    let dragDiv:HTMLDivElement;
    let isdragging:boolean = false;
    let dragStartX:number;
    let dragStartY:number;

    async function dragStart(e:MouseEvent) {
        const target = e.target as HTMLElement;
        if(target.id == "drag" || target.id == "itemCon"){
            tempSelect = [];
            showContext = false;

            dragStartX = e.pageX;
            dragStartY = e.pageY;

            dragDiv.style.width = `0px`;
            dragDiv.style.height = `0px`;
            dragDiv.style.top = `${dragStartY}px`;
            dragDiv.style.left = `${dragStartX}px`;
            isdragging = true;
        }
    }

    async function dragging(e:MouseEvent) {
        if(!isdragging) return;

        const currentX = e.pageX;
        const currentY = e.pageY;
        const width = currentX - dragStartX;
        const height = currentY - dragStartY;

        if(width >= 0){
            dragDiv.style.width = `${width}px`;
        }else{
            dragDiv.style.left = `${currentX}px`;
            dragDiv.style.width = `${-width}px`;
        }

        if(height >= 0){
            dragDiv.style.height = `${height}px`;
        }else{
            dragDiv.style.top = `${currentY}px`;
            dragDiv.style.height = `${-height}px`;
        }

        const dragRect = dragDiv.getBoundingClientRect();
        const items = document.querySelectorAll('.item');

        tempSelect = [];

        items.forEach((item, index) => {
            const itemRect = item.getBoundingClientRect();

            if(
                dragRect.left < itemRect.right &&
                dragRect.right > itemRect.left &&
                dragRect.top < itemRect.bottom &&
                dragRect.bottom > itemRect.top
            ){
                tempSelect.push(index);
            }
        });
    }

    async function dragEnd() {
        isdragging = false;
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

    function handleItemDragStart(e: DragEvent, index: number) {
        e.dataTransfer?.setData("application/x-nas-drag", JSON.stringify({
            indices: tempSelect.includes(index) ? tempSelect : [index]
        }));

        e.dataTransfer!.effectAllowed = "move";
    }

    function handleFolderDrop(e: DragEvent, folderIndex: number) {
        const data = e.dataTransfer?.getData("application/x-nas-drag");
        if(!data) return;

        const { indices } = JSON.parse(data);
        const targetFolder = sortedFiles[folderIndex].name;

        moveFilesToFolder(indices, targetFolder);
    }

    async function moveFilesToFolder(indices: number[], targetFolder: string) {
        const loginCheck = await checkLogin("COPY");
        if(!loginCheck) return;

        const promises = [];
        let targetLoc = currentPath.join("/");

        if(targetFolder === "..BACK_PARENT"){
            targetLoc = currentPath.slice(0, -1).join("/");
            targetFolder = "";
        }else{
            targetLoc = currentPath.join("/") + "/" + targetFolder;
        }

        for(const i of indices){
            promises.push(
                fetch(`/server/move?originLoc=/${encodeURIComponent(currentPath.join("/"))}&fileName=${encodeURIComponent(sortedFiles[i].name)}&targetLoc=/${encodeURIComponent(targetLoc)}&token=${$useAuth.token}`)
            );
        }

        await Promise.all(promises);
        loading = getFiles();
    }

    function handleFolderDragOver(e: DragEvent, folderIndex: number) {
        e.preventDefault();
        dragOverIndex = folderIndex;
    }

    function handleFolderDragLeave(e: DragEvent, folderIndex: number) {
        if(dragOverIndex === folderIndex) dragOverIndex = null;
    }

    function handleBackDragOver(e: DragEvent) {
        e.preventDefault();
        dragOverBack = true;
    }

    function handleBackDragLeave(e: DragEvent) {
        dragOverBack = false;
    }

    function handleBackDrop(e: DragEvent) {
        dragOverBack = false;
        const data = e.dataTransfer?.getData("application/x-nas-drag");
        if(!data) return;

        const { indices } = JSON.parse(data);

        if(currentPath.length > 0){
            moveFilesToFolder(indices, "..BACK_PARENT");
        }
    }

    loading = getFiles();
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<main id="main">
    <main on:mousedown={dragStart} on:mouseup={dragEnd} on:mousemove={dragging} on:contextmenu={context} id="drag" bind:this={itemCon}>
        <div class="title" id="marginTop">📂 File Explorer</div>
        <div class="subtitle">Current Location : /{currentPath.join("/")}</div>
        <div id="itemCon">
            {#await loading}
            <div class="subtitle">Fetching Data...</div>
            {:then value}
                <div id="functionsCon">
                    <div id="uploadbtn" on:click={upload}>Upload File</div>
                    <div id="uploadbtn" on:click={makedir}>New Dir</div>
                    <div id="searchCon">
                        <img src="/svg/search_black.svg" alt="">
                        <input type="search" placeholder="Search in current dir" id="searchInput" bind:value={sortValue} on:input={sort} autocomplete="off">
                    </div>
                </div>
                <div id="additional_input">
                    <div id="searchCon">
                        <img src="/svg/search_black.svg" alt="">
                        <input type="search" placeholder="Search In All Files" id="searchInput" bind:value={searchValue} on:input={search} autocomplete="off">
                    </div>
                </div>
                {#if status !== "clear"}
                    {#if status == "need_login"}
                        <div class="subtitle">You need to login to use this service.</div>
                    {:else if status == "need_reLogin"}
                        <div class="subtitle">There is a problem with your token.<br>Please log out and re-log in.</div>
                    {:else if status == "no_intents"}
                        <div class="subtitle">You don't have permission to view files</div>
                    {/if}
                {/if}
                {#if searchValue == ""}
                    {#if currentPath.length !== 0}
                        <div
                            class="back {tempSelect.length == 1 && tempSelect[0] == -1 ? 'tempSelect' : ''} {dragOverBack ? 'drag-over' : ''}"
                            on:click={() => {click(-1)}}
                            on:dragover={handleBackDragOver}
                            on:dragleave={handleBackDragLeave}
                            on:drop={handleBackDrop}
                        >
                            <img src='/png/back.png' alt="" class="item_icon">
                            <div class="item_name">..</div>
                        </div>
                    {/if}
                    {#if preparingUpload}
                        <div class="subtitle">Preparing Upload... Please wait...</div>
                    {/if}
                    {#each uploadingFiles as file}
                        <div class="uploading {file.action == "zipping" ? "action_zipping" : file.action == "uploading" ? "action_uploading" : file.action == "unzipping" ? "action_unzipping" : "action_error"}" style="--progress: {file.percent}%">
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
                        <div
                            class="item {data.isFolder ? 'folder' : 'file'} {tempSelect.includes(index) ? 'tempSelect' : ''} {dragOverIndex === index ? 'drag-over' : ''}"
                            draggable="true"
                            on:dragstart={(e) => handleItemDragStart(e, index)}
                            on:dragover|preventDefault={data.isFolder ? (e) => handleFolderDragOver(e, index) : null}
                            on:dragleave={data.isFolder ? (e) => handleFolderDragLeave(e, index) : null}
                            on:drop={data.isFolder ? (e) => { handleFolderDrop(e, index); dragOverIndex = null; } : null}
                            on:click={() => {click(index)}}
                            data-index={index}
                        >
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
        <div id="dragCon" bind:this={dragDiv} class={isdragging ? "" : "none"}></div>
        <div id="contextCon" class={showContext ? "" : "none"} bind:this={contextCon}>
            {#if contextType == "file"}
                <div class="menu topRadius" on:click={deleteFile}>
                    <img src="/svg/delete.svg" alt="" class="context_icon">
                    <dir class="context_text">
                        Delete
                    </dir>
                </div>
                <div class="menu" on:click={downloadFile}>
                    <img src="/svg/download.svg" alt="" class="context_icon">
                    <dir class="context_text">
                        Download
                    </dir>
                </div>
                <div class="menu" on:click={copyDonwloadLink}>
                    <img src="/svg/copy_download_link.svg" alt="" class="context_icon">
                    <dir class="context_text">
                        Copy Download Link
                    </dir>
                </div>
                {#if tempSelect.length == 1}
                    <div class="menu" on:click={renameFile}>
                        <img src="/svg/rename.svg" alt="" class="context_icon">
                        <dir class="context_text">
                            Rename
                        </dir>
                    </div>
                    {#if sortedFiles[tempSelect[0]]?.isFolder}
                        <div class="menu" on:click={zipFiles}>
                        <img src="/svg/zip.svg" alt="" class="context_icon">
                        <dir class="context_text">
                            Zip Folder
                        </dir>
                    </div>
                    {/if}
                {/if}
                {#if tempSelect.length > 1 }
                    <div class="menu" on:click={zipFiles}>
                        <img src="/svg/zip.svg" alt="" class="context_icon">
                        <dir class="context_text">
                            Zip Files
                        </dir>
                    </div>
                {/if}
                {#if tempSelect.length == 1 && sortedFiles[tempSelect[0]]?.extensions == "zip"}
                    <div class="menu" on:click={unzip}>
                        <img src="/svg/unzip.svg" alt="" class="context_icon">
                        <dir class="context_text">
                            Unzip File
                        </dir>
                    </div>
                {/if}
                <div class="menu" on:click={copyFiles}>
                    <img src="/svg/copy.svg" alt="" class="context_icon">
                    <dir class="context_text">
                        Copy File(s)
                    </dir>
                </div>
            {:else}
                <div class="menu topRadius" on:click={makedir}>
                    <img src="/svg/folder_create.svg" alt="" class="context_icon">
                    <dir class="context_text">
                        New Dir
                    </dir>
                </div>
                <div class="menu bottomRadius" on:click={pasteFiles}>
                    <img src="/svg/paste.svg" alt="" class="context_icon">
                    <dir class="context_text">
                        Paste
                    </dir>
                </div>
            {/if}
            {#if contextType == "file"}
                {#if tempSelect.length == 1}
                    <div class="menu bottomRadius" on:click={getfileInformation}>
                        <img src="/svg/wrench.svg" alt="" class="context_icon">
                        <dir class="context_text">
                            Information
                        </dir>
                    </div>
                {/if}
            {/if}
        </div>
    </main>
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
    img {
        -webkit-user-select: none;
        -khtml-user-select: none;
        -moz-user-select: none;
        -o-user-select: none;
        user-select: none;
        -webkit-user-drag: none;
        -khtml-user-drag: none;
        -moz-user-drag: none;
        -o-user-drag: none;
    }

    #main{
        height: 100vh;
        user-select: none;
    }

    #drag{
        border-right: 2px solid #121212;
        border-left: 2px solid #121212;
        height: 100vh;
        overflow-y: scroll;
        padding-left: 38px;
        padding-right: 38px;
        transition: all 0.2s ease;
    }

    #drag::-webkit-scrollbar {
        width: 5px;
        position: absolute;
        top: 0;
        right: 0;
    }

    #drag::-webkit-scrollbar-track {
        background-color: #121212;
        border-radius: 5px;
    }

    #drag::-webkit-scrollbar-thumb { 
        background-color: #565656;
        border-radius: 5px;
    }

    #drag::-webkit-scrollbar-button {
        display: none;
    }

    #marginTop{
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
        margin-bottom: 300px;
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

    .back:hover {
        cursor: default;
        background-color: #383838;
    }

    .uploading {
        position: relative;
        overflow: hidden;
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        align-items: center;
        border: 1px solid #121212;
        color: white;
        padding: 10px;
    }

    .uploading:hover {
        cursor: default;
    }

    .uploading::before {
        content: "";
        position: absolute;
        top: 0;
        left: 0;
        width: var(--progress, 0%);
        height: 100%;
        background-color: rgba(255, 255, 255, 0.2);
        z-index: 1;
        transition: width 0.3s ease;
    }

    .uploading > * {
        position: relative;
        z-index: 2;
    }

    .action_zipping {
        cursor: default;
        background-color: #7869ff;
    }

    .action_uploading {
        background-color: #4e86ff;
    }

    .action_uploading:hover {
        cursor: default;
        background-color: #6999ff;
    }

    .action_unzipping {
        background-color: #2ab400;
    }

    .action_unzipping:hover {
        cursor: default;
        background-color: #32d401;
    }

    .action_error {
        background-color: #ff1818;
    }

    .action_error:hover {
        cursor: default;
        background-color: #ff3f3f;
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

    .item:hover {
        cursor: default;
        background-color: #383838;
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

    #dragCon{
        position: absolute;
        background-color: rgba($color: #4e86ff, $alpha: 0.5);
        border: 2px solid #4e86ff;
        border-radius: 5px;
    }

    #contextCon{
        position: absolute;
        background-color: #383838;
        box-shadow: 4px 4px 4px #000000;
        border-radius: 10px;
    }

    .menu{
        display: flex;
        flex-direction: row;
        gap: 10px;
        color: white;
        font-weight: bolder;
        font-size: medium;
        padding: 10px;
    }

    .menu:hover{
        cursor: pointer;
        background-color: rgb(92, 92, 92);
    }

    .context_icon{
        width: 20px;
        height: 20px;
    }

    .context_text{
        padding: 0;
        margin: 0;
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

    .back.drag-over {
        border: 1px dotted #FFFFFF;
    }

    .item.drag-over {
        border: 1px dotted #FFFFFF;
    }
</style>