package packets

import (
	"encoding/json"
	"log"
	"net"
	"sync/atomic"
)

type AtomicBool struct {
	value int32
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

type PacketListenerCallback = func(kind PacketKind, addr net.Addr, data interface{}, customData interface{})

type PacketListener struct {
	shouldListen *AtomicBool
	callbacks    map[PacketKind]PacketListenerCallback
	customData   interface{}
}

func NewPacketListener(customData interface{}) PacketListener {
	return PacketListener{
		shouldListen: &AtomicBool{
			value: 1,
		},
		callbacks:  make(map[PacketKind]PacketListenerCallback),
		customData: customData,
	}
}

func (packetListener PacketListener) ShutDown() {
	packetListener.shouldListen.Set(false)
}

func (packetListener *PacketListener) Register(kind PacketKind, callback PacketListenerCallback) {
	packetListener.callbacks[kind] = callback
}

func (packetListener *PacketListener) Listen(conn *net.UDPConn) {
	for packetListener.shouldListen.Get() {
		addr, buffer := ReceivePacketWithAddr(conn)
		kind := PacketKindFromBytes(buffer)
		callback, ok := packetListener.callbacks[kind]
		if !ok {
			log.Fatalf("Kind: %v not defined in callbacks\n", kind)
		}

		switch kind {
		case Hello:
			callCallback[HelloPacketData](packetListener, addr, buffer, callback, kind)
			break
		case Welcome:
			callCallback[WelcomePacketData](packetListener, addr, buffer, callback, kind)
			break
		case PlayerUpdate:
			callCallback[PlayerUpdatePacketData](packetListener, addr, buffer, callback, kind)
			break
		case ServerUpdate:
			callCallback[ServerUpdatePacketData](packetListener, addr, buffer, callback, kind)
			break
		case Bye:
			callCallback[ByePacketData](packetListener, addr, buffer, callback, kind)
			break
		case ByeAck:
			callCallback[ByeAckPacketData](packetListener, addr, buffer, callback, kind)
			break
		}
	}
}

func callCallback[T PacketData](packetListener *PacketListener, addr net.Addr, buffer []byte, callback PacketListenerCallback, kind PacketKind) {
	var packet Packet[T]
	unmarshalPacket(buffer, &packet)
	packetData := packet.Data
	callback(kind, addr, packetData, packetListener.customData)
}

func unmarshalPacket[T PacketData](buffer []byte, packet *Packet[T]) {
	err := json.Unmarshal(buffer, &packet)
	if err != nil {
		log.Fatalln("Error when deserializing packet")
	}
}
