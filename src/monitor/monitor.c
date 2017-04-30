#include <stdio.h>
#include <string.h>
#include <stdlib.h>
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

#define TOKENS_MAX	(4)

void
help(void)
{
	printf("help\tthis help\r\n" \
	    "d\tdisassemble\r\n");
}

#define LINE_MAX	(80)
void
parse(char *l)
{
	char		lines, i, n, x, *p;
	char		*tokens[TOKENS_MAX];
	uint16_t	address;

	x = 0;
	p = strtok(l, " ");

	while (p != NULL && x < TOKENS_MAX) {
		tokens[x] = p;
		p = strtok(NULL, " ");
		x++;
	}

	if (x == 0) {
		return;
	}

	if (!strcmp(tokens[0], "help")) {
		help();
	} else if (!strcmp(tokens[0], "d")) {
		if (x > 1) {
			address = atoi(tokens[1]);
		} else {
			printf("usage: d <address> [count]\r\n");
			return;
		}
		if (x > 2) {
			lines = atoi(tokens[2]);
		} else {
			lines = 1;
		}
		for (i =0; i < lines; i++) {
			n = disassemble(address);
			address += n;
		}
	} else {
		printf("invalid command\r\n");
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
		case 27: // Esc
			quit = 1;
			break;
		case '\r':
			/* FALLTHROUGH */
		case '\n':
			printf("\r\n");

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
