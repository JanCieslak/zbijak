package packets

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
	"time"
)

type PacketKind uint8

const (
	Hello PacketKind = iota
	Welcome
	PlayerUpdate
	ServerUpdate
)

type HelloPacketData struct{}

type WelcomePacketData struct {
	ClientId uint8
}

type PlayerUpdateData struct {
	ClientId uint8
	X, Y     float64
	InDash   bool
}

type PlayerData struct {
	ClientId uint8
	X, Y     float64
	InDash   bool
}

type ServerUpdateData struct {
	PlayersData map[uint8]PlayerData
	Timestamp   time.Time
}

type PacketData interface {
	HelloPacketData | WelcomePacketData | PlayerUpdateData | ServerUpdateData
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
