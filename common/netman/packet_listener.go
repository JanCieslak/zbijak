package netman

import (
	"encoding/json"
	"log"
	"net"
	"sync"
)

type packetListenerUDPCallback = func(kind PacketKind, addr net.Addr, data interface{}, customData interface{})
type packetListenerTCPCallback = func(kind PacketKind, tcpConn *net.TCPConn, data interface{}, customData interface{})

type packetListener struct {
	shouldListen    *AtomicBool
	udpCallbacks    map[PacketKind]packetListenerUDPCallback
	tcpCallbacks    map[PacketKind]packetListenerTCPCallback
	tcpClients      []*net.TCPConn
	tcpClientsMutex sync.Mutex
	customData      interface{}
}

func NewPacketListener(customData interface{}) packetListener {
	return packetListener{
		shouldListen: &AtomicBool{
			value: 1,
		},
		udpCallbacks: make(map[PacketKind]packetListenerUDPCallback),
		tcpCallbacks: make(map[PacketKind]packetListenerTCPCallback),
		customData:   customData,
	}
}

func (packetListener *packetListener) registerUDP(kind PacketKind, callback packetListenerUDPCallback) {
	packetListener.udpCallbacks[kind] = callback
}

func (packetListener *packetListener) registerTCP(kind PacketKind, callback packetListenerTCPCallback) {
	packetListener.tcpCallbacks[kind] = callback
}

func (packetListener *packetListener) listenUDP() {
	for packetListener.shouldListen.Get() {
		buffer, addr := ReceiveBytesWithAddrUnreliable()
		kind := packetKindFromBytes(buffer)
		callback, ok := packetListener.udpCallbacks[kind]
		if !ok {
			log.Fatalf("Kind: %v not defined in callbacks\n", kind)
		}

		// TODO Can it be removed ? is there any way for it to be just callUDPCallback(...)
		switch kind {
		case Hello:
			callUDPCallback[HelloPacketData](packetListener, addr, buffer, callback, kind)
			break
		case Welcome:
			callUDPCallback[WelcomePacketData](packetListener, addr, buffer, callback, kind)
			break
		case PlayerUpdate:
			callUDPCallback[PlayerUpdatePacketData](packetListener, addr, buffer, callback, kind)
			break
		case ServerUpdate:
			callUDPCallback[ServerUpdatePacketData](packetListener, addr, buffer, callback, kind)
			break
		case Fire:
			callUDPCallback[FirePacketData](packetListener, addr, buffer, callback, kind)
			break
		case Bye:
			callUDPCallback[ByePacketData](packetListener, addr, buffer, callback, kind)
			break
		case ByeAck:
			callUDPCallback[ByeAckPacketData](packetListener, addr, buffer, callback, kind)
			break
		default:
			log.Fatalln("Should define switch branch")
		}
	}
}

func (packetListener *packetListener) acceptNewTCPConnections() {
	for {
		conn, err := tcpListener.AcceptTCP()
		if err != nil {
			log.Fatalln("TCP accept:", err)
		}

		packetListener.tcpClientsMutex.Lock()
		packetListener.tcpClients = append(packetListener.tcpClients, conn)
		packetListener.tcpClientsMutex.Unlock()

		go packetListener.listenTCP(conn)
	}
}

func (packetListener *packetListener) listenTCP(conn *net.TCPConn) {
	for packetListener.shouldListen.Get() {
		buffer := ReceiveBytesReliable(conn)
		kind := packetKindFromBytes(buffer)
		callback, ok := packetListener.tcpCallbacks[kind]
		if !ok {
			log.Fatalf("Kind: %v not defined in callbacks\n", kind)
		}

		switch kind {
		case Hello:
			callTCPCallback[HelloPacketData](packetListener, conn, buffer, callback, kind)
			break
		case Welcome:
			callTCPCallback[WelcomePacketData](packetListener, conn, buffer, callback, kind)
			break
		case PlayerUpdate:
			callTCPCallback[PlayerUpdatePacketData](packetListener, conn, buffer, callback, kind)
			break
		case ServerUpdate:
			callTCPCallback[ServerUpdatePacketData](packetListener, conn, buffer, callback, kind)
			break
		case Fire:
			callTCPCallback[FirePacketData](packetListener, conn, buffer, callback, kind)
			break
		case Bye:
			callTCPCallback[ByePacketData](packetListener, conn, buffer, callback, kind)
			break
		case ByeAck:
			callTCPCallback[ByeAckPacketData](packetListener, conn, buffer, callback, kind)
			break
		default:
			log.Fatalln("Should define switch branch")
		}
	}
}

func (packetListener packetListener) shutDown() {
	packetListener.shouldListen.Set(false)
}

func callUDPCallback[T PacketData](packetListener *packetListener, addr net.Addr, buffer []byte, callback packetListenerUDPCallback, kind PacketKind) {
	var packet Packet[T]
	unmarshalPacket(buffer, &packet)
	packetData := packet.Data
	callback(kind, addr, packetData, packetListener.customData)
}

func callTCPCallback[T PacketData](packetListener *packetListener, conn *net.TCPConn, buffer []byte, callback packetListenerTCPCallback, kind PacketKind) {
	var packet Packet[T]
	unmarshalPacket(buffer, &packet)
	packetData := packet.Data
	callback(kind, conn, packetData, packetListener.customData)
}

func unmarshalPacket[T PacketData](buffer []byte, packet *Packet[T]) {
	err := json.Unmarshal(buffer, &packet)
	if err != nil {
		log.Fatalln("Error when deserializing packet")
	}
}
