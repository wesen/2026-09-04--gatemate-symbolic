#!/usr/bin/env bash
set -euo pipefail
source /home/manuel/fpga/oss-cad-suite/environment
ticket="$(cd "$(dirname "$0")/.." && pwd)"
repo="$(git -C "$ticket" rev-parse --show-toplevel)"
out="$ticket/reference/validation"
trap 'printf "%s\n" "$?" > "$out/P6-checks.exit"' EXIT
cd "$repo"
go test -race ./... -count=1 > "$out/P6-go-tests.log" 2>&1
pnpm --dir web test > "$out/P6-web-tests.log" 2>&1
go generate ./internal/microscope > "$out/P6-web-build.log" 2>&1
go build ./... > "$out/P6-default-build.log" 2>&1
go build -tags embed ./... > "$out/P6-embed-build.log" 2>&1
make lint > "$out/P6-lint.log" 2>&1
make govulncheck > "$out/P6-vulnerabilities.log" 2>&1
make -C queens_rollback test > "$out/P6-queens-tests.log" 2>&1
