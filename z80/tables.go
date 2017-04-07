// Copyright (c) 2012 Andrea Fazzi
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package z80

const (
	FLAG_C = byte(carry)
	FLAG_N = byte(addsub)
	FLAG_V = byte(parity) // V == P
	FLAG_P = byte(parity) // V == P
	FLAG_3 = byte(unused)
	FLAG_H = byte(halfCarry)
	FLAG_5 = byte(unused2)
	FLAG_Z = byte(zero)
	FLAG_S = byte(sign)
)

func ternB(cond bool, ret1, ret2 byte) byte {
	if cond {
		return ret1
	}
	return ret2
}

// Whether a half carry occurred or not can be determined by looking at the 3rd
// bit of the two arguments and the result; these are hashed into this table in
// the form r12, where r is the 3rd bit of the result, 1 is the 3rd bit of the
// 1st argument and 2 is the third bit of the 2nd argument; the tables differ
// for add and subtract operations
var halfcarryAddTable = []byte{0, FLAG_H, FLAG_H, FLAG_H, 0, 0, 0, FLAG_H}
var halfcarrySubTable = []byte{0, 0, FLAG_H, 0, FLAG_H, 0, FLAG_H, FLAG_H}

// Similarly, overflow can be determined by looking at the 7th bits; again the
// hash into this table is r12
var overflowAddTable = []byte{0, 0, 0, FLAG_V, FLAG_V, 0, 0, 0}
var overflowSubTable = []byte{0, FLAG_V, 0, 0, 0, 0, FLAG_V, 0}

var sz53Table, sz53pTable, parityTable [0x100]byte

func (z *z80) adc16(val uint16) {
	t := uint(z.hl) + uint(val) + uint(z.af&carry)
	lookup := byte(z.hl&0x8800>>11 | val&0x8800>>10 | uint16(t&0x8800>>9))
	z.hl = uint16(t)

	f := ternB(t&0x10000 != 0, FLAG_C, 0) | overflowAddTable[lookup>>4] |
		byte(z.hl>>8)&(FLAG_3|FLAG_5|FLAG_S) |
		halfcarryAddTable[lookup&0x07] |
		ternB(z.hl != 0, 0, FLAG_Z)
	z.af = z.af&0xff00 | uint16(f)
}

func (z *z80) add16(v1, v2 uint16) uint16 {
	t := uint(v1) + uint(v2)
	lookup := byte(v1&0x0800>>11 | v2&0x0800>>10 | uint16(t&0x0800>>9))
	f := z.af&uint16(FLAG_V|FLAG_Z|FLAG_S) |
		uint16(ternB(t&0x10000 != 0, FLAG_C, 0)|
			(byte(t>>8)&(FLAG_3|FLAG_5))|halfcarryAddTable[lookup])
	z.af = z.af&0xff00 | f
	return uint16(t)
}

func (z *z80) add(val byte) {
	t := uint16(z.af>>8) + uint16(val)
	lookup := byte(z.af>>8)&0x88>>3 | val&0x88>>2 | byte(t)&0x88>>1
	z.af = t<<8 | uint16(ternB(t&0x100 != 0, FLAG_C, 0)|
		halfcarryAddTable[lookup&0x07]|overflowAddTable[lookup>>4]|
		sz53Table[byte(t)])
}

func (z *z80) adc(val byte) {
	a := byte(z.af >> 8)
	t := uint16(a) + uint16(val) + uint16(z.af&carry)
	lookup := uint16(a)&0x88>>3 | uint16(val)&0x88>>2 | t&0x88>>1
	f := ternB(t&0x100 != 0, FLAG_C, 0) | halfcarryAddTable[lookup&0x07] |
		overflowAddTable[lookup>>4] | sz53Table[byte(t)]
	z.af = t<<8 | uint16(f)
}

func (z *z80) and(val byte) {
	a := byte(z.af>>8) & val
	z.af = uint16(a)<<8 | halfCarry | uint16(sz53pTable[a])
}

