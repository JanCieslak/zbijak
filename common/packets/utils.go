package packets

import (
	"encoding/binary"
	"encoding/json"
	"github.com/JanCieslak/zbijak/common/vector"
	"log"
	"net"
	"sync/atomic"
	"time"
)

type PacketKind uint8

const (
	Hello PacketKind = iota
	Welcome
	PlayerUpdate
	ServerUpdate
	Bye
	ByeAck
)

type HelloPacketData struct{}

type WelcomePacketData struct {
	ClientId uint8
}

type PlayerUpdateData struct {
	ClientId uint8
	Pos      vector.Vec2
	InDash   bool
}

type PlayerData struct {
	ClientId uint8
	Pos      vector.Vec2
	InDash   bool
}

type ServerUpdateData struct {
	PlayersData map[uint8]PlayerData
	Timestamp   time.Time
}

type ByePacketData struct {
	ClientId uint8
}

type ByeAckPacketData struct {
	ClientId uint8
}

type PacketData interface {
	HelloPacketData | WelcomePacketData | PlayerUpdateData | ServerUpdateData | ByePacketData | ByeAckPacketData
}

type Packet[T PacketData] struct {
	Kind PacketKind
	Data T
}

func ReceivePacketWithAddr(conn net.PacketConn) (net.Addr, []byte) {
	buffer := make([]byte, 512)

	log.Printf("[%v] Receiving packet...", conn.LocalAddr())
	_, addr, err := conn.ReadFrom(buffer)
	if err != nil {
		log.Fatalln("Error when reading packet:", err)
	}

	dataLen := binary.BigEndian.Uint16(buffer[0:])
	return addr, buffer[2 : dataLen+2]
}

func SendPacketTo[T PacketData](conn net.PacketConn, toAddr net.Addr, packetKind PacketKind, packetData T) {
	var packet Packet[T]
	packet.Kind = packetKind
	packet.Data = packetData
	data := Serialize(packet)

	buffer := make([]byte, 2)
	dataLen := uint16(len(data))
	binary.BigEndian.PutUint16(buffer, dataLen)

	buffer = append(buffer, data...)

	_, err := conn.WriteTo(buffer, toAddr)
	if err != nil {
		log.Fatalf("Sending packet to %v error: %v", toAddr, err)
	}
}

func Send[T PacketData](conn *net.UDPConn, packetKind PacketKind, packetData T) {
	var packet Packet[T]
	packet.Kind = packetKind
	packet.Data = packetData
	data := Serialize(packet)

	buffer := make([]byte, 2)
	dataLen := uint16(len(data))
	binary.BigEndian.PutUint16(buffer, dataLen)

	buffer = append(buffer, data...)

	_, err := conn.Write(buffer)
	if err != nil {
		log.Fatalln("Sending packet error:", err)
	}
}

func ReceivePacket[T PacketData](client bool, conn *net.UDPConn, packet *Packet[T]) {
	buffer := make([]byte, 512)

	log.Printf("[%v] Receiving packet...", conn.LocalAddr())
	if client {
		_, err := conn.Read(buffer)
		if err != nil {
			log.Fatalln("Receive Packet Client error:", err)
		}
	} else {
		_, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatalln("Receive Packet Server error:", err)
		}
	}

	dataLen := binary.BigEndian.Uint16(buffer[0:])
	bytes := buffer[2 : dataLen+2]

	err := json.Unmarshal(bytes, &packet)
	if err != nil {
		log.Fatalln("Error when deserializing packet")
	}
}

func Serialize(packet any) []byte {
	data, err := json.Marshal(packet)
	if err != nil {
		log.Fatalln("Error when serializing packet:", packet)
	}
	return data
}

func PacketKindFromBytes(bytes []byte) PacketKind {
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

type PacketListenerCallback = func(kind PacketKind, data interface{}, customData interface{})

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

func (packetListener *PacketListener) Register(kind PacketKind, callback PacketListenerCallback) {
	packetListener.callbacks[kind] = callback
}

func (packetListener *PacketListener) Listen(conn *net.UDPConn) {
	for packetListener.shouldListen.Get() {
		_, buffer := ReceivePacketWithAddr(conn)
		kind := PacketKindFromBytes(buffer)
		callback, ok := packetListener.callbacks[kind]
		if !ok {
			log.Fatalf("Kind: %s not defined in callbacks\n", string(kind))
		}

		switch kind {
		case ServerUpdate:
			var packet Packet[ServerUpdateData]
			unmarshalPacket(buffer, &packet)
			packetData := packet.Data
			callback(kind, packetData, packetListener.customData)
			break
		case ByeAck:
			var packet Packet[ByeAckPacketData]
			unmarshalPacket(buffer, &packet)
			packetData := packet.Data
			callback(kind, packetData, packetListener.customData)
			break
		}
	}
}

func unmarshalPacket[T PacketData](buffer []byte, packet *Packet[T]) {
	err := json.Unmarshal(buffer, &packet)
	if err != nil {
		log.Fatalln("Error when deserializing packet")
	}
}
