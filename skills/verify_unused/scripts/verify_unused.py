#!/usr/bin/env python3
# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""Script to verify if a GKE/Kubernetes cluster is active or unused before deletion.

Evaluates External Exposure, Persistent Data, and Active Compute criteria with
low-overhead queries (chunking/early return) and fail-close timeout handling.
"""

import argparse
import json
import subprocess
import sys


def run_kubectl(args, timeout_sec=5.0):
  """Runs a kubectl command and parses its JSON output.

  Handles missing API resource types gracefully (returns None).
  Enforces request timeouts and fails close on network or API errors.
  """
  cmd = [
      "kubectl",
  ] + args + [
      "-o",
      "json",
      f"--request-timeout={timeout_sec}s",
  ]
  try:
    result = subprocess.run(
        cmd,
        capture_output=True,
        text=True,
        check=True,
        timeout=timeout_sec + 2.0,
    )
    return json.loads(result.stdout)
  except subprocess.TimeoutExpired as e:
    raise RuntimeError(
        f"Kubectl query timed out after {timeout_sec}s"
    ) from e
  except subprocess.CalledProcessError as e:
    err_msg = (e.stderr or "").lower()
    # Handle resource type not found errors gracefully (e.g., if Gateway or MultiClusterIngress CRDs are not installed)
    if (
        "doesn't have a resource type" in err_msg
        or "not found" in err_msg
        or "the server could not find the requested resource" in err_msg
    ):
      return None
    raise RuntimeError(
        f"Kubectl query failed (exit code {e.returncode}): {(e.stderr or '').strip()}"
    ) from e


def verify_unused(kubeconfig=None, context=None, timeout_sec=5.0):
  base_args = []
  if kubeconfig:
    base_args.extend(["--kubeconfig", kubeconfig])
  if context:
    base_args.extend(["--context", context])

  active_reasons = []

  print("Evaluating GKE cluster safety criteria...")

  try:
    # 1. External Exposure Checks
    # Check Services of type LoadBalancer
    svcs = run_kubectl(base_args + ["get", "services", "-A", "--chunk-size=500"], timeout_sec=timeout_sec)
    if svcs:
      for item in svcs.get("items", []):
        spec = item.get("spec", {})
        if spec.get("type") == "LoadBalancer":
          metadata = item.get("metadata", {})
          active_reasons.append(
              f"External Exposure: Service {metadata.get('namespace')}/{metadata.get('name')} is of type LoadBalancer"
          )
          break  # Early exit on first active match to minimize processing overhead

    # Check Ingresses
    if not active_reasons:
      ingresses = run_kubectl(base_args + ["get", "ingress", "-A", "--chunk-size=500"], timeout_sec=timeout_sec)
      if ingresses and ingresses.get("items"):
        metadata = ingresses["items"][0].get("metadata", {})
        active_reasons.append(
            f"External Exposure: Active Ingress resource present ({metadata.get('namespace')}/{metadata.get('name')})"
        )

    # Check Gateways
    if not active_reasons:
      gateways = run_kubectl(base_args + ["get", "gateway", "-A", "--chunk-size=500"], timeout_sec=timeout_sec)
      if gateways and gateways.get("items"):
        metadata = gateways["items"][0].get("metadata", {})
        active_reasons.append(
            f"External Exposure: Active Gateway resource present ({metadata.get('namespace')}/{metadata.get('name')})"
        )

    # Check MultiClusterIngresses
    if not active_reasons:
      mcis = run_kubectl(base_args + ["get", "multiclusteringress", "-A", "--chunk-size=500"], timeout_sec=timeout_sec)
      if mcis and mcis.get("items"):
        metadata = mcis["items"][0].get("metadata", {})
        active_reasons.append(
            f"External Exposure: Active MultiClusterIngress resource present ({metadata.get('namespace')}/{metadata.get('name')})"
        )

    # 2. Persistent Data Checks
    # Check PVCs in Bound state
    if not active_reasons:
      pvcs = run_kubectl(base_args + ["get", "pvc", "-A", "--chunk-size=500"], timeout_sec=timeout_sec)
      if pvcs:
        for item in pvcs.get("items", []):
          status = item.get("status", {})
          if status.get("phase") == "Bound":
            metadata = item.get("metadata", {})
            active_reasons.append(
                f"Persistent Data: PVC {metadata.get('namespace')}/{metadata.get('name')} is in Bound state"
            )
            break

    # 3. Active Compute Checks
    # Check Pods in Running or Pending state in user namespaces (excluding kube-system and gke-*)
    if not active_reasons:
      pods = run_kubectl(base_args + ["get", "pods", "-A", "--chunk-size=500"], timeout_sec=timeout_sec)
      if pods:
        for item in pods.get("items", []):
          metadata = item.get("metadata", {})
          ns = metadata.get("namespace", "default")
          if ns == "kube-system" or ns.startswith("gke-"):
            continue
          status = item.get("status", {})
          phase = status.get("phase")
          if phase in ("Running", "Pending"):
            active_reasons.append(
                f"Active Compute: Pod {ns}/{metadata.get('name')} is in {phase} state"
            )
            break

  except Exception as e:
    print("=" * 70)
    print("\033[91m[FAIL-CLOSE] Cluster safety check failed due to query error or timeout!\033[0m")
    print(f"Error details: {e}")
    print("To prevent accidental cluster deletion during control-plane outages or network latency, deletion is BLOCKED.")
    print("=" * 70)
    sys.exit(2)

  # Final Verdict
  print("=" * 70)
  if active_reasons:
    print("\033[91m[ACTIVE] Cluster is currently in use! Deletion blocked.\033[0m")
    print("Active workloads/resources detected:")
    for reason in active_reasons:
      print(f"  - {reason}")
    print("=" * 70)
    sys.exit(1)
  else:
    print("\033[92m[UNUSED] Cluster is verified unused (no active compute, exposure, or persistent data).\033[0m")
    print("It is safe to proceed with cluster deletion.")
    print("=" * 70)
    sys.exit(0)


def main():
  parser = argparse.ArgumentParser(
      description="Verify if a GKE/Kubernetes cluster is unused before deletion."
  )
  parser.add_argument(
      "--kubeconfig",
      help="Path to the kubeconfig file to use for verification.",
  )
  parser.add_argument(
      "--context",
      help="Name of the kubeconfig context to verify.",
  )
  parser.add_argument(
      "--timeout",
      type=float,
      default=5.0,
      help="Timeout in seconds for synchronous API queries (defaults to 5.0s).",
  )
  args = parser.parse_args()
  verify_unused(
      kubeconfig=args.kubeconfig,
      context=args.context,
      timeout_sec=args.timeout,
  )


if __name__ == "__main__":
  main()
