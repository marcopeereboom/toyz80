package main

import "fmt"

func bitBR(reg, bit int) {
	// 0x40: {
	// 	mnemonic: []string{"bit", ""},
	// 	dst:      implied,
	// 	dstR:     []string{"0"},
	// 	src:      register,
	// 	srcR:     []string{"b"},
	// 	noBytes:  2,
	// 	noCycles: 8,
	// },
	//
	// Bit Tested b Register r
	//   0      000      B 000
	//   1      001      C 001
	//   2      010      D 010
	//   3      011      E 011
	//   4      100      H 100
	//   5      101      L 101
	//   6      110      A 111
	//   7      111
	//
	// 7 6 5 4 3 2 1 0
	// 0 1 r r r b b b
	register := []string{"b", "c", "d", "e", "h", "l", "", "a"}
	bitShift := uint(3)
	fmt.Printf("0x%02x: {\n", 0x40|reg|bit<<bitShift)
	fmt.Printf("\tmnemonic: []string{\"bit\", \"\"},\n")
	fmt.Printf("\tdst:      implied,\n")
	fmt.Printf("\tdstR:     []string{\"%v\"},\n", bit)
	fmt.Printf("\tsrc:      register,\n")
	fmt.Printf("\tsrcR:     []string{\"%v\"},\n",
		register[reg])
	fmt.Printf("\tnoBytes:  2,\n")
	fmt.Printf("\tnoCycles: 8,\n")
	fmt.Printf("\t},\n")
}

func bitBIR(reg, bit int) {
	bitShift := uint(3)
	fmt.Printf("0x%02x: {\n", 0x40|reg|bit<<bitShift)
	fmt.Printf("\tmnemonic: []string{\"bit\", \"\"},\n")
	fmt.Printf("\tdst:      implied,\n")
	fmt.Printf("\tdstR:     []string{\"%v\"},\n", bit)
	fmt.Printf("\tsrc:      registerIndirect,\n")
	fmt.Printf("\tsrcR:     []string{\"hl\"},\n")
	fmt.Printf("\tnoBytes:  2,\n")
	fmt.Printf("\tnoCycles: 12,\n")
	fmt.Printf("\t},\n")
}

func bitBRCode(reg, bit int) {
	bitShift := uint(3)
	register := []string{"b", "c", "d", "e", "h", "l", "", "a"}
	actualReg := []string{
		"byte(z.bc>>8)",
		"byte(z.bc)",
		"byte(z.de>>8)",
		"byte(z.de)",
		"byte(z.hl>>8)",
		"byte(z.hl)",
		"",
		"byte(z.af>>8)"}
	fmt.Printf("case 0x%02x: // bit %v,%v\n", 0x40|reg|bit<<bitShift, bit,
		register[reg])
	fmt.Printf("\tz.bit(%v, %v)\n", bit, actualReg[reg])
}

func bitBIRCode(reg, bit int) {
	bitShift := uint(3)
	fmt.Printf("case 0x%02x: // bit %v,(hl)\n", 0x40|reg|bit<<bitShift, bit)
	fmt.Printf("\tz.bit(%v, z.bus.Read(z.hl))\n", bit)
}

// ---		BIT	0,(IX+index)	DDCBindex46	Z flag <- NOT 0b

// 0xdd 0xcb INDEX 0x46
// 0x40: {
// 	mnemonic: []string{"bit"},
// 	dst:      implied,
// 	dstR:     []string{"0"},
// 	src:      indexedStupid,
// 	srcR:     []string{"b"},
// 	noBytes:  2,
// 	noCycles: 8,
// },
func main() {
	fmt.Println("// opcodes 0xcbXX bit b,r")
	for bit := 0; bit < 8; bit++ {
		for reg := 0; reg < 8; reg++ {
			if reg == 6 {
				bitBIR(reg, bit)
				continue
			}
			bitBR(reg, bit)
		}
	}

	fmt.Println("// code 0xcbXX bit b,r")
	for bit := 0; bit < 8; bit++ {
		for reg := 0; reg < 8; reg++ {
			if reg == 6 {
				bitBIRCode(reg, bit)
				continue
			}
			bitBRCode(reg, bit)
		}
	}

	fmt.Println()
	fmt.Println("// 0xddcb bit b,(ix+d)")
}
