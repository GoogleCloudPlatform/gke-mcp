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

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

function run_task() {
    local task_path=$1
    echo -e "${BLUE}Running ${task_path}...${NC}"
    if "${task_path}"; then
        echo -e "${GREEN}✓ ${task_path} passed.${NC}"
    else
        echo -e "${RED}✗ ${task_path} failed.${NC}"
        return 1
    fi
}

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

run_task "./dev/ci/presubmits/go-build.sh"
run_task "./dev/ci/presubmits/go-test.sh"
run_task "./dev/ci/presubmits/go-vet.sh"
run_task "./dev/ci/presubmits/docker-build.sh"

# Run golangci-lint if available
if command -v golangci-lint &> /dev/null; then
    run_task "./dev/ci/presubmits/golangci-lint.sh"
else
    echo "Warning: golangci-lint not found. Install it with:"
    echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"
fi

run_task "./dev/tasks/format.sh"
run_task "./dev/tasks/gomod.sh"
run_task "./dev/tasks/super-linter.sh"

echo -e "${GREEN}Local presubmit checks complete, commit any changed files.${NC}"
