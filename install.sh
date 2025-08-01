#!/usr/bin/env bash
# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

# Check for required commands
for cmd in curl tar; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Error: $cmd is not installed. Please install $cmd to proceed."
    exit 1
  fi
done

REPO="GoogleCloudPlatform/gke-mcp"
BINARY="gke-mcp"

# Detect OS
sysOS="$(uname | tr '[:upper:]' '[:lower:]')"
case "$sysOS" in
linux) OS="Linux" ;;
darwin) OS="Darwin" ;;
*)
  echo "If you are on an unsupported OS, please follow the manual installation instructions at:"
  echo "https://github.com/${REPO}#installation"
  exit 1
  ;;
esac

# Detect ARCH
ARCH="$(uname -m)"
case "${ARCH}" in
x86_64 | amd64) ARCH="x86_64" ;;
arm64 | aarch64) ARCH="arm64" ;;
*)
  echo "If you are on an unsupported architecture, please follow the manual installation instructions at:"
  echo "https://github.com/${REPO}#installation"
  exit 1
  ;;
esac

# Get latest version tag from GitHub API, Use GITHUB_TOKEN if available to avoid potential rate limit
if [ -n "${GITHUB_TOKEN:-}" ]; then
  auth_hdr="Authorization: token $GITHUB_TOKEN"
else
  auth_hdr=""
fi
LATEST_TAG=$(curl -s -H "$auth_hdr" \
  "https://api.github.com/repos/${REPO}/releases/latest" |
  sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')
if [ -z "${LATEST_TAG}" ]; then
  echo "Failed to fetch latest release tag."
  exit 1
fi

# Compose download URL
TARBALL="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${TARBALL}"

# Download and extract
echo "Downloading ${URL} ..."
curl -fSL --retry 3 "${URL}" -o "${TARBALL}"
tar --no-same-owner -xzf "${TARBALL}"

# Move binary to /usr/local/bin (may require sudo)
echo "Installing ${BINARY} to /usr/local/bin (may require sudo)..."
install -m 0755 "${BINARY}" /usr/local/bin/ || sudo install -m 0755 "${BINARY}" /usr/local/bin/

# Clean up
rm "${TARBALL}"

echo "✅ ${BINARY} installed successfully! Run '${BINARY} --help' to get started."
