CC	?= sdcc
MAKEBIN	?= makebin

ifeq ($(CC),sdcc)
CFLAGS	?= -mz80
LDFLAGS ?= --code-loc 0x1000 --data-loc 0x0200
all: monitor.bin monitor.ihx

%.rel: %.c
	${CC} -c ${CFLAGS} $<

monitor.ihx: monitor.rel i8251.rel disassembler.rel
	${CC} ${CFLAGS} ${LDFLAGS} $^

monitor.bin: monitor.ihx
	${MAKEBIN} $< > $@

else
CFLAGS	?=-g -O2
monitor: monitor.c disassembler.c
	${CC} ${CFLAGS} -o $@ $^

all: monitor
endif

clean:
	rm -rf monitor
	rm -rf *.{asm,bin,ihx,lk,lst,map,mem,noi,rel,rst,sym}
