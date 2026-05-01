#!/usr/bin/env bash
set -e

# --- Configuration ---
LOCAL_BIN="$HOME/.local/bin"
TARBALL_URL="https://github.com/dshnayder/gke-mcp/archive/refs/heads/main.tar.gz"

# --- Pre-flight Checks ---
if ! command -v openclaw >/dev/null 2>&1; then
  echo "Error: 'openclaw' CLI is required. Please install OpenClaw first." >&2
  exit 1
fi
if ! command -v curl >/dev/null 2>&1; then
  echo "Error: 'curl' is required but not installed." >&2
  exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "Error: 'jq' is required but not installed." >&2
  exit 1
fi

SKIP_MCP=0
echo "[gke-agent] Verifying Google Cloud SDK (gcloud) setup..."
if ! gcloud container clusters list >/dev/null 2>&1; then
  echo "Warning: 'gcloud container clusters list' failed." >&2
  echo "         Please ensure 'gcloud' is installed and you are authenticated to a GCP project." >&2
  echo "         Skipping gke-mcp binary installation and MCP server registration." >&2
  SKIP_MCP=1
fi

# TODO: Enable MCP installation in a future update.
SKIP_MCP=1
echo ""
echo "========================================================================"
echo " NOTICE: Live GKE MCP server integration is an upcoming feature!"
echo "         Agents and skills are being installed now."
echo "         Real-time cluster operations via the gke-mcp server will be"
echo "         enabled in a future release. In the meantime, agents can still"
echo "         perform operations using standard gcloud and kubectl commands."
echo "========================================================================"
echo ""

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
REPO_TARBALL="$TMP_DIR/repo.tar.gz"
REPO_NAME="gke-mcp-main"

# Download repo tarball
echo "[gke-agent] Downloading repository assets..."
if ! curl -sSL "$TARBALL_URL" -o "$REPO_TARBALL"; then
  echo "Error: Failed to download repository tarball from $TARBALL_URL." >&2
  rm -rf "$TMP_DIR"
  exit 1
fi

# Extract necessary parts
echo "[gke-agent] Extracting assets..."
mkdir -p "$TMP_DIR/agents"
mkdir -p "$TMP_DIR/skills"

# Extract install.sh from root
tar -xzf "$REPO_TARBALL" -C "$TMP_DIR" "$REPO_NAME/install.sh" --strip-components=1 2>/dev/null || true
# Extract agents
tar -xzf "$REPO_TARBALL" -C "$TMP_DIR/agents" "$REPO_NAME/openclaw/agents" --strip-components=3 2>/dev/null || true
# Extract skills
tar -xzf "$REPO_TARBALL" -C "$TMP_DIR/skills" "$REPO_NAME/skills" --strip-components=2 2>/dev/null || true

# --- Phase 1: Install gke-mcp Binary ---
if [ "$SKIP_MCP" -eq 0 ]; then
  echo "--- Phase 1: Installing gke-mcp ---"

  mkdir -p "$LOCAL_BIN"

  if [ -f "$LOCAL_BIN/gke-mcp" ]; then
    echo "[gke-agent] gke-mcp is already installed at $LOCAL_BIN/gke-mcp"
  else
    echo "[gke-agent] Installing gke-mcp binary..."
    if [ -f "$TMP_DIR/install.sh" ]; then
      # Patch and execute the local install.sh
      cat "$TMP_DIR/install.sh" | \
           sed "s|/usr/local/bin|$LOCAL_BIN|g" | \
           sed 's/|| sudo install .*//g' | \
           sed 's/curl -fSL/curl -s -fSL/g' | \
           (cd "$TMP_DIR" && bash) || {
        echo "Error: Execution of gke-mcp install script failed." >&2
        rm -rf "$TMP_DIR"
        exit 1
      }
      echo "✅ gke-mcp binary installation complete."
    else
       echo "Error: Failed to extract root install.sh from tarball." >&2
       rm -rf "$TMP_DIR"
       exit 1
    fi
  fi
else
  echo "--- Phase 1: Skipped (Upcoming Feature) ---"
  echo "    The 'gke-mcp' binary provides the core capability for agents to read"
  echo "    cluster states, inspect resources, and view logs. This functionality"
  echo "    will be unavailable in this release."
fi

# --- Phase 2: Register Agents in OpenClaw ---
echo "--- Phase 2: Registering OpenClaw Agents ---"

AGENTS=()

