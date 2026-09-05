#!/usr/bin/env python3
from pathlib import Path
import hashlib,json,re
root=Path(__file__).resolve().parents[1]
validation=root/'reference/validation'
slips=['parser-plan']+[f'parser-s{n}-{stage}' for n in range(1,4) for stage in ['start','done']]
for name in slips:assert 'printed: true' in (validation/(name+'-print.log')).read_text(),name
fuzz=(validation/'parser-fuzz.log').read_text();assert '\nPASS\n' in fuzz
counts=re.findall(r'execs: (\d+)',fuzz)
coverage=(validation/'parser-s3-tests.log').read_text();assert 'coverage: 94.5%' in coverage
logs={}
for name in ['parser-s1-tests','parser-s2-tests','parser-s3-tests','parser-repository-tests','parser-race','parser-vet','parser-build','parser-fuzz']:
 p=validation/(name+'.log');text=p.read_text();assert 'FAIL' not in text,name
 logs[name]=hashlib.sha256(p.read_bytes()).hexdigest()
print(json.dumps({'implemented':'lexer and recursive-descent/Pratt parser','printed_slips':len(slips),'fuzz_executions':int(counts[-1]),'statement_coverage_percent':94.5,'checks':['focused tests','repository tests','parser race tests','go vet ./...','go build ./...','bounded fuzzing'],'log_sha256':logs,'remaining_I1':['binding resolution','monomorphic type checking','independent semantic evaluator']},indent=2))
