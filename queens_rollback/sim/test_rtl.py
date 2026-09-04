from pathlib import Path
import subprocess
import pytest
from queens_model import Machine
from oracle import solve_rows, unpack

ROOT=Path(__file__).resolve().parents[1]

def run_core(tmp_path,first=False,choices=8,stall=0,seed=1,trail=False,capacity=64):
    binary=tmp_path/'core.vvp'
    cmd=['iverilog','-g2012','-s','tb_queens','-o',str(binary),
         f'-Ptb_queens.FIRST_ONLY={int(first)}',f'-Ptb_queens.CHOICE_CAPACITY={choices}',
         f'-Ptb_queens.USE_TRAIL={int(trail)}',f'-Ptb_queens.TRAIL_CAPACITY={capacity}',
         str(ROOT/'rtl/queens_types_pkg.sv'),str(ROOT/'rtl/queens_core.sv'),
         str(ROOT.parent/'symbolic_eval/rtl/sync_sdp_ram.sv'),str(ROOT/'sim/tb_queens.sv')]
    subprocess.run(cmd,check=True,capture_output=True,text=True)
    run=subprocess.run(['vvp',str(binary),f'+stall={stall}',f'+seed={seed}'],
                       check=True,capture_output=True,text=True,timeout=45)
    model=Machine(trail=trail,first_only=first,choice_capacity=choices,trail_capacity=capacity).run()
    actual=[line for line in run.stdout.splitlines() if line.startswith(('E ','C','T'))]
    expected='\n'.join(model.events).splitlines()
    assert actual==expected, next(((i,a,b) for i,(a,b) in enumerate(zip(actual,expected)) if a!=b),
                                (len(actual),len(expected)))
    results=[int(line.split()[1],16) for line in run.stdout.splitlines() if line.startswith('RESULT ')]
    assert results==model.output
    return run.stdout

@pytest.mark.parametrize('first',[False,True])
@pytest.mark.parametrize('stall',[0,1])
def test_snapshot(tmp_path,first,stall):
    run_core(tmp_path,first=first,stall=stall)

@pytest.mark.parametrize('choices',[0,1,5])
def test_snapshot_capacity(tmp_path,choices):
    run_core(tmp_path,choices=choices)

@pytest.mark.parametrize('first',[False,True])
@pytest.mark.parametrize('stall',[0,1])
def test_trail(tmp_path,first,stall):
    run_core(tmp_path,first=first,stall=stall,trail=True)

@pytest.mark.parametrize('capacity',[0,1,7,31,32])
def test_trail_capacity_rtl(tmp_path,capacity):
    run_core(tmp_path,trail=True,capacity=capacity)

@pytest.mark.parametrize('choices',[0,1,5,6])
def test_trail_choice_capacity_rtl(tmp_path,choices):
    run_core(tmp_path,trail=True,choices=choices)
