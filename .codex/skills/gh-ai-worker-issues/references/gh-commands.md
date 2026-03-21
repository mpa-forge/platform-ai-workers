# GitHub CLI commands for AI worker issue management

Use these commands as templates and substitute the real repo, issue number, and worker id.

## Inspect a live issue

```powershell
gh issue view 18 --repo mpa-forge/backend-api --json number,title,state,labels,projectItems
```

Use this first to confirm labels and any linked project item status.

## Check lane blockers

List current in-progress work for the lane:

```powershell
gh issue list --repo mpa-forge/backend-api --state open --label worker:worker-01 --label ai:in-progress --json number,title,labels
```

List review backlog for the lane:

```powershell
gh issue list --repo mpa-forge/backend-api --state open --label worker:worker-01 --label ai:ready-for-review --json number,title,labels
```

List queued ready work for the lane:

```powershell
gh issue list --repo mpa-forge/backend-api --state open --label worker:worker-01 --label ai:ready --json number,title,labels
```

## Queue a new issue for pickup

Add the lane label and `ai:ready`:

```powershell
gh issue edit 18 --repo mpa-forge/backend-api --add-label worker:worker-01 --add-label ai:ready
```

If the issue has a conflicting AI state, remove it in the same command:

```powershell
gh issue edit 18 --repo mpa-forge/backend-api --add-label worker:worker-01 --add-label ai:ready --remove-label ai:failed
```

## Requeue a failed issue

```powershell
gh issue edit 18 --repo mpa-forge/backend-api --add-label ai:ready --remove-label ai:failed
```

Preserve the existing `worker:<id>` label unless the lane assignment itself is changing.

## Mark review-driven rework

Only do this when a human reviewer has already requested changes:

```powershell
gh issue edit 18 --repo mpa-forge/backend-api --add-label ai:rework-requested --remove-label ai:ready-for-review
```

Keep the worker lane label unchanged so the same lane can pick the issue back up.

## Verify the final state

```powershell
gh issue view 18 --repo mpa-forge/backend-api --json number,title,labels,projectItems
```

Do not rely on the previous command's success message alone; always reread the issue.
