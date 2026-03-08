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

## Locking logic

The lane lock is implemented in `internal/workspace/workspace.go` as a remote branch lease:

- branch name: `ai-lock/<worker-id>`
- payload file committed on that branch: `.ai-worker-lock.json`

The lock record stores:

- `worker_id`
- `repo`
- `run_id`
- optional `event_id`
- `acquired_at`
- `base_branch`

### Acquisition sequence

1. Fetch remote refs.
2. Check whether `origin/ai-lock/<worker-id>` already exists.
3. If it exists:
   - read `.ai-worker-lock.json`
   - if the lock is younger than `LOCK_STALE_AFTER`, reject the run
   - if the lock is older than `LOCK_STALE_AFTER`, delete the stale lock branch
4. Create a local lock commit containing `.ai-worker-lock.json`
5. Push `ai-lock/<worker-id>` to origin

### Important concurrency detail

The read/check step is not atomic.

Two workers can both observe "no current lock" and both try to acquire one. In that race, the effective owner is the one whose push of `ai-lock/<worker-id>` succeeds. The other run must treat push failure as "lock not acquired".

So the current implementation does **not** guarantee atomicity at the "read remote state" step. It relies on remote branch update success as the ownership decision point.

This is acceptable for the current phase because it prevents concurrent steady-state ownership without introducing extra infrastructure, but it is still a lightweight distributed lock approximation rather than a strict lease service.

### Stale lock reclaim

If a previous run died without releasing the branch lock:

- a later worker reads `.ai-worker-lock.json`
- compares `acquired_at` to `LOCK_STALE_AFTER`
- deletes the stale lock branch
- then attempts normal acquisition

This makes timeout/crash recovery possible, but stale-lock reclaim is the most race-prone part of the current design.

### Current guarantees

The branch lock is intended to provide:

- one active lane owner per `WORKER_ID` in normal operation
- a visible remote record of lane ownership
- simple crash recovery through stale lock expiry

It does **not** provide:

- a fully atomic distributed lock
- transactional stale-lock handoff
- hard guarantees equal to a dedicated lease store such as Firestore, GCS generation locks, or Cloud SQL row locks

If stronger guarantees are needed later, the lock mechanism should move to a dedicated remote lease primitive.

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
