package packets

import (
	"encoding/binary"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestSerde(t *testing.T) {
	packet := Packet[WelcomePacketData]{
		Kind: Welcome,
		Data: WelcomePacketData{
			ClientId: 123,
		},
	}
	data := Serialize(packet)

	var jsonPacket Packet[WelcomePacketData]
	err := json.Unmarshal(data, &jsonPacket)
	if err != nil {
		log.Fatalln("Error when deserializing packet")
	}

	assert.Equal(t, packet.Kind, jsonPacket.Kind)

	welcomeData := packet.Data
	jsonWelcomeData := jsonPacket.Data
	assert.Equal(t, welcomeData.ClientId, jsonWelcomeData.ClientId)
}

func TestBinary(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 0, 0, 0, 0, 0}
	dataLen := uint16(len(data))

	buffer := make([]byte, 2)
	binary.BigEndian.PutUint16(buffer, dataLen)

	assert.Equal(t, []byte{0, 15}, buffer)
}

func TestPackets(t *testing.T) {
	//addr, err := net.ResolveUDPAddr("udp", ":8083")
	//if err != nil {
	//	log.Fatalln("Resolve Addr error:", err)
	//}
	//
	//serverConn, err := net.ListenUDP("udp", addr)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//clientConn, err := net.DialUDP("udp", nil, addr)
	//if err != nil {
	//	log.Fatalln("Dial creation:", err)
	//}
	//
	//packet := Packet[WelcomePacketData]{
	//	Kind: Welcome,
	//	Data: WelcomePacketData{
	//		ClientId: 123,
	//	},
	//}
	//data := Serialize(packet)
	//Send(clientConn, nil, data)
	//
	//bytes := ReceivePacket(false, serverConn)
	//var jsonPacket Packet[WelcomePacketData]
	//err = json.Unmarshal(bytes, &jsonPacket)
	//if err != nil {
	//	log.Fatalln("Error when deserializing packet")
	//}
	//
	//assert.Equal(t, packet.Kind, jsonPacket.Kind)
	//
	//welcomeData := packet.Data
	//jsonWelcomeData := jsonPacket.Data
	//assert.Equal(t, welcomeData.ClientId, jsonWelcomeData.ClientId)
}
