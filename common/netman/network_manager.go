package netman

import (
	"encoding/binary"
	"log"
	"net"
)

const (
	bufferSize = 1024
)

var (
	udpConn     *net.UDPConn
	tcpListener *net.TCPListener
	tcpConn     *net.TCPConn
	listener    packetListener
)

// TODO should be initialized on client / server side
func InitializeServerSockets(udpAddr string, tcpAddr string, listenerCustomData interface{}) {
	udpServerAddr, err := net.ResolveUDPAddr("udp", udpAddr)
	if err != nil {
		log.Fatalln("Udp address:", err)
	}

	udpConn, err = net.ListenUDP("udp", udpServerAddr)
	if err != nil {
		log.Fatalln("UDP dial creation:", err)
	}

	tcpServerAddr, err := net.ResolveTCPAddr("tcp", tcpAddr)
	if err != nil {
		log.Fatalln("TCP address:", err)
	}

	tcpListener, err = net.ListenTCP("tcp", tcpServerAddr)
	if err != nil {
		log.Fatalln("TCP listener creation:", err)
	}

	listener = NewPacketListener(listenerCustomData)
}

// TODO should be initialized on client / server side
func InitializeClientSockets(udpAddr string, tcpAddr string, listenerCustomData interface{}) {
	udpClientAddr, err := net.ResolveUDPAddr("udp", udpAddr)
	if err != nil {
		log.Fatalln("Udp address:", err)
	}

	udpConn, err = net.DialUDP("udp", nil, udpClientAddr)
	if err != nil {
		log.Fatalln("UDP Dial creation:", err)
	}

	tcpClientAddr, err := net.ResolveTCPAddr("tcp", tcpAddr)
	if err != nil {
		log.Fatalln("TCP address:", err)
	}

	tcpConn, err = net.DialTCP("tcp", nil, tcpClientAddr)
	if err != nil {
		log.Fatalln("TCP dial creation:", err)
	}

	listener = NewPacketListener(listenerCustomData)
}

// TODO Should packet_listener and network_manager be merged ?
func RegisterUDP(kind PacketKind, callback packetListenerUDPCallback) {
	listener.registerUDP(kind, callback)
}

func RegisterTCP(kind PacketKind, callback packetListenerTCPCallback) {
	listener.registerTCP(kind, callback)
}

func ListenUDP() {
	listener.listenUDP()
}

func AcceptNewTCPConnections() {
	listener.acceptNewTCPConnections()
}

func ListenTCP() {
	listener.listenTCP(tcpConn)
}

func ShutDown() {
	udpConn.Close()

	if tcpConn != nil {
		tcpConn.Close()
	}

	if tcpListener != nil {
		tcpListener.Close()
	}

	listener.shutDown()
}

func SendUnreliable[T PacketData](kind PacketKind, packetData T) {
	send(kind, packetData, udpConn)
}

func SendToUnreliable[T PacketData](to net.Addr, kind PacketKind, packetData T) {
	_, err := udpConn.WriteTo(prepareRequestBuffer(kind, packetData), to)
	if err != nil {
		log.Fatalln("Sending packet error:", err)
	}
}

func ReceiveUnreliable[T PacketData](packet *Packet[T]) {
	receive(udpConn, packet)
}

func ReceiveBytesWithAddrUnreliable() ([]byte, net.Addr) {
	buffer := make([]byte, bufferSize)

	_, addr, err := udpConn.ReadFrom(buffer)
	if err != nil {
		log.Fatalln("Error when reading packet:", err)
	}

	dataLen := binary.BigEndian.Uint16(buffer[0:])
	return buffer[2 : dataLen+2], addr
}

func SendReliable[T PacketData](kind PacketKind, packetData T) {
	log.Println("Sending Reliable data", kind, packetData)
	send(kind, packetData, tcpConn)
}

func SendReliableWithConn[T PacketData](conn *net.TCPConn, kind PacketKind, packetData T) {
	log.Println("Sending Reliable data", kind, packetData)
	send(kind, packetData, conn)
}

func ReceiveReliable[T PacketData](packet *Packet[T]) {
	receive(tcpConn, packet)
}

func ReceiveBytesReliable(conn *net.TCPConn) []byte {
	buffer := make([]byte, bufferSize)

	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatalln("Error when reading packet:", err)
	}

	dataLen := binary.BigEndian.Uint16(buffer[0:])
	return buffer[2 : dataLen+2]
}

func send[T PacketData](kind PacketKind, packetData T, conn net.Conn) {
	_, err := conn.Write(prepareRequestBuffer(kind, packetData))
	if err != nil {
		log.Fatalln("Sending packet error:", err)
	}
}

func receive[T PacketData](conn net.Conn, packet *Packet[T]) {
	buffer := make([]byte, bufferSize)

	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatalln("Receive Packet Client error:", err)
	}

	dataLen := binary.BigEndian.Uint16(buffer[0:])
	bytes := buffer[2 : dataLen+2]

	deserialize(bytes, packet)
}

func prepareRequestBuffer[T PacketData](kind PacketKind, packetData T) []byte {
	var packet Packet[T]
	packet.Kind = kind
	packet.Data = packetData
	data := serialize(packet)

	buffer := make([]byte, 2)
	dataLen := uint16(len(data))
	binary.BigEndian.PutUint16(buffer, dataLen)

	return append(buffer, data...)
}

func BroadcastReliable[T PacketData](kind PacketKind, packetData T) {
	listener.tcpClientsMutex.Lock()
	for _, conn := range listener.tcpClients {
		SendReliableWithConn(conn, kind, packetData)
	}
	listener.tcpClientsMutex.Unlock()
}
