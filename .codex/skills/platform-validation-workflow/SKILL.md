---
name: platform-validation-workflow
description: Apply the platform blueprint validation workflow. Use when deciding which repo-local checks to run, when rerunning formatting and validation, how to fall back from missing make targets, or when scripted/autonomous work should run pre-commit before committing.
---

# Platform Validation Workflow

Use this skill when code or config work needs the platform's validation order applied consistently.

## Default Workflow

1. Run the strongest repo-local validation entrypoints that exist for the change.
2. Prefer documented repo-local commands over ad hoc command invention.
3. If formatting is needed, run the formatter and rerun validation.
4. If a standard command is unavailable, fall back to the documented equivalent in `README.md`.
5. If the repo uses pre-commit and the work is scripted or autonomous, run pre-commit before committing.

## Default Validation Order

When available:

- `make lint`
- `make test`
- `make format-check`

This is the preferred order, not a requirement that every repo expose every target.

## Formatting Rule

- If formatting is required and the repo exposes a formatter, run it.
- After formatting, rerun the relevant validation commands.

## Fallback Rule

- If the repo does not expose one of the standard commands, use the documented equivalent from `README.md` or the repo-local docs.
- Prefer repo-local entrypoints and documented workflows over direct tool invocation.

## Pre-Commit Rule

If the repo has `.pre-commit-config.yaml` and the change is being made by scripted or autonomous workflow:

- prefer `make precommit-run` when available
- otherwise use `python -m pre_commit run --all-files`

## Review Prompt

Before finishing, ask:

- did I run the strongest repo-local checks available for this change?
- did formatting change files after validation, requiring a rerun?
- did I skip a repo-local entrypoint in favor of an ad hoc command without a good reason?
