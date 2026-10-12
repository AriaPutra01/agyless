#!/usr/bin/env bash
# agyless installer for macOS / Linux.
#   curl -fsSL https://raw.githubusercontent.com/AriaPutra01/agyless/main/scripts/install.sh | bash

set -euo pipefail

OWNER="AriaPutra01"
REPO="agyless"
DEST="${HOME}/.local/bin"

ok()   { printf '\033[32m✔\033[0m %s\n' "$*"; }
warn() { printf '\033[33m!\033[0m %s\n' "$*"; }
err()  { printf '\033[31m✖\033[0m %s\n' "$*" >&2; }

case "$(uname -s)" in
  Linux*)  os="linux" ;;
  Darwin*) os="darwin" ;;
  *) err "Unsupported OS: $(uname -s)"; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64)  arch="x64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) err "Unsupported architecture: $(uname -m)"; exit 1 ;;
esac

asset="agyless-${os}-${arch}"
url="https://github.com/${OWNER}/${REPO}/releases/latest/download/${asset}"

mkdir -p "$DEST"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

printf '\033[36m↓\033[0m Downloading %s from %s/%s…\n' "$asset" "$OWNER" "$REPO"
if curl -fSL --progress-bar -o "$tmp" "$url" 2>/dev/null && [ -s "$tmp" ]; then
  chmod +x "$tmp"
  install -m 0755 "$tmp" "${DEST}/agyless"
  ok "installed agyless → ${DEST}/agyless"
elif command -v go >/dev/null 2>&1; then
  warn "GitHub release asset not found. Building from source via go install..."
  go install "github.com/${OWNER}/${REPO}/cmd/agyless@latest"
  gobin="$(go env GOPATH)/bin/agyless"
  if [ -f "$gobin" ]; then
    install -m 0755 "$gobin" "${DEST}/agyless"
    ok "compiled and installed agyless → ${DEST}/agyless"
  else
    err "Compilation failed."
    exit 1
  fi
else
  err "Download failed ($asset) and 'go' is not installed."
  err "See https://github.com/${OWNER}/${REPO}/releases"
  exit 1
fi

case ":${PATH}:" in
  *":${DEST}:"*) : ;;
  *) warn "Add ${DEST} to PATH if your shell does not already include it: export PATH=\"${DEST}:\$PATH\"" ;;
esac

printf '\n\033[36m⚙\033[0m Running initial setup...\n'
"${DEST}/agyless" init
