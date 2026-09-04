CORE ?= trail
FIRST_ONLY ?= 0
TRAIL_CAPACITY ?= 64
CHOICE_CAPACITY ?= 8
PNR_SEED ?= 2
USE_TRAIL = $(if $(filter trail,$(CORE)),1,0)

.PHONY: test synth bit load versions
test:
	python3 -m pytest sim -q
versions:
	yosys -V
	nextpnr-himbaechel --version
	iverilog -V
synth:
	@test '$(CORE)' = trail -o '$(CORE)' = snapshot
	mkdir -p build
	python3 scripts/make_synth.py --trail $(USE_TRAIL) --first-only $(FIRST_ONLY) --trail-capacity $(TRAIL_CAPACITY) --choice-capacity $(CHOICE_CAPACITY)
	yosys -ql build/yosys.log build/synth.ys
bit: synth
	nextpnr-himbaechel -l build/nextpnr.log --device CCGM1A1 --json build/top.json -o ccf=../symbolic_eval/constraints/olimex_gatematea1_evb.ccf -o out=build/top.txt --sdc ../symbolic_eval/constraints/top.sdc --router router2 --seed $(PNR_SEED)
	gmpack build/top.txt build/top.bit
load: bit
	openFPGALoader -b olimex_gatemateevb build/top.bit
