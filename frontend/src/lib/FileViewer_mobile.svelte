<script lang="ts">
    import * as monaco from "monaco-editor";
    import loader from "@monaco-editor/loader";
    import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
    import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
    import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
    import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker';
    import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';
    import { onMount } from "svelte";
    
    interface Ifile{
        name: string;
        loc: string;
        extensions: string;
    }
    
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
    let file:Ifile;
    let extensions:string;
    let loading = new Promise(res => {});

    async function loadEditor() {
        const res = await fetch(`/server/getTextFile?loc=/${encodeURIComponent(file.loc)}&name=${encodeURIComponent(file.name)}`);
        const data = await res.json();
        const content = data.content

        if(editor){
            const currentModel = editor.getModel();

            if(currentModel){
                currentModel.dispose();
            }

            editor = Monaco.editor.create(editorContainer, {
                value: content,
                language: translate[file.extensions] || "plaintext",
                theme: "vs-dark",
                automaticLayout: true
            });

            editor.onKeyDown(e => {
                e.preventDefault();
                alert("모바일은 편집을 지원하지 않습니다.");
                return;
            })
    
            return () => {
                editor.dispose();
            };

            // const newModel = Monaco.editor.createModel(
            //     content,
            //     translate[file.extensions] || "plaintext"
            // );

            // editor.setModel(newModel);
        }else{
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
                e.preventDefault();
                alert("모바일은 편집을 지원하지 않습니다.");
                return;
            })
    
            return () => {
                editor.dispose();
            };
        }
    }

    async function loadData() {
        const fileloc = location.href.split("loc=")[1].split("&")[0];
        const filename = location.href.split("name=")[1];
        const match = filename.match(/\.([^.]+)$/);
        if(match){
            const extensions = match[1];
            file = {
                name: filename,
                loc: fileloc == "" ? "/" : fileloc,
                extensions
            }
        }

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
    }

    onMount(async() => {
        loading = loadData();
    })

    $:{
        fileType = fileType;
    }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_media_has_caption -->
<main id="container">
    {#await loading}
        <div class="title">Fetching file data from server...</div>
    {:then value} 
        {#if file}
            {#if fileType == "img"}
                <img src="/server/getImageData?loc=/{encodeURIComponent(file.loc)}&name={encodeURIComponent(file.name)}" alt="">
            {:else if fileType == "audio"}
                <audio src="/server/getAudioData?loc=/{encodeURIComponent(file.loc)}&name={encodeURIComponent(file.name)}" controls></audio>
            {:else if fileType == "video"}
                <video src="/server/getVideoData?loc=/{encodeURIComponent(file.loc)}&name={encodeURIComponent(file.name)}" controls></video>
            {:else if fileType == "text"}
                <div id="editorContainer" bind:this={editorContainer}></div>
            {:else}
                <div class="title">해당 파일은 이곳에서 불러올 수 없습니다.</div>
            {/if}
        {/if}
    {/await}
</main>

<style lang="scss">
    #container{
        width: 100%;
        height: 100vh;
        display: flex;
        justify-content: center;
        align-items: center;
        overflow-y: hidden;
    }

    img{
        width: 100%;
    }

    video{
        width: 100%;
    }

    #editorContainer{
        width: 100%;
        height: 100%;
    }

    .title{
        color: white;
        font-size: x-large;
        font-weight: bolder;
    }
</style>