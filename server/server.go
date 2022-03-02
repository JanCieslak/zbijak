package main

import (
	"encoding/json"
	"fmt"
	"github.com/JanCieslak/zbijak/common/packets"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type Server struct {
	players      sync.Map
	nextClientId uint32
}

type RemotePlayer struct {
	addr net.Addr
	x, y float64
}

func main() {
	log.SetPrefix("Server - ")
	//log.SetOutput(ioutil.Discard)

	// TODO Changge to UDPConn (it implements ListenPacket)
	packetConn, err := net.ListenPacket("udp", ":8083")
	if err != nil {
		log.Fatalln("PacketConn error", err)
	}

	log.Println("Listening on: 8083")

	s := Server{
		players:      sync.Map{},
		nextClientId: 0,
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

		fmt.Println("Packet of type ", packet.Kind)

		switch packet.Kind {
		case packets.Hello:
			var welcomePacket packets.Packet[packets.WelcomePacketData]
			welcomePacket.Kind = packets.ServerUpdate
			welcomePacket.Data = packets.WelcomePacketData{
				ClientId: uint8(s.nextClientId),
			}
			fmt.Println("Id: ", s.nextClientId)
			atomic.AddUint32(&s.nextClientId, 1)
			fmt.Println("Id after: ", s.nextClientId)
			packets.SendPacketTo(packetConn, remoteAddr, packets.Serialize(welcomePacket))
			break
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
