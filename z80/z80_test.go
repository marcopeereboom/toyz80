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
		// 0x06
		{

			name: "ld b,n",
			opc:  "ld",
			dst:  "b",
			src:  "$55",
			data: []byte{0x06, 0x55},
			init: func(z *z80) { z.bc = 0x1122 },
			expect: func(z *z80) bool {
				return z.bc == 0x5522 && z.pc == 0x0002
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
		// 0x0b
		{
			name: "dec bc",
			opc:  "dec",
			dst:  "bc",
			src:  "",
			data: []byte{0x0b},
			init: func(z *z80) { z.bc = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0x1121
			},
		},
		{
			name: "dec bc == 0",
			opc:  "dec",
			dst:  "bc",
			src:  "",
			data: []byte{0x0b},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0xffff
			},
		},
		{
			name: "dec bc == 0x8000",
			opc:  "dec",
			dst:  "bc",
			src:  "",
			data: []byte{0x0b},
			init: func(z *z80) { z.bc = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0x7fff
			},
		},
		// 0x0e
		{

			name: "ld c,n",
			opc:  "ld",
			dst:  "c",
			src:  "$55",
			data: []byte{0x0e, 0x55},
			init: func(z *z80) { z.bc = 0x2233 },
			expect: func(z *z80) bool {
				return z.bc == 0x2255 && z.pc == 0x0002
			},
		},
		// 0x11 ld de,nn
		{

			name: "ld de,nn",
			opc:  "ld",
			dst:  "de",
			src:  "$55aa",
			data: []byte{0x11, 0xaa, 0x55},
			expect: func(z *z80) bool {
				return 0x55aa == z.de && z.pc == 0x0003
			},
		},
		// 0x12
		{

			name: "ld (de),a",
			opc:  "ld",
			dst:  "(de)",
			src:  "a",
			data: []byte{0x12},
			init: func(z *z80) { z.af = 0xffee; z.de = 0x1122 },
			expect: func(z *z80) bool {
				return z.af == 0xffee && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xff
			},
		},
		// 0x13
		{
			name: "inc de",
			opc:  "inc",
			dst:  "de",
			src:  "",
			data: []byte{0x13},
			init: func(z *z80) { z.de = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x1123
			},
		},
		{
			name: "inc de == -1",
			opc:  "inc",
			dst:  "de",
			src:  "",
			data: []byte{0x13},
			init: func(z *z80) { z.de = 0xffff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x0
			},
		},
		{
			name: "inc de == 0x7fff",
			opc:  "inc",
			dst:  "de",
			src:  "",
			data: []byte{0x13},
			init: func(z *z80) { z.de = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x8000
			},
		},
		// 0x16
		{

			name: "ld d,n",
			opc:  "ld",
			dst:  "d",
			src:  "$55",
			data: []byte{0x16, 0x55},
			init: func(z *z80) { z.de = 0x2233 },
			expect: func(z *z80) bool {
				return z.de == 0x5533 && z.pc == 0x0002
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
		// 0x1b
		{
			name: "dec de",
			opc:  "dec",
			dst:  "de",
			src:  "",
			data: []byte{0x1b},
			init: func(z *z80) { z.de = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x1121
			},
		},
		{
			name: "dec de == 0",
			opc:  "dec",
			dst:  "de",
			src:  "",
			data: []byte{0x1b},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0xffff
			},
		},
		{
			name: "dec de == 0x8000",
			opc:  "dec",
			dst:  "de",
			src:  "",
			data: []byte{0x1b},
			init: func(z *z80) { z.de = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x7fff
			},
		},
		// 0x1e
		{

			name: "ld e,n",
			opc:  "ld",
			dst:  "e",
			src:  "$55",
			data: []byte{0x1e, 0x55},
			init: func(z *z80) { z.de = 0x2233 },
			expect: func(z *z80) bool {
				return z.de == 0x2255 && z.pc == 0x0002
			},
		},
		// 0x21 ld hl,nn
		{

			name: "ld hl,nn",
			opc:  "ld",
			dst:  "hl",
			src:  "$55aa",
			data: []byte{0x21, 0xaa, 0x55},
			expect: func(z *z80) bool {
				return 0x55aa == z.hl && z.pc == 0x0003
			},
		},
		// 0x23
		{
			name: "inc hl",
			opc:  "inc",
			dst:  "hl",
			src:  "",
			data: []byte{0x23},
			init: func(z *z80) { z.hl = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x1123
			},
		},
		{
			name: "inc hl == -1",
			opc:  "inc",
			dst:  "hl",
			src:  "",
			data: []byte{0x23},
			init: func(z *z80) { z.hl = 0xffff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x0
			},
		},
		{
			name: "inc hl == 0x7fff",
			opc:  "inc",
			dst:  "hl",
			src:  "",
			data: []byte{0x23},
			init: func(z *z80) { z.hl = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x8000
			},
		},
		// 0x26
		{

			name: "ld h,n",
			opc:  "ld",
			dst:  "h",
			src:  "$55",
			data: []byte{0x26, 0x55},
			init: func(z *z80) { z.hl = 0x2233 },
			expect: func(z *z80) bool {
				return z.hl == 0x5533 && z.pc == 0x0002
			},
		},
		// 0x28
		{
			name: "jr z,negative",
			opc:  "jr",
			dst:  "z",
			src:  "$ffff",
			data: []byte{0x28, 0xfd},
			init: func(z *z80) { z.af = zero },
			expect: func(z *z80) bool {
				return z.pc == 0xffff
			},
			dontSkipPC: true,
		},
		{
			name: "jr z,positive",
			opc:  "jr",
			dst:  "z",
			src:  "$0005",
			data: []byte{0x28, 0x3},
			init: func(z *z80) { z.af = zero },
			expect: func(z *z80) bool {
				return z.pc == 0x0005
			},
			dontSkipPC: true,
		},
		{
			name: "jr z,negative don't follow",
			opc:  "jr",
			dst:  "z",
			src:  "$ffff",
			data: []byte{0x28, 0xfd},
			expect: func(z *z80) bool {
				return z.pc == 0x0002
			},
			dontSkipPC: true,
		},
		{
			name: "jr z,positive don't follow",
			opc:  "jr",
			dst:  "z",
			src:  "$0005",
			data: []byte{0x28, 0x3},
			expect: func(z *z80) bool {
				return z.pc == 0x0002
			},
			dontSkipPC: true,
		},
		// 0x2b
		{
			name: "dec hl",
			opc:  "dec",
			dst:  "hl",
			src:  "",
			data: []byte{0x2b},
			init: func(z *z80) { z.hl = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x1121
			},
		},
		{
			name: "dec hl == 0",
			opc:  "dec",
			dst:  "hl",
			src:  "",
			data: []byte{0x2b},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xffff
			},
		},
		{
			name: "dec hl == 0x8000",
			opc:  "dec",
			dst:  "hl",
			src:  "",
			data: []byte{0x2b},
			init: func(z *z80) { z.hl = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x7fff
			},
		},
		// 0x2e
		{

			name: "ld l,n",
			opc:  "ld",
			dst:  "l",
			src:  "$55",
			data: []byte{0x2e, 0x55},
			init: func(z *z80) { z.hl = 0x2233 },
			expect: func(z *z80) bool {
				return z.hl == 0x2255 && z.pc == 0x0002
			},
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
		// 0x32
		{

			name: "ld (nn),a",
			opc:  "ld",
			dst:  "$ffee",
			src:  "a",
			data: []byte{0x32, 0xee, 0xff},
			init: func(z *z80) {
				z.af = 0x1122
				z.bus.Write(0xffee, 0x55)
			},
			expect: func(z *z80) bool {
				return z.af == 0x1122 && z.pc == 0x0003 &&
					z.bus.Read(0xffee) == 0x11
			},
		},
		// 0x33
		{
			name: "inc sp",
			opc:  "inc",
			dst:  "sp",
			src:  "",
			data: []byte{0x33},
			init: func(z *z80) { z.sp = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0x1123
			},
		},
		{
			name: "inc sp == -1",
			opc:  "inc",
			dst:  "sp",
			src:  "",
			data: []byte{0x33},
			init: func(z *z80) { z.sp = 0xffff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0x0
			},
		},
		{
			name: "inc sp == 0x7fff",
			opc:  "inc",
			dst:  "sp",
			src:  "",
			data: []byte{0x33},
			init: func(z *z80) { z.sp = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0x8000
			},
		},
		// 0x36
		{

			name: "ld (hl),n",
			opc:  "ld",
			dst:  "(hl)",
			src:  "$55",
			data: []byte{0x36, 0x55},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bus.Write(z.hl, 0xaa)
			},
			expect: func(z *z80) bool {
				return z.hl == 0x1122 && z.pc == 0x0002 &&
					z.bus.Read(0x1122) == 0x55
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
		// 0x3b
		{
			name: "dec sp",
			opc:  "dec",
			dst:  "sp",
			src:  "",
			data: []byte{0x3b},
			init: func(z *z80) { z.sp = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0x1121
			},
		},
		{
			name: "dec sp == 0",
			opc:  "dec",
			dst:  "sp",
			src:  "",
			data: []byte{0x3b},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xffff
			},
		},
		{
			name: "dec sp == 0x8000",
			opc:  "dec",
			dst:  "sp",
			src:  "",
			data: []byte{0x3b},
			init: func(z *z80) { z.sp = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0x7fff
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
			init: func(z *z80) { z.af = 0x2233 },
			expect: func(z *z80) bool {
				return z.af == 0x5533 && z.pc == 0x0002
			},
		},
		// 0x40
		{
			name: "ld b,b",
			opc:  "ld",
			dst:  "b",
			src:  "b",
			data: []byte{0x40},
			init: func(z *z80) { z.bc = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2233 == z.bc && z.pc == 0x0001
			},
		},
		// 0x41
		{
			name: "ld b,c",
			opc:  "ld",
			dst:  "b",
			src:  "c",
			data: []byte{0x41},
			init: func(z *z80) { z.bc = 0x2233 },
			expect: func(z *z80) bool {
				return 0x3333 == z.bc && z.pc == 0x0001
			},
		},
		// 0x42
		{
			name: "ld b,d",
			opc:  "ld",
			dst:  "b",
			src:  "d",
			data: []byte{0x42},
			init: func(z *z80) { z.bc = 0x2233; z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x4433 == z.bc && z.pc == 0x0001
			},
		},
		// 0x43
		{
			name: "ld b,e",
			opc:  "ld",
			dst:  "b",
			src:  "e",
			data: []byte{0x43},
			init: func(z *z80) { z.bc = 0x2233; z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x5533 == z.bc && z.pc == 0x0001
			},
		},
		// 0x44
		{
			name: "ld b,h",
			opc:  "ld",
			dst:  "b",
			src:  "h",
			data: []byte{0x44},
			init: func(z *z80) { z.bc = 0x2233; z.hl = 0x4455 },
			expect: func(z *z80) bool {
				return 0x4433 == z.bc && z.pc == 0x0001
			},
		},
		// 0x45
		{
			name: "ld b,l",
			opc:  "ld",
			dst:  "b",
			src:  "l",
			data: []byte{0x45},
			init: func(z *z80) { z.bc = 0x2233; z.hl = 0x4455 },
			expect: func(z *z80) bool {
				return 0x5533 == z.bc && z.pc == 0x0001
			},
		},
		// 0x46
		{

			name: "ld b,(hl)",
			opc:  "ld",
			dst:  "b",
			src:  "(hl)",
			data: []byte{0x46},
			init: func(z *z80) {
				z.bc = 0x3344
				z.hl = 0x1122
				z.bus.Write(0x1122, 0xaa)
			},
			expect: func(z *z80) bool {
				return z.bc == 0xaa44 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xaa
			},
		},
		// 0x47
		{
			name: "ld b,a",
			opc:  "ld",
			dst:  "b",
			src:  "a",
			data: []byte{0x47},
			init: func(z *z80) { z.af = 0x1144; z.bc = 0x2233 },
			expect: func(z *z80) bool {
				return 0x1133 == z.bc && z.pc == 0x0001
			},
		},
		// 0x48
		{
			name: "ld c,b",
			opc:  "ld",
			dst:  "c",
			src:  "b",
			data: []byte{0x48},
			init: func(z *z80) { z.af = 0x1144; z.bc = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2222 == z.bc && z.pc == 0x0001
			},
		},
		// 0x49
		{
			name: "ld c,c",
			opc:  "ld",
			dst:  "c",
			src:  "c",
			data: []byte{0x49},
			init: func(z *z80) { z.bc = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2233 == z.bc && z.pc == 0x0001
			},
		},
		// 0x4a
		{
			name: "ld c,d",
			opc:  "ld",
			dst:  "c",
			src:  "d",
			data: []byte{0x4a},
			init: func(z *z80) { z.bc = 0x2233; z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2244 == z.bc && z.pc == 0x0001
			},
		},
		// 0x4b
		{
			name: "ld c,e",
			opc:  "ld",
			dst:  "c",
			src:  "e",
			data: []byte{0x4b},
			init: func(z *z80) { z.bc = 0x2233; z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2255 == z.bc && z.pc == 0x0001
			},
		},
		// 0x4c
		{
			name: "ld c,h",
			opc:  "ld",
			dst:  "c",
			src:  "h",
			data: []byte{0x4c},
			init: func(z *z80) { z.bc = 0x2233; z.hl = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2244 == z.bc && z.pc == 0x0001
			},
		},
		// 0x4d
		{
			name: "ld c,l",
			opc:  "ld",
			dst:  "c",
			src:  "l",
			data: []byte{0x4d},
			init: func(z *z80) { z.bc = 0x2233; z.hl = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2255 == z.bc && z.pc == 0x0001
			},
		},
		// 0x4e
		{

			name: "ld c,(hl)",
			opc:  "ld",
			dst:  "c",
			src:  "(hl)",
			data: []byte{0x4e},
			init: func(z *z80) {
				z.bc = 0x3344
				z.hl = 0x1122
				z.bus.Write(0x1122, 0xaa)
			},
			expect: func(z *z80) bool {
				return z.bc == 0x33aa && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xaa
			},
		},
		// 0x4f
		{
			name: "ld c,a",
			opc:  "ld",
			dst:  "c",
			src:  "a",
			data: []byte{0x4f},
			init: func(z *z80) { z.af = 0x1144; z.bc = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2211 == z.bc && z.pc == 0x0001
			},
		},
		// 0x50
		{
			name: "ld d,b",
			opc:  "ld",
			dst:  "d",
			src:  "b",
			data: []byte{0x50},
			init: func(z *z80) { z.bc = 0x2233; z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2255 == z.de && z.pc == 0x0001
			},
		},
		// 0x51
		{
			name: "ld d,c",
			opc:  "ld",
			dst:  "d",
			src:  "c",
			data: []byte{0x51},
			init: func(z *z80) { z.bc = 0x2233; z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x3355 == z.de && z.pc == 0x0001
			},
		},
		// 0x52
		{
			name: "ld d,d",
			opc:  "ld",
			dst:  "d",
			src:  "d",
			data: []byte{0x52},
			init: func(z *z80) { z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x4455 == z.de && z.pc == 0x0001
			},
		},
		// 0x53
		{
			name: "ld d,e",
			opc:  "ld",
			dst:  "d",
			src:  "e",
			data: []byte{0x53},
			init: func(z *z80) { z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x5555 == z.de && z.pc == 0x0001
			},
		},
		// 0x54
		{
			name: "ld d,h",
			opc:  "ld",
			dst:  "d",
			src:  "h",
			data: []byte{0x54},
			init: func(z *z80) { z.de = 0x2233; z.hl = 0x4455 },
			expect: func(z *z80) bool {
				return 0x4433 == z.de && z.pc == 0x0001
			},
		},
		// 0x55
		{
			name: "ld d,l",
			opc:  "ld",
			dst:  "d",
			src:  "l",
			data: []byte{0x55},
			init: func(z *z80) { z.de = 0x2233; z.hl = 0x4455 },
			expect: func(z *z80) bool {
				return 0x5533 == z.de && z.pc == 0x0001
			},
		},
		// 0x56
		{

			name: "ld d,(hl)",
			opc:  "ld",
			dst:  "d",
			src:  "(hl)",
			data: []byte{0x56},
			init: func(z *z80) {
				z.de = 0x3344
				z.hl = 0x1122
				z.bus.Write(0x1122, 0xaa)
			},
			expect: func(z *z80) bool {
				return z.de == 0xaa44 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xaa
			},
		},
		// 0x57
		{
			name: "ld d,a",
			opc:  "ld",
			dst:  "d",
			src:  "a",
			data: []byte{0x57},
			init: func(z *z80) { z.de = 0x2233; z.af = 0x4455 },
			expect: func(z *z80) bool {
				return 0x4433 == z.de && z.pc == 0x0001
			},
		},
		// 0x58
		{
			name: "ld e,b",
			opc:  "ld",
			dst:  "e",
			src:  "b",
			data: []byte{0x58},
			init: func(z *z80) { z.de = 0x2233; z.bc = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2244 == z.de && z.pc == 0x0001
			},
		},
		// 0x59
		{
			name: "ld e,c",
			opc:  "ld",
			dst:  "e",
			src:  "c",
			data: []byte{0x59},
			init: func(z *z80) { z.de = 0x2233; z.bc = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2255 == z.de && z.pc == 0x0001
			},
		},
		// 0x5a
		{
			name: "ld e,d",
			opc:  "ld",
			dst:  "e",
			src:  "d",
			data: []byte{0x5a},
			init: func(z *z80) { z.de = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2222 == z.de && z.pc == 0x0001
			},
		},
		// 0x5b
		{
			name: "ld e,e",
			opc:  "ld",
			dst:  "e",
			src:  "e",
			data: []byte{0x5b},
			init: func(z *z80) { z.de = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2233 == z.de && z.pc == 0x0001
			},
		},
		// 0x5c
		{
			name: "ld e,h",
			opc:  "ld",
			dst:  "e",
			src:  "h",
			data: []byte{0x5c},
			init: func(z *z80) { z.de = 0x2233; z.hl = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2244 == z.de && z.pc == 0x0001
			},
		},
		// 0x5d
		{
			name: "ld e,h",
			opc:  "ld",
			dst:  "e",
			src:  "l",
			data: []byte{0x5d},
			init: func(z *z80) { z.de = 0x2233; z.hl = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2255 == z.de && z.pc == 0x0001
			},
		},
		// 0x5e
		{

			name: "ld e,(hl)",
			opc:  "ld",
			dst:  "e",
			src:  "(hl)",
			data: []byte{0x5e},
			init: func(z *z80) {
				z.de = 0x3344
				z.hl = 0x1122
				z.bus.Write(0x1122, 0xaa)
			},
			expect: func(z *z80) bool {
				return z.de == 0x33aa && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xaa
			},
		},
		// 0x5f
		{
			name: "ld e,a",
			opc:  "ld",
			dst:  "e",
			src:  "a",
			data: []byte{0x5f},
			init: func(z *z80) { z.de = 0x2233; z.af = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2244 == z.de && z.pc == 0x0001
			},
		},
		// 0x60
		{
			name: "ld h,b",
			opc:  "ld",
			dst:  "h",
			src:  "b",
			data: []byte{0x60},
			init: func(z *z80) { z.hl = 0x2233; z.bc = 0x4455 },
			expect: func(z *z80) bool {
				return 0x4433 == z.hl && z.pc == 0x0001
			},
		},
		// 0x61
		{
			name: "ld h,c",
			opc:  "ld",
			dst:  "h",
			src:  "c",
			data: []byte{0x61},
			init: func(z *z80) { z.hl = 0x2233; z.bc = 0x4455 },
			expect: func(z *z80) bool {
				return 0x5533 == z.hl && z.pc == 0x0001
			},
		},
		// 0x62
		{
			name: "ld h,d",
			opc:  "ld",
			dst:  "h",
			src:  "d",
			data: []byte{0x62},
			init: func(z *z80) { z.hl = 0x2233; z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x4433 == z.hl && z.pc == 0x0001
			},
		},
		// 0x63
		{
			name: "ld h,e",
			opc:  "ld",
			dst:  "h",
			src:  "e",
			data: []byte{0x63},
			init: func(z *z80) { z.hl = 0x2233; z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x5533 == z.hl && z.pc == 0x0001
			},
		},
		// 0x64
		{
			name: "ld h,h",
			opc:  "ld",
			dst:  "h",
			src:  "h",
			data: []byte{0x64},
			init: func(z *z80) { z.hl = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2233 == z.hl && z.pc == 0x0001
			},
		},
		// 0x65
		{
			name: "ld h,l",
			opc:  "ld",
			dst:  "h",
			src:  "l",
			data: []byte{0x65},
			init: func(z *z80) { z.hl = 0x2233 },
			expect: func(z *z80) bool {
				return 0x3333 == z.hl && z.pc == 0x0001
			},
		},
		// 0x66
		{

			name: "ld h,(hl)",
			opc:  "ld",
			dst:  "h",
			src:  "(hl)",
			data: []byte{0x66},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bus.Write(0x1122, 0xaa)
			},
			expect: func(z *z80) bool {
				return z.hl == 0xaa22 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xaa
			},
		},
		// 0x67
		{
			name: "ld h,a",
			opc:  "ld",
			dst:  "h",
			src:  "a",
			data: []byte{0x67},
			init: func(z *z80) { z.hl = 0x2233; z.af = 0x4455 },
			expect: func(z *z80) bool {
				return 0x4433 == z.hl && z.pc == 0x0001
			},
		},
		// 0x68
		{
			name: "ld l,b",
			opc:  "ld",
			dst:  "l",
			src:  "b",
			data: []byte{0x68},
			init: func(z *z80) { z.hl = 0x2233; z.bc = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2244 == z.hl && z.pc == 0x0001
			},
		},
		// 0x69
		{
			name: "ld l,c",
			opc:  "ld",
			dst:  "l",
			src:  "c",
			data: []byte{0x69},
			init: func(z *z80) { z.hl = 0x2233; z.bc = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2255 == z.hl && z.pc == 0x0001
			},
		},
		// 0x6a
		{
			name: "ld l,d",
			opc:  "ld",
			dst:  "l",
			src:  "d",
			data: []byte{0x6a},
			init: func(z *z80) { z.hl = 0x2233; z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2244 == z.hl && z.pc == 0x0001
			},
		},
		// 0x6b
		{
			name: "ld l,e",
			opc:  "ld",
			dst:  "l",
			src:  "e",
			data: []byte{0x6b},
			init: func(z *z80) { z.hl = 0x2233; z.de = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2255 == z.hl && z.pc == 0x0001
			},
		},
		// 0x6c
		{
			name: "ld l,h",
			opc:  "ld",
			dst:  "l",
			src:  "h",
			data: []byte{0x6c},
			init: func(z *z80) { z.hl = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2222 == z.hl && z.pc == 0x0001
			},
		},
		// 0x6d
		{
			name: "ld l,l",
			opc:  "ld",
			dst:  "l",
			src:  "l",
			data: []byte{0x6d},
			init: func(z *z80) { z.hl = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2233 == z.hl && z.pc == 0x0001
			},
		},
		// 0x6e
		{

			name: "ld l,(hl)",
			opc:  "ld",
			dst:  "l",
			src:  "(hl)",
			data: []byte{0x6e},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bus.Write(0x1122, 0xaa)
			},
			expect: func(z *z80) bool {
				return z.hl == 0x11aa && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xaa
			},
		},
		// 0x6f
		{
			name: "ld l,a",
			opc:  "ld",
			dst:  "l",
			src:  "a",
			data: []byte{0x6f},
			init: func(z *z80) { z.hl = 0x2233; z.af = 0x4455 },
			expect: func(z *z80) bool {
				return 0x2244 == z.hl && z.pc == 0x0001
			},
		},
		// 0x70
		{

			name: "ld (hl),b",
			opc:  "ld",
			dst:  "(hl)",
			src:  "b",
			data: []byte{0x70},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bc = 0x3344
			},
			expect: func(z *z80) bool {
				return z.hl == 0x1122 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x33
			},
		},
		// 0x71
		{

			name: "ld (hl),c",
			opc:  "ld",
			dst:  "(hl)",
			src:  "c",
			data: []byte{0x71},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bc = 0x3344
			},
			expect: func(z *z80) bool {
				return z.hl == 0x1122 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x44
			},
		},
		// 0x72
		{

			name: "ld (hl),d",
			opc:  "ld",
			dst:  "(hl)",
			src:  "d",
			data: []byte{0x72},
			init: func(z *z80) {
				z.hl = 0x1122
				z.de = 0x3344
			},
			expect: func(z *z80) bool {
				return z.hl == 0x1122 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x33
			},
		},
		// 0x73
		{

			name: "ld (hl),e",
			opc:  "ld",
			dst:  "(hl)",
			src:  "e",
			data: []byte{0x73},
			init: func(z *z80) {
				z.hl = 0x1122
				z.de = 0x3344
			},
			expect: func(z *z80) bool {
				return z.hl == 0x1122 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x44
			},
		},
		// 0x74
		{

			name: "ld (hl),h",
			opc:  "ld",
			dst:  "(hl)",
			src:  "h",
			data: []byte{0x74},
			init: func(z *z80) {
				z.hl = 0x1122
			},
			expect: func(z *z80) bool {
				return z.hl == 0x1122 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x11
			},
		},
		// 0x75
		{

			name: "ld (hl),l",
			opc:  "ld",
			dst:  "(hl)",
			src:  "l",
			data: []byte{0x75},
			init: func(z *z80) {
				z.hl = 0x1122
			},
			expect: func(z *z80) bool {
				return z.hl == 0x1122 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x22
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
		// 0x77
		{

			name: "ld (hl),a",
			opc:  "ld",
			dst:  "(hl)",
			src:  "a",
			data: []byte{0x77},
			init: func(z *z80) {
				z.hl = 0x1122
				z.af = 0x3344
			},
			expect: func(z *z80) bool {
				return z.hl == 0x1122 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x33
			},
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
		// 0x79
		{
			name: "ld a,c",
			opc:  "ld",
			dst:  "a",
			src:  "c",
			data: []byte{0x79},
			init: func(z *z80) { z.af = 0x1100; z.bc = 0x2233 },
			expect: func(z *z80) bool {
				return 0x3300 == z.af && z.pc == 0x0001
			},
		},
		// 0x7a
		{
			name: "ld a,d",
			opc:  "ld",
			dst:  "a",
			src:  "d",
			data: []byte{0x7a},
			init: func(z *z80) { z.af = 0x1100; z.de = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2200 == z.af && z.pc == 0x0001
			},
		},
		// 0x7b
		{
			name: "ld a,e",
			opc:  "ld",
			dst:  "a",
			src:  "e",
			data: []byte{0x7b},
			init: func(z *z80) { z.af = 0x1100; z.de = 0x2233 },
			expect: func(z *z80) bool {
				return 0x3300 == z.af && z.pc == 0x0001
			},
		},
		// 0x7c
		{
			name: "ld a,h",
			opc:  "ld",
			dst:  "a",
			src:  "h",
			data: []byte{0x7c},
			init: func(z *z80) { z.af = 0x1100; z.hl = 0x2233 },
			expect: func(z *z80) bool {
				return 0x2200 == z.af && z.pc == 0x0001
			},
		},
		// 0x7d
		{
			name: "ld a,l",
			opc:  "ld",
			dst:  "a",
			src:  "l",
			data: []byte{0x7d},
			init: func(z *z80) { z.af = 0x1100; z.hl = 0x2233 },
			expect: func(z *z80) bool {
				return 0x3300 == z.af && z.pc == 0x0001
			},
		},
		// 0x7e
		{

			name: "ld a,(hl)",
			opc:  "ld",
			dst:  "a",
			src:  "(hl)",
			data: []byte{0x7e},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bus.Write(0x1122, 0xaa)
			},
			expect: func(z *z80) bool {
				return z.af == 0xaa00 && z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xaa
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
		// 0xa7
		{
			name: "and a",
			opc:  "and",
			dst:  "a",
			data: []byte{0xa7},
			init: func(z *z80) { z.af = 0xa500 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xa500 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "and 0",
			opc:  "and",
			dst:  "a",
			data: []byte{0xa7},
			init: func(z *z80) { z.af = 0x0000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "and $7f",
			opc:  "and",
			dst:  "a",
			data: []byte{0xa7, 0x7f},
			init: func(z *z80) { z.af = 0xaf00 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xaf00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xc1
		{
			name: "pop bc",
			opc:  "pop",
			dst:  "bc",
			src:  "",
			data: []byte{0xc1},
			init: func(z *z80) {
				z.bc = 0x1122
				z.sp = 0xaa55
				z.bus.Write(0xaa55, 0xff)
				z.bus.Write(0xaa56, 0xee)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa57 &&
					z.bc == 0xeeff
			},
		},
		// 0xc2
		{
			name: "jp nz,nn (Z set)",
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
			name: "jp nz,nn (Z clear)",
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
		// 0xc5
		{
			name: "push bc",
			opc:  "push",
			dst:  "bc",
			src:  "",
			data: []byte{0xc5},
			init: func(z *z80) { z.bc = 0x1122; z.sp = 0xaa55 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa53 &&
					z.bus.Read(0xaa53) == 0x22 &&
					z.bus.Read(0xaa54) == 0x11
			},
		},
		// 0xc8
		{
			name: "ret z (Z set)",
			opc:  "ret",
			dst:  "z",
			data: []byte{0xc8},
			init: func(z *z80) {
				z.af |= zero
				z.sp = 0xaa55
				z.bus.Write(0xaa55, 0xff)
				z.bus.Write(0xaa56, 0xee)
			},
			expect: func(z *z80) bool {
				return z.pc == 0xeeff && z.sp == 0xaa57
			},
			dontSkipPC: true,
		},
		{
			name: "ret z (Z clear)",
			opc:  "ret",
			dst:  "z",
			data: []byte{0xc8},
			init: func(z *z80) {
				z.sp = 0xaa55
				z.bus.Write(0xaa55, 0xff)
				z.bus.Write(0xaa56, 0xee)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa55
			},
			dontSkipPC: true,
		},
		// 0xc9
		{
			name: "ret",
			opc:  "ret",
			data: []byte{0xc9},
			init: func(z *z80) {
				z.sp = 0xaa55
				z.bus.Write(0xaa55, 0xff)
				z.bus.Write(0xaa56, 0xee)
			},
			expect: func(z *z80) bool {
				return z.pc == 0xeeff && z.sp == 0xaa57
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
		// 0xcd
		{
			name: "call nn",
			opc:  "call",
			dst:  "$1122",
			src:  "",
			data: []byte{0xcd, 0x22, 0x11},
			init: func(z *z80) { z.sp = 0x5566 },
			expect: func(z *z80) bool {
				return z.pc == 0x1122 && z.sp == 0x5564 &&
					z.bus.Read(0x5564) == 0x03 &&
					z.bus.Read(0x5565) == 0x00
			},
			dontSkipPC: true,
		},
		// 0xd1
		{
			name: "pop de",
			opc:  "pop",
			dst:  "de",
			src:  "",
			data: []byte{0xd1},
			init: func(z *z80) {
				z.de = 0x1122
				z.sp = 0xaa55
				z.bus.Write(0xaa55, 0xff)
				z.bus.Write(0xaa56, 0xee)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa57 &&
					z.de == 0xeeff
			},
		},
		// 0xd2
		{
			name: "jp nc,nn (C set)",
			opc:  "jp",
			dst:  "nc",
			src:  "$1122",
			data: []byte{0xd2, 0x22, 0x11},
			init: func(z *z80) { z.af = 0xff00 | carry },
			expect: func(z *z80) bool {
				return z.pc == 0x0003
			},
			dontSkipPC: true,
		},
		{
			name: "jp nc,nn (C clear)",
			opc:  "jp",
			dst:  "nc",
			src:  "$1122",
			data: []byte{0xd2, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x1122
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
		// 0xd5
		{
			name: "push de",
			opc:  "push",
			dst:  "de",
			src:  "",
			data: []byte{0xd5},
			init: func(z *z80) { z.de = 0x1122; z.sp = 0xaa55 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa53 &&
					z.bus.Read(0xaa53) == 0x22 &&
					z.bus.Read(0xaa54) == 0x11
			},
		},
		// 0xda
		{
			name: "jp c,nn (C set)",
			opc:  "jp",
			dst:  "c",
			src:  "$1122",
			data: []byte{0xda, 0x22, 0x11},
			init: func(z *z80) { z.af = 0xff00 | carry },
			expect: func(z *z80) bool {
				return z.pc == 0x1122
			},
			dontSkipPC: true,
		},
		{
			name: "jp c,nn (C clear)",
			opc:  "jp",
			dst:  "c",
			src:  "$1122",
			data: []byte{0xda, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x0003
			},
			dontSkipPC: true,
		},
		// 0xdb
		{
			name: "in a,(n)",
			opc:  "in",
			dst:  "a",
			src:  "($aa)",
			data: []byte{0xdb, 0xaa},
			init: func(z *z80) { z.af = 0xff00 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002
			},
		},
		// 0xdd 0x23
		{
			name: "inc ix",
			opc:  "inc",
			dst:  "ix",
			src:  "",
			data: []byte{0xdd, 0x23},
			init: func(z *z80) { z.ix = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.ix == 0x1123
			},
		},
		{
			name: "inc ix == -1",
			opc:  "inc",
			dst:  "ix",
			src:  "",
			data: []byte{0xdd, 0x23},
			init: func(z *z80) { z.ix = 0xffff },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.ix == 0x0
			},
		},
		{
			name: "inc ix == 0x7fff",
			opc:  "inc",
			dst:  "ix",
			src:  "",
			data: []byte{0xdd, 0x23},
			init: func(z *z80) { z.ix = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.ix == 0x8000
			},
		},
		// 0xdd 0x2b
		{
			name: "dec ix",
			opc:  "dec",
			dst:  "ix",
			src:  "",
			data: []byte{0xdd, 0x2b},
			init: func(z *z80) { z.ix = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.ix == 0x1121
			},
		},
		{
			name: "dec ix == 0",
			opc:  "dec",
			dst:  "ix",
			src:  "",
			data: []byte{0xdd, 0x2b},
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.ix == 0xffff
			},
		},
		{
			name: "dec ix == 0x8000",
			opc:  "dec",
			dst:  "ix",
			src:  "",
			data: []byte{0xdd, 0x2b},
			init: func(z *z80) { z.ix = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.ix == 0x7fff
			},
		},
		// 0xdd e1
		{
			name: "pop af",
			opc:  "pop",
			dst:  "ix",
			src:  "",
			data: []byte{0xdd, 0xe1},
			init: func(z *z80) {
				z.ix = 0x1122
				z.sp = 0xaa55
				z.bus.Write(0xaa55, 0xff)
				z.bus.Write(0xaa56, 0xee)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa57 &&
					z.ix == 0xeeff
			},
		},
		// 0xdd 0xe5
		{
			name: "push ix",
			opc:  "push",
			dst:  "ix",
			src:  "",
			data: []byte{0xdd, 0xe5},
			init: func(z *z80) { z.ix = 0x1122; z.sp = 0xaa55 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.sp == 0xaa53 &&
					z.bus.Read(0xaa53) == 0x22 &&
					z.bus.Read(0xaa54) == 0x11
			},
		},
		// 0xe1
		{
			name: "pop hl",
			opc:  "pop",
			dst:  "hl",
			src:  "",
			data: []byte{0xe1},
			init: func(z *z80) {
				z.hl = 0x1122
				z.sp = 0xaa55
				z.bus.Write(0xaa55, 0xff)
				z.bus.Write(0xaa56, 0xee)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa57 &&
					z.hl == 0xeeff
			},
		},
		// 0xe2
		{
			name: "jp po,nn (P set)",
			opc:  "jp",
			dst:  "po",
			src:  "$1122",
			data: []byte{0xe2, 0x22, 0x11},
			init: func(z *z80) { z.af = 0xff00 | parity },
			expect: func(z *z80) bool {
				return z.pc == 0x0003
			},
			dontSkipPC: true,
		},
		{
			name: "jp po,nn (P clear)",
			opc:  "jp",
			dst:  "po",
			src:  "$1122",
			data: []byte{0xe2, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x1122
			},
			dontSkipPC: true,
		},
		// 0xe5
		{
			name: "push hl",
			opc:  "push",
			dst:  "hl",
			src:  "",
			data: []byte{0xe5},
			init: func(z *z80) { z.hl = 0x1122; z.sp = 0xaa55 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa53 &&
					z.bus.Read(0xaa53) == 0x22 &&
					z.bus.Read(0xaa54) == 0x11
			},
		},
		// 0xe6
		{
			name: "and n",
			opc:  "and",
			dst:  "$f0",
			data: []byte{0xe6, 0xf0},
			init: func(z *z80) { z.af = 0xa500 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.af&0xff00 == 0xa000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "and 0",
			opc:  "and",
			dst:  "$00",
			data: []byte{0xe6, 0x00},
			init: func(z *z80) { z.af = 0xa500 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "and $7f",
			opc:  "and",
			dst:  "$7f",
			data: []byte{0xe6, 0x7f},
			init: func(z *z80) { z.af = 0xaf00 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.af&0xff00 == 0x2f00 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xe9
		{
			// jp (hl) DOES NOT dereference hl but according to
			// Zilog the mnemonic uses parenthesis
			name: "jp (hl)",
			opc:  "jp",
			dst:  "(hl)",
			src:  "",
			data: []byte{0xe9},
			init: func(z *z80) { z.hl = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x1122
			},
			dontSkipPC: true,
		},
		// 0xea
		{
			name: "jp pe,nn (P set)",
			opc:  "jp",
			dst:  "pe",
			src:  "$1122",
			data: []byte{0xea, 0x22, 0x11},
			init: func(z *z80) { z.af = 0xff00 | parity },
			expect: func(z *z80) bool {
				return z.pc == 0x1122
			},
			dontSkipPC: true,
		},
		{
			name: "jp pe,nn (Z clear)",
			opc:  "jp",
			dst:  "pe",
			src:  "$1122",
			data: []byte{0xea, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x0003
			},
			dontSkipPC: true,
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
		// 0xf1
		{
			name: "pop af",
			opc:  "pop",
			dst:  "af",
			src:  "",
			data: []byte{0xf1},
			init: func(z *z80) {
				z.af = 0x1122
				z.sp = 0xaa55
				z.bus.Write(0xaa55, 0xff)
				z.bus.Write(0xaa56, 0xee)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa57 &&
					z.af == 0xeeff
			},
		},
		// 0xf2
		{
			name: "jp p,nn (S set)",
			opc:  "jp",
			dst:  "p",
			src:  "$1122",
			data: []byte{0xf2, 0x22, 0x11},
			init: func(z *z80) { z.af = 0xff00 | sign },
			expect: func(z *z80) bool {
				return z.pc == 0x0003
			},
			dontSkipPC: true,
		},
		{
			name: "jp p,nn (S clear)",
			opc:  "jp",
			dst:  "p",
			src:  "$1122",
			data: []byte{0xf2, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x1122
			},
			dontSkipPC: true,
		},
		// 0xf5
		{
			name: "push af",
			opc:  "push",
			dst:  "af",
			src:  "",
			data: []byte{0xf5},
			init: func(z *z80) { z.af = 0x1122; z.sp = 0xaa55 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa53 &&
					z.bus.Read(0xaa53) == 0x22 &&
					z.bus.Read(0xaa54) == 0x11
			},
		},
		// 0xfa
		{
			name: "jp m,nn (S set)",
			opc:  "jp",
			dst:  "m",
			src:  "$1122",
			data: []byte{0xfa, 0x22, 0x11},
			init: func(z *z80) { z.af = 0xff00 | sign },
			expect: func(z *z80) bool {
				return z.pc == 0x1122
			},
			dontSkipPC: true,
		},
		{
			name: "jp m,nn (S clear)",
			opc:  "jp",
			dst:  "m",
			src:  "$1122",
			data: []byte{0xfa, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x0003
			},
			dontSkipPC: true,
		},
		// 0xfd 0x23
		{
			name: "inc iy",
			opc:  "inc",
			dst:  "iy",
			src:  "",
			data: []byte{0xfd, 0x23},
			init: func(z *z80) { z.iy = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.iy == 0x1123
			},
		},
		{
			name: "inc iy == -1",
			opc:  "inc",
			dst:  "iy",
			src:  "",
			data: []byte{0xfd, 0x23},
			init: func(z *z80) { z.iy = 0xffff },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.iy == 0x0
			},
		},
		{
			name: "inc iy == 0x7fff",
			opc:  "inc",
			dst:  "iy",
			src:  "",
			data: []byte{0xfd, 0x23},
			init: func(z *z80) { z.iy = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.iy == 0x8000
			},
		},
		// 0xfd 0x2b
		{
			name: "dec iy",
			opc:  "dec",
			dst:  "iy",
			src:  "",
			data: []byte{0xfd, 0x2b},
			init: func(z *z80) { z.iy = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.iy == 0x1121
			},
		},
		{
			name: "dec iy == 0",
			opc:  "dec",
			dst:  "iy",
			src:  "",
			data: []byte{0xfd, 0x2b},
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.iy == 0xffff
			},
		},
		{
			name: "dec iy == 0x8000",
			opc:  "dec",
			dst:  "iy",
			src:  "",
			data: []byte{0xfd, 0x2b},
			init: func(z *z80) { z.iy = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.iy == 0x7fff
			},
		},
		// 0xfd e1
		{
			name: "pop af",
			opc:  "pop",
			dst:  "iy",
			src:  "",
			data: []byte{0xfd, 0xe1},
			init: func(z *z80) {
				z.iy = 0x1122
				z.sp = 0xaa55
				z.bus.Write(0xaa55, 0xff)
				z.bus.Write(0xaa56, 0xee)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa57 &&
					z.iy == 0xeeff
			},
		},
		// 0xfd 0xe5
		{
			name: "push iy",
			opc:  "push",
			dst:  "iy",
			src:  "",
			data: []byte{0xfd, 0xe5},
			init: func(z *z80) { z.iy = 0x1122; z.sp = 0xaa55 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.sp == 0xaa53 &&
					z.bus.Read(0xaa53) == 0x22 &&
					z.bus.Read(0xaa54) == 0x11
			},
		},
	}

	seen := make(map[uint16]int)
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
		//t.Logf("%v %v %v %v", opc, dst, src, x)
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

		o := uint16(test.data[0])
		if opcodes[o].multiByte {
			o = o<<8 | uint16(test.data[1])
		}
		seen[o]++
	}
	t.Logf("opcodes seen: %v", len(seen))

	// Minimal test to verify there is a unit test implemented.
	for o := range opcodes {
		// XXX add 2 byte opcodes
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
