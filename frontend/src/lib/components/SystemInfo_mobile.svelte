<script lang="ts">
    interface systemInfo{
        cpu: string;
        memory: string;
        uptime: string;
        disk: {
            total: string;
            used: string;
            available: string;
            usagePercentage: string;
        };
    }

    let systemData:systemInfo;
    let loading = new Promise(res => {});

    async function getSystemData() {
        const res = await fetch(`/server/getSystemInfo`);
        systemData = await res.json();
        return;
    }

    loading = getSystemData();
</script>

<main>
    <div class="title">🖥️ System Info</div>
    {#await loading}
        <div class="subtitle">Fetching Data...</div>
    {:then value}
        <div id="contents">
            <div class="content">
                <div class="text">CPU USAGE</div>
                <div class="content_text">{systemData.cpu}</div>
            </div>
            <div class="content">
                <div class="text">MEMORY USAGE</div>
                <div class="content_text">{systemData.memory}</div>
            </div>
            <div class="content">
                <div class="text">SYSTEM UPTIME</div>
                <div class="content_text">{systemData.uptime}</div>
            </div>
            <div class="content">
                <div class="text">DISK USAGE</div>
                <div class="content_text">{systemData.disk.usagePercentage} ({systemData.disk.used}B / {systemData.disk.total}B)</div>
            </div>
        </div>
    {/await}
</main>

<style lang="scss">
    main{
        // height: 100vh;
        overflow-y: auto;
        padding: 30px 30px 200px 30px;
        // user-select: none;
    }

    #contents{
        display: flex;
        width: 100%;
        flex-direction: column;
        gap: 40px;
    }

    .title{
        font-size: xx-large;
        color: white;
        font-weight: bolder;
        margin-bottom: 40px;
    }

    .subtitle{
        font-size: x-large;
        color: white;
        font-weight: bolder;
    }

    .content{
        display: flex;
        flex-direction: column;
        gap: 10px;
        color: white;
        font-weight: bolder;
        font-size: x-large;
    }

    .content_text{
        text-align: left;
        color: white;
        font-weight: bolder;
        font-size: x-large;
    }

    .text{
        text-align: left;
    }
</style>