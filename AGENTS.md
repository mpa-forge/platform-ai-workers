# Agent Context

## Purpose

This repository is part of the platform blueprint workspace. Treat the checked-out repository as the source of truth for code, commands, and validation.

## Always Load

Before making changes:

1. Read `README.md`.
2. Read `Makefile` if present.
3. Check for repo-specific docs under `docs/` that affect the task.

## Working Rules

- Keep changes scoped to the task.
- Prefer repo-local commands over ad hoc alternatives.
- Do not bypass branch protection or write directly to `main`.
- Leave the repository worktree clean when finished.
- If generated artifacts or temporary files are created during the task, remove them before finishing unless they are intended outputs.

## Validation Baseline

Run the strongest repo-local validation entrypoints that exist for the change:

- `make lint`
- `make test`
- `make format-check`

If formatting is required and the repo exposes a formatter, run it and then rerun validation.

If the repo does not expose one of the commands above, fall back to the documented equivalent in `README.md`.

## GitHub Flow

When acting autonomously:

1. Create or use a short-lived task branch.
2. Commit only after validation passes.
3. Push the branch with `git push`.
4. Use `gh` to create or update a draft PR.
5. Do not merge the PR.

## Priority of Instructions

Use this file as the fixed baseline context.

If repo-specific instructions in `README.md`, `docs/`, or task materials conflict with this file, the more specific repo/task instruction wins.
