"""Board-level tests: top.sv (core + ROM + rv_reg + value printer + UART)
simulated end to end. The UART byte stream must spell the model's EMIT
values in the documented T<tag>:<payload-hex>\\r\\n format, and the machine
must reach the book's exit criteria states.
"""

import os
import subprocess
import sys

import pytest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "tools"))

from asm20 import assemble  # noqa: E402
from stack_model import run_program  # noqa: E402
from test_directed import _ensure_hex  # noqa: E402

ROOT = os.path.join(os.path.dirname(__file__), "..")
BUILD = os.path.join(ROOT, "build")

RTL = ["rtl/symbolic_types_pkg.sv", "rtl/sync_sdp_ram.sv",
       "rtl/stack_core_bram.sv", "rtl/program_rom.sv", "rtl/rv_reg.sv",
       "rtl/reset_sync.sv", "rtl/uart_tx.sv", "rtl/top.sv"]


def _run_top(hex_path, max_cycles=300000):
    vvp = os.path.join(BUILD, "tb_top.vvp")
    subprocess.run([
        "iverilog", "-g2012", "-s", "tb_top", "-o", vvp,
        *[os.path.join(ROOT, f) for f in RTL],
        os.path.join(ROOT, "sim", "CC_USR_RSTN.sv"),
        os.path.join(ROOT, "sim", "tb_top.sv"),
    ], check=True, capture_output=True)
    out = subprocess.run(["vvp", vvp, f"+rom={hex_path}",
                          f"+max_cycles={max_cycles}"],
                         check=True, capture_output=True, text=True,
                         timeout=300)
    assert "FAIL" not in out.stdout, out.stdout
    return out.stdout


def _bytes_to_str(hex_bytes):
    return "".join(chr(int(h, 16)) for h in hex_bytes)


def _uart_lines(stdout):
    return [l.split()[1] for l in stdout.splitlines()
            if l.startswith("UARTBYTE")]


def _model_emits(name):
    with open(os.path.join(ROOT, "programs", f"{name}.asm")) as f:
        words, _, _ = assemble(f.read())
    m = run_program(words)
    return m.output


def test_program_a_board_stream():
    stdout = _run_top(_ensure_hex("arith"))
    got = _bytes_to_str(_uart_lines(stdout))
    assert got == "T1:00000001\r\n"
    assert "TOPDONE 1 0 8 0" in stdout


def test_program_b_board_stream():
    stdout = _run_top(_ensure_hex("typefault"))
    got = _bytes_to_str(_uart_lines(stdout))
    assert got == ""                      # no output transfer
    assert "TOPDONE 0 1 2 2" in stdout    # faulted, pc=2, depth=2


def test_smoke_board_stream_matches_model():
    stdout = _run_top(_ensure_hex("smoke"))
    got = _bytes_to_str(_uart_lines(stdout))
    expected = "".join(
        f"T{v.tag:x}:{v.payload:08X}\r\n" for v in _model_emits("smoke"))
    assert got == expected
    assert "TOPDONE 1 0 10 0" in stdout


def test_deep_board_stream_matches_model():
    stdout = _run_top(_ensure_hex("deep"))
    got = _bytes_to_str(_uart_lines(stdout))
    expected = "".join(
        f"T{v.tag:x}:{v.payload:08X}\r\n" for v in _model_emits("deep"))
    assert got == expected
    assert "TOPDONE 1 0 18 1" in stdout
