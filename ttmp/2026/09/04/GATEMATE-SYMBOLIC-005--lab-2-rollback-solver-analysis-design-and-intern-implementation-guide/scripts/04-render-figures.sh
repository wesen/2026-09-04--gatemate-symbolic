#!/usr/bin/env bash
set -euo pipefail
ticket_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "$ticket_dir/design-doc/assets"
dot -Tpng -Gdpi=150 "$ticket_dir/scripts/02-architecture.dot" -o "$ticket_dir/design-doc/assets/architecture.png"
dot -Tpng -Gdpi=150 "$ticket_dir/scripts/03-rollback.dot" -o "$ticket_dir/design-doc/assets/rollback.png"
