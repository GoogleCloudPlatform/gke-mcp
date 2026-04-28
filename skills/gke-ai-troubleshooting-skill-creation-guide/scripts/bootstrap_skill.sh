#!/bin/bash
# Bootstrap a new GKE troubleshooting skill directory
set -euo pipefail

# --- Configuration & Input ---
REQUIRED_TOOLS=("gum" "gh")
SKILL_NAME="${1:-}"

# --- Utility Functions ---
check_dependencies() {
  local missing_tools=()
  for tool in "${REQUIRED_TOOLS[@]}"; do
    if ! command -v "$tool" >/dev/null 2>&1; then
      missing_tools+=("$tool")
    fi
  done

  if [ ${#missing_tools[@]} -ne 0 ]; then
    echo "❌ Error: Missing required dependencies: ${missing_tools[*]}"
    echo "Please install them before continuing."
    exit 1
  fi
}

# --- Main Logic ---
if [ -z "$SKILL_NAME" ]; then
  echo "Usage: $0 <skill-name>"
  exit 1
fi

check_dependencies

# Refine Title: Remove common prefixes and avoid double "Troubleshooting"
CLEAN_NAME=$(echo "$SKILL_NAME" | sed -E 's/^(gke-)?(ai-)?(troubleshooting-)?//g' | sed -E 's/-troubleshooting$//g')
TITLE_BASE=$(echo "$CLEAN_NAME" | tr '-' ' ' | awk '{for(i=1;i<=NF;i++)sub(/./,toupper(substr($i,1,1)),$i)}1')
FINAL_TITLE="${TITLE_BASE} Troubleshooting"

TARGET_DIR="skills/$SKILL_NAME"
mkdir -p "$TARGET_DIR/references"
mkdir -p "$TARGET_DIR/scripts"

# Create SKILL.md placeholder
cat <<EOF > "$TARGET_DIR/SKILL.md"
---
name: $SKILL_NAME
description: Enter a clear description here.
---

# $FINAL_TITLE

## 🔍 Diagnostic Workflow
### Step 0: Context Acquisition
- **Mandatory**: <project_id>, <location>, <cluster_name>, <timestamp>.
- **Optional**: <node_name>, <workload_name>, <workload_namespace>, <nodepool_name>.

### Step 1: [Low Risk] Investigation via Cloud Logging (LQL)
- **Action**: Call \`query_logs\`.
- **Filter**: 
  \`\`\`
  resource.type="k8s_container" 
  AND resource.labels.project_id="<project_id>"
  AND resource.labels.location="<location>"
  AND resource.labels.cluster_name="<cluster_name>"
  AND textPayload:"ERROR_SIGNATURE"
  \`\`\`

### Step 2: [Low Risk] Investigation via Cloud Monitoring (PromQL)
- **Action**: Call any available monitoring tool or provide PromQL for manual verification.
- **Example Query**:
  \`\`\`promql
  # Fetch container memory usage
  container_memory_usage_bytes{project_id="<project_id>", location="<location>", cluster_name="<cluster_name>", container_name="OPTIONAL_NAME"}
  \`\`\`
EOF

# Create validate_queries.sh template
cat <<EOF > "$TARGET_DIR/scripts/validate_queries.sh"
#!/bin/bash
set -euo pipefail
echo "Validating filters for $SKILL_NAME..."
# gcloud logging read "YOUR_FILTER" --limit=1 > /dev/null
EOF

chmod +x "$TARGET_DIR/scripts/validate_queries.sh"

echo "✅ Success: Skill $SKILL_NAME bootstrapped."
ls -R "$TARGET_DIR"
