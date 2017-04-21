package z80

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/marcopeereboom/toyz80/bus"
)

func newZ80(imageName string) (*z80, chan string, error) {
	// Just a bunch of RAM.
	devices := []bus.Device{
		{
			Type:  bus.DeviceRAM,
			Name:  "RAM",
			Start: 0x0000,
			Size:  64 * 1024,
		},
	}

	shutdown := make(chan string)
	bus, err := bus.New(devices, shutdown)
	if err != nil {
		return nil, nil, err
	}

	z, err := New(ModeZ80, bus)
	if err != nil {
		return nil, nil, err
	}

	image, err := ioutil.ReadFile(imageName)
	if err != nil {
		return nil, nil, err
	}

	err = bus.WriteMemory(0x100, image)
	if err != nil {
		return nil, nil, err
	}

	// Patch memory, halt when 0x0000 is called.
	z.bus.WriteMemory(0, []byte{0x76 /* halt */})

	// Patch memory, print string in callback and return to caller.
	z.bus.WriteMemory(5, []byte{0xc9 /* ret */})

	return z, shutdown, nil
}

func (z *z80) printChar() error {
	// Emulate CP/M call 5; function is in register C.
	//
	// Function 2: print char in register E
	// Function 9: print $ terminated string pointer in DE
	switch byte(z.bc) {
	case 2:
		fmt.Printf("%c", byte(z.de))
	case 9:
		// Just panic if we overflow
		for addr := z.de; ; addr++ {
			ch := z.bus.Read(addr)
			if ch == '$' {
				break
			}
			fmt.Printf("%c", ch)
		}
	}

	return nil
}

func TestZexDoc(t *testing.T) {
	z, _, err := newZ80("zex/zexdoc.com")
	if err != nil {
		t.Fatal(err)
	}

	z.SetBreakPoint(0x5, z.printChar)
	z.SetPC(0x100)
	for {
		err = z.Step()
		if err != nil {
			switch bp := err.(type) {
			case BreakpointError:
				err2 := bp.Callback()
				if err2 != nil {
					t.Fatal(err2)
				}
			case HaltError:
				return
			default:
				t.Fatal(err)
			}
		}
	}
}
