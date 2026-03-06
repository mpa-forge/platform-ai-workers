# platform-ai-workers

AI worker runtime repository for task-to-code automation in the platform blueprint.

## Structure
- cmd/: worker runtime entrypoints
- internal/: private orchestration and execution logic
- pkg/: shareable public packages
- deploy/: deployment manifests and job packaging assets
- docs/: worker-specific documentation
- scripts/: local utility and developer scripts

## Toolchain
- GNU Make (or a compatible make implementation)
- Go 1.24.12
- Version pin source: .tool-versions and go.mod

## Setup
Before running bootstrap:
- Required: GNU Make (or a compatible make implementation)
- Recommended: mise or sdf for automatic tool installation from .tool-versions
- Fallback: manually install the pinned tool versions listed above

Run the bootstrap command from the repository root:
- Make: make bootstrap

Bootstrap validates the pinned Go toolchain and runs go mod download.
If mise or sdf is available, the script will use it to install the pinned toolchain automatically.

## Run
No runnable worker entrypoint exists yet.
The baseline runtime implementation will be added in P1-T11 and P1-T12.

## Test
No automated test suite is configured yet.
Linting, formatting, and test commands will be introduced in P1-T04 and later tasks.
