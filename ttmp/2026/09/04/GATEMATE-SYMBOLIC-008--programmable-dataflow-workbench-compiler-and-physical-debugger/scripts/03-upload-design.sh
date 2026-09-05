#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
remarquee upload bundle design-doc/01-programmable-dataflow-workbench-intern-analysis-design-and-implementation-guide.md --name 'GATEMATE 008 Programmable Dataflow Workbench Design' --remote-dir /ai/2026/09/04/GATEMATE-SYMBOLIC-008 --toc-depth 2 --non-interactive > reference/validation/p1-upload.log 2>&1
