#!/usr/bin/env sh
# Antigravity CLI Status Line installer (Linux / macOS / Windows)
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | sh
#   curl -fsSL .../install.sh | VERSION=v1.0.0 sh
set -eu

REPO="${REPO:-Joacohbc/antigravity-cli-statusline}"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.gemini/antigravity-cli}"
SETTINGS_FILE="${INSTALL_DIR}/settings.json"

err() { printf 'error: %s\n' "$*" >&2; exit 1; }
info() { printf '==> %s\n' "$*"; }

uname_s=$(uname -s)
uname_m=$(uname -m)

case "$uname_s" in
  Linux*)  os="linux" ;;
  Darwin*) os="darwin" ;;
  MINGW*|MSYS*|CYGWIN*) os="windows" ;;
  *) err "unsupported OS: $uname_s" ;;
esac

case "$uname_m" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) err "unsupported architecture: $uname_m" ;;
esac

target="${os}-${arch}"
asset="statusline-${target}"
binary_dest="${INSTALL_DIR}/statusline"

if [ "$os" = "windows" ]; then
  asset="${asset}.exe"
  binary_dest="${binary_dest}.exe"
fi

if [ "$VERSION" = "latest" ]; then
  url="https://github.com/${REPO}/releases/latest/download/${asset}"
else
  url="https://github.com/${REPO}/releases/download/${VERSION}/${asset}"
fi

command -v curl >/dev/null 2>&1 || err "curl is required to download statusline"

mkdir -p "$INSTALL_DIR"
tmp_path="${binary_dest}.download"

info "Downloading $asset ($VERSION)"
curl -fSL --progress-bar -o "$tmp_path" "$url" || err "download failed: $url"

# Verify the published sha256 before trusting the binary
sum_url="${url}.sha256"
sum_path="${tmp_path}.sha256"
info "Verifying SHA256 checksum..."
if curl -fsSL -o "$sum_path" "$sum_url" 2>/dev/null; then
  expected=$(awk '{print $1}' "$sum_path" | head -n1)
  if [ -n "$expected" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
      actual=$(sha256sum "$tmp_path" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then
      actual=$(shasum -a 256 "$tmp_path" | awk '{print $1}')
    else
      actual=""
    fi

    if [ -n "$actual" ] && [ "$expected" != "$actual" ]; then
      rm -f "$tmp_path" "$sum_path"
      err "checksum mismatch: expected $expected, got $actual"
    fi
  fi
  rm -f "$sum_path"
fi

mv -f "$tmp_path" "$binary_dest"
chmod +x "$binary_dest"

info "Configuring Antigravity CLI (${SETTINGS_FILE})..."
if [ -f "${SETTINGS_FILE}" ]; then
  if command -v jq >/dev/null 2>&1; then
    TEMP_SETTINGS=$(mktemp "${SETTINGS_FILE}.tmp.XXXXXX")
    jq --arg cmd "${binary_dest}" '.statusLine = {"type": "command", "command": $cmd, "enabled": true}' "${SETTINGS_FILE}" > "${TEMP_SETTINGS}"
    mv -f "${TEMP_SETTINGS}" "${SETTINGS_FILE}"
  else
    printf 'Notice: jq not found. Ensure %s has statusLine set to %s\n' "$SETTINGS_FILE" "$binary_dest"
  fi
else
  cat <<EOF > "${SETTINGS_FILE}"
{
  "statusLine": {
    "type": "command",
    "command": "${binary_dest}",
    "enabled": true
  }
}
EOF
fi

info "Antigravity Status Line installed successfully to ${binary_dest}"
