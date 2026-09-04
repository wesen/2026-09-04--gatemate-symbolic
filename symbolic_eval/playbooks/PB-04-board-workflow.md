# PB-04: Board workflow (build, load, capture, restart)

Everything runs from the Makefile. The program image is baked into the ROM at
synthesis time — changing programs means re-running `make bit`.

## Build and load

```bash
source ~/fpga/oss-cad-suite/environment
cd symbolic_eval
make versions              # record tool versions
make test                  # 198 tests, including state and synthesized constructors
make asm PROG=fib          # assemble programs/fib.asm -> build/fib.hex
make bit PROG=fib          # synth -> PnR -> pack (copies fib.hex to prog.hex)
make load PROG=fib         # keep PROG: load rebuilds its selected image
```

Budget check after synthesis: `grep -E "CC_BRAM_20K|CC_MULT" build/yosys.log`
and the `Device utilisation` / `Max frequency` lines in `build/nextpnr.log`.
Stop-build budget: 2 BRAM blocks, ~2000 CPEs (before debug instrumentation).

## Capturing UART output

**The program runs at configuration time** — its output is on the wire within
milliseconds of `openFPGALoader` finishing. Attach the reader *before* loading:

```bash
stty -F /dev/ttyACM0 115200 raw -echo
rm -f /tmp/uart.log
(timeout 6 cat /dev/ttyACM0 > /tmp/uart.log &)
sleep 0.3
openFPGALoader -b olimex_gatemateevb build/top.bit
sleep 3.5
od -c /tmp/uart.log        # e.g. T0:00000037 = INT(55)
```

`/dev/ttyACM0` is the RP2040 CDC UART; ACM1 is the JTAG channel.

Output format: one EMIT value = 13 bytes,
`'T' <tag hex> ':' <8 payload hex digits, MSB first> CR LF`.
Examples: `T1:00000001` = BOOL(true), `T0:00000037` = INT(55).

## LED codes (active-low: pin LOW = lit)

- slow blink (~0.6 Hz): running
- solid lit: halted (success)
- fast blink (~2.4 Hz): faulted

## Restarting a run

The FPGA button (active-low) is wired as a global experiment abort: press =
synchronized assertion, release = synchronous restart from the initial image. No reload
needed — press it while a terminal is attached to `ACM0` and watch the output
again.

## Interpreting "no output"

For fault-demo programs (`typefault.asm`, `retunderflow.asm`) an empty capture
**is** the expected result — the precise-fault exit criterion includes "no
output transfer". Distinguish "no output because faulted" (LED fast-blink) from
"no output because broken" (run the same program in `make test` / `tb_top`
first; the board sim is a pytest suite).
