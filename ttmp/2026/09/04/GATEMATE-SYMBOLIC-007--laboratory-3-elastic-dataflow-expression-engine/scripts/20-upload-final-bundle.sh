#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
ticket=ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine
remarquee upload bundle \
 "$ticket/design-doc/01-elastic-dataflow-engine-intern-analysis-design-and-implementation-guide.md" \
 "$ticket/design-doc/02-dataflow-ide-go-react-architecture-and-intern-implementation-guide.md" \
 "$ticket/reference/02-dataflow-api-and-debug-register-reference.md" \
 "$ticket/reference/03-implemented-laboratory-handoff-and-screenshot-atlas.md" \
 --name 'GATEMATE 007 Implemented Dataflow Lab and IDE' \
 --remote-dir /ai/2026/09/04/GATEMATE-SYMBOLIC-007 --toc-depth 2 --non-interactive "$@"
