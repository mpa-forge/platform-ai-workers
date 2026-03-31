---
name: platform-git-release-workflow
description: Apply the platform blueprint git and release policy. Use when creating branches, preparing PRs, choosing merge strategy, planning releases or hotfixes, tagging versions, or handling contract-version compatibility across repos.
---

# Platform Git And Release Workflow

Use this skill when work touches git flow, release policy, or versioning decisions.

## Default Workflow

1. Work on a short-lived branch.
2. Merge into `main` only through a PR.
3. Keep `main` protected and never push to it directly.
4. Never create commits on local `main`; branch before the first commit.
5. Use squash merge on `main`.
6. Treat release tags as the production promotion boundary.

## Branch Policy

- Protected branch:
  - `main`
- Local working rule:
  - `main` must stay aligned with `origin/main`
  - create or switch to a short-lived branch before making the first commit
- Short-lived branches:
  - `feat/<short-description>`
  - `fix/<short-description>`
  - `chore/<short-description>`
  - `hotfix/<short-description>`

## PR Policy

- every change reaches `main` via PR
- PR must be human-reviewed and approved
- required checks must pass before merge
- do not merge the PR yourself unless the task explicitly requires merge handling

Baseline required checks:

- lint
- tests
- code quality gates
- security and secret scans
- contract checks where applicable

## Merge Strategy

- squash merge only on `main`

Use this to keep history clean, make rollback easier by PR unit, and simplify AI-generated iterative work.

## Versioning

- version independently per repository
- use SemVer: `MAJOR.MINOR.PATCH`
- use plain git tags:
  - `vX.Y.Z`

SemVer meaning:

- `PATCH`:
  - fixes and internal changes with no contract break
- `MINOR`:
  - backward-compatible additive behavior
- `MAJOR`:
  - breaking behavior or contract changes

## Release Promotion

- merge to `main`:
  - auto-deploy to `rc`
- promote to `prod`:
  - create release tag `vX.Y.Z`
  - run prod promotion workflow with required approvals and smoke gates

## Hotfix Flow

- use `hotfix/*`
- still require PR, checks, and human approval
- release a new patch tag after merge

## Working Tree And Cleanup Rule

- leave the repository worktree clean when finished
- remove temporary files and generated scratch artifacts unless they are intended outputs
- do not leave abandoned local branches behind

After squash-merged work:

- switch to `main`
- pull latest `origin/main`
- delete the local feature branch if it still exists
- verify the worktree is clean

## Contract Compatibility Rule

For `platform-contracts`:

- require `buf lint`
- require `buf breaking`
- do not make breaking changes in place
- create a new major namespace/package version for breaking changes
- keep generated TypeScript package version aligned to the release tag

## Escalate to Planning Docs

Update platform planning docs or ADRs when the workflow or release policy itself changes, not just when you are following it.
