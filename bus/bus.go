package bus

import (
	"errors"
	"fmt"

	"github.com/marcopeereboom/toyz80/device"
	"github.com/marcopeereboom/toyz80/device/console"
	"github.com/marcopeereboom/toyz80/device/dummy"
)

const (
	MemoryMax   = 65536
	MemoryUnit  = 1024
	MemoryShift = 10
	IOMax       = 255
)

const (
	memoryFlagRead  = 1 << 1
	memoryFlagWrite = 1 << 2
)

var (
	ErrInvalidDeviceType = errors.New("invalid device type")
	ErrInvalidUnit       = errors.New("invalid memory unit")
	ErrInvalidSize       = errors.New("invalid memory size")
	ErrInvalidMemoryType = errors.New("invalid memory type")
	ErrInvalidImageSize  = errors.New("invalid image size")
)

type BusDeviceType int

const (
	DeviceInvalid = iota
	DeviceRAM
	DeviceROM
	DeviceSerialConsole
	DeviceDummy
)

// Bus glues the memory map and devices.
type Bus struct {
	memoryFlags []byte        // Memory attributes array
	memory      []byte        // Memory space
	io          []interface{} // I/O device lookup array
	ioStart     []byte        // I/O device start location
}

type Device struct {
	Name  string // User definable name
	Start uint16
	Size  int
	Type  BusDeviceType
	Image []byte
}

func New(devices []Device, shutdown chan string) (*Bus, error) {
	// Hardcode memory and I/O space sizes for now.
	bus := &Bus{
		memory:      make([]byte, MemoryMax),
		memoryFlags: make([]byte, MemoryMax/MemoryUnit),
		io:          make([]interface{}, IOMax),
		ioStart:     make([]byte, IOMax),
	}

	for _, d := range devices {
		// Make sure we don't overlap memory regions.
		switch d.Type {
		case DeviceRAM, DeviceROM:
			err := bus.newMemoryRegion(d)
			if err != nil {
				return nil, err
			}
		case DeviceSerialConsole:
			// Console device uses 2 ports
			if int(d.Start)+d.Size+2 > IOMax {
				return nil, ErrInvalidSize
			}
			cons, err := console.New(shutdown)
			if err != nil {
				return nil, err
			}
			// XXX rethink this
			bus.io[d.Start] = cons
			bus.io[d.Start+1] = cons
			bus.ioStart[d.Start] = byte(d.Start)
			bus.ioStart[d.Start+1] = byte(d.Start)
		case DeviceDummy:
			dummyDev, err := dummy.New()
			if err != nil {
				return nil, err
			}
			bus.io[d.Start] = dummyDev
			bus.ioStart[d.Start] = byte(d.Start)
		default:
			return nil, ErrInvalidDeviceType
		}
	}

	return bus, nil
}

func (b *Bus) newMemoryRegion(d Device) error {
	// Make sure we have a proper sized unit
	if d.Size%MemoryUnit != 0 {
		return ErrInvalidUnit
	}

	if int(d.Start)+d.Size > MemoryMax {
		return ErrInvalidSize
	}

	if len(d.Image) > d.Size {
		return ErrInvalidImageSize
	}

	var flags byte
	switch d.Type {
	case DeviceRAM:
		flags = memoryFlagRead | memoryFlagWrite
	case DeviceROM:
		flags = memoryFlagRead
	default:
		return ErrInvalidMemoryType
	}
	a := d.Start >> MemoryShift
	c := uint16(d.Size / MemoryUnit)
	for i := a; i < a+c; i++ {
		b.memoryFlags[i] = flags
	}

	// Fill memory with provided image.
	copy(b.memory[d.Start:], d.Image)

	return nil
}

func (b *Bus) Read(address uint16) byte {
	idx := address >> MemoryShift
	if b.memoryFlags[idx]&memoryFlagRead == 0 {
		panic(fmt.Sprintf("invalid read location: 0x%04x", address))
	}
	return b.memory[address]
}

func (b *Bus) Write(address uint16, data byte) {
	idx := address >> MemoryShift
	if b.memoryFlags[idx]&memoryFlagWrite == 0 {
		panic(fmt.Sprintf("invalid write location: 0x%04x", address))
	}
	b.memory[address] = data
}

func (b *Bus) IORead(address byte) byte {
	x := b.io[address].(device.Device).Read(address - b.ioStart[address])
	return x
}

func (b *Bus) IOWrite(address, data byte) {
	b.io[address].(device.Device).Write(address-b.ioStart[address], data)
}

func (b *Bus) Shutdown() {
	for i := range b.io {
		dev, ok := b.io[i].(device.Device)
		if !ok {
			continue
		}
		dev.Shutdown()
	}
}

func (b *Bus) WriteMemory(address uint16, data []byte) error {
	if int(address)+len(data) > len(b.memory) {
		return ErrInvalidImageSize
	}
	copy(b.memory[address:], data[:])
	return nil
}

// Dump returns a dump of memory starting at the provided address and length.
func (b *Bus) Dump(addr, count uint16) []byte {
	buf := make([]byte, count)
	copy(buf, b.memory[addr:addr+count])
	return buf
}
