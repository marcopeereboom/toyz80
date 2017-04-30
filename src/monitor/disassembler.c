#include <stdio.h>
#include <string.h>
#include <stdint.h>

#define X_SHIFT		(6)
#define Y_SHIFT		(3)

#define displacement(a, b)	(uint16_t)(a+2+(int8_t)(b))

#if !defined (__SDCC)
uint8_t		memory[64 * 1024];
#endif

const char *alu[] = {"add", "adc", "sub", "sbc", "and", "xor", "or", "cp"};
const char *r[] = {"b", "c", "d", "e", "h", "l", "(hl)", "a"};
const char *rp[] = {"bc", "de", "hl", "sp"};
const char *rp2[] = {"bc", "de", "hl", "af"};
const char *cc[] = {"nz", "z", "nc", "c", "po", "pe", "p", "m"};

// often used menmonics
const char *add = "add";
const char *call = "call";
const char *dec = "dec";
const char *ex = "ex";
const char *in = "in";
const char *inc = "inc";
const char *jp = "jp";
const char *jr = "jr";
const char *ld = "ld";
const char *out = "out";
const char *ret = "ret";

// disassemble disassembles address and returns number of bytes consumed.  A
// return of 0 means an error occured.  Unknown opcodes shall return 1 byte so
// that the disassembler can continue.
//
// This function does a couple of odd things in order to remain as small as
// possible.
char
disassemble(uint16_t address)
{
	char		scratch[24+1];
	char		hex[12+1];
	const char	*mnemonic = "unknown";
	const char	*operands = &scratch[0];
	char		bytes = 1;
	uint8_t		*m;
	char		p, q, x, y, z; // decoding values
	int		i;

#if defined (__SDCC)
	m = (uint8_t *)address;
#else
	m = &memory[address];
#endif
	scratch[0] = 0;

	/*
	 * http://www.z80.info/decoding.htm
	 *
	 * First byte bits:
	 * 7 6 5 4 3 2 1 0
	 * x x y y y z z z
	 *     p p q
	 */
	x = (m[0]&0xc0) >> X_SHIFT;
	y = (m[0]&0x38) >> Y_SHIFT;
	z = m[0]&0x07;
	q = y&0x01;
	p = y>>1;

	switch (x) {
	case 0:
		switch (z) {
		case 0: // Relative jumps and assorted ops
			switch (y) {
			case 0: // nop
				mnemonic = "nop";
				break;
			case 1: // ex af,af'
				mnemonic = ex;
				operands = "af,af'";
				break;
			case 2: // djnz
				mnemonic = "djnz";
				sprintf(&scratch[0], "$%04x",
				    displacement(address, m[1]));
				bytes = 2;
				break;
			case 3: // jr
				mnemonic = jr;
				sprintf(&scratch[0], "$%04x",
				    displacement(address, m[1]));
				bytes = 2;
				break;
			default: // jr cc
				mnemonic = jr;
				sprintf(&scratch[0], "%s,$%04x",
				    cc[y-4],
				    displacement(address, m[1]));
				bytes = 2;
				break;
			}
			break;
		case 1: // 16-bit load immediate/add
			switch (q) {
			case 0:
				mnemonic = ld;
				sprintf(&scratch[0], "%s,$%04x",
				    rp[p], m[1] | ((uint16_t)m[2])<<8);
				bytes = 3;
				break;
			case 1:
				mnemonic = add;
				sprintf(&scratch[0], "hl,%s", rp[p]);
				break;
			}
			break;
		case 2: // Indirect loading
			mnemonic = ld;
			switch (q) {
			case 0:
				switch (p) {
				case 0:
					operands = "(bc),a";
					break;
				case 1:
					operands = "(de),a";
					break;
				case 2:
					sprintf(&scratch[0], "($%04x),hl",
					    m[1] | ((uint16_t)m[2])<<8);
					bytes = 3;
					break;
				case 3:
					sprintf(&scratch[0], "($%04x),a",
					    m[1] | ((uint16_t)m[2])<<8);
					bytes = 3;
					break;
				}
				break;
			case 1:
				switch (p) {
				case 0:
					operands = "a,(bc)";
					break;
				case 1:
					operands = "a,(de)";
					break;
				case 2:
					sprintf(&scratch[0], "hl,($%04x)",
					    m[1] | ((uint16_t)m[2])<<8);
					bytes = 3;
					break;
				case 3:
					sprintf(&scratch[0], "a,($%04x)",
					    m[1] | ((uint16_t)m[2])<<8);
					bytes = 3;
					break;
				}
			}
			break;
		case 3: // 16-bit INC/DEC
			operands = rp[p];
			switch (q) {
			case 0:
				mnemonic = inc;
				break;
			case 1:
				mnemonic = dec;
				break;
			}
			break;
		case 4: // 8-bit INC
			mnemonic = inc;
			operands = r[y];
			break;
		case 5: // 8-bit DEC
			mnemonic = dec;
			operands = r[y];
			break;
		case 6: // 8-bit load immediate
			mnemonic = ld;
			sprintf(&scratch[0], "%s,$%02x", r[y], m[1]);
			bytes = 2;
			break;
		case 7:
			switch (y) {
			case 0:
				mnemonic = "rlca";
				break;
			case 1:
				mnemonic = "rrca";
				break;
			case 2:
				mnemonic = "rla";
				break;
			case 3:
				mnemonic = "rra";
				break;
			case 4:
				mnemonic = "daa";
				break;
			case 5:
				mnemonic = "cpl";
				break;
			case 6:
				mnemonic = "scf";
				break;
			case 7:
				mnemonic = "ccf";
				break;
			}
			break;
		}
		break;
	case 1:
		switch (z) {
		case 6: // Exception (replaces LD (HL), (HL))
			mnemonic = "halt";
			break;
		default: // 8-bit loading
			mnemonic = ld;
			sprintf(&scratch[0], "%s,%s",r[y], r[z]);
			break;
		}
		break;
	case 2: // Operate on accumulator and register/memory location
		// some mnemonics prefix with a,; we ignore that; all or none!
		mnemonic = alu[y];
		operands = r[z];
		break;
	case 3:
		switch (z) {
		case 0: // Conditional return
			mnemonic = ret;
			operands = cc[y];
			break;
		case 1: // POP & various ops
			switch (q) {
			case 0:
				mnemonic = "pop";
				operands = rp2[p];
				break;
			case 1:
				switch (p) {
				case 0:
					mnemonic = ret;
					break;
				case 1:
					mnemonic = "exx";
					break;
				case 2:
					// zilog says (hl) but it is hl
					mnemonic = jp;
					operands = "hl";
					break;
				case 3:
					mnemonic = ld;
					operands = "sp,hl";
					break;
				}
				break;
			}
			break;
		case 2: // Conditional jump
			mnemonic = jp;
			sprintf(&scratch[0], "%s,$%04x",cc[y],
			    m[1] | ((uint16_t)m[2])<<8);
			bytes = 3;
			break;
		case 3: // Assorted operations
			switch (y) {
			case 0:
				mnemonic = jp;
				sprintf(&scratch[0], "$%04x",
				    m[1] | ((uint16_t)m[2])<<8);
				bytes = 3;
				break;
			case 1:
					// XXX CB prefix
					goto bad;
				break;
			case 2:
				mnemonic = out;
				sprintf(&scratch[0], "($%02x),a", m[1]);
				bytes = 2;
				break;
			case 3:
				mnemonic = in;
				sprintf(&scratch[0], "a,($%02x)", m[1]);
				bytes = 2;
				break;
			case 4:
				mnemonic = ex;
				operands = "(sp),hl";
				break;
			case 5:
				mnemonic = ex;
				operands = "de,hl";
				break;
			case 6:
				mnemonic = "di";
				break;
			case 7:
				mnemonic = "ei";
				break;
			}
			break;
		case 4: // Conditional call
			mnemonic = call;
			sprintf(&scratch[0], "%s,$%04x", cc[y],
			    m[1] | ((uint16_t)m[2])<<8);
			bytes = 3;
			break;
		case 5: // PUSH & various ops
			switch (q) {
			case 0:
				mnemonic = "push";
				operands = rp2[p];
				break;
			case 1:
				switch (p) {
				case 0:
					mnemonic = call;
					sprintf(&scratch[0], "$%04x",
					    m[1] | ((uint16_t)m[2])<<8);
					bytes = 3;
					break;
				case 1:
					// XXX DD prefix
					goto bad;
					break;
				case 2:
					// XXX ED prefix
					goto bad;
					break;
				case 3:
					// XXX FD prefix
					goto bad;
					break;
				}
				break;
			}
			break;
		case 6: // Operate on accumulator and immediate operand
			mnemonic = alu[y];
			sprintf(&scratch[0], "$%02x", m[1]);
			bytes = 2;
			break;
		case 7: // Restart
			mnemonic = "rst";
			sprintf(&scratch[0], "$%02x", y<<3);
			break;
		}
		break;
	default:
	bad:
		sprintf(&scratch[0], "x $%02x y $%02x z $%02x", x, y, z);
		break;
	}

	// calculate hex
	for (i = 0; i < bytes; i++) {
		sprintf(&hex[i*3], "%02x ", m[i]);
	}

	printf("%04x: %-12s%-12s%s\r\n", address, hex, mnemonic, operands);
	return (bytes);
}

