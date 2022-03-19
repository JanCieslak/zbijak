package main

import (
	"fmt"
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
	conn         *net.UDPConn
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
	serverAddress, err := net.ResolveUDPAddr("udp", ":8083")
	if err != nil {
		log.Fatalln("Udp address:", err)
	}

	conn, err := net.ListenUDP("udp", serverAddress)
	if err != nil {
		log.Fatalln("Dial creation:", err)
	}

	log.Println("Listening on: 8083")

	server := &Server{
		players:      sync.Map{},
		nextClientId: 0,
		conn:         conn,
	}

	go func() {
		tickTime := time.Second / constants.TickRate

		for {
			start := time.Now()
			players := map[uint8]packets.PlayerData{}

			server.players.Range(func(key, value any) bool {
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

				server.players.Range(func(key, value any) bool {
					player := value.(*RemotePlayer)
					log.Println("Sending server update with players:", players)
					packets.SendPacketTo(conn, player.addr, packets.ServerUpdate, packets.ServerUpdatePacketData{
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

	packetListener := packets.NewPacketListener(server)
	packetListener.Register(packets.Hello, handleHelloPacket)
	packetListener.Register(packets.PlayerUpdate, handlePlayerUpdatePacket)
	packetListener.Register(packets.Bye, handleByePacket)
	packetListener.Listen(server.conn)
}

func handleHelloPacket(kind packets.PacketKind, addr net.Addr, data interface{}, server interface{}) {
	serverData := server.(*Server)
	packets.SendPacketTo(serverData.conn, addr, packets.Welcome, packets.WelcomePacketData{
		ClientId: uint8(serverData.nextClientId),
	})
	atomic.AddUint32(&serverData.nextClientId, 1)
}

func handlePlayerUpdatePacket(kind packets.PacketKind, addr net.Addr, data interface{}, server interface{}) {
	playerUpdatePacketData := data.(packets.PlayerUpdatePacketData)
	serverData := server.(*Server)

	serverData.players.Store(playerUpdatePacketData.ClientId, &RemotePlayer{
		addr:   addr,
		pos:    playerUpdatePacketData.Pos,
		inDash: playerUpdatePacketData.InDash,
	})
}

func handleByePacket(kind packets.PacketKind, addr net.Addr, data interface{}, server interface{}) {
	byePacketData := data.(packets.ByePacketData)
	serverData := server.(*Server)

	fmt.Println("BYE", byePacketData.ClientId)
	serverData.players.Delete(byePacketData.ClientId)

	serverData.players.Range(func(key, value any) bool {
		player := value.(*RemotePlayer)
		packets.SendPacketTo(serverData.conn, player.addr, packets.ByeAck, packets.ByeAckPacketData{
			ClientId: byePacketData.ClientId,
		})
		return true
	})
}
