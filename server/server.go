package main

import (
	"encoding/json"
	"github.com/JanCieslak/zbijak/common/packets"
	"log"
	"net"
)

func main() {
	log.SetPrefix("Server - ")

	addr, err := net.ResolveUDPAddr("udp", ":8083")
	if err != nil {
		log.Fatalln("Resolve Addr error:", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Listening on:", addr)

	for {
		remoteAddr, buffer := packets.ReceivePacketWithAddr(false, conn)

		var packet packets.Packet[any]
		err = json.Unmarshal(buffer, &packet)
		if err != nil {
			log.Fatalln("Error when deserializing packet")
		}

		switch packet.Kind {
		case packets.PlayerUpdate:
			var playerUpdatePacket packets.Packet[packets.PlayerUpdateData]
			err = json.Unmarshal(buffer, &playerUpdatePacket)
			if err != nil {
				log.Fatalln("Error when deserializing packet")
			}
			playerUpdateData := playerUpdatePacket.Data

			players := make([]packets.PlayerData, 0)
			players = append(players, packets.PlayerData{
				ClientId: playerUpdateData.ClientId,
				X:        playerUpdateData.X,
				Y:        playerUpdateData.Y,
			})
			var serverUpdatePacket packets.Packet[packets.ServerUpdateData]
			serverUpdatePacket.Kind = packets.ServerUpdate
			serverUpdatePacket.Data = packets.ServerUpdateData{
				PlayersData: players,
			}
			packets.SendPacket(conn, remoteAddr, packets.Serialize(serverUpdatePacket))

			break
		default:
			log.Fatalln("Something went wrong")
		}
	}
}
