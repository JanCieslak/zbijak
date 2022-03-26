package main

import (
	"fmt"
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/JanCieslak/zbijak/common/vec"
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
	balls        []RemoteBall
}

type RemotePlayer struct {
	clientId uint8
	addr     net.Addr
	pos      vec.Vec2
	rotation float64
	inDash   bool
}

type RemoteBall struct {
	id      uint8
	pos     vec.Vec2
	ownerId uint8
}

func main() {
	log.SetPrefix("Server - ")
	log.SetOutput(ioutil.Discard)

	serverAddress, err := net.ResolveUDPAddr("udp", ":8083")
	if err != nil {
		log.Fatalln("Udp address:", err)
	}

	conn, err := net.ListenUDP("udp", serverAddress)
	if err != nil {
		log.Fatalln("Dial creation:", err)
	}

	log.Println("Listening on: 8083")

	balls := make([]RemoteBall, 0)
	balls = append(balls, RemoteBall{
		id:      0,
		pos:     vec.Vec2{X: 300, Y: 300},
		ownerId: 0,
	})

	server := &Server{
		players:      sync.Map{},
		nextClientId: 0,
		conn:         conn,
		balls:        balls,
	}

	go server.Update()

	packetListener := packets.NewPacketListener(server)
	packetListener.Register(packets.Hello, handleHelloPacket)
	packetListener.Register(packets.PlayerUpdate, handlePlayerUpdatePacket)
	packetListener.Register(packets.Bye, handleByePacket)
	packetListener.Register(packets.Fire, handleFirePacket)
	packetListener.Listen(server.conn)
}

func (s *Server) Update() {
	tickTime := time.Second / constants.TickRate

	for {
		start := time.Now()

		s.CheckCollisions()
		s.SendServerUpdate()

		if time.Since(start) < tickTime {
			time.Sleep(tickTime - time.Since(start))
		}
	}
}

func (s *Server) CheckCollisions() {
	s.players.Range(func(key, value any) bool {
		remotePlayer := value.(*RemotePlayer)

		if remotePlayer.pos.AddVecRet(vec.NewVec2(16, 16)).IsWithinRadius(s.balls[0].pos, 25) {
			s.balls[0].ownerId = remotePlayer.clientId
		}

		return true
	})
}

func (s *Server) SendServerUpdate() {
	players := map[uint8]packets.PlayerData{}
	s.players.Range(func(key, value any) bool {
		clientId := key.(uint8)
		player := value.(*RemotePlayer)

		players[clientId] = packets.PlayerData{
			ClientId: clientId,
			Pos:      player.pos,
			Rotation: player.rotation,
			InDash:   player.inDash,
		}

		return true
	})

	if len(players) > 0 {
		timeStamp := time.Now()

		ballsData := make([]packets.BallData, 0)

		for _, ball := range s.balls {
			ballsData = append(ballsData, packets.BallData{
				Id:    ball.id,
				Owner: ball.ownerId,
				Pos:   ball.pos,
			})
		}

		s.players.Range(func(key, value any) bool {
			player := value.(*RemotePlayer)

			packets.SendPacketTo(s.conn, player.addr, packets.ServerUpdate, packets.ServerUpdatePacketData{
				PlayersData: players,
				Balls:       ballsData,
				Timestamp:   timeStamp,
			})

			return true
		})
	}
}

func handleHelloPacket(_ packets.PacketKind, addr net.Addr, _ interface{}, server interface{}) {
	serverData := server.(*Server)
	packets.SendPacketTo(serverData.conn, addr, packets.Welcome, packets.WelcomePacketData{
		ClientId: uint8(serverData.nextClientId),
	})
	atomic.AddUint32(&serverData.nextClientId, 1)
}

func handlePlayerUpdatePacket(_ packets.PacketKind, addr net.Addr, data interface{}, server interface{}) {
	playerUpdatePacketData := data.(packets.PlayerUpdatePacketData)
	serverData := server.(*Server)

	serverData.players.Store(playerUpdatePacketData.ClientId, &RemotePlayer{
		clientId: playerUpdatePacketData.ClientId,
		addr:     addr,
		pos:      playerUpdatePacketData.Pos,
		rotation: playerUpdatePacketData.Rotation,
		inDash:   playerUpdatePacketData.InDash,
	})
}

func handleByePacket(_ packets.PacketKind, _ net.Addr, data interface{}, server interface{}) {
	byePacketData := data.(packets.ByePacketData)
	serverData := server.(*Server)

	log.Println("Bye:", byePacketData.ClientId)
	serverData.players.Delete(byePacketData.ClientId)

	serverData.players.Range(func(key, value any) bool {
		player := value.(*RemotePlayer)
		packets.SendPacketTo(serverData.conn, player.addr, packets.ByeAck, packets.ByeAckPacketData{
			ClientId: byePacketData.ClientId,
		})
		return true
	})
}

func handleFirePacket(_ packets.PacketKind, _ net.Addr, data interface{}, server interface{}) {
	//firePacketData := data.(packets.FirePacketData)
	serverData := server.(*Server)

	serverData.balls[0].ownerId = 255 // TODO search ball by firePacketData ownerId
	fmt.Println("FIRED")
}
