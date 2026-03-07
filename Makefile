SHELL := bash

GO_VERSION := 1.24.12
GOLANGCI_LINT_VERSION := v1.64.8

.PHONY: help bootstrap install-tools check-tools print-toolchain install-dev-tools precommit-install precommit-run lint format format-check repo-lint repo-format repo-format-check

help:
	@echo "Targets:"
	@echo "  bootstrap         Install toolchain when possible and run baseline setup"
	@echo "  install-tools     Install pinned tools with mise/asdf if available"
	@echo "  check-tools       Validate pinned tool versions"
	@echo "  print-toolchain   Print pinned tool versions"
	@echo "  install-dev-tools Install Python development tooling"
	@echo "  precommit-install Install git pre-commit hooks"
	@echo "  precommit-run     Run the configured pre-commit checks on all files"
	@echo "  lint              Run repo lint checks"
	@echo "  format            Apply repo formatting"
	@echo "  format-check      Check repo formatting without writing changes"

bootstrap: install-tools check-tools install-dev-tools
	@if find . -name '*.go' -not -path './vendor/*' | grep -q .; then \
		go mod download; \
	else \
		echo "No Go files yet; skipping go mod download."; \
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

repo-lint:
	@if find . -name '*.go' -not -path './vendor/*' | grep -q .; then \
		unformatted="$$(find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -l)"; \
		if [[ -n "$$unformatted" ]]; then \
			echo "The following Go files need formatting:"; \
			echo "$$unformatted"; \
			exit 1; \
		fi; \
		go vet ./...; \
		go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...; \
	else \
		echo "No Go files yet; skipping Go lint."; \
	fi

repo-format:
	@if find . -name '*.go' -not -path './vendor/*' | grep -q .; then \
		find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -w; \
	else \
		echo "No Go files yet; skipping Go format."; \
	fi

repo-format-check:
	@if find . -name '*.go' -not -path './vendor/*' | grep -q .; then \
		unformatted="$$(find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -l)"; \
		if [[ -n "$$unformatted" ]]; then \
			echo "The following Go files need formatting:"; \
			echo "$$unformatted"; \
			exit 1; \
		fi; \
	else \
		echo "No Go files yet; skipping Go format check."; \
	fi
