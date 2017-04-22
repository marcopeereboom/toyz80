#include <stdio.h>
#include <stdint.h>
#include <math.h>

#if defined(__clang__)
#define init_console()
#else
#define I8251_ADDRESS	(0x02)
#define	I8251_DATA	(I8251_ADDRESS )
#define	I8251_STATUS	(I8251_ADDRESS + 1)

/* start of day mode settings */
#define	I8251_MODE_BAUD600		(0x01)
#define	I8251_MODE_BAUD9600		(0x02)
#define	I8251_MODE_BAUD38400		(0x03)
#define	I8251_MODE_BITS5		(0x04)
#define	I8251_MODE_BITS6		(0x06)
#define	I8251_MODE_BITS7		(0x08)
#define	I8251_MODE_BITS8		(0x0c)
#define	I8251_MODE_PARITYDISABLE	(0x00)
#define	I8251_MODE_PARITYODD		(0x10)
#define	I8251_MODE_PARITYDISABLE2	(0x20)
#define	I8251_MODE_PARITYEVEN		(0x30)
#define	I8251_MODE_STOP			(0x00)
#define	I8251_MODE_STOP1		(0x40)
#define	I8251_MODE_STOP15		(0x80)
#define	I8251_MODE_STOP2		(0xc0)

#define	I8251_CMD_TXEN			(1<<0)
#define	I8251_CMD_DTR			(1<<1)
#define	I8251_CMD_RXEN			(1<<2)
#define	I8251_CMD_SBRK			(1<<3)
#define	I8251_CMD_ER			(1<<4)
#define	I8251_CMD_RTS			(1<<5)
#define	I8251_CMD_IR			(1<<6)
#define	I8251_CMD_HUNT			(1<<7)
		// bit 0 TXEN
		//	00 disable
		//	01 transmit enable
		// bit 1 DTR (low active)
		//	00 DTR = 1
		//	02 DTR = 0
		// bit 2 RXE
		//	00 disable
		//	04 receive enable
		// bit 3 SBRK
		//	08 send SBRK
		//	00 normal operation
		// bit 4 ER
		//	10 reset error flag
		//	00 normal operation
		// bit 5 RTS (low active)
		//	00 RTS = 1
		//	20 RTS = 0
		// bit 6 IR
		//	40 internal reset
		//	00 normal operation
		// bit 7 EH
		//	80 hunt mode
		//	00 normal operation
// override putchar from stdlib
void
putchar(unsigned char ch)
{
	ch; // shut compiler up
__asm
	// ch is in register a
	// ld	hl,#2		// add 2 to sp to get to a
	// add	hl,sp
	// ld	a,(hl)		// a = ch
	out	(I8251_DATA),a	// write a to UART
__endasm;
}

// Can't be wrapped beacuse asm.
#define I8251_MODE_DEFAULT	(I8251_MODE_STOP1|I8251_MODE_PARITYDISABLE|I8251_MODE_BITS8|I8251_MODE_BAUD9600)
#define I8251_CMD_DEFAULT	(I8251_CMD_TXEN|I8251_CMD_RXEN|I8251_CMD_DTR|I8251_CMD_RTS|I8251_CMD_ER)
void
init_console()
{
__asm
	ld	a,#I8251_MODE_DEFAULT	// initialize 8251A UART
	out	(I8251_STATUS),a	// 1 stop bit, no parity, 8-bit
	ld	a,#I8251_CMD_DEFAULT	// enable rx/tx/dtr/rts
	out	(I8251_STATUS),a	// and reset error flag
__endasm;
}
#endif /* __clang__ */

void
space(int x)
{
	int	i;
	for (i = 0; i < x; i++) {
		printf(" ");
	}
}

int
main(int argc, char *argv[])
{
	float	x;

	argc; argv; // shut compiler up

	init_console();
	printf("Hello Z80 C world! int %d string %s hex %x\r\n\r\n",
	    1, "haha", 0xff);

	x = 0;
	do {
		space((int)(40+sinf(x)*20));
		printf("*\r\n");
		x += 0.2;
	} while (x <= 6.2);

	return (0);
}
