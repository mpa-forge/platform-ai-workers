---
name: gh-ai-worker-issues
description: Manage GitHub issues and linked project items for label-driven AI workers with GitHub CLI. Use when Codex needs to inspect or update worker-pickup state, add or remove `ai:*` and `worker:*` labels, check lane backlog or in-progress blockers, verify whether an issue is eligible for pickup, or reconcile project-board status with the labels the worker actually reads.
---

# AI Worker Issues

Use `gh` as the source of truth for live issue state.

Prefer the smallest safe change that makes the issue eligible for the intended worker behavior, then verify the result with a fresh read.

## Confirm the control model

- Read repo-local docs and runtime code first when the worker behavior is repo-specific.
- Determine whether workers ingest from issue labels, project fields, or both.
- In the platform blueprint workspace, assume issue labels are authoritative unless the target repo says otherwise.
- When the task depends on lane naming or queue rules, inspect files such as `README.md`, `.env.example`, runtime docs, or worker client code before editing live issues.

## Inspect before editing

- Read the issue first with `gh issue view <number> --repo <owner/repo> --json number,title,state,labels,projectItems`.
- Identify the current AI state label, if any, and the current worker lane label, if any.
- Check whether the lane is blocked by an existing `ai:in-progress` issue or a full `ai:ready-for-review` backlog before promising immediate pickup.
- If a project URL or item id is provided, inspect project fields only after you confirm whether the worker actually reads them.

Load [gh-commands.md](references/gh-commands.md) for concrete commands.

## Apply safe state transitions

- Maintain exactly one AI execution state label at a time.
- Maintain exactly one worker lane label when the worker system is lane-bound.
- For new work, set the issue to `ai:ready` plus the correct `worker:<id>` label.
- For rework, use `ai:rework-requested` only when human review has already asked for changes.
- Remove conflicting `ai:*` state labels in the same `gh issue edit` operation when needed.
- Do not guess a lane label if the repo does not expose one; derive it from docs, config, runtime code, or an existing issue pattern first.
- Do not treat project-board status as sufficient if the worker runtime selects only by labels.

## Verify and report

- Re-read the issue after editing instead of assuming the change stuck.
- Report the exact labels added and removed.
- Report whether the issue is immediately eligible, merely queued, or still blocked by backlog or an existing in-progress item.
- If you leave a project status unchanged, say why.
