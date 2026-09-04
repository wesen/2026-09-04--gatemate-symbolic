from pathlib import Path
import subprocess
import pytest
from queens_model import Machine

ROOT=Path(__file__).resolve().parents[1]
RTL=['queens_types_pkg.sv','queens_core.sv','queens_result_printer.sv','queens_top.sv']
SHARED=['sync_sdp_ram.sv','reset_sync.sv','uart_tx.sv']

@pytest.mark.parametrize('trail,first,capacity,restart',[
    (0,0,64,0),(1,0,64,0),(0,1,64,0),(1,1,64,0),
    (1,0,0,0),(1,0,31,0),(1,1,64,1),(0,1,64,1)])
def test_board_stream(tmp_path,trail,first,capacity,restart):
    binary=tmp_path/'board.vvp'
    cmd=['iverilog','-g2012','-s','tb_queens_top','-o',str(binary),
         f'-Ptb_queens_top.USE_TRAIL={trail}',f'-Ptb_queens_top.FIRST_ONLY={first}',
         f'-Ptb_queens_top.TRAIL_CAPACITY={capacity}',
         *[str(ROOT/'rtl'/f) for f in RTL],
         *[str(ROOT.parent/'symbolic_eval/rtl'/f) for f in SHARED],
         str(ROOT.parent/'symbolic_eval/sim/CC_USR_RSTN.sv'),str(ROOT/'sim/tb_queens_top.sv')]
    subprocess.run(cmd,check=True,capture_output=True,text=True)
    result=subprocess.run(['vvp',str(binary),f'+restart={restart}'],check=True,
                          capture_output=True,text=True,timeout=45)
    stdout=result.stdout.split('RESTART\n')[-1]
    got=bytes(int(line.split()[1],16) for line in stdout.splitlines() if line.startswith('BYTE '))
    model=Machine(trail=bool(trail),first_only=bool(first),trail_capacity=capacity).run()
    expected=''.join(f'Q:{word:06X}\r\n' for word in model.output)
    expected+=f'F:{int(model.fault):02X}\r\n' if model.fault else f'D:{len(model.output):08X}\r\n'
    assert got==expected.encode(),(got,expected)
    assert f'TOPDONE {int(model.done)} {int(model.fault)} {len(model.output)}' in stdout
