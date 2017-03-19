package device

type Device interface {
	Write(byte, byte) // Write single byte to address
	Read(byte) byte   // Read single byte to address
	// DirectAccess() []byte // Returns a zero based pointer to backing memory
	// Interrupt() bool      // Interrupt signal
	// Ack() bool            // Acknowledge interrupt
}
