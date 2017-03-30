package console

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/marcopeereboom/toyz80/device"
)

var (
	ErrInvalidAddress = errors.New("invalid address")
	socketName        = "/tmp/toyz80.socket"
)

// Console is a i8251A serial console.  This implementation is incomplete and
// it only needs to emulate the bare necessities.
type Console struct {
	sync.Mutex

	address byte
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

	beenShutdown   bool
	shutdownReason string
	shutdownC      chan string

	socket net.Conn
}

var (
	_ device.Device = (*Console)(nil)
)

func (c *Console) Write(address, data byte) {
	switch address {
	case 0x00:
		if c.cold || c.errorFlag || !c.enableTx {
			return
		}
		buf := []byte{data}
		c.socket.Write(buf)
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

func (c *Console) Read(address byte) byte {
	if c.cold || c.errorFlag || !c.enableRx {
		return 0xff
	}

	switch address {
	case 0x00: // data
		c.Lock()
		defer c.Unlock()
		if c.data != 0xff {
			a := c.data
			c.data = 0xff
			return a
		}
		return 0xff
	case 0x01: // status
		var rv byte
		select {
		case data := <-c.dataC:
			c.Lock()
			c.data = data
			c.Unlock()
			rv = 0x03
		default:
			rv = 0x01
		}
		return rv //0x01 //| 0x02 // TXRDY | RXRDY
	default:
	}

	return 0xff
}

func (c *Console) Shutdown() {
	c.Lock()
	defer c.Unlock()

	if c.beenShutdown {
		return
	}

	c.beenShutdown = true
	c.shutdownC <- c.shutdownReason
}

func New(shutdownC chan string) (interface{}, error) {
	c := &Console{
		errorFlag: true,
		cold:      true,
		dataC:     make(chan byte, 0),
		shutdownC: shutdownC,
	}

	l, err := net.Listen("unix", socketName)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	fmt.Printf("awaiting console connection on: %v\n", socketName)
	c.socket, err = l.Accept()
	if err != nil {
		return nil, err
	}

	go func() {
		buf := make([]byte, 1)
		for {
			// exit if we were shut down.
			c.Lock()
			if c.beenShutdown {
				c.Unlock()
				return
			}
			c.Unlock()

			_, err := c.socket.Read(buf)
			if err != nil {
				c.shutdownC <- fmt.Sprintf("console: %v\n", err)
				return
			}

			c.dataC <- buf[0]
		}
	}()

	return c, nil
}
