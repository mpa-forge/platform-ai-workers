# AI Task Automation Workflow (P0-T12, P0-T14)

## Purpose

Define the full automation workflow for AI-managed tasks, from initial pickup through review and rework.

## Scope

This document is the source of truth for:

- AI task execution states
- worker loop behavior
- review and rework transitions
- branch and PR update rules

Detailed cloud wake-up implementation remains in:

- `docs/automation/ai-comment-trigger-cloud-run-jobs.md`

## State Model

Each AI-eligible task must have exactly one AI execution state:

- `ai:ready`
- `ai:in-progress`
- `ai:ready-for-review`
- `ai:rework-requested`
- `ai:failed`

State exclusivity rule:

- a task may not hold more than one of the states above at the same time

Lane assignment rule:

- each AI-managed task must also carry exactly one `worker:<id>` label
- a worker may only pick tasks whose `worker:<id>` matches its own lane id

## Worker Start Conditions

A worker may start:

- on schedule
- after a state change for an AI-eligible task

Concurrency rule:

- if that worker lane is already running, do not start another run for the same worker
- no two worker processes with the same `worker:<id>` may run simultaneously

## End-to-End Workflow

1. A human marks a task as `ai:ready`.
2. The matching worker starts on schedule or wake-up event.
3. The worker ingests eligible work for its own `worker:<id>`.
4. The worker moves the task to `ai:in-progress`.
5. The worker performs the coding work.
6. If work fails, the task moves to `ai:failed`.
7. If work succeeds, the worker commits and pushes changes, opens or updates the PR, and moves the task to `ai:ready-for-review`.
8. A human reviewer reviews the PR.
9. If changes are required, the reviewer comments on the PR and moves the task to `ai:rework-requested`.
10. The worker picks that rework first, moves the task back to `ai:in-progress`, updates the same PR branch, and the cycle repeats.

## Worker Loop

The worker loop is:

1. `ingest`
2. `work`
3. `push`

After `push`, the worker checks review backlog and either stops or starts again at `ingest`.

## Ingest Step

Selection priority:

1. tasks in `ai:rework-requested` for the current `worker:<id>`
2. tasks in `ai:ready` for the current `worker:<id>`

When a task is selected:

- remove its current eligible state
- move it to `ai:in-progress`

## Work Step

During work, the worker executes the coding task for the selected issue/PR context.

Failure rule:

- if work fails, move task from `ai:in-progress` to `ai:failed`
- if work succeeds, continue to `push`

## Push Step

During push, the worker:

- commits changes
- pushes branch updates
- opens or updates the PR

Success transition:

- move task from `ai:in-progress` to `ai:ready-for-review`

## Review and Rework

Rework starts only after:

- the worker has pushed changes
- the task state is `ai:ready-for-review`
- a human reviewer reviews the PR and requests changes

When changes are needed, the reviewer:

- adds comments to the PR
- moves the task state from `ai:ready-for-review` to `ai:rework-requested`

Rework rule:

- `ai:rework-requested` always has priority over `ai:ready`
- rework may be picked during a triggered run or a scheduled run
- rework uses the same execution loop as initial work

## PR Update Rule

- Work and rework always update the existing PR branch
- Rework does not create a new PR for the same task

This keeps review history and reviewer comments attached to a single PR thread.

## Pending Review Limit

Review backlog is measured as the number of tasks in `ai:ready-for-review` for that worker lane.

Stop rule:

- if tasks in `ai:ready-for-review` are at the configured limit, stop the worker
- otherwise, return to `ingest`

## Trigger Sources

Cloud wake-up sources may include:

- PR review `changes_requested`
- maintainer command comment such as `/ai rework`
- scheduled run catching already-labeled `ai:rework-requested`

## Governance Rules

- only a human reviewer moves reviewed work into `ai:rework-requested`
- rework remains subject to the same branch protection and required checks
- the worker may push updates, but may not merge

## Determinism Rules

- `ai:rework-requested` always has priority over `ai:ready`
- only one worker lane may own a task at a time
- a running worker lane is not started again concurrently
- at most one active worker process may exist for a given `worker:<id>`

## Manual/Human Interaction Points

- humans move tasks into `ai:ready`
- review feedback moves tasks into `ai:rework-requested`
- humans review tasks in `ai:ready-for-review`
- humans may inspect and re-queue tasks in `ai:failed`
