#!/usr/bin/env python3
"""Fail if RTL constants or simulation names drift from Python ISA metadata."""
from pathlib import Path
import re
import sys
ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / 'tools'))
from opcodes import BY_MNEMONIC, Fault, Tag, EVENT_NAMES


def check():
    package = (ROOT / 'rtl/symbolic_types_pkg.sv').read_text()
    values = {}
    for name, radix, raw in re.findall(r'\b((?:OP|TAG|F|EV)_\w+)\s*=\s*\d+\x27([hd])([0-9a-fA-F]+)', package):
        values[name] = int(raw, 16 if radix == 'h' else 10)
    fault_names = {'BAD_BRANCH_TARGET': 'BAD_BRANCH', 'NONCANONICAL_BOOL': 'NONCANON_BOOL'}
    expected = {'OP_'+n: i.opcode for n, i in BY_MNEMONIC.items()}
    expected.update({'TAG_'+t.name: int(t) for t in Tag})
    expected.update({'F_'+fault_names.get(f.name, f.name): int(f) for f in Fault})
    expected.update({'EV_'+name: code for code, name in EVENT_NAMES.items()})
    assert values == expected, (values, expected)
    for name in ['tb_stack_core.sv', 'tb_stack_core_bram.sv']:
        bench = (ROOT/'sim'/name).read_text()
        ops = {int(code, 16): text for code, text in re.findall(r"5'h([0-9a-fA-F]+): return \"(\w+)\"", bench)}
        faults = {int(code): text for code, text in re.findall(r'4\x27d(\d+): return "(\w+)"', bench)}
        assert ops == {i.opcode: n for n, i in BY_MNEMONIC.items()}, name
        assert faults == {int(f): f.name for f in Fault}, name


if __name__ == '__main__':
    check()
    print('ISA constants and testbench names match Python metadata')
