package buffers

import (
	"constraints"
	"reflect"
)

type OutputBitStream struct {
	bitHead uint32
	buffer  []byte
}

func NewOutputBitStream() *OutputBitStream {
	return &OutputBitStream{
		bitHead: 0,
		buffer:  make([]byte, 32),
	}
}

func (s *OutputBitStream) Reallocate(newSize int) {
	newBuffer := make([]byte, newSize)
	copy(newBuffer, s.buffer)
	s.buffer = newBuffer
}

func (s *OutputBitStream) Write(data interface{}) {
	dataType := reflect.TypeOf(data).Kind()

	// TODO signed types ? bool type, enum types ?

	if dataType == reflect.Uint8 {
		writeBits(s, data.(uint8), 8)
	} else if dataType == reflect.Uint16 {
		writeNBytes(s, data.(uint16), 2)
	} else if dataType == reflect.Uint32 {
		writeNBytes(s, data.(uint32), 4)
	} else if dataType == reflect.Uint64 {
		writeNBytes(s, data.(uint64), 8)
	}
}

func writeNBytes[T constraints.Unsigned](stream *OutputBitStream, data T, nBytes uint8) {
	var i uint8
	for i = 0; i < nBytes; i++ {
		writeBits(stream, uint8(data>>(i*8)), 8)
	}
}

func writeBits(stream *OutputBitStream, data uint8, bitCount uint32) {
	nextBitHead := stream.bitHead + bitCount

	if nextBitHead > uint32(cap(stream.buffer)*8) {
		stream.Reallocate(2 * cap(stream.buffer))
	}

	byteOffset := stream.bitHead / 8
	bitOffset := stream.bitHead & 0b111
	var mask byte = ^(0xff << bitOffset)

	stream.buffer[byteOffset] = (stream.buffer[byteOffset] & mask) | (data << bitOffset)

	bitsUsed := 8 - bitOffset

	if bitsUsed < bitCount {
		stream.buffer[byteOffset+1] = data >> bitsUsed
	}

	stream.bitHead = nextBitHead
}
