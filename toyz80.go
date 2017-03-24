package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/marcopeereboom/toyz80/bus"
	"github.com/marcopeereboom/toyz80/z80"
)

// z80comp is our fictious z80 based computer
//
// Memory map
// ROM	0x0000-0x0fff boot
// RAM	0x1000-0xffff working memory
//
// IO space
// 0x00	console status
// 0x01	console data
type z80comp struct {
	// bus Bus
	// cpu CPU
}

func _main() error {
	//var memory []byte
	var (
		bootImage = flag.String("boot", "", "boot ROM image")
		ramImage  = flag.String("ramimage", "", "main RAM image")
		logFile   = flag.String("log", "stderr", "log trace")
		traceFlag = flag.Bool("trace", false, "trace execution")
		boot      []byte
		ram       []byte
		err       error
	)
	flag.Parse()

	if *bootImage == "" {
		// minimal rom
		boot = []byte{
			0x31, 0x00, 0xee, // ld sp,$e000
			0xc3, 0x00, 0x10, // jp $1000
		}
		ram = []byte{
			0x76, // halt
		}
	} else {
		boot, err = ioutil.ReadFile(*bootImage)
		if err != nil {
			return err
		}
	}

	if *ramImage != "" {
		ram, err = ioutil.ReadFile(*ramImage)
		if err != nil {
			return err
		}
	}

	trace := *traceFlag

	f := os.Stderr
	if *logFile != "stderr" {
		f, err = os.Create(*logFile)
		if err != nil {
			return err
		}
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
			Size:  0xe000,
			Type:  bus.DeviceRAM,
			Image: ram,
		},
		// I/O space
		{
			Name:  "serial console",
			Start: 0x02, // 0x02 data 0x03 UART status
			Size:  0x02,
			Type:  bus.DeviceSerialConsole,
		},
	}
	bus, err := bus.New(devices)
	if err != nil {
		return err
	}

	z, err := z80.New(z80.ModeZ80, bus)
	if err != nil {
		return err
	}

	var prefix string
	for {
		if trace {
			s, pc, _, err := z.DisassemblePC(true)
			if err != nil {
				return err
			}
			prefix = fmt.Sprintf("%04x: %v", pc, s)
		}

		err := z.Step()

		if trace {
			fmt.Fprintf(f, "%-35s%s\n", prefix, z.DumpRegisters())
		}

		if err != nil && err != z80.ErrHalt {
			return err
		}
	}

	return nil
}

func main() {
	err := _main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
