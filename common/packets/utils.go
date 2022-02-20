package packets

import (
	"encoding/binary"
	"github.com/JanCieslak/zbijak/common"
	"log"
	"math"
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

//type PlayerUpdatePacket struct {
//	ClientId uint8
//	X, Y     float64
//}
//
//type ServerUpdatePacket struct {
//	ClientId uint8
//	X, Y     float64
//}

func ParsePlayerUpdatePacket(buffer []byte) common.RemotePlayer {
	log.Println("Received Player update packet")
	clientId := buffer[1]
	xBits := binary.BigEndian.Uint64(buffer[2:10])
	yBits := binary.BigEndian.Uint64(buffer[10:18])
	return common.RemotePlayer{
		ClientId: clientId,
		X:        float64(xBits),
		Y:        float64(yBits),
	}
}

func ParseServerUpdatePacket(buffer []byte) []common.RemotePlayer {
	playersCount := buffer[0]
	log.Println("SERVER UPDATE COUNT", playersCount)
	players := make([]common.RemotePlayer, playersCount)
	for i := 0; i < int(playersCount); i++ {
		bufferIndex := i + 1
		clientId := buffer[bufferIndex+1]
		xBits := binary.BigEndian.Uint64(buffer[bufferIndex+2 : bufferIndex+10])
		x := math.Float64frombits(xBits)
		yBits := binary.BigEndian.Uint64(buffer[bufferIndex+10 : bufferIndex+18])
		y := math.Float64frombits(yBits)
		bufferIndex += 17
		players[i] = common.RemotePlayer{
			ClientId: clientId,
			X:        x,
			Y:        y,
		}
	}
	return players
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

func SendPlayerUpdatePacket(conn *net.UDPConn, addr *net.UDPAddr, clientId uint8, x, y float64) {
	buffer := make([]byte, 17)
	buffer[0] = clientId
	binary.BigEndian.PutUint64(buffer[1:9], math.Float64bits(x))
	binary.BigEndian.PutUint64(buffer[9:17], math.Float64bits(y))
	sendPacket(conn, addr, Packet{
		Kind: PlayerUpdate,
		Data: buffer,
	})
}

func SendServerUpdatePacket(conn *net.UDPConn, addr *net.UDPAddr, players []common.RemotePlayer) {
	buffer := make([]byte, 0)
	buffer = append(buffer, uint8(len(players)))
	for _, player := range players {
		buffer = append(buffer, player.ClientId)

		xBuffer := make([]byte, 8)
		binary.BigEndian.PutUint64(xBuffer[:], math.Float64bits(player.X))

		yBuffer := make([]byte, 8)
		binary.BigEndian.PutUint64(yBuffer[:], math.Float64bits(player.Y))

		buffer = append(buffer, xBuffer...)
		buffer = append(buffer, yBuffer...)
	}
	sendPacket(conn, addr, Packet{
		Kind: ServerUpdate,
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

func ReceivePlayerUpdatePacket(client bool, conn *net.UDPConn) common.RemotePlayer {
	packet := receivePacket(client, conn, 19)
	return ParsePlayerUpdatePacket(packet.Data)
}

func ReceiveServerUpdatePacket(conn *net.UDPConn) []common.RemotePlayer {
	packet := receivePacket(true, conn, 256)
	return ParseServerUpdatePacket(packet.Data)
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
