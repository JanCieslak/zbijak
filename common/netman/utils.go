package netman

import (
	"encoding/json"
	"log"
	"sync/atomic"
)

type AtomicBool struct {
	value int32
}

func NewAtomicBool(value bool) *AtomicBool {
	atomicBool := AtomicBool{}
	atomicBool.Set(value)
	return &atomicBool
}

func (b *AtomicBool) Set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}
	atomic.StoreInt32(&b.value, i)
}

func (b *AtomicBool) Get() bool {
	if atomic.LoadInt32(&b.value) != 0 {
		return true
	}
	return false
}

// TODO Abstract to insert better serializer
func serialize(packet any) []byte {
	data, err := json.Marshal(packet)
	if err != nil {
		log.Fatalln("Error when serializing packet:", packet)
	}
	return data
}

func deserialize[T PacketData](bytes []byte, packet *Packet[T]) {
	err := json.Unmarshal(bytes, packet)
	if err != nil {
		log.Fatalln("Error when deserializing packet")
	}
}

func packetKindFromBytes(bytes []byte) PacketKind {
	type AnyKindPacket struct {
		Kind PacketKind
		Data any
	}
	var packet AnyKindPacket
	err := json.Unmarshal(bytes, &packet)
	if err != nil {
		log.Fatalln("Error when deserializing packet")
	}
	return packet.Kind
}
