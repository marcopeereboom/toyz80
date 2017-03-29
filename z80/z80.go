package z80

import (
	"errors"
	"fmt"

	"github.com/marcopeereboom/toyz80/bus"
)

var (
	ErrDisassemble = errors.New("could not disassemble")
	ErrHalt        = errors.New("halt")
	//ErrInvalidInstruction = errors.New("invalid instruction")
)

type CPUMode int

const (
	ModeZ80 CPUMode = iota
	Mode8080
)

const (
	carry     uint16 = 1 << 0 // C Carry Flag
	addsub    uint16 = 1 << 1 // N Add/Subtract
	parity    uint16 = 1 << 2 // P/V Parity/Overflow Flag
	unused    uint16 = 1 << 3 // X Not Used
	halfCarry uint16 = 1 << 4 // H Half Carry Flag
	unused2   uint16 = 1 << 5 // X Not Used
	zero      uint16 = 1 << 6 // Z Zero Flag
	sign      uint16 = 1 << 7 // S Sign Flag
)

// z80 describes a z80/8080 CPU.
type z80 struct {
	af  uint16 // A & Flags
	af_ uint16 // A' & Flags'
	bc  uint16 // B & C
	bc_ uint16 // B' & C'
	de  uint16 // D & E
	de_ uint16 // D' & E'
	hl  uint16 // H & L
	hl_ uint16 // H' & L'
	ix  uint16 // index register X
	iy  uint16 // index register Y

	pc uint16 // program counter
	sp uint16 // stack pointer

	iff1 byte // iff1 flip-flop
	iff2 byte // iff2 flip-flop

	bus *bus.Bus // System bus

	totalCycles uint64 // Total cycles used

	mode CPUMode // Mode CPU is running
}

// DumpRegisters returns a dump of all registers.
func (z *z80) DumpRegisters() string {
	flags := ""
	if z.af&sign == sign {
		flags += "S"
	} else {
		flags += "-"
	}
	if z.af&zero == zero {
		flags += "Z"
	} else {
		flags += "-"
	}
	if z.af&unused2 == unused2 {
		flags += "1"
	} else {
		flags += "0"
	}
	if z.af&halfCarry == halfCarry {
		flags += "H"
	} else {
		flags += "-"
	}
	if z.af&unused == unused {
		flags += "1"
	} else {
		flags += "0"
	}
	if z.af&parity == parity {
		flags += "P"
	} else {
		flags += "-"
	}
	if z.af&addsub == addsub {
		flags += "N"
	} else {
		flags += "-"
	}
	if z.af&carry == carry {
		flags += "C"
	} else {
		flags += "-"
	}
	return fmt.Sprintf("af $%04x bc $%04x de $%04x hl $%04x ix $%04x "+
		"iy $%04x pc $%04x sp $%04x f %v ", uint16(z.af),
		uint16(z.bc), uint16(z.de), uint16(z.hl), uint16(z.ix),
		uint16(z.iy), uint16(z.pc), uint16(z.sp), flags)
}

// New returns a cold reset Z80 CPU struct.
func New(mode CPUMode, bus *bus.Bus) (*z80, error) {
	return &z80{
		mode: mode,
		bus:  bus,
	}, nil
}

// Reset resets the CPU.  If cold is true then memory is zeroed.
func (z *z80) Reset(cold bool) {
	if cold {
		// toss memory.
		//z.bus.Reset()
	}

	//The program counter is reset to 0000h
	z.pc = 0

	//Interrupt mode 0.

	//Interrupt are dissabled.

	//The register I = 00h
	//The register R = 00h
}

func (z *z80) evalZ(src byte) {
	if src == 0x00 {
		z.af |= zero
	} else {
		z.af &^= zero
	}
}

func (z *z80) evalS(src byte) {
	if src&0x80 == 0x80 {
		z.af |= sign
	} else {
		z.af &^= sign
	}
}

func (z *z80) evalH(src, increment byte) {
	h := (src&0x0f + increment&0x0f) & 0x10
	if h != 0 {
		z.af |= halfCarry
	} else {
		z.af &^= halfCarry
	}
}

func (z *z80) genericPostInstruction(o *opcode) {
	// panic for now
	if o.noBytes == 0 || o.noCycles == 0 {
		panic(fmt.Sprintf("opcode missing or invalid: pc %04x", z.pc))
	}

	z.totalCycles += o.noCycles
	z.pc += uint16(o.noBytes)
}

