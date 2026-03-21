SHELL := bash

GO_VERSION := 1.25.1
GOLANGCI_LINT_VERSION := v1.64.8
DOCKER_IMAGE := platform-ai-workers:local

.PHONY: help bootstrap doctor install-tools check-tools print-toolchain install-dev-tools precommit-install precommit-run lint format format-check repo-lint repo-format repo-format-check run test build-image container-codex-dir container-codex-login run-container

help:
	@echo "Targets:"
	@echo "  bootstrap         Install toolchain when possible and run baseline setup"
	@echo "  doctor            Run shared workstation checks from sibling platform-blueprint-specs"
	@echo "  install-tools     Install pinned tools with mise/asdf if available"
	@echo "  check-tools       Validate pinned tool versions"
	@echo "  print-toolchain   Print pinned tool versions"
	@echo "  install-dev-tools Install Python development tooling"
	@echo "  precommit-install Install git pre-commit hooks"
	@echo "  precommit-run     Run the configured pre-commit checks on all files"
	@echo "  lint              Run repo lint checks"
	@echo "  format            Apply repo formatting"
	@echo "  format-check      Check repo formatting without writing changes"
	@echo "  run               Start the worker with the current environment"
	@echo "  test              Run Go tests"
	@echo "  build-image       Build the local worker container image"
	@echo "  container-codex-dir   Create the persistent Codex auth directory for Docker runs"
	@echo "  container-codex-login Start a one-off container to log Codex in with device auth"
	@echo "  run-container     Run the worker container with .env.local and the persistent Codex mount"

bootstrap: install-tools check-tools install-dev-tools
	@if find . -name '*.go' -not -path './vendor/*' -not -path './.workspaces/*' | grep -q .; then \
		go mod download; \
	else \
		echo "No Go files yet; skipping go mod download."; \
	fi

doctor:
	@if [[ -f ../platform-blueprint-specs/scripts/windows-tooling-doctor.ps1 ]]; then \
		powershell -ExecutionPolicy Bypass -File ../platform-blueprint-specs/scripts/windows-tooling-doctor.ps1; \
	else \
		echo "Shared doctor script not found at ../platform-blueprint-specs/scripts/windows-tooling-doctor.ps1" >&2; \
		echo "Keep platform-blueprint-specs as a sibling checkout to use make doctor in this workspace." >&2; \
		exit 1; \
	fi

install-tools:
	@if command -v mise >/dev/null 2>&1; then \
		echo "Installing pinned tools with mise..."; \
		mise install; \
	elif command -v asdf >/dev/null 2>&1; then \
		echo "Installing pinned tools with asdf..."; \
		asdf install; \
	else \
		echo "No supported version manager detected. Validating local tools only."; \
	fi

check-tools:
	@actual_go="$$(go version 2>/dev/null || true)"; \
	if [[ -z "$$actual_go" ]]; then \
		echo "Go is required but not installed. Expected $(GO_VERSION)." >&2; \
		exit 1; \
	fi; \
	if [[ "$$actual_go" != *"$(GO_VERSION)"* ]]; then \
		echo "Go version mismatch. Expected $(GO_VERSION), got: $$actual_go" >&2; \
		exit 1; \
	fi

print-toolchain:
	@echo "Go $(GO_VERSION)"
	@echo "golangci-lint $(GOLANGCI_LINT_VERSION)"

install-dev-tools:
	python -m pip install --user -r requirements-dev.txt

precommit-install: install-dev-tools
	python -m pre_commit install

precommit-run:
	python -m pre_commit run --all-files --show-diff-on-failure

lint: repo-lint

format: repo-format

format-check: repo-format-check

run:
	go run ./cmd/worker run

test:
	go test ./...

build-image:
	docker build -t $(DOCKER_IMAGE) .

container-codex-dir:
	@docker_codex_dir="$${HOME:-$${USERPROFILE}}/.docker-platform-ai-workers/codex"; \
		mkdir -p "$$docker_codex_dir"; \
		echo "Persistent Codex Docker state directory: $$docker_codex_dir"

container-codex-login: build-image container-codex-dir
	@docker_codex_dir="$${HOME:-$${USERPROFILE}}/.docker-platform-ai-workers/codex"; \
		MSYS_NO_PATHCONV=1 MSYS2_ARG_CONV_EXCL='*' docker run --rm -it \
			-e HOME=/root \
			-e CODEX_HOME=/root/.codex \
			-v "$$docker_codex_dir:/root/.codex" \
			--entrypoint bash $(DOCKER_IMAGE) \
			-lc "codex login --device-auth && codex login status"

run-container: build-image container-codex-dir
	@test -f .env.local || (echo ".env.local is required. Copy .env.example and set GITHUB_TOKEN before running the container." >&2; exit 1)
	@mkdir -p .workspaces; \
		docker_codex_dir="$${HOME:-$${USERPROFILE}}/.docker-platform-ai-workers/codex"; \
		MSYS_NO_PATHCONV=1 MSYS2_ARG_CONV_EXCL='*' docker run --rm \
			-e HOME=/root \
			-e CODEX_HOME=/root/.codex \
			--env-file .env.local \
			-v "$$docker_codex_dir:/root/.codex" \
			-v "$$(pwd)/.workspaces:/app/.workspaces" \
			$(DOCKER_IMAGE) run

repo-lint:
	@if find . -name '*.go' -not -path './vendor/*' -not -path './.workspaces/*' | grep -q .; then \
		go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...; \
	else \
		echo "No Go files yet; skipping Go lint."; \
	fi

repo-format:
	@if find . -name '*.go' -not -path './vendor/*' -not -path './.workspaces/*' | grep -q .; then \
		find . -name '*.go' -not -path './vendor/*' -not -path './.workspaces/*' -print0 | xargs -0 gofmt -w; \
	else \
		echo "No Go files yet; skipping Go format."; \
	fi

repo-format-check:
	@if find . -name '*.go' -not -path './vendor/*' -not -path './.workspaces/*' | grep -q .; then \
		unformatted="$$(find . -name '*.go' -not -path './vendor/*' -not -path './.workspaces/*' -print0 | xargs -0 gofmt -l)"; \
		if [[ -n "$$unformatted" ]]; then \
			echo "The following Go files need formatting:"; \
			echo "$$unformatted"; \
			exit 1; \
		fi; \
	else \
		echo "No Go files yet; skipping Go format check."; \
	fi
