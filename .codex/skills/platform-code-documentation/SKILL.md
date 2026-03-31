---
name: platform-code-documentation
description: Apply the platform blueprint code documentation standard. Use when adding or updating comments, package docs, README/docs content, runbooks, contract comments, or reviewing whether a change needs documentation in Go, TypeScript, Terraform, protobuf, or shared platform docs.
---

# Platform Code Documentation

Use this skill when code or docs work needs the platform's documentation rules applied consistently.

## Core Rule

Document why, contracts, constraints, and non-obvious behavior.
Do not document obvious mechanics.

## Default Workflow

1. Identify the smallest durable documentation layer that matches the change.
2. Update code and documentation in the same change.
3. Prefer language-native doc styles first.
4. Add repo docs only when behavior, runtime flow, architecture, or operations would be hard to infer from code alone.
5. Escalate to planning docs only when the behavior affects multiple repos or a platform policy.

## Cross-Repo Documentation Rule

When a change affects other repositories or the platform contract between repositories, also update `platform-blueprint-specs` in the same task.

Cross-repo examples:

- API contract, protobuf, auth, or runtime changes that affect frontend, workers, infra, or AI automation
- generated client or package usage changes
- deployment, observability, or environment model changes that affect more than one repo
- workflow or operational changes that alter how another repo must be built, validated, deployed, or integrated

When cross-repo impact exists, update the smallest durable planning artifact that matches the change:

- standards in `common/standards/` and matching skills in `.codex/skills/` when they exist
- architecture or automation docs in planning docs or repo-owned docs linked from planning artifacts
- phase or task files when scope, sequencing, or acceptance criteria change
- `platform-specification.md` when a locked stack or architecture decision changes
- ADRs when the change is a platform-level decision or tradeoff

Do not keep cross-repo behavioral knowledge only in a code repo if other repos depend on it.

## Documentation Placement

- Inline comment:
  - one local non-obvious behavior, invariant, retry rule, concurrency rule, or side effect
- Doc comment:
  - exported symbol or reusable internal contract
- Package/module doc:
  - subsystem responsibility, lifecycle, or boundaries
- Repo docs:
  - setup, runtime, architecture, operations, troubleshooting
- Planning docs or ADRs:
  - cross-repo standards, platform policy, architecture decisions

## What Good Documentation Covers

- why something exists
- important constraints or invariants
- side effects and failure behavior
- lifecycle, retry, timeout, or concurrency expectations
- public contract behavior

## What to Avoid

- line-by-line restatements of the code
- comments that only rename the statement underneath
- duplicating the same explanation in every layer
- leaving stale docs behind after behavior changes

## Contract-Focused Areas

Be explicit when documenting:

- environment variables and config coupling
- auth behavior and error semantics
- protobuf services, RPCs, messages, enums, and fields
- retries, idempotency, locking, and distributed coordination
- generated code ownership and generation source

## Language Baselines

- Go:
  - document exported identifiers that form package contracts
  - add package comments when boundaries are not obvious
- TypeScript:
  - document exported modules, hooks, utilities, and non-obvious public contracts selectively
- Terraform:
  - prefer clear names first, then document variables/outputs when coupling or ownership is non-obvious
- Protobuf:
  - treat `.proto` comments as part of the interface contract

## Review Rule

When reviewing or changing code, ask:

- did the change introduce non-obvious behavior without explanation?
- are public/runtime contracts still documented accurately?
- should the explanation live in code, repo docs, or planning docs?
