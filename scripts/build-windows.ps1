param(
    [string]$Output = "dist/TS3Portable.exe"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$payload = Join-Path $repoRoot "payload/ts3-client-win64.zip"
$embedDir = Join-Path $repoRoot "internal/payload/assets"
$embeddedPayload = Join-Path $embedDir "ts3-client-win64.zip"
$icon = Join-Path $repoRoot "assets/windows/app.ico"
$manifest = Join-Path $repoRoot "assets/windows/app.manifest"
$versionFile = Join-Path $repoRoot "VERSION"
$resourceJson = Join-Path $repoRoot "cmd/ts3-portable-launcher/resource.json"
$gopath = & go env GOPATH
$goversioninfo = Join-Path $gopath "bin/goversioninfo"

if (-not (Test-Path $payload)) {
    throw "Missing payload archive: $payload"
}
if (-not (Test-Path $icon)) {
    throw "Missing icon file: $icon"
}
if (-not (Test-Path $manifest)) {
    throw "Missing manifest file: $manifest"
}
if (-not (Test-Path $goversioninfo)) {
    throw "Missing goversioninfo tool: $goversioninfo"
}

$outputPath = Join-Path $repoRoot $Output
$outputDir = Split-Path -Parent $outputPath
$version = (Get-Content $versionFile -Raw).Trim()
$versionParts = $version.Split(".")
$major = if ($versionParts.Length -gt 0) { $versionParts[0] } else { "0" }
$minor = if ($versionParts.Length -gt 1) { $versionParts[1] } else { "0" }
$patch = if ($versionParts.Length -gt 2) { $versionParts[2] } else { "0" }

New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
New-Item -ItemType Directory -Force -Path $embedDir | Out-Null
Copy-Item -Force $payload $embeddedPayload

@"
{
  "FixedFileInfo": {
    "FileVersion": {
      "Major": $major,
      "Minor": $minor,
      "Patch": $patch,
      "Build": 0
    },
    "ProductVersion": {
      "Major": $major,
      "Minor": $minor,
      "Patch": $patch,
      "Build": 0
    },
    "FileFlagsMask": "3f",
    "FileFlags": "00",
    "FileOS": "040004",
    "FileType": "01",
    "FileSubType": "00"
  },
  "StringFileInfo": {
    "CompanyName": "teamSpeaker",
    "FileDescription": "Portable TeamSpeak 3 launcher",
    "FileVersion": "$version",
    "InternalName": "TS3Portable",
    "LegalCopyright": "Personal use packaging helper",
    "OriginalFilename": "TS3Portable.exe",
    "ProductName": "TS3 Portable Launcher",
    "ProductVersion": "$version"
  },
  "IconPath": "$icon",
  "ManifestPath": "$manifest"
}
"@ | Set-Content -Path $resourceJson -Encoding ascii

Push-Location (Join-Path $repoRoot "cmd/ts3-portable-launcher")
& $goversioninfo -64 -o resource.syso resource.json
Pop-Location

$env:GOOS = "windows"
$env:GOARCH = "amd64"

go build `
    -tags bundle `
    -ldflags "-H=windowsgui -s -w" `
    -o $outputPath `
    ./cmd/ts3-portable-launcher
