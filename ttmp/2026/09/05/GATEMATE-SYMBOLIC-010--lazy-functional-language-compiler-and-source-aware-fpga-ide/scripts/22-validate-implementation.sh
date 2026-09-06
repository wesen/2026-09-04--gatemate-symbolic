#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
ticket=ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide
export GOCACHE=/tmp/gatemate009-go-cache
go generate ./internal/lazylanguageide > "$ticket/reference/validation/final-frontend-build.log" 2>&1
go test ./... -count=1 > "$ticket/reference/validation/final-go-tests.log" 2>&1
go test -race ./pkg/lazylang/... ./internal/lazylanguageide > "$ticket/reference/validation/final-go-race.log" 2>&1
go build ./... > "$ticket/reference/validation/final-go-build.log" 2>&1
go build -tags embed ./... > "$ticket/reference/validation/final-embedded-build.log" 2>&1
go vet ./... > "$ticket/reference/validation/final-go-vet.log" 2>&1
pnpm --dir web test > "$ticket/reference/validation/final-frontend-tests.log" 2>&1
go run ./cmd/lazy-language --source examples/lazylang/shared.lazy --action reference > "$ticket/reference/validation/cli-shared-reference.json"
go run ./cmd/lazy-language --source examples/lazylang/squares.lazy --action run > "$ticket/reference/validation/cli-squares-model.json"
go run ./cmd/lazy-language --help > "$ticket/reference/validation/cli-help.txt"
go run ./cmd/lazy-language-ide --help > "$ticket/reference/validation/ide-help.txt"
git diff --check
