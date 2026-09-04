"""Regression cases from GATEMATE-SYMBOLIC-002; both physical stacks."""
import pytest

from asm20 import assemble
from stack_model import run_program
from test_directed import _run_rtl
from test_bram import _run_rtl_bram


def compare_program(tmp_path, source, core, total_depth=8, stall_seed=7):
    words, _, _ = assemble(source)
    image = tmp_path / 'program.hex'
    image.write_text(''.join(f'{w:05x}\n' for w in words + [0] * (1024-len(words))))
    model = run_program(words, total_depth=total_depth)
    expected = [r.line() for r in model.trace] + [model.final_line()]
    runner = _run_rtl if core == 'reg' else _run_rtl_bram
    depth = total_depth if core == 'reg' else total_depth-2
    assert runner(str(image), depth, stall_seed=stall_seed) == expected
    return model


@pytest.mark.parametrize('core', ['reg', 'bram'])
@pytest.mark.parametrize('source', [
    'DUP\nHALT',
    'PUSH_TRUE\nDROP\nDUP\nHALT',
    'PUSH_TRUE\nEMIT\nDUP\nHALT',
    'PUSH_S15 7\nEMIT\nJMP end\nend: HALT',
    'PUSH_S15 1\nPUSH_S15 2\nPUSH_S15 3\nEMIT\nSWAP\nEMIT\nEMIT\nHALT',
    'PUSH_TRUE\n' * 8 + 'DUP\nHALT',
    'PUSH_S15 9\nDUP\nEMIT\nEMIT\nHALT',
])
def test_instruction_guards(tmp_path, core, source):
    compare_program(tmp_path, source, core)


@pytest.mark.parametrize('line', ['PUSH_S15 1 2', 'HALT extra tokens', 'label:: HALT', '@bad'])
def test_malformed_assembly_is_rejected(line):
    with pytest.raises(ValueError, match='line 2'):
        assemble('; valid comment\n' + line + '\nHALT')


@pytest.mark.parametrize('source', ['WORD 0\nWORD 0\nWORD 0', 'HALT\nWORD 0\nWORD 0',
                                  'WORD 0\nPUSH_S15 1\nWORD 0'])
def test_all_emission_paths_enforce_capacity(source):
    with pytest.raises(ValueError, match='ROM depth 2'):
        assemble(source, rom_depth=2)


def test_exact_capacity_and_comments():
    words, symbols, _ = assemble('; comment\n\nstart: WORD 0\nHALT ; done', rom_depth=2)
    assert len(words) == 2 and symbols == {'start': 0}


@pytest.mark.parametrize('core', ['reg', 'bram'])
@pytest.mark.parametrize('ending', ['WORD 0xF8000', 'WORD 0x5ffff', 'RET', 'JZ 0', 'WORD 0x7ffff'])
def test_deep_fault_context(tmp_path, core, ending):
    compare_program(tmp_path, 'PUSH_TRUE\nPUSH_S15 2\nPUSH_S15 3\nDROP\n' + ending, core)


@pytest.mark.parametrize('core', ['reg', 'bram'])
def test_call_return_retirement(tmp_path, core):
    compare_program(tmp_path, 'CALL sub\nHALT\nsub: CALL leaf\nRET\nleaf: RET', core)
