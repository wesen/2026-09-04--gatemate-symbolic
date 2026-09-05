#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
ticket=ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine
check() {
 name="$1"; shift
 "$@" > "$ticket/reference/validation/P7-$name.log" 2>&1
 printf 'PASS %s\n' "$name"
}
check generate go generate ./internal/microscope ./internal/dataflowide
check go-race go test -race ./... -count=1
check go-build go build ./...
check go-embed-build go build -tags embed ./...
check go-vet go vet ./...
check glazed-lint make glazed-lint
check govulncheck make govulncheck
check frontend-tests pnpm --dir web test
check frontend-types pnpm --dir web check
check whitespace git diff --check
