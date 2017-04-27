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

#define	I8251_S_RXRDY			(1<<1)

void	init_console(void);
void	putchar(unsigned char);
