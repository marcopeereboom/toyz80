#include <stdio.h>
#include <string.h>
#include <stdint.h>

#include "i8251.h"

/*
 * cross compile
 * works in build environment
 * disassembly
 * memory dump and set
 * proper line editor
 *
 * maybe:
 *	- serial load
 */

const char	*version = "0.1";

#if defined(__SDCC)
#define ECHO_CHAR
#else
uint8_t		memory[64 * 1024];
#endif

#if defined ECHO_CHAR
#define ECHO(c) \
	printf("%c", c);
#else
#define ECHO(c)
#endif

#define LINE_MAX	(80)
void
parse(char *l)
{
	if (!strcmp(l, "moo")) {
		printf("got moo");
	}
}

#if defined (__SDCC)
int
main(int argc, char *argv[])
{
	int c, i= 0, quit = 0;
	char l[LINE_MAX];

	argc; argv; // shut compiler up

	init_console();

	printf("Z80-monitor %s\r\n", version);

	do {
		c = getchar();
		switch (c) {
		case 27:
			quit = 1;
			break;
		// deal with backspace here as well
		case '\n':
			printf("got (%d): %s\n", i, l);
			parse(l);
			i = 0;
			l[i] = '\0';
			break;
		default:
			if (i > LINE_MAX-2) {
				// silently drop for now
				continue;
			}
			l[i++] = c;
			l[i] = '\0';
			ECHO(c);
			break;
		}
	} while (!quit);

	return (0);
}
#else

int
main(int argc, char *argv[])
{
	//char m[]= {0x08, 0x08};
	memset(memory, 0, sizeof(memory));
	memory[0] = 0xe9;
	memory[1] = 0x01;
	memory[2] = 0x10;
	disassemble(0);
	//printf("x %d\n", x);

	//int c, i= 0, quit = 0;
	//char l[LINE_MAX];
	//printf("Z80-monitor %s\n", version);
	//do {
	//	c = getchar();
	//	switch (c) {
	//	// deal with backspace here as well
	//	case '\n':
	//		printf("got (%d): %s\n", i, l);
	//		parse(l);
	//		i = 0;
	//		l[i] = '\0';
	//		break;
	//	default:
	//		if (i > LINE_MAX-2) {
	//			// silently drop for now
	//			continue;
	//		}
	//		l[i++] = c;
	//		l[i] = '\0';
	//		ECHO(c);
	//		break;
	//	}
	//} while (!quit);

	return (0);
}
#endif
