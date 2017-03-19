package z80

import (
	"errors"
	"fmt"
)

var (
	ErrHalt               = errors.New("halt")
	ErrInvalidInstruction = errors.New("invalid instruction")
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
	af uint16 // A & Flags
	bc uint16 // B & C
	de uint16 // D & E
	hl uint16 // H & L
	ix uint16 // index register X
	iy uint16 // index register Y

	pc uint16 // program counter
	sp uint16 // stack pointer

	memory []byte

	totalCycles uint64
	mode        CPUMode
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
func New(mode CPUMode) *z80 {
	return &z80{
		mode:   mode,
		memory: make([]byte, 65536),
	}
}

// Reset resets the CPU.  If cold is true then memory is zeroed.
func (z *z80) Reset(cold bool) {
	if cold {
		// toss memory.
		for i := 0; i < len(z.memory); i++ {
			z.memory[i] = 0
		}
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
	z.totalCycles += o.noCycles
	z.pc += uint16(o.noBytes)
}

// inc8H adds 1 to the low byte of *p and sets the flags.
func (z *z80) inc8L(p *uint16) {
	oldB := byte(*p & 0x00ff)
	newB := oldB + 1
	*p = uint16(newB) | *p&0xff00
	//Condition Bits Affected
	//S is set if result is negative; otherwise, it is reset.
	//Z is set if result is 0; otherwise, it is reset.
	//H is set if carry from bit 3; otherwise, it is reset.
	//P/V is set if r was 7Fh before operation; otherwise, it is reset.
	//N is reset.
	//C is not affected
	z.evalS(newB)
	z.evalZ(newB)
	if oldB == 0x7f {
		z.af |= parity
	} else {
		z.af &^= parity
	}
	z.evalH(oldB, 1)
	z.af &^= addsub
}

// inc8H adds 1 to the high byte of *p and sets the flags.
func (z *z80) inc8H(p *uint16) {
	oldB := byte(*p >> 8)
	newB := oldB + 1
	*p = uint16(newB)<<8 | *p&0x00ff
	//Condition Bits Affected
	//S is set if result is negative; otherwise, it is reset.
	//Z is set if result is 0; otherwise, it is reset.
	//H is set if carry from bit 3; otherwise, it is reset.
	//P/V is set if r was 7Fh before operation; otherwise, it is reset.
	//N is reset.
	//C is not affected
	z.evalS(newB)
	z.evalZ(newB)
	if oldB == 0x7f {
		z.af |= parity
	} else {
		z.af &^= parity
	}
	z.evalH(oldB, 1)
	z.af &^= addsub
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
	opc := z.memory[z.pc]
	opcodeStruct := &opcodes[opc]
	pi := z.genericPostInstruction

	switch opc {
	case 0x00: // nop
	case 0x01: // ld bc,nn
		z.bc = uint16(z.memory[z.pc+1]) | uint16(z.memory[z.pc+2])<<8
	case 0x02: // ld (bc),a
		z.memory[z.bc] = byte(z.af >> 8)
	case 0x03: //inc bc
		z.bc += 1
		// Condition bits are no affected
	case 0x04: //inc b
		z.inc8H(&z.bc)
	case 0x0a: // ld a,(bc)
		z.af = uint16(z.memory[z.bc])<<8 | z.af&0x00ff
	case 0x1a: // ld a,(de)
		z.af = uint16(z.memory[z.de])<<8 | z.af&0x00ff
	case 0x18:
		// We just won the world championship of dumb casting.
		z.pc = z.pc + 2 + uint16(int8(z.memory[z.pc+1]))
		z.totalCycles += opcodeStruct.noCycles
		return nil
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
		z.sp = uint16(z.memory[z.pc+1]) | uint16(z.memory[z.pc+2])<<8
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
		z.af = uint16(z.memory[uint16(z.memory[z.pc+1])|uint16(z.memory[z.pc+2])<<8]) << 8
	case 0x3c: // inc a
		z.inc8L(&z.af)
	case 0x3e: // ld a,n
		z.af = uint16(z.memory[z.pc+1]) << 8
	case 0x76: // halt
		z.totalCycles += opcodeStruct.noCycles
		return ErrHalt
	case 0x78: // ld a,b
		z.af = z.af&0x00ff | z.bc&0xff00
	case 0x7f: // ld a,a
		// basically nop since it doesn't affect flags
	case 0xc2: // jmp nz,nn
		if z.af&zero == 0 {
			z.pc = uint16(z.memory[z.pc+1]) | uint16(z.memory[z.pc+2])<<8
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
	case 0xc3: // jmp nn
		z.pc = uint16(z.memory[z.pc+1]) | uint16(z.memory[z.pc+2])<<8
		z.totalCycles += opcodeStruct.noCycles
		return nil
	case 0xca: // jmp z,nn
		if z.af&zero == zero {
			z.pc = uint16(z.memory[z.pc+1]) | uint16(z.memory[z.pc+2])<<8
			z.totalCycles += opcodeStruct.noCycles
			return nil
		}
	case 0xeb: // ex de,hl
		t := z.hl
		z.hl = z.de
		z.de = t
	case 0xed: // z80 only
		switch z.memory[z.pc+1] {
		case 0x44: // neg
			// Condition Bits Affected
			// S is set if result is negative; otherwise, it is reset.
			// Z is set if result is 0; otherwise, it is reset.
			// H is set if borrow from bit 4; otherwise, it is reset.
			// P/V is set if Accumulator was 80h before operation; otherwise, it is reset.
			// N is set.
			// C is set if Accumulator was not 00h before operation; otherwise, it is reset.
			opcodeStruct = &opcodesED[0x44]
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
			return ErrInvalidInstruction
		}
	default:
		//fmt.Printf("opcode %x\n", opcode)
		return ErrInvalidInstruction
	}

	pi(opcodeStruct)

	return nil
}

// Disassemble disassembles the instruction at the provided address and also
// returns the number of bytes consumed.
func (z *z80) Disassemble(address uint16) (string, int) {
	opc, dst, src, noBytes := z.DisassembleComponents(address)

	if dst != "" && src != "" {
		src = "," + src
	}
	dst += src

	s := fmt.Sprintf("%-6v%-4v", opc, dst)

	return s, noBytes
}

// DisassembleComponents disassmbles the instruction at the provided address
// and returns all compnonts of the instruction (opcode, destination, source).
func (z *z80) DisassembleComponents(address uint16) (opc string, dst string, src string, noBytes int) {
	o := &opcodes[z.memory[address]]
	if o.multiByte {
		switch z.memory[address] {
		case 0xed:
			o = &opcodesED[z.memory[address+1]]
		}
	}
	switch o.dst {
	case condition:
		dst = o.dstR[z.mode]
	case displacement:
		dst = fmt.Sprintf("$%04x", address+2+
			uint16(int8(z.memory[address+1])))
	case registerIndirect:
		if z.mode == Mode8080 {
			dst = fmt.Sprintf("%v", o.dstR[z.mode])
		} else {
			dst = fmt.Sprintf("(%v)", o.dstR[z.mode])
		}
	case immediateExtended:
		dst = fmt.Sprintf("$%04x", uint16(z.memory[address+1])|
			uint16(z.memory[address+2])<<8)
	case register:
		dst = o.dstR[z.mode]
	}

	switch o.src {
	case registerIndirect:
		if z.mode == Mode8080 {
			src = fmt.Sprintf("%v", o.srcR[z.mode])
		} else {
			src = fmt.Sprintf("(%v)", o.srcR[z.mode])
		}
	case extended:
		src = fmt.Sprintf("($%04x)", uint16(z.memory[address+1])|
			uint16(z.memory[address+2])<<8)
	case immediate:
		src = fmt.Sprintf("$%02x", z.memory[address+1])
	case immediateExtended:
		src = fmt.Sprintf("$%04x", uint16(z.memory[address+1])|
			uint16(z.memory[address+2])<<8)
	case register:
		src = o.srcR[z.mode]
	}

	noBytes = int(o.noBytes)
	opc = o.mnemonic[z.mode]

	// if opcode is invalid skip it.
	if opc == "" {
		opc = "INVALID"
		noBytes = 1
	}

	return
}

// block describes an origination address and a data payload for the loader.
type block struct {
	org  uint16
	data []byte
}

// Load loads the provided blocks into the CPU memory.
func (z *z80) Load(blocks []block) error {
	for i := range blocks {
		if len(blocks[i].data)+int(blocks[i].org) > len(z.memory) {
			return fmt.Errorf("block %v out of bounds", i)
		}

		copy(z.memory[blocks[i].org:], blocks[i].data)
	}

	return nil
}

func (z *z80) Trace() ([]string, []string, error) {
	trace := make([]string, 0, 1024)
	registers := make([]string, 0, 1024)

	for {
		s, _ := z.Disassemble(z.pc)
		trace = append(trace, fmt.Sprintf("%04x: %v", z.pc, s))
		err := z.Step()
		registers = append(registers, z.DumpRegisters())
		if err != nil {
			return trace, registers, err
		}
	}
	return trace, registers, nil
}

//func main() {
//	z := New()
//	err := z.Load([]block{
//		{
//			org: 0x0,
//			data: []byte{
//				0x31, 0xc4, 0x07, // ld sp,$07c4
//				0xc3, 0x00, 0x10, // jp $1000
//			},
//		},
//		{
//			org: 0x1000,
//			data: []byte{
//				0x3a, 0xaa, 0x55, // ld a,($55aa)
//				0x0a,       // ld a,(bc)
//				0x1a,       // ld a,(de)
//				0x7f,       // ld a,a
//				0x78,       // ld a,b
//				0x3e, 0xff, // ld a,$ff
//				0x02,             // ld (bc),a
//				0x01, 0x00, 0x10, // ld bc,$1000
//				0x00, // nop
//				0xeb, // ex de,hl
//				0x76, // halt
//			},
//		},
//		{
//			org:  0x55aa,
//			data: []byte{0x23, 0x45, 0x67},
//		},
//	})
//	z.bc = 0x55ab
//	z.de = 0x55ac
//	if err != nil {
//		fmt.Printf("load: %v\n", err)
//		return
//	}
//
//	trace, registers, err := z.Trace(ModeZ80)
//	//trace, registers, err := z.Trace(Mode8080)
//	if err != nil && err != ErrHalt {
//		fmt.Printf("trace: %v\n", err)
//		return
//	}
//	for i, line := range trace {
//		fmt.Printf("%-25s%s\n", line, registers[i])
//	}
//
//	fmt.Printf("cycles: %v\n", z.totalCycles)
//}
