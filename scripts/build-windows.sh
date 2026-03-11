#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PAYLOAD="$ROOT_DIR/payload/ts3-client-win64.zip"
EMBED_DIR="$ROOT_DIR/internal/payload/assets"
EMBED_PAYLOAD="$EMBED_DIR/ts3-client-win64.zip"
ICON="$ROOT_DIR/assets/windows/app.ico"
MANIFEST="$ROOT_DIR/assets/windows/app.manifest"
VERSION_FILE="$ROOT_DIR/VERSION"
RESOURCE_JSON="$ROOT_DIR/cmd/ts3-portable-launcher/resource.json"
OUT="${1:-$ROOT_DIR/dist/TS3Portable.exe}"

if [[ ! -f "$PAYLOAD" ]]; then
  echo "Missing payload archive: $PAYLOAD" >&2
  exit 1
fi

if [[ ! -f "$ICON" ]]; then
  echo "Missing icon file: $ICON" >&2
  exit 1
fi

if [[ ! -f "$MANIFEST" ]]; then
  echo "Missing manifest file: $MANIFEST" >&2
  exit 1
fi

GOVERSIONINFO_BIN="${GOVERSIONINFO_BIN:-$(go env GOPATH)/bin/goversioninfo}"
if [[ ! -x "$GOVERSIONINFO_BIN" ]]; then
  echo "goversioninfo not found: $GOVERSIONINFO_BIN" >&2
  echo "Install it with: go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest" >&2
  exit 1
fi

VERSION="$(tr -d '[:space:]' < "$VERSION_FILE")"
IFS='.' read -r major minor patch <<< "$VERSION"
major="${major:-0}"
minor="${minor:-0}"
patch="${patch:-0}"

mkdir -p "$EMBED_DIR" "$(dirname "$OUT")"
cp "$PAYLOAD" "$EMBED_PAYLOAD"

cat > "$RESOURCE_JSON" <<EOF
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
    "FileVersion": "$VERSION",
    "InternalName": "TS3Portable",
    "LegalCopyright": "Personal use packaging helper",
    "OriginalFilename": "TS3Portable.exe",
    "ProductName": "TS3 Portable Launcher",
    "ProductVersion": "$VERSION"
  },
  "IconPath": "$ICON",
  "ManifestPath": "$MANIFEST"
}
EOF

(
  cd "$ROOT_DIR/cmd/ts3-portable-launcher"
  "$GOVERSIONINFO_BIN" -64 -o resource.syso resource.json
)

GOOS=windows GOARCH=amd64 go build \
  -tags bundle \
  -ldflags "-H=windowsgui -s -w" \
  -o "$OUT" \
  ./cmd/ts3-portable-launcher
