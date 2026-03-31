# Agent Context

## Local Entry Point

This file is the repo-local entry point for agent context.

## Always Load

Before making changes:

1. Read `README.md`.
2. Read `Makefile` if present.
3. Read `docs/worker-runtime.md` when the task affects runtime flow, locking, or execution semantics.
4. Run `make sync-agent-skills` before starting major changes or when shared skill guidance may have changed.
5. Read `../platform-blueprint-specs/common/AGENTS.md`.
6. Read `docs/automation/ai-task-to-code-architecture.md`.
7. Read `docs/automation/ai-task-automation-workflow.md`.
8. Read `docs/automation/ai-worker-local-cloud-parity.md`.
9. Read `docs/security/ai-worker-credentials.md`.
10. Read `docs/automation/alert-ai-webhook-spec.md` when the task affects alert intake, webhook validation, or AI incident-summary flow.

## Repo Role

- Own the AI task-to-code automation runtime.
- The Go worker is the control plane; the coding agent runs as a subprocess CLI.
- Local and cloud runs must use the same codepath with environment-specific behavior limited to config and adapters.

## Relevant Shared Constraints

- Worker state machine is label-driven: `ai:ready`, `ai:in-progress`, `ai:ready-for-review`, `ai:rework-requested`, `ai:failed`.
- One active worker is allowed per `worker:<id>` lane.
- Branch + draft PR is the mandatory output path.
- Review and rework happen on the same PR branch.

## Consult Conditionally

- `docs/automation/alert-ai-webhook-spec.md` when the task affects alert intake, webhook validation, or AI incident-summary flow.

## Typical Validation

- `make lint`
- `make test`
- `make format-check`

## Priority of Instructions

Repo-local instructions override shared planning docs.

If local repo docs conflict with a shared planning file, the more specific repo or task instruction wins.
