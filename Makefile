# Makefile at project root

GOLANGCI_LINT_VERSION := v2.0.2
GOLANGCI_LINT_BIN := ./bin/golangci-lint

.PHONY: lint install-linter tidy test build

install-linter:
	@echo "Installing golangci-lint..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b ./bin $(GOLANGCI_LINT_VERSION)

lint: install-linter
	@echo "Running golangci-lint..."
	$(GOLANGCI_LINT_BIN) run ./...

tidy:
	go mod tidy

test:
	go test ./...

format:
	@echo "Running golangci-lint with auto-fix (formatting)..."
	$(GOLANGCI_LINT_BIN) fmt ./...