# Discover agents
if [ -d "$TMP_DIR/agents" ]; then
  for AGENT_DIR in "$TMP_DIR/agents"/*; do
    [ -e "$AGENT_DIR" ] || continue
    if [ -d "$AGENT_DIR" ]; then
      AGENT_NAME=$(basename "$AGENT_DIR")
      AGENTS+=("$AGENT_NAME")
      WORKSPACE_DIR="$HOME/.openclaw/workspace/agents/$AGENT_NAME"
      
      echo "Processing agent: $AGENT_NAME"

      if openclaw agents list | grep -q "^- $AGENT_NAME$"; then
        echo "[gke-agent] Agent '$AGENT_NAME' is already registered in OpenClaw."
      else
        echo "[gke-agent] Adding agent '$AGENT_NAME' to OpenClaw..."
        if ! openclaw agents add "$AGENT_NAME" --workspace "$WORKSPACE_DIR" --non-interactive; then
           echo "Error: Failed to add agent using OpenClaw CLI." >&2
           exit 1
        fi
      fi

      echo "[gke-agent] Copying agent assets to workspace ($WORKSPACE_DIR)..."
      mkdir -p "$WORKSPACE_DIR"
      
      # Copy all agent files from temp extraction
      cp -a "$AGENT_DIR/." "$WORKSPACE_DIR/"

      # Fetch skills listed in skills.list
      SKILLS_LIST_FILE="$WORKSPACE_DIR/skills.list"
      if [ -f "$SKILLS_LIST_FILE" ]; then
        mkdir -p "$WORKSPACE_DIR/skills"
        while IFS= read -r SKILL || [ -n "$SKILL" ]; do
          [[ -z "$SKILL" || "$SKILL" == \#* ]] && continue
          
          if [ -d "$TMP_DIR/skills/$SKILL" ]; then
            echo "  -> Copying skill $SKILL..."
            cp -a "$TMP_DIR/skills/$SKILL" "$WORKSPACE_DIR/skills/"
          else
            echo "Warning: Skill $SKILL not found in repository tarball."
          fi
        done < "$SKILLS_LIST_FILE"
        rm -f "$SKILLS_LIST_FILE"
      else
        echo "Info: No skills.list found for $AGENT_NAME"
      fi
      
      # Identity setup assumes files are present in the workspace
      if [ -f "$WORKSPACE_DIR/IDENTITY.md" ]; then
        echo "[gke-agent] Applying identity from IDENTITY.md for $AGENT_NAME..."
        if ! openclaw agents set-identity --agent "$AGENT_NAME" --workspace "$WORKSPACE_DIR" --from-identity; then
           echo "Warning: Failed to set identity for $AGENT_NAME." >&2
        fi
      fi
    fi
  done
else
  echo "Warning: No agents directory found in tarball."
fi

# --- Phase 3: Register MCP Server ---
if [ "$SKIP_MCP" -eq 0 ]; then
  echo "--- Phase 3: Registering MCP Server (gke-mcp) ---"
  if openclaw mcp list | grep -q "^- gke-mcp$"; then
    echo "[gke-agent] MCP server 'gke-mcp' is already registered."
  else
    echo "[gke-agent] Adding MCP server 'gke-mcp'..."
    # Use JSON string for the server configuration
    MCP_CONFIG="{\"command\":\"$LOCAL_BIN/gke-mcp\",\"args\":[],\"env\":{}}"
    if ! openclaw mcp set gke-mcp "$MCP_CONFIG"; then
      echo "Error: Failed to register MCP server." >&2
    fi
  fi
else
  echo "--- Phase 3: Skipped (Upcoming Feature) ---"
  echo "    The OpenClaw gateway will not bridge the GKE MCP tools to your agents yet."
fi

# --- Phase 4: Configure Semantic Routing ---
echo "--- Phase 4: Configuring Semantic Routing ---"
if [ ${#AGENTS[@]} -gt 1 ]; then
  # Get the current allowAgents array (defaulting to empty array if not set)
  CURRENT_ALLOW_AGENTS=$(openclaw config get agents.defaults.subagents.allowAgents 2>/dev/null || echo "[]")

  # Use jq to add all agents to the array
  AGENTS_JSON_ARRAY=$(printf '%s\n' "${AGENTS[@]}" | jq -R . | jq -s -c .)
  NEW_ALLOW_AGENTS=$(echo "$CURRENT_ALLOW_AGENTS" | jq -c ". + $AGENTS_JSON_ARRAY | unique")

  # Patch the configuration with the updated array
  echo "{\"agents\":{\"defaults\":{\"subagents\":{\"allowAgents\":$NEW_ALLOW_AGENTS}}}}" | openclaw config patch --stdin
else
  echo "No agents to configure for semantic routing."
fi

# Cleanup
rm -rf "$TMP_DIR"

echo "--- Installation Complete ---"
if [ ${#AGENTS[@]} -gt 0 ]; then
  echo "You can now start the gateway and interact with your new GKE agents."
fi