// Step executes the instruction as pointed at by PC.
func (z *z80) Step() error {
	// This is a little messy because of multi-byte opcodes.  We assume the
	// opcode is one byte and we change in the switch statement to contain
	// the *actual* opcode in order to calculate cycles etc.
	//
	// Instructions that return early shall handle pc and noCycles.
	//
	// Reference used: http://zilog.com/docs/z80/um0080.pdf
	opc := z.bus.Read(z.pc)
	opcodeStruct := &opcodes[opc]
	pi := z.genericPostInstruction

	// Move all code into opcodes array for extra vroom vroom.
	switch opc {
	case 0x00: // nop
		// nothing to do
	case 0x01: // ld bc,nn
		z.bc = uint16(z.bus.Read(z.pc+1)) | uint16(z.bus.Read(z.pc+2))<<8
	case 0x02: // ld (bc),a
		z.bus.Write(z.bc, byte(z.af>>8))
	case 0x03: // inc bc
		z.bc += 1
	case 0x04: // inc b
		z.bc = uint16(z.inc(byte(z.bc>>8)))<<8 | z.bc&0x00ff
	case 0x05: // dec b
		z.bc = uint16(z.dec(byte(z.bc>>8)))<<8 | z.bc&0x00ff
	case 0x06: // ld b,n
		z.bc = uint16(z.bus.Read(z.pc+1))<<8 | z.bc&0x00ff
	case 0x07: // rlca
		a := z.af >> 8 << 1
		f := z.af&(sign|zero|parity|unused|unused2) | a>>8
		z.af = (a>>8|a)<<8 | f
	case 0x08: // ex af,af'
		t := z.af
		z.af = z.af_
		z.af_ = t
	case 0x09:
		z.hl = z.add16(z.hl, z.bc)
	case 0x0a: // ld a,(bc)
		z.af = uint16(z.bus.Read(z.bc))<<8 | z.af&0x00ff
	case 0x0b: //dec bc
		z.bc -= 1
	case 0x0c: // inc c
		z.bc = uint16(z.inc(byte(z.bc))) | z.bc&0xff00
	case 0x0d: // dec c
		z.bc = uint16(z.dec(byte(z.bc))) | z.bc&0xff00
	case 0x0e: // ld c,n
		z.bc = uint16(z.bus.Read(z.pc+1)) | z.bc&0xff00
	case 0x0f: // rrca
		a := z.af & 0xff00 >> 1
		a = a<<8 | a
		f := z.af&(sign|zero|parity|unused|unused2) |
			uint16(ternB(a&0x80 == 0x80, byte(carry), 0))
		z.af = a&0xff00 | uint16(f)
	case 0x11: // ld de,nn
		z.de = uint16(z.bus.Read(z.pc+1)) | uint16(z.bus.Read(z.pc+2))<<8
	case 0x12: // ld (de),a
		z.bus.Write(z.de, byte(z.af>>8))
	case 0x13: // inc de
		z.de += 1
	case 0x14: // inc d
		z.de = uint16(z.inc(byte(z.de>>8)))<<8 | z.de&0x00ff
	case 0x15: // dec d
		z.de = uint16(z.dec(byte(z.de>>8)))<<8 | z.de&0x00ff
	case 0x16: // ld d,n
		z.de = uint16(z.bus.Read(z.pc+1))<<8 | z.de&0x00ff
	case 0x18: // jr d
		z.pc = z.pc + 2 + uint16(int8(z.bus.Read(z.pc+1)))
		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0x19: // add hl,de
		z.hl = z.add16(z.hl, z.de)
	case 0x1a: // ld a,(de)
		z.af = uint16(z.bus.Read(z.de))<<8 | z.af&0x00ff
	case 0x1b: // dec de
		z.de -= 1
	case 0x1c: // inc e
		z.de = uint16(z.inc(byte(z.de))) | z.de&0xff00
	case 0x1d: // dec e
		z.de = uint16(z.dec(byte(z.de))) | z.de&0xff00
	case 0x1e: // ld e,n
		z.de = uint16(z.bus.Read(z.pc+1)) | z.de&0xff00
	case 0x1f: // rar
		a := z.af&0xff00>>1 | z.af<<15
		f := z.af&(sign|zero|parity) | uint16(ternB(a&0x80 == 0x80,
			byte(carry), 0))
		z.af = a | f
	case 0x21: // ld hl,nn
		z.hl = uint16(z.bus.Read(z.pc+1)) | uint16(z.bus.Read(z.pc+2))<<8
	case 0x22: // ld (nn),hl
		addr := uint16(z.bus.Read(z.pc+1)) | uint16(z.bus.Read(z.pc+2))<<8
		z.bus.Write(addr, byte(z.hl))
		z.bus.Write(addr+1, byte(z.hl>>8))
	case 0x23: // inc hl
		z.hl += 1
	case 0x24: // inc h
		z.hl = uint16(z.inc(byte(z.hl>>8)))<<8 | z.hl&0x00ff
	case 0x25: // dec h
		z.hl = uint16(z.dec(byte(z.hl>>8)))<<8 | z.hl&0x00ff
	case 0x26: // ld h,n
		z.hl = uint16(z.bus.Read(z.pc+1))<<8 | z.hl&0x00ff
	case 0x28: // jr z,d
		if z.af&zero == zero {
			z.pc = z.pc + 2 + uint16(int8(z.bus.Read(z.pc+1)))
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
		// XXX make this generic
		z.totalCycles += 7
	case 0x29: // add hl,hl
		z.hl = z.add16(z.hl, z.hl)
	case 0x2a: // ld (hl),nn
		addr := uint16(z.bus.Read(z.pc+1)) | uint16(z.bus.Read(z.pc+2))<<8
		z.hl = uint16(z.bus.Read(addr)) | uint16(z.bus.Read(addr+1))<<8
	case 0x2b: //dec hl
		z.hl -= 1
	case 0x2c: // inc l
		z.hl = uint16(z.inc(byte(z.hl))) | z.hl&0xff00
	case 0x2d: // dec l
		z.hl = uint16(z.dec(byte(z.hl))) | z.hl&0xff00
	case 0x2e: // ld l,n
		z.hl = uint16(z.bus.Read(z.pc+1)) | z.hl&0xff00
	case 0x2f: // cpl
		z.af = z.af&0x00ff | ^z.af&0xff00

		// Condition Bits Affected
		// S is not affected.
		// Z is not affected.
		// H is set.
		// P/V is not affected.
		// N is set.
		// C is not affected.
		z.af |= halfCarry
		z.af |= addsub
	case 0x31: // ld sp,nn
		z.sp = uint16(z.bus.Read(z.pc+1)) | uint16(z.bus.Read(z.pc+2))<<8
	case 0x32: // ld (nn),a
		z.bus.Write(uint16(z.bus.Read(z.pc+1))|
			uint16(z.bus.Read(z.pc+2))<<8, byte(z.af>>8))
	case 0x33: //inc sp
		z.sp += 1
	case 0x34: // inc (hl)
		z.bus.Write(z.hl, z.inc(z.bus.Read(z.hl)))
	case 0x35: // dec (hl)
		z.bus.Write(z.hl, z.dec(z.bus.Read(z.hl)))
	case 0x36: // ld (hl),n
		z.bus.Write(z.hl, z.bus.Read(z.pc+1))
	case 0x37: // scf
		// Condition Bits Affected
		// S is not affected.
		// Z is not affected.
		// H is reset.
		// P/V is not affected.
		// N is reset.
		// C is set.
		z.af &^= halfCarry
		z.af &^= addsub
		z.af |= carry
	case 0x39: // add hl,sp
		z.hl = z.add16(z.hl, z.sp)
	case 0x3d: // dec a
		z.af = uint16(z.dec(byte(z.af>>8)))<<8 | z.af&0x00ff
	case 0x3f: // ccf
		// Condition Bits Affected
		// S is not affected.
		// Z is not affected.
		// H, previous carry is copied.
		// P/V is not affected.
		// N is reset.
		// C is set if CY was 0 before operation; otherwise, it is reset.
		if z.af&carry == carry {
			z.af |= halfCarry
			z.af &^= carry // invert carry
		} else {
			z.af &^= halfCarry
			z.af |= carry // invert carry
		}
		z.af &^= addsub
	case 0x3a: // ld a,(nn)
		z.af = uint16(z.bus.Read(uint16(z.bus.Read(z.pc+1))|
			uint16(z.bus.Read(z.pc+2))<<8)) << 8
	case 0x3b: //dec sp
		z.sp -= 1
	case 0x3c: // inc a
		z.af = uint16(z.inc(byte(z.af>>8)))<<8 | z.af&0x00ff
	case 0x3e: // ld a,n
		z.af = uint16(z.bus.Read(z.pc+1))<<8 | z.af&0x00ff
	case 0x40: //ld b,b
		// nothing to do
	case 0x41: //ld b,c
		z.bc = z.bc&0x00ff | z.bc<<8
	case 0x42: //ld b,d
		z.bc = z.bc&0x00ff | z.de&0xff00
	case 0x43: //ld b,e
		z.bc = z.bc&0x00ff | z.de<<8
	case 0x44: //ld b,h
		z.bc = z.bc&0x00ff | z.hl&0xff00
	case 0x45: //ld b,l
		z.bc = z.bc&0x00ff | z.hl<<8
	case 0x46: //ld b,(hl)
		z.bc = uint16(z.bus.Read(z.hl))<<8 | z.bc&0x00ff
	case 0x47: // ld b,a
		z.bc = z.af&0xff00 | z.bc&0x00ff
	case 0x48: // ld c,b
		z.bc = z.bc>>8 | z.bc&0xff00
	case 0x49: //ld c,c
		// nothing to do
	case 0x4a: // ld c,d
		z.bc = z.bc&0xff00 | z.de>>8
	case 0x4b: // ld c,e
		z.bc = z.bc&0xff00 | z.de&0x00ff
	case 0x4c: // ld c,h
		z.bc = z.bc&0xff00 | z.hl>>8
	case 0x4d: // ld c,l
		z.bc = z.bc&0xff00 | z.hl&0x00ff
	case 0x4e: //ld c,(hl)
		z.bc = uint16(z.bus.Read(z.hl)) | z.bc&0xff00
	case 0x4f: // ld c,a
		z.bc = z.bc&0xff00 | z.af>>8
	case 0x50: // ld d,b
		z.de = z.bc&0xff00 | z.de&0x00ff
	case 0x51: // ld d,c
		z.de = z.bc<<8 | z.de&0x00ff
	case 0x52: // ld d,d
		// nothing to do
	case 0x53: // ld d,e
		z.de = z.de&0x00ff | z.de<<8
	case 0x54: // ld d,h
		z.de = z.hl&0xff00 | z.de&0x00ff
	case 0x55: // ld d,l
		z.de = z.hl<<8 | z.de&0x00ff
	case 0x56: // ld d,(hl)
		z.de = uint16(z.bus.Read(z.hl))<<8 | z.de&0x00ff
	case 0x57: // ld d,a
		z.de = z.af&0xff00 | z.de&0x00ff
	case 0x58: // ld e,b
		z.de = z.bc>>8 | z.de&0xff00
	case 0x59: // ld e,c
		z.de = z.bc&0x00ff | z.de&0xff00
	case 0x5a: // ld e,d
		z.de = z.de>>8 | z.de&0xff00
	case 0x5b: // ld e,e
		// nothing to do
	case 0x5c: // ld e,h
		z.de = z.hl>>8 | z.de&0xff00
	case 0x5d: // ld e,l
		z.de = z.hl&0x00ff | z.de&0xff00
	case 0x5e: // ld e,(hl)
		z.de = uint16(z.bus.Read(z.hl)) | z.de&0xff00
	case 0x5f: // ld e,a
		z.de = z.af>>8 | z.de&0xff00
	case 0x60: // ld h,b
		z.hl = z.hl&0x00ff | z.bc&0xff00
	case 0x61: // ld h,c
		z.hl = z.hl&0x00ff | z.bc<<8
	case 0x62: // ld h,d
		z.hl = z.hl&0x00ff | z.de&0xff00
	case 0x63: // ld h,e
		z.hl = z.hl&0x00ff | z.de<<8
	case 0x64: // ld h,h
		// nothing to do
	case 0x65: // ld h,l
		z.hl = z.hl&0x00ff | z.hl<<8
	case 0x66: // ld h,(hl)
		z.hl = uint16(z.bus.Read(z.hl))<<8 | z.hl&0x00ff
	case 0x67: // ld h,a
		z.hl = z.hl&0x00ff | z.af&0xff00
	case 0x68: // ld l,b
		z.hl = z.hl&0xff00 | z.bc>>8
	case 0x69: // ld l,c
		z.hl = z.hl&0xff00 | z.bc&0x00ff
	case 0x6a: // ld l,d
		z.hl = z.hl&0xff00 | z.de>>8
	case 0x6b: // ld l,e
		z.hl = z.hl&0xff00 | z.de&0x00ff
	case 0x6c: // ld l,h
		z.hl = z.hl>>8 | z.hl&0xff00
	case 0x6d: // ld l,l
		// nothing to do
	case 0x6e: // ld l,(hl)
		z.hl = uint16(z.bus.Read(z.hl)) | z.hl&0xff00
	case 0x6f: // ld l,a
		z.hl = z.hl&0xff00 | z.af>>8
	case 0x70: // ld (hl),b
		z.bus.Write(z.hl, byte(z.bc>>8))
	case 0x71: // ld (hl),c
		z.bus.Write(z.hl, byte(z.bc))
	case 0x72: // ld (hl),d
		z.bus.Write(z.hl, byte(z.de>>8))
	case 0x73: // ld (hl),e
		z.bus.Write(z.hl, byte(z.de))
	case 0x74: // ld (hl),h
		z.bus.Write(z.hl, byte(z.hl>>8))
	case 0x75: // ld (hl),l
		z.bus.Write(z.hl, byte(z.hl))
	case 0x76: // halt
		z.totalCycles += opcodeStruct.noCycles
		return ErrHalt
	case 0x77: // ld (hl),a
		z.bus.Write(z.hl, byte(z.af>>8))
	case 0x78: // ld a,b
		z.af = z.af&0x00ff | z.bc&0xff00
	case 0x79: // ld a,c
		z.af = z.af&0x00ff | z.bc<<8
	case 0x7a: // ld a,d
		z.af = z.af&0x00ff | z.de&0xff00
	case 0x7b: // ld a,e
		z.af = z.af&0x00ff | z.de<<8
	case 0x7c: // ld a,h
		z.af = z.af&0x00ff | z.hl&0xff00
	case 0x7d: // ld a,l
		z.af = z.af&0x00ff | z.hl<<8
	case 0x7e: // ld a,(hl)
		z.af = uint16(z.bus.Read(z.hl))<<8 | z.af&0x00ff
	case 0x7f: // ld a,a
		// nothing to do
	case 0x80: // add a,b
		z.add(byte(z.bc >> 8))
	case 0x81: // add a,c
		z.add(byte(z.bc))
	case 0x82: // add a,d
		z.add(byte(z.de >> 8))
	case 0x83: // add a,e
		z.add(byte(z.de))
	case 0x84: // add a,h
		z.add(byte(z.hl >> 8))
	case 0x85: // add a,l
		z.add(byte(z.hl))
	case 0x86: // add a,(hl)
		z.add(z.bus.Read(z.hl))
	case 0x87: // add a,a
		z.add(byte(z.af >> 8))
	case 0x88: // adc a,b
		z.adc(byte(z.bc >> 8))
	case 0x89: // adc a,c
		z.adc(byte(z.bc))
	case 0x8a: // adc a,d
		z.adc(byte(z.de) >> 8)
	case 0x8b: // adc a,e
		z.adc(byte(z.de))
	case 0x8c: // adc a,h
		z.adc(byte(z.hl >> 8))
	case 0x8d: // adc a,l
		z.adc(byte(z.hl))
	case 0x8e: // adc a,(hl)
		z.adc(z.bus.Read(z.hl))
	case 0x8f: // adc a,a
		z.adc(byte(z.af >> 8))
	case 0x90: // sub a,b
		z.sub(byte(z.bc >> 8))
	case 0x91: // sub a,c
		z.sub(byte(z.bc))
	case 0x92: // sub a,d
		z.sub(byte(z.de >> 8))
	case 0x93: // sub a,e
		z.sub(byte(z.de))
	case 0x94: // sub a,h
		z.sub(byte(z.hl >> 8))
	case 0x95: // sub a,l
		z.sub(byte(z.hl))
	case 0x96: // sub a,(hl)
		z.sub(z.bus.Read(z.hl))
	case 0x97: // sub a
		z.sub(byte(z.af >> 8))
	case 0x98: // sbc a,b
		z.sbc(byte(z.bc >> 8))
	case 0x99: // sbc a,c
		z.sbc(byte(z.bc))
	case 0x9a: // sbc a,d
		z.sbc(byte(z.de >> 8))
	case 0x9b: // sbc a,e
		z.sbc(byte(z.de))
	case 0x9c: // sbc a,h
		z.sbc(byte(z.hl >> 8))
	case 0x9d: // sbc a,l
		z.sbc(byte(z.hl))
	case 0x9e: // sbc a,(hl)
		z.sbc(z.bus.Read(z.hl))
	case 0x9f: // sbc a
		z.sbc(byte(z.af >> 8))
	case 0xa0: // and a,b
		z.and(byte(z.bc >> 8))
	case 0xa1: // and a,c
		z.and(byte(z.bc))
	case 0xa2: // and a,d
		z.and(byte(z.de >> 8))
	case 0xa3: // and a,e
		z.and(byte(z.de))
	case 0xa4: // and a,h
		z.and(byte(z.hl >> 8))
	case 0xa5: // and a,l
		z.and(byte(z.hl))
	case 0xa7: // and a,a
		z.and(byte(z.af >> 8))
	case 0xa8: // xor a,b
		z.xor(byte(z.bc >> 8))
	case 0xa9: // xor a,c
		z.xor(byte(z.bc))
	case 0xaa: // xor a,d
		z.xor(byte(z.de >> 8))
	case 0xab: // xor a,d
		z.xor(byte(z.de))
	case 0xac: // xor a,h
		z.xor(byte(z.hl >> 8))
	case 0xad: // xor a,l
		z.xor(byte(z.hl))
	case 0xae: // xor a,(hl)
		z.xor(z.bus.Read(z.hl))
	case 0xaf: // xor a
		z.xor(byte(z.af >> 8))
	case 0xb0: // or b
		z.or(byte(z.bc >> 8))
	case 0xb1: // or c
		z.or(byte(z.bc))
	case 0xb2: // or d
		z.or(byte(z.de >> 8))
	case 0xb3: // or e
		z.or(byte(z.de))
	case 0xb4: // or h
		z.or(byte(z.hl >> 8))
	case 0xb5: // or l
		z.or(byte(z.hl))
	case 0xb6: // or (hl)
		z.or(z.bus.Read(z.hl))
	case 0xb7: // or a
		z.or(byte(z.af >> 8))
	case 0xb8: // cp b
		z.cp(byte(z.bc >> 8))
	case 0xb9: // cp c
		z.cp(byte(z.bc))
	case 0xba: // cp d
		z.cp(byte(z.de >> 8))
	case 0xbb: // cp e
		z.cp(byte(z.de))
	case 0xbc: // cp h
		z.cp(byte(z.hl >> 8))
	case 0xbd: // cp h
		z.cp(byte(z.hl))
	case 0xbe: // cp (hl)
		z.cp(z.bus.Read(z.hl))
	case 0xbf: // cp a
		z.cp(byte(z.af >> 8))
	case 0xc0: // ret nz
		if z.af&zero == 0 {
			pc := uint16(z.bus.Read(z.sp))
			z.sp++
			pc = uint16(z.bus.Read(z.sp))<<8 | pc&0x00ff
			z.sp++
			z.totalCycles += 11 // XXX
			z.pc = pc
			return nil
		}
	case 0xc1: // pop bc
		z.bc = uint16(z.bus.Read(z.sp)) | z.bc&0xff00
		z.sp++
		z.bc = uint16(z.bus.Read(z.sp))<<8 | z.bc&0x00ff
		z.sp++
	case 0xc2: // jp nz,nn
		if z.af&zero == 0 {
			z.pc = uint16(z.bus.Read(z.pc+1)) |
				uint16(z.bus.Read(z.pc+2))<<8
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
	case 0xc3: // jp nn
		z.pc = uint16(z.bus.Read(z.pc+1)) |
			uint16(z.bus.Read(z.pc+2))<<8
		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0xc5: // push bc
		z.sp--
		z.bus.Write(z.sp, byte(z.bc>>8))
		z.sp--
		z.bus.Write(z.sp, byte(z.bc))
	case 0xc4: //call nz
		if z.af&zero == 0 {
			retPC := z.pc + opcodeStruct.noBytes
			z.sp--
			z.bus.Write(z.sp, byte(retPC>>8))
			z.sp--
			z.bus.Write(z.sp, byte(retPC))

			z.pc = uint16(z.bus.Read(z.pc+1)) |
				uint16(z.bus.Read(z.pc+2))<<8

			z.totalCycles += 17
			return nil
		}
	case 0xc6: // add a,i
		z.add(z.bus.Read(z.pc + 1))
	case 0xc7: // rst $0
		retPC := z.pc + opcodeStruct.noBytes
		z.sp--
		z.bus.Write(z.sp, byte(retPC>>8))
		z.sp--
		z.bus.Write(z.sp, byte(retPC))

		z.pc = 0x00

		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0xc8: // ret z
		if z.af&zero == zero {
			pc := uint16(z.bus.Read(z.sp))
			z.sp++
			pc = uint16(z.bus.Read(z.sp))<<8 | pc&0x00ff
			z.sp++
			z.totalCycles += 11 // XXX
			z.pc = pc
			return nil
		}
	case 0xc9: // ret
		pc := uint16(z.bus.Read(z.sp))
		z.sp++
		pc = uint16(z.bus.Read(z.sp))<<8 | pc&0x00ff
		z.sp++
		z.totalCycles += opcodeStruct.noCycles
		z.pc = pc
		return nil
	case 0xca: // jp z,nn
		if z.af&zero == zero {
			z.pc = uint16(z.bus.Read(z.pc+1)) |
				uint16(z.bus.Read(z.pc+2))<<8
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
	case 0xcb: // z80 only
		byte2 := z.bus.Read(z.pc + 1)
		opcodeStruct = &opcodesCB[byte2]
		switch byte2 {
		case 0x27: // sla a
			z.af = uint16(z.sla(byte(z.af>>8)))<<8 | z.af&0x00ff
		case 0x3f: // srl a
			z.af = uint16(z.srl(byte(z.af>>8)))<<8 | z.af&0x00ff
		default:
			return fmt.Errorf("invalid instruction: 0x%02x "+
				"0x%02x @ 0x%04x", opc, z.bus.Read(z.pc+1),
				z.pc)
			//return ErrInvalidInstruction
		}
	case 0xcd: //call nn
		retPC := z.pc + opcodeStruct.noBytes
		z.sp--
		z.bus.Write(z.sp, byte(retPC>>8))
		z.sp--
		z.bus.Write(z.sp, byte(retPC))

		z.pc = uint16(z.bus.Read(z.pc+1)) |
			uint16(z.bus.Read(z.pc+2))<<8

		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0xce: // adc a,i
		z.adc(z.bus.Read(z.pc + 1))
	case 0xcf: // rst $08
		retPC := z.pc + opcodeStruct.noBytes
		z.sp--
		z.bus.Write(z.sp, byte(retPC>>8))
		z.sp--
		z.bus.Write(z.sp, byte(retPC))

		z.pc = 0x08

		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0xd0: // ret nc
		if z.af&carry == 0 {
			pc := uint16(z.bus.Read(z.sp))
			z.sp++
			pc = uint16(z.bus.Read(z.sp))<<8 | pc&0x00ff
			z.sp++
			z.totalCycles += 11 // XXX
			z.pc = pc
			return nil
		}
	case 0xd1: // pop de
		z.de = uint16(z.bus.Read(z.sp)) | z.de&0xff00
		z.sp++
		z.de = uint16(z.bus.Read(z.sp))<<8 | z.de&0x00ff
		z.sp++
	case 0xd2: // jp nc,nn
		if z.af&carry == 0 {
			z.pc = uint16(z.bus.Read(z.pc+1)) |
				uint16(z.bus.Read(z.pc+2))<<8
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
	case 0xd3: // out (n), a
		z.bus.IOWrite(z.bus.Read(z.pc+1), byte(z.af>>8))
	case 0xd5: // push de
		z.sp--
		z.bus.Write(z.sp, byte(z.de>>8))
		z.sp--
		z.bus.Write(z.sp, byte(z.de))
	case 0xd6: // sub i
		z.sub(z.bus.Read(z.pc + 1))
	case 0xd7: // rst $10
		retPC := z.pc + opcodeStruct.noBytes
		z.sp--
		z.bus.Write(z.sp, byte(retPC>>8))
		z.sp--
		z.bus.Write(z.sp, byte(retPC))

		z.pc = 0x10

		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0xd8: // ret c
		if z.af&carry == carry {
			pc := uint16(z.bus.Read(z.sp))
			z.sp++
			pc = uint16(z.bus.Read(z.sp))<<8 | pc&0x00ff
			z.sp++
			z.totalCycles += 11 // XXX
			z.pc = pc
			return nil
		}
	case 0xda: // jp c,nn
		if z.af&carry == carry {
			z.pc = uint16(z.bus.Read(z.pc+1)) |
				uint16(z.bus.Read(z.pc+2))<<8
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
	case 0xdb: // in a,(n)
		z.af = uint16(z.bus.IORead(z.bus.Read(z.pc+1)))<<8 | z.af&0x00ff
	case 0xdd: // z80 only
		byte2 := z.bus.Read(z.pc + 1)
		opcodeStruct = &opcodesDD[byte2]
		switch byte2 {
		case 0x23: // inc ix
			z.ix += 1
		case 0x2b: // dec ix
			z.ix -= 1
		case 0xe1: // pop ix
			z.ix = uint16(z.bus.Read(z.sp)) | z.ix&0xff00
			z.sp++
			z.ix = uint16(z.bus.Read(z.sp))<<8 | z.ix&0x00ff
			z.sp++
		case 0xe5: // push ix
			z.sp--
			z.bus.Write(z.sp, byte(z.ix>>8))
			z.sp--
			z.bus.Write(z.sp, byte(z.ix))
		default:
			return fmt.Errorf("invalid instruction: 0x%02x "+
				"0x%02x @ 0x%04x", opc, z.bus.Read(z.pc+1),
				z.pc)
			//return ErrInvalidInstruction
		}
	case 0xdf: // rst $18
		retPC := z.pc + opcodeStruct.noBytes
		z.sp--
		z.bus.Write(z.sp, byte(retPC>>8))
		z.sp--
		z.bus.Write(z.sp, byte(retPC))

		z.pc = 0x18

		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0xe1: // pop hl
		z.hl = uint16(z.bus.Read(z.sp)) | z.hl&0xff00
		z.sp++
		z.hl = uint16(z.bus.Read(z.sp))<<8 | z.hl&0x00ff
		z.sp++
	case 0xe2: // jp po,nn
		if z.af&parity == 0 {
			z.pc = uint16(z.bus.Read(z.pc+1)) |
				uint16(z.bus.Read(z.pc+2))<<8
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
	case 0xe3: // ex (sp),hl
		h := z.bus.Read(z.sp + 1)
		l := z.bus.Read(z.sp)
		z.bus.Write(z.sp+1, byte(z.hl>>8))
		z.bus.Write(z.sp, byte(z.hl))
		z.hl = uint16(h)<<8 | uint16(l)
	case 0xe5: // push hl
		z.sp--
		z.bus.Write(z.sp, byte(z.hl>>8))
		z.sp--
		z.bus.Write(z.sp, byte(z.hl))
	case 0xe6: // and n
		z.and(z.bus.Read(z.pc + 1))
	case 0xe7: // rst $20
		retPC := z.pc + opcodeStruct.noBytes
		z.sp--
		z.bus.Write(z.sp, byte(retPC>>8))
		z.sp--
		z.bus.Write(z.sp, byte(retPC))

		z.pc = 0x20

		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0xe9: // jp (hl)
		// but we don't dereference, *sigh* zilog
		z.pc = z.hl
		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0xea: // jp pe,nn
		if z.af&parity == parity {
			z.pc = uint16(z.bus.Read(z.pc+1)) |
				uint16(z.bus.Read(z.pc+2))<<8
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
	case 0xeb: // ex de,hl
		t := z.hl
		z.hl = z.de
		z.de = t
	case 0xed: // z80 only
		byte2 := z.bus.Read(z.pc + 1)
		opcodeStruct = &opcodesED[byte2]
		switch byte2 {
		case 0x44: // neg
			// Condition Bits Affected
			// S is set if result is negative; otherwise, it is reset.
			// Z is set if result is 0; otherwise, it is reset.
			// H is set if borrow from bit 4; otherwise, it is reset.
			// P/V is set if Accumulator was 80h before operation; otherwise, it is reset.
			// N is set.
			// C is set if Accumulator was not 00h before operation; otherwise, it is reset.
			oldA := byte(z.af & 0xff00 >> 8)
			newA := 0 - oldA
			z.af = uint16(newA) << 8
			z.evalS(newA)
			z.evalZ(newA)
			// XXX figure out how to handle the H flag
			if oldA == 0x80 {
				z.af |= parity
			} else {
				z.af &^= parity
			}
			z.af |= addsub
			if oldA != 0x00 {
				z.af |= carry
			} else {
				z.af &^= carry
			}
		default:
			return fmt.Errorf("invalid instruction: 0x%02x "+
				"0x%02x @ 0x%04x", opc, z.bus.Read(z.pc+1),
				z.pc)
			//return ErrInvalidInstruction
		}
	case 0xef: // rst $28
		retPC := z.pc + opcodeStruct.noBytes
		z.sp--
		z.bus.Write(z.sp, byte(retPC>>8))
		z.sp--
		z.bus.Write(z.sp, byte(retPC))

		z.pc = 0x28

		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0xf0: // ret p
		if z.af&parity == parity {
			pc := uint16(z.bus.Read(z.sp))
			z.sp++
			pc = uint16(z.bus.Read(z.sp))<<8 | pc&0x00ff
			z.sp++
			z.totalCycles += 11 // XXX
			z.pc = pc
			return nil
		}
	case 0xf1: // pop af
		z.af = uint16(z.bus.Read(z.sp)) | z.af&0xff00
		z.sp++
		z.af = uint16(z.bus.Read(z.sp))<<8 | z.af&0x00ff
		z.sp++
	case 0xf2: // jp p,nn
		if z.af&sign == 0 {
			z.pc = uint16(z.bus.Read(z.pc+1)) |
				uint16(z.bus.Read(z.pc+2))<<8
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
	case 0xf3: // di
		z.iff1 = 0
		z.iff2 = 0
	case 0xf5: // push af
		z.sp--
		z.bus.Write(z.sp, byte(z.af>>8))
		z.sp--
		z.bus.Write(z.sp, byte(z.af))
	case 0xf6: // or n
		z.or(z.bus.Read(z.pc + 1))
	case 0xf7: // rst $30
		retPC := z.pc + opcodeStruct.noBytes
		z.sp--
		z.bus.Write(z.sp, byte(retPC>>8))
		z.sp--
		z.bus.Write(z.sp, byte(retPC))

		z.pc = 0x30

		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0xfa: // jp m,nn
		if z.af&sign == sign {
			z.pc = uint16(z.bus.Read(z.pc+1)) |
				uint16(z.bus.Read(z.pc+2))<<8
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
	case 0xfb: // ei
		z.iff1 = 1
		z.iff2 = 1
	case 0xfd: // z80 only
		byte2 := z.bus.Read(z.pc + 1)
		opcodeStruct = &opcodesFD[byte2]
		switch byte2 {
		case 0x23: // inc iy
			z.iy += 1
		case 0x2b: // dec iy
			z.iy -= 1
		case 0xe1: // pop iy
			z.iy = uint16(z.bus.Read(z.sp)) | z.iy&0xff00
			z.sp++
			z.iy = uint16(z.bus.Read(z.sp))<<8 | z.iy&0x00ff
			z.sp++
		case 0xe5: // push iy
			z.sp--
			z.bus.Write(z.sp, byte(z.iy>>8))
			z.sp--
			z.bus.Write(z.sp, byte(z.iy))
		default:
			return fmt.Errorf("invalid instruction: 0x%02x "+
				"0x%02x @ 0x%04x", opc, z.bus.Read(z.pc+1),
				z.pc)
			//return ErrInvalidInstruction
		}
	case 0xfe: // cp i
		z.cp(z.bus.Read(z.pc + 1))
	case 0xff: // rst $38
		retPC := z.pc + opcodeStruct.noBytes
		z.sp--
		z.bus.Write(z.sp, byte(retPC>>8))
		z.sp--
		z.bus.Write(z.sp, byte(retPC))

		z.pc = 0x38

		z.totalCycles += opcodeStruct.noCycles
		return nil
	default:
		//fmt.Printf("opcode %x\n", opcode)
		//return ErrInvalidInstruction
		// XXX make this a generic ErrInvalidInstruction
		return fmt.Errorf("invalid instruction: 0x%02x @ 0x%04x", opc, z.pc)
	}

	pi(opcodeStruct)

	return nil
}

// Disassemble disassembles the instruction at the provided address and also
// returns the address and the number of bytes consumed.
func (z *z80) Disassemble(address uint16, loud bool) (string, uint16, int, error) {
	mn, dst, src, opc, noBytes, err := z.DisassembleComponents(address)

	if dst != "" && src != "" {
		src = "," + src
	}
	dst += src

	var s string
	if loud {
		s = fmt.Sprintf("%-12v%-6v%-4v", opc, mn, dst)
	} else {
		s = fmt.Sprintf("%-6v%-4v", mn, dst)
	}
	return s, address, noBytes, err
}

// Disassemble disassembles the instruction at the current program counter.
func (z *z80) DisassemblePC(loud bool) (string, uint16, int, error) {
	return z.Disassemble(z.pc, loud)
}

// DisassembleComponents disassmbles the instruction at the provided address
// and returns all compnonts of the instruction (opcode, destination, source).
func (z *z80) DisassembleComponents(address uint16) (mnemonic string, dst string, src string, opc string, noBytes int, retErr error) {
	p := make([]byte, 4)
	p[0] = z.bus.Read(address)
	o := &opcodes[p[0]]
	start := uint16(1)
	if o.multiByte {
		p[1] = z.bus.Read(address + 1)
		switch p[0] {
		case 0xcb:
			o = &opcodesCB[p[1]]
		case 0xdd:
			o = &opcodesDD[p[1]]
		case 0xed:
			o = &opcodesED[p[1]]
		case 0xfd:
			o = &opcodesFD[p[1]]
		}
		start = 2
	}

	// get remaining bytes.
	for i := start; i < uint16(o.noBytes); i++ {
		p[i] = z.bus.Read(address + i)
	}

	switch o.dst {
	case condition:
		dst = o.dstR[z.mode]
	case displacement:
		dst = fmt.Sprintf("$%04x", address+2+uint16(int8(p[1])))
	case registerIndirect:
		if z.mode == Mode8080 {
			dst = fmt.Sprintf("%v", o.dstR[z.mode])
		} else {
			dst = fmt.Sprintf("(%v)", o.dstR[z.mode])
		}
	case extended:
		dst = fmt.Sprintf("($%04x)", uint16(p[1])|uint16(p[2])<<8)
	case immediate:
		dst = fmt.Sprintf("$%02x", p[1])
	case immediateExtended:
		dst = fmt.Sprintf("$%04x", uint16(p[1])|uint16(p[2])<<8)
	case register:
		dst = o.dstR[z.mode]
	case indirect:
		dst = fmt.Sprintf("($%02x)", p[1])
	case implied:
		dst = o.dstR[z.mode]
	}

	switch o.src {
	case displacement:
		src = fmt.Sprintf("$%04x", address+2+uint16(int8(p[1])))
	case registerIndirect:
		if z.mode == Mode8080 {
			src = fmt.Sprintf("%v", o.srcR[z.mode])
		} else {
			src = fmt.Sprintf("(%v)", o.srcR[z.mode])
		}
	case extended:
		src = fmt.Sprintf("($%04x)", uint16(p[1])|uint16(p[2])<<8)
	case immediate:
		src = fmt.Sprintf("$%02x", p[1])
	case immediateExtended:
		src = fmt.Sprintf("$%04x", uint16(p[1])|uint16(p[2])<<8)
	case register:
		src = o.srcR[z.mode]
	case indirect:
		src = fmt.Sprintf("($%02x)", p[1])
	}

	noBytes = int(o.noBytes)
	retErr = nil
	if len(o.mnemonic) == 0 {
		switch p[0] {
		case 0xcb, 0xdd, 0xed, 0xfd:
			mnemonic = fmt.Sprintf("%02x %02x", p[0], p[1])
			noBytes = 2
		default:
			mnemonic = fmt.Sprintf("%02x", p[0])
			noBytes = 1
		}
		retErr = fmt.Errorf("%04x: %v INVALID", address, mnemonic)
	} else {
		switch noBytes {
		case 1:
			opc = fmt.Sprintf("%02x", p[0])
		case 2:
			opc = fmt.Sprintf("%02x %02x", p[0], p[1])
		case 3:
			opc = fmt.Sprintf("%02x %02x %02x", p[0], p[1], p[2])
		case 4:
			opc = fmt.Sprintf("%02x %02x %02x %02x", p[0], p[1],
				p[2], p[3])
		default:
			opc = "OPCINV"
		}
		mnemonic = o.mnemonic[z.mode]
	}

	return
}

func (z *z80) Trace() ([]string, []string, error) {
	trace := make([]string, 0, 1024)
	registers := make([]string, 0, 1024)

	for {
		s, _, _, err := z.Disassemble(z.pc, true)
		trace = append(trace, fmt.Sprintf("%04x: %v", z.pc, s))
		if err != nil {
			return trace, registers, err
		}
		//fmt.Printf("%04x: %v\n", z.pc, s)
		err = z.Step()
		registers = append(registers, z.DumpRegisters())
		//fmt.Printf("\t%v\n", z.DumpRegisters())
		if err != nil {
			return trace, registers, err
		}
	}
	return trace, registers, nil
}
