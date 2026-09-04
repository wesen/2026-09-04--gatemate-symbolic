≡ ![GateMateA1-EVB - Open Source Hardware Board](/Products/FPGA/GateMate/GateMateA1-EVB/images/thumbs/310x230/GateMateA1-EVB.jpg) ![GateMateA1-EVB - Open Source Hardware Board](/Products/FPGA/GateMate/GateMateA1-EVB/images/thumbs/100x75/GateMateA1-EVB.jpg) ![GateMateA1-EVB - Open Source Hardware Board](/Products/FPGA/GateMate/GateMateA1-EVB/images/thumbs/100x75/GateMateA1-EVB-1.jpg) ![GateMateA1-EVB - Open Source Hardware Board](/Products/FPGA/GateMate/GateMateA1-EVB/images/thumbs/100x75/GateMateA1-EVB-3.jpg)

**Development board for CCGM1A1**

#### Select Product Variant

- GateMateA1-EVB
- GateMateA1-EVB-2M

| Price | 50.00 EUR |
| --- | --- |

GateMateA1-EVB is [OSHW certified](https://certification.oshwa.org/list.html) Open Source Hardware with UID BG0000117

GateMateA1-EVB-2M has U3 2MB of SPI flash and the surrounding components soldered on board.

#### FEATURES

- CCGM1A1 FPGA with 20480 logic cells
- PSRAM 64Mbit
- RP2040 processor for programing and debugging
- 2MB configuration Flash for RP2040
- 4 buttons
- USB-C for power supply and programming
- PS2 connector
- VGA connector
- 4 Banks with signals with selectable levels 1.2V 1.8V 2.5V
- PMOD with level shifters
- UEXT with level shifters
- Power LED
- User LED
- 4 sections configuration slide switch
- Dimensions: 120 x 80 mm

#### HARDWARE

- [GateMateA1-EVB schematic in PDF format](https://github.com/OLIMEX/GateMateA1-EVB/blob/main/HARDWARE/GateMateA1-EVB-Rev.C/GateMateA1-EVB_Rev_C.pdf)
- [GateMateA1-EVB repository with sources in KiCAD format](https://github.com/OLIMEX/GateMateA1-EVB)

#### SOFTWARE

- [GateMate toolchain installation guide](https://www.colognechip.com/docs/ug1002-toolchain-install-latest.pdf)
- [pico-dirtyJtag binary and README](https://github.com/OLIMEX/GateMateA1-EVB/tree/main/SOFTWARE/dirtyJtag)
- [Gatemate ILA (Integrated Logic Analyzer) repository](https://github.com/colognechip/gatemate_ila)

#### COMMUNITY

- [Michael Schröder's 8080 emulation in Verilog, running Space Invaders](https://gitlab.com/x653/spaceinvaders-fpga)
- [Yuri Zaporozhets IBM PC emulation](https://gitlab.com/gatemate/pc)

#### FAQ

- How to program this board?
- Via the USB connector. The board comes with dirtyJtag bootloader already programmed and we use openFPGALoader software to program it. The command we use to program it here is:  
	  
	sudo./openFPGALoader -c dirtyJtag blink\_00.cfg\_LED\_BUT.bit  
	  
	The openFPGALoader reposiotry can be found here: https://github.com/trabucayre/openFPGALoader  
	  
	Maybe check the instructions here: https://trabucayre.github.io/openFPGALoader/guide/install.html  
	  
	The dirtyJtag bootloader repo is here: https://github.com/phdussud/pico-dirtyJtag