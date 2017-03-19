package bus

import (
	"errors"
	"fmt"

	"github.com/marcopeereboom/toyz80/device"
)

const (
	MemoryMax   = 65536
	MemoryUnit  = 1024
	MemoryShift = 10
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
)

type BusDeviceType int

const (
	DeviceInvalid = iota
	DeviceRAM
	DeviceROM
	DeviceSimpleConsole
)

// Bus glues the memory map and devices.
type Bus struct {
	memoryFlags []byte           // Memory attributes array
	memory      []byte           // Memory space
	ioL         []*device.Device // I/O lookup array
	io          []byte           // I/O space
}

type Device struct {
	Name  string // User definable name
	Start uint16
	Size  int
	Type  BusDeviceType
	Image []byte
}

func New(devices []Device) (*Bus, error) {
	// Hardcode memory and I/O space sizes for now.
	bus := &Bus{
		memory:      make([]byte, MemoryMax),
		memoryFlags: make([]byte, MemoryMax/MemoryUnit),
		io:          make([]byte, 256),
	}

	for _, d := range devices {
		// Make sure we don't overlap memory regions.
		switch d.Type {
		case DeviceRAM, DeviceROM:
			err := bus.newMemoryRegion(d)
			if err != nil {
				return nil, err
			}
		case DeviceSimpleConsole:
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
