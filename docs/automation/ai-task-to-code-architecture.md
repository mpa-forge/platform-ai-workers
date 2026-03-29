# AI Task-to-Code Architecture (P0-T11)

## Purpose

Define the baseline architecture and control model for AI-driven task execution across repositories.

## Objective

Convert selected GitHub tasks into implementation PRs while preserving:

- deterministic execution
- least-privilege credentials
- mandatory human review
- reuse of the same worker runtime locally and in cloud

## System Components

- GitHub Issues + Projects:
  - source of truth for task selection and status
- GitHub Actions:
  - wakes cloud workers from task/review events
  - authenticates to GCP via WIF
- Cloud Run Jobs:
  - bounded cloud execution runtime for worker lanes
- Local worker runtime:
  - long-running poll loop for development/debugging
- `platform-ai-workers` repository:
  - source of the shared worker container and runtime logic

## Core Architecture Decisions

- One dedicated repository: `platform-ai-workers`
- One shared poll-loop logic path across local and cloud
- One worker lane per target repository
- One worker output model: branch + PR
- Human review is always required before merge
- Cloud execution is event-woken first, scheduler-backed second
- Worker runtime implementation: Go application orchestrator
- Coding agent integration mode: subprocess CLI invocation from the Go worker

## Worker Lane Model

Each deployed worker lane is bound to one target repository.

Required lane configuration:

- `WORKER_RUNTIME_MODE`
- `WORKER_ID`
- `TARGET_REPO`
- `MAX_PENDING_REVIEW`
- `POLL_INTERVAL`
- credential references

Lane ownership rule:

- a worker only processes items explicitly mapped to its lane (`worker:<id>` or equivalent deterministic mapping)

## Execution Flow

1. Worker wakes up locally or in Cloud Run.
2. Worker checks outstanding items already waiting for review.
3. If pending-review cap is reached:
   - local mode waits and polls again
   - cloud mode exits
4. Worker selects next eligible work item from GitHub.
5. Worker claims the item and moves it to in-progress state.
6. Worker runs the coding agent subprocess CLI against the checked-out repository workspace and applies repository changes.
7. Worker pushes a branch and opens or updates a PR.
8. Worker moves task state to review-ready on success, or failed on error.

## Agent Execution Model

- The worker runtime is a Go application.
- The Go worker is the control plane:
  - clone repository
  - prepare task/rework context
  - invoke the coding agent
  - inspect results
  - run validation commands
  - commit, push, and update GitHub state
- The coding agent is not embedded as custom model orchestration inside the Go code for the baseline.
- The coding agent is invoked as a subprocess CLI from the Go worker process.

Baseline execution contract:

- the Go worker prepares a task instruction bundle in the workspace
- the Go worker invokes the agent CLI with `os/exec` or equivalent subprocess control
- the agent CLI reads and edits files in the checked-out repository workspace
- the Go worker captures exit code, stdout/stderr, and resulting git diff
- the Go worker decides success/failure and performs git/PR/state transitions

Deferred alternatives:

- direct model API integration in Go
- long-running sidecar agent service
- custom in-process SDK orchestration

These alternatives are explicitly deferred unless a later ADR replaces the subprocess CLI baseline.

## Runtime Modes

- `local`
  - continuous loop
  - manual startup
  - same execution logic, different lifecycle behavior
- `cloud`
  - bounded Cloud Run Job execution
  - wakes from GitHub events
  - exits on idle, pending-review cap, or timeout

Detailed parity contract:

- `./ai-worker-local-cloud-parity.md`

## Trigger Model

Primary triggers:

- task becomes ready for AI execution
- PR review sets `changes_requested`
- maintainer comment `/ai rework`

Cloud trigger execution path:

1. GitHub event arrives.
2. GitHub Actions validates actor and target eligibility.
3. Workflow updates labels/state as needed.
4. Workflow authenticates to GCP with WIF.
5. Workflow executes the mapped Cloud Run Job.

Detailed rework trigger contract:

- `./ai-comment-trigger-cloud-run-jobs.md`

## Governance and Control Boundaries

- AI workers may push commits and update PRs.
- AI workers may not merge PRs.
- AI workers may not bypass required checks.
- AI workers may not write directly to protected branches.
- Human review and approval remain mandatory.

## Security Boundaries

- Cloud workers use GCP-native runtime identity and GSM-backed secrets.
- GitHub-trigger workflows use WIF, not static cloud keys.
- Credentials are scoped per lane/repository where possible.
- Full credential model is defined separately in `P0-T13`.

## Dependencies on Follow-Up Docs

- `P0-T12` and `P0-T14`: combined task automation workflow, including state machine and rework protocol
- `P0-T13`: credential and secret model

## Delivery Dependencies

This architecture must be consumed by:

- Phase 1:
  - worker repo bootstrap and local runtime baseline
- Phase 4:
  - AI PR governance and event-driven trigger workflows
- Phase 5:
  - Cloud Run Job, IAM, GSM, and execute permissions

## Acceptance Boundary

`P0-T11` is complete when:

- architecture boundaries are documented
- local/cloud parity is explicitly referenced
- worker lane model is defined
- governance rules are explicit
- downstream phases can implement against this contract without re-deciding the architecture
