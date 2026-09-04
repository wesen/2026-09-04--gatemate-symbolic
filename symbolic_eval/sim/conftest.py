"""Make tools/ importable from the pytest suites (MATE-16 pattern)."""

import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "tools"))


def pytest_sessionfinish(session, exitstatus):
    import json
    from pathlib import Path
    from state_checks import COVERAGE
    out = Path(__file__).resolve().parents[1] / 'build' / 'verification-coverage.json'
    out.parent.mkdir(exist_ok=True)
    out.write_text(json.dumps(COVERAGE, indent=2, sort_keys=True) + '\n')
