Z=	zasm
AFLAGS=	-u -v2

tinybasic2.rom: tinybasic2.asm
	${Z} --asm8080 ${AFLAGS} $<

CPM22.rom: CPM22.asm
	${Z} --asm8080 ${AFLAGS} $<

z80_cbios.rom: z80_cbios.asm
	${Z} ${AFLAGS} $<

2K_ROM_8.rom: 2K_ROM_8.asm
	${Z} ${AFLAGS} $<

clean:
	rm *.rom *.lst
