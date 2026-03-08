# Worker Runtime

## Purpose

Explain the runtime shape of `platform-ai-workers` without forcing the reader to reconstruct it from `internal/`.

## Main flow

1. `cmd/worker/main.go` loads env-backed config and starts the worker.
2. `internal/app/app.go`:
   - prepares the reusable workspace clone
   - acquires a per-lane lock
   - runs the shared poll loop
3. `internal/githubcli/client.go` reads and mutates issue/PR state through `gh`.
4. `internal/workspace/workspace.go` manages:
   - clean workspace resets
   - remote lock branch lease
5. `internal/prompt/prompt.go` renders the task prompt from `prompts/task.md.tmpl`.
6. `internal/agent/codex.go` invokes Codex CLI as a subprocess in the checked-out repo.

## Selection order

The worker processes one lane only:

- lane label: `worker:<id>`

Within the lane, selection is:

1. existing `ai:in-progress` issue
2. `ai:rework-requested`
3. `ai:ready`

This keeps interrupted work ahead of new work.

## Safety model

The current phase uses two lightweight controls:

- remote branch lock:
  - `ai-lock/<worker-id>`
- issue comment event markers:
  - `automation-marker ... automation-event-id:<id>`

Together they provide:

- one active lane owner at a time
- dedupe of repeated GitHub event wake-ups

## Deliberate boundary

The Go worker is the control plane.

It owns:

- issue selection
- label transitions
- workspace reset
- lane lock
- event dedupe
- success/failure handling

Codex owns only the repository task execution inside the checked-out workspace:

- edit files
- run repo validation commands
- commit/push
- create/update draft PR

That split is intentional and matches the planning docs for `P1-T11` and `P1-T12`.
