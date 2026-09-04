"""Differential RTL tests for the BRAM core (two-entry top cache over
sync_sdp_ram). Same programs as the register core, plus the refinement
invariant (depth == tc + dc, checked in the testbench) and random
legal/illegal program generation (book Lab 1 "Verification plan", random
test generation) compared model-vs-RTL with random output backpressure.
"""

import os
import random
import subprocess
import sys

import pytest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "tools"))

from opcodes import INSTRUCTIONS, encode  # noqa: E402
from stack_model import run_program  # noqa: E402
from state_checks import assert_architectural_states, initial_stack_args
from test_directed import PROGRAMS, _ensure_hex  # noqa: E402

ROOT = os.path.join(os.path.dirname(__file__), "..")
BUILD = os.path.join(ROOT, "build")

# BRAM core configs: (deep_depth, model total_depth, stall_seeds)
BRAM_PROGRAMS = [(name, deep, seeds)
                 for (name, _d, seeds) in PROGRAMS
                 for deep in ([30] if name != "stackoverflow" else [6])]


def _run_rtl_bram(hex_path, deep_depth, stall_seed=None, rom_depth=1024, initial_stack=()):
    tag = f"tb_bram_d{deep_depth}"
    vvp = os.path.join(BUILD, f"{tag}.vvp")
    subprocess.run([
        "iverilog", "-g2012", "-s", "tb_stack_core_bram", "-o", vvp,
        os.path.join(ROOT, "rtl", "symbolic_types_pkg.sv"),
        os.path.join(ROOT, "rtl", "sync_sdp_ram.sv"),
        os.path.join(ROOT, "rtl", "stack_core_bram.sv"),
        os.path.join(ROOT, "sim", "tb_stack_core_bram.sv"),
        f"-Ptb_stack_core_bram.DEEP_DEPTH={deep_depth}",
        f"-Ptb_stack_core_bram.ROM_DEPTH={rom_depth}",
    ], check=True, capture_output=True)
    cmd = ["vvp", vvp, f"+rom={hex_path}"]
    cmd += initial_stack_args(hex_path, initial_stack, deep_depth+2)
    if stall_seed is not None:
        cmd.append(f"+stall_seed={stall_seed}")
    out = subprocess.run(cmd, check=True, capture_output=True, text=True,
                         timeout=120)
    lines = [l for l in out.stdout.splitlines()
             if l.startswith(("TRACE", "FINAL", "TB_", "FAIL",
                             "ASSERT_FAIL"))]
    assert any(l.startswith("TB_PASS") for l in lines), out.stdout
    assert_architectural_states(out.stdout, hex_path, deep_depth+2, rom_depth, initial_stack)
    return [l for l in lines if l.startswith(("TRACE", "FINAL"))]


def _model_lines_for(words, total_depth):
    m = run_program(words, total_depth=total_depth)
    return [r.line() for r in m.trace] + [m.final_line()]


def _words_of(name):
    from asm20 import assemble
    with open(os.path.join(ROOT, "programs", f"{name}.asm")) as f:
        words, _, _ = assemble(f.read())
    return words


@pytest.mark.parametrize("name,deep,seeds", BRAM_PROGRAMS)
def test_directed_bram(name, deep, seeds):
    hex_path = _ensure_hex(name)
    total = deep + 2
    expected = _model_lines_for(_words_of(name), total)
    got = _run_rtl_bram(hex_path, deep)
    assert got == expected, _diff(got, expected)
    for seed in seeds:
        got_s = _run_rtl_bram(hex_path, deep, stall_seed=seed)
        assert got_s == expected, f"stall_seed={seed} changed the trace"


def _diff(got, expected):
    import difflib
    return "\n".join(difflib.unified_diff(
        expected, got, "model", "rtl", lineterm=""))


# ---------------------------------------------------------------------------
# Random program generation (book: maintain a typed software stack, choose
# only legal instructions, occasionally inject one illegal instruction).
# ---------------------------------------------------------------------------

from program_generation import gen_program


@pytest.mark.parametrize("seed", range(24))
def test_random_programs(seed):
    rng = random.Random(seed)
    for total_depth, deep in ((8, 6), (12, 10)):
        words, _ = gen_program(rng, length=120, total_depth=total_depth,
                               illegal_p=0 if seed % 2 == 0 else 0.06)
        hex_path = os.path.join(BUILD, f"rand_{seed}_{total_depth}.hex")
        os.makedirs(BUILD, exist_ok=True)
        with open(hex_path, "w") as f:
            for w in words:
                f.write(f"{w:05x}\n")
            for _ in range(1024 - len(words)):
                f.write("00000\n")
        expected = _model_lines_for(words, total_depth)
        got = _run_rtl_bram(hex_path, deep, stall_seed=seed)
        assert got == expected, _diff(got, expected)


@pytest.mark.parametrize("seed", range(12))
def test_random_programs_register_core(seed):
    """The register core must pass the same random programs (STACK_DEPTH
    equals the model total_depth)."""
    from test_directed import _run_rtl as _run_reg
    rng = random.Random(1000 + seed)
    words, _ = gen_program(rng, length=120, total_depth=8,
                           illegal_p=0 if seed % 2 == 0 else 0.06)
    hex_path = os.path.join(BUILD, f"randreg_{seed}.hex")
    os.makedirs(BUILD, exist_ok=True)
    with open(hex_path, "w") as f:
        for w in words:
            f.write(f"{w:05x}\n")
        for _ in range(1024 - len(words)):
            f.write("00000\n")
    expected = _model_lines_for(words, 8)
    got = _run_reg(hex_path, 8, stall_seed=seed)
    assert got == expected, _diff(got, expected)
