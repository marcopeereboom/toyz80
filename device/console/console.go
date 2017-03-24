package console

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/marcopeereboom/toyz80/device"
)

var (
	ErrInvalidAddress = errors.New("invalid address")
)

// Console is a i8251A serial console.  This implementation is incomplete and
// it only needs to emulate the bare necessities.
type Console struct {
	address byte
	status  byte
	data    byte
	dataC   chan byte
	mode    byte

	errorFlag bool
	enableTx  bool
	enableRx  bool
	// Set during cold boot. 8251A waits for "Mode Instruction" to instruct
	// speed, parity etc. See:
	// http://www.electronics.dit.ie/staff/tscarff/8251usart/8251.htm
	cold bool
}

var (
	_ device.Device = (*Console)(nil)
)

func (c *Console) Write(address, data byte) {
	switch address {
	case 0x00:
		fmt.Printf("%c", data&0x7f)
	case 0x01:
		if c.cold {
			// We are in cold boot.  Receice Mode.
			// bit 0..1 baud multiplier Xi
			//	00 not implemented
			//	01 1x
			//	02 16x 9600bps
			//	11 64x
			// bit 2..3 byte length
			//	00 5 bits
			//	04 6 bits
			//	08 7 bits
			//	0c 8 bits
			// bit 4..5 parity
			//	00 disable
			//	10 odd
			//	20 disable
			//	30 even
			// bit 6..7 stop bit length
			//	00 inhabit
			//	40 1 bit
			//	80 1.5 bits
			//	c0 2 bits
			c.mode = data
			c.cold = false
			return
		}
		// Command
		// bit 0 TXEN
		//	00 disable
		//	01 transmit enable
		// bit 1 DTR (low active)
		//	00 DTR = 1
		//	02 DTR = 0
		// bit 2 RXE
		//	00 disable
		//	04 receive enable
		// bit 3 SBRK
		//	08 send SBRK
		//	00 normal operation
		// bit 4 ER
		//	10 reset error flag
		//	00 normal operation
		// bit 5 RTS (low active)
		//	00 RTS = 1
		//	20 RTS = 0
		// bit 6 IR
		//	40 internal reset
		//	00 normal operation
		// bit 7 EH
		//	80 hunt mode
		//	00 normal operation
		if data&0x01 == 0x01 {
			c.enableTx = true
		}
		if data&0x04 == 0x04 {
			c.enableRx = true
		}
		if data&0x10 == 0x10 {
			c.errorFlag = false
		}

	default:
		panic(fmt.Sprintf("can't access address 0x%02x", address))
	}
}

// Read is not reentrant.
func (c *Console) Read(address byte) byte {
	switch address {
	case 0x00:
		//panic("console read data")
		if c.data != 0xff {
			a := c.data
			//fmt.Printf("read %02x  ", a)
			c.data = 0xff
			return a
		}
		return 0xff
	case 0x01:
		var rv byte
		select {
		case c.data = <-c.dataC:
			rv = 0x03
		default:
			rv = 0x01
		}
		//if c.data != 0xff {
		//	return 0x03
		//}
		return rv //0x01 //| 0x02 // TXRDY | RXRDY
	default:
	}

	return 0xff
}

func New() (interface{}, error) {
	// disable input buffering
	err := exec.Command("stty", "-f", "/dev/tty", "cbreak", "min", "1").Run()
	if err != nil {
		return nil, err
	}
	// do not display entered characters on the screen
	err = exec.Command("stty", "-f", "/dev/tty", "-echo").Run()
	if err != nil {
		return nil, err
	}

	defer exec.Command("stty", "-f", "/dev/tty", "echo").Run()

	c := &Console{
		errorFlag: true,
		cold:      true,
		dataC:     make(chan byte, 1),
	}
	go func() {
		var b []byte = make([]byte, 1)
		for {
			os.Stdin.Read(b)
			// see if we need to translate \r to \n
			if b[0] == 0x0a {
				b[0] = 0x0d
			}
			c.dataC <- b[0]
		}
	}()
	return c, nil
}
