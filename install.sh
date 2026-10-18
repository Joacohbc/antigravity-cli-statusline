#!/usr/bin/env sh
# Antigravity CLI Status Line installer & updater (Linux / macOS / Windows)
# Usage:
#   Install latest:
#     curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | sh
#   Update or install specific version:
#     curl -fsSL https://raw.githubusercontent.com/Joacohbc/antigravity-cli-statusline/main/install.sh | VERSION=v1.1.0 sh
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

is_update=0
if [ -f "$binary_dest" ]; then
  is_update=1
  current_version=$("$binary_dest" version 2>/dev/null || echo "installed")
  info "Existing installation detected (${current_version})"
  info "Updating statusline to ${VERSION} (${asset})..."
else
  info "Installing statusline ${VERSION} (${asset})..."
fi

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

# Atomic replacement with safe backup
if [ "$is_update" = "1" ]; then
  cp -f "$binary_dest" "${binary_dest}.bak" 2>/dev/null || true
fi

mv -f "$tmp_path" "$binary_dest"
chmod +x "$binary_dest"
rm -f "${binary_dest}.bak" 2>/dev/null || true

# Antigravity CLI configuration
if [ -f "${SETTINGS_FILE}" ]; then
  already_configured=0
  if command -v jq >/dev/null 2>&1; then
    existing_cmd=$(jq -r '.statusLine.command // empty' "${SETTINGS_FILE}" 2>/dev/null || true)
    existing_enabled=$(jq -r '.statusLine.enabled // empty' "${SETTINGS_FILE}" 2>/dev/null || true)
    if [ "$existing_cmd" = "$binary_dest" ] && [ "$existing_enabled" = "true" ]; then
      already_configured=1
    else
      TEMP_SETTINGS=$(mktemp "${SETTINGS_FILE}.tmp.XXXXXX")
      jq --arg cmd "${binary_dest}" '.statusLine = {"type": "command", "command": $cmd, "enabled": true}' "${SETTINGS_FILE}" > "${TEMP_SETTINGS}"
      mv -f "${TEMP_SETTINGS}" "${SETTINGS_FILE}"
    fi
  fi
  if [ "$already_configured" = "1" ]; then
    info "Antigravity CLI settings already verified (${SETTINGS_FILE})"
  else
    info "Updated Antigravity CLI settings (${SETTINGS_FILE})"
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
  info "Created Antigravity CLI settings (${SETTINGS_FILE})"
fi

active_version=$("$binary_dest" version 2>/dev/null || echo "$VERSION")
if [ "$is_update" = "1" ]; then
  info "Antigravity Status Line successfully updated to ${active_version}!"
else
  info "Antigravity Status Line successfully installed: ${active_version} (${binary_dest})"
fi
