#include <stdio.h>
#include <string.h>
#include <stdint.h>

#include "disassembler.h"

#if defined(__SDCC)
#include "i8251.h"

#define ECHO_CHAR
#endif

#if defined ECHO_CHAR
#define ECHO(c) \
	printf("%c", c);
#else
#define ECHO(c)
#endif

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

#define LINE_MAX	(80)
void
parse(char *l)
{
	if (!strcmp(l, "moo")) {
		printf("got moo");
	}
}

void
monitor(void)
{
	char l[LINE_MAX];
	int c, i= 0, quit = 0;

	printf("Z80-monitor %s\r\n", version);

	do {
		// getchar on UNIX is buffered but we ignore that.
		c = getchar();
		switch (c) {
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
}

int
main(int argc, char *argv[])
{
#if defined (__SDCC)
	argc; argv; // shut compiler up
	init_console();
#else
	memset(memory, 0, sizeof(memory));
#endif

	monitor();

	return (0);
}
