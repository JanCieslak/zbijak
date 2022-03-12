package main

import (
	"encoding/json"
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/JanCieslak/zbijak/common/vector"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	players      sync.Map
	nextClientId uint32
}

type RemotePlayer struct {
	addr   net.Addr
	pos    vector.Vec2
	inDash bool
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
		players:      sync.Map{},
		nextClientId: 0,
	}

	go func() {
		tickTime := time.Second / constants.ServerTickRate

		for {
			start := time.Now()
			players := map[uint8]packets.PlayerData{}

			s.players.Range(func(key, value any) bool {
				clientId := key.(uint8)
				player := value.(*RemotePlayer)

				players[clientId] = packets.PlayerData{
					ClientId: clientId,
					Pos:      player.pos,
					InDash:   player.inDash,
				}

				return true
			})

			if len(players) > 0 {
				timeStamp := time.Now()

				s.players.Range(func(key, value any) bool {
					player := value.(*RemotePlayer)
					log.Println("Sending server update with players:", players)
					packets.SendPacketTo(packetConn, player.addr, packets.ServerUpdate, packets.ServerUpdateData{
						PlayersData: players,
						Timestamp:   timeStamp,
					})
					return true
				})
			}

			if time.Since(start) < tickTime {
				time.Sleep(tickTime - time.Since(start))
			}
		}
	}()

	for {
		remoteAddr, buffer := packets.ReceivePacketWithAddr(packetConn)

		packetKind := packets.PacketKindFromBytes(buffer)
		log.Println("Received packet of type:", packetKind)

		switch packetKind {
		case packets.Hello:
			packets.SendPacketTo(packetConn, remoteAddr, packets.Welcome, packets.WelcomePacketData{
				ClientId: uint8(s.nextClientId),
			})
			atomic.AddUint32(&s.nextClientId, 1)
			break
		case packets.PlayerUpdate:
			var playerUpdatePacket packets.Packet[packets.PlayerUpdateData]
			err = json.Unmarshal(buffer, &playerUpdatePacket)
			if err != nil {
				log.Fatalln("Error when deserializing packet")
			}
			playerUpdateData := playerUpdatePacket.Data

			s.players.Store(playerUpdateData.ClientId, &RemotePlayer{
				addr:   remoteAddr,
				pos:    playerUpdateData.Pos,
				inDash: playerUpdateData.InDash,
			})
			break
		default:
			log.Fatalln("Something went wrong")
		}
	}
}
