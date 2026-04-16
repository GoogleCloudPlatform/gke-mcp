---
name: gke-app-onboarding
description: Workflows for containerizing and deploying applications to GKE for the first time.
---

# GKE App Onboarding Skill

This skill provides workflows for preparing applications that are not yet running on Kubernetes and deploying them to GKE for the first time.

## Workflow

### 1. App Assessment

Before containerizing, assess the application's requirements:

- **Language & Framework**: Identify the tech stack.
- **Dependencies**: List required libraries and services.
- **Configuration**: Determine how the app is configured (e.g., environment variables, config files).
- **Statefulness**: Identify if the app needs persistent storage (databases, file storage).
- **Networking**: Determine port mapping and protocol (HTTP, TCP, etc.).

### 2. Containerization

Create a container image suitable for the application:

- **Dockerfile**: Create a `Dockerfile` in the project root.
- **Multi-stage Builds**: Recommend multi-stage builds to keep the production image small and secure.
- **Logging**: Ensure the application logs to `stdout` and `stderr` for proper log collection.
- **Alternatives**: Consider using **Cloud Native Buildpacks** or **Skaffold** for automated containerization and development workflows without writing Dockerfiles.

### 3. Image Management

Build and store the container image:

- **Build**: Build the image locally or using a CI/CD pipeline.
- **Repository**: Push the image to **Google Artifact Registry**.
- **Vulnerability Scanning**: Enable automatic vulnerability scanning in Artifact Registry to detect security issues in base images and dependencies.

### 4. Manifest Generation

Generate Kubernetes manifests for the application:

- **Deployment**: Create a `Deployment` manifest.
  - Include resource requests and limits.
  - Configure liveness and readiness probes.
- **Compute Classes (Autopilot)**: If your app has specific hardware requirements (e.g., high memory, GPU), use `nodeSelector` to target the appropriate **Compute Class** in GKE Autopilot.
- **Service**: Create a Service manifest (e.g., ClusterIP for internal apps, LoadBalancer for external access). For advanced L7 routing, consider using the [Gateway API](../gke-networking-edge/SKILL.md).
- **Security & Identity**: 
  - Use **Workload Identity** to map Kubernetes Service Accounts to Google Service Accounts for secure API access.
  - Use the **Secret Manager CSI driver** to mount sensitive data directly from Secret Manager as volumes.

### 5. Initial Deployment

Apply the manifests and verify the deployment. **GKE Autopilot** is the recommended target for new applications due to its hands-off operations and security defaults.

- **Apply**: Use `kubectl apply -f <manifest-file>`.
- **Verify**: Check pod status with `kubectl get pods` and ensure the service is accessible.

## Next Steps

Once the application is running, use the [gke-productionize](../gke-productionize/SKILL.md) skill to assess its readiness for production.
