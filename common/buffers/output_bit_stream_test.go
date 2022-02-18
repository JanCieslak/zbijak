package buffers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteBitsOneByte(t *testing.T) {
	stream := NewOutputBitStream()
	var data1 uint8 = 0b1010
	var data2 uint8 = 0b1000
	writeBits(stream, data1, 4)
	writeBits(stream, data2, 4)
	assert.Equal(t, uint8(0b10001010), stream.buffer[0])
}

func TestWriteBitsTwoBytes(t *testing.T) {
	stream := NewOutputBitStream()
	var data1 uint8 = 0b1010
	var data2 uint8 = 0b100011
	writeBits(stream, data1, 4)
	writeBits(stream, data2, 6)
	assert.Equal(t, uint8(0b00111010), stream.buffer[0])
	assert.Equal(t, uint8(0b10), stream.buffer[1])
}

func TestWriteBitsZeroByte(t *testing.T) {
	stream := NewOutputBitStream()
	var data uint8 = 0
	writeBits(stream, data, 0)
	assert.Equal(t, uint32(0), stream.bitHead)
	assert.Equal(t, uint8(0), stream.buffer[0])
}

func TestWriteBitsAnyPrimitiveType(t *testing.T) {
	stream := NewOutputBitStream()
	var uint8data uint8 = 0x80
	var uint16data uint16 = 0x8000
	var uint32data uint32 = 0x80000000
	var uint64data uint64 = 0x80000000_00000000
	stream.Write(uint8data)
	stream.Write(uint16data)
	stream.Write(uint32data)
	stream.Write(uint64data)
	assert.Equal(t, uint8(128), stream.buffer[0])
	assert.Equal(t, uint8(128), stream.buffer[2])
	assert.Equal(t, uint8(128), stream.buffer[6])
	assert.Equal(t, uint8(128), stream.buffer[14])
}
