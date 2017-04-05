package bus

import (
	"crypto/rand"
	"testing"
)

func fakeBus(start uint16, size int, image []byte) (*Bus, error) {
	devices := []Device{
		{
			Name:  "working memory",
			Start: start,
			Size:  size,
			Type:  DeviceRAM,
			Image: image,
		},
	}

	shutdown := make(chan string)
	b, err := New(devices, shutdown)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func TestRam(t *testing.T) {
	b, err := fakeBus(0x0000, 0x10000, []byte{})
	if err != nil {
		t.Fatalf("%v", err)
	}

	for addr := 0x0000; addr < 0x10000; addr++ {
		b.Write(uint16(addr), byte(addr%256))
	}
	for addr := 0x0000; addr < 0x10000; addr++ {
		if x := b.Read(uint16(addr)); x != byte(addr%256) {
			t.Fatalf("memory corruption at: %04x got %02x, "+
				"expected %02x\n", uint16(addr), x,
				byte(addr%256))
		}
	}

	// test backing memory as well.
	for addr, value := range b.memory {
		if addr%256 != int(value) {
			t.Fatalf("memory corruption at: %04x got %02x, "+
				"expected %02x\n", uint16(addr), value,
				byte(addr%256))
		}
	}
}

func assertPanic(t *testing.T, name string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("%v: did not panic", name)
		}
	}()
	f()
}
func TestRamBounds(t *testing.T) {
	start := uint16(0x1000)
	size := 0x1000
	b, err := fakeBus(start, size, []byte{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	// lower bounds
	assertPanic(t, "lower bounds", func() {
		b.Read(start - 1)
	})
	b.Read(start)

	// upper bounds
	assertPanic(t, "upper bounds", func() {
		b.Read(start + uint16(size-1) + 1)
	})
	b.Read(start + uint16(size-1))
}

func TestRamImage(t *testing.T) {
	tests := []struct {
		name      string
		start     uint16
		size      int
		imageSize int
		expected  error
	}{
		{
			name:      "0 1k",
			start:     0,
			size:      1024,
			imageSize: 1024,
		},
		{
			name:      "16k 8k",
			start:     16 * 1024,
			size:      8 * 1024,
			imageSize: 8 * 1024,
		},
		{
			name:      "60k 4k",
			start:     60 * 1024,
			size:      4 * 1024,
			imageSize: 4 * 1024,
		},
		{
			name:      "1k 2k",
			start:     1024,
			size:      1024,
			imageSize: 2 * 1024,
			expected:  ErrInvalidImageSize,
		},
		{
			name:     "63k 2k",
			start:    63 * 1024,
			size:     2 * 1024,
			expected: ErrInvalidSize,
		},
	}

	for _, test := range tests {
		image := make([]byte, test.imageSize)
		_, err := rand.Read(image)
		if err != nil {
			t.Fatalf("%v: %v", test.name, err)
		}

		b, err := fakeBus(test.start, test.size, image)
		if err != test.expected {
			t.Fatalf("%v: %v", test.name, err)
		}
		for addr := test.start; int(addr) < test.size; addr++ {
			if x := b.Read(uint16(addr)); x != b.memory[addr] {
				t.Fatalf("%v: memory corruption at: %04x "+
					"got %02x, expected %02x\n",
					uint16(addr), x, b.memory[addr])
			}
		}
	}
}
