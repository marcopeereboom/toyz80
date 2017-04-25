CC	?=	sdcc

ifeq ($(CC),sdcc)
CFLAGS	?=	-mz80
else
CFLAGS	?=	-g -O2
endif
monitor: monitor.c
	${CC} ${CFLAGS} -o $@ $<

all: monitor

clean:
	rm monitor monitor.{asm,lk,lst,map,noi,rel,sym}
