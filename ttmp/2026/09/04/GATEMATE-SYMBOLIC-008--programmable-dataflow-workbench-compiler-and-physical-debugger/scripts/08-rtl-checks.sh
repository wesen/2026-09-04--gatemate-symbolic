#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
source /home/manuel/fpga/oss-cad-suite/environment
elastic_dataflow/scripts/test.sh
elastic_dataflow/scripts/test-link.sh
elastic_dataflow/scripts/test-programmable.sh
