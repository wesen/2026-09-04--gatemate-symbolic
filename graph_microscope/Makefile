PNR_SEED ?= 2
.PHONY: synth bit load
synth:
	mkdir -p build
	yosys -ql build/yosys.log scripts/synth.ys
bit: synth
	nextpnr-himbaechel -l build/nextpnr.log --device CCGM1A1 --json build/top.json -o ccf=../symbolic_eval/constraints/olimex_gatematea1_evb.ccf -o out=build/top.txt --sdc ../symbolic_eval/constraints/top.sdc --router router2 --seed $(PNR_SEED)
	gmpack build/top.txt build/top.bit
load: bit
	openFPGALoader -b olimex_gatemateevb build/top.bit
