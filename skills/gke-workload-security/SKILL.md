---
name: gke-workload-security
description: Workflows for auditing and hardening the security of GKE workloads.
---

# GKE Workload Security

This skill provides workflows and best practices for securing GKE workloads. It
covers security auditing, Identity and Access Management (Workload Identity),
Network Security (Network Policies), and Node Security.

## Workflows

### 1. Security Audit

Assess the current security posture of your cluster using the provided audit
script.

**Capabilities:**

- Checks for Workload Identity.
- Verifies Network Policy is enabled.
- Checks if Shielded Nodes are enabled.
- Checks if Binary Authorization is enabled.
- Checks for Private Cluster configuration.

**Command:**

```bash
./.agent/skills/gke-workload-security/scripts/audit_cluster.sh <cluster-name> <region> <project-id>
```

### 2. Configure Workload Identity

Workload Identity allows Kubernetes Service Accounts (KSAs) to impersonate
Google Service Accounts (GSAs). This is the recommended method for workloads to
access Google Cloud APIs.

**Steps:**

1. **Create Namespace and KSA:**

   ```bash
   kubectl create namespace workload-identity-test-ns
   kubectl create serviceaccount <ksa-name> \
       --namespace workload-identity-test-ns
   ```

2. **Bind KSA to GSA:**

   ```bash
   gcloud iam service-accounts add-iam-policy-binding <gsa-name>@<project-id>.iam.gserviceaccount.com \
       --role roles/iam.workloadIdentityUser \
       --member "serviceAccount:<project-id>.svc.id.goog[workload-identity-test-ns/<ksa-name>]"
   ```

3. **Annotate KSA:**

   ```bash
   kubectl annotate serviceaccount <ksa-name> \
       --namespace workload-identity-test-ns \
       iam.gke.io/gcp-service-account=<gsa-name>@<project-id>.iam.gserviceaccount.com
   ```

4. **Verify Example Pod:**
   Use existing asset `assets/workload-identity-pod.yaml` to test the
   configuration. Update the `<ksa-name>` in the file first.

   ```bash
   kubectl apply -f .agent/skills/gke-workload-security/assets/workload-identity-pod.yaml
   ```

### 3. Implement Network Policies

Control traffic flow between Pods using Network Policies. By default, all
traffic is allowed.

**Enable Network Policy Enforcement:**

```bash
gcloud container clusters update <cluster-name> \
    --update-addons=NetworkPolicy=ENABLED \
    --region <region>
```

> [!NOTE]
> If your cluster uses Dataplane V2 (`--enable-dataplane-v2`), Network Policy enforcement is built-in and this step is not required (and may fail).

**Apply Default Deny Policy:**
Isolate namespaces by denying all ingress and egress traffic by default.

**Replace <target-namespace> with the namespace you want to isolate.**

kubectl apply -f .agent/skills/gke-workload-security/assets/default-deny-netpol.yaml -n <target-namespace>

### 4. Enable Shielded Nodes

Ensure nodes are running with verifiable integrity.

**Command:**

```bash
gcloud container clusters update <cluster-name> \
    --enable-shielded-nodes \
    --region <region>
```

### 5. GKE Sandbox (gVisor)

Run untrusted workloads in a sandbox for extra isolation.

**Enable GKE Sandbox:**

```bash
gcloud container clusters update <cluster-name> \
    --enable-gke-sandbox \
    --region <region>
```

**Run a Sandboxed Pod:**
Add `runtimeClassName: gvisor` to your Pod spec.

### 6. Pod Security Standards (Admission)

Enforce security policies on namespaces using labels. GKE Autopilot enforces the `baseline` profile by default.

**Enforce Restricted Profile:**

```bash
kubectl label --overwrite ns <namespace> \
    pod-security.kubernetes.io/enforce=restricted \
    pod-security.kubernetes.io/enforce-version=latest
```

**Audit a Namespace:**

```bash
kubectl label --overwrite ns <namespace> \
    pod-security.kubernetes.io/warn=restricted \
    pod-security.kubernetes.io/warn-version=latest
```

