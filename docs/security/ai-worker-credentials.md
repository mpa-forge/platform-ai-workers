# AI Worker Credentials and Secret Model (P0-T13)

## Purpose

Define the credential and secret model for AI worker lanes in local and cloud runtimes.

## Baseline Decision

- Primary GitHub credential model for baseline implementation: fine-grained PATs
- One or more fine-grained PATs may exist under a user account
- PATs must be scoped only to the repositories a worker lane needs
- GitHub App remains the planned migration path for a later hardening phase

## GitHub Credential Model

### Baseline

- Use fine-grained PATs stored in Google Secret Manager
- PATs are repository-scoped, not org-wide broad tokens
- A worker lane must only use the PAT that corresponds to its target repository or approved repo set

### Why this baseline

- Faster to bootstrap than GitHub App
- Sufficient for single-maintainer baseline implementation
- Keeps credential scope narrower than classic PATs

### Later migration

- Migrate to GitHub App when automation surface area grows and stronger automation identity boundaries are needed

## Required GitHub Permissions

Fine-grained PATs should include only the minimum required repository permissions:

- Issues: read/write
- Pull requests: read/write
- Contents: read/write
- Metadata: read

No admin or organization-management permissions are required for worker execution.

## GCP Authentication Model

- GitHub Actions trigger workflows authenticate to GCP using Workload Identity Federation
- No static GCP service account keys in repositories
- Trigger workflow identity is separate from worker runtime identity

## Runtime Identities

### Trigger workflow identity

- Used by GitHub Actions to execute Cloud Run Jobs
- Permissions limited to:
  - execute mapped Cloud Run Job
  - read only the minimum metadata/secrets needed by the trigger workflow

### Worker runtime identity

- Cloud Run Job service account per workload class or per lane where practical
- Permissions limited to:
  - read only its mapped GSM secrets
  - write telemetry/logs as required
  - access only required runtime dependencies

## Secret Source of Truth

- Google Secret Manager is the source of truth for worker credentials and runtime secrets

## Secret Layout

Recommended secret layout per environment:

- GitHub PAT:
  - `ai-worker-github-pat-<worker-id>-rc`
  - `ai-worker-github-pat-<worker-id>-prod`
- Agent/provider secret:
  - `ai-worker-agent-key-<worker-id>-rc`
  - `ai-worker-agent-key-<worker-id>-prod`

Cloud Run runtime env mapping contract:

- `GITHUB_TOKEN` <- mapped GSM GitHub PAT secret (required)
- `OPENAI_API_KEY` <- mapped GSM provider key secret (optional; only for `AGENT_AUTH_MODE=api`)

Rules:

- separate secret per environment
- separate secret per worker lane where repo scope differs
- no plaintext credentials in git

## Local vs Cloud Secret Delivery

### Cloud

- Cloud Run Jobs receive secret values from GSM
- Secret references are injected through runtime configuration
- Trigger workflows should only pass non-secret run context (`TARGET_PR`, `EVENT_ID`, and similar) at execution time

### Local

- Local worker uses developer-provided environment file or shell env vars
- Local values must never be committed
- Naming should match cloud secret intent to reduce drift

## Least-Privilege Boundaries

- A worker may only use credentials mapped to its own `worker:<id>`
- A worker PAT must not grant access to unrelated repositories
- Trigger workflow credentials must not be reused as worker execution credentials

## Rotation and Ownership

- Credential owner: human maintainer
- PAT rotation is manual in the baseline phase
- Rotation process:
  1. create replacement PAT
  2. store new secret version in GSM
  3. redeploy or restart affected worker lane
  4. revoke old PAT

## Non-Goals

- Broad org-scoped classic PATs
- Static GCP keys in repo or CI
- Shared single PAT for all worker lanes when repo scopes differ
