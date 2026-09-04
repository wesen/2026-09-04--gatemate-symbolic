import pytest
from oracle import solve_rows, valid_board, unpack
from queens_model import Machine, Fault, attack, singleton


def test_oracle():
    boards = solve_rows()
    assert len(boards) == len(set(boards)) == 92
    assert boards[0] == (0,4,7,5,2,6,1,3)
    assert all(valid_board(board) for board in boards)


def test_helpers():
    for mask in range(256):
        assert singleton(mask) == (mask.bit_count() == 1)
    for source in range(8):
        for row in range(8):
            for target in range(8):
                expected = sum(1 << r for r in range(8)
                               if r == row or abs(r-row) == abs(target-source))
                assert attack(source,row,target) == expected


@pytest.mark.parametrize('trail',[False,True])
@pytest.mark.parametrize('first_only',[False,True])
def test_enumeration(trail,first_only):
    machine = Machine(trail=trail,first_only=first_only).run()
    assert not machine.fault and machine.done
    assert list(map(unpack,machine.output)) == (solve_rows()[:1] if first_only else solve_rows())
    assert not machine.choices
    assert machine.base == len(machine.trail) if first_only else not machine.trail
    assert machine.max_trail <= 64 and machine.max_choices <= 8


@pytest.mark.parametrize('trail',[False,True])
def test_blocked_result_and_cut(trail):
    m = Machine(trail=trail,first_only=True)
    while not m.waiting: m.step(False)
    before = (m.domains.copy(),m.propagated,len(m.choices),m.trail.copy(),m.pending)
    for _ in range(100):
        assert m.step(False) is None
        assert before == (m.domains,m.propagated,len(m.choices),m.trail,m.pending)
        assert not m.done and not m.output
    m.step(True)
    assert m.output == [0x672be0] and not m.choices and m.base == len(m.trail)
    m.run()
    assert m.done and len(m.output) == 1


@pytest.mark.parametrize('capacity',[0,1,7,31])
def test_trail_capacity(capacity):
    m = Machine(trail_capacity=capacity).run()
    assert m.fault == Fault.TRAIL_FULL
    assert len(m.trail) == capacity
    if capacity == 0: assert m.domains == [255]*8 and not m.choices
    if capacity == 1: assert m.domains == [1]+[255]*7


@pytest.mark.parametrize('capacity',[0,1,5])
def test_choice_capacity(capacity):
    m = Machine(choice_capacity=capacity).run()
    assert m.fault == Fault.CHOICE_FULL and len(m.choices) == capacity


def test_minimum_sufficient_measured_capacity():
    assert len(Machine(trail_capacity=32,choice_capacity=6).run().output) == 92
