package system

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type DiskInfo struct {
	Total       string  `json:"total"`
	Used        string  `json:"used"`
	Free        string  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

type SystemInfo struct {
	CPU    float64  `json:"cpu"`
	Memory float64  `json:"memory"`
	Uptime string   `json:"uptime"`
	Disk   DiskInfo `json:"disk"`
}

func Snapshot() (*SystemInfo, error) {
	info := &SystemInfo{}

	cpuPercent, err := cpu.Percent(200*time.Millisecond, false)
	if err == nil && len(cpuPercent) > 0 {
		info.CPU = round2(cpuPercent[0])
	}

	memInfo, err := mem.VirtualMemory()
	if err == nil {
		info.Memory = round2(memInfo.UsedPercent)
	}

	uptimeSeconds, err := host.Uptime()
	if err == nil {
		info.Uptime = formatUptime(time.Duration(uptimeSeconds) * time.Second)
	}

	diskPath := "/"
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	}
	usage, err := disk.Usage(diskPath)
	if err == nil {
		info.Disk = DiskInfo{
			Total:       formatSize(usage.Total),
			Used:        formatSize(usage.Used),
			Free:        formatSize(usage.Free),
			UsedPercent: round2(usage.UsedPercent),
		}
	}
	return info, nil
}

func GetSystemInfoHandler(w http.ResponseWriter, r *http.Request) {
	info, err := Snapshot()
	if err != nil {
		http.Error(w, "system info failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func round2(value float64) float64 {
	return float64(int(value*100)) / 100
}

func formatUptime(duration time.Duration) string {
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	return formatPlural(days, "day") + " " + formatPlural(hours, "hour") + " " + formatPlural(minutes, "min")
}

func formatPlural(count int, unit string) string {
	if count == 1 {
		return "1 " + unit
	}
	return intToString(count) + " " + unit + "s"
}

func intToString(value int) string {
	if value == 0 {
		return "0"
	}
	result := ""
	if value < 0 {
		result = "-"
		value = -value
	}
	digits := ""
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	return result + digits
}

func formatSize(bytes uint64) string {
	const (
		kb uint64 = 1024
		mb        = 1024 * kb
		gb        = 1024 * mb
		tb        = 1024 * gb
	)
	switch {
	case bytes >= tb:
		return formatFloat(float64(bytes)/float64(tb)) + " TB"
	case bytes >= gb:
		return formatFloat(float64(bytes)/float64(gb)) + " GB"
	case bytes >= mb:
		return formatFloat(float64(bytes)/float64(mb)) + " MB"
	case bytes >= kb:
		return formatFloat(float64(bytes)/float64(kb)) + " KB"
	default:
		return intToString(int(bytes)) + " B"
	}
}

func formatFloat(value float64) string {
	whole := int(value)
	fraction := int((value - float64(whole)) * 100)
	return intToString(whole) + "." + pad2(fraction)
}

func pad2(value int) string {
	if value < 10 {
		return "0" + intToString(value)
	}
	return intToString(value)
}
