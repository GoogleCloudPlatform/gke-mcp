#!/usr/bin/env bash
# Copyright 2025 Google LLC
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

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

failed=0

for skill_dir in skills/*; do
    if [ -d "$skill_dir" ]; then
        skill_file="$skill_dir/SKILL.md"
        if [ -f "$skill_file" ]; then
            # Check if file starts with ---
            if ! head -n 1 "$skill_file" | grep -q "^---$"; then
                echo "Error: $skill_file does not start with ---"
                failed=1
                continue
            fi
            
            # Extract frontmatter content (between first and second ---)
            fm_content=$(awk '
                BEGIN { count=0; content="" }
                /^---$/ { count++; next }
                count==1 { content = content $0 "\n" }
                count==2 { exit }
                END { if (count >= 2) printf "%s", content }
            ' "$skill_file")
            
            if [ -z "$fm_content" ]; then
                echo "Error: $skill_file has empty or invalid frontmatter (missing closing ---?)"
                failed=1
                continue
            fi
            
            if ! echo "$fm_content" | grep -q "^name:[[:space:]]*[^[:space:]]"; then
                echo "Error: $skill_file is missing a non-empty 'name' in frontmatter"
                failed=1
            fi

            if echo "$fm_content" | grep -q "^name:"; then
                skill_name=$(echo "$fm_content" | grep "^name:" | sed 's/^name:[[:space:]]*//' | xargs)
                dir_name=$(basename "$skill_dir")
                if [ "$skill_name" != "$dir_name" ]; then
                    echo "Error: Skill name '$skill_name' does not match directory name '$dir_name' in $skill_file"
                    failed=1
                fi
            fi
            
            if ! echo "$fm_content" | grep -q "^description:[[:space:]]*[^[:space:]]"; then
                echo "Error: $skill_file is missing a non-empty 'description' in frontmatter"
                failed=1
            fi
            
        else
            echo "Error: Missing SKILL.md in $skill_dir"
            failed=1
        fi
    fi
done

if [ $failed -ne 0 ]; then
    echo "Skills validation failed."
    exit 1
fi

echo "All skills validated successfully."
