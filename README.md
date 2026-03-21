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
- Go `1.25.1`
- Version pin source: `.tool-versions` and `go.mod`

## Setup

Before running bootstrap:

- Shared workspace requirement: keep `platform-blueprint-specs` checked out as a sibling directory if you want to use `make doctor`.
- Required: GNU Make (or a compatible `make` implementation) and a bash-compatible shell
- Recommended: `mise` or `asdf` for automatic tool installation from `.tool-versions`
- Fallback: manually install the pinned tool versions listed above

Run the setup commands from the repository root:

- Workstation checks: `make doctor`
- Bootstrap: `make bootstrap`

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
- For local container runs with ChatGPT account auth:

  - keep `AGENT_AUTH_MODE=chatgpt`
  - set `GITHUB_TOKEN` in `.env.local`
  - leave `OPENAI_API_KEY` empty unless you intentionally switch to `AGENT_AUTH_MODE=api`
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
  - `RUN_ID`
  - `LOCK_STALE_AFTER`
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
- Prepare persistent Codex auth for local container runs:

  - `make build-image`
  - `make container-codex-login`
  - sign in with your ChatGPT account inside the one-off container
- Container execution:

  - `make run-container`

## Local Container Auth

Use this flow when you want the worker container to spend ChatGPT account usage instead of API credits:

1. Create or update `.env.local` from `.env.example`.
2. Set `GITHUB_TOKEN` in `.env.local`.
3. Keep `AGENT_AUTH_MODE=chatgpt`.
4. Run `make build-image`.
5. Run `make container-codex-login` once to populate the persistent Docker mount at `$HOME/.docker-platform-ai-workers/codex`.
6. Run `make run-container`.

Notes:

- `make run-container` mounts the persistent Codex state directory into `/root/.codex` inside the container.
- The current worker still requires `GITHUB_TOKEN`; GitHub account login through `gh auth login` alone does not satisfy startup validation.
- `OPENAI_API_KEY` is only required when `AGENT_AUTH_MODE=api`.

Current baseline behavior:

- one Go runtime entrypoint for local and cloud execution
- shared GitHub poll loop for `ai:rework-requested` then `ai:ready`
- issue state transitions:

  - `ai:ready` -> `ai:in-progress` -> `ai:ready-for-review`
  - `ai:rework-requested` -> `ai:in-progress` -> `ai:ready-for-review`
  - failures move the issue to `ai:failed`
- per-worker remote lane lock via `ai-lock/<worker-id>` branch in the target repo
- existing `ai:in-progress` work is resumed before new ready/rework selection
- event-triggered reruns are deduplicated through `EVENT_ID` markers stored on the issue
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
- For Docker-based local runs, prefer `make container-codex-login` and `make run-container` so the container reuses a persistent Codex login mount.
- The baseline keeps branch/PR operations inside the agent prompt so the same workflow can run locally and later in Cloud Run with the same Codex CLI contract.
- Runtime structure notes: `docs/worker-runtime.md`
- Branch-lock semantics and limitations are documented in `docs/worker-runtime.md#locking-logic`
