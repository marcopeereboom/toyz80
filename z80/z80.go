package z80

import (
	"errors"
	"fmt"

	"github.com/marcopeereboom/toyz80/bus"
)

var (
	ErrDisassemble = errors.New("could not disassemble")
	//ErrHalt        = errors.New("halt")
	//ErrInvalidInstruction = errors.New("invalid instruction")
)

type BreakpointError struct {
	PC       uint16
	Callback func() error
}

func (bp BreakpointError) Error() string {
	return fmt.Sprintf("breakpoint: $%04x", bp.PC)
}

type HaltError struct {
	PC uint16
}

func (hp HaltError) Error() string {
	return fmt.Sprintf("halt: $%04x", hp.PC)
}

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

	debug bool                    // debug mode enabled
	bp    map[uint16]func() error // break points with optional callback
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
		bp:   make(map[uint16]func() error),
	}, nil
}

func (z *z80) GetBreakPoints() []uint16 {
	if !z.debug {
		return []uint16{}
	}

	bps := make([]uint16, 0, len(z.bp))
	for address := range z.bp {
		bps = append(bps, address)
	}
	return bps
}

func (z *z80) SetBreakPoint(address uint16, f func() error) {
	z.bp[address] = f
	z.debug = true
}

func (z *z80) DelBreakPoint(address uint16) {
	delete(z.bp, address)
	if len(z.bp) == 0 {
		z.debug = false
	}
}

