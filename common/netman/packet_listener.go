package netman

import (
	"encoding/json"
	"log"
	"net"
)

type packetListenerCallback = func(kind PacketKind, addr net.Addr, data interface{}, customData interface{})

type packetListener struct {
	shouldListen *AtomicBool
	callbacks    map[PacketKind]packetListenerCallback
	customData   interface{}
}

func NewPacketListener(customData interface{}) packetListener {
	return packetListener{
		shouldListen: &AtomicBool{
			value: 1,
		},
		callbacks:  make(map[PacketKind]packetListenerCallback),
		customData: customData,
	}
}

func (packetListener *packetListener) Register(kind PacketKind, callback packetListenerCallback) {
	packetListener.callbacks[kind] = callback
}

func (packetListener *packetListener) Listen() {
	for packetListener.shouldListen.Get() {
		buffer, addr := ReceiveBytesFromUnreliable()
		kind := packetKindFromBytes(buffer)
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
		case Fire:
			callCallback[FirePacketData](packetListener, addr, buffer, callback, kind)
			break
		case Bye:
			callCallback[ByePacketData](packetListener, addr, buffer, callback, kind)
			break
		case ByeAck:
			callCallback[ByeAckPacketData](packetListener, addr, buffer, callback, kind)
			break
		default:
			log.Fatalln("Should define switch branch")
		}
	}
}

func (packetListener packetListener) ShutDown() {
	packetListener.shouldListen.Set(false)
}

func callCallback[T PacketData](packetListener *packetListener, addr net.Addr, buffer []byte, callback packetListenerCallback, kind PacketKind) {
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
