#!/usr/bin/env python3
"""Build three configurations and compare physical UART records with the model."""
from pathlib import Path
import hashlib
import json
import subprocess
import sys
import time

TICKET = Path(__file__).resolve().parents[1]
REPO = next(p for p in TICKET.parents if (p/'queens_rollback').is_dir())
PROJECT = REPO/'queens_rollback'
OUT = TICKET/'reference/validation'
sys.path.insert(0, str(PROJECT/'tools'))
from queens_model import Machine

results = []
try:
    (OUT/'P5-build-commit.txt').write_text(subprocess.check_output(['git','rev-parse','HEAD'], cwd=REPO, text=True))
    for name, trail, first in [('snapshot',False,False),('trail',True,False),('trail-first',True,True)]:
        with (OUT/f'P5-{name}-build.log').open('w') as log:
            subprocess.run(['make','bit',f'CORE={"trail" if trail else "snapshot"}',f'FIRST_ONLY={int(first)}'], cwd=PROJECT, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=300)
        for filename in ['yosys.log','nextpnr.log']:
            (OUT/f'P5-{name}-{filename}').write_bytes((PROJECT/'build'/filename).read_bytes())
        image = (PROJECT/'build/top.bit').read_bytes()
        model = Machine(trail=trail, first_only=first).run()
        expected = (''.join(f'Q:{word:06X}\r\n' for word in model.output)+f'D:{len(model.output):08X}\r\n').encode()
        subprocess.run(['stty','-F','/dev/ttyACM0','115200','raw','-echo'],check=True)
        with (OUT/f'P5-{name}-uart.bin').open('wb') as capture:
            reader = subprocess.Popen(['timeout','8','cat','/dev/ttyACM0'], stdout=capture, stderr=subprocess.PIPE)
            try:
                time.sleep(0.3)
                with (OUT/f'P5-{name}-load.log').open('w') as log:
                    subprocess.run(['openFPGALoader','-b','olimex_gatemateevb','build/top.bit'], cwd=PROJECT, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=45)
                _, error = reader.communicate(timeout=12)
                assert reader.returncode in (0,124), error.decode()
            finally:
                if reader.poll() is None:
                    reader.terminate()
                    reader.wait()
        actual = (OUT/f'P5-{name}-uart.bin').read_bytes()
        result = dict(configuration=name, image_sha256=hashlib.sha256(image).hexdigest(), expected=expected.decode(), actual=actual.decode(errors='replace'), match=actual==expected)
        results.append(result)
        (OUT/'P5-hardware.json').write_text(json.dumps(results,indent=2)+'\n')
        print(json.dumps(result),flush=True)
        assert actual == expected, result
    (OUT/'P5-hardware.exit').write_text('0\n')
except Exception:
    (OUT/'P5-hardware.exit').write_text('1\n')
    raise
