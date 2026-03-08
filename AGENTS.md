# Agent Context

## Local Entry Point

This file is the repo-local entry point for agent context.

## Always Load

Before making changes:

1. Read `README.md`.
2. Read `Makefile` if present.
3. Read `docs/worker-runtime.md` when the task affects runtime flow, locking, or execution semantics.
4. Read `../platform-blueprint-specs/docs/shared/agent-common-operating-rules.md`.
5. Read `../platform-blueprint-specs/docs/shared/agent-platform-workspace-map.md`.
6. Read `../platform-blueprint-specs/docs/shared/repo-context/platform-ai-workers.md`.

## Priority of Instructions

Repo-local instructions override shared planning docs.

If local repo docs conflict with a shared planning file, the more specific repo or task instruction wins.
