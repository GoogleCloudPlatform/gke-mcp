# Test Plan for gke-tpu-dynamic-slices-monitoring

## Manual Verification

### Test Case 1: Triggering and Context Acquisition

1.  **Prompt**: "My dynamic TPU slice is failing to deploy. Can you help me troubleshoot?"
2.  **Expected Output**: The agent should identify the `gke-tpu-dynamic-slices-monitoring` skill and ask for the following context:
    - Project ID
    - Cluster Name
    - Region/Zone
    - Slice Name
    - Optional: Issue Time

### Test Case 2: Diagnostic Recommendation

1.  **Prompt**: Provide dummy values for the requested context (e.g. Project ID: `my-tpu-proj`, Cluster: `tpu-cluster-1`, Region/Zone: `us-central1-a`, Slice Name: `my-slice`).
2.  **Expected Output**: The agent should recommend checking the slice status by providing the command:
    ```bash
    kubectl describe slice my-slice
    ```
    It should also outline the potential states (`SliceNotCreated`, `SliceCreationFailed`, `ACTIVATING`, `ACTIVE`, `FAILED`, etc.) and explain what they mean.

### Test Case 3: Workload Manifest Verification

1.  **Prompt**: "Here is my Job manifest: [paste a Kubernetes Job manifest missing required annotations/nodeSelectors]. Does this look right for dynamic slicing?"
2.  **Expected Output**: The agent should inspect the manifest and point out the missing annotations (e.g., `cloud.google.com/gke-tpu-slice-topology`) and nodeSelectors (e.g., `cloud.google.com/gke-tpu-slice: my-slice`), offering the correct YAML template.
