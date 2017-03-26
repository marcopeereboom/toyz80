package dummy

import (
	"errors"

	"github.com/marcopeereboom/toyz80/device"
)

var (
	ErrInvalidAddress = errors.New("invalid address")
)

// Dummy is a dummy device for testing.
type Dummy struct {
	last byte
}

var (
	_ device.Device = (*Dummy)(nil)
)

func (d *Dummy) Write(address, data byte) {
	d.last = data
}

func (d *Dummy) Read(address byte) byte {
	return d.last
}

func (d *Dummy) Shutdown() {
}

func New() (interface{}, error) {
	return &Dummy{last: 0xff}, nil
}
