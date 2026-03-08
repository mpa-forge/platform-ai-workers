# platform-ai-workers

AI worker runtime repository for task-to-code automation in the platform blueprint.

## Structure
- `cmd/`: worker runtime entrypoints
- `internal/`: private orchestration and execution logic
- `pkg/`: shareable public packages
- `deploy/`: deployment manifests and job packaging assets
- `docs/`: worker-specific documentation
- `scripts/`: local utility and developer scripts

## Toolchain
- GNU Make (or a compatible `make` implementation) and a bash-compatible shell
- Go `1.24.12`
- Version pin source: `.tool-versions` and `go.mod`

## Setup
Before running bootstrap:
- Required: GNU Make (or a compatible `make` implementation) and a bash-compatible shell
- Recommended: `mise` or `asdf` for automatic tool installation from `.tool-versions`
- Fallback: manually install the pinned tool versions listed above

Run the bootstrap command from the repository root:
- Make: `make bootstrap`

Bootstrap validates the pinned Go toolchain and runs `go mod download`.
If `mise` or `asdf` is available, the script will use it to install the pinned toolchain automatically.

## Lint and Format
- Install git hooks: `make precommit-install`
- Run all pre-commit checks manually: `make precommit-run`
- Run repo lint checks: `make lint`
- Apply formatting: `make format`
- Check formatting only: `make format-check`

## Environment
- Copy `.env.example` to `.env.local` for local development
- Required local baseline variables:
  - `APP_ENV`
  - `LOG_LEVEL`
  - `WORKER_RUNTIME_MODE`
  - `WORKER_ID`
  - `TARGET_REPO`
  - `MAX_PENDING_REVIEW`
  - `POLL_INTERVAL`
  - `GITHUB_TOKEN`
- Additional runtime variables for the worker baseline:
  - `BASE_BRANCH`
  - `WORKSPACE_ROOT`
  - `AGENT_CLI`
  - `AGENT_AUTH_MODE`
  - `AGENT_MODEL`
  - `PROMPT_TEMPLATE`
  - `GITHUB_TOKEN`
  - `OPENAI_API_KEY`
  - `TRIGGER_SOURCE`
  - `TARGET_ISSUE`
  - `TARGET_PR`
  - `EVENT_ID`
  - `DRY_RUN`

## Run
- Local host execution:
  - `make run`
- Container execution:
  - `docker run --rm --env-file .env.local platform-ai-workers:local run`

Current baseline behavior:
- one Go runtime entrypoint for local and cloud execution
- shared GitHub poll loop for `ai:rework-requested` then `ai:ready`
- issue state transitions:
  - `ai:ready` -> `ai:in-progress` -> `ai:ready-for-review`
  - `ai:rework-requested` -> `ai:in-progress` -> `ai:ready-for-review`
  - failures move the issue to `ai:failed`
- worker-owned reusable clone under `WORKSPACE_ROOT`
- Codex CLI subprocess execution against the checked-out target repository
- prompt-template-driven task instructions in `prompts/task.md.tmpl`
- agent is instructed to run `make lint`, commit, push, and create/update the draft PR with `gh`
- local mode keeps polling; cloud mode exits on `no_work` or `pending_review_limit_reached`

## Test
- Unit tests: `make test`
- Lint: `make lint`
- Formatting: `make format` / `make format-check`

## Implementation Notes
- Local and cloud use the same worker binary and the same runtime codepath.
- `AGENT_AUTH_MODE=chatgpt` assumes Codex CLI is already logged in on the machine or inside the container.
- `AGENT_AUTH_MODE=api` requires `OPENAI_API_KEY`.
- The baseline keeps branch/PR operations inside the agent prompt so the same workflow can run locally and later in Cloud Run with the same Codex CLI contract.
