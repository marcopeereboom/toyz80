package console

import (
	"errors"
	"fmt"

	"github.com/marcopeereboom/toyz80/device"
)

var (
	ErrInvalidAddress = errors.New("invalid address")
)

const (
	WriteReady = 1 << 1
	ReadReady  = 1 << 2
)

// Console is a simple 7 bit console.
type Console struct {
	address byte
	status  byte
	data    byte
}

var (
	_ device.Device = (*Console)(nil)
)

func (c *Console) Write(address, data byte) {
	switch address {
	case 0x00:
		panic("can't write to status register")
	case 0x01:
		if c.status&WriteReady == WriteReady {
			fmt.Printf("%c", data&0x7f)
		}
	default:
		panic(fmt.Sprintf("can't access address 0x%02x", address))
	}
}

func (c *Console) Read(address byte) byte {
	if c.status&ReadReady == ReadReady {
		return c.data
	}
	return 0xff
}

func New() (interface{}, error) {
	return &Console{
		status: WriteReady,
	}, nil
}
