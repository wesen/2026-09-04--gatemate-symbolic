#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
remarquee upload bundle design-doc/01-lazy-graph-reducer-intern-analysis-design-and-implementation-guide.md reference/02-implemented-reducer-api-and-qualification-handoff.md --name 'GATEMATE 009 Implemented Lazy Reducer and Inspector' --remote-dir /ai/2026/09/05/GATEMATE-SYMBOLIC-009 --toc-depth 2 --non-interactive > reference/validation/p5-handoff-upload.log 2>&1
