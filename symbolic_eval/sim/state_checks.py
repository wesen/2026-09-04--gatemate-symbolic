"""Compare the complete abstraction, independently of the concise TRACE format."""
import difflib
from collections import Counter
from pathlib import Path
from stack_model import Machine, load_hex

COVERAGE = {name: Counter() for name in ['opcodes', 'faults', 'branches', 'cache_transitions']}


def assert_architectural_states(stdout, hex_path, total_depth, rom_depth, initial_stack=()):
    m = Machine(program=load_hex(hex_path), total_depth=total_depth,
                rom_depth=rom_depth, stack=list(initial_stack))
    expected = []
    for _ in range(100000):
        record = m.step()
        if record is None:
            break
        data = ''.join(f' {v.word():010x}' for v in m.stack)
        returns = ''.join(f' {pc}' for pc in m.rstack)
        expected.append(f'STATE {record.seq} PC {m.pc} D {m.depth}{data} R {len(m.rstack)}{returns}')
    assert m.halted or m.fault is not None, 'model watchdog expired'
    actual = [line for line in stdout.splitlines() if line.startswith('STATE ')]
    assert actual == expected, '\n'.join(difflib.unified_diff(expected, actual, 'model state', 'RTL state'))
    transfers = [line for line in stdout.splitlines() if line.startswith('XFER ')]
    assert transfers == [f'XFER {v.word():010x}' for v in m.output]
    trace = [line for line in stdout.splitlines() if line.startswith(('TRACE ', 'FINAL '))]
    assert trace == [r.line() for r in m.trace] + [m.final_line()]
    for record in m.trace:
        COVERAGE['opcodes'][record.op] += 1
        if record.fault_code is not None:
            COVERAGE['faults'][record.fault_code.name] += 1
        if record.op == 'JZ' and record.fault_code is None:
            COVERAGE['branches']['taken' if record.pc_new != record.pc_old+1 else 'not_taken'] += 1
    for line in stdout.splitlines():
        if line.startswith('CACHE '):
            COVERAGE['cache_transitions'][line[6:]] += 1


def initial_stack_args(hex_path, initial_stack, total_depth):
    if not initial_stack:
        return []
    path = Path(hex_path).with_suffix('.stack.hex')
    words = [v.word() for v in initial_stack]
    path.write_text(''.join(f'{w:010x}\n' for w in words + [0]*(total_depth-len(words))))
    return [f'+stack={path}', f'+stack_count={len(words)}']