func (z *z80) SetPC(address uint16) {
	z.pc = address
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

func (z *z80) res(bit, val byte) byte {
	mask := byte(^(1 << bit))
	return val & mask
}

func (z *z80) set(bit, val byte) byte {
	x := byte(1 << bit)
	return val | x
}

func (z *z80) ddcb() error {
	// zilog really is crazy, 4th byte + bit 7&6
	// descriminates the instruction type
	byte4 := z.bus.Read(z.pc + 3)
	xx := byte4 >> 6
	yy := 0x07 & (byte4 >> 3)
	zz := 0x07 & byte4
	fmt.Printf("x %x y %x z %x dd cb %02x %02x ->", xx, yy, zz,
		z.bus.Read(z.pc+2), byte4)
	switch xx {
	case 0:
		if zz == 6 {
			// rot[y],(IX+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			val := z.bus.Read(z.ix + displacement)
			fmt.Printf("val %02x ", val)
			switch yy {
			case 0x00:
				fmt.Printf("rlc ")
				z.bus.Write(z.ix+displacement, z.rlc(val))
			case 0x01:
				fmt.Printf("rrc ")
				z.bus.Write(z.ix+displacement, z.rrc(val))
			case 0x02:
				fmt.Printf("rl ")
				z.bus.Write(z.ix+displacement, z.rl(val))
			case 0x03:
				fmt.Printf("rr ")
				z.bus.Write(z.ix+displacement, z.rr(val))
			case 0x04:
				fmt.Printf("sla ")
				z.bus.Write(z.ix+displacement, z.sla(val))
			case 0x05:
				fmt.Printf("sra ")
				z.bus.Write(z.ix+displacement, z.sra(val))
			case 0x06:
				fmt.Printf("sll ")
				z.bus.Write(z.ix+displacement, z.sll(val))
			case 0x07:
				fmt.Printf("srl ")
				z.bus.Write(z.ix+displacement, z.srl(val))
			}
			fmt.Printf("new val %02x\n", z.bus.Read(z.ix+displacement))
			z.totalCycles += 1 // XXX
			z.pc += 4
			return nil
		}
		// ld r[z], rot[y] (IX+d)
		panic("ld")
	case 0x01:
		if zz == 6 {
			bit := byte4 & 0x38 >> 3
			displacement := uint16(z.bus.Read(z.pc + 2))
			val := z.bus.Read(z.ix + displacement)
			z.bit(bit, val)

			z.totalCycles += 1 // XXX
			z.pc += 4
			return nil
		}
	case 0x02:
		var val byte
		switch zz {
		case 0x00:
			val = byte(z.bc >> 8)
		case 0x01:
			val = byte(z.bc)
		case 0x02:
			val = byte(z.de >> 8)
		case 0x03:
			val = byte(z.de)
		case 0x04:
			val = byte(z.hl >> 8)
		case 0x05:
			val = byte(z.hl)
		case 0x06:
			val = z.bus.Read(z.hl)
		case 0x07:
			val = byte(z.af >> 8)
		}

		switch yy {
		case 0x00:
			z.add(val)
		case 0x01:
			z.adc(val)
		case 0x02:
			z.sub(val)
		case 0x03:
			z.sbc(val)
		case 0x04:
			z.and(val)
		case 0x05:
			z.xor(val)
		case 0x06:
			z.or(val)
		case 0x07:
			z.cp(val)
		}

		z.totalCycles += 1 // XXX
		z.pc += 4
		return nil
	default:
		fmt.Printf("x %x y %x z %x dd cb %02x %02x\n", xx, yy, zz, z.bus.Read(z.pc+2), byte4)
		panic("xxx")
	}

	//switch byte4 & 0xc0 {
	//case 0x00: // rot/shft
	//	// operation lives in bit 5, 4 & 3
	//	// Index	0	1	2	3	4	5	6	7
	//	// Value	RLC	RRC	RL	RR	SLA	SRA	SLL	SRL
	//	// Index	0	1	2	3	4	5	6	7
	//	// Value	B	C	D	E	H	L	(HL)	A
	//	op := byte4 & 38 >> 3
	//	switch op {
	//	case 0x00:
	//		reg := byte4 & 7
	//		switch reg {
	//		case 0x00:
	//			z.bc = z.bc&0x00ff | uint16(z.rlc(byte(z.bc>>8)))<<8
	//		case 0x01:
	//			z.bc = z.bc&0xff00 | uint16(z.rlc(byte(z.bc)))
	//		case 0x02:
	//			z.de = z.de&0x00ff | uint16(z.rlc(byte(z.de>>8)))<<8
	//		case 0x03:
	//			z.de = z.de&0xff00 | uint16(z.rlc(byte(z.de)))
	//		case 0x04:
	//			z.hl = z.de&0x00ff | uint16(z.rlc(byte(z.de>>8)))<<8
	//		case 0x05:
	//			z.hl = z.de&0xff00 | uint16(z.rlc(byte(z.de)))
	//		case 0x06:
	//			t := z.rlc(z.bus.Read(z.hl))
	//			z.bus.Write(z.hl, t)
	//		case 0x07:
	//			z.af = z.af&0x00ff | uint16(z.rlc(byte(z.af>>8)))<<8
	//		default:
	//			panic(fmt.Sprintf("---d %02x %02x %02x %02x %02x",
	//				byte4&0xc0,
	//				z.bus.Read(z.pc+0),
	//				z.bus.Read(z.pc+1),
	//				z.bus.Read(z.pc+2),
	//				z.bus.Read(z.pc+3)))
	//		}
	//	case 0x04:
	//		reg := byte4 & 7
	//		switch reg {
	//		case 0x00:
	//			z.bc = z.bc&0x00ff | uint16(z.rl(byte(z.bc>>8)))<<8
	//		case 0x01:
	//			z.bc = z.bc&0xff00 | uint16(z.rl(byte(z.bc)))
	//		case 0x02:
	//			z.de = z.de&0x00ff | uint16(z.rl(byte(z.de>>8)))<<8
	//		case 0x03:
	//			z.de = z.de&0xff00 | uint16(z.rl(byte(z.de)))
	//		case 0x04:
	//			z.hl = z.de&0x00ff | uint16(z.rl(byte(z.de>>8)))<<8
	//		case 0x05:
	//			z.hl = z.de&0xff00 | uint16(z.rl(byte(z.de)))
	//		case 0x06:
	//			t := z.rl(z.bus.Read(z.hl))
	//			z.bus.Write(z.hl, t)
	//		case 0x07:
	//			z.af = z.af&0x00ff | uint16(z.rl(byte(z.af>>8)))<<8
	//		default:
	//			panic(fmt.Sprintf("---d %02x %02x %02x %02x %02x",
	//				byte4&0xc0,
	//				z.bus.Read(z.pc+0),
	//				z.bus.Read(z.pc+1),
	//				z.bus.Read(z.pc+2),
	//				z.bus.Read(z.pc+3)))
	//		}
	//	default:
	//		// XXX should become a nop
	//		panic(fmt.Sprintf("d %02x op %02x %02x %02x %02x %02x",
	//			byte4&0xc0,
	//			op,
	//			z.bus.Read(z.pc+0),
	//			z.bus.Read(z.pc+1),
	//			z.bus.Read(z.pc+2),
	//			z.bus.Read(z.pc+3)))
	//	}
	//case 0x40: // bit b,(ix+d)
	//	bit := byte4 & 0x38 >> 3
	//	displacement := uint16(z.bus.Read(z.pc + 2))
	//	val := z.bus.Read(z.ix + displacement)
	//	z.bit(bit, val)
	//case 0x80: // res b,(ix+d)
	//	bit := byte4 & 0x38 >> 3
	//	displacement := uint16(z.bus.Read(z.pc + 2))
	//	val := z.bus.Read(z.ix + displacement)
	//	z.bus.Write(z.ix+displacement, z.res(bit, val))
	//case 0xc0: // set b,(ix+d)
	//	bit := byte4 & 0x38 >> 3
	//	displacement := uint16(z.bus.Read(z.pc + 2))
	//	val := z.bus.Read(z.ix + displacement)
	//	z.bus.Write(z.ix+displacement, z.set(bit, val))
	//default:
	//	panic(fmt.Sprintf("+++d %02x %02x %02x %02x %02x",
	//		byte4&0xc0,
	//		z.bus.Read(z.pc+0),
	//		z.bus.Read(z.pc+1),
	//		z.bus.Read(z.pc+2),
	//		z.bus.Read(z.pc+3)))
	//}

	return nil
}

func (z *z80) genericPostInstruction(o *opcode) {
	// panic for now
	if o.noBytes == 0 || o.noCycles == 0 {
		panic(fmt.Sprintf("opcode missing or invalid: pc %04x "+
			"%02x %02x %02x %02x", z.pc,
			z.bus.Read(z.pc+0), z.bus.Read(z.pc+1),
			z.bus.Read(z.pc+2), z.bus.Read(z.pc+3)))
	}

	z.totalCycles += o.noCycles
	z.pc += uint16(o.noBytes)
}

func (z *z80) Step() error {
	err := z.step()
	if err != nil {
		return err
	}

	if !z.debug {
		return nil
	}

	// see if we hit a break point
	if cb, ok := z.bp[z.pc]; ok {
		return BreakpointError{PC: z.pc, Callback: cb}
	}

	return nil
}

// Step executes the instruction as pointed at by PC.
func (z *z80) step() error {

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
		a := byte(z.af >> 8)
		a = a<<1 | a>>7
		f := byte(z.af)&(FLAG_P|FLAG_Z|FLAG_S) |
			a&(FLAG_C|FLAG_3|FLAG_5)
		z.af = uint16(a)<<8 | uint16(f)
	case 0x08: // ex af,af'
		t := z.af
		z.af = z.af_
		z.af_ = t
	case 0x09: // add hl,bc
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
		a := byte(z.af >> 8)
		f := byte(z.af)&(FLAG_P|FLAG_Z|FLAG_S) | a&FLAG_C
		a = a>>1 | a<<7
		f |= a & (FLAG_3 | FLAG_5)
		z.af = uint16(a)<<8 | uint16(f)
	case 0x10: // djnz
		b := byte(z.bc>>8) - 1
		z.bc = z.bc&0x00ff | uint16(b)<<8
		if b != 0 {
			z.pc = z.pc + 2 + uint16(int8(z.bus.Read(z.pc+1)))
			z.totalCycles += 13
			return nil
		}
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
	case 0x17: // rla
		t := byte(z.af >> 8)
		a := t<<1 | byte(z.af)&FLAG_C
		f := byte(z.af)&(FLAG_P|FLAG_Z|FLAG_S) | a&(FLAG_3|FLAG_5) |
			t>>7
		z.af = uint16(a)<<8 | uint16(f)
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
	case 0x1f: // rra
		t := byte(z.af >> 8)
		a := t>>1 | byte(z.af)<<7
		f := byte(z.af)&(FLAG_P|FLAG_Z|FLAG_S) |
			a&(FLAG_3|FLAG_5) | t&FLAG_C
		z.af = uint16(a)<<8 | uint16(f)
	case 0x20: // jr nz,d
		if z.af&zero == 0 {
			z.pc = z.pc + 2 + uint16(int8(z.bus.Read(z.pc+1)))
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
		// XXX make this generic
		z.totalCycles += 7
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
	case 0x27: // daa
		z.daa()
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
		a := byte(z.af >> 8)
		a ^= 0xff
		f := byte(z.af)&(FLAG_C|FLAG_P|FLAG_Z|FLAG_S) |
			a&(FLAG_3|FLAG_5) |
			FLAG_N | FLAG_H
		z.af = uint16(a)<<8 | uint16(f)
	case 0x30: // jr nc,d
		if z.af&carry == 0 {
			z.pc = z.pc + 2 + uint16(int8(z.bus.Read(z.pc+1)))
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
		// XXX make this generic
		z.totalCycles += 7
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
		a := byte(z.af >> 8)
		f := byte(z.af)&(FLAG_P|FLAG_Z|FLAG_S) |
			a&(FLAG_3|FLAG_5) |
			FLAG_C
		z.af = z.af&0xff00 | uint16(f)
	case 0x38: // jr c,d
		if z.af&carry == carry {
			z.pc = z.pc + 2 + uint16(int8(z.bus.Read(z.pc+1)))
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
		// XXX make this generic
		z.totalCycles += 7
	case 0x39: // add hl,sp
		z.hl = z.add16(z.hl, z.sp)
	case 0x3d: // dec a
		z.af = uint16(z.dec(byte(z.af>>8)))<<8 | z.af&0x00ff
	case 0x3f: // ccf
		a := byte(z.af >> 8)
		f := byte(z.af)&(FLAG_P|FLAG_Z|FLAG_S) |
			ternB(byte(z.af)&FLAG_C != 0, FLAG_H, FLAG_C) |
			a&(FLAG_3|FLAG_5)
		z.af = z.af&0xff00 | uint16(f)
	case 0x3a: // ld a,(nn)
		z.af = uint16(z.bus.Read(uint16(z.bus.Read(z.pc+1))|
			uint16(z.bus.Read(z.pc+2))<<8))<<8 | z.af&0x00ff
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
		z.bc = z.bc&0x00ff | uint16(z.bus.Read(z.hl))<<8
	case 0x47: // ld b,a
		z.bc = z.bc&0x00ff | z.af&0xff00
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
		z.bc = z.bc&0xff00 | uint16(z.bus.Read(z.hl))
	case 0x4f: // ld c,a
		z.bc = z.bc&0xff00 | z.af>>8
	case 0x50: // ld d,b
		z.de = z.de&0x00ff | z.bc&0xff00
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
		return HaltError{PC: z.pc}
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
		z.adc(byte(z.de >> 8))
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
	case 0xa6: // and (hl)
		z.and(z.bus.Read(z.hl))
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
		case 0x00: // rlc b
			z.bc = uint16(z.rlc(byte(z.bc>>8)))<<8 | z.bc&0x00ff
		case 0x01: // rlc c
			z.bc = uint16(z.rlc(byte(z.bc))) | z.bc&0xff00
		case 0x02: // rlc d
			z.de = uint16(z.rlc(byte(z.de>>8)))<<8 | z.de&0x00ff
		case 0x03: // rlc e
			z.de = uint16(z.rlc(byte(z.de))) | z.de&0xff00
		case 0x04: // rlc h
			z.hl = uint16(z.rlc(byte(z.hl>>8)))<<8 | z.hl&0x00ff
		case 0x05: // rlc l
			z.hl = uint16(z.rlc(byte(z.hl))) | z.hl&0xff00
		case 0x06: // rlc (hl)
			t := z.rlc(z.bus.Read(z.hl))
			z.bus.Write(z.hl, t)
		case 0x07: // rlc a
			z.af = uint16(z.rlc(byte(z.af>>8)))<<8 | z.af&0x00ff
		case 0x08: // rrc b
			z.bc = uint16(z.rrc(byte(z.bc>>8)))<<8 | z.bc&0x00ff
		case 0x09: // rrc c
			z.bc = uint16(z.rrc(byte(z.bc))) | z.bc&0xff00
		case 0x0a: // rrc d
			z.de = uint16(z.rrc(byte(z.de>>8)))<<8 | z.de&0x00ff
		case 0x0b: // rrc e
			z.de = uint16(z.rrc(byte(z.de))) | z.de&0xff00
		case 0x0c: // rrc h
			z.hl = uint16(z.rrc(byte(z.hl>>8)))<<8 | z.hl&0x00ff
		case 0x0d: // rrc l
			z.hl = uint16(z.rrc(byte(z.hl))) | z.hl&0xff00
		case 0x0e: // rrc (hl)
			t := z.rrc(z.bus.Read(z.hl))
			z.bus.Write(z.hl, t)
		case 0x0f: // rrc a
			z.af = uint16(z.rrc(byte(z.af>>8)))<<8 | z.af&0x00ff
		case 0x10: // rl b
			z.bc = uint16(z.rl(byte(z.bc>>8)))<<8 | z.bc&0x00ff
		case 0x11: // rl c
			z.bc = uint16(z.rl(byte(z.bc))) | z.bc&0xff00
		case 0x12: // rl d
			z.de = uint16(z.rl(byte(z.de>>8)))<<8 | z.de&0x00ff
		case 0x13: // rl e
			z.de = uint16(z.rl(byte(z.de))) | z.de&0xff00
		case 0x14: // rl h
			z.hl = uint16(z.rl(byte(z.hl>>8)))<<8 | z.hl&0x00ff
		case 0x15: // rl l
			z.hl = uint16(z.rl(byte(z.hl))) | z.hl&0xff00
		case 0x16: // rl (hl)
			t := z.rl(z.bus.Read(z.hl))
			z.bus.Write(z.hl, t)
		case 0x17: // rl a
			z.af = uint16(z.rl(byte(z.af>>8)))<<8 | z.af&0x00ff
		case 0x18: // rr b
			z.bc = uint16(z.rr(byte(z.bc>>8)))<<8 | z.bc&0x00ff
		case 0x19: // rr c
			z.bc = uint16(z.rr(byte(z.bc))) | z.bc&0xff00
		case 0x1a: // rr d
			z.de = uint16(z.rr(byte(z.de>>8)))<<8 | z.de&0x00ff
		case 0x1b: // rr e
			z.de = uint16(z.rr(byte(z.de))) | z.de&0xff00
		case 0x1c: // rr h
			z.hl = uint16(z.rr(byte(z.hl>>8)))<<8 | z.hl&0x00ff
		case 0x1d: // rr l
			z.hl = uint16(z.rr(byte(z.hl))) | z.hl&0xff00
		case 0x1e: // rr (hl)
			t := z.rr(z.bus.Read(z.hl))
			z.bus.Write(z.hl, t)
		case 0x1f: // rr a
			z.af = uint16(z.rr(byte(z.af>>8)))<<8 | z.af&0x00ff
		case 0x20: // sla b
			z.bc = uint16(z.sla(byte(z.bc>>8)))<<8 | z.bc&0x00ff
		case 0x21: // sla c
			z.bc = uint16(z.sla(byte(z.bc))) | z.bc&0xff00
		case 0x22: // sla d
			z.de = uint16(z.sla(byte(z.de>>8)))<<8 | z.de&0x00ff
		case 0x23: // sla e
			z.de = uint16(z.sla(byte(z.de))) | z.de&0xff00
		case 0x24: // sla h
			z.hl = uint16(z.sla(byte(z.hl>>8)))<<8 | z.hl&0x00ff
		case 0x25: // sla l
			z.hl = uint16(z.sla(byte(z.hl))) | z.hl&0xff00
		case 0x26: // sla (hl)
			t := z.sla(z.bus.Read(z.hl))
			z.bus.Write(z.hl, t)
		case 0x27: // sla a
			z.af = uint16(z.sla(byte(z.af>>8)))<<8 | z.af&0x00ff
		case 0x28: // sra b
			z.bc = uint16(z.sra(byte(z.bc>>8)))<<8 | z.bc&0x00ff
		case 0x29: // sra c
			z.bc = uint16(z.sra(byte(z.bc))) | z.bc&0xff00
		case 0x2a: // sra d
			z.de = uint16(z.sra(byte(z.de>>8)))<<8 | z.de&0x00ff
		case 0x2b: // sra e
			z.de = uint16(z.sra(byte(z.de))) | z.de&0xff00
		case 0x2c: // sra h
			z.hl = uint16(z.sra(byte(z.hl>>8)))<<8 | z.hl&0x00ff
		case 0x2d: // sra l
			z.hl = uint16(z.sra(byte(z.hl))) | z.hl&0xff00
		case 0x2e: // sra (hl)
			t := z.sra(z.bus.Read(z.hl))
			z.bus.Write(z.hl, t)
		case 0x2f: // sra a
			z.af = uint16(z.sra(byte(z.af>>8)))<<8 | z.af&0x00ff
		case 0x30: // sll b
			z.bc = uint16(z.sll(byte(z.bc>>8)))<<8 | z.bc&0x00ff
		case 0x31: // sll c
			z.bc = uint16(z.sll(byte(z.bc))) | z.bc&0xff00
		case 0x32: // sll d
			z.de = uint16(z.sll(byte(z.de>>8)))<<8 | z.de&0x00ff
		case 0x33: // sll e
			z.de = uint16(z.sll(byte(z.de))) | z.de&0xff00
		case 0x34: // sll h
			z.hl = uint16(z.sll(byte(z.hl>>8)))<<8 | z.hl&0x00ff
		case 0x35: // sll l
			z.hl = uint16(z.sll(byte(z.hl))) | z.hl&0xff00
		case 0x36: // sll (hl)
			t := z.sll(z.bus.Read(z.hl))
			z.bus.Write(z.hl, t)
		case 0x37: // sll a
			z.af = uint16(z.sll(byte(z.af>>8)))<<8 | z.af&0x00ff
		case 0x38: // srl b
			z.bc = uint16(z.srl(byte(z.bc>>8)))<<8 | z.bc&0x00ff
		case 0x39: // srl c
			z.bc = uint16(z.srl(byte(z.bc))) | z.bc&0xff00
		case 0x3a: // srl d
			z.de = uint16(z.srl(byte(z.de>>8)))<<8 | z.de&0x00ff
		case 0x3b: // srl e
			z.de = uint16(z.srl(byte(z.de))) | z.de&0xff00
		case 0x3c: // srl h
			z.hl = uint16(z.srl(byte(z.hl>>8)))<<8 | z.hl&0x00ff
		case 0x3d: // srl l
			z.hl = uint16(z.srl(byte(z.hl))) | z.hl&0xff00
		case 0x3e: // srl (hl)
			t := z.srl(z.bus.Read(z.hl))
			z.bus.Write(z.hl, t)
		case 0x3f: // srl a
			z.af = uint16(z.srl(byte(z.af>>8)))<<8 | z.af&0x00ff
		case 0x40: // bit 0,b
			z.bit(0, byte(z.bc>>8))
		case 0x41: // bit 0,c
			z.bit(0, byte(z.bc))
		case 0x42: // bit 0,d
			z.bit(0, byte(z.de>>8))
		case 0x43: // bit 0,e
			z.bit(0, byte(z.de))
		case 0x44: // bit 0,h
			z.bit(0, byte(z.hl>>8))
		case 0x45: // bit 0,l
			z.bit(0, byte(z.hl))
		case 0x46: // bit 0,(hl)
			z.bit(0, z.bus.Read(z.hl))
		case 0x47: // bit 0,a
			z.bit(0, byte(z.af>>8))
		case 0x48: // bit 1,b
			z.bit(1, byte(z.bc>>8))
		case 0x49: // bit 1,c
			z.bit(1, byte(z.bc))
		case 0x4a: // bit 1,d
			z.bit(1, byte(z.de>>8))
		case 0x4b: // bit 1,e
			z.bit(1, byte(z.de))
		case 0x4c: // bit 1,h
			z.bit(1, byte(z.hl>>8))
		case 0x4d: // bit 1,l
			z.bit(1, byte(z.hl))
		case 0x4e: // bit 1,(hl)
			z.bit(1, z.bus.Read(z.hl))
		case 0x4f: // bit 1,a
			z.bit(1, byte(z.af>>8))
		case 0x50: // bit 2,b
			z.bit(2, byte(z.bc>>8))
		case 0x51: // bit 2,c
			z.bit(2, byte(z.bc))
		case 0x52: // bit 2,d
			z.bit(2, byte(z.de>>8))
		case 0x53: // bit 2,e
			z.bit(2, byte(z.de))
		case 0x54: // bit 2,h
			z.bit(2, byte(z.hl>>8))
		case 0x55: // bit 2,l
			z.bit(2, byte(z.hl))
		case 0x56: // bit 2,(hl)
			z.bit(2, z.bus.Read(z.hl))
		case 0x57: // bit 2,a
			z.bit(2, byte(z.af>>8))
		case 0x58: // bit 3,b
			z.bit(3, byte(z.bc>>8))
		case 0x59: // bit 3,c
			z.bit(3, byte(z.bc))
		case 0x5a: // bit 3,d
			z.bit(3, byte(z.de>>8))
		case 0x5b: // bit 3,e
			z.bit(3, byte(z.de))
		case 0x5c: // bit 3,h
			z.bit(3, byte(z.hl>>8))
		case 0x5d: // bit 3,l
			z.bit(3, byte(z.hl))
		case 0x5e: // bit 3,(hl)
			z.bit(3, z.bus.Read(z.hl))
		case 0x5f: // bit 3,a
			z.bit(3, byte(z.af>>8))
		case 0x60: // bit 4,b
			z.bit(4, byte(z.bc>>8))
		case 0x61: // bit 4,c
			z.bit(4, byte(z.bc))
		case 0x62: // bit 4,d
			z.bit(4, byte(z.de>>8))
		case 0x63: // bit 4,e
			z.bit(4, byte(z.de))
		case 0x64: // bit 4,h
			z.bit(4, byte(z.hl>>8))
		case 0x65: // bit 4,l
			z.bit(4, byte(z.hl))
		case 0x66: // bit 4,(hl)
			z.bit(4, z.bus.Read(z.hl))
		case 0x67: // bit 4,a
			z.bit(4, byte(z.af>>8))
		case 0x68: // bit 5,b
			z.bit(5, byte(z.bc>>8))
		case 0x69: // bit 5,c
			z.bit(5, byte(z.bc))
		case 0x6a: // bit 5,d
			z.bit(5, byte(z.de>>8))
		case 0x6b: // bit 5,e
			z.bit(5, byte(z.de))
		case 0x6c: // bit 5,h
			z.bit(5, byte(z.hl>>8))
		case 0x6d: // bit 5,l
			z.bit(5, byte(z.hl))
		case 0x6e: // bit 5,(hl)
			z.bit(5, z.bus.Read(z.hl))
		case 0x6f: // bit 5,a
			z.bit(5, byte(z.af>>8))
		case 0x70: // bit 6,b
			z.bit(6, byte(z.bc>>8))
		case 0x71: // bit 6,c
			z.bit(6, byte(z.bc))
		case 0x72: // bit 6,d
			z.bit(6, byte(z.de>>8))
		case 0x73: // bit 6,e
			z.bit(6, byte(z.de))
		case 0x74: // bit 6,h
			z.bit(6, byte(z.hl>>8))
		case 0x75: // bit 6,l
			z.bit(6, byte(z.hl))
		case 0x76: // bit 6,(hl)
			z.bit(6, z.bus.Read(z.hl))
		case 0x77: // bit 6,a
			z.bit(6, byte(z.af>>8))
		case 0x78: // bit 7,b
			z.bit(7, byte(z.bc>>8))
		case 0x79: // bit 7,c
			z.bit(7, byte(z.bc))
		case 0x7a: // bit 7,d
			z.bit(7, byte(z.de>>8))
		case 0x7b: // bit 7,e
			z.bit(7, byte(z.de))
		case 0x7c: // bit 7,h
			z.bit(7, byte(z.hl>>8))
		case 0x7d: // bit 7,l
			z.bit(7, byte(z.hl))
		case 0x7e: // bit 7,(hl)
			z.bit(7, z.bus.Read(z.hl))
		case 0x7f: // bit 7,a
			z.bit(7, byte(z.af>>8))
		case 0x80: // res 0,b
			z.bc = z.bc&0x00ff | uint16(z.res(0, byte(z.bc>>8)))<<8
		case 0x81: // res 0,c
			z.bc = z.bc&0xff00 | uint16(z.res(0, byte(z.bc)))
		case 0x82: // res 0,d
			z.de = z.de&0x00ff | uint16(z.res(0, byte(z.de>>8)))<<8
		case 0x83: // res 0,e
			z.de = z.de&0xff00 | uint16(z.res(0, byte(z.de)))
		case 0x84: // res 0,h
			z.hl = z.hl&0x00ff | uint16(z.res(0, byte(z.hl>>8)))<<8
		case 0x85: // res 0,l
			z.hl = z.hl&0xff00 | uint16(z.res(0, byte(z.hl)))
		case 0x86: // res 0,(hl)
			val := z.res(0, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0x87: // res 0,a
			z.af = z.af&0x00ff | uint16(z.res(0, byte(z.af>>8)))<<8
		case 0x88: // res 1,b
			z.bc = z.bc&0x00ff | uint16(z.res(1, byte(z.bc>>8)))<<8
		case 0x89: // res 1,c
			z.bc = z.bc&0xff00 | uint16(z.res(1, byte(z.bc)))
		case 0x8a: // res 1,d
			z.de = z.de&0x00ff | uint16(z.res(1, byte(z.de>>8)))<<8
		case 0x8b: // res 1,e
			z.de = z.de&0xff00 | uint16(z.res(1, byte(z.de)))
		case 0x8c: // res 1,h
			z.hl = z.hl&0x00ff | uint16(z.res(1, byte(z.hl>>8)))<<8
		case 0x8d: // res 1,l
			z.hl = z.hl&0xff00 | uint16(z.res(1, byte(z.hl)))
		case 0x8e: // res 1,(hl)
			val := z.res(1, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0x8f: // res 1,a
			z.af = z.af&0x00ff | uint16(z.res(1, byte(z.af>>8)))<<8
		case 0x90: // res 2,b
			z.bc = z.bc&0x00ff | uint16(z.res(2, byte(z.bc>>8)))<<8
		case 0x91: // res 2,c
			z.bc = z.bc&0xff00 | uint16(z.res(2, byte(z.bc)))
		case 0x92: // res 2,d
			z.de = z.de&0x00ff | uint16(z.res(2, byte(z.de>>8)))<<8
		case 0x93: // res 2,e
			z.de = z.de&0xff00 | uint16(z.res(2, byte(z.de)))
		case 0x94: // res 2,h
			z.hl = z.hl&0x00ff | uint16(z.res(2, byte(z.hl>>8)))<<8
		case 0x95: // res 2,l
			z.hl = z.hl&0xff00 | uint16(z.res(2, byte(z.hl)))
		case 0x96: // res 2,(hl)
			val := z.res(2, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0x97: // res 2,a
			z.af = z.af&0x00ff | uint16(z.res(2, byte(z.af>>8)))<<8
		case 0x98: // res 3,b
			z.bc = z.bc&0x00ff | uint16(z.res(3, byte(z.bc>>8)))<<8
		case 0x99: // res 3,c
			z.bc = z.bc&0xff00 | uint16(z.res(3, byte(z.bc)))
		case 0x9a: // res 3,d
			z.de = z.de&0x00ff | uint16(z.res(3, byte(z.de>>8)))<<8
		case 0x9b: // res 3,e
			z.de = z.de&0xff00 | uint16(z.res(3, byte(z.de)))
		case 0x9c: // res 3,h
			z.hl = z.hl&0x00ff | uint16(z.res(3, byte(z.hl>>8)))<<8
		case 0x9d: // res 3,l
			z.hl = z.hl&0xff00 | uint16(z.res(3, byte(z.hl)))
		case 0x9e: // res 3,(hl)
			val := z.res(3, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0x9f: // res 3,a
			z.af = z.af&0x00ff | uint16(z.res(3, byte(z.af>>8)))<<8
		case 0xa0: // res 4,b
			z.bc = z.bc&0x00ff | uint16(z.res(4, byte(z.bc>>8)))<<8
		case 0xa1: // res 4,c
			z.bc = z.bc&0xff00 | uint16(z.res(4, byte(z.bc)))
		case 0xa2: // res 4,d
			z.de = z.de&0x00ff | uint16(z.res(4, byte(z.de>>8)))<<8
		case 0xa3: // res 4,e
			z.de = z.de&0xff00 | uint16(z.res(4, byte(z.de)))
		case 0xa4: // res 4,h
			z.hl = z.hl&0x00ff | uint16(z.res(4, byte(z.hl>>8)))<<8
		case 0xa5: // res 4,l
			z.hl = z.hl&0xff00 | uint16(z.res(4, byte(z.hl)))
		case 0xa6: // res 4,(hl)
			val := z.res(4, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xa7: // res 4,a
			z.af = z.af&0x00ff | uint16(z.res(4, byte(z.af>>8)))<<8
		case 0xa8: // res 5,b
			z.bc = z.bc&0x00ff | uint16(z.res(5, byte(z.bc>>8)))<<8
		case 0xa9: // res 5,c
			z.bc = z.bc&0xff00 | uint16(z.res(5, byte(z.bc)))
		case 0xaa: // res 5,d
			z.de = z.de&0x00ff | uint16(z.res(5, byte(z.de>>8)))<<8
		case 0xab: // res 5,e
			z.de = z.de&0xff00 | uint16(z.res(5, byte(z.de)))
		case 0xac: // res 5,h
			z.hl = z.hl&0x00ff | uint16(z.res(5, byte(z.hl>>8)))<<8
		case 0xad: // res 5,l
			z.hl = z.hl&0xff00 | uint16(z.res(5, byte(z.hl)))
		case 0xae: // res 5,(hl)
			val := z.res(5, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xaf: // res 5,a
			z.af = z.af&0x00ff | uint16(z.res(5, byte(z.af>>8)))<<8
		case 0xb0: // res 6,b
			z.bc = z.bc&0x00ff | uint16(z.res(6, byte(z.bc>>8)))<<8
		case 0xb1: // res 6,c
			z.bc = z.bc&0xff00 | uint16(z.res(6, byte(z.bc)))
		case 0xb2: // res 6,d
			z.de = z.de&0x00ff | uint16(z.res(6, byte(z.de>>8)))<<8
		case 0xb3: // res 6,e
			z.de = z.de&0xff00 | uint16(z.res(6, byte(z.de)))
		case 0xb4: // res 6,h
			z.hl = z.hl&0x00ff | uint16(z.res(6, byte(z.hl>>8)))<<8
		case 0xb5: // res 6,l
			z.hl = z.hl&0xff00 | uint16(z.res(6, byte(z.hl)))
		case 0xb6: // res 6,(hl)
			val := z.res(6, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xb7: // res 6,a
			z.af = z.af&0x00ff | uint16(z.res(6, byte(z.af>>8)))<<8
		case 0xb8: // res 7,b
			z.bc = z.bc&0x00ff | uint16(z.res(7, byte(z.bc>>8)))<<8
		case 0xb9: // res 7,c
			z.bc = z.bc&0xff00 | uint16(z.res(7, byte(z.bc)))
		case 0xba: // res 7,d
			z.de = z.de&0x00ff | uint16(z.res(7, byte(z.de>>8)))<<8
		case 0xbb: // res 7,e
			z.de = z.de&0xff00 | uint16(z.res(7, byte(z.de)))
		case 0xbc: // res 7,h
			z.hl = z.hl&0x00ff | uint16(z.res(7, byte(z.hl>>8)))<<8
		case 0xbd: // res 7,l
			z.hl = z.hl&0xff00 | uint16(z.res(7, byte(z.hl)))
		case 0xbe: // res 7,(hl)
			val := z.res(7, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xbf: // res 7,a
			z.af = z.af&0x00ff | uint16(z.res(7, byte(z.af>>8)))<<8
		case 0xc0: // set 0,b
			z.bc = z.bc&0x00ff | uint16(z.set(0, byte(z.bc>>8)))<<8
		case 0xc1: // set 0,c
			z.bc = z.bc&0xff00 | uint16(z.set(0, byte(z.bc)))
		case 0xc2: // set 0,d
			z.de = z.de&0x00ff | uint16(z.set(0, byte(z.de>>8)))<<8
		case 0xc3: // set 0,e
			z.de = z.de&0xff00 | uint16(z.set(0, byte(z.de)))
		case 0xc4: // set 0,h
			z.hl = z.hl&0x00ff | uint16(z.set(0, byte(z.hl>>8)))<<8
		case 0xc5: // set 0,l
			z.hl = z.hl&0xff00 | uint16(z.set(0, byte(z.hl)))
		case 0xc6: // set 0,(hl)
			val := z.set(0, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xc7: // set 0,a
			z.af = z.af&0x00ff | uint16(z.set(0, byte(z.af>>8)))<<8
		case 0xc8: // set 1,b
			z.bc = z.bc&0x00ff | uint16(z.set(1, byte(z.bc>>8)))<<8
		case 0xc9: // set 1,c
			z.bc = z.bc&0xff00 | uint16(z.set(1, byte(z.bc)))
		case 0xca: // set 1,d
			z.de = z.de&0x00ff | uint16(z.set(1, byte(z.de>>8)))<<8
		case 0xcb: // set 1,e
			z.de = z.de&0xff00 | uint16(z.set(1, byte(z.de)))
		case 0xcc: // set 1,h
			z.hl = z.hl&0x00ff | uint16(z.set(1, byte(z.hl>>8)))<<8
		case 0xcd: // set 1,l
			z.hl = z.hl&0xff00 | uint16(z.set(1, byte(z.hl)))
		case 0xce: // set 1,(hl)
			val := z.set(1, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xcf: // set 1,a
			z.af = z.af&0x00ff | uint16(z.set(1, byte(z.af>>8)))<<8
		case 0xd0: // set 2,b
			z.bc = z.bc&0x00ff | uint16(z.set(2, byte(z.bc>>8)))<<8
		case 0xd1: // set 2,c
			z.bc = z.bc&0xff00 | uint16(z.set(2, byte(z.bc)))
		case 0xd2: // set 2,d
			z.de = z.de&0x00ff | uint16(z.set(2, byte(z.de>>8)))<<8
		case 0xd3: // set 2,e
			z.de = z.de&0xff00 | uint16(z.set(2, byte(z.de)))
		case 0xd4: // set 2,h
			z.hl = z.hl&0x00ff | uint16(z.set(2, byte(z.hl>>8)))<<8
		case 0xd5: // set 2,l
			z.hl = z.hl&0xff00 | uint16(z.set(2, byte(z.hl)))
		case 0xd6: // set 2,(hl)
			val := z.set(2, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xd7: // set 2,a
			z.af = z.af&0x00ff | uint16(z.set(2, byte(z.af>>8)))<<8
		case 0xd8: // set 3,b
			z.bc = z.bc&0x00ff | uint16(z.set(3, byte(z.bc>>8)))<<8
		case 0xd9: // set 3,c
			z.bc = z.bc&0xff00 | uint16(z.set(3, byte(z.bc)))
		case 0xda: // set 3,d
			z.de = z.de&0x00ff | uint16(z.set(3, byte(z.de>>8)))<<8
		case 0xdb: // set 3,e
			z.de = z.de&0xff00 | uint16(z.set(3, byte(z.de)))
		case 0xdc: // set 3,h
			z.hl = z.hl&0x00ff | uint16(z.set(3, byte(z.hl>>8)))<<8
		case 0xdd: // set 3,l
			z.hl = z.hl&0xff00 | uint16(z.set(3, byte(z.hl)))
		case 0xde: // set 3,(hl)
			val := z.set(3, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xdf: // set 3,a
			z.af = z.af&0x00ff | uint16(z.set(3, byte(z.af>>8)))<<8
		case 0xe0: // set 4,b
			z.bc = z.bc&0x00ff | uint16(z.set(4, byte(z.bc>>8)))<<8
		case 0xe1: // set 4,c
			z.bc = z.bc&0xff00 | uint16(z.set(4, byte(z.bc)))
		case 0xe2: // set 4,d
			z.de = z.de&0x00ff | uint16(z.set(4, byte(z.de>>8)))<<8
		case 0xe3: // set 4,e
			z.de = z.de&0xff00 | uint16(z.set(4, byte(z.de)))
		case 0xe4: // set 4,h
			z.hl = z.hl&0x00ff | uint16(z.set(4, byte(z.hl>>8)))<<8
		case 0xe5: // set 4,l
			z.hl = z.hl&0xff00 | uint16(z.set(4, byte(z.hl)))
		case 0xe6: // set 4,(hl)
			val := z.set(4, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xe7: // set 4,a
			z.af = z.af&0x00ff | uint16(z.set(4, byte(z.af>>8)))<<8
		case 0xe8: // set 5,b
			z.bc = z.bc&0x00ff | uint16(z.set(5, byte(z.bc>>8)))<<8
		case 0xe9: // set 5,c
			z.bc = z.bc&0xff00 | uint16(z.set(5, byte(z.bc)))
		case 0xea: // set 5,d
			z.de = z.de&0x00ff | uint16(z.set(5, byte(z.de>>8)))<<8
		case 0xeb: // set 5,e
			z.de = z.de&0xff00 | uint16(z.set(5, byte(z.de)))
		case 0xec: // set 5,h
			z.hl = z.hl&0x00ff | uint16(z.set(5, byte(z.hl>>8)))<<8
		case 0xed: // set 5,l
			z.hl = z.hl&0xff00 | uint16(z.set(5, byte(z.hl)))
		case 0xee: // set 5,(hl)
			val := z.set(5, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xef: // set 5,a
			z.af = z.af&0x00ff | uint16(z.set(5, byte(z.af>>8)))<<8
		case 0xf0: // set 6,b
			z.bc = z.bc&0x00ff | uint16(z.set(6, byte(z.bc>>8)))<<8
		case 0xf1: // set 6,c
			z.bc = z.bc&0xff00 | uint16(z.set(6, byte(z.bc)))
		case 0xf2: // set 6,d
			z.de = z.de&0x00ff | uint16(z.set(6, byte(z.de>>8)))<<8
		case 0xf3: // set 6,e
			z.de = z.de&0xff00 | uint16(z.set(6, byte(z.de)))
		case 0xf4: // set 6,h
			z.hl = z.hl&0x00ff | uint16(z.set(6, byte(z.hl>>8)))<<8
		case 0xf5: // set 6,l
			z.hl = z.hl&0xff00 | uint16(z.set(6, byte(z.hl)))
		case 0xf6: // set 6,(hl)
			val := z.set(6, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xf7: // set 6,a
			z.af = z.af&0x00ff | uint16(z.set(6, byte(z.af>>8)))<<8
		case 0xf8: // set 7,b
			z.bc = z.bc&0x00ff | uint16(z.set(7, byte(z.bc>>8)))<<8
		case 0xf9: // set 7,c
			z.bc = z.bc&0xff00 | uint16(z.set(7, byte(z.bc)))
		case 0xfa: // set 7,d
			z.de = z.de&0x00ff | uint16(z.set(7, byte(z.de>>8)))<<8
		case 0xfb: // set 7,e
			z.de = z.de&0xff00 | uint16(z.set(7, byte(z.de)))
		case 0xfc: // set 7,h
			z.hl = z.hl&0x00ff | uint16(z.set(7, byte(z.hl>>8)))<<8
		case 0xfd: // set 7,l
			z.hl = z.hl&0xff00 | uint16(z.set(7, byte(z.hl)))
		case 0xfe: // set 7,(hl)
			val := z.set(7, z.bus.Read(z.hl))
			z.bus.Write(z.hl, val)
		case 0xff: // set 7,a
			z.af = z.af&0x00ff | uint16(z.set(7, byte(z.af>>8)))<<8
		default:
			return fmt.Errorf("invalid instruction: 0x%02x "+
				"0x%02x @ 0x%04x", opc, z.bus.Read(z.pc+1),
				z.pc)
			//return ErrInvalidInstruction
		}
	case 0xcc: //call z,nn
		if z.af&zero == zero {
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
	case 0xdc: //call c,nn
		if z.af&carry == carry {
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
	case 0xdd: // z80 only
		byte2 := z.bus.Read(z.pc + 1)
		opcodeStruct = &opcodesDD[byte2]
		switch byte2 {
		case 0x09: // add ix,bc
			z.ix = z.add16(z.ix, z.bc)
		case 0x19: // add ix,de
			z.ix = z.add16(z.ix, z.de)
		case 0x21: // ld ix,nn
			z.ix = uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
		case 0x22: // ld (nn),ix
			addr := uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
			z.bus.Write(addr, byte(z.ix))
			z.bus.Write(addr+1, byte(z.ix>>8))
		case 0x23: // inc ix
			z.ix += 1
		case 0x24: // inc ixh XXX this is supposed to be undocumented
			z.ix = uint16(z.inc(byte(z.ix>>8)))<<8 | z.ix&0x00ff
		case 0x25: // dec ixh XXX this is supposed to be undocumented
			z.ix = uint16(z.dec(byte(z.ix>>8)))<<8 | z.ix&0x00ff
		case 0x26: // ld ixh,n XXX this is supposed to be undocumented
			z.ix = z.ix&0x00ff | uint16(z.bus.Read(z.pc+2))<<8
		case 0x29: // add ix,ix
			z.ix = z.add16(z.ix, z.ix)
		case 0x2a: // ld ix,(nn)
			addr := uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
			z.ix = uint16(z.bus.Read(addr)) |
				uint16(z.bus.Read(addr+1))<<8
		case 0x2b: // dec ix
			z.ix -= 1
		case 0x2c: // inc ixl XXX this is supposed to be undocumented
			z.ix = uint16(z.inc(byte(z.ix))) | z.ix&0xff00
		case 0x2d: // dec ixl XXX this is supposed to be undocumented
			z.ix = uint16(z.dec(byte(z.ix))) | z.ix&0xff00
		case 0x2e: // ld ixl,n XXX this is supposed to be undocumented
			z.ix = z.ix&0xff00 | uint16(z.bus.Read(z.pc+2))
		case 0x34: // inc (ix+d)
			x := z.inc(z.bus.Read(z.ix + uint16(z.bus.Read(z.pc+2))))
			z.bus.Write(z.ix+uint16(z.bus.Read(z.pc+2)), x)
		case 0x35: // dec (ix+d)
			x := z.dec(z.bus.Read(z.ix + uint16(z.bus.Read(z.pc+2))))
			z.bus.Write(z.ix+uint16(z.bus.Read(z.pc+2)), x)
		case 0x36: // ld (ix+d),n
			val := z.bus.Read(z.pc + 3)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.ix+displacement, val)
		case 0x39: // add ix,sp
			z.ix = z.add16(z.ix, z.sp)
		case 0x40, 0x41, 0x42, 0x43: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x44: // ld b,ixh
			z.bc = z.bc&0x00ff | z.ix&0xff00
		case 0x45: // ld b,ixl
			z.bc = z.bc&0x00ff | z.ix<<8
		case 0x46: // ld b,(ix+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bc = z.bc&0x00ff |
				uint16(z.bus.Read(z.ix+displacement))<<8
		case 0x47, 0x48, 0x49, 0x4a, 0x4b: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x4c: // ld c,ixh
			z.bc = z.bc&0xff00 | z.ix>>8
		case 0x4d: // ld c,ixl
			z.bc = z.bc&0xff00 | z.ix&0x00ff
		case 0x4e: // ld c,(ix+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bc = z.bc&0xff00 |
				uint16(z.bus.Read(z.ix+displacement))
		case 0x4f, 0x50, 0x51, 0x52, 0x53: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x54: // ld d,ixh
			z.de = z.de&0x00ff | z.ix&0xff00
		case 0x55: // ld d,ixl
			z.de = z.de&0x00ff | z.ix<<8
		case 0x56: // ld d,(ix+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.de = z.de&0x00ff |
				uint16(z.bus.Read(z.ix+displacement))<<8
		case 0x57, 0x58, 0x59, 0x5a, 0x5b: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x5c: // ld e,ixh
			z.de = z.de&0xff00 | z.ix>>8
		case 0x5d: // ld e,ixl
			z.de = z.de&0xff00 | z.ix&0x00ff
		case 0x5e: // ld e,(ix+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.de = z.de&0xff00 |
				uint16(z.bus.Read(z.ix+displacement))
		case 0x5f: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x60: // ld ixh,b
			z.ix = z.ix&0x00ff | z.bc&0xff00
		case 0x61: // ld ixh,c
			z.ix = z.ix&0x00ff | z.bc<<8
		case 0x62: // ld ixh,d
			z.ix = z.ix&0x00ff | z.de&0xff00
		case 0x63: // ld ixh,e
			z.ix = z.ix&0x00ff | z.de<<8
		case 0x64: // ld ixh,ixh
			// nop
		case 0x65: // ld ixh,ixl
			z.ix = z.ix&0x00ff | z.ix<<8
		case 0x66: // ld h,(ix+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.hl = z.hl&0x00ff |
				uint16(z.bus.Read(z.ix+displacement))<<8
		case 0x67: // ld ixh,a
			z.ix = z.ix&0x00ff | z.af&0xff00
		case 0x68: // ld ixl,b
			z.ix = z.ix&0xff00 | z.bc>>8
		case 0x69: // ld ixl,c
			z.ix = z.ix&0xff00 | z.bc&0x00ff
		case 0x6a: // ld ixl,d
			z.ix = z.ix&0xff00 | z.de>>8
		case 0x6b: // ld ixl,e
			z.ix = z.ix&0xff00 | z.de&0x00ff
		case 0x6c: // ld ixl,ixh
			z.ix = z.ix&0xff00 | z.ix>>8
		case 0x6d: // ld ixl,ixl
			// nop
		case 0x6e: // ld l,(ix+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.hl = z.hl&0xff00 |
				uint16(z.bus.Read(z.ix+displacement))
		case 0x6f: // ld ixl,a
			z.ix = z.ix&0xff00 | z.af>>8
		case 0x70: // ld (ix+d),b
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.ix+displacement, byte(z.bc>>8))
		case 0x71: // ld (ix+d),c
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.ix+displacement, byte(z.bc))
		case 0x72: // ld (ix+d),d
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.ix+displacement, byte(z.de>>8))
		case 0x73: // ld (ix+d),e
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.ix+displacement, byte(z.de))
		case 0x74: // ld (ix+d),h
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.ix+displacement, byte(z.hl>>8))
		case 0x75: // ld (ix+d),l
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.ix+displacement, byte(z.hl))
		case 0x76: // ld (ix+d),n
			val := z.bus.Read(z.pc + 3)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.ix+displacement, val)
		case 0x77: // ld (ix+d),a
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.ix+displacement, byte(z.af>>8))
		case 0x78, 0x79, 0x7a, 0x7b: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x7c: // ld a,ixh
			z.af = z.af&0x00ff | z.ix&0xff00
		case 0x7d: // ld a,ixl
			z.af = z.af&0x00ff | z.ix<<8
		case 0x7e: // ld a,(ix+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.af = z.af&0x00ff |
				uint16(z.bus.Read(z.ix+displacement))<<8
		case 0x7f: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x84: // add a,ixh XXX this is supposed to be undocumented
			z.add(byte(z.ix >> 8))
		case 0x85: // add a,ixl XXX this is supposed to be undocumented
			z.add(byte(z.ix))
		case 0x86: // add a,(ixl+d)
			z.add(z.bus.Read(z.ix + uint16(z.bus.Read(z.pc+2))))
		case 0x8c: // adc a,ixh XXX this is supposed to be undocumented
			z.adc(byte(z.ix >> 8))
		case 0x8d: // add a,ixl XXX this is supposed to be undocumented
			z.adc(byte(z.ix))
		case 0x8e: // adc a,(ixl+d)
			z.adc(z.bus.Read(z.ix + uint16(z.bus.Read(z.pc+2))))
		case 0x94: // sub a,ixh XXX this is supposed to be undocumented
			z.sub(byte(z.ix >> 8))
		case 0x95: // sub a,ixl XXX this is supposed to be undocumented
			z.sub(byte(z.ix))
		case 0x96: // sub a,(ixl+d)
			z.sub(z.bus.Read(z.ix + uint16(z.bus.Read(z.pc+2))))
		case 0x9c: // sbc a,ixh XXX this is supposed to be undocumented
			z.sbc(byte(z.ix >> 8))
		case 0x9d: // sbc a,ixl XXX this is supposed to be undocumented
			z.sbc(byte(z.ix))
		case 0x9e: // sbc a,(ixl+d)
			z.sbc(z.bus.Read(z.ix + uint16(z.bus.Read(z.pc+2))))
		case 0xa4: // and a,ixh XXX this is supposed to be undocumented
			z.and(byte(z.ix >> 8))
		case 0xa5: // and a,ixl XXX this is supposed to be undocumented
			z.and(byte(z.ix))
		case 0xa6: // and (ixl+d)
			z.and(z.bus.Read(z.ix + uint16(z.bus.Read(z.pc+2))))
		case 0xac: // xor a,ixh XXX this is supposed to be undocumented
			z.xor(byte(z.ix >> 8))
		case 0xad: // xor a,ixl XXX this is supposed to be undocumented
			z.xor(byte(z.ix))
		case 0xae: // xor (ixl+d)
			z.xor(z.bus.Read(z.ix + uint16(z.bus.Read(z.pc+2))))
		case 0xb4: // or a,ixh XXX this is supposed to be undocumented
			z.or(byte(z.ix >> 8))
		case 0xb5: // or a,ixl XXX this is supposed to be undocumented
			z.or(byte(z.ix))
		case 0xb6: // or (ixl+d)
			z.or(z.bus.Read(z.ix + uint16(z.bus.Read(z.pc+2))))
		case 0xbc: // cp a,ixh XXX this is supposed to be undocumented
			z.cp(byte(z.ix >> 8))
		case 0xbd: // cp a,ixl XXX this is supposed to be undocumented
			z.cp(byte(z.ix))
		case 0xbe: // cp (ixl+d)
			z.cp(z.bus.Read(z.ix + uint16(z.bus.Read(z.pc+2))))
		case 0xcb:
			return z.ddcb()
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
		case 0xf9: // ld sp, ix
			z.sp = z.ix
		default:
			return fmt.Errorf("invalid instruction: 0x%02x "+
				"0x%02x 0x%02x 0x%02x @ 0x%04x", opc,
				z.bus.Read(z.pc+1), z.bus.Read(z.pc+2),
				z.bus.Read(z.pc+3), z.pc)
			//return ErrInvalidInstruction
		}
	case 0xde: // sbc a,i
		z.sbc(z.bus.Read(z.pc + 1))
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
		case 0x42: // sbc hl,bc
			z.sbc16(z.bc)
		case 0x43: // ld (nn),bc
			addr := uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
			z.bus.Write(addr, byte(z.bc))
			z.bus.Write(addr+1, byte(z.bc>>8))
		case 0x44: // neg
			t := byte(z.af >> 8)
			z.af = z.af & 0x00ff
			z.sub(t)
		case 0x4a: // adc hl,bc
			z.adc16(z.bc)
		case 0x4b: // ld bc,(nn)
			addr := uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
			z.bc = uint16(z.bus.Read(addr)) |
				uint16(z.bus.Read(addr+1))<<8
		case 0x4d: // reti
			z.iff1 = z.iff2

			pc := uint16(z.bus.Read(z.sp))
			z.sp++
			pc = uint16(z.bus.Read(z.sp))<<8 | pc&0x00ff
			z.sp++
			z.totalCycles += 14 // XXX
			z.pc = pc
			return nil
		case 0x52: // sbc hl,de
			z.sbc16(z.de)
		case 0x53: // ld (nn),de
			addr := uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
			z.bus.Write(addr, byte(z.de))
			z.bus.Write(addr+1, byte(z.de>>8))
		case 0x5a: // adc hl,de
			z.adc16(z.de)
		case 0x5b: // ld de,(nn)
			addr := uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
			z.de = uint16(z.bus.Read(addr)) |
				uint16(z.bus.Read(addr+1))<<8
		case 0x62: // sbc hl,hl
			z.sbc16(z.hl)
		case 0x67: // rrd
			z.rrd()
		case 0x6a: // adc hl,hl
			z.adc16(z.hl)
		case 0x6b: // ld (nn),hl
			addr := uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
			z.bus.Write(addr, byte(z.hl))
			z.bus.Write(addr+1, byte(z.hl>>8))
		case 0x6f: // rld
			z.rld()
		case 0x72: // sbc hl,sp
			z.sbc16(z.sp)
		case 0x73: // ld (nn),sp
			addr := uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
			z.bus.Write(addr, byte(z.sp))
			z.bus.Write(addr+1, byte(z.sp>>8))
		case 0x7a: // adc hl,sp
			z.adc16(z.sp)
		case 0x7b: // ld sp,(nn)
			addr := uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
			z.sp = uint16(z.bus.Read(addr)) |
				uint16(z.bus.Read(addr+1))<<8
		case 0xa0: // ldi
			z.ldi()
		case 0xa1: // cpi
			z.cpi()
		case 0xa8: // ldd
			z.ldd()
		case 0xa9: // cpd
			z.cpd()
		case 0xb0: // ldir
			t := z.bus.Read(z.hl)
			z.bus.Write(z.de, t)
			t += byte(z.af >> 8)
			z.bc--
			f := byte(z.af)&(FLAG_C|FLAG_Z|FLAG_S) |
				ternB(z.bc != 0, FLAG_V, 0) | (t & FLAG_3) |
				ternB((t&0x02 != 0), FLAG_5, 0)
			z.af = z.af&0xff00 | uint16(f)
			z.de++
			z.hl++
			if z.bc != 0 {
				// don't move pc
				z.totalCycles += 21
				return nil
			}
		case 0xb1: // cpir
			a := byte(z.af >> 8)
			val := z.bus.Read(z.hl)
			t := a - val
			lookup := a&0x08>>3 | val&0x08>>2 | t&0x08>>1
			z.bc--
			z.hl++
			f := byte(z.af)&FLAG_C |
				ternB(z.bc != 0, FLAG_V|FLAG_N, FLAG_N) |
				halfcarrySubTable[lookup] |
				ternB(t != 0, 0, FLAG_Z) | t&FLAG_S
			if f&FLAG_H != 0 {
				t--
			}
			f |= t&FLAG_3 | ternB(t&0x02 != 0, FLAG_5, 0)
			z.af = z.af&0xff00 | uint16(f)
			if f&(FLAG_V|FLAG_Z) == FLAG_V {
				// don't move pc
				z.totalCycles += 21
				return nil
			}
		case 0xb8: // lddr
			t := z.bus.Read(z.hl)
			z.bus.Write(z.de, t)
			z.bc--
			t += byte(z.af >> 8)
			f := byte(z.af)&(FLAG_C|FLAG_Z|FLAG_S) |
				ternB(z.bc != 0, FLAG_V, 0) | t&FLAG_3 |
				ternB(t&0x02 != 0, FLAG_5, 0)
			z.af = z.af&0xff00 | uint16(f)
			z.hl--
			z.de--
			if z.bc != 0 {
				// don't move pc
				z.totalCycles += 21
				return nil
			}
		case 0xb9: // cpdr
			a := byte(z.af >> 8)
			val := z.bus.Read(z.hl)
			t := a - val
			lookup := a&0x08>>3 | val&0x08>>2 | t&0x08>>1
			z.bc--
			z.hl--
			f := byte(z.af)&FLAG_C |
				ternB(z.bc != 0, FLAG_V|FLAG_N, FLAG_N) |
				halfcarrySubTable[lookup] |
				ternB(t != 0, 0, FLAG_Z) | t&FLAG_S
			if f&FLAG_H != 0 {
				t--
			}
			f |= t&FLAG_3 | ternB(t&0x02 != 0, FLAG_5, 0)
			z.af = z.af&0xff00 | uint16(f)
			if f&(FLAG_V|FLAG_Z) == FLAG_V {
				// don't move pc
				z.totalCycles += 21
				return nil
			}
		default:
			return fmt.Errorf("invalid instruction: 0x%02x "+
				"0x%02x @ 0x%04x", opc, z.bus.Read(z.pc+1),
				z.pc)
			//return ErrInvalidInstruction
		}
	case 0xee: // xor n
		z.xor(z.bus.Read(z.pc + 1))
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
	case 0xf9: // ld sp,hl
		z.sp = z.hl
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
		case 0x09: // add iy,bc
			z.iy = z.add16(z.iy, z.bc)
		case 0x19: // add iy,de
			z.iy = z.add16(z.iy, z.de)
		case 0x21: // ld iy,nn
			z.iy = uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
		case 0x22: // ld (nn),iy
			addr := uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
			z.bus.Write(addr, byte(z.iy))
			z.bus.Write(addr+1, byte(z.iy>>8))
		case 0x23: // inc iy
			z.iy += 1
		case 0x24: // inc iyh XXX this is supposed to be undocumented
			z.iy = uint16(z.inc(byte(z.iy>>8)))<<8 | z.iy&0x00ff
		case 0x25: // dec iyh XXX this is supposed to be undocumented
			z.iy = uint16(z.dec(byte(z.iy>>8)))<<8 | z.iy&0x00ff
		case 0x26: // ld iyh,n XXX this is supposed to be undocumented
			z.iy = z.iy&0x00ff | uint16(z.bus.Read(z.pc+2))<<8
		case 0x29: // add iy,iy
			z.iy = z.add16(z.iy, z.iy)
		case 0x2a: // ld iy,(nn)
			addr := uint16(z.bus.Read(z.pc+2)) |
				uint16(z.bus.Read(z.pc+3))<<8
			z.iy = uint16(z.bus.Read(addr)) |
				uint16(z.bus.Read(addr+1))<<8
		case 0x2b: // dec iy
			z.iy -= 1
		case 0x2c: // inc iyl XXX this is supposed to be undocumented
			z.iy = uint16(z.inc(byte(z.iy))) | z.iy&0xff00
		case 0x2d: // dec iyl XXX this is supposed to be undocumented
			z.iy = uint16(z.dec(byte(z.iy))) | z.iy&0xff00
		case 0x2e: // ld iyl,n XXX this is supposed to be undocumented
			z.iy = z.iy&0xff00 | uint16(z.bus.Read(z.pc+2))
		case 0x34: // inc (iy+d)
			x := z.inc(z.bus.Read(z.iy + uint16(z.bus.Read(z.pc+2))))
			z.bus.Write(z.iy+uint16(z.bus.Read(z.pc+2)), x)
		case 0x35: // dec (iy+d)
			x := z.dec(z.bus.Read(z.iy + uint16(z.bus.Read(z.pc+2))))
			z.bus.Write(z.iy+uint16(z.bus.Read(z.pc+2)), x)
		case 0x36: // ld (iy+d),n
			val := z.bus.Read(z.pc + 3)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.iy+displacement, val)
		case 0x39: // add iy,sp
			z.iy = z.add16(z.iy, z.sp)
		case 0x40, 0x41, 0x42, 0x43: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x44: // ld b,iyh
			z.bc = z.bc&0x00ff | z.iy&0xff00
		case 0x45: // ld b,iyl
			z.bc = z.bc&0x00ff | z.iy<<8
		case 0x46: // ld b,(iy+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bc = z.bc&0x00ff |
				uint16(z.bus.Read(z.iy+displacement))<<8
		case 0x47, 0x48, 0x49, 0x4a, 0x4b: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x4c: // ld c,iyh
			z.bc = z.bc&0xff00 | z.iy>>8
		case 0x4d: // ld c,iyl
			z.bc = z.bc&0xff00 | z.iy&0x00ff
		case 0x4e: // ld c,(iy+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bc = z.bc&0xff00 |
				uint16(z.bus.Read(z.iy+displacement))
		case 0x4f, 0x50, 0x51, 0x52, 0x53: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x54: // ld d,iyh
			z.de = z.de&0x00ff | z.iy&0xff00
		case 0x55: // ld d,iyl
			z.de = z.de&0x00ff | z.iy<<8
		case 0x56: // ld d,(iy+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.de = z.de&0x00ff |
				uint16(z.bus.Read(z.iy+displacement))<<8
		case 0x57, 0x58, 0x59, 0x5a, 0x5b: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x5c: // ld e,iyh
			z.de = z.de&0xff00 | z.iy>>8
		case 0x5d: // ld e,iyl
			z.de = z.de&0xff00 | z.iy&0x00ff
		case 0x5e: // ld e,(iy+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.de = z.de&0xff00 |
				uint16(z.bus.Read(z.iy+displacement))
		case 0x5f: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x60: // ld iyh,b
			z.iy = z.iy&0x00ff | z.bc&0xff00
		case 0x61: // ld iyh,c
			z.iy = z.iy&0x00ff | z.bc<<8
		case 0x62: // ld iyh,d
			z.iy = z.iy&0x00ff | z.de&0xff00
		case 0x63: // ld iyh,e
			z.iy = z.iy&0x00ff | z.de<<8
		case 0x64: // ld iyh,iyh
			// nop
		case 0x65: // ld iyh,iyl
			z.iy = z.iy&0x00ff | z.iy<<8
		case 0x66: // ld h,(iy+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.hl = z.hl&0x00ff |
				uint16(z.bus.Read(z.iy+displacement))<<8
		case 0x67: // ld iyh,a
			z.iy = z.iy&0x00ff | z.af&0xff00
		case 0x68: // ld iyl,b
			z.iy = z.iy&0xff00 | z.bc>>8
		case 0x69: // ld iyl,c
			z.iy = z.iy&0xff00 | z.bc&0x00ff
		case 0x6a: // ld iyl,d
			z.iy = z.iy&0xff00 | z.de>>8
		case 0x6b: // ld iyl,e
			z.iy = z.iy&0xff00 | z.de&0x00ff
		case 0x6c: // ld iyl,iyh
			z.iy = z.iy&0xff00 | z.iy>>8
		case 0x6d: // ld iyl,iyl
			// nop
		case 0x6e: // ld l,(iy+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.hl = z.hl&0xff00 |
				uint16(z.bus.Read(z.iy+displacement))
		case 0x6f: // ld iyl,a
			z.iy = z.iy&0xff00 | z.af>>8
		case 0x70: // ld (iy+d),b
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.iy+displacement, byte(z.bc>>8))
		case 0x71: // ld (iy+d),c
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.iy+displacement, byte(z.bc))
		case 0x72: // ld (iy+d),d
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.iy+displacement, byte(z.de>>8))
		case 0x73: // ld (iy+d),e
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.iy+displacement, byte(z.de))
		case 0x74: // ld (iy+d),h
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.iy+displacement, byte(z.hl>>8))
		case 0x75: // ld (iy+d),l
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.iy+displacement, byte(z.hl))
		case 0x76: // ld (iy+d),n
			val := z.bus.Read(z.pc + 3)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.iy+displacement, val)
		case 0x77: // ld (iy+d),a
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.bus.Write(z.iy+displacement, byte(z.af>>8))
		case 0x78, 0x79, 0x7a, 0x7b: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x7c: // ld a,iyh
			z.af = z.af&0x00ff | z.iy&0xff00
		case 0x7d: // ld a,iyl
			z.af = z.af&0x00ff | z.iy<<8
		case 0x7e: // ld a,(iy+d)
			displacement := uint16(z.bus.Read(z.pc + 2))
			z.af = z.af&0x00ff |
				uint16(z.bus.Read(z.iy+displacement))<<8
		case 0x7f: // noni
			z.pc += 1
			z.totalCycles += 4 // XXX
			return nil
		case 0x84: // add a,iyh XXX this is supposed to be undocumented
			z.add(byte(z.iy >> 8))
		case 0x85: // add a,iyl XXX this is supposed to be undocumented
			z.add(byte(z.iy))
		case 0x86: // add a,(iyl+d)
			z.add(z.bus.Read(z.iy + uint16(z.bus.Read(z.pc+2))))
		case 0x8c: // adc a,iyh XXX this is supposed to be undocumented
			z.adc(byte(z.iy >> 8))
		case 0x8d: // add a,iyl XXX this is supposed to be undocumented
			z.adc(byte(z.iy))
		case 0x8e: // adc a,(iyl+d)
			z.adc(z.bus.Read(z.iy + uint16(z.bus.Read(z.pc+2))))
		case 0x94: // sub a,iyh XXX this is supposed to be undocumented
			z.sub(byte(z.iy >> 8))
		case 0x95: // sub a,iyl XXX this is supposed to be undocumented
			z.sub(byte(z.iy))
		case 0x96: // sub a,(iyl+d)
			z.sub(z.bus.Read(z.iy + uint16(z.bus.Read(z.pc+2))))
		case 0x9c: // sbc a,iyh XXX this is supposed to be undocumented
			z.sbc(byte(z.iy >> 8))
		case 0x9d: // sbc a,iyl XXX this is supposed to be undocumented
			z.sbc(byte(z.iy))
		case 0x9e: // sbc a,(iyl+d)
			z.sbc(z.bus.Read(z.iy + uint16(z.bus.Read(z.pc+2))))
		case 0xa4: // and a,iyh XXX this is supposed to be undocumented
			z.and(byte(z.iy >> 8))
		case 0xa5: // and a,iyl XXX this is supposed to be undocumented
			z.and(byte(z.iy))
		case 0xa6: // and (iyl+d)
			z.and(z.bus.Read(z.iy + uint16(z.bus.Read(z.pc+2))))
		case 0xac: // xor a,iyh XXX this is supposed to be undocumented
			z.xor(byte(z.iy >> 8))
		case 0xad: // xor a,iyl XXX this is supposed to be undocumented
			z.xor(byte(z.iy))
		case 0xae: // xor (iyl+d)
			z.xor(z.bus.Read(z.iy + uint16(z.bus.Read(z.pc+2))))
		case 0xb4: // or a,iyh XXX this is supposed to be undocumented
			z.or(byte(z.iy >> 8))
		case 0xb5: // or a,iyl XXX this is supposed to be undocumented
			z.or(byte(z.iy))
		case 0xb6: // or (iyl+d)
			z.or(z.bus.Read(z.iy + uint16(z.bus.Read(z.pc+2))))
		case 0xbc: // cp a,iyh XXX this is supposed to be undocumented
			z.cp(byte(z.iy >> 8))
		case 0xbd: // cp a,iyl XXX this is supposed to be undocumented
			z.cp(byte(z.iy))
		case 0xbe: // cp (iyl+d)
			z.cp(z.bus.Read(z.iy + uint16(z.bus.Read(z.pc+2))))
		case 0xcb: // bit b,(iy+d)
			bit := z.bus.Read(z.pc+3) & 0x38 >> 3
			displacement := uint16(z.bus.Read(z.pc + 2))
			val := z.bus.Read(z.iy + displacement)
			z.bit(bit, val)
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
		case 0xf9: // ld sp,iy
			z.sp = z.iy
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
		dst = fmt.Sprintf("$%04x", address+2+uint16(int8(p[start])))
	case registerIndirect:
		if z.mode == Mode8080 {
			dst = fmt.Sprintf("%v", o.dstR[z.mode])
		} else {
			dst = fmt.Sprintf("(%v)", o.dstR[z.mode])
		}
	case extended:
		dst = fmt.Sprintf("($%04x)", uint16(p[start])|uint16(p[start+1])<<8)
	case immediate:
		dst = fmt.Sprintf("$%02x", p[start])
	case immediateExtended:
		dst = fmt.Sprintf("$%04x", uint16(p[start])|uint16(p[start+1])<<8)
	case register:
		dst = o.dstR[z.mode]
	case indirect:
		dst = fmt.Sprintf("($%02x)", p[start])
	case implied:
		dst = o.dstR[z.mode]
	case indexed:
		dst = fmt.Sprintf("(%v+$%02x)", o.dstR[z.mode], p[start])
	case bitIndexed:
		// we assume this is the 4th byte
		dst = fmt.Sprintf("%v", p[start+1]&0x38>>3)
	}

	switch o.src {
	case displacement:
		src = fmt.Sprintf("$%04x", address+2+uint16(int8(p[start])))
	case registerIndirect:
		if z.mode == Mode8080 {
			src = fmt.Sprintf("%v", o.srcR[z.mode])
		} else {
			src = fmt.Sprintf("(%v)", o.srcR[z.mode])
		}
	case extended:
		src = fmt.Sprintf("($%04x)", uint16(p[start])|uint16(p[start+1])<<8)
	case immediate:
		src = fmt.Sprintf("$%02x", p[start])
	case immediateExtended:
		src = fmt.Sprintf("$%04x", uint16(p[start])|uint16(p[start+1])<<8)
	case register:
		src = o.srcR[z.mode]
	case indirect:
		src = fmt.Sprintf("($%02x)", p[start])
	case indexed:
		src = fmt.Sprintf("(%v+$%02x)", o.srcR[z.mode], p[start])
	}

	noBytes = int(o.noBytes)
	retErr = nil
	if len(o.mnemonic) == 0 {
		noBytes = 1
		switch p[0] {
		case 0xcb, 0xdd, 0xed, 0xfd:
			opc = fmt.Sprintf("%02x %02x", p[0], p[1])
			noBytes = 2
		default:
			opc = fmt.Sprintf("%02x", p[0])
		}
		mnemonic = "INVALID"
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
