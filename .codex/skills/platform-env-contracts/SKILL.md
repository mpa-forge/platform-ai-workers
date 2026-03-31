---
name: platform-env-contracts
description: Apply the platform blueprint environment variable strategy. Use when creating or updating .env.example files, documenting required or optional variables, deciding naming/default/secret rules, or adding startup validation for environment-backed configuration across repos.
---

# Platform Environment Contracts

Use this skill when work touches environment-variable naming, documentation, or validation contracts.

## Core Rule

Repo-local `.env.example` files are the source of truth for concrete variable lists.
This skill defines the shared cross-repo rules for how those contracts should be shaped.

## Default Workflow

1. Update the repo-local `.env.example` first.
2. Separate required and optional variables clearly.
3. Keep placeholders for secrets; never commit real values.
4. Promote required runtime variables into typed startup validation when the service is runnable.
5. Reflect cross-repo patterns back into shared planning docs only when the rule is genuinely shared.

## File Policy

- commit `.env.example`
- do not commit `.env`
- do not commit `.env.local`
- use local env files only for developer-specific values

## Naming Rules

- shared service variables:
  - all-caps snake case
- frontend browser-exposed variables:
  - `VITE_*`
- durations:
  - Go-style duration strings like `30s`, `5m`
- booleans:
  - lowercase `true` or `false`

## Design Rules

- defaults are acceptable only for low-risk local convenience
- secrets must always stay as placeholders in committed examples
- required variables should become typed startup config and fail-fast validation once runtime startup exists
- mutually dependent variables should be documented together

## Repo Baseline

Use the current repo-local `.env.example` and runtime docs for the authoritative variable set.
Use this skill to keep the shape of those contracts consistent across repos.
