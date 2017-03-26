package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

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
		logFile   = flag.String("log", "stderr", "log trace")
		traceFlag = flag.Bool("trace", false, "trace execution")
		err       error
	)
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nusage: %v device={rom|ram|console},"+
			"origin-size[,image] load=origin,image\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "example: %v device=console,0x02-0x02 "+
			"ram,0x0000-0x10000 load=mysuper.rom\n",
			os.Args[0])
	}
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		return nil
	}

	// logging
	trace := *traceFlag
	f := os.Stderr
	if *logFile != "stderr" {
		f, err = os.Create(*logFile)
		if err != nil {
			return err
		}
	}

	// devices
	var devices []bus.Device
	var loads []string
	for _, args := range flag.Args() {
		// format rom,0x1000-0x1000,image
		cmd := strings.Split(args, "=")
		if len(cmd) != 2 {
			return fmt.Errorf("expected device={rom|ram|console}," +
				"origin-size[,image] load=origin,image")
		}
		switch cmd[0] {
		case "load":
			loads = append(loads, cmd[1])
			continue
		case "device":
		default:
			flag.Usage()
			return nil
		}

		// device
		a := strings.Split(cmd[1], ",")

		// memory type
		var d bus.BusDeviceType
		switch a[0] {
		case "rom":
			d = bus.DeviceROM
		case "ram":
			d = bus.DeviceRAM
		case "console":
			d = bus.DeviceSerialConsole
		default:
			return fmt.Errorf("invalid device type: %v", a[0])
		}

		// origin address
		o := strings.Split(a[1], "-")
		if len(o) != 2 {
			return fmt.Errorf("invalid origin-size: %v", a[1])
		}
		origin, err := strconv.ParseUint(o[0], 0, 16)
		if err != nil {
			return fmt.Errorf("invalid origin: %v", err)
		}

		// size
		size, err := strconv.ParseUint(o[1], 0, 32)
		if err != nil {
			return fmt.Errorf("invalid size: %v", err)
		}
		if origin+size > 65536 {
			return fmt.Errorf("size out of bounds: %v", o[1])
		}

		// warn user if there is no rom image
		if a[0] == "rom" && len(a) != 3 {
			fmt.Printf("warning rom @ 0x%04x does not have an "+
				"image\n", origin)
		}

		// load image
		var image []byte
		if len(a) == 3 {
			image, err = ioutil.ReadFile(a[2])
			if err != nil {
				return err
			}
		}
		if int(origin)+len(image) > 65536 {
			return fmt.Errorf("image out of bounds: %v", a[2])
		}

		devices = append(devices, bus.Device{
			Name:  a[0],
			Start: uint16(origin),
			Size:  int(size),
			Type:  d,
			Image: image,
		})
	}

	shutdown := make(chan string)
	bus, err := bus.New(devices, shutdown)
	if err != nil {
		return err
	}

	z, err := z80.New(z80.ModeZ80, bus)
	if err != nil {
		return err
	}

	// load memory
	for _, load := range loads {
		a := strings.Split(load, ",")
		var image []byte
		if len(a) != 2 {
			flag.Usage()
			return fmt.Errorf("invalid load: %v", load)
		}
		origin, err := strconv.ParseUint(a[0], 0, 16)
		if err != nil {
			return err
		}
		if int(origin)+len(image) > 65536 {
			return fmt.Errorf("image out of bounds: %v",
				a[1])
		}

		image, err = ioutil.ReadFile(a[1])
		if err != nil {
			return err
		}

		err = bus.WriteMemory(uint16(origin), image)
		if err != nil {
			return err
		}
	}

	var prefix string
	for {
		select {
		case reason := <-shutdown:
			return fmt.Errorf("shutdown requested: %v", reason)
		default:
		}

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

		if err != nil {
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
