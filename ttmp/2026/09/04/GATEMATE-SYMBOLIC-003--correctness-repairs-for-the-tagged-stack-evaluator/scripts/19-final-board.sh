#!/usr/bin/env bash
set -euo pipefail
ticket_scripts="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
bash "$ticket_scripts/09-build.sh"
bash "$ticket_scripts/11-hardware.sh"
