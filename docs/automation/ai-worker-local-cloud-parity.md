# AI Worker Local/Cloud Runtime Parity

## Purpose

Define how `platform-ai-workers` runs locally and in Cloud Run Jobs using the same poll-based codepath and container image.

## Parity Requirement

- One runtime entrypoint and poll loop must be used in both contexts.
- Environment-specific behavior is runtime lifecycle only, not separate task-selection or task-execution logic.
- Local runs must exercise the same task-claim/state-transition/PR-update logic used in cloud runs.

## Execution Model

- Build one worker container image.
- Use the same command/entrypoint for local and Cloud Run execution.
- Implement the worker runtime as a Go application.
- Package the coding agent as a CLI available inside the worker runtime environment.
- Invoke the coding agent from Go as a subprocess CLI against the checked-out repository workspace.
- Shared loop behavior (both local and cloud):
  1. Read current lane state and count outstanding tasks/PRs in `ai:ready-for-review`.
  2. If outstanding count is `>= MAX_PENDING_REVIEW`:
     - local mode: sleep `POLL_INTERVAL` and continue.
     - cloud mode: exit with reason `pending_review_limit_reached`.
  3. Select next work item deterministically from GitHub:
     - ready tasks (`ai:ready` + `worker:<id>`),
     - rework tasks (`ai:rework-requested` and/or PRs with unresolved review comments/changes-requested state mapped to this label).
  4. Claim and move to `ai:in-progress` (idempotent).
  5. Run agent subprocess CLI, apply changes, push branch updates, open/update draft PR.
  6. Move to `ai:ready-for-review` on success, or `ai:failed` on failure.
  7. If no candidate exists:
     - local mode: sleep `POLL_INTERVAL` and continue.
     - cloud mode: exit with reason `no_work`.

## Runtime Modes

- `WORKER_RUNTIME_MODE=local`
  - Long-running process on developer machine.
  - Never exits for idle/limit conditions; waits and polls again.
- `WORKER_RUNTIME_MODE=cloud`
  - Bounded Cloud Run Job execution.
  - Exits on `no_work`, `pending_review_limit_reached`, hard timeout, or configured max-loop/max-runtime limit.
  - Re-invoked by GitHub event-trigger workflows (task/rework changes); optional low-frequency scheduler is a safety backstop.

## Required Runtime Inputs

- `WORKER_ID`
- `TARGET_REPO`
- `MAX_PENDING_REVIEW`
- `POLL_INTERVAL`
- `WORKER_RUNTIME_MODE` (`local` or `cloud`)
- optional trigger context:
  - `TRIGGER_SOURCE` (`event`, `scheduled`, `manual`)
  - `TARGET_ISSUE`, `TARGET_PR`, `EVENT_ID`

## Configuration Sources

- Cloud Run: env vars + Secret Manager mounted/injected values.
- Local: `.env.local` (or equivalent local env file) with non-production credentials/tokens.
- Secret names/keys should match across local and cloud where possible to reduce mapping drift.

## Adapter Boundaries

- Wake-up adapter:
  - Cloud Run is started by GitHub event-trigger workflows (and optional scheduler).
  - Local process is started manually once and keeps polling.
- Credentials adapter:
  - Cloud: WIF + GSM-backed secrets.
  - Local: developer-scoped credentials (GitHub App token/PAT and agent key), never committed to git.
- Observability adapter:
  - Keep the same structured logs and run-id fields locally for debugging parity.

## Agent Boundary

- The agent execution path must be identical in local and cloud modes:
  - same worker image
  - same Go runtime entrypoint
  - same subprocess CLI invocation contract
- The worker is responsible for subprocess timeout, exit-code handling, log capture, and post-run diff inspection.
- The agent CLI is responsible for repository file edits only; it does not own task state transitions or merge policy.

## Local Run Commands (Baseline)

- Continuous local loop:
  - `docker run --rm --env-file .env.local -e WORKER_RUNTIME_MODE=local <worker-image> run`
- Optional targeted debug run:
  - `docker run --rm --env-file .env.local -e WORKER_RUNTIME_MODE=local -e TRIGGER_SOURCE=manual -e TARGET_PR=<n> -e EVENT_ID=<id> <worker-image> run`

Equivalent host-native command may be provided for faster iteration, but it must call the same runtime package.

## Validation Requirements

- Local dry-run must prove:
  - continuous polling loop behavior,
  - task selection and claim,
  - deterministic state transitions,
  - branch update and draft PR behavior,
  - idempotent handling for repeated `EVENT_ID`.
- Cloud run must prove the same behavior with the same image tag/entrypoint and cloud lifecycle exits (`no_work`, `pending_review_limit_reached`).

## Non-Goals

- Separate local-only worker implementation.
- Mock-only local flow that bypasses GitHub state transitions.
