CC= sdcc
MAKEBIN= makebin
CFLAGS= -mz80 --code-loc 0x0180 --data-loc 0x1000

all: hello.bin hello.ihx

hello.bin: hello.ihx
	${MAKEBIN} $< > $@

hello.ihx: hello.c
	${CC} ${CFLAGS} $<

clean:
	rm hello.{asm,bin,ihx,lk,lst,map,noi,rel,sym}
