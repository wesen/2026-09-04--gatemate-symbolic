"""Differential RTL tests: every program runs through the reference model
AND the RTL (iverilog), and the TRACE/FINAL lines must match exactly.

Book verification plan (Laboratory 1): directed tests 1-8 plus output stalls.
The random-stall runs prove that backpressure changes timing but never the
commit trace (Delayed Irreversible Store + ready/valid ownership).
"""

import os
import subprocess
import sys

import pytest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "tools"))

from asm20 import assemble  # noqa: E402
from stack_model import run_program  # noqa: E402

ROOT = os.path.join(os.path.dirname(__file__), "..")
BUILD = os.path.join(ROOT, "build")

PROGRAMS = [
    # (name, stack_depth, stall_seeds to also run)
    ("arith", 32, []),
    ("typefault", 32, []),
    ("smoke", 32, [7, 123]),
    ("deep", 32, [3, 99]),
    ("branch", 32, [5]),
    ("muloverflow", 32, []),
    ("underflow", 32, []),
    ("badop", 32, []),
    ("badbranch", 32, []),
    ("jztype", 32, []),
    ("stackoverflow", 8, []),
    ("fib", 32, [11]),
    ("sq", 32, []),
    ("countdown", 32, [4]),
    ("retunderflow", 32, []),
    ("calloverflow", 32, []),
]


def _ensure_hex(name: str) -> str:
    os.makedirs(BUILD, exist_ok=True)
    hex_path = os.path.join(BUILD, f"{name}.hex")
    with open(os.path.join(ROOT, "programs", f"{name}.asm")) as f:
        words, _, _ = assemble(f.read())
    with open(hex_path, "w") as f:
        for w in words:
            f.write(f"{w:05x}\n")
        for _ in range(1024 - len(words)):
            f.write("00000\n")
    return hex_path


def _run_rtl(hex_path: str, stack_depth: int, stall_seed=None):
    """Compile (per stack-depth config) and run the register-core testbench."""
    tag = f"tb_reg_d{stack_depth}"
    vvp = os.path.join(BUILD, f"{tag}.vvp")
    compile_cmd = [
        "iverilog", "-g2012", "-s", "tb_stack_core", "-o", vvp,
        os.path.join(ROOT, "rtl", "symbolic_types_pkg.sv"),
        os.path.join(ROOT, "rtl", "stack_core.sv"),
        os.path.join(ROOT, "sim", "tb_stack_core.sv"),
        f"-Ptb_stack_core.STACK_DEPTH={stack_depth}",
    ]
    subprocess.run(compile_cmd, check=True, capture_output=True)
    cmd = ["vvp", vvp, f"+rom={hex_path}"]
    if stall_seed is not None:
        cmd.append(f"+stall_seed={stall_seed}")
    out = subprocess.run(cmd, check=True, capture_output=True, text=True,
                          timeout=120)
    lines = [l for l in out.stdout.splitlines()
             if l.startswith(("TRACE", "FINAL", "TB_", "FAIL",
                             "ASSERT_FAIL"))]
    tb_pass = any(l.startswith("TB_PASS") for l in lines)
    assert tb_pass, f"testbench reported errors:\n{out.stdout}"
    return [l for l in lines if l.startswith(("TRACE", "FINAL"))]


def _model_lines(name: str, stack_depth: int):
    with open(os.path.join(ROOT, "programs", f"{name}.asm")) as f:
        words, _, _ = assemble(f.read())
    m = run_program(words, total_depth=stack_depth)
    return [r.line() for r in m.trace] + [m.final_line()]


@pytest.mark.parametrize("name,depth,seeds", PROGRAMS)
def test_directed(name, depth, seeds):
    hex_path = _ensure_hex(name)
    expected = _model_lines(name, depth)
    got = _run_rtl(hex_path, depth)
    assert got == expected, _diff(got, expected)

    for seed in seeds:
        got_s = _run_rtl(hex_path, depth, stall_seed=seed)
        assert got_s == expected, f"stall_seed={seed} changed the trace"


def _diff(got, expected):
    import difflib
    return "\n".join(difflib.unified_diff(
        expected, got, "model", "rtl", lineterm=""))


def test_exit_criteria_book():
    """Book hardware evidence: one BOOL(true) output then HALTED; and the
    bad program faults precisely with no output."""
    hex_a = _ensure_hex("arith")
    rtl = _run_rtl(hex_a, 32)
    assert rtl[-3] == "TRACE 7 7 8 EMIT 0 OUTPUT 1 00000001"
    assert rtl[-2] == "TRACE 8 8 8 HALT 0 COMMIT"
    assert rtl[-1] == "FINAL 1 NONE 8 0 1 0"
    rtl_b = _run_rtl(_ensure_hex("typefault"), 32)
    assert rtl_b[-1] == "FINAL 0 TYPE_FAULT 2 2 0 0"
