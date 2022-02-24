package main

import (
	"encoding/json"
	"github.com/JanCieslak/zbijak/common/packets"
	"io/ioutil"
	"log"
	"net"
	"sync"
)

type Server struct {
	players sync.Map
	//player  *RemotePlayer
}

type RemotePlayer struct {
	addr net.Addr
	x, y float64
}

func main() {
	log.SetPrefix("Server - ")
	log.SetOutput(ioutil.Discard)

	// TODO Changge to UDPConn (it implements ListenPacket)
	packetConn, err := net.ListenPacket("udp", ":8083")
	if err != nil {
		log.Fatalln("PacketConn error", err)
	}

	log.Println("Listening on: 8083")

	s := Server{
		players: sync.Map{},
	}

	go func() {
		for {
			players := make([]packets.PlayerData, 0)
			s.players.Range(func(key, value any) bool {
				player := value.(*RemotePlayer)

				players = append(players, packets.PlayerData{
					ClientId: key.(uint8),
					X:        player.x,
					Y:        player.y,
				})

				return true
			})

			if len(players) > 0 {
				s.players.Range(func(key, value any) bool {
					player := value.(*RemotePlayer)

					var serverUpdatePacket packets.Packet[packets.ServerUpdateData]
					serverUpdatePacket.Kind = packets.ServerUpdate
					serverUpdatePacket.Data = packets.ServerUpdateData{
						PlayersData: players,
					}
					log.Println("Sending server update with players:", players)
					packets.SendPacketTo(packetConn, player.addr, packets.Serialize(serverUpdatePacket))
					return true
				})
			}
		}
	}()

	for {
		remoteAddr, buffer := packets.ReceivePacketWithAddr(packetConn)

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

			s.players.Store(playerUpdateData.ClientId, &RemotePlayer{
				addr: remoteAddr,
				x:    playerUpdateData.X,
				y:    playerUpdateData.Y,
			})
			break
		default:
			log.Fatalln("Something went wrong")
		}
	}
}
