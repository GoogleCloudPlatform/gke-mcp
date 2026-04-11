# GKE Storage Best Practices

This skill provides guidance on managing storage in Google Kubernetes Engine (GKE) clusters.

## Overview

GKE supports various storage options, from Persistent Disks to Cloud Storage. Choosing the right storage type and configuring it correctly is essential for performance and reliability.

## Workflows

### 1. Configure Storage Classes

StorageClasses allow you to describe the "classes" of storage you offer. Different classes might map to quality-of-service levels, or to backup policies.

**Example StorageClass Manifest:**

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: premium-rwo
provisioner: pd.csi.storage.gke.io
parameters:
  type: pd-ssd
  replication-type: regional-pd
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```

Setting `allowVolumeExpansion: true` is highly recommended for production.

### 2. Use CSI Drivers

GKE includes container storage interface (CSI) drivers for dynamic provisioning of storage.

- **Compute Engine Persistent Disk CSI Driver**: Default for block storage.
- **Google Cloud Filestore CSI Driver**: For managed NFS (ReadWriteMany).
- **Cloud Storage FUSE CSI Driver**: For mounting GCS buckets as volumes.

**Example using Filestore CSI Driver:**

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: filestore-pvc
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: standard-rwm # Pre-defined for Filestore
  resources:
    requests:
      storage: 1Ti
```

### 3. Implement Volume Expansion

If `allowVolumeExpansion` is true in the StorageClass, you can resize a volume by updating the PVC manifest.

**Steps:**

1. Edit the PVC manifest and increase the storage request.
2. Apply the changes.

Kubernetes will automatically resize the file system on the volume.

### 4. Volume Snapshots

Capture the state of your persistent volumes at a specific point in time.

**Steps:**

1. **Create a VolumeSnapshotClass**:
   ```yaml
   apiVersion: snapshot.storage.k8s.io/v1
   kind: VolumeSnapshotClass
   metadata:
     name: my-snapshot-class
   driver: pd.csi.storage.gke.io
   deletionPolicy: Delete
   ```

2. **Create a VolumeSnapshot**:
   ```yaml
   apiVersion: snapshot.storage.k8s.io/v1
   kind: VolumeSnapshot
   metadata:
     name: my-snapshot
   spec:
     volumeSnapshotClassName: my-snapshot-class
     source:
       persistentVolumeClaimName: my-pvc
   ```

### 5. High-Performance Local SSDs

For extreme performance, use Local SSDs attached directly to the node.

**Example: Requesting Local SSDs in Standard Node Pool**
```bash
gcloud container node-pools create <pool-name> \
    --cluster=<cluster-name> \
    --local-ssd-count=2 \
    --region <region>
```

**Using Local SSDs in Pods (Autopilot)**
Autopilot automatically provisions nodes with Local SSDs when requested.
```yaml
spec:
  nodeSelector:
    cloud.google.com/gke-local-nvme-ssd: "true"
  containers:
  - name: my-cache
    volumeMounts:
    - name: ssd-storage
      mountPath: /cache
  volumes:
  - name: ssd-storage
    emptyDir:
      medium: LocalSSD
```

## Best Practices

1. **Use CSI Drivers**: Always use the official Google Cloud CSI drivers for best integration and performance.
2. **Enable Volume Expansion**: Always set `allowVolumeExpansion: true` in your StorageClasses to allow for growth.
3. **Choose the Right Disk Type**: Use `pd-ssd` or `pd-extreme` for I/O intensive workloads, and `pd-standard` or `pd-balanced` for others.
4. **Local SSDs**: For ephemeral, high-performance storage (e.g., caching, scratch space), use **Local SSDs**. Note that data is lost when the node is deleted.
5. **Backup & DR**: Ensure all persistent data is protected using [Backup for GKE](../gke-backup-dr/SKILL.md).
6. **Use ReadWriteMany Carefully**: Filestore (NFS) is great for sharing data among multiple Pods, but be aware of file locking and consistency semantics.
