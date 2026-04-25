# Environment Variables

This document is the authoritative reference for the worker runtime environment.

Use it together with [`.env.example`](../.env.example) when creating `.env.local`.

## How to read this reference

- `Required`: whether startup fails if the variable is missing
- `Default`: the value used when the variable is unset
- `Current usage`: whether the variable affects runtime behavior today or is mainly reserved context

## Core runtime

| Variable | Required | Default | Meaning | Current usage |
| --- | --- | --- | --- | --- |
| `APP_ENV` | No | `local` | Free-form environment name for the runtime, such as `local` or `cloud`. | Parsed into config but not currently used for branching behavior. |
| `LOG_LEVEL` | No | `info` | Intended log verbosity setting. | Parsed into config but not currently used to change logging output. |
| `WORKER_RUNTIME_MODE` | No | `local` | Runtime lifecycle mode. Valid values: `local`, `cloud`. | Actively used. `local` keeps polling after idle or backpressure; `cloud` exits on terminal outcomes such as `no_work` and `pending_review_limit_reached`. |
| `WORKER_ID` | Yes | None | Lane identifier for this worker, for example `worker-01`. | Actively used for lane selection, issue labels (`worker:<id>`), branch naming, lock ownership, and generated `RUN_ID` values. |
| `TARGET_REPO` | Yes | None | GitHub repository the worker controls, in `owner/repo` form. | Actively used for GitHub issue and PR operations and for workspace preparation. |
| `MAX_PENDING_REVIEW` | No | `3` | Maximum number of open `ai:ready-for-review` issues allowed for the lane before the worker stops or waits. Must be a positive integer. | Actively used in the poll loop to apply backpressure. |
| `POLL_INTERVAL` | No | `30s` | Sleep duration between loop iterations in `local` mode. Must be a positive Go duration string. | Actively used in local polling mode. |
| `BASE_BRANCH` | No | `main` | Base branch the reusable workspace is reset to before work starts. | Actively used for workspace preparation and lock metadata. |
| `WORKSPACE_ROOT` | No | `.workspaces` | Directory under which the worker keeps its reusable clone. | Actively used when preparing the target repo workspace. |
| `DRY_RUN` | No | `false` | When `true`, the worker claims the issue and renders the prompt but skips Codex execution and PR verification. | Actively used for debugging the control plane without running the coding agent. |

## Agent execution

| Variable | Required | Default | Meaning | Current usage |
| --- | --- | --- | --- | --- |
| `AGENT_CLI` | No | `codex` | Executable name or path for the coding agent CLI. | Actively used when spawning the agent subprocess. |
| `AGENT_AUTH_MODE` | No | `chatgpt` | Authentication mode for the agent. Valid values: `chatgpt`, `api`. | Actively used. `chatgpt` relies on an existing Codex login; `api` injects `OPENAI_API_KEY` into the agent process. |
| `AGENT_MODEL` | No | empty | Optional model override passed to the agent CLI. | Actively used only when set; otherwise the agent uses its default configured model. |
| `PROMPT_TEMPLATE` | No | `prompts/task.md.tmpl` | Template path used to render the task prompt given to Codex. | Actively used when generating the prompt file for the agent. |
| `OPENAI_API_KEY` | Conditionally | None | OpenAI API key for `AGENT_AUTH_MODE=api`. | Required only in `api` mode. Ignored in `chatgpt` mode. For Cloud Run Jobs, inject from the Phase 5 GSM secret catalog instead of plaintext env values. |

## GitHub and trigger context

| Variable | Required | Default | Meaning | Current usage |
| --- | --- | --- | --- | --- |
| `GITHUB_TOKEN` | Yes | None | GitHub token used by both the worker control plane and the agent subprocess for issue, PR, and git operations. | Actively used everywhere GitHub access is required. Startup fails without it. For Cloud Run Jobs, inject from the Phase 5 GSM secret catalog instead of plaintext env values. |
| `TRIGGER_SOURCE` | No | `manual` | Intended source label for the current run, such as `manual`, `event`, or `scheduled`. | Parsed into config but not currently used for branching behavior. |
| `TARGET_ISSUE` | No | empty | Optional issue number for a targeted run. Must be a positive integer when set. | Actively used. If set, the worker tries that issue before normal queue selection, but only if the issue is eligible for the current lane. |
| `TARGET_PR` | No | empty | Optional PR number for a targeted or event-driven run. Must be a positive integer when set. | Parsed and validated, but not currently consumed by the worker flow. Reserved for future trigger-specific behavior. |
| `EVENT_ID` | No | empty | Unique identifier for an event-triggered run. | Actively used for deduplication and machine-readable issue comments such as `automation-event-id:<id>`. Leave empty for normal local polling runs. |

## Run identity and locking

| Variable | Required | Default | Meaning | Current usage |
| --- | --- | --- | --- | --- |
| `RUN_ID` | No | auto-generated | Optional unique identifier for a worker run. If unset, the worker generates `<worker-id>-<unix-nanoseconds>`. | Actively used in remote lock metadata and automation marker comments for traceability. |
| `LOCK_STALE_AFTER` | No | `15m` | Staleness threshold for the remote lane lock branch. Must be a positive Go duration string. | Actively used when deciding whether an existing lock can be reclaimed after a crash or timeout. |

## Notes by workflow

### Cloud Run Jobs (Phase 5 baseline)

Secret contract for cloud execution:

- `GITHUB_TOKEN` is required and must be delivered as a Secret Manager-backed env var.
- `OPENAI_API_KEY` is optional and only delivered when `AGENT_AUTH_MODE=api`.
- Keep secret ownership and names in the Phase 5 catalog managed by `platform-infra`; this repo only consumes runtime env var values.
- Do not pass token or key literals through Terraform env values, checked-in manifests, or `gcloud run jobs execute --update-env-vars`.

### Normal local polling

Typical local runs need:

- `WORKER_ID`
- `TARGET_REPO`
- `GITHUB_TOKEN`
- optional local overrides such as `POLL_INTERVAL`

Leave these blank unless you are debugging a specific scenario:

- `RUN_ID`
- `TARGET_ISSUE`
- `TARGET_PR`
- `EVENT_ID`

### Local container run with ChatGPT account auth

Recommended settings:

- `AGENT_AUTH_MODE=chatgpt`
- `GITHUB_TOKEN=<real token>`
- `OPENAI_API_KEY=` (leave blank)

In this mode, GitHub auth still comes from `GITHUB_TOKEN`, while Codex auth comes from the persisted `/root/.codex` mount described in [worker-runtime.md](./worker-runtime.md).

### Event-driven or cloud-style targeted runs

These variables are mainly useful for triggered reruns and debug sessions:

- `TARGET_ISSUE`
- `TARGET_PR`
- `EVENT_ID`
- `RUN_ID`

`EVENT_ID` is the most important of the four for deduping repeated webhook-driven retries.
