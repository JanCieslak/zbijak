package packets

import (
	"encoding/binary"
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

type Packet struct {
	Kind PacketKind
	Data []byte
}

type WelcomePacket struct {
	ClientId uint8
	Addr     *net.UDPAddr
}

type PlayerUpdatePacket struct {
	X, Y float64
}

func SendHelloPacket(conn *net.UDPConn, addr *net.UDPAddr) {
	sendPacket(conn, addr, Packet{
		Kind: Hello,
		Data: []byte{},
	})
}

func SendWelcomePacket(conn *net.UDPConn, addr *net.UDPAddr, clientSocketAddr string, clientId uint8) {
	buffer := []byte{clientId, byte(len(clientSocketAddr))}
	buffer = append(buffer, []byte(clientSocketAddr)...)
	sendPacket(conn, addr, Packet{
		Kind: Welcome,
		Data: buffer,
	})
}

func SendPlayerUpdatePacket(conn *net.UDPConn, addr *net.UDPAddr, x, y float64) {
	buffer := make([]byte, 16)
	binary.BigEndian.PutUint64(buffer, uint64(x))
	binary.BigEndian.PutUint64(buffer, uint64(y))
	sendPacket(conn, addr, Packet{
		Kind: PlayerUpdate,
		Data: buffer,
	})
}

func sendPacket(conn *net.UDPConn, addr *net.UDPAddr, packet Packet) {
	buffer := []byte{byte(packet.Kind)}
	if len(packet.Data) > 0 {
		buffer = append(buffer, packet.Data...)
	}

	log.Printf("Sending packet of kind %d sent to %v, buffer %v", packet.Kind, addr, buffer)
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
	log.Printf("Packet of kind %d sent to %v", packet.Kind, addr)
}

func ReceiveWelcomePacket(client bool, conn *net.UDPConn) WelcomePacket {
	packet := receivePacket(client, conn, 48)
	clientId := packet.Data[0]
	addrLen := packet.Data[1]
	log.Printf("Packet: %v, clientId: %d, addrLen: %d, addr: (%s)", packet, clientId, addrLen, string(packet.Data[2:addrLen+3]))
	addr, err := net.ResolveUDPAddr("udp", string(packet.Data[2:addrLen+3]))
	if err != nil {
		log.Fatalln("Resolve addr error:", err)
	}
	return WelcomePacket{
		ClientId: clientId,
		Addr:     addr,
	}
}

func ReceivePlayerUpdatePacket(client bool, conn *net.UDPConn) PlayerUpdatePacket {
	packet := receivePacket(client, conn, 18)
	xBits := binary.BigEndian.Uint64(packet.Data[1:9])
	yBits := binary.BigEndian.Uint64(packet.Data[9:17])
	return PlayerUpdatePacket{
		X: float64(xBits),
		Y: float64(yBits),
	}
}

func receivePacket(client bool, conn *net.UDPConn, bufSize uint) Packet {
	buffer := make([]byte, bufSize)

	log.Printf("[%v] Receiving packet...", conn.LocalAddr())
	if client {
		_, err := conn.Read(buffer)
		log.Println("Buffer:", buffer)
		if err != nil {
			log.Fatalln("Receive Packet Client error:", err)
		}
	} else {
		_, _, err := conn.ReadFromUDP(buffer)
		log.Println("Buffer:", buffer)
		if err != nil {
			log.Fatalln("Receive Packet Server error:", err)
		}
	}
	log.Printf("Packet of kind %d received, buffer %v", PacketKind(buffer[0]), buffer)

	return Packet{
		Kind: PacketKind(buffer[0]),
		Data: buffer[1:],
	}
}
