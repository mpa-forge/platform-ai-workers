# AI Rework Trigger: GitHub Comment/Review -> Cloud Run Job

## Purpose

Define how human review feedback triggers immediate AI rework runs without waiting for scheduler cadence.
Runtime parity requirements for local/cloud execution are defined in `docs/automation/ai-worker-local-cloud-parity.md`.
This trigger wakes the cloud worker poller loop; it does not define a separate worker logic path.

## Trigger Sources

- `pull_request_review` with `state=changes_requested`
- `issue_comment` on a PR thread with maintainer command `/ai rework`

## End-to-End Flow

1. Reviewer submits `changes requested` or `/ai rework`.
2. GitHub Actions workflow validates actor permissions and PR eligibility (`ai-generated` label or equivalent).
3. Workflow updates labels/state (`ai:rework-requested`, remove `ai:ready-for-review`).
4. Workflow authenticates to GCP through Workload Identity Federation.
5. Workflow executes mapped Cloud Run Job on-demand (wake-up).
6. Worker enters the standard poll loop, prioritizes eligible rework items for its lane, pushes updates to the same PR branch, and sets resulting state (`ai:ready-for-review` or `ai:failed`).
7. Worker exits by normal cloud lifecycle rules (`no_work`, `pending_review_limit_reached`, or timeout), then waits for next wake-up.

## GitHub Actions Events and Conditions

- `on.pull_request_review.types: [submitted]` and guard `github.event.review.state == 'changes_requested'`
- `on.issue_comment.types: [created]` and guard:
  - comment is on PR (`github.event.issue.pull_request != null`)
  - body starts with `/ai rework`
  - actor is trusted (`OWNER`, `MEMBER`, `COLLABORATOR`)

## Cloud Run Job Execution Contract

Pass wake-up context as runtime env overrides when executing job:

- `WORKER_RUNTIME_MODE=cloud`
- `TRIGGER_SOURCE=rework`
- `TARGET_PR=<pr_number>`
- `EVENT_ID=<review_or_comment_id>`
- optional: `TARGET_ISSUE=<issue_number>`

Secret delivery boundary:

- `GITHUB_TOKEN` and optional `OPENAI_API_KEY` are provisioned on the Cloud Run Job as Secret Manager-backed env vars.
- The trigger workflow must not send token/key literals via `--update-env-vars`.

Worker requirements:

- Idempotency keyed by `EVENT_ID`
- Deterministic worker lane ownership (`WORKER_ID`)
- Rework updates the existing draft PR branch by default
- Same poll/execution logic as local mode; only lifecycle differs.

## IAM and Security

- GitHub Actions uses OIDC/WIF (`id-token: write`) to impersonate a dedicated service account.
- Service account permissions are least privilege:
  - execute Cloud Run Jobs for mapped worker jobs
  - read required secret references only if needed by trigger workflow
- No static GCP keys in repository.

## Governance Rules

- Rework automation is only allowed for AI-managed PRs.
- Branch protection, CODEOWNERS, and required checks still gate merge.
- Automation may push commits but cannot bypass review/merge policy.

## Operational Notes

- `issue_comment` workflows run only when workflow file is present on default branch.
- Use `EVENT_ID` dedup to avoid duplicate rework runs from retries or duplicate webhooks.
- Add audit comments to PR with run id and result summary.

## Minimal Workflow Skeleton (Reference)

```yaml
name: ai-rework-trigger
on:
  pull_request_review:
    types: [submitted]
  issue_comment:
    types: [created]

permissions:
  id-token: write
  contents: read
  pull-requests: write
  issues: write

jobs:
  trigger-rework:
    if: >
      (github.event_name == 'pull_request_review' && github.event.review.state == 'changes_requested') ||
      (github.event_name == 'issue_comment' && github.event.issue.pull_request &&
       startsWith(github.event.comment.body, '/ai rework'))
    runs-on: ubuntu-latest
    steps:
      - uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: ${{ secrets.GCP_WIF_PROVIDER }}
          service_account: ${{ secrets.GCP_WIF_SERVICE_ACCOUNT }}
      - uses: google-github-actions/setup-gcloud@v2
      - run: |
          gcloud run jobs execute "$AI_WORKER_JOB" \
            --region "$GCP_REGION" \
            --wait \
            --update-env-vars "TRIGGER_SOURCE=rework,TARGET_PR=$PR_NUMBER,EVENT_ID=$EVENT_ID"
```
