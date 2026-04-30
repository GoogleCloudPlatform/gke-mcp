# Test Plan for gke-ai-troubleshooting-tpu-connection-failure-vbar-oom

## Manual Verification

### Test Case 1: Triggering and Context Acquisition

1. **Prompt**: "I have a problem with my TPU v6e nodes crashing. I think it might be a VBAR OOM."
2. **Expected Output**: The agent should identify this skill and ask for:
   - Project ID
   - Cluster Name
   - Node Name
   - Issue Time
   - Workload Name (JobSet Name)
   - Workload Namespace

### Test Case 2: Diagnostic Execution

1. **Prompt**: Provide the requested context (dummy values are fine for simulation).
2. **Expected Output**: The agent should suggest running `query_logs` with the filter specified in Step 1 of the skill, using the calculated time window.

### Test Case 3: Relative Time Rejection

1. **Prompt**: "My nodes crashed about 5 minutes ago."
2. **Expected Output**: The agent should reject the relative time and request an exact timestamp or a specific time window as per the "Time Handling Rules".