func (z *z80) bit(bit, val byte) {
	f := byte(z.af)&FLAG_C | FLAG_H | val&(FLAG_3|FLAG_5)
	if val&(0x01<<bit) == 0 {
		f |= FLAG_P | FLAG_Z
	}
	if bit == 7 && (val&0x80) != 0 {
		f |= FLAG_S
	}
	z.af = z.af&0xff00 | uint16(f)
}

func (z *z80) cp(val byte) {
	a := byte(z.af >> 8)
	aTmp := uint16(a) - uint16(val)
	lookup := ((a & 0x88) >> 3) | ((val & 0x88) >> 2) | byte((aTmp&0x88)>>1)
	f := ternB((aTmp&0x100) != 0, FLAG_C, ternB(aTmp != 0, 0, FLAG_Z)) |
		FLAG_N | halfcarrySubTable[lookup&0x07] |
		overflowSubTable[lookup>>4] | (val & (FLAG_3 | FLAG_5)) |
		byte(aTmp)&FLAG_S
	z.af = z.af&0xff00 | uint16(f)
}

func (z *z80) cpd() {
	a := byte(z.af >> 8)
	val := z.bus.Read(z.hl)
	t := a - val
	lookup := a&0x08>>3 | val&0x08>>2 | t&0x08>>1
	z.hl--
	z.bc--
	f := byte(z.af)&FLAG_C | ternB(z.bc != 0, FLAG_V|FLAG_N, FLAG_N) |
		halfcarrySubTable[lookup] | ternB(t != 0, 0, FLAG_Z) | t&FLAG_S
	if (f & FLAG_H) != 0 {
		t--
	}
	f |= t&FLAG_3 | ternB(t&0x02 != 0, FLAG_5, 0)
	z.af = z.af&0xff00 | uint16(f)
}

func (z *z80) cpi() {
	a := byte(z.af >> 8)
	val := z.bus.Read(z.hl)
	t := a - val
	lookup := a&0x08>>3 | val&0x08>>2 | t&0x08>>1
	z.hl++
	z.bc--
	f := byte(z.af)&FLAG_C | ternB(z.bc != 0, FLAG_V|FLAG_N, FLAG_N) |
		halfcarrySubTable[lookup] | ternB(t != 0, 0, FLAG_Z) |
		(t & FLAG_S)
	if f&FLAG_H != 0 {
		t--
	}
	f |= (t & FLAG_3) | ternB(t&0x02 != 0, FLAG_5, 0)
	z.af = z.af&0xff00 | uint16(f)
}

func (z *z80) daa() {
	a := byte(z.af >> 8)
	f := byte(z.af)
	add := byte(0)
	carryIn := f & FLAG_C
	if f&FLAG_H != 0 || a&0x0f > 9 {
		add = 6
	}
	if carryIn != 0 || a > 0x99 {
		add |= 0x60
	}
	if a > 0x99 {
		carryIn = FLAG_C
	}
	if f&FLAG_N != 0 {
		z.sub(add)
	} else {
		z.add(add)
	}
	temp := byte(f&^(FLAG_C|FLAG_P)) | carryIn | parityTable[a]
	z.af = z.af&0xff00 | uint16(temp)
}

func (z *z80) dec(val byte) byte {
	f := byte(z.af)&FLAG_C | ternB(val&0x0f != 0, 0, FLAG_H) | FLAG_N
	val--
	f |= ternB(val == 0x7f, FLAG_V, 0) | sz53Table[val]
	z.af = z.af&0xff00 | uint16(f)
	return val
}

func (z *z80) inc(val byte) byte {
	val++
	f := byte(z.af)&FLAG_C | ternB(val == 0x80, FLAG_V, 0) |
		ternB(val&0x0f != 0, 0, FLAG_H) | sz53Table[val]
	z.af = z.af&0xff00 | uint16(f)
	return val
}

func (z *z80) or(val byte) {
	a := byte(z.af>>8) | val
	z.af = uint16(a)<<8 | uint16(sz53pTable[a])
}

func (z *z80) ldd() {
	t := z.bus.Read(z.hl)
	z.bc--
	z.bus.Write(z.de, t)
	z.de--
	z.hl--
	t += byte(z.af >> 8)
	f := byte(z.af)&(FLAG_C|FLAG_Z|FLAG_S) |
		ternB(z.bc != 0, FLAG_V, 0) | t&FLAG_3 |
		ternB(t&0x02 != 0, FLAG_5, 0)
	z.af = z.af&0xff00 | uint16(f)
}

