#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
defuddle parse 'https://www.microsoft.com/en-us/research/publication/implementing-lazy-functional-languages-on-stock-hardware-the-spineless-tagless-g-machine/' --md -o sources/stg-publication.md
curl -fL 'https://www.microsoft.com/en-us/research/wp-content/uploads/1992/04/spineless-tagless-gmachine.pdf' -o sources/spineless-tagless-gmachine.pdf
pdftotext -layout sources/spineless-tagless-gmachine.pdf sources/spineless-tagless-gmachine.txt
defuddle parse 'https://arxiv.org/abs/0907.4640' --md -o sources/call-by-need-semantics.md
