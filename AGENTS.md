# AGENTS.md

## Project Overview

This repository contains the GKE MCP server and Gemini CLI extension. The main
server is written in Go. Web UI assets live under `ui/` and are built into
checked-in single-file HTML bundles under `ui/dist/apps/`.

## Repository Layout

- `cmd/`, `pkg/`: Go source for the MCP server, tools, config, and agents.
- `ui/`: Vite/React TypeScript apps and shared UI code.
- `ui/apps/*/index.html`: source app HTML entrypoints.
- `ui/dist/apps/*/index.html`: generated, checked-in UI bundles embedded by Go.
- `dev/tasks/`: local maintenance tasks that may update files.
- `dev/ci/presubmits/`: CI-equivalent verification scripts.

## Development Commands

- Build everything: `make build`
- Build Go only: `go build -v ./...`
- Run Go tests: `go test -v ./...`
- Run Go vet: `go vet ./...`
- Install UI dependencies: `npm --prefix ui install`
- Build UI bundles: `npm --prefix ui run build`
- Run UI tests: `npm --prefix ui run test`
- Run UI lint: `npm --prefix ui run lint`
- Format all supported files: `./dev/tasks/format.sh`
- Format Markdown files after editing: `npx prettier --write <files>`
- Run local presubmit: `./dev/tasks/presubmit.sh` or `make presubmit`

Run the smallest relevant check for the files you changed. For broad changes,
prefer `make presubmit` when feasible.

## Generated Files

The files under `ui/dist/apps/` are generated but checked in because Go embeds
them. Do not edit them by hand. Change source files under `ui/apps/`,
`ui/shared/`, or UI build configuration, then regenerate with:

```sh
npm --prefix ui run build
```

To verify generated UI bundles are current, run:

```sh
./dev/ci/presubmits/verify-ui.sh
```

If `verify-ui.sh` reports changes after an intentional UI update, include the
updated `ui/dist/apps/*/index.html` files in the same change.

## Go Guidelines

- Keep Go code formatted with `gofmt`.
- If dependencies change, run `go mod tidy` or `./dev/tasks/gomod.sh`.
- Add or update focused tests near changed behavior.
- Prefer existing package patterns in `pkg/tools`, `pkg/config`, and
  `pkg/agents` before adding new abstractions.

## UI Guidelines

- The UI uses Vite, React, TypeScript, MUI, and Vitest.
- Keep source changes in `ui/apps/` or `ui/shared/` when possible.
- Run `npm --prefix ui run lint` after TypeScript or React changes.
- Run `npm --prefix ui run test` for UI logic changes.
- Regenerate `ui/dist` after UI source or build configuration changes.

## Pull Request Hygiene

- Perform a self-review of the changes before sending or updating PRs.
- Keep changes scoped to the request.
- Do not commit unrelated formatting or generated output.
- Use Conventional Commits for commit messages.
- Push PR branches to a fork, not to the upstream repository.
- Use `.github/PULL_REQUEST_TEMPLATE.md` for PR body structure and level of
  detail.
- When updating Markdown files, run `npx prettier --write <files>` on the
  changed Markdown files before committing.
- Run `./dev/tasks/presubmit.sh` locally before creating a PR, and only create
  the PR once presubmit passes.
- The presubmit script can take several minutes; when tooling supports it, run
  it as a background or otherwise non-blocking task while continuing other PR
  preparation.
- Mention any checks that could not be run and why.
- For generated bundle changes, explain which source or build configuration
  change caused the regenerated output.
