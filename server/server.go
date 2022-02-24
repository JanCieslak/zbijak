package main

import (
	"encoding/json"
	"github.com/JanCieslak/zbijak/common/packets"
	"log"
	"net"
	"sync"
)

type Server struct {
	players sync.Map
	player  *RemotePlayer
}

type RemotePlayer struct {
	addr *net.Addr
	x, y float64
}

func main() {
	log.SetPrefix("Server - ")

	packetConn, err := net.ListenPacket("udp", ":8083")
	if err != nil {
		log.Fatalln("PacketConn error", err)
	}

	log.Println("Listening on: 8083")

	s := Server{
		players: sync.Map{},
		player: &RemotePlayer{
			addr: nil,
			x:    0,
			y:    0,
		},
	}

	for {
		remoteAddr, buffer := packets.ReceivePacketWithAddr(packetConn)

		var packet packets.Packet[any]
		err = json.Unmarshal(buffer, &packet)
		if err != nil {
			log.Fatalln("Error when deserializing packet")
		}

		switch packet.Kind {
		case packets.PlayerUpdate:
			// Receive
			var playerUpdatePacket packets.Packet[packets.PlayerUpdateData]
			err = json.Unmarshal(buffer, &playerUpdatePacket)
			if err != nil {
				log.Fatalln("Error when deserializing packet")
			}
			playerUpdateData := playerUpdatePacket.Data

			s.player.addr = &remoteAddr
			s.player.x = playerUpdateData.X
			s.player.y = playerUpdateData.Y

			// Send
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
			packets.SendPacketTo(packetConn, remoteAddr, packets.Serialize(serverUpdatePacket))

			break
		default:
			log.Fatalln("Something went wrong")
		}
	}
}
