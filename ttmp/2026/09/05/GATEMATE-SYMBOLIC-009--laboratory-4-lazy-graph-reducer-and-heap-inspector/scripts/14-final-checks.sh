#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
ticket=ttmp/2026/09/05/GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector
export GOCACHE=/tmp/gatemate009-go-cache
: > "$ticket/reference/validation/p4-check-summary.log"
check(){ local name=$1;shift;if "$@" > "$ticket/reference/validation/p4-$name.log" 2>&1;then echo "PASS $name" >> "$ticket/reference/validation/p4-check-summary.log";else echo "FAIL $name" >> "$ticket/reference/validation/p4-check-summary.log";return 1;fi; }
check go-race go test -race ./... -count=1
check go-build go build ./...
check go-embed go build -tags embed ./...
check go-vet go vet ./...
check glazed-lint make glazed-lint
check govulncheck make govulncheck
check frontend-types pnpm --dir web check
check frontend-tests pnpm --dir web test
