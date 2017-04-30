#include <stdint.h>

#if !defined(__SDCC)
extern char	memory[64 *1024];
#endif

char	disassemble(uint16_t);
