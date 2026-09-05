#!/usr/bin/env python3
from pathlib import Path
root=Path(__file__).resolve().parents[1];repo=root.parents[4]
p=repo/'web/src/lazy/style.css';p.write_text(p.read_text().replace('.lazy-source{','textarea.lazy-source{'))
p=repo/'web/src/lazy/types.ts';s=p.read_text();s=s.replace("const t=Number(tag);let payload=0;", "const t=Number(tag);const allowed=t===0?['tag','value']:t===13?['tag','code']:[1,2].includes(t)?['tag','a','b']:[3,4].includes(t)?['tag','a']:['tag'];if(Object.keys(n).some(k=>!allowed.includes(k)))throw Error('Unexpected node field.');let payload=0;");p.write_text(s)
p=root/'scripts/11-browser-model.js';s=p.read_text().replace("await page.getByLabel('Graph source').fill(old);", "await page.getByLabel('Graph source').fill(old);await page.getByLabel('Cycles to advance').fill('1');await change('Advance');");p.write_text(s)
