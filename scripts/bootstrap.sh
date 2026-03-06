#!/usr/bin/env bash
set -euo pipefail

read_tool_version() {
  local tool="$1"
  awk -v target="$tool" '$1 == target { print $2 }' .tool-versions
}

ensure_command_version() {
  local command="$1"
  local expected="$2"
  local display_name="$3"
  shift 3

  if ! command -v "$command" >/dev/null 2>&1; then
    echo "$display_name is required but not installed. Expected version $expected." >&2
    exit 1
  fi

  local current
  current="$($command "$@" 2>/dev/null | head -n 1)"
  if [[ -z "$current" ]]; then
    echo "Unable to determine $display_name version. Expected $expected." >&2
    exit 1
  fi

  if [[ "$current" != *"$expected"* ]]; then
    echo "$display_name version mismatch. Expected $expected, got: $current" >&2
    exit 1
  fi
}

if command -v mise >/dev/null 2>&1; then
  echo 'Installing pinned tools with mise...'
  mise install
elif command -v asdf >/dev/null 2>&1; then
  echo 'Installing pinned tools with asdf...'
  asdf install
else
  echo 'No supported version manager detected. Validating locally installed tools against .tool-versions...'
fi

if grep -q '^nodejs ' .tool-versions; then
  ensure_command_version node "$(read_tool_version nodejs)" 'Node.js' --version
  ensure_command_version npm '11.8.0' 'npm' --version
fi

if grep -q '^golang ' .tool-versions; then
  ensure_command_version go "$(read_tool_version golang)" 'Go' version
fi

if grep -q '^terraform ' .tool-versions; then
  ensure_command_version terraform "$(read_tool_version terraform)" 'Terraform' version
fi

if grep -q '^buf ' .tool-versions; then
  ensure_command_version buf "$(read_tool_version buf)" 'Buf' --version
fi

if [[ -f package.json ]]; then
  echo 'Installing npm dependencies...'
  npm install
fi

if [[ -f go.mod ]]; then
  echo 'Downloading Go module dependencies...'
  go mod download
fi

echo 'Bootstrap completed.'
