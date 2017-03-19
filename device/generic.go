package device

type Device interface {
	Write(uint16, byte)   // Write single byte to address
	Read(uint16) byte     // Read single byte to address
	DirectAccess() []byte // Returns a zero based pointer to backing memory
	// Interrupt() bool      // Interrupt signal
	// Ack() bool            // Acknowledge interrupt
}
