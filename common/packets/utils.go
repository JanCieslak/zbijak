package packets

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
)

func ReceivePacketWithAddr(conn net.PacketConn) (net.Addr, []byte) {
	buffer := make([]byte, 512)

	//log.Printf("[%v] Receiving packet...", conn.LocalAddr())
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

	//log.Printf("[%v] Receiving packet...", conn.LocalAddr())
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

func Lerp(start, end, p float64) float64 {
	return start + (end-start)*p
}

//func SLerp(start, end vec.Vec2, p float64) vec.Vec2 {
//	angle := math.Acos(start.Dot(end))
//	one := math.Sin((1-p)*angle) / math.Sin(angle)
//	two := math.Sin(p*angle) / math.Sin(angle)
//	return start.MulRet(one).AddVecRet(end.MulRet(two))
//}
