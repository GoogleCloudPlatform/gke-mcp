#!/usr/bin/env python3
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import sys

try:
    import yaml
except ImportError:
    print("Error: PyYAML is required. Install it with 'pip install PyYAML'.")
    sys.exit(1)

def check_workflow(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        try:
            wf = yaml.safe_load(f)
        except yaml.YAMLError as e:
            print(f"Error parsing {file_path}: {e}")
            return False

    if not wf:
        return True

    if not isinstance(wf, dict):
        return True

    triggers = wf.get('on', {})
    has_pr_target = False
    if isinstance(triggers, str):
        has_pr_target = (triggers == 'pull_request_target')
    elif isinstance(triggers, list):
        has_pr_target = 'pull_request_target' in triggers
    elif isinstance(triggers, dict):
        has_pr_target = 'pull_request_target' in triggers

    if not has_pr_target:
        return True

    print(f"Checking {file_path} for insecure pull_request_target usage...")
    failed = False

    jobs = wf.get('jobs', {})
    if not isinstance(jobs, dict):
        return True
    for job_name, job in jobs.items():
        if not isinstance(job, dict):
            continue
        steps = job.get('steps', [])
        if not isinstance(steps, list):
            continue
        for step in steps:
            if not isinstance(step, dict):
                continue
            uses = step.get('uses', '')
            if not isinstance(uses, str):
                continue
            if uses.startswith('actions/checkout@') or uses == 'actions/checkout':
                with_params = step.get('with')
                if not isinstance(with_params, dict):
                    with_params = {}
                ref = str(with_params.get('ref', ''))
                repo = str(with_params.get('repository', ''))

                ref_lower = ref.lower()
                repo_lower = repo.lower()

                if (("pull_request" in ref_lower and "head" in ref_lower) or 
                    "head_ref" in ref_lower or 
                    "refs/pull/" in ref_lower):
                    print(f"Error: {file_path} (job: {job_name}) uses pull_request_target and checks out untrusted code (ref: {ref})")
                    failed = True
                if "pull_request" in repo_lower and "head" in repo_lower:
                    print(f"Error: {file_path} (job: {job_name}) uses pull_request_target and checks out an untrusted repository (repository: {repo})")
                    failed = True

    return not failed

def main():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    workflow_dir = os.path.abspath(os.path.join(script_dir, '../../..', '.github/workflows'))
    if not os.path.isdir(workflow_dir):
        print(f"No workflows directory found at {workflow_dir}")
        sys.exit(0)

    failed = False
    for filename in os.listdir(workflow_dir):
        if filename.endswith('.yml') or filename.endswith('.yaml'):
            file_path = os.path.join(workflow_dir, filename)
            if not check_workflow(file_path):
                failed = True

    if failed:
        print("Workflow security check failed.")
        sys.exit(1)

    print("Workflow security check passed.")
    sys.exit(0)

if __name__ == '__main__':
    main()
