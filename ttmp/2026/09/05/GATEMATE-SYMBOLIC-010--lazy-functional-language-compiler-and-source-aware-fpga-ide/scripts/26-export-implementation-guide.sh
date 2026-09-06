#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
case "${1:?use pdf or upload}" in
 pdf)
  remarquee upload bundle sources/implementation-guide.md --name 'GATEMATE 010 Implemented Language and IDE Guide' --toc-depth 2 --mermaid=false --pdf-only --output-dir reference/validation --non-interactive > reference/validation/implementation-pdf-render.log 2>&1
  ;;
 upload)
  remarquee upload bundle sources/implementation-guide.md --name 'GATEMATE 010 Implemented Language and IDE Guide' --remote-dir /ai/2026/09/05/GATEMATE-SYMBOLIC-010 --toc-depth 2 --mermaid=false --non-interactive > reference/validation/implementation-remarkable-upload.log 2>&1
  ;;
 *) exit 2;;
esac
