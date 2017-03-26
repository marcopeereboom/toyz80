package device

type Device interface {
	Write(byte, byte) // Write single byte to address
	Read(byte) byte   // Read single byte to address
	Shutdown()        // Nicely shut device down
}
