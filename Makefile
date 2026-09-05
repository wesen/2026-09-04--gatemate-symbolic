GLAZED_LINT_BIN ?= /tmp/gatemate-glazed-lint
GLAZED_VERSION = $(shell go list -m -f '{{.Version}}' github.com/go-go-golems/glazed)
GOVULNCHECK_BIN ?= /tmp/gatemate-govulncheck

.PHONY: test frontend build lint glazed-lint glazed-lint-build govulncheck dev-backend dev-frontend
test:
	go test ./... -count=1
frontend:
	pnpm --dir web install --frozen-lockfile
	go generate ./internal/microscope ./internal/dataflowide
build: frontend
	go build -tags embed ./...
lint: glazed-lint
	go vet ./...
glazed-lint-build:
	GOBIN=/tmp go install github.com/go-go-golems/glazed/cmd/tools/glazed-lint@$(GLAZED_VERSION)
	cp /tmp/glazed-lint $(GLAZED_LINT_BIN)
glazed-lint: glazed-lint-build
	go vet -vettool=$(GLAZED_LINT_BIN) ./cmd/... ./internal/...
govulncheck:
	GOBIN=/tmp go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
	cp /tmp/govulncheck $(GOVULNCHECK_BIN)
	$(GOVULNCHECK_BIN) ./...
dev-backend:
	go run ./cmd/search-microscope
dev-frontend:
	pnpm --dir web dev

.PHONY: dataflow-frontend dataflow-dev
dataflow-frontend:
	go generate ./internal/dataflowide
dataflow-dev:
	go run ./cmd/dataflow-ide
