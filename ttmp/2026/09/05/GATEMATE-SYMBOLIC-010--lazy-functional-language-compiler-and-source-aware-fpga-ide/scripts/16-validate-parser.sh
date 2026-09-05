#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
ticket=ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide
export GOCACHE=/tmp/gatemate009-go-cache
case "${1:?use fuzz or repository}" in
 fuzz) go test ./pkg/lazylang/syntax -run '^$' -fuzz '^FuzzParse$' -fuzztime 15s -parallel 2 > "$ticket/reference/validation/parser-fuzz.log" 2>&1 ;;
 repository)
  go test ./... -count=1 > "$ticket/reference/validation/parser-repository-tests.log" 2>&1
  go test -race ./pkg/lazylang/syntax -count=1 > "$ticket/reference/validation/parser-race.log" 2>&1
  go vet ./... > "$ticket/reference/validation/parser-vet.log" 2>&1
  go build ./... > "$ticket/reference/validation/parser-build.log" 2>&1
  ;;
 *) exit 2;;
esac
