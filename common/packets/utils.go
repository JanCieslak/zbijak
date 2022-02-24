package packets

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
)

type PacketKind uint8

const (
	Hello PacketKind = iota
	Welcome
	PlayerUpdate
	ServerUpdate
)

type Packet[T any] struct {
	Kind PacketKind
	Data T
}

type HelloPacketData struct{}

type WelcomePacketData struct {
	ClientId uint8
}

type PlayerUpdateData struct {
	ClientId uint8
	X, Y     float64
}

type PlayerData struct {
	ClientId uint8
	X, Y     float64
}

type ServerUpdateData struct {
	PlayersData []PlayerData
}

func SendPacket(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	buffer := make([]byte, 2)
	dataLen := uint16(len(data))
	binary.BigEndian.PutUint16(buffer, dataLen)

	buffer = append(buffer, data...)

	if addr == nil {
		_, err := conn.Write(buffer)
		if err != nil {
			log.Fatalln("Sending packet (addr) error:", err)
		}
	} else {
		_, err := conn.WriteToUDP(buffer, addr)
		if err != nil {
			log.Fatalln("Sending packet (nil) error:", err)
		}
	}
}

func ReceivePacketWithAddr(client bool, conn *net.UDPConn) (*net.UDPAddr, []byte) {
	buffer := make([]byte, 512)

	log.Printf("[%v] Receiving packet...", conn.LocalAddr())
	var addr *net.UDPAddr
	if client {
		_, err := conn.Read(buffer)
		if err != nil {
			log.Fatalln("Receive Packet Client error:", err)
		}
	} else {
		_, ra, err := conn.ReadFromUDP(buffer)
		addr = ra
		if err != nil {
			log.Fatalln("Receive Packet Server error:", err)
		}
	}

	dataLen := binary.BigEndian.Uint16(buffer[0:])
	return addr, buffer[2 : dataLen+2]
}

func ReceivePacket(client bool, conn *net.UDPConn) []byte {
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
	return buffer[2 : dataLen+2]
}

func Serialize(packet any) []byte {
	data, err := json.Marshal(packet)
	if err != nil {
		log.Fatalln("Error when serializing packet:", packet)
	}
	return data
}
