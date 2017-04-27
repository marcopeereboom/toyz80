#include "i8251.h"

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

// override getchar from stdlib
int
getchar(void)
{
__asm
l1:	in	a,(I8251_STATUS)	// get status
	and	#I8251_S_RXRDY		// check RxRDY bit
	jp	z,l1			// not ready, loop
	in	a,(I8251_DATA)		// get char
	ld	h,#0			// reset high return value
	ld	l,a			// set low return value
__endasm;
	// we eat the warning to save 3 bytes.
}

// Can't be wrapped because of sdcc.
#define I8251_MODE_DEFAULT	(I8251_MODE_STOP1|I8251_MODE_PARITYDISABLE|I8251_MODE_BITS8|I8251_MODE_BAUD9600)
#define I8251_CMD_DEFAULT	(I8251_CMD_TXEN|I8251_CMD_RXEN|I8251_CMD_DTR|I8251_CMD_RTS|I8251_CMD_ER)
void
init_console(void)
{
__asm
	ld	a,#I8251_MODE_DEFAULT	// initialize 8251A UART
	out	(I8251_STATUS),a	// 1 stop bit, no parity, 8-bit
	ld	a,#I8251_CMD_DEFAULT	// enable rx/tx/dtr/rts
	out	(I8251_STATUS),a	// and reset error flag
__endasm;
}
