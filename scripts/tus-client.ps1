#!/usr/bin/env pwsh
# Minimal tus 1.0.0 client for stress testing the Go backend.
# Usage:
#   ./tus-client.ps1 -SourcePath <file> -Filename <dest-name> -Loc <subdir> -Token <jwt>
#                    -BaseUrl <url> -ChunkSize <bytes> [-StopAfterBytes <n>] [-ResumeFrom <offset> -UploadUrl <url>]
[CmdletBinding()]
param(
    [Parameter(Mandatory)] [string] $SourcePath,
    [Parameter(Mandatory)] [string] $Filename,
    [Parameter(Mandatory)] [string] $Token,
    [Parameter(Mandatory)] [string] $BaseUrl,
    [string] $Loc = "",
    [int] $ChunkSize = 4MB,
    [long] $StopAfterBytes = -1,
    [long] $ResumeFrom = -1,
    [string] $UploadUrl = ""
)

$ErrorActionPreference = "Stop"

function Encode-Base64 {
    param([string] $value)
    return [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($value))
}

function New-Upload {
    param([long] $totalSize)
    $metadata = "filename $(Encode-Base64 $Filename),loc $(Encode-Base64 $Loc)"
    $headers = @{
        "Tus-Resumable"   = "1.0.0"
        "Upload-Length"   = $totalSize.ToString()
        "Upload-Metadata" = $metadata
        "Authorization"   = "Bearer $Token"
    }
    $response = Invoke-WebRequest -Uri "$BaseUrl/files/" -Method POST -Headers $headers -UseBasicParsing
    if ($response.StatusCode -ne 201) {
        throw "Create failed: $($response.StatusCode)"
    }
    $location = $response.Headers.Location
    if ($location -is [Array]) { $location = $location[0] }
    return $location
}

function Get-HeaderLong {
    param($response, [string] $name)
    $value = $response.Headers[$name]
    if ($value -is [Array]) { $value = $value[0] }
    return [long] $value
}

function Get-Offset {
    param([string] $url)
    $headers = @{"Tus-Resumable" = "1.0.0"; "Authorization" = "Bearer $Token"}
    $response = Invoke-WebRequest -Uri $url -Method HEAD -Headers $headers -UseBasicParsing
    return Get-HeaderLong -response $response -name "Upload-Offset"
}

function Send-Chunk {
    param([string] $url, [long] $offset, [byte[]] $chunk)
    $headers = @{
        "Tus-Resumable"  = "1.0.0"
        "Content-Type"   = "application/offset+octet-stream"
        "Upload-Offset"  = $offset.ToString()
        "Authorization"  = "Bearer $Token"
    }
    $response = Invoke-WebRequest -Uri $url -Method PATCH -Headers $headers -Body $chunk -UseBasicParsing
    if ($response.StatusCode -ne 204) {
        throw "PATCH failed at offset $offset"
    }
    return Get-HeaderLong -response $response -name "Upload-Offset"
}

$totalSize = (Get-Item $SourcePath).Length
Write-Output "Source: $SourcePath ($totalSize bytes)"

if ($UploadUrl -ne "") {
    Write-Output "Resuming upload at $UploadUrl"
    $serverOffset = Get-Offset $UploadUrl
    Write-Output "Server offset: $serverOffset"
    $url = $UploadUrl
    $offset = $serverOffset
} else {
    $url = New-Upload -totalSize $totalSize
    Write-Output "Created upload: $url"
    $offset = 0
}

$stream = [IO.File]::OpenRead($SourcePath)
try {
    $stream.Seek($offset, "Begin") | Out-Null
    $buffer = New-Object byte[] $ChunkSize
    $start = Get-Date
    while ($offset -lt $totalSize) {
        $remaining = $totalSize - $offset
        $toRead = [int][Math]::Min([long]$ChunkSize, $remaining)
        $read = $stream.Read($buffer, 0, $toRead)
        if ($read -eq 0) { break }
        $chunk = $buffer[0..($read - 1)]

        if ($StopAfterBytes -gt 0 -and $offset + $read -gt $StopAfterBytes) {
            Write-Output "STOP signal at offset $($offset + $read)"
            break
        }

        $offset = Send-Chunk -url $url -offset $offset -chunk $chunk
        $elapsed = (Get-Date) - $start
        $mbPerSec = if ($elapsed.TotalSeconds -gt 0) { ($offset / 1MB) / $elapsed.TotalSeconds } else { 0 }
        Write-Progress -Activity "Uploading $Filename" -Status "$([math]::Round($offset / 1MB, 1)) / $([math]::Round($totalSize / 1MB, 1)) MB at $([math]::Round($mbPerSec, 1)) MB/s" -PercentComplete (($offset * 100) / $totalSize)
    }
}
finally {
    $stream.Close()
}

$elapsed = (Get-Date) - $start
Write-Output "Final offset: $offset / $totalSize"
Write-Output "Elapsed: $([math]::Round($elapsed.TotalSeconds, 2))s"
Write-Output "Throughput: $([math]::Round(($offset / 1MB) / $elapsed.TotalSeconds, 2)) MB/s"
Write-Output "UPLOAD_URL=$url"
