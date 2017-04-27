CC	?= sdcc
MAKEBIN	?= makebin

ifeq ($(CC),sdcc)
CFLAGS	?= -mz80 --code-loc 0x1000 --data-loc 0x0200

all: monitor.bin monitor.ihx

monitor.bin: monitor.ihx
	${MAKEBIN} $< > $@

i8251.rel: i8251.c
	${CC} -c -mz80 $<

disassembler.rel: disassembler.c
	${CC} -c -mz80 $<

monitor.ihx: monitor.c i8251.rel disassembler.rel
	${CC} ${CFLAGS} monitor.c i8251.rel disassembler.rel

else
CFLAGS	?=-g -O2
monitor: monitor.c
	${CC} ${CFLAGS} -o $@ $<

all: monitor
endif

clean:
	rm -rf monitor
	rm -rf *.{asm,bin,ihx,lk,lst,map,mem,noi,rel,rst,sym}
