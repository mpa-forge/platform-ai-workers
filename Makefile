SHELL := bash

GO_VERSION := 1.24.12

.PHONY: help bootstrap install-tools check-tools print-toolchain

help:
	@echo "Targets:"
	@echo "  bootstrap       Install toolchain when possible and run baseline setup"
	@echo "  install-tools    Install pinned tools with mise/asdf if available"
	@echo "  check-tools      Validate pinned tool versions"
	@echo "  print-toolchain  Print pinned tool versions"

bootstrap: install-tools check-tools
	go mod download

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
