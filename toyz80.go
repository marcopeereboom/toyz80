package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
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

var completer = readline.NewPrefixCompleter(
	// edit modes
	readline.PcItem("mode",
		readline.PcItem("emacs"),
		readline.PcItem("vi"),
	),

	// emulator commands
	readline.PcItem("bp",
		readline.PcItem("set"),
		readline.PcItem("del")),
	readline.PcItem("continue"),
	readline.PcItem("disassemble"),
	readline.PcItem("dump"),
	readline.PcItem("help"),
	readline.PcItem("pause"),
	readline.PcItem("registers"),
	readline.PcItem("step"),
	readline.PcItem("pc"),
)

func help() {
	h := [][]string{
		{"bp <set|del>", "Breakpoint, leave empty to list."},
		{"continue", "Resume execution."},
		{"disassemble [address[ count]]",
			"Disassemble starting at provided address."},
		{"dump [address[ count]]",
			"Dump memory starting at provided address."},
		{"help", "This help."},
		{"mode <emacs|vi>", "Set edit mode."},
		{"pause", "Pause execution."},
		{"pc <address>", "Set program counter to address."},
		{"registers", "Print registers."},
		{"step [count]", "Execute next instruction."},
	}
	for i := range h {
		fmt.Printf("%-32v%v\n", h[i][0], h[i][1])
	}
}

func filterInput(r rune) (rune, bool) {
	//switch r {
	//// block CtrlZ feature
	//case readline.CharCtrlZ:
	//	return r, false
	//}
	return r, true
}

// hexify converts $ into 0x in order to simulate oldskool hex.
func hexify(s string) string {
	if strings.HasPrefix(s, "$") {
		return "0x" + s[1:]
	}
	return s
}

