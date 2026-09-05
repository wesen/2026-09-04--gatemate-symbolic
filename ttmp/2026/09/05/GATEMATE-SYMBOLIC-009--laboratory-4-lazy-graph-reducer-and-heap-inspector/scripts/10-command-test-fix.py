#!/usr/bin/env python3
from pathlib import Path
repo=Path(__file__).resolve().parents[6]
s=(repo/'cmd/dataflow-lab/main_test.go').read_text().replace('"book", "copy", "fault", "cancel"','"shared", "cycle", "indirection", "indirection-cycle", "overflow"')
(repo/'cmd/lazy-lab/main_test.go').write_text(s)
