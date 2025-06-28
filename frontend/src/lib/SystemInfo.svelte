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
        <table>
            <tbody>
                <tr>
                    <td class="content_title">
                        <div class="text">CPU USAGE</div>
                        <div class="text">:</div>
                    </td>
                    <td class="content_text">{systemData.cpu}</td>
                </tr>
                <tr>
                    <td class="content_title">
                        <div class="text">MEMORY USAGE</div>
                        <div class="text">:</div>
                    </td>
                    <td class="content_text">{systemData.memory}</td>
                </tr>
                <tr>
                    <td class="content_title">
                        <div class="text">SYSTEM UPTIME</div>
                        <div class="text">:</div>
                    </td>
                    <td class="content_text">{systemData.uptime}</td>
                </tr>
                <tr>
                    <td class="content_title">
                        <div class="text">DISK USAGE</div>
                        <div class="text">:</div>
                    </td>
                    <td class="content_text">{systemData.disk.usagePercentage} ({systemData.disk.used}B / {systemData.disk.total}B)</td>
                </tr>
            </tbody>
        </table>
    {/await}
</main>

<style lang="scss">
    main{
        // height: 100vh;
        overflow-y: auto;
        padding: 40px;
        // user-select: none;
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

    table{
        margin-bottom: 20px;
    }

    .content_title{
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        gap: 10px;
        color: white;
        font-weight: bolder;
        font-size: larger;
    }

    .content_text{
        padding-left: 10px;
        color: white;
        font-weight: bolder;
        font-size: large;
    }
</style>