import os from "os";
import { execSync, exec } from 'child_process';
import fs from "fs/promises";
import path from "path";
import { diskPath } from "../config/config";

interface DiskUsage{
    total: string;
    used: string;
    available: string;
    usagePercentage: string;
};

interface FileInfo {
    name: string;
    isFolder: boolean;
    extensions: string;
    loc: string;
}

export function formatSize(size: number): string {
    if (size >= 1_073_741_824) return (size / 1_073_741_824).toFixed(2) + " GB";
    if (size >= 1_048_576) return (size / 1_048_576).toFixed(2) + " MB";
    if (size >= 1024) return (size / 1024).toFixed(2) + " KB";
    return size + " B";
}

export function formatDate(date: Date): string {
    return date.toISOString().split("T")[0];
}

export function getCpuUsage(): string {
    const cpus = os.cpus();
    let totalIdle = 0, totalTick = 0;

    cpus.forEach(cpu => {
        for (const type in cpu.times) {
        totalTick += cpu.times[type as keyof typeof cpu.times];
        }
        totalIdle += cpu.times.idle;
    });

    const idle = totalIdle / cpus.length;
    const total = totalTick / cpus.length;
    const usage = 100 - Math.floor((idle / total) * 100);

    return `${usage}%`;
}

export function getMemoryUsage(): string {
    const totalMemory = os.totalmem();
    const freeMemory = os.freemem();
    const usedMemory = totalMemory - freeMemory;
    const usagePercentage = (usedMemory / totalMemory) * 100;
  
    const totalMemoryGB = (totalMemory / (1024 ** 3)).toFixed(2);
    const usedMemoryGB = (usedMemory / (1024 ** 3)).toFixed(2);
  
    return `${usagePercentage.toFixed(2)}% (${usedMemoryGB}GB / ${totalMemoryGB}GB)`;
}

export function getUptime(): string {
    const uptime = os.uptime();
    const days = Math.floor(uptime / 86400);
    const hours = Math.floor((uptime % 86400) / 3600);
    const minutes = Math.floor((uptime % 3600) / 60);
    return `${days}d ${hours}h ${minutes}m`;
}

export function getDiskUsage(): Promise<DiskUsage> {
    return new Promise((resolve, reject) => {
      exec(`df -h --output=size,used,avail,pcent ${diskPath}`, (error, stdout, stderr) => {
            if (error) {
                reject(`Error executing df: ${stderr || error.message}`);
                return;
            }
    
            const lines = stdout.trim().split('\n');
            if (lines.length < 2) {
                reject('Unexpected df output format');
                return;
            }
    
            const data = lines[1].trim().split(/\s+/); // Splits the second line
            if (data.length < 4) {
                reject('Failed to parse df output');
                return;
            }
    
            resolve({
                total: data[0],
                used: data[1],
                available: data[2],
                usagePercentage: data[3],
            });
        });
    });
}

export async function searchFilesInDir(dir: string, query: string, result: FileInfo[]) {
    try{
        const files = await fs.readdir(dir, { withFileTypes: true });

        for(const file of files){
            const filePath = path.join(dir, file.name);
            const isFolder = file.isDirectory();
            const extMatch = file.name.match(/\.([^.]+)$/);
            const ext = isFolder ? "" : extMatch ? extMatch[1] : "";

            if(file.name.toLowerCase().includes(query.toLowerCase())){
                result.push({
                    name: file.name,
                    isFolder,
                    extensions: ext,
                    loc: filePath.split("/nas-data")[1].startsWith("/") ? filePath.split("/nas-data/")[1] : filePath.split("/nas-data/")[0]
                });
            }

            if(isFolder){
                await searchFilesInDir(filePath, query, result);
            }
        }
    }catch(err){
        return;
    }
}