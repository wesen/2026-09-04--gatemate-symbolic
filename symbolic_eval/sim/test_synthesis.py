"""Regression for constant packed-struct constructors after Yosys lowering."""
from pathlib import Path
import subprocess


def test_synthesized_constructors(tmp_path):
    root = Path(__file__).resolve().parents[1]
    netlist = tmp_path / 'constructors.v'
    script = (f'read_verilog -sv rtl/symbolic_types_pkg.sv sim/constructor_probe.sv; '
              f'hierarchy -top constructor_probe; proc; opt; write_verilog -noattr {netlist}')
    subprocess.run(['yosys', '-Q', '-q', '-p', script], cwd=root, check=True, capture_output=True)
    binary = tmp_path / 'constructors.vvp'
    subprocess.run(['iverilog', '-g2012', '-s', 'tb_constructor', '-o', str(binary),
                    str(netlist), 'sim/tb_constructor.sv'], cwd=root, check=True, capture_output=True)
    result = subprocess.run(['vvp', str(binary)], text=True, capture_output=True)
    assert result.returncode == 0, result.stdout + result.stderr
    assert 'CONSTRUCTORS_PASS' in result.stdout
