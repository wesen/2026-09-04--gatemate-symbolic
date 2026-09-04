"""Ensure generated legal execution and metadata checks enforce their claims."""
from collections import Counter
import importlib.util
from pathlib import Path
import random
import pytest
from opcodes import Fault, INSTRUCTIONS
from stack_model import run_program
from program_generation import gen_program
from test_repairs import compare_words


def test_isa_mirrors():
    path = Path(__file__).resolve().parents[1]/'scripts/check_isa.py'
    spec = importlib.util.spec_from_file_location('check_isa', path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    module.check()


def test_legal_generator_coverage():
    counts = Counter()
    branches = Counter()
    for seed in range(24):
        words, _ = gen_program(random.Random(seed), length=120, total_depth=8, illegal_p=0)
        m = run_program(words, total_depth=8)
        assert m.halted and m.fault is None
        assert all(r.pc_old < len(words) for r in m.trace)
        counts.update(r.op for r in m.trace)
        for r in m.trace:
            if r.op == 'JZ':
                branches[r.pc_new != r.pc_old+1] += 1
    assert set(counts) == {i.mnemonic for i in INSTRUCTIONS.values()}
    assert branches[True] > 0 and branches[False] > 0


@pytest.mark.parametrize('seed', range(4))
def test_fault_injection_is_a_fault(seed):
    words, _ = gen_program(random.Random(seed), illegal_p=1)
    assert run_program(words).fault is not None


@pytest.mark.parametrize('overflow', [False, True])
def test_production_bram_capacity(tmp_path, overflow):
    from opcodes import encode
    words = [encode(0, n % 16000) for n in range(514)]
    # Fill physical capacity; drain as far as the 1024-word ROM permits.
    words += [encode(8)] if overflow else [encode(9)] * 509 + [encode(0x0e)]
    m = compare_words(tmp_path, words, 'bram', total_depth=514)
    assert m.fault == Fault.STACK_OVERFLOW if overflow else m.halted and m.depth == 5


@pytest.mark.parametrize('core', ['reg', 'bram'])
@pytest.mark.parametrize('case', ['add_overflow', 'sub_overflow', 'mul_overflow', 'noncanonical', 'flags', 'signed_lt'])
def test_constructed_architectural_states(tmp_path, core, case):
    from opcodes import encode
    from stack_model import Value40, make_int
    from test_directed import _run_rtl
    from test_bram import _run_rtl_bram
    cases = {
        'add_overflow': ([make_int(2147483647), make_int(1)], [encode(3)]),
        'sub_overflow': ([make_int(-2147483648), make_int(1)], [encode(4)]),
        'mul_overflow': ([make_int(-2147483648), make_int(-1)], [encode(5)]),
        'noncanonical': ([Value40(1, 5)], [encode(0x0c, 0)]),
        'flags': ([Value40(0, 7, 2), Value40(0, 7, 3)],
                  [encode(8), encode(0x0d), encode(6), encode(0x0d)]),
        'signed_lt': ([make_int(-2147483648), make_int(2147483647)], [encode(7), encode(0x0d)]),
    }
    stack, words = cases[case]
    words += [encode(0x0e)]
    image = tmp_path/'initial.hex'
    image.write_text(''.join(f'{w:05x}\n' for w in words+[0]*(1024-len(words))))
    runner = _run_rtl if core == 'reg' else _run_rtl_bram
    runner(str(image), 8 if core == 'reg' else 6, initial_stack=stack, stall_seed=7)
