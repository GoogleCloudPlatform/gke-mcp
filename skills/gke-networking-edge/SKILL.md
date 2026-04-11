---
name: gke-networking-edge
description: Workflows for configuring edge networking, ingress, and security on GKE.
---

# GKE Networking Edge Skill

This skill provides workflows for exposing applications running on GKE securely to the internet or internal networks.

## Workflows

### 1. Configure Gateway API (Recommended)

The Gateway API is the modern way to manage routing in Kubernetes.

**Prerequisites**: Gateway API must be enabled on the cluster (enabled by default in GKE 1.24+).

**Example Gateway Manifest:**

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: my-gateway
  namespace: my-namespace
spec:
  gatewayClassName: gke-l7-global-external-managed # GKE managed external L7 load balancer
  listeners:
    - name: http
      protocol: HTTP
      port: 80
```

**Example HTTPRoute Manifest:**

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: my-route
  namespace: my-namespace
spec:
  parentRefs:
    - name: my-gateway
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: my-service
          port: 80
```

### 2. Configure Standard GKE Ingress

Use standard Ingress for simpler use cases or legacy setups.

**Example Ingress Manifest:**

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
  namespace: my-namespace
  annotations:
    kubernetes.io/ingress.class: "gce"
spec:
  rules:
    - http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-service
                port:
                  number: 80
```

### 3. Secure with Cloud Armor

Cloud Armor provides WAF and DDoS protection.

**Enable Cloud Armor via BackendConfig:**

1. Create a Security Policy in Cloud Armor (usually via gcloud or Terraform).
2. Reference it in a `BackendConfig` in GKE.

**Example BackendConfig:**

```yaml
apiVersion: cloud.google.com/v1
kind: BackendConfig
metadata:
  name: my-backend-config
  namespace: my-namespace
spec:
  securityPolicy:
    name: my-cloud-armor-policy
```

3. Associate `BackendConfig` with your `Service` via annotations.

### 4. Configure Google-Managed SSL Certificates

Automatically provision and renew SSL certificates.

**Example ManagedCertificate (Legacy Ingress):**

```yaml
apiVersion: networking.gke.io/v1
kind: ManagedCertificate
metadata:
  name: my-certificate
spec:
  domains:
    - example.com
```

Reference it in Ingress annotations: `networking.gke.io/managed-certificates: my-certificate`.

**Gateway API Approach:**
Use the `gateway.networking.k8s.io` API with `CertificateManager` integration for more advanced certificate management.

### 5. Secure with Identity-Aware Proxy (IAP)

IAP provides identity-based access control for your applications.

**Enable IAP via BackendConfig:**

1. Configure OAuth Consent Screen and credentials in Google Cloud Console.
2. Create a Kubernetes Secret containing your OAuth client ID and secret.

```bash
kubectl create secret generic my-iap-secret \
    --from-literal=client_id=<client-id> \
    --from-literal=client_secret=<client-secret>
```

3. Create a `BackendConfig` to enable IAP.

**Example BackendConfig:**

```yaml
apiVersion: cloud.google.com/v1
kind: BackendConfig
metadata:
  name: my-iap-config
spec:
  iap:
    enabled: true
    oauthclientCredentials:
      secretName: my-iap-secret
```

4. Associate with your `Service` via annotations.

### 6. Service Mesh with Gateway API (GAMMA)

Use the Gateway API to manage internal service-to-service traffic (GAMMA).

**Example HTTPRoute for internal traffic:**

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: internal-route
spec:
  parentRefs:
    - name: my-service # Reference the Service itself as the parent for internal routing
      kind: Service
      group: ""
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /v2
      backendRefs:
        - name: my-service-v2
          port: 8080
```

### 7. Enable Container-Native Load Balancing (Recommended)

Container-native load balancing allows load balancers to target Kubernetes Pods directly, rather than targeting nodes. This improves latency and distribution.

**Prerequisites**: Cluster must be VPC-native.

**How it works**:

- For GKE Ingress and Gateway API, container-native load balancing is enabled by default via Network Endpoint Groups (NEGs).
- To verify or explicitly enable it for a Service, use the `cloud.google.com/neg` annotation.

**Example Service Manifest:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
  annotations:
    cloud.google.com/neg: '{"ingress": true}' # Enabled for Ingress
spec:
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  selector:
    app: my-app
  type: ClusterIP
```

### 8. Configure Private Service Connect (PSC)

Private Service Connect allows you to expose services in one VPC to consumers in another VPC securely, without VPC peering.

**Steps:**

1. Create an internal load balancer for your service.
2. Create a `ServiceAttachment` referencing the load balancer.

**Example ServiceAttachment Manifest:**

```yaml
apiVersion: networking.gke.io/v1
kind: ServiceAttachment
metadata:
  name: my-psc-attachment
  namespace: my-namespace
spec:
  connectionPreference: ACCEPT_AUTOMATIC
  natSubnets:
    - my-psc-nat-subnet # Subnet dedicated for PSC NAT
  targetService:
    name: my-service
    namespace: my-namespace
```

Share the `ServiceAttachment` URI with consumers to create a PSC endpoint in their VPC.

## Best Practices

1. **Prefer Gateway API**: It offers more flexibility and role separation than Ingress. Use it for both external LBs and internal Service Mesh (GAMMA).
2. **Enable Cloud Armor**: Always protect public-facing endpoints with Cloud Armor.
3. **Use Identity-Aware Proxy (IAP)**: For internal tools or back-office apps, use IAP to enforce identity-based access control.
4. **Use Certificate Manager**: Prefer Google Cloud Certificate Manager with Gateway API for scalable certificate management.
5. **Use Container-Native Load Balancing**: Always use NEGs for HTTP(S) load balancing to reduce latency and improve traffic distribution.
