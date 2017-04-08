package main

import "fmt"

func setBR(reg, bit int) {
	// 0xc0: {
	// 	mnemonic: []string{"set", ""},
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
	fmt.Printf("0x%02x: {\n", 0xc0|reg|bit<<bitShift)
	fmt.Printf("\tmnemonic: []string{\"set\", \"\"},\n")
	fmt.Printf("\tdst:      implied,\n")
	fmt.Printf("\tdstR:     []string{\"%v\"},\n", bit)
	fmt.Printf("\tsrc:      register,\n")
	fmt.Printf("\tsrcR:     []string{\"%v\"},\n",
		register[reg])
	fmt.Printf("\tnoBytes:  2,\n")
	fmt.Printf("\tnoCycles: 8,\n")
	fmt.Printf("\t},\n")
}

func setBIR(reg, bit int) {
	bitShift := uint(3)
	fmt.Printf("0x%02x: {\n", 0xc0|reg|bit<<bitShift)
	fmt.Printf("\tmnemonic: []string{\"set\", \"\"},\n")
	fmt.Printf("\tdst:      implied,\n")
	fmt.Printf("\tdstR:     []string{\"%v\"},\n", bit)
	fmt.Printf("\tsrc:      registerIndirect,\n")
	fmt.Printf("\tsrcR:     []string{\"hl\"},\n")
	fmt.Printf("\tnoBytes:  2,\n")
	fmt.Printf("\tnoCycles: 12,\n")
	fmt.Printf("\t},\n")
}

func setBRCode(reg, bit int) {
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
		"byte(z.af>>8)",
	}
	assignReg := []string{
		"z.bc",
		"z.bc",
		"z.de",
		"z.de",
		"z.hl",
		"z.hl",
		"",
		"z.af",
	}
	maskReg := []string{
		"z.bc&0x00ff",
		"z.bc&0xff00",
		"z.de&0x00ff",
		"z.de&0xff00",
		"z.hl&0x00ff",
		"z.hl&0xff00",
		"",
		"z.af&0x00ff",
	}
	shiftReg := []string{
		"<<8",
		"",
		"<<8",
		"",
		"<<8",
		"",
		"",
		"<<8",
	}
	fmt.Printf("case 0x%02x: // set %v,%v\n", 0xc0|reg|bit<<bitShift, bit,
		register[reg])
	fmt.Printf("\t%v = %v | uint16(z.set(%v, %v))%v\n", assignReg[reg],
		maskReg[reg], bit, actualReg[reg], shiftReg[reg])
}

func setBIRCode(reg, bit int) {
	bitShift := uint(3)
	fmt.Printf("case 0x%02x: // set %v,(hl)\n", 0xc0|reg|bit<<bitShift, bit)
	fmt.Printf("\tval := z.set(%v, z.bus.Read(z.hl))\n", bit)
	fmt.Printf("\tz.bus.Write(z.hl, val)\n")
}

func main() {
	fmt.Println("// opcodes 0xcbXX set b,r")
	for bit := 0; bit < 8; bit++ {
		for reg := 0; reg < 8; reg++ {
			if reg == 6 {
				setBIR(reg, bit)
				continue
			}
			setBR(reg, bit)
		}
	}

	fmt.Println("// code 0xcbXX set b,r")
	for bit := 0; bit < 8; bit++ {
		for reg := 0; reg < 8; reg++ {
			if reg == 6 {
				setBIRCode(reg, bit)
				continue
			}
			setBRCode(reg, bit)
		}
	}
}
