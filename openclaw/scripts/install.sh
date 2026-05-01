#!/usr/bin/env bash
set -e

# --- Configuration ---
LOCAL_BIN="$HOME/.local/bin"
BASE_URL="https://raw.githubusercontent.com/dshnayder/gke-mcp/main"
# List of agents managed by this plugin installer
AGENTS=("gke-sre")
AGENT_FILES=("SOUL.md" "IDENTITY.md")

# --- Pre-flight Checks ---
if ! command -v openclaw >/dev/null 2>&1; then
  echo "Error: 'openclaw' CLI is required. Please install OpenClaw first." >&2
  exit 1
fi

# --- Phase 1: Install gke-mcp Binary ---
echo "--- Phase 1: Installing gke-mcp ---"

TMP_DIR="$HOME/.local/tmp/gke-mcp-installer"

mkdir -p "$LOCAL_BIN"
mkdir -p "$TMP_DIR"

if [ -f "$LOCAL_BIN/gke-mcp" ]; then
  echo "[gke-agent] gke-mcp is already installed at $LOCAL_BIN/gke-mcp"
else
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

  echo "✅ gke-mcp binary installation complete."
fi

# Cleanup
rm -rf "$TMP_DIR"

# Ensure the local bin is in PATH for the current session (informational)
if [[ ":$PATH:" != *":$LOCAL_BIN:"* ]]; then
    echo "Warning: $LOCAL_BIN is not in your PATH. You may need to add it."
fi

# --- Phase 2: Register Agents in OpenClaw ---
echo "--- Phase 2: Registering OpenClaw Agents ---"

for AGENT_NAME in "${AGENTS[@]}"; do
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

  # Populate workspace and identity (Remote fetching)
  echo "[gke-agent] Fetching agent assets to workspace ($WORKSPACE_DIR)..."
  mkdir -p "$WORKSPACE_DIR"
  mkdir -p "$WORKSPACE_DIR/skills"

  for FILE in "${AGENT_FILES[@]}"; do
    echo "  -> Downloading $FILE..."
    if ! curl -sSLf "$BASE_URL/openclaw/agents/$AGENT_NAME/$FILE" -o "$WORKSPACE_DIR/$FILE"; then
       echo "Warning: Failed to download $FILE for $AGENT_NAME"
    fi
  done

  # Fetch skills list
  SKILLS_LIST_FILE="$WORKSPACE_DIR/skills.list"
  echo "  -> Downloading skills list..."
  if curl -sSLf "$BASE_URL/openclaw/agents/$AGENT_NAME/skills.list" -o "$SKILLS_LIST_FILE"; then
    while IFS= read -r SKILL || [ -n "$SKILL" ]; do
      # Skip empty lines and comments
      [[ -z "$SKILL" || "$SKILL" == \#* ]] && continue
      
      echo "  -> Downloading skill $SKILL..."
      mkdir -p "$WORKSPACE_DIR/skills/$SKILL"
      if ! curl -sSLf "$BASE_URL/skills/$SKILL/SKILL.md" -o "$WORKSPACE_DIR/skills/$SKILL/SKILL.md"; then
         echo "Warning: Failed to download skill $SKILL for $AGENT_NAME"
      fi
    done < "$SKILLS_LIST_FILE"
    # Optional: cleanup the list file
    rm -f "$SKILLS_LIST_FILE"
  else
    echo "Warning: No skills.list found or failed to download for $AGENT_NAME"
  fi
  
  # Identity setup assumes files are present in the workspace
  if [ -f "$WORKSPACE_DIR/IDENTITY.md" ]; then
    echo "[gke-agent] Applying identity from IDENTITY.md for $AGENT_NAME..."
    if [ "$AGENT_NAME" = "gke-sre" ]; then
      if ! openclaw agents set-identity --agent "$AGENT_NAME" --workspace "$WORKSPACE_DIR" --from-identity --name "GKE Expert SRE" --theme "blue"; then
         echo "Warning: Failed to set identity." >&2
      fi
    else
      if ! openclaw agents set-identity --agent "$AGENT_NAME" --workspace "$WORKSPACE_DIR" --from-identity --theme "blue"; then
         echo "Warning: Failed to set identity." >&2
      fi
    fi
  fi
done

# --- Phase 3: Register MCP Server ---
echo "--- Phase 3: Registering MCP Server (gke-mcp) ---"
if openclaw mcp list | grep -q "^- gke-mcp$"; then
  echo "[gke-agent] MCP server 'gke-mcp' is already registered."
else
  echo "[gke-agent] Adding MCP server 'gke-mcp'..."
  # Use JSON string for the server configuration
  MCP_CONFIG="{\"command\":\"$LOCAL_BIN/gke-mcp\",\"args\":[],\"env\":{}}"
  if ! openclaw mcp set gke-mcp "$MCP_CONFIG"; then
    echo "Error: Failed to register MCP server." >&2
    exit 1
  fi
fi

# --- Phase 4: Configure Semantic Routing ---
echo "--- Phase 4: Configuring Semantic Routing ---"
if [ ${#AGENTS[@]} -gt 0 ]; then
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

echo "--- Installation Complete ---"
if [ ${#AGENTS[@]} -gt 0 ]; then
  echo "You can now start the gateway and chat with your agents using commands like:"
  echo "  openclaw chat ${AGENTS[0]}"
fi