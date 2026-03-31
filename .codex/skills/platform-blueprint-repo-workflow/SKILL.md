---
name: platform-blueprint-repo-workflow
description: Work from the `platform-blueprint-specs` planning repository while making coordinated changes in sibling repositories such as `frontend-web`, `backend-api`, `backend-worker`, `platform-ai-workers`, `platform-contracts`, and `platform-infra`. Use when tasks are tracked in the planning repo but implementation or configuration changes must be applied across one or more sibling repos, and when branch hygiene, PR-based delivery, and clean git trees must be maintained throughout the session.
---

# Platform Blueprint Repo Workflow

## Overview

Use this skill when the active session starts in `platform-blueprint-specs` but the work spans the sibling implementation repositories in the same parent directory. Treat the planning repo as the control plane for task status and evidence, and treat each sibling repo as an independently protected delivery unit.

## Workflow

### 1. Discover the affected repos

- Start in `platform-blueprint-specs`.
- Identify which sibling repos in the parent directory are relevant to the task.
- Assume the standard sibling repo layout:
  - `../frontend-web`
  - `../backend-api`
  - `../backend-worker`
  - `../platform-ai-workers`
  - `../platform-contracts`
  - `../platform-infra`
- If a required repo is missing locally, clone it before making changes.

### 2. Inspect git state before editing

- Check `git status --short --branch` in the planning repo and in every affected sibling repo before doing work.
- If a repo has unexpected local changes that were not created in the current task, stop and ask before proceeding.
- Do not overwrite, reset, or clean another repo's worktree just to make your task easier.

## Clean Tree Rule

- Keep a clean git tree in every repo involved in the session.
- Before finishing work in a repo:
  - either commit the intended changes, or
  - explicitly leave the repo untouched
- Do not leave behind:
  - abandoned feature branches
  - half-applied patches
  - uncommitted generated files
  - stale local branches whose remote branch has already been merged and deleted
- If PRs are merged via squash, refresh or realign the local repo before continuing more work.
- At the end of the task, verify `git status --short` is clean in each repo you changed unless the user explicitly wants local uncommitted changes.

### 3. Use PR flow, not direct pushes to protected `main`

- Assume sibling repos use protected `main`.
- Create a task-specific branch in each implementation repo that needs changes.
- Push the branch, open a PR, and merge through GitHub.
- Do not push directly to `main`.
- If a stale local branch causes conflicts after squash merges, recreate a clean branch from current remote `main` instead of forcing through broken history.

### 4. Keep planning changes separate from implementation changes

- Record planning/task/evidence updates in `platform-blueprint-specs`.
- Record code/config changes in the target sibling repo only.
- Commit planning-repo updates separately from implementation-repo changes.
- When a task completes across several repos, update the planning repo after the implementation PRs are merged so the evidence reflects reality.

### 5. Prefer deterministic repo-local entrypoints

- If the workflow standardizes a reusable pattern across repos, keep the canonical template or reference in the planning repo and copy it into sibling repos as needed.
- Repo-local files should remain the executable source of truth for that repo.
- Avoid hidden coupling where a repo can only work if the planning repo is present locally.

## Branch Hygiene

- Use short-lived branches.
- Delete local feature branches after they are merged and no longer needed.
- Delete or close stale PR branches that were superseded by cleaner replacements.
- After cleanup, leave each local sibling repo on `main` tracking `origin/main` unless there is an active reason not to.

## Safety Constraints

- Never use destructive git commands such as `git reset --hard` or `git checkout --` on user work without explicit approval.
- If branch cleanup is needed, only delete branches that are clearly stale, merged, and have no uncommitted work.
- If you need to rebuild a feature branch after a squash merge, rebuild it from the current remote `main`, not from a stale local branch.

## Completion Checklist

Before ending a session that uses this skill:

- confirm the target sibling repos received the intended changes
- confirm PRs were created and merged where required
- update planning docs/tasks/evidence in `platform-blueprint-specs`
- clean up stale local branches in touched repos
- verify clean git trees in touched repos unless the user asked to keep changes local
