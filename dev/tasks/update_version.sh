#!/usr/bin/env bash
# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http: //www.apache.org/licenses/LICENSE-2.0
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

git fetch upstream main
git checkout upstream/main

TAG=$(git tag --points-at HEAD)
VERSION="${TAG/v/}"

git branch -D "update-version-${TAG}" || true
git checkout -b "update-version-${TAG}"

tmp=$(mktemp)
jq --arg ver "${VERSION}" '.version = $ver' gemini-extension.json > "${tmp}" && mv "${tmp}" gemini-extension.json

git add gemini-extension.json
git commit -m "chore: Update Gemini extension JSON to ${TAG}"
git push -u origin "update-version-${TAG}"

gh pr create -f
