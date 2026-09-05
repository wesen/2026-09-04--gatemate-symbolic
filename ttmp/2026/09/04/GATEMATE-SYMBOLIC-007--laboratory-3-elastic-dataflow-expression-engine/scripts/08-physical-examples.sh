#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
ticket=ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine
for example in book copy fault cancel; do
 go run ./cmd/dataflow-lab --engine serial --device /dev/ttyACM0 --example "$example" --format json --wire-log "$ticket/reference/validation/P4-$example-wire.log" > "$ticket/reference/validation/P4-$example-physical.json"
done
printf 'PASS all four physical examples\n'
