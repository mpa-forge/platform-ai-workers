# platform-ai-workers

AI worker runtime repository for task-to-code automation in the platform blueprint.

## Structure
- `cmd/`: worker runtime entrypoints
- `internal/`: private orchestration and execution logic
- `pkg/`: shareable public packages
- `deploy/`: deployment manifests and job packaging assets
- `docs/`: worker-specific documentation
- `scripts/`: local utility and developer scripts

## Setup
This repository is currently at the skeleton stage.
Go version pinning, bootstrap commands, and dependency wiring will be added in `P1-T03`.

## Run
No runnable worker entrypoint exists yet.
The baseline runtime implementation will be added in `P1-T11` and `P1-T12`.

## Test
No automated test suite is configured yet.
Linting, formatting, and test commands will be introduced in `P1-T04` and later tasks.
