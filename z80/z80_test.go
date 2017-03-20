package z80

import (
	"testing"

	"github.com/marcopeereboom/toyz80/bus"
)

func TestInstructions(t *testing.T) {
	tests := []struct {
		name       string
		opc        string
		dst        string
		src        string
		data       []byte
		init       func(z *z80)
		expect     func(z *z80) bool
		err        error
		dontSkipPC bool
	}{
		// 0x00
		{
			name:   "nop",
			opc:    "nop",
			data:   []byte{0x00},
			expect: func(z *z80) bool { return z.pc == 0x0001 },
		},
		// 0x01
		{
			name: "ld bc,nn",
			opc:  "ld",
			dst:  "bc",
			src:  "$55aa",
			data: []byte{0x01, 0xaa, 0x55},
			expect: func(z *z80) bool {
				return 0x55aa == z.bc && z.pc == 0x0003
			},
		},
		{
			name: "ld bc,nn",
			opc:  "ld",
			dst:  "bc",
			src:  "$ffff",
			data: []byte{0x01, 0xff, 0xff},
			expect: func(z *z80) bool {
				return 0xffff == z.bc && z.pc == 0x0003
			},
		},
		// 0x02
		{

			name: "ld (bc),a",
			opc:  "ld",
			dst:  "(bc)",
			src:  "a",
			data: []byte{0x02},
			init: func(z *z80) { z.af = 0xff00; z.bc = 0x1122 },
			expect: func(z *z80) bool {
				return z.af == 0xff00 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xff
			},
		},
		// 0x03
		{
			name: "inc bc",
			opc:  "inc",
			dst:  "bc",
			src:  "",
			data: []byte{0x03},
			init: func(z *z80) { z.bc = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0x1123
			},
		},
		{
			name: "inc bc == -1",
			opc:  "inc",
			dst:  "bc",
			src:  "",
			data: []byte{0x03},
			init: func(z *z80) { z.bc = 0xffff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0x0
			},
		},
		{
			name: "inc bc == 0x7fff",
			opc:  "inc",
			dst:  "bc",
			src:  "",
			data: []byte{0x03},
			init: func(z *z80) { z.bc = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0x8000
			},
		},
		// 0x04
		{
			name: "inc b",
			opc:  "inc",
			dst:  "b",
			src:  "",
			data: []byte{0x04},
			init: func(z *z80) { z.bc = 0x11a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0x12a5 &&
					z.af&sign == 0 && z.af&zero == 0 &&
					z.af&parity == 0 && z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc b == -1",
			opc:  "inc",
			dst:  "b",
			src:  "",
			data: []byte{0x04},
			init: func(z *z80) { z.bc = 0xffa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0x00a5 &&
					z.af&sign == 0 && z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc b == 0x7f",
			opc:  "inc",
			dst:  "b",
			src:  "",
			data: []byte{0x04},
			init: func(z *z80) { z.bc = 0x7fa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0x80a5 &&
					z.af&sign == sign && z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		// 0x0a
		{

			name: "ld a,(bc)",
			opc:  "ld",
			dst:  "a",
			src:  "(bc)",
			data: []byte{0x0a},
			init: func(z *z80) {
				z.bc = 0x1122
				z.bus.Write(0x1122, 0xaa)
			},
			expect: func(z *z80) bool {
				return z.af == 0xaa00 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xaa
			},
		},
		// 0x18
		{
			name: "jr positive",
			opc:  "jr",
			dst:  "$0005",
			data: []byte{0x18, 0x03},
			expect: func(z *z80) bool {
				return z.pc == 0x0005
			},
			dontSkipPC: true,
		},
		{
			name: "jr negative",
			opc:  "jr",
			dst:  "$ffff",
			data: []byte{0x18, 0xfd},
			expect: func(z *z80) bool {
				return z.pc == 0xffff
			},
			dontSkipPC: true,
		},
		// 0x2f
		{

			name: "cpl",
			opc:  "cpl",
			data: []byte{0x2f},
			init: func(z *z80) { z.af = 0xa500 },
			expect: func(z *z80) bool {
				return z.af&0xff00 == 0x5a00 && z.pc == 0x0001 &&
					z.af&addsub == addsub &&
					z.af&halfCarry == halfCarry
			},
		},
		// 0x37
		{

			name: "scf",
			opc:  "scf",
			data: []byte{0x37},
			init: func(z *z80) { z.af = 0xff00 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x3f
		{

			name: "ccf (0xff)",
			opc:  "ccf",
			data: []byte{0x3f},
			init: func(z *z80) { z.af = 0x00ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{

			name: "ccf (0x00)",
			opc:  "ccf",
			data: []byte{0x3f},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x1a
		{

			name: "ld a,(de)",
			opc:  "ld",
			dst:  "a",
			src:  "(de)",
			data: []byte{0x1a},
			init: func(z *z80) {
				z.de = 0x1122
				z.bus.Write(0x1122, 0xaa)
			},
			expect: func(z *z80) bool {
				return z.af == 0xaa00 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xaa
			},
		},
		// 0x31 ld sp,nn
		{

			name: "ld sp,nn",
			opc:  "ld",
			dst:  "sp",
			src:  "$55aa",
			data: []byte{0x31, 0xaa, 0x55},
			expect: func(z *z80) bool {
				return 0x55aa == z.sp && z.pc == 0x0003
			},
		},
		// 0x3a
		{

			name: "ld a,nn",
			opc:  "ld",
			dst:  "a",
			src:  "($55aa)",
			data: []byte{0x3a, 0xaa, 0x55},
			init: func(z *z80) {
				z.bus.Write(0x55aa, 0xff)
			},
			expect: func(z *z80) bool {
				return z.bus.Read(0x55aa) == byte(z.af>>8) &&
					z.af == 0xff00 && z.pc == 0x0003
			},
		},
		// 0x3e
		{

			name: "ld a,n",
			opc:  "ld",
			dst:  "a",
			src:  "$55",
			data: []byte{0x3e, 0x55},
			expect: func(z *z80) bool {
				return z.af == 0x5500 && z.pc == 0x0002
			},
		},
		// 0x76
		{

			name: "halt",
			opc:  "halt",
			dst:  "",
			src:  "",
			data: []byte{0x76},
			expect: func(z *z80) bool {
				return z.pc == 0x0000
			},
			err:        ErrHalt,
			dontSkipPC: true,
		},
		// 0x78
		{
			name: "ld a,b",
			opc:  "ld",
			dst:  "a",
			src:  "b",
			data: []byte{0x78},
			init: func(z *z80) { z.af = 0x1100; z.bc = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2200 == z.af && z.pc == 0x0001
			},
		},
		// 0x7f
		{
			name: "ld a,a",
			opc:  "ld",
			dst:  "a",
			src:  "a",
			data: []byte{0x7f},
			init: func(z *z80) { z.af = 0x11a5 },
			expect: func(z *z80) bool {
				return 0x11a5 == z.af && z.pc == 0x0001
			},
		},
		// 0xc2
		{
			name: "jmp nz,nn (Z set)",
			opc:  "jp",
			dst:  "nz",
			src:  "$1122",
			data: []byte{0xc2, 0x22, 0x11},
			init: func(z *z80) { z.af = 0xff00 | zero },
			expect: func(z *z80) bool {
				return z.pc == 0x0003
			},
			dontSkipPC: true,
		},
		{
			name: "jmp nz,nn (Z clear)",
			opc:  "jp",
			dst:  "nz",
			src:  "$1122",
			data: []byte{0xc2, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x1122
			},
			dontSkipPC: true,
		},
		// 0xc3
		{
			name: "jp nn",
			opc:  "jp",
			dst:  "$1122",
			src:  "",
			data: []byte{0xc3, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x1122
			},
			dontSkipPC: true,
		},
		// 0xca
		{
			name: "jp z,nn (Z set)",
			opc:  "jp",
			dst:  "z",
			src:  "$1122",
			data: []byte{0xca, 0x22, 0x11},
			init: func(z *z80) { z.af = 0xff00 | zero },
			expect: func(z *z80) bool {
				return z.pc == 0x1122
			},
			dontSkipPC: true,
		},
		{
			name: "jp z,nn (Z clear)",
			opc:  "jp",
			dst:  "z",
			src:  "$1122",
			data: []byte{0xca, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x0003
			},
			dontSkipPC: true,
		},
		// 0xd3
		{
			name: "out (n),a",
			opc:  "out",
			dst:  "($aa)",
			src:  "a",
			data: []byte{0xd3, 0xaa},
			init: func(z *z80) { z.af = 0xff00 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002
			},
		},
		// 0xeb
		{
			name: "ex de,hl",
			opc:  "ex",
			dst:  "de",
			src:  "hl",
			data: []byte{0xeb},
			init: func(z *z80) { z.de = 0x1122; z.hl = 0x3344 },
			expect: func(z *z80) bool {
				return 0x1122 == z.hl && 0x3344 == z.de &&
					z.pc == 0x0001
			},
		},
		// 0xed 0x44 neg
		// XXX add more test cases for all the flags
		{
			name: "neg 0",
			opc:  "neg",
			dst:  "",
			src:  "",
			data: []byte{0xed, 0x44},
			expect: func(z *z80) bool {
				return 0x0000 == z.af&0xff00 && z.pc == 0x0002 &&
					z.af&zero == zero && z.af&sign == 0 &&
					z.af&addsub == addsub
			},
		},
		{
			name: "neg 1",
			opc:  "neg",
			dst:  "",
			src:  "",
			data: []byte{0xed, 0x44},
			init: func(z *z80) { z.af = 0x0100 },
			expect: func(z *z80) bool {
				return 0xff00 == z.af&0xff00 &&
					z.pc == 0x0002 && z.af&zero == 0 &&
					z.af&sign == sign &&
					z.af&addsub == addsub
			},
		},
		{
			name: "neg -1",
			opc:  "neg",
			dst:  "",
			src:  "",
			data: []byte{0xed, 0x44},
			init: func(z *z80) { z.af = 0xff00 },
			expect: func(z *z80) bool {
				return 0x0100 == z.af&0xff00 &&
					z.pc == 0x0002 && z.af&zero == 0 &&
					z.af&sign == 0
			},
		},
	}

	for _, test := range tests {
		t.Logf("running: %v", test.name)

		devices := []bus.Device{
			bus.Device{
				Name:  "RAM",
				Start: 0x0000,
				Size:  65536,
				Type:  bus.DeviceRAM,
				Image: test.data,
			},
			bus.Device{
				Name:  "Dummy",
				Start: 0xaa,
				Size:  1,
				Type:  bus.DeviceDummy,
			},
		}
		bus, err := bus.New(devices)
		if err != nil {
			t.Fatalf("%v: bus %v", test.name, err)
		}

		z, err := New(ModeZ80, bus)
		if err != nil {
			t.Fatalf("%v: z80 %v", test.name, err)
		}

		if test.init != nil {
			test.init(z)
		}

		err = z.Step()
		if err != test.err {
			t.Fatalf("%v: step %v", test.name, err)
		}

		opc, dst, src, x := z.DisassembleComponents(0)
		if opc != test.opc {
			t.Fatalf("%v: invalid opcode got %v expected %v",
				test.name, opc, test.opc)
		}
		if dst != test.dst {
			t.Fatalf("%v: invalid destination got %v expected %v",
				test.name, dst, test.dst)
		}
		if src != test.src {
			t.Fatalf("%v: invalid source got %v expected %v",
				test.name, src, test.src)
		}

		if test.dontSkipPC == false && uint16(x) != z.pc {
			t.Fatalf("%v: invalid pc got $%04x expected $%04x",
				test.name, x, z.pc)
		}

		if !test.expect(z) {
			t.Fatalf("%v: failed %v", test.name, z.DumpRegisters())
		}
	}

	// Minimal test to verify there is a unit test implemented.
	for o := range opcodes {
		if len(opcodes[o].mnemonic) == 0 {
			continue
		}

		for _, test := range tests {
			if byte(o) == test.data[0] {
				goto next
			}
		}
		t.Fatalf("not implemented: 0x%02x", o)
	next:
	}
}
