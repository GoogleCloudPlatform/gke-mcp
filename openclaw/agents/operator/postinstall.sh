#!/usr/bin/env bash
set -e

AGENT_NAME="operator"
CRON_NAME="operator-heartbeat-5m"

if openclaw cron list | grep -q "$CRON_NAME"; then
  echo "[gke-agent] Heartbeat cronjob for $AGENT_NAME already exists. Skipping."
else
  echo "[gke-agent] Adding heartbeat cronjob for $AGENT_NAME..."
  openclaw cron add \
    --name "$CRON_NAME" \
    --agent "$AGENT_NAME" \
    --every 5m \
    --session isolated \
    --message "Read HEARTBEAT.md if it exists (workspace context). Follow it strictly. Do not infer or repeat old tasks from prior chats. If nothing needs attention, reply exactly NO_REPLY." \
    --announce \
    --channel last || echo "Warning: Failed to add cronjob for $AGENT_NAME." >&2
fi
