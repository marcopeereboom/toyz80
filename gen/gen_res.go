package main

import "fmt"

func resBR(reg, bit int) {
	// 0x80: {
	// 	mnemonic: []string{"res", ""},
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
	fmt.Printf("0x%02x: {\n", 0x80|reg|bit<<bitShift)
	fmt.Printf("\tmnemonic: []string{\"res\", \"\"},\n")
	fmt.Printf("\tdst:      implied,\n")
	fmt.Printf("\tdstR:     []string{\"%v\"},\n", bit)
	fmt.Printf("\tsrc:      register,\n")
	fmt.Printf("\tsrcR:     []string{\"%v\"},\n",
		register[reg])
	fmt.Printf("\tnoBytes:  2,\n")
	fmt.Printf("\tnoCycles: 8,\n")
	fmt.Printf("\t},\n")
}

func resBIR(reg, bit int) {
	bitShift := uint(3)
	fmt.Printf("0x%02x: {\n", 0x80|reg|bit<<bitShift)
	fmt.Printf("\tmnemonic: []string{\"res\", \"\"},\n")
	fmt.Printf("\tdst:      implied,\n")
	fmt.Printf("\tdstR:     []string{\"%v\"},\n", bit)
	fmt.Printf("\tsrc:      registerIndirect,\n")
	fmt.Printf("\tsrcR:     []string{\"hl\"},\n")
	fmt.Printf("\tnoBytes:  2,\n")
	fmt.Printf("\tnoCycles: 12,\n")
	fmt.Printf("\t},\n")
}

func resBRCode(reg, bit int) {
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
	fmt.Printf("case 0x%02x: // res %v,%v\n", 0x80|reg|bit<<bitShift, bit,
		register[reg])
	fmt.Printf("\t%v = %v | uint16(z.res(%v, %v))%v\n", assignReg[reg],
		maskReg[reg], bit, actualReg[reg], shiftReg[reg])
}

func resBIRCode(reg, bit int) {
	bitShift := uint(3)
	fmt.Printf("case 0x%02x: // res %v,(hl)\n", 0x80|reg|bit<<bitShift, bit)
	fmt.Printf("\tval := z.res(%v, z.bus.Read(z.hl))\n", bit)
	fmt.Printf("\tz.bus.Write(z.hl, val)\n")
}

func main() {
	fmt.Println("// opcodes 0xcbXX res b,r")
	for bit := 0; bit < 8; bit++ {
		for reg := 0; reg < 8; reg++ {
			if reg == 6 {
				resBIR(reg, bit)
				continue
			}
			resBR(reg, bit)
		}
	}

	fmt.Println("// code 0xcbXX res b,r")
	for bit := 0; bit < 8; bit++ {
		for reg := 0; reg < 8; reg++ {
			if reg == 6 {
				resBIRCode(reg, bit)
				continue
			}
			resBRCode(reg, bit)
		}
	}
}
