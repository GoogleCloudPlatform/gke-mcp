#!/usr/bin/env bash
set -e

LOCAL_BIN="$HOME/.local/bin"
TMP_DIR="$HOME/.local/tmp/gke-mcp-installer"

mkdir -p "$LOCAL_BIN"
mkdir -p "$TMP_DIR"

if [ -f "$LOCAL_BIN/gke-mcp" ]; then
  echo "[gke-agent] gke-mcp is already installed at $LOCAL_BIN/gke-mcp"
  rm -rf "$TMP_DIR"
  exit 0
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "Error: 'curl' is required but not installed." >&2
  exit 1
fi

echo "[gke-agent] Downloading and installing gke-mcp..."

# Fetch the installation script, patch it in memory, and execute it directly via pipe
# We wrap the pipe execution in a subshell that cd's into TMP_DIR first.
if ! curl -fsSL https://raw.githubusercontent.com/GoogleCloudPlatform/gke-mcp/main/install.sh 2>/dev/null | \
     sed "s|/usr/local/bin|$LOCAL_BIN|g" | \
     sed 's/|| sudo install .*//g' | \
     sed 's/curl -fSL/curl -s -fSL/g' | \
     (cd "$TMP_DIR" && bash); then
  echo "Error: Execution of gke-mcp install script failed." >&2
  rm -rf "$TMP_DIR"
  exit 1
fi

# Cleanup
rm -rf "$TMP_DIR"

echo "✅ gke-mcp binary installation complete."
