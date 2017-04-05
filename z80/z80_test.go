package z80

import (
	"fmt"
	"strings"
	"testing"

	"github.com/marcopeereboom/toyz80/bus"
)

func TestInstructions(t *testing.T) {
	tests := []struct {
		name       string
		mn         string
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
			mn:     "nop",
			data:   []byte{0x00},
			expect: func(z *z80) bool { return z.pc == 0x0001 },
		},
		// 0x01
		{
			name: "ld bc,nn",
			mn:   "ld",
			dst:  "bc",
			src:  "$55aa",
			data: []byte{0x01, 0xaa, 0x55},
			expect: func(z *z80) bool {
				return 0x55aa == z.bc && z.pc == 0x0003
			},
		},
		{
			name: "ld bc,nn",
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "inc",
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
			mn:   "inc",
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
			mn:   "inc",
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
			mn:   "inc",
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
			mn:   "inc",
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
			mn:   "inc",
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
		// 0x05
		{
			name: "dec b",
			mn:   "dec",
			dst:  "b",
			src:  "",
			data: []byte{0x05},
			init: func(z *z80) { z.af = 0x9988; z.bc = 0x80a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0x7fa5 &&
					z.af&0xff00 == 0x9900 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec b == 1",
			mn:   "dec",
			dst:  "b",
			src:  "",
			data: []byte{0x05},
			init: func(z *z80) { z.bc = 0x01a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0x00a5 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec b == 0",
			mn:   "dec",
			dst:  "b",
			src:  "",
			data: []byte{0x05},
			init: func(z *z80) { z.bc = 0x00a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0xffa5 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		// 0x07
		{
			name: "rlca",
			mn:   "rlca",
			data: []byte{0x07},
			init: func(z *z80) { z.af = 0xa5ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x4b00 &&
					z.af&sign == sign &&
					z.af&zero == zero &&
					z.af&parity == parity &&
					z.af&carry == carry &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		// 0x06
		{

			name: "ld b,n",
			mn:   "ld",
			dst:  "b",
			src:  "$55",
			data: []byte{0x06, 0x55},
			init: func(z *z80) { z.bc = 0x1122 },
			expect: func(z *z80) bool {
				return z.bc == 0x5522 && z.pc == 0x0002
			},
		},
		// 0x07
		{

			name: "rlca $a5",
			mn:   "rlca",
			data: []byte{0x07},
			init: func(z *z80) { z.af = 0xa5ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x4b00 &&
					z.af&sign == sign &&
					z.af&zero == zero &&
					z.af&parity == parity &&
					z.af&carry == carry &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{

			name: "rlca $80",
			mn:   "rlca",
			data: []byte{0x07},
			init: func(z *z80) { z.af = 0x80ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0100 &&
					z.af&sign == sign &&
					z.af&zero == zero &&
					z.af&parity == parity &&
					z.af&carry == carry &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{

			name: "rlca $ff",
			mn:   "rlca",
			data: []byte{0x07},
			init: func(z *z80) { z.af = 0xffff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == zero &&
					z.af&parity == parity &&
					z.af&carry == carry &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{

			name: "rlca $7f",
			mn:   "rlca",
			data: []byte{0x07},
			init: func(z *z80) { z.af = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0xfe00 &&
					z.af&sign == sign &&
					z.af&zero == zero &&
					z.af&parity == parity &&
					z.af&carry == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		// 0x08
		{

			name: "ex af,af'",
			mn:   "ex",
			dst:  "af",
			src:  "af'",
			data: []byte{0x08},
			init: func(z *z80) {
				z.af = 0x1122
				z.af_ = 0x3344
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af == 0x3344 &&
					z.af_ == 0x1122
			},
		},
		// 0x09
		{
			name: "add hl,bc",
			mn:   "add",
			dst:  "hl",
			src:  "bc",
			data: []byte{0x09},
			init: func(z *z80) { z.bc = 0x1000; z.hl = 0x1000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add hl,bc 0x7fff",
			mn:   "add",
			dst:  "hl",
			src:  "bc",
			data: []byte{0x09},
			init: func(z *z80) { z.bc = 0x7fff; z.hl = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xfffe &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0x0a
		{

			name: "ld a,(bc)",
			mn:   "ld",
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
			mn:   "dec",
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
			mn:   "dec",
			dst:  "bc",
			src:  "",
			data: []byte{0x0b},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0xffff
			},
		},
		{
			name: "dec bc == 0x8000",
			mn:   "dec",
			dst:  "bc",
			src:  "",
			data: []byte{0x0b},
			init: func(z *z80) { z.bc = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0x7fff
			},
		},
		// 0x0c
		{
			name: "inc c",
			mn:   "inc",
			dst:  "c",
			src:  "",
			data: []byte{0x0c},
			init: func(z *z80) { z.bc = 0xa511 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0xa512 &&
					z.af&sign == 0 && z.af&zero == 0 &&
					z.af&parity == 0 && z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc c == -1",
			mn:   "inc",
			dst:  "c",
			src:  "",
			data: []byte{0x0c},
			init: func(z *z80) { z.bc = 0xa5ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0xa500 &&
					z.af&sign == 0 && z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc c == 0x7f",
			mn:   "inc",
			dst:  "c",
			src:  "",
			data: []byte{0x0c},
			init: func(z *z80) { z.bc = 0xa57f },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0xa580 &&
					z.af&sign == sign && z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		// 0x0d
		{
			name: "dec c",
			mn:   "dec",
			dst:  "c",
			src:  "",
			data: []byte{0x0d},
			init: func(z *z80) { z.af = 0x9988; z.bc = 0xa580 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0xa57f &&
					z.af&0xff00 == 0x9900 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec c == 1",
			mn:   "dec",
			dst:  "c",
			src:  "",
			data: []byte{0x0d},
			init: func(z *z80) { z.bc = 0xa501 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0xa500 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec c == 0",
			mn:   "dec",
			dst:  "c",
			src:  "",
			data: []byte{0x0d},
			init: func(z *z80) { z.bc = 0xa500 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.bc == 0xa5ff &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		// 0x0e
		{

			name: "ld c,n",
			mn:   "ld",
			dst:  "c",
			src:  "$55",
			data: []byte{0x0e, 0x55},
			init: func(z *z80) { z.bc = 0x2233 },
			expect: func(z *z80) bool {
				return z.bc == 0x2255 && z.pc == 0x0002
			},
		},
		// 0x0f
		{

			name: "rrca $a5",
			mn:   "rrca",
			data: []byte{0x0f},
			init: func(z *z80) { z.af = 0xa5ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0xd200 &&
					z.af&sign == sign &&
					z.af&zero == zero &&
					z.af&parity == parity &&
					z.af&carry == carry &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{

			name: "rrca $80",
			mn:   "rrca",
			data: []byte{0x0f},
			init: func(z *z80) { z.af = 0x80ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x4000 &&
					z.af&sign == sign &&
					z.af&zero == zero &&
					z.af&parity == parity &&
					z.af&carry == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{

			name: "rrca $ff",
			mn:   "rrca",
			data: []byte{0x0f},
			init: func(z *z80) { z.af = 0xffff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == zero &&
					z.af&parity == parity &&
					z.af&carry == carry &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{

			name: "rrca $7f",
			mn:   "rrca",
			data: []byte{0x0f},
			init: func(z *z80) { z.af = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0xbf00 &&
					z.af&sign == sign &&
					z.af&zero == zero &&
					z.af&parity == parity &&
					z.af&carry == carry &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		// 0x11 ld de,nn
		{

			name: "ld de,nn",
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "inc",
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
			mn:   "inc",
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
			mn:   "inc",
			dst:  "de",
			src:  "",
			data: []byte{0x13},
			init: func(z *z80) { z.de = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x8000
			},
		},
		// 0x14
		{
			name: "inc d",
			mn:   "inc",
			dst:  "d",
			src:  "",
			data: []byte{0x14},
			init: func(z *z80) { z.de = 0x11a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x12a5 &&
					z.af&sign == 0 && z.af&zero == 0 &&
					z.af&parity == 0 && z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc d == -1",
			mn:   "inc",
			dst:  "d",
			src:  "",
			data: []byte{0x14},
			init: func(z *z80) { z.de = 0xffa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x00a5 &&
					z.af&sign == 0 && z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc d == 0x7f",
			mn:   "inc",
			dst:  "d",
			src:  "",
			data: []byte{0x14},
			init: func(z *z80) { z.de = 0x7fa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x80a5 &&
					z.af&sign == sign && z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		// 0x15
		{
			name: "dec d",
			mn:   "dec",
			dst:  "d",
			src:  "",
			data: []byte{0x15},
			init: func(z *z80) { z.af = 0x9988; z.de = 0x80a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x7fa5 &&
					z.af&0xff00 == 0x9900 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec d == 1",
			mn:   "dec",
			dst:  "d",
			src:  "",
			data: []byte{0x15},
			init: func(z *z80) { z.de = 0x01a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x00a5 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec d == 0",
			mn:   "dec",
			dst:  "d",
			src:  "",
			data: []byte{0x15},
			init: func(z *z80) { z.de = 0x00a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0xffa5 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		// 0x16
		{

			name: "ld d,n",
			mn:   "ld",
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
			mn:   "jr",
			dst:  "$0005",
			data: []byte{0x18, 0x03},
			expect: func(z *z80) bool {
				return z.pc == 0x0005
			},
			dontSkipPC: true,
		},
		{
			name: "jr negative",
			mn:   "jr",
			dst:  "$ffff",
			data: []byte{0x18, 0xfd},
			expect: func(z *z80) bool {
				return z.pc == 0xffff
			},
			dontSkipPC: true,
		},
		// 0x19
		{
			name: "add hl,de",
			mn:   "add",
			dst:  "hl",
			src:  "de",
			data: []byte{0x19},
			init: func(z *z80) { z.de = 0x1234; z.hl = 0x4321 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x5555 &&
					z.de == 0x1234 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add hl,de 0x7fff",
			mn:   "add",
			dst:  "hl",
			src:  "de",
			data: []byte{0x19},
			init: func(z *z80) { z.de = 0x7fff; z.hl = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xfffe &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0x1a
		{

			name: "ld a,(de)",
			mn:   "ld",
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
			mn:   "dec",
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
			mn:   "dec",
			dst:  "de",
			src:  "",
			data: []byte{0x1b},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0xffff
			},
		},
		{
			name: "dec de == 0x8000",
			mn:   "dec",
			dst:  "de",
			src:  "",
			data: []byte{0x1b},
			init: func(z *z80) { z.de = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0x7fff
			},
		},
		// 0x1c
		{
			name: "inc e",
			mn:   "inc",
			dst:  "e",
			src:  "",
			data: []byte{0x1c},
			init: func(z *z80) { z.de = 0xa511 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0xa512 &&
					z.af&sign == 0 && z.af&zero == 0 &&
					z.af&parity == 0 && z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc e == -1",
			mn:   "inc",
			dst:  "e",
			src:  "",
			data: []byte{0x1c},
			init: func(z *z80) { z.de = 0xa5ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0xa500 &&
					z.af&sign == 0 && z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc e == 0x7f",
			mn:   "inc",
			dst:  "e",
			src:  "",
			data: []byte{0x1c},
			init: func(z *z80) { z.de = 0xa57f },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0xa580 &&
					z.af&sign == sign && z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		// 0x1d
		{
			name: "dec e",
			mn:   "dec",
			dst:  "e",
			src:  "",
			data: []byte{0x1d},
			init: func(z *z80) { z.af = 0x9988; z.de = 0xa580 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0xa57f &&
					z.af&0xff00 == 0x9900 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec e == 1",
			mn:   "dec",
			dst:  "e",
			src:  "",
			data: []byte{0x1d},
			init: func(z *z80) { z.de = 0xa501 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0xa500 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec e == 0",
			mn:   "dec",
			dst:  "e",
			src:  "",
			data: []byte{0x1d},
			init: func(z *z80) { z.de = 0xa500 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.de == 0xa5ff &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		// 0x1e
		{

			name: "ld e,n",
			mn:   "ld",
			dst:  "e",
			src:  "$55",
			data: []byte{0x1e, 0x55},
			init: func(z *z80) { z.de = 0x2233 },
			expect: func(z *z80) bool {
				return z.de == 0x2255 && z.pc == 0x0002
			},
		},
		// 0x1f
		{
			name: "rra 01",
			mn:   "rra",
			data: []byte{0x1f},
			init: func(z *z80) { z.af = 0x0100 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&carry == carry
			},
		},
		{
			name: "rra 01 + Carry",
			mn:   "rra",
			data: []byte{0x1f},
			init: func(z *z80) { z.af = 0x0100 | carry },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&carry == carry
			},
		},
		// 0x21 ld hl,nn
		{

			name: "ld hl,nn",
			mn:   "ld",
			dst:  "hl",
			src:  "$55aa",
			data: []byte{0x21, 0xaa, 0x55},
			expect: func(z *z80) bool {
				return 0x55aa == z.hl && z.pc == 0x0003
			},
		},
		// 0x22
		{
			name: "ld (nn),hl",
			mn:   "ld",
			dst:  "($b229)",
			src:  "hl",
			data: []byte{0x22, 0x29, 0xb2},
			init: func(z *z80) {
				z.hl = 0x483a
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0003 && z.hl == 0x483a &&
					z.bus.Read(0xb229) == 0x3a &&
					z.bus.Read(0xb22a) == 0x48
			},
		},
		// 0x23
		{
			name: "inc hl",
			mn:   "inc",
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
			mn:   "inc",
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
			mn:   "inc",
			dst:  "hl",
			src:  "",
			data: []byte{0x23},
			init: func(z *z80) { z.hl = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x8000
			},
		},
		// 0x24
		{
			name: "inc h",
			mn:   "inc",
			dst:  "h",
			src:  "",
			data: []byte{0x24},
			init: func(z *z80) { z.hl = 0x11a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x12a5 &&
					z.af&sign == 0 && z.af&zero == 0 &&
					z.af&parity == 0 && z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc h == -1",
			mn:   "inc",
			dst:  "h",
			src:  "",
			data: []byte{0x24},
			init: func(z *z80) { z.hl = 0xffa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x00a5 &&
					z.af&sign == 0 && z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc h == 0x7f",
			mn:   "inc",
			dst:  "h",
			src:  "",
			data: []byte{0x24},
			init: func(z *z80) { z.hl = 0x7fa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x80a5 &&
					z.af&sign == sign && z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		// 0x25
		{
			name: "dec h",
			mn:   "dec",
			dst:  "h",
			src:  "",
			data: []byte{0x25},
			init: func(z *z80) { z.af = 0x9988; z.hl = 0x80a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x7fa5 &&
					z.af&0xff00 == 0x9900 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec h == 1",
			mn:   "dec",
			dst:  "h",
			src:  "",
			data: []byte{0x25},
			init: func(z *z80) { z.hl = 0x01a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x00a5 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec h == 0",
			mn:   "dec",
			dst:  "h",
			src:  "",
			data: []byte{0x25},
			init: func(z *z80) { z.hl = 0x00a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xffa5 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		// 0x26
		{

			name: "ld h,n",
			mn:   "ld",
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
			mn:   "jr",
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
			mn:   "jr",
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
			mn:   "jr",
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
			mn:   "jr",
			dst:  "z",
			src:  "$0005",
			data: []byte{0x28, 0x3},
			expect: func(z *z80) bool {
				return z.pc == 0x0002
			},
			dontSkipPC: true,
		},
		// 0x29
		{
			name: "add hl,hl",
			mn:   "add",
			dst:  "hl",
			src:  "hl",
			data: []byte{0x29},
			init: func(z *z80) { z.af = 0xffff; z.hl = 0x1000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x2000 &&
					z.af&sign == sign &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add hl,hl 0x7fff",
			mn:   "add",
			dst:  "hl",
			src:  "hl",
			data: []byte{0x29},
			init: func(z *z80) { z.af = 0xffff; z.hl = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xfffe &&
					z.af&sign == sign &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0x2a
		{
			name: "ld hl,(nn)",
			mn:   "ld",
			dst:  "hl",
			src:  "($4545)",
			data: []byte{0x2a, 0x45, 0x45},
			init: func(z *z80) {
				z.hl = 0xa55a
				z.bus.Write(0x4545, 0x37)
				z.bus.Write(0x4546, 0xa1)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0003 && z.hl == 0xa137
			},
		},
		// 0x2b
		{
			name: "dec hl",
			mn:   "dec",
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
			mn:   "dec",
			dst:  "hl",
			src:  "",
			data: []byte{0x2b},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xffff
			},
		},
		{
			name: "dec hl == 0x8000",
			mn:   "dec",
			dst:  "hl",
			src:  "",
			data: []byte{0x2b},
			init: func(z *z80) { z.hl = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x7fff
			},
		},
		// 0x2c
		{
			name: "inc l",
			mn:   "inc",
			dst:  "l",
			src:  "",
			data: []byte{0x2c},
			init: func(z *z80) { z.hl = 0xa511 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xa512 &&
					z.af&sign == 0 && z.af&zero == 0 &&
					z.af&parity == 0 && z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc l == -1",
			mn:   "inc",
			dst:  "l",
			src:  "",
			data: []byte{0x2c},
			init: func(z *z80) { z.hl = 0xa5ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xa500 &&
					z.af&sign == 0 && z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc l == 0x7f",
			mn:   "inc",
			dst:  "l",
			src:  "",
			data: []byte{0x2c},
			init: func(z *z80) { z.hl = 0xa57f },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xa580 &&
					z.af&sign == sign && z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		// 0x2d
		{
			name: "dec l",
			mn:   "dec",
			dst:  "l",
			src:  "",
			data: []byte{0x2d},
			init: func(z *z80) { z.af = 0x9988; z.hl = 0xa580 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xa57f &&
					z.af&0xff00 == 0x9900 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec l == 1",
			mn:   "dec",
			dst:  "l",
			src:  "",
			data: []byte{0x2d},
			init: func(z *z80) { z.hl = 0xa501 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xa500 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec l == 0",
			mn:   "dec",
			dst:  "l",
			src:  "",
			data: []byte{0x2d},
			init: func(z *z80) { z.hl = 0xa500 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xa5ff &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		// 0x2e
		{

			name: "ld l,n",
			mn:   "ld",
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
			mn:   "cpl",
			data: []byte{0x2f},
			init: func(z *z80) { z.af = 0xa5ff &^ (addsub | halfCarry) },
			expect: func(z *z80) bool {
				return z.af&0xffff == 0x5aff && z.pc == 0x0001
			},
		},
		// 0x31
		{

			name: "ld sp,nn",
			mn:   "ld",
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
			mn:   "ld",
			dst:  "($ffee)",
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
			mn:   "inc",
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
			mn:   "inc",
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
			mn:   "inc",
			dst:  "sp",
			src:  "",
			data: []byte{0x33},
			init: func(z *z80) { z.sp = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0x8000
			},
		},
		// 0x34
		{

			name: "inc (hl)",
			mn:   "inc",
			dst:  "(hl)",
			data: []byte{0x34},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bus.Write(0x1122, 0x11)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x12 &&
					z.af&sign == 0 && z.af&zero == 0 &&
					z.af&parity == 0 && z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc (hl) == -1",
			mn:   "inc",
			dst:  "(hl)",
			data: []byte{0x34},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bus.Write(0x1122, 0xff)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x00 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc (hl) == 7f",
			mn:   "inc",
			dst:  "(hl)",
			data: []byte{0x34},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bus.Write(0x1122, 0x7f)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x80 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		// 0x35
		{

			name: "dec (hl)",
			mn:   "dec",
			dst:  "(hl)",
			data: []byte{0x35},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bus.Write(0x1122, 0x11)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x10 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec (hl) == 0",
			mn:   "dec",
			dst:  "(hl)",
			data: []byte{0x35},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bus.Write(0x1122, 0x00)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0xff &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec (hl) == 80",
			mn:   "dec",
			dst:  "(hl)",
			data: []byte{0x35},
			init: func(z *z80) {
				z.hl = 0x1122
				z.bus.Write(0x1122, 0x80)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.bus.Read(0x1122) == 0x7f &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		// 0x36
		{

			name: "ld (hl),n",
			mn:   "ld",
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
			mn:   "scf",
			data: []byte{0x37},
			init: func(z *z80) { z.af = 0xff00 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x39
		{
			name: "add hl,sp",
			mn:   "add",
			dst:  "hl",
			src:  "sp",
			data: []byte{0x39},
			init: func(z *z80) { z.sp = 0x1000; z.hl = 0x1000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add hl,sp 0x7fff",
			mn:   "add",
			dst:  "hl",
			src:  "sp",
			data: []byte{0x39},
			init: func(z *z80) { z.sp = 0x7fff; z.hl = 0x7fff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.hl == 0xfffe &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0x3b
		{
			name: "dec sp",
			mn:   "dec",
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
			mn:   "dec",
			dst:  "sp",
			src:  "",
			data: []byte{0x3b},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xffff
			},
		},
		{
			name: "dec sp == 0x8000",
			mn:   "dec",
			dst:  "sp",
			src:  "",
			data: []byte{0x3b},
			init: func(z *z80) { z.sp = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0x7fff
			},
		},
		// 0x3c
		{
			name: "inc a",
			mn:   "inc",
			dst:  "a",
			src:  "",
			data: []byte{0x3c},
			init: func(z *z80) { z.af = 0x11a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x1200 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc a == -1",
			mn:   "inc",
			dst:  "a",
			src:  "",
			data: []byte{0x3c},
			init: func(z *z80) { z.af = 0xffa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		{
			name: "inc a == 0x7f",
			mn:   "inc",
			dst:  "a",
			src:  "",
			data: []byte{0x3c},
			init: func(z *z80) { z.af = 0x7fa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0
			},
		},
		// 0x3d
		{
			name: "dec a",
			mn:   "dec",
			dst:  "a",
			data: []byte{0x3d},
			init: func(z *z80) { z.af = 0x80a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x7f00 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec a == 1",
			mn:   "dec",
			dst:  "a",
			data: []byte{0x3d},
			init: func(z *z80) { z.af = 0x01a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == addsub
			},
		},
		{
			name: "dec a == 0",
			mn:   "dec",
			dst:  "a",
			src:  "",
			data: []byte{0x3d},
			init: func(z *z80) { z.af = 0x00a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub
			},
		},
		// 0x3f
		{

			name: "ccf (0xff)",
			mn:   "ccf",
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
			mn:   "ccf",
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
			mn:   "ld",
			dst:  "a",
			src:  "($55aa)",
			data: []byte{0x3a, 0xaa, 0x55},
			init: func(z *z80) {
				z.af = 0x3344
				z.bus.Write(0x55aa, 0xff)
			},
			expect: func(z *z80) bool {
				return z.bus.Read(0x55aa) == 0xff &&
					z.af == 0xff44 && z.pc == 0x0003
			},
		},
		// 0x3e
		{

			name: "ld a,n",
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "halt",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
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
			mn:   "ld",
			dst:  "a",
			src:  "a",
			data: []byte{0x7f},
			init: func(z *z80) { z.af = 0x11a5 },
			expect: func(z *z80) bool {
				return 0x11a5 == z.af && z.pc == 0x0001
			},
		},
		// 0x80
		{
			name: "add a,b",
			mn:   "add",
			dst:  "a",
			src:  "b",
			data: []byte{0x80},
			init: func(z *z80) { z.af = 0x10ff; z.bc = 0x10a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,b == $7f",
			mn:   "add",
			dst:  "a",
			src:  "b",
			data: []byte{0x80},
			init: func(z *z80) { z.af = 0x0100; z.bc = 0x7fa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,b == $ff",
			mn:   "add",
			dst:  "a",
			src:  "b",
			data: []byte{0x80},
			init: func(z *z80) { z.af = 0x0100; z.bc = 0xffa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x81
		{
			name: "add a,c",
			mn:   "add",
			dst:  "a",
			src:  "c",
			data: []byte{0x81},
			init: func(z *z80) { z.af = 0x10ff; z.bc = 0xa510 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,c == $7f",
			mn:   "add",
			dst:  "a",
			src:  "c",
			data: []byte{0x81},
			init: func(z *z80) { z.af = 0x0100; z.bc = 0xa57f },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,c == $ff",
			mn:   "add",
			dst:  "a",
			src:  "c",
			data: []byte{0x81},
			init: func(z *z80) { z.af = 0x0100; z.bc = 0xa5ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x82
		{
			name: "add a,d",
			mn:   "add",
			dst:  "a",
			src:  "d",
			data: []byte{0x82},
			init: func(z *z80) { z.af = 0x10ff; z.de = 0x10a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,d == $7f",
			mn:   "add",
			dst:  "a",
			src:  "d",
			data: []byte{0x82},
			init: func(z *z80) { z.af = 0x0100; z.de = 0x7fa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,d == $ff",
			mn:   "add",
			dst:  "a",
			src:  "d",
			data: []byte{0x82},
			init: func(z *z80) { z.af = 0x0100; z.de = 0xffa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x83
		{
			name: "add a,e",
			mn:   "add",
			dst:  "a",
			src:  "e",
			data: []byte{0x83},
			init: func(z *z80) { z.af = 0x10ff; z.de = 0xa510 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,e == $7f",
			mn:   "add",
			dst:  "a",
			src:  "e",
			data: []byte{0x83},
			init: func(z *z80) { z.af = 0x0100; z.de = 0xa57f },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,e == $ff",
			mn:   "add",
			dst:  "a",
			src:  "e",
			data: []byte{0x83},
			init: func(z *z80) { z.af = 0x0100; z.de = 0xa5ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x84
		{
			name: "add a,h",
			mn:   "add",
			dst:  "a",
			src:  "h",
			data: []byte{0x84},
			init: func(z *z80) { z.af = 0x10ff; z.hl = 0x10a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,h == $7f",
			mn:   "add",
			dst:  "a",
			src:  "h",
			data: []byte{0x84},
			init: func(z *z80) { z.af = 0x0100; z.hl = 0x7fa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,h == $ff",
			mn:   "add",
			dst:  "a",
			src:  "h",
			data: []byte{0x84},
			init: func(z *z80) { z.af = 0x0100; z.hl = 0xffa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x85
		{
			name: "add a,l",
			mn:   "add",
			dst:  "a",
			src:  "l",
			data: []byte{0x85},
			init: func(z *z80) { z.af = 0x10ff; z.hl = 0xa510 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,l == $7f",
			mn:   "add",
			dst:  "a",
			src:  "l",
			data: []byte{0x85},
			init: func(z *z80) { z.af = 0x0100; z.hl = 0xa57f },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,l == $ff",
			mn:   "add",
			dst:  "a",
			src:  "l",
			data: []byte{0x85},
			init: func(z *z80) { z.af = 0x0100; z.hl = 0xa5ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x86
		{
			name: "add a,(hl)",
			mn:   "add",
			dst:  "a",
			src:  "(hl)",
			data: []byte{0x86},
			init: func(z *z80) {
				z.af = 0x10ff
				z.hl = 0x1122
				z.bus.Write(0x1122, 0x10)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,i == $7f",
			mn:   "add",
			dst:  "a",
			src:  "(hl)",
			data: []byte{0x86},
			init: func(z *z80) {
				z.af = 0x01ff
				z.hl = 0x1122
				z.bus.Write(0x1122, 0x7f)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,(hl) == $ff",
			mn:   "add",
			dst:  "a",
			src:  "(hl)",
			data: []byte{0x86},
			init: func(z *z80) {
				z.af = 0x01ff
				z.hl = 0x1122
				z.bus.Write(0x1122, 0xff)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x87
		{
			name: "add a,a",
			mn:   "add",
			dst:  "a",
			src:  "a",
			data: []byte{0x87},
			init: func(z *z80) { z.af = 0x10ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,a == $7f",
			mn:   "add",
			dst:  "a",
			src:  "a",
			data: []byte{0x87},
			init: func(z *z80) { z.af = 0x7f00 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0xfe00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,a == $40",
			mn:   "add",
			dst:  "a",
			src:  "a",
			data: []byte{0x87},
			init: func(z *z80) { z.af = 0x4000; z.hl = 0xffa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0x88
		{
			name: "adc a,b",
			mn:   "adc",
			dst:  "a",
			src:  "b",
			data: []byte{0x88},
			init: func(z *z80) { z.af = 0x10ff; z.bc = 0x10a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x2100 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "adc a,b == $7f",
			mn:   "adc",
			dst:  "a",
			src:  "b",
			data: []byte{0x88},
			init: func(z *z80) { z.af = 0x0100; z.bc = 0x7fa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "adc a,b == $ff",
			mn:   "adc",
			dst:  "a",
			src:  "b",
			data: []byte{0x88},
			init: func(z *z80) { z.af = 0x0100; z.bc = 0xffa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x89
		{
			name: "adc a,c",
			mn:   "adc",
			dst:  "a",
			src:  "c",
			data: []byte{0x89},
			init: func(z *z80) { z.af = 0x10ff; z.bc = 0xa510 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x2100 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "adc a,c == $7f",
			mn:   "adc",
			dst:  "a",
			src:  "c",
			data: []byte{0x89},
			init: func(z *z80) { z.af = 0x0100; z.bc = 0xa57f },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "adc a,c == $ff",
			mn:   "adc",
			dst:  "a",
			src:  "c",
			data: []byte{0x89},
			init: func(z *z80) { z.af = 0x0100; z.bc = 0xa5ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0x8a
		{
			name: "adc a,d",
			mn:   "adc",
			dst:  "a",
			src:  "d",
			data: []byte{0x8a},
			init: func(z *z80) { z.af = 0x10ff; z.de = 0x10a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x2100 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "adc a,d == $7f",
			mn:   "adc",
			dst:  "a",
			src:  "d",
			data: []byte{0x8a},
			init: func(z *z80) { z.af = 0x0100; z.de = 0x7fa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "adc a,d == $ff",
			mn:   "adc",
			dst:  "a",
			src:  "d",
			data: []byte{0x8a},
			init: func(z *z80) { z.af = 0x0100; z.de = 0xffa5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0xa7
		{
			name: "and a",
			mn:   "and",
			dst:  "a",
			src:  "a",
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
			name: "and a 0",
			mn:   "and",
			dst:  "a",
			src:  "a",
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
			name: "and a $7f",
			mn:   "and",
			dst:  "a",
			src:  "a",
			data: []byte{0xa7},
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
		// 0xb0
		{
			name: "or b",
			mn:   "or",
			dst:  "b",
			data: []byte{0xb0},
			init: func(z *z80) { z.af = 0xa500; z.bc = 0xf0a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xf500 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "or b == 0",
			mn:   "or",
			dst:  "b",
			data: []byte{0x0b0},
			init: func(z *z80) { z.af = 0xa500; z.bc = 0x00a5 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xa500 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "or b 0 on 0",
			mn:   "or",
			dst:  "b",
			data: []byte{0xb0},
			init: func(z *z80) { z.af = 0x0000; z.bc = 0x005a },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "or b $7f",
			mn:   "or",
			dst:  "b",
			data: []byte{0xb0},
			init: func(z *z80) { z.af = 0x8500; z.bc = 0x7f5a },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xb1
		{
			name: "or c",
			mn:   "or",
			dst:  "c",
			data: []byte{0xb1},
			init: func(z *z80) { z.af = 0xa500; z.bc = 0xa5f0 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xf500 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "or c == 0",
			mn:   "or",
			dst:  "c",
			data: []byte{0x0b1},
			init: func(z *z80) { z.af = 0xa500; z.bc = 0xa500 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xa500 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "or c 0 on 0",
			mn:   "or",
			dst:  "c",
			data: []byte{0xb1},
			init: func(z *z80) { z.af = 0x0000; z.bc = 0x5a00 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "or c $7f",
			mn:   "or",
			dst:  "c",
			data: []byte{0xb1},
			init: func(z *z80) { z.af = 0x8500; z.bc = 0x5a7f },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xb2
		{
			name: "or d",
			mn:   "or",
			dst:  "d",
			data: []byte{0xb2},
			init: func(z *z80) { z.af = 0xa5ff; z.de = 0x5aff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xb3
		{
			name: "or e",
			mn:   "or",
			dst:  "e",
			data: []byte{0xb3},
			init: func(z *z80) { z.af = 0xa5ff; z.de = 0xff5a },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xb4
		{
			name: "or h",
			mn:   "or",
			dst:  "h",
			data: []byte{0xb4},
			init: func(z *z80) { z.af = 0xa5ff; z.hl = 0x5aff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xb5
		{
			name: "or l",
			mn:   "or",
			dst:  "l",
			data: []byte{0xb5},
			init: func(z *z80) { z.af = 0xa5ff; z.hl = 0xff5a },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xb6
		{
			name: "or (hl)",
			mn:   "or",
			dst:  "(hl)",
			data: []byte{0xb6},
			init: func(z *z80) {
				z.af = 0xa5ff
				z.hl = 0x1122
				z.bus.Write(0x1122, 0x5a)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xb7
		{
			name: "or a",
			mn:   "or",
			dst:  "a",
			data: []byte{0xb7},
			init: func(z *z80) { z.af = 0xa5ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xa500 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xb8
		{
			name: "cp b <",
			mn:   "cp",
			dst:  "b",
			data: []byte{0xb8},
			init: func(z *z80) { z.af = 0x1000; z.bc = 0x2000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x1000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		{
			name: "cp b >",
			mn:   "cp",
			dst:  "b",
			data: []byte{0xb8},
			init: func(z *z80) { z.af = 0x2000; z.bc = 0x1000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		{
			name: "cp b =",
			mn:   "cp",
			dst:  "b",
			data: []byte{0xb8},
			init: func(z *z80) { z.af = 0x2000; z.bc = 0x2000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		// 0xb9
		{
			name: "cp c <",
			mn:   "cp",
			dst:  "c",
			data: []byte{0xb9},
			init: func(z *z80) { z.af = 0x1000; z.bc = 0x0020 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x1000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		{
			name: "cp c >",
			mn:   "cp",
			dst:  "c",
			data: []byte{0xb9},
			init: func(z *z80) { z.af = 0x2000; z.bc = 0x0010 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		{
			name: "cp c =",
			mn:   "cp",
			dst:  "c",
			data: []byte{0xb9},
			init: func(z *z80) { z.af = 0x2000; z.bc = 0x0020 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		// 0xba
		{
			name: "cp d <",
			mn:   "cp",
			dst:  "d",
			data: []byte{0xba},
			init: func(z *z80) { z.af = 0x1000; z.de = 0x2000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x1000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		{
			name: "cp d >",
			mn:   "cp",
			dst:  "d",
			data: []byte{0xba},
			init: func(z *z80) { z.af = 0x2000; z.de = 0x1000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		{
			name: "cp d =",
			mn:   "cp",
			dst:  "d",
			data: []byte{0xba},
			init: func(z *z80) { z.af = 0x2000; z.de = 0x2000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		// 0xbb
		{
			name: "cp e <",
			mn:   "cp",
			dst:  "e",
			data: []byte{0xbb},
			init: func(z *z80) { z.af = 0x1000; z.de = 0x0020 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x1000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		{
			name: "cp e >",
			mn:   "cp",
			dst:  "e",
			data: []byte{0xbb},
			init: func(z *z80) { z.af = 0x2000; z.de = 0x0010 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		{
			name: "cp e =",
			mn:   "cp",
			dst:  "e",
			data: []byte{0xbb},
			init: func(z *z80) { z.af = 0x2000; z.de = 0x0020 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		// 0xbc
		{
			name: "cp h <",
			mn:   "cp",
			dst:  "h",
			data: []byte{0xbc},
			init: func(z *z80) { z.af = 0x1000; z.hl = 0x2000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x1000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		{
			name: "cp h >",
			mn:   "cp",
			dst:  "h",
			data: []byte{0xbc},
			init: func(z *z80) { z.af = 0x2000; z.hl = 0x1000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		{
			name: "cp h =",
			mn:   "cp",
			dst:  "h",
			data: []byte{0xbc},
			init: func(z *z80) { z.af = 0x2000; z.hl = 0x2000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		// 0xbd
		{
			name: "cp l <",
			mn:   "cp",
			dst:  "l",
			data: []byte{0xbd},
			init: func(z *z80) { z.af = 0x1000; z.hl = 0x0020 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x1000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		{
			name: "cp l >",
			mn:   "cp",
			dst:  "l",
			data: []byte{0xbd},
			init: func(z *z80) { z.af = 0x2000; z.hl = 0x10 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		{
			name: "cp l =",
			mn:   "cp",
			dst:  "l",
			data: []byte{0xbd},
			init: func(z *z80) { z.af = 0x2000; z.hl = 0x0020 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		// 0xbe
		{
			name: "cp (hl) <",
			mn:   "cp",
			dst:  "(hl)",
			data: []byte{0xbe},
			init: func(z *z80) {
				z.af = 0x1000
				z.hl = 0x1122
				z.bus.Write(0x1122, 0x20)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x1000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		{
			name: "cp (hl) >",
			mn:   "cp",
			dst:  "(hl)",
			data: []byte{0xbe},
			init: func(z *z80) {
				z.af = 0x2000
				z.hl = 0x1122
				z.bus.Write(0x1122, 0x10)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		{
			name: "cp (hl) =",
			mn:   "cp",
			dst:  "(hl)",
			data: []byte{0xbe},
			init: func(z *z80) {
				z.af = 0x2000
				z.hl = 0x1122
				z.bus.Write(0x1122, 0x20)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		// 0xbf
		{
			name: "cp a",
			mn:   "cp",
			dst:  "a",
			data: []byte{0xbf},
			init: func(z *z80) { z.af = 0xaf00 },
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.af&0xff00 == 0xaf00 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		// 0xc0
		{
			name: "ret nz (Z clear)",
			mn:   "ret",
			dst:  "nz",
			data: []byte{0xc0},
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
		{
			name: "ret nz (z set)",
			mn:   "ret",
			dst:  "nz",
			data: []byte{0xc0},
			init: func(z *z80) {
				z.af |= zero
				z.sp = 0xaa55
				z.bus.Write(0xaa55, 0xff)
				z.bus.Write(0xaa56, 0xee)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa55
			},
			dontSkipPC: true,
		},
		// 0xc1
		{
			name: "pop bc",
			mn:   "pop",
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
			mn:   "jp",
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
			mn:   "jp",
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
			mn:   "jp",
			dst:  "$1122",
			src:  "",
			data: []byte{0xc3, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x1122
			},
			dontSkipPC: true,
		},
		// 0xc4
		{
			name: "call nz,nn (Z set)",
			mn:   "call",
			dst:  "nz",
			src:  "$1122",
			data: []byte{0xc4, 0x22, 0x11},
			init: func(z *z80) { z.af = zero; z.sp = 0x5566 },
			expect: func(z *z80) bool {
				return z.pc == 0x0003 && z.sp == 0x5566
			},
			dontSkipPC: true,
		},
		{
			name: "call nz,nn (Z clear)",
			mn:   "call",
			dst:  "nz",
			src:  "$1122",
			data: []byte{0xc4, 0x22, 0x11},
			init: func(z *z80) { z.sp = 0x5566 },
			expect: func(z *z80) bool {
				return z.pc == 0x1122 && z.sp == 0x5564 &&
					z.bus.Read(0x5564) == 0x03 &&
					z.bus.Read(0x5565) == 0x00
			},
			dontSkipPC: true,
		},
		// 0xc5
		{
			name: "push bc",
			mn:   "push",
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
		// 0xc6
		{
			name: "add a,i",
			mn:   "add",
			dst:  "a",
			src:  "$10",
			data: []byte{0xc6, 0x10},
			init: func(z *z80) { z.af = 0x10ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,i == $7f",
			mn:   "add",
			dst:  "a",
			src:  "$7f",
			data: []byte{0xc6, 0x7f},
			init: func(z *z80) { z.af = 0x0100 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add a,i == $ff",
			mn:   "add",
			dst:  "a",
			src:  "$ff",
			data: []byte{0xc6, 0xff},
			init: func(z *z80) { z.af = 0x0100 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0xc7
		{
			name: "rst $00",
			mn:   "rst",
			dst:  "$00",
			data: []byte{0xc7},
			init: func(z *z80) {
				z.sp = 0x5566
				z.bus.Write(0x5564, 0x64)
				z.bus.Write(0x5565, 0x65)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0000 && z.sp == 0x5564 &&
					z.bus.Read(0x5564) == 0x01 &&
					z.bus.Read(0x5565) == 0x00
			},
			dontSkipPC: true,
		},
		// 0xc8
		{
			name: "ret z (Z set)",
			mn:   "ret",
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
			mn:   "ret",
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
			mn:   "ret",
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
			mn:   "jp",
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
			mn:   "jp",
			dst:  "z",
			src:  "$1122",
			data: []byte{0xca, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x0003
			},
			dontSkipPC: true,
		},
		// 0xcb 0x27
		{
			name: "sla a",
			mn:   "sla",
			dst:  "a",
			data: []byte{0xcb, 0x27},
			init: func(z *z80) { z.af = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x2200 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "sla a = 0x80",
			mn:   "sla",
			dst:  "a",
			data: []byte{0xcb, 0x27},
			init: func(z *z80) { z.af = 0x8022 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		{
			name: "sla a = 0xff",
			mn:   "sla",
			dst:  "a",
			data: []byte{0xcb, 0x27},
			init: func(z *z80) { z.af = 0xff22 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0xfe00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0xcb 0x3f
		{
			name: "srl a",
			mn:   "srl",
			dst:  "a",
			data: []byte{0xcb, 0x3f},
			init: func(z *z80) { z.af = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x0800 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		{
			name: "srl a = 0x80",
			mn:   "srl",
			dst:  "a",
			data: []byte{0xcb, 0x3f},
			init: func(z *z80) { z.af = 0x8022 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x4000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "srl a = 0xff",
			mn:   "srl",
			dst:  "a",
			data: []byte{0xcb, 0x3f},
			init: func(z *z80) { z.af = 0xff22 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x7f00 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0xcd
		{
			name: "call nn",
			mn:   "call",
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
		// 0xce
		{
			name: "adc a,i",
			mn:   "adc",
			dst:  "a",
			src:  "$10",
			data: []byte{0xce, 0x10},
			init: func(z *z80) { z.af = 0x10ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x2100 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "adc a,i == $7f",
			mn:   "adc",
			dst:  "a",
			src:  "$7f",
			data: []byte{0xce, 0x7f},
			init: func(z *z80) { z.af = 0x0100 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x8000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == parity &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "adc a,i == $ff",
			mn:   "adc",
			dst:  "a",
			src:  "$ff",
			data: []byte{0xce, 0xff},
			init: func(z *z80) { z.af = 0x0000 | carry },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0xcf
		{
			name: "rst $08",
			mn:   "rst",
			dst:  "$08",
			data: []byte{0xcf},
			init: func(z *z80) {
				z.sp = 0x5566
				z.bus.Write(0x5564, 0x64)
				z.bus.Write(0x5565, 0x65)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0008 && z.sp == 0x5564 &&
					z.bus.Read(0x5564) == 0x01 &&
					z.bus.Read(0x5565) == 0x00
			},
			dontSkipPC: true,
		},
		// 0xd0
		{
			name: "ret nc (C clear)",
			mn:   "ret",
			dst:  "nc",
			data: []byte{0xd0},
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
		{
			name: "ret nc (C set)",
			mn:   "ret",
			dst:  "nc",
			data: []byte{0xd0},
			init: func(z *z80) {
				z.af |= carry
				z.sp = 0xaa55
				z.bus.Write(0xaa55, 0xff)
				z.bus.Write(0xaa56, 0xee)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.sp == 0xaa55
			},
			dontSkipPC: true,
		},
		// 0xd1
		{
			name: "pop de",
			mn:   "pop",
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
			mn:   "jp",
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
			mn:   "jp",
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
			mn:   "out",
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
			mn:   "push",
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
		// 0xd6
		{
			name: "sub i",
			mn:   "sub",
			dst:  "$10",
			data: []byte{0xd6, 0x10},
			init: func(z *z80) { z.af = 0x10ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					z.af&parity == 0 &&
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		// 0xd7
		{
			name: "rst $10",
			mn:   "rst",
			dst:  "$10",
			data: []byte{0xd7},
			init: func(z *z80) {
				z.sp = 0x5566
				z.bus.Write(0x5564, 0x64)
				z.bus.Write(0x5565, 0x65)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0010 && z.sp == 0x5564 &&
					z.bus.Read(0x5564) == 0x01 &&
					z.bus.Read(0x5565) == 0x00
			},
			dontSkipPC: true,
		},
		// 0xd8
		{
			name: "ret c (C set)",
			mn:   "ret",
			dst:  "c",
			data: []byte{0xd8},
			init: func(z *z80) {
				z.af |= carry
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
			name: "ret c (C clear)",
			mn:   "ret",
			dst:  "c",
			data: []byte{0xd8},
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
		// 0xda
		{
			name: "jp c,nn (C set)",
			mn:   "jp",
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
			mn:   "jp",
			dst:  "c",
			src:  "$1122",
			data: []byte{0xda, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x0003
			},
			dontSkipPC: true,
		},
		// 0xdc
		{
			name: "call c,nn (C set)",
			mn:   "call",
			dst:  "c",
			src:  "$1122",
			data: []byte{0xdc, 0x22, 0x11},
			init: func(z *z80) { z.af = carry; z.sp = 0x5566 },
			expect: func(z *z80) bool {
				return z.pc == 0x1122 && z.sp == 0x5564 &&
					z.bus.Read(0x5564) == 0x03 &&
					z.bus.Read(0x5565) == 0x00
			},
			dontSkipPC: true,
		},
		{
			name: "call c,nn (C clear)",
			mn:   "call",
			dst:  "c",
			src:  "$1122",
			data: []byte{0xdc, 0x22, 0x11},
			init: func(z *z80) { z.sp = 0x5566 },
			expect: func(z *z80) bool {
				return z.pc == 0x0003 && z.sp == 0x5566
			},
			dontSkipPC: true,
		},
		// 0xdb
		{
			name: "in a,(n)",
			mn:   "in",
			dst:  "a",
			src:  "($aa)",
			data: []byte{0xdb, 0xaa},
			init: func(z *z80) { z.af = 0xff00 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002
			},
		},
		// 0xdd 0x09 add ix,bc
		{
			name: "add ix,bc",
			mn:   "add",
			dst:  "ix",
			src:  "bc",
			data: []byte{0xdd, 0x09},
			init: func(z *z80) { z.ix = 0x3344; z.bc = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.ix == 0x4466 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add ix,bc (-1+1)",
			mn:   "add",
			dst:  "ix",
			src:  "bc",
			data: []byte{0xdd, 0x09},
			init: func(z *z80) { z.ix = 0xffff; z.bc = 0x0001 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.ix == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		{
			name: "add ix,bc (-1+2)",
			mn:   "add",
			dst:  "ix",
			src:  "bc",
			data: []byte{0xdd, 0x09},
			init: func(z *z80) {
				z.ix = 0xffff
				z.bc = 0x0002
				z.af |= carry
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.ix == 0x0001 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0xdd 0x23
		{
			name: "inc ix",
			mn:   "inc",
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
			mn:   "inc",
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
			mn:   "inc",
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
			mn:   "dec",
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
			mn:   "dec",
			dst:  "ix",
			src:  "",
			data: []byte{0xdd, 0x2b},
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.ix == 0xffff
			},
		},
		{
			name: "dec ix == 0x8000",
			mn:   "dec",
			dst:  "ix",
			src:  "",
			data: []byte{0xdd, 0x2b},
			init: func(z *z80) { z.ix = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.ix == 0x7fff
			},
		},
		// 0xdd 0x86
		{
			name: "add a,(ix+d)",
			mn:   "add",
			dst:  "a",
			src:  "(ix+$11)",
			data: []byte{0xdd, 0x86, 0x11},
			init: func(z *z80) {
				z.af = 0x1100
				z.ix = 0x3344
				z.bus.Write(0x3344+0x11, 0x55)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0003 &&
					z.af&0xff00 == 0x6600 &&
					z.ix == 0x3344 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xdd 0xe1
		{
			name: "pop ix",
			mn:   "pop",
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
				return z.pc == 0x0002 && z.sp == 0xaa57 &&
					z.ix == 0xeeff
			},
		},
		// 0xdd 0xe5
		{
			name: "push ix",
			mn:   "push",
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
		// 0xde
		{
			name: "sbc a,i",
			mn:   "sbc",
			dst:  "a",
			src:  "$10",
			data: []byte{0xde, 0x10},
			init: func(z *z80) { z.af = 0x10ff },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		{
			name: "sbc a,i == $7f",
			mn:   "sbc",
			dst:  "a",
			src:  "$7f",
			data: []byte{0xde, 0x7f},
			init: func(z *z80) { z.af = 0x0100 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0x8200 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		{
			name: "sbc a,i == $ff",
			mn:   "sbc",
			dst:  "a",
			src:  "$ff",
			data: []byte{0xde, 0xff},
			init: func(z *z80) { z.af = 0xff00 | carry },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&parity == 0 &&
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		// 0xdf
		{
			name: "rst $18",
			mn:   "rst",
			dst:  "$18",
			data: []byte{0xdf},
			init: func(z *z80) {
				z.sp = 0x5566
				z.bus.Write(0x5564, 0x64)
				z.bus.Write(0x5565, 0x65)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0018 && z.sp == 0x5564 &&
					z.bus.Read(0x5564) == 0x01 &&
					z.bus.Read(0x5565) == 0x00
			},
			dontSkipPC: true,
		},
		// 0xe1
		{
			name: "pop hl",
			mn:   "pop",
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
			mn:   "jp",
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
			mn:   "jp",
			dst:  "po",
			src:  "$1122",
			data: []byte{0xe2, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x1122
			},
			dontSkipPC: true,
		},
		// 0xe3
		{
			name: "ex (sp),hl",
			mn:   "ex",
			dst:  "(sp)",
			src:  "hl",
			data: []byte{0xe3},
			init: func(z *z80) {
				z.sp = 0x8856
				z.hl = 0x7012
				z.bus.Write(0x8856, 0x11)
				z.bus.Write(0x8857, 0x22)
			},
			expect: func(z *z80) bool {
				return z.hl == 0x2211 && z.pc == 0x0001 &&
					z.sp == 0x8856 &&
					z.bus.Read(0x8856) == 0x12 &&
					z.bus.Read(0x8857) == 0x70
			},
			dontSkipPC: true,
		},
		// 0xe5
		{
			name: "push hl",
			mn:   "push",
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
			mn:   "and",
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
			mn:   "and",
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
			mn:   "and",
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
		// 0xe7
		{
			name: "rst $20",
			mn:   "rst",
			dst:  "$20",
			data: []byte{0xe7},
			init: func(z *z80) {
				z.sp = 0x5566
				z.bus.Write(0x5564, 0x64)
				z.bus.Write(0x5565, 0x65)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0020 && z.sp == 0x5564 &&
					z.bus.Read(0x5564) == 0x01 &&
					z.bus.Read(0x5565) == 0x00
			},
			dontSkipPC: true,
		},
		// 0xe9
		{
			// jp (hl) DOES NOT dereference hl but according to
			// Zilog the mnemonic uses parenthesis
			name: "jp (hl)",
			mn:   "jp",
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
			mn:   "jp",
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
			mn:   "jp",
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
			mn:   "ex",
			dst:  "de",
			src:  "hl",
			data: []byte{0xeb},
			init: func(z *z80) { z.de = 0x1122; z.hl = 0x3344 },
			expect: func(z *z80) bool {
				return 0x1122 == z.hl && 0x3344 == z.de &&
					z.pc == 0x0001
			},
		},
		// 0xed 0x42 sbc hl,bc
		{
			name: "sbc hl,bc",
			mn:   "sbc",
			dst:  "hl",
			src:  "bc",
			data: []byte{0xed, 0x42},
			init: func(z *z80) { z.hl = 0x3344; z.bc = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.hl == 0x2222 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == addsub
			},
		},
		{
			name: "sbc hl,bc (1-1)",
			mn:   "sbc",
			dst:  "hl",
			src:  "bc",
			data: []byte{0xed, 0x42},
			init: func(z *z80) { z.hl = 0x0001; z.bc = 0x0001 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.hl == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == addsub
			},
		},
		{
			name: "sbc hl,bc (1-1-1)",
			mn:   "sbc",
			dst:  "hl",
			src:  "bc",
			data: []byte{0xed, 0x42},
			init: func(z *z80) {
				z.hl = 0x0001
				z.bc = 0x0001
				z.af |= carry
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.hl == 0xffff &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		// 0xed 0x44 neg
		// XXX add more test cases for all the flags
		{
			name: "neg 0",
			mn:   "neg",
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
			mn:   "neg",
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
			mn:   "neg",
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
		// 0xed 0x4a adc hl,bc
		{
			name: "adc hl,bc",
			mn:   "adc",
			dst:  "hl",
			src:  "bc",
			data: []byte{0xed, 0x4a},
			init: func(z *z80) { z.hl = 0x3344; z.bc = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.hl == 0x4466 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "adc hl,bc (-1+1)",
			mn:   "adc",
			dst:  "hl",
			src:  "bc",
			data: []byte{0xed, 0x4a},
			init: func(z *z80) { z.hl = 0xffff; z.bc = 0x0001 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.hl == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		{
			name: "adc hl,bc (-1+1+1)",
			mn:   "adc",
			dst:  "hl",
			src:  "bc",
			data: []byte{0xed, 0x4a},
			init: func(z *z80) {
				z.hl = 0xffff
				z.bc = 0x0001
				z.af |= carry
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.hl == 0x0001 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0xed 0x73 ld (nn),sp
		{
			name: "ld (nn),sp",
			mn:   "ld",
			dst:  "($1000)",
			src:  "sp",
			data: []byte{0xed, 0x73, 0x00, 0x10},
			init: func(z *z80) { z.sp = 0x4644 },
			expect: func(z *z80) bool {
				return z.pc == 0x0004 &&
					z.bus.Read(0x1000) == 0x44 &&
					z.bus.Read(0x1001) == 0x46 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		// 0xed 0x7b ld sp,(nn)
		{
			name: "ld sp,(nn)",
			mn:   "ld",
			dst:  "sp",
			src:  "($2130)",
			data: []byte{0xed, 0x7b, 0x30, 0x21},
			init: func(z *z80) {
				z.bus.Write(0x2130, 0x65)
				z.bus.Write(0x2131, 0x78)
				z.sp = 0x4644
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0004 &&
					z.sp == 0x7865 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0
			},
		},
		// 0xef
		{
			name: "rst $28",
			mn:   "rst",
			dst:  "$28",
			data: []byte{0xef},
			init: func(z *z80) {
				z.sp = 0x5566
				z.bus.Write(0x5564, 0x64)
				z.bus.Write(0x5565, 0x65)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0028 && z.sp == 0x5564 &&
					z.bus.Read(0x5564) == 0x01 &&
					z.bus.Read(0x5565) == 0x00
			},
			dontSkipPC: true,
		},
		// 0xf0
		{
			name: "ret p (P set)",
			mn:   "ret",
			dst:  "p",
			data: []byte{0xf0},
			init: func(z *z80) {
				z.af |= parity
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
			name: "ret p (P clear)",
			mn:   "ret",
			dst:  "p",
			data: []byte{0xf0},
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
		// 0xf1
		{
			name: "pop af",
			mn:   "pop",
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
		// 0xf3
		{
			name: "di",
			mn:   "di",
			data: []byte{0xf3},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.iff1 == 0 &&
					z.iff2 == 0
			},
		},
		// 0xf2
		{
			name: "jp p,nn (S set)",
			mn:   "jp",
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
			mn:   "jp",
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
			mn:   "push",
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
		// 0xf6
		{
			name: "or n",
			mn:   "or",
			dst:  "$f0",
			data: []byte{0xf6, 0xf0},
			init: func(z *z80) { z.af = 0xa500 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.af&0xff00 == 0xf500 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "or 0",
			mn:   "or",
			dst:  "$00",
			data: []byte{0xf6, 0x00},
			init: func(z *z80) { z.af = 0xa500 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.af&0xff00 == 0xa500 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "or 0 on 0",
			mn:   "or",
			dst:  "$00",
			data: []byte{0xf6, 0x00},
			init: func(z *z80) { z.af = 0x0000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.af&0xff00 == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "or $7f",
			mn:   "or",
			dst:  "$7f",
			data: []byte{0xf6, 0x7f},
			init: func(z *z80) { z.af = 0x8500 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.af&0xff00 == 0xff00 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xf7
		{
			name: "rst $30",
			mn:   "rst",
			dst:  "$30",
			data: []byte{0xf7},
			init: func(z *z80) {
				z.sp = 0x5566
				z.bus.Write(0x5564, 0x64)
				z.bus.Write(0x5565, 0x65)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0030 && z.sp == 0x5564 &&
					z.bus.Read(0x5564) == 0x01 &&
					z.bus.Read(0x5565) == 0x00
			},
			dontSkipPC: true,
		},
		// 0xfa
		{
			name: "jp m,nn (S set)",
			mn:   "jp",
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
			mn:   "jp",
			dst:  "m",
			src:  "$1122",
			data: []byte{0xfa, 0x22, 0x11},
			expect: func(z *z80) bool {
				return z.pc == 0x0003
			},
			dontSkipPC: true,
		},
		// 0xfb
		{
			name: "ei",
			mn:   "ei",
			data: []byte{0xfb},
			expect: func(z *z80) bool {
				return z.pc == 0x0001 && z.iff1 == 1 &&
					z.iff2 == 1
			},
		},
		// 0xfd 0x09 add iy,bc
		{
			name: "add iy,bc",
			mn:   "add",
			dst:  "iy",
			src:  "bc",
			data: []byte{0xfd, 0x09},
			init: func(z *z80) { z.iy = 0x3344; z.bc = 0x1122 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.iy == 0x4466 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		{
			name: "add iy,bc (-1+1)",
			mn:   "add",
			dst:  "iy",
			src:  "bc",
			data: []byte{0xfd, 0x09},
			init: func(z *z80) { z.iy = 0xffff; z.bc = 0x0001 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.iy == 0x0000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		{
			name: "add iy,bc (-1+2)",
			mn:   "add",
			dst:  "iy",
			src:  "bc",
			data: []byte{0xfd, 0x09},
			init: func(z *z80) {
				z.iy = 0xffff
				z.bc = 0x0002
				z.af |= carry
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0002 &&
					z.iy == 0x0001 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0 &&
					z.af&carry == carry
			},
		},
		// 0xfd 0x23
		{
			name: "inc iy",
			mn:   "inc",
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
			mn:   "inc",
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
			mn:   "inc",
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
			mn:   "dec",
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
			mn:   "dec",
			dst:  "iy",
			src:  "",
			data: []byte{0xfd, 0x2b},
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.iy == 0xffff
			},
		},
		{
			name: "dec iy == 0x8000",
			mn:   "dec",
			dst:  "iy",
			src:  "",
			data: []byte{0xfd, 0x2b},
			init: func(z *z80) { z.iy = 0x8000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.iy == 0x7fff
			},
		},
		// 0xfd 0x86
		{
			name: "add a,(iy+d)",
			mn:   "add",
			dst:  "a",
			src:  "(iy+$11)",
			data: []byte{0xfd, 0x86, 0x11},
			init: func(z *z80) {
				z.af = 0x1100
				z.iy = 0x3344
				z.bus.Write(0x3344+0x11, 0x55)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0003 &&
					z.af&0xff00 == 0x6600 &&
					z.iy == 0x3344 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&parity == 0 &&
					z.af&halfCarry == 0 &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xfd 0xa6
		{
			name: "and (iy+d)",
			mn:   "and",
			dst:  "(iy+$11)",
			data: []byte{0xfd, 0xa6, 0x11},
			init: func(z *z80) {
				z.af = 0xa500
				z.iy = 0x3344
				z.bus.Write(0x3344+0x11, 0xf0)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0003 &&
					z.af&0xff00 == 0xa000 &&
					z.iy == 0x3344 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&parity == parity &&
					z.af&halfCarry == halfCarry &&
					z.af&addsub == 0 &&
					z.af&carry == 0
			},
		},
		// 0xfd e1
		{
			name: "pop iy",
			mn:   "pop",
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
				return z.pc == 0x0002 && z.sp == 0xaa57 &&
					z.iy == 0xeeff
			},
		},
		// 0xfd 0xe5
		{
			name: "push iy",
			mn:   "push",
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
		// 0xfe
		{
			name: "cp i <",
			mn:   "cp",
			dst:  "$20",
			data: []byte{0xfe, 0x20},
			init: func(z *z80) { z.af = 0x1000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.af&0xff00 == 0x1000 &&
					z.af&sign == sign &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == carry
			},
		},
		{
			name: "cp i >",
			mn:   "cp",
			dst:  "$10",
			data: []byte{0xfe, 0x10},
			init: func(z *z80) { z.af = 0x2000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == 0 &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		{
			name: "cp i =",
			mn:   "cp",
			dst:  "$20",
			data: []byte{0xfe, 0x20},
			init: func(z *z80) { z.af = 0x2000 },
			expect: func(z *z80) bool {
				return z.pc == 0x0002 && z.af&0xff00 == 0x2000 &&
					z.af&sign == 0 &&
					z.af&zero == zero &&
					z.af&halfCarry == 0 &&
					// PV
					z.af&addsub == addsub &&
					z.af&carry == 0
			},
		},
		// 0xff
		{
			name: "rst $38",
			mn:   "rst",
			dst:  "$38",
			data: []byte{0xff},
			init: func(z *z80) {
				z.sp = 0x5566
				z.bus.Write(0x5564, 0x64)
				z.bus.Write(0x5565, 0x65)
			},
			expect: func(z *z80) bool {
				return z.pc == 0x0038 && z.sp == 0x5564 &&
					z.bus.Read(0x5564) == 0x01 &&
					z.bus.Read(0x5565) == 0x00
			},
			dontSkipPC: true,
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
		bus, err := bus.New(devices, make(chan string))
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

		mn, dst, src, opc, x, err := z.DisassembleComponents(0)
		if mn != test.mn {
			t.Fatalf("%v: invalid mnemonic got %v expected %v opcode %v",
				test.name, mn, test.mn, opc)
		}
		var opcExpected string
		for _, x := range test.data {
			opcExpected += fmt.Sprintf("%02x ", x)
		}
		if opc != strings.Trim(opcExpected, " ") {
			t.Fatalf("%v: invalid opcode got '%v' expected '%v'",
				test.name, opc, strings.Trim(opcExpected, " "))
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
		if opcodes[o].noBytes == 0 {
			continue
		}
		for _, test := range tests {
			if byte(o) == test.data[0] {
				goto next
			}
		}
		t.Logf("not implemented: 0x%02x", o)
	next:
	}
}
