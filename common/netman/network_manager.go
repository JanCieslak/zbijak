package netman

import (
	"encoding/binary"
	"log"
	"net"
)

const (
	bufferSize = 1024
)

var udpConn *net.UDPConn

func SetDefaultDestination(addr string) {
	serverAddress, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalln("Udp address:", err)
	}

	udpConn, err = net.DialUDP("udp", nil, serverAddress)
	if err != nil {
		log.Fatalln("Dial creation:", err)
	}
}

func SendUnreliable[T PacketData](kind PacketKind, packetData T) {
	var packet Packet[T]
	packet.Kind = kind
	packet.Data = packetData
	data := serialize(packet)

	buffer := make([]byte, 2)
	dataLen := uint16(len(data))
	binary.BigEndian.PutUint16(buffer, dataLen)

	buffer = append(buffer, data...)

	_, err := udpConn.Write(buffer)
	if err != nil {
		log.Fatalln("Sending packet error:", err)
	}
}

func ReceiveUnreliable[T PacketData](packet *Packet[T]) {
	buffer := make([]byte, bufferSize)

	_, err := udpConn.Read(buffer)
	if err != nil {
		log.Fatalln("Receive Packet Client error:", err)
	}

	dataLen := binary.BigEndian.Uint16(buffer[0:])
	bytes := buffer[2 : dataLen+2]

	deserialize(bytes, packet)
}

func ReceiveBytesFromUnreliable() ([]byte, net.Addr) {
	buffer := make([]byte, 1024)

	//log.Printf("[%v] Receiving packet...", conn.LocalAddr())
	_, addr, err := udpConn.ReadFrom(buffer)
	if err != nil {
		log.Fatalln("Error when reading packet:", err)
	}

	dataLen := binary.BigEndian.Uint16(buffer[0:])
	return buffer[2 : dataLen+2], addr
}