func parseUint(s string, size int) (uint64, error) {
	return strconv.ParseUint(hexify(strings.TrimSpace(s)), 0, size)
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
		origin, err := parseUint(o[0], 16)
		if err != nil {
			return fmt.Errorf("invalid origin: %v", err)
		}

		// size
		size, err := parseUint(o[1], 32)
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
		origin, err := parseUint(a[0], 16)
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

	// setup readline
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return err
	}
	defer l.Close()

	log.SetOutput(l.Stderr())

	// setup main loop
	restart := make(chan string)
	pause := true
	firstPause := false
	interactive := true
	stepCount := uint64(0)
	go func() {
		var prefix string
		for {
			if pause {
				if firstPause {
					firstPause = false
					fmt.Fprintf(l.Stdout(), "%v\n",
						z.DumpRegisters())
				}
				if stepCount == 0 {
					select {
					case action := <-restart:
						switch action {
						case "registers":
							fmt.Fprintf(l.Stdout(),
								"%v\n",
								z.DumpRegisters())
							continue
						}
					}
				}
			}
			select {
			case reason := <-shutdown:
				fmt.Fprintf(l.Stdout(),
					"shutdown requested: %v", reason)
				return
			default:
			}

			if trace || stepCount > 0 {
				s, pc, _, err := z.DisassemblePC(true)
				if err != nil {
					fmt.Fprintf(l.Stdout(),
						"DisassemblePC: %v", err)
					return
				}
				prefix = fmt.Sprintf("%04x: %v", pc, s)
			}

			err := z.Step()

			if trace || stepCount > 0 {
				output := fmt.Sprintf("%-35s%s", prefix,
					z.DumpRegisters())
				if interactive {
					fmt.Fprintf(l.Stdout(), "%v\n", output)
				}
				if trace {
					fmt.Fprintf(f, "%v\n", output)
				}
			}

			if err != nil {
				switch err.(type) {
				case z80.HaltError:
					fmt.Fprintf(l.Stdout(), "CPU halted\n")
					pause = true
					fmt.Fprintf(l.Stdout(), "%v\n", err)
					fmt.Fprintf(l.Stdout(), "%v\n",
						z.DumpRegisters())
				case z80.BreakpointError:
					pause = true
					fmt.Fprintf(l.Stdout(), "%v\n", err)
					fmt.Fprintf(l.Stdout(), "%v\n",
						z.DumpRegisters())
				default:
					fmt.Fprintf(l.Stdout(),
						"CPU error: %v\n", err)
				}
			}
			if stepCount > 0 {
				stepCount--
			}
		}
	}()

	var (
		last     string
		lastAddr uint16
	)
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		replay := false
		if line == "" {
			replay = true
			line = last
		}
		switch {
		case strings.HasPrefix(line, "mode "):
			switch line[5:] {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			default:
				println("invalid mode:", line[5:])
			}
		case line == "mode":
			if l.IsVimMode() {
				println("current mode: vim")
			} else {
				println("current mode: emacs")
			}
		case line == "help":
			help()
		case line == "continue" || line == "run":
			if !pause {
				fmt.Printf("CPU is currently running\n")
				continue
			}
			pause = false
			restart <- ""
		case line == "pause":
			if pause {
				fmt.Printf("CPU is not running\n")
				continue
			}
			pause = true
			firstPause = true // dump registers
		case line == "disassemble",
			strings.HasPrefix(line, "disassemble "):
			if line == "disassemble" {
				line += " " + strconv.Itoa(int(lastAddr))
			}
			a := strings.Split(line[12:], " ")

			// line count
			lines := 24
			if len(a) == 2 {
				l, err := parseUint(a[1], 16)
				if err != nil {
					fmt.Printf("invalid line count: %v\n",
						err)
					continue
				}
				lines = int(l)
			}

			// address
			address, err := parseUint(a[0], 16)
			if err != nil {
				fmt.Printf("invalid address: %v\n", err)
				continue
			}

			// continue
			var addr uint16
			if replay {
				addr = lastAddr
			} else {
				addr = uint16(address)
			}

			// actually disassemble
			for i := 0; i < lines; i++ {
				s, _, count, _ := z.Disassemble(addr, true)
				fmt.Printf("%04x: %v\n", addr, s)
				addr += uint16(count)
			}
			lastAddr = addr
		case line == "dump", strings.HasPrefix(line, "dump "):
			if line == "dump" {
				line += " " + strconv.Itoa(int(lastAddr))
			}
			a := strings.Split(line[5:], " ")

			// line count
			lines := 24
			if len(a) == 2 {
				l, err := parseUint(a[1], 16)
				if err != nil {
					fmt.Printf("invalid line count: %v\n",
						err)
					continue
				}
				lines = int(l)
			}

			// address
			address, err := parseUint(a[0], 16)
			if err != nil {
				fmt.Printf("invalid address: %v\n", err)
				continue
			}

			// continue
			var addr uint16
			if replay {
				addr = lastAddr
			} else {
				addr = uint16(address)
			}

			// actually disassemble
			for i := 0; i < lines; i++ {
				s := bus.Dump(addr, 16)
				sd := hex.Dump(s)
				sd = sd[10 : len(sd)-1] // chop of addr and \n
				fmt.Printf("%04x: %v\n", addr, sd)
				addr += 16
			}
			lastAddr = addr
		case strings.HasPrefix(line, "pc "):
			x, err := parseUint(line[3:], 16)
			if err != nil {
				fmt.Printf("invalid PC: %v\n", err)
				continue
			}
			if pause == false {
				fmt.Printf("CPU is currently running\n")
				continue
			}
			z.SetPC(uint16(x))
		case strings.HasPrefix(line, "step "):
			x, err := parseUint(line[5:], 32)
			if err != nil {
				fmt.Printf("invalid step count: %v\n", err)
				continue
			}
			if pause == false {
				fmt.Printf("CPU is currently running\n")
				continue
			}
			stepCount = x
			restart <- ""
		case line == "step":
			if pause == false {
				fmt.Printf("CPU is currently running\n")
				continue
			}
			stepCount = 1
			restart <- ""
		case line == "registers":
			if pause == false {
				fmt.Printf("CPU is currently running\n")
				continue
			}
			restart <- "registers"

		case strings.HasPrefix(line, "bp "):
			a := strings.Split(line[3:], " ")
			if len(a) != 2 {
				fmt.Printf("bp [set address][del address]\n")
				continue
			}
			x, err := parseUint(a[1], 16)
			if err != nil {
				fmt.Printf("invalid address: %v\n", err)
				continue
			}
			switch a[0] {
			case "set":
				z.SetBreakPoint(uint16(x), nil)
			case "del":
				z.DelBreakPoint(uint16(x))
			default:
				fmt.Printf("bp [set address][del address]\n")
				continue
			}
		case line == "bp":
			bps := z.GetBreakPoints()
			fmt.Printf("Breakpoints:\n")
			for _, bp := range bps {
				fmt.Printf("$%04x\n", bp)
			}
		default:
			fmt.Printf("invalid command %v\n", line)
		}
		last = line
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
