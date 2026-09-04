#!/usr/bin/env python3
"""Build each program, arm serial capture before load, and compare board bytes."""
from pathlib import Path
import hashlib
import json
import subprocess
import sys
import time
TICKET=Path(__file__).resolve().parents[1]
PROJECT=next(p for p in TICKET.parents if (p/'symbolic_eval').is_dir())/'symbolic_eval'
OUT=TICKET/'reference/validation'
sys.path.insert(0,str(PROJECT/'tools'))
from asm20 import assemble
from stack_model import run_program
results=[]
try:
    for program in ['fib','arith','typefault','countdown']:
        if program!='fib':
            with (OUT/f'P5-{program}-build.log').open('w') as log:
                subprocess.run(['make','bit',f'PROG={program}'],cwd=PROJECT,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=180)
        image=(PROJECT/'build/top.bit').read_bytes()
        words,_,_=assemble((PROJECT/f'programs/{program}.asm').read_text())
        model=run_program(words,total_depth=514)
        expected=''.join(f'T{v.tag:x}:{v.payload:08X}\r\n' for v in model.output).encode()
        subprocess.run(['stty','-F','/dev/ttyACM0','115200','raw','-echo'],check=True)
        with (OUT/f'P5-{program}-uart.bin').open('wb') as capture:
            reader=subprocess.Popen(['timeout','8','cat','/dev/ttyACM0'],stdout=capture,stderr=subprocess.PIPE)
            try:
                time.sleep(0.3)
                with (OUT/f'P5-{program}-load.log').open('w') as log:
                    subprocess.run(['openFPGALoader','-b','olimex_gatemateevb','build/top.bit'],cwd=PROJECT,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=45)
                _,err=reader.communicate(timeout=12)
                assert reader.returncode in (0,124),err.decode()
            finally:
                if reader.poll() is None:
                    reader.terminate();reader.wait()
        actual=(OUT/f'P5-{program}-uart.bin').read_bytes()
        result={'program':program,'image_sha256':hashlib.sha256(image).hexdigest(),
                'expected':expected.decode(),'actual':actual.decode(errors='replace'),
                'match':actual==expected,'model_fault':model.fault.name if model.fault else None,
                'evidence_limit':'UART silence alone does not expose precise hardware fault state' if model.fault else None}
        results.append(result)
        (OUT/'P5-hardware.json').write_text(json.dumps(results,indent=2)+'\n')
        print(json.dumps(result),flush=True)
        assert actual==expected,result
        for name in ['yosys.log','nextpnr.log']:
            (OUT/f'P5-{program}-{name}').write_text((PROJECT/'build'/name).read_text())
    (OUT/'P5-hardware.exit').write_text('0\n')
except Exception:
    (OUT/'P5-hardware.exit').write_text('1\n')
    raise
