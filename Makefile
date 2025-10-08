GOBIN ?= $(shell go env GOPATH)/bin
LINTER := $(GOBIN)/golangci-lint

.PHONY: lint lint-fix install-linter

lint:
	@command -v $(LINTER) >/dev/null 2>&1 || { \
		echo "golangci-lint not found. Run 'make install-linter' or install via Homebrew."; \
		exit 1; \
	}
	@$(LINTER) run ./...

lint-fix:
	@command -v $(LINTER) >/dev/null 2>&1 || { \
		echo "golangci-lint not found. Run 'make install-linter' or install via Homebrew."; \
		exit 1; \
	}
	@$(LINTER) run --fix ./...

install-linter:
	@echo "Installing golangci-lint to $(GOBIN)"
	@GOBIN=$(GOBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installed at: $(LINTER)"
	@echo "If you want to call it without Makefile, add $(GOBIN) to your PATH."
