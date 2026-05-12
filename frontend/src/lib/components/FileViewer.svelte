<script lang="ts">
    import * as monaco from "monaco-editor";
    import loader from "@monaco-editor/loader";
    import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
    import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
    import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
    import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker';
    import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';
    import { onMount } from "svelte";
    import { useAuth } from "../store/store";
    
    interface Ifile{
        name: string;
        loc: string;
        extensions: string;
        modified: boolean;
    }

    export let fileList:Ifile[];
    export let file:Ifile;
    let fileType:"video"|"audio"|"img"|"text"|"none";
    const translate:any = {
        "c": "c",
        "cpp": "cpp",
        "cc": "cpp",
        "cxx": "cpp",
        "cs": "csharp",
        "java": "java",
        "py": "python",
        "js": "javascript",
        "ts": "typescript",
        "html": "html",
        "htm": "html",
        "css": "css",
        "xml": "xml",
        "md": "markdown",
        "json": "json",
        "sh": "shell",
        "yaml": "yaml",
        "yml": "yaml",
        "env": "properties",
        "properties": "properties",
        "sql": "sql",
        "gitignore": "git",
        "txt": "plaintext"
    }
    const imgType = ["png", "jpg", "jpeg", "svg"];
    const audioType = ["mp3", "wav"];
    const videoType = ["mp4", "mov"];
    let editorContainer: HTMLDivElement;
    let editor: monaco.editor.IStandaloneCodeEditor;
    let Monaco:any;
    let origin_content:string;

    async function addModified() {
        const value = editor.getValue();
        if(value == origin_content){
            file.modified = false;
        }else{
            file.modified = true;
        }
        fileList = fileList;
    }

    async function removeModified() {
        file.modified = false;
        fileList = fileList;
    }

    async function loadEditor() {
        const res = await fetch(`/server/getTextFile?loc=/${encodeURIComponent(file.loc)}&name=${encodeURIComponent(file.name)}&token=${$useAuth.token}`);
        const data = await res.json();
        const content = data.content
        origin_content = data.content;

        self.MonacoEnvironment = {
            getWorker: function (_moduleId: any, label: string) {
                if (label === 'json') {
                    return new jsonWorker();
                }
                if (label === 'css' || label === 'scss' || label === 'less') {
                    return new cssWorker();
                }
                if (label === 'html' || label === 'handlebars' || label === 'razor') {
                    return new htmlWorker();
                }
                if (label === 'typescript' || label === 'javascript') {
                    return new tsWorker();
                }
                return new editorWorker();
            }
        };

        Monaco = await import('monaco-editor');
        editor = Monaco.editor.create(editorContainer, {
            value: content,
            language: translate[file.extensions] || "plaintext",
            theme: "vs-dark",
            automaticLayout: true
        });

        editor.onKeyDown(e => {
            setTimeout(() => {
                addModified();    
            }, 1);
        })

        return () => {
            editor.dispose();
        };
    }

    onMount(async() => {
        if(imgType.includes(file.extensions)){
            fileType = "img";
        }else if(audioType.includes(file.extensions)){
            fileType = "audio";
        }else if(videoType.includes(file.extensions)){
            fileType = "video";
        }else if(file.extensions in translate){
            fileType = "text";
        }else{
            fileType = "none";
        }

        if (fileType === "text") {
            loadEditor();
        }

        window.addEventListener("keydown", async e => {
            if(e.ctrlKey && e.key === "s" || e.ctrlKey && e.key === "S"){
                e.preventDefault();

                if(!file.modified) return;
    
                const content = editor.getValue();
                const res = await fetch(`/server/saveTextFile?loc=/${encodeURIComponent(file.loc)}&name=${encodeURIComponent(file.name)}&token=${$useAuth.token}`, {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json"
                    },
                    body: JSON.stringify({
                        text: content
                    })
                })
                const data = await res.text();
                if(data != "complete"){
                    alert("저장 실패");
                    return;
                }

                removeModified();
            }
        })
    })
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_media_has_caption -->
<main id="container">
    {#if file}
        {#if fileType == "img"}
            <img src="/server/getImageData?loc=/{encodeURIComponent(file.loc)}&name={encodeURIComponent(file.name)}&token={$useAuth.token}" alt="">
        {:else if fileType == "audio"}
            <audio src="/server/getAudioData?loc=/{encodeURIComponent(file.loc)}&name={encodeURIComponent(file.name)}&token={$useAuth.token}" controls></audio>
        {:else if fileType == "video"}
            <video src="/server/getVideoData?loc=/{encodeURIComponent(file.loc)}&name={encodeURIComponent(file.name)}&token={$useAuth.token}" controls></video>
        {:else if fileType == "text"}
            <div id="editorContainer" bind:this={editorContainer}></div>
        {:else}
            <div class="title">해당 파일은 이곳에서 불러올 수 없습니다.</div>
        {/if}
    {/if}
</main>

<style lang="scss">
    #container{
        width: 100%;
        height: 95.9vh;
        display: flex;
        justify-content: center;
        align-items: center;
        overflow-y: hidden;
    }

    video{
        width: 30vw;
    }

    img{
        width: 30vw;
    }

    #editorContainer{
        width: 100%;
        height: 100%;
    }

    .title{
        color: white;
        font-weight: bolder;
        font-size: x-large;
    }
</style>