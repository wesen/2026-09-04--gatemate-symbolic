#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
export GOCACHE=/tmp/gatemate008-go-cache
ticket=ttmp/2026/09/04/GATEMATE-SYMBOLIC-008--programmable-dataflow-workbench-compiler-and-physical-debugger
check() { name="$1"; shift; "$@" > "$ticket/reference/validation/p6-$name.log" 2>&1; printf 'PASS %s\n' "$name"; }
check go-race go test -race ./... -count=1
check go-build go build ./...
check go-embed-build go build -tags embed ./...
check go-vet go vet ./...
check glazed-lint make glazed-lint
check govulncheck make govulncheck
check frontend-types pnpm --dir web check
check frontend-tests pnpm --dir web test
