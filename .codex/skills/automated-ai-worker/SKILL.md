# Automated AI Worker

Use this skill when the repository is being changed by an automated AI worker or by a human trying to follow the same autonomous workflow.

## Workflow

1. Load repository context:
   - `AGENTS.md`
   - `README.md`
   - `Makefile`
2. Understand the task before editing files.
3. Make the smallest change that satisfies the task.
4. Run repo validation:
   - `make lint` if present
   - `make test` if present
   - `make format-check` if present
5. If formatting or lint fixes are needed, apply them and rerun validation.
6. Prepare GitHub delivery:
   - ensure work is on a short-lived branch
   - create a focused commit
   - push the branch
   - create or update a draft PR with `gh`
7. Leave the repo in a clean state.

## GitHub CLI Guidance

Preferred flow:

1. Check current branch and status.
2. Commit with a concise task-focused message.
3. Push with `git push -u origin <branch>` when needed.
4. Create or update a draft PR with `gh pr create` or `gh pr edit`.

Do not:

- push directly to `main`
- merge the PR
- ignore failed validation checks

## Validation Rule

If a validation command fails, fix the reported issue before finishing unless the task explicitly asks for a failing reproduction.
