package main

import (
	"fmt"

	"github.com/marcopeereboom/toyz80/bus"
	"github.com/marcopeereboom/toyz80/z80"
)

// z80comp is our fictious z80 based computer
//
// Memory map
// ROM	0x0000-0x0fff boot
// RAM	0x1000-0xcfff working memory
// ROM  0xe000-0xefff basic rom
// RAM	0xf000-0xffff basic scratch space
//
// IO space
// 0x00	console status
// 0x01	console data
type z80comp struct {
	// bus Bus
	// cpu CPU
}

func main() {
	//var memory []byte

	boot := []byte{
		0x31, 0xc4, 0x07, // ld sp,$07c4
		0xc3, 0x00, 0x10, // jp $1000
	}

	ram := []byte{
		0x3e, 'a', // ld a, 'a'
		0xd3, 0x81, // out ($01), a
		0x76, // halt
	}

	devices := []bus.Device{
		bus.Device{
			Name:  "Boot ROM",
			Start: 0x0000,
			Size:  0x1000,
			Type:  bus.DeviceROM,
			Image: boot,
		},
		{
			Name:  "working memory",
			Start: 0x1000,
			Size:  0x1000,
			Type:  bus.DeviceRAM,
			Image: ram,
		},
		// I/O space
		{
			Name:  "console",
			Start: 0x80, // 0x00 status 0x01 data
			Size:  0x02,
			Type:  bus.DeviceSimpleConsole,
		},
	}
	bus, err := bus.New(devices)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	cpu, err := z80.New(z80.ModeZ80, bus)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	trace, registers, err := cpu.Trace()
	if err != nil && err != z80.ErrHalt {
		fmt.Printf("trace: %v\n", err)
		return
	}
	for i, line := range trace {
		fmt.Printf("%-25s%s\n", line, registers[i])
	}
}
