"""Make tools/ importable from the pytest suites (MATE-16 pattern)."""

import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "tools"))
