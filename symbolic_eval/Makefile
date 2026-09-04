# Makefile — symbolic_eval: Laboratory 1 tagged stack evaluator
# Source book: Composable_Hardware_Patterns_for_Symbolic_Computers.md (Lab 1)
# Toolchain: OSS CAD Suite — source its `environment` before building:
#     source ~/fpga/oss-cad-suite/environment
#
# Phases (design doc §11): P0 bootstrap, P1 model, P2 assembler, P3 register
# RTL, P4 BRAM RTL, P5 board. Every stage fails immediately when an earlier
# stage fails.

TOP      ?= top
DEVICE   ?= CCGM1A1
BOARD    ?= olimex_gatemateevb
CCF      ?= constraints/olimex_gatematea1_evb.ccf
SDC      ?= constraints/top.sdc

RTL_DIR  := rtl
SIM_DIR  := sim
TOOLS_DIR := tools
BUILD    := build

PROG     ?= arith

# All synthesizable RTL. top.sv arrives in P5; until then blink_top is the
# synthesis target (P0 evidence).
RTL_SYNTH := $(RTL_DIR)/reset_sync.sv $(RTL_DIR)/blink_top.sv

.PHONY: all versions test sim synth pnr bit load asm clean
all: sim bit

versions:
	@mkdir -p $(BUILD)
	@{ \
	  date -u; \
	  echo "OSS CAD Suite: $$(cat ~/fpga/oss-cad-suite/VERSION 2>/dev/null)"; \
	  yosys -V; \
	  nextpnr-himbaechel --version 2>&1; \
	  openFPGALoader --version 2>&1; \
	  iverilog -V 2>&1 | head -n 2; \
	  python3 --version 2>&1; \
	} | tee $(BUILD)/tool-versions.txt

# Pure-software tests: opcode metadata + reference model + assembler (P1/P2),
# then differential RTL tests (P3/P4) which invoke iverilog themselves.
test:
	python3 -m pytest sim/ -q

# Assemble a program: programs/$(PROG).asm -> build/$(PROG).{hex,bin,lst,sym.json}
asm:
	python3 $(TOOLS_DIR)/asm20.py programs/$(PROG).asm -o $(BUILD) -n $(PROG)

# Simulation of the P0 blink top (uses the sim-only CC_USR_RSTN model).
sim:
	@mkdir -p $(BUILD)
	iverilog -g2012 -s tb_blink -o $(BUILD)/blink.vvp \
	  $(RTL_DIR)/blink_top.sv $(RTL_DIR)/reset_sync.sv $(SIM_DIR)/CC_USR_RSTN.sv $(SIM_DIR)/tb_blink.sv
	vvp $(BUILD)/blink.vvp

synth:
	@mkdir -p $(BUILD)
	yosys -ql $(BUILD)/yosys.log -p 'read_verilog -defer -sv $(RTL_SYNTH); synth_gatemate -top blink_top -luttree -nomx8; stat; write_json $(BUILD)/blink.json;'

pnr: synth
	nextpnr-himbaechel -l $(BUILD)/nextpnr.log --device $(DEVICE) \
	  --json $(BUILD)/blink.json \
	  -o ccf=$(CCF) -o out=$(BUILD)/blink.txt \
	  --sdc $(SDC) --router router2

bit: pnr
	gmpack $(BUILD)/blink.txt $(BUILD)/blink.bit

load: bit
	openFPGALoader -b $(BOARD) $(BUILD)/blink.bit

clean:
	rm -f $(BUILD)/*.{json,txt,bit,vcd,log,vvp,hex,bin,lst,out}