func (z *z80) ldi() {
	t := z.bus.Read(z.hl)
	z.bc--
	z.bus.Write(z.de, t)
	z.de++
	z.hl++
	t += byte(z.af >> 8)
	f := byte(z.af)&(FLAG_C|FLAG_Z|FLAG_S) |
		ternB(z.bc != 0, FLAG_V, 0) | t&FLAG_3 |
		ternB(t&0x02 != 0, FLAG_5, 0)
	z.af = z.af&0xff00 | uint16(f)
}

func (z *z80) rrd() {
	a := byte(z.af >> 8)
	t := z.bus.Read(z.hl)
	z.bus.Write(z.hl, a<<4|t>>4)
	a = a&0xf0 | t&0x0f
	f := byte(z.af)&FLAG_C | sz53pTable[a]
	z.af = uint16(a)<<8 | uint16(f)
}

func (z *z80) rld() {
	a := byte(z.af >> 8)
	t := z.bus.Read(z.hl)
	z.bus.Write(z.hl, t<<4|a&0x0f)
	a = a&0xf0 | t>>4
	f := byte(z.af)&FLAG_C | sz53pTable[a]
	z.af = uint16(a)<<8 | uint16(f)
}

func (z *z80) sla(val byte) byte {
	f := val >> 7
	val <<= 1
	z.af = z.af&0xff00 | uint16(f) | uint16(sz53pTable[val])
	return val
}

func (z *z80) srl(val byte) byte {
	f := val & FLAG_C
	val >>= 1
	z.af = uint16(f) | uint16(sz53pTable[val])
	return val
}

func (z *z80) sbc(val byte) {
	a := byte(z.af >> 8)
	t := uint16(a) - uint16(val) - z.af&carry
	lookup := a&0x88>>3 | val&0x88>>2 | byte(t&0x88>>1)
	f := ternB((t&0x100) != 0, FLAG_C, 0) | FLAG_N |
		halfcarrySubTable[lookup&0x07] |
		overflowSubTable[lookup>>4] |
		sz53Table[byte(t)]
	z.af = uint16(t)<<8 | uint16(f)
}

func (z *z80) sbc16(val uint16) {
	t := uint(z.hl) - uint(val) - uint(z.af&carry)
	lookup := byte(z.hl&0x8800>>11 | val&0x8800>>10 | uint16(t)&0x8800>>9)
	z.hl = uint16(t)

	f := ternB(t&0x10000 != 0, FLAG_C, 0) | FLAG_N |
		overflowSubTable[lookup>>4] |
		byte(z.hl&0xff00>>8)&(FLAG_3|FLAG_5|FLAG_S) |
		halfcarrySubTable[lookup&0x07] | ternB(z.hl != 0, 0, FLAG_Z)
	z.af = z.af&0xff00 | uint16(f)
}

func (z *z80) sub(val byte) {
	a := byte(z.af >> 8)
	t := uint16(a) - uint16(val)
	lookup := a&0x88>>3 | val&0x88>>2 | byte(t&0x88>>1)
	a = byte(t)
	z.af = uint16(a)<<8 | uint16(ternB(t&0x100 != 0, FLAG_C, 0)|FLAG_N|
		halfcarrySubTable[lookup&0x07]|overflowSubTable[lookup>>4]|
		sz53Table[a])
}

func (z *z80) xor(val byte) {
	a := byte(z.af>>8) ^ val
	z.af = uint16(a)<<8 | uint16(sz53pTable[a])
}

func init() {
	var i int16
	var j, k byte
	var p byte

	for i = 0; i < 0x100; i++ {
		sz53Table[i] = byte(i) & (0x08 | 0x20 | 0x80)
		j = byte(i)
		p = 0
		for k = 0; k < 8; k++ {
			p ^= j & 1
			j >>= 1
		}
		parityTable[i] = ternB(p != 0, 0, 0x04)
		sz53pTable[i] = sz53Table[i] | parityTable[i]
	}

	sz53Table[0] |= 0x40
	sz53pTable[0] |= 0x40
}