### 7. Binary Authorization

Ensure only trusted images are deployed to your cluster.

**Enable Binary Authorization:**

```bash
gcloud container clusters update <cluster-name> \
    --enable-binauthz \
    --region <region>
```

**Configure Policy:**
Edit the default policy to require attestation or whitelist specific registries.

```bash
gcloud container binauthz policy import policy.yaml
```

**Example policy.yaml:**
```yaml
admissionWhitelistPatterns:
- namePattern: gcr.io/google-containers/*
defaultAdmissionRule:
  evaluationMode: ALWAYS_DENY
  enforcementMode: ENFORCED_BLOCK_AND_AUDIT_LOG
  matchingMode: ALWAYS_MATCH
```

### 8. Secret Manager Integration (CSI Driver)

Mount secrets from Google Cloud Secret Manager directly as volumes in your pods.

**Prerequisites**: Secret Manager CSI driver must be enabled on the cluster.

**Example SecretProviderClass:**

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: my-secret-provider
spec:
  provider: gcp
  parameters:
    secrets: |
      - resourceName: "projects/<project-id>/secrets/my-secret/versions/latest"
        fileName: "my-secret-file"
```

**Example Pod Spec excerpt:**

```yaml
spec:
  containers:
    - name: my-app
      volumeMounts:
        - name: secrets-store-inline
          mountPath: "/mnt/secrets"
          readOnly: true
  volumes:
    - name: secrets-store-inline
      csi:
        driver: secrets-store.csi.k8s.io
        readOnly: true
        volumeAttributes:
          secretProviderClass: "my-secret-provider"
```

### 9. Enable Network Policy Logging

If using GKE Dataplane V2, you can log allowed and denied connections.

**Steps:**

1. Configure the `NetworkLogging` custom resource.

**Example NetworkLogging Manifest:**

```yaml
apiVersion: networking.gke.io/v1alpha1
kind: NetworkLogging
metadata:
  name: default
spec:
  cluster:
    allow:
      log: true
      delegate: true
    deny:
      log: true
      delegate: true
```

This will log connection details to Cloud Logging.

### 10. Policy Controller (Gatekeeper)

Enforce custom policies and compliance across your cluster using Policy Controller (based on Open Policy Agent Gatekeeper).

**Enable Policy Controller:**

```bash
gcloud container clusters update <cluster-name> \
    --enable-managed-anthos-identity \
    --region <region>
# Then enable the feature
gcloud container fleet policy-controller enable --memberships=<membership-name>
```

**Example: Enforce 'No Privileged Containers' Policy**

1. **Install a Template**: Use a pre-built constraint template.
2. **Apply a Constraint**:

```yaml
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sPSPPrivilegedContainer
metadata:
  name: psp-privileged-container
spec:
  match:
    kinds:
      - apiGroups: [""]
        kinds: ["Pod"]
    excludedNamespaces: ["kube-system"]
```

**Capabilities:**
- **Audit**: Detect existing resources that violate policies.
- **Enforce**: Block new resources that violate policies.
- **Mutation**: Automatically modify resources (e.g., add default labels) at creation time.

## Best Practices

1. **Autopilot First**: Use GKE Autopilot to benefit from managed node security, automatic patching, and pre-configured security defaults.
2. **Least Privilege:** Always use Workload Identity with minimal IAM roles. Avoid using Node default service accounts.
3. **Network Isolation:** Use Network Policies to restrict Pod-to-Pod communication. Enable Network Policy Logging for visibility.
4. **Image Security:** Use Binary Authorization with attestations from your CI/CD pipeline.
5. **Secret Management**: Use Secret Manager CSI driver instead of default Kubernetes secrets for sensitive data.
6. **Pod Security**: Enforce `restricted` Pod Security Standards on all non-system namespaces.
7. **Policy Enforcement**: Consider using **Policy Controller** (Gatekeeper) to enforce custom security and compliance policies across the cluster.
