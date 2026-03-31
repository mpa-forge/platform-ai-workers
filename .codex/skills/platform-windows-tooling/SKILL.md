---
name: platform-windows-tooling
description: Apply the platform blueprint Windows workstation tooling baseline. Use when setting up or diagnosing a Windows developer machine, validating shell/tool availability, fixing PATH issues, or troubleshooting make/bash/python/gh/docker/gcloud problems in this workspace.
---

# Platform Windows Tooling

Use this skill when a Windows workstation needs to be validated or repaired for this workspace.

## Goal

Avoid shell and tool drift that breaks bootstrap, validation, or automation behavior.

## Required Baseline

- Git for Windows with `bash.exe` on `PATH`
- GNU Make with POSIX shell compatibility
- Python where `python` resolves to a real interpreter
- `gh`
- Docker Desktop or equivalent Docker engine

Required for cloud and deployment work:

- `gcloud`

Strongly recommended:

- `mise`
- `rg`

## Unsupported

- `GnuWin32` `make`
- Windows Store `python` aliases that do not resolve to a real Python install

## PATH Expectations

The shell should resolve:

- `bash`
- `make`
- `python`
- `gh`
- `docker`

And, when installed:

- `gcloud`
- `rg`
- `go`
- `node`
- `npm`
- `terraform`
- `buf`

## Default Workflow

1. Check tool resolution from the active shell.
2. Verify `make` is a GNU Make build compatible with Git Bash.
3. Verify `python` resolves to a real interpreter, not a store alias.
4. Verify Docker and `gh` are usable from the same shell environment.
5. Run the shared doctor script for workspace validation.

## Doctor Script

Use:

- `scripts/windows-tooling-doctor.ps1` from `platform-blueprint-specs`

Prefer the doctor script when you need a repeatable check instead of one-off manual probing.
