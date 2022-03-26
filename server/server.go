package main

import (
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/JanCieslak/zbijak/common/vec"
	"io/ioutil"
	"log"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	NoOwner = 255 // TODO Limit players ?
)

type Server struct {
	players      sync.Map
	nextClientId uint32
	conn         *net.UDPConn
	balls        sync.Map
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
	vel     vec.Vec2
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

	balls := sync.Map{}
	balls.Store(0, &RemoteBall{
		id:      0,
		pos:     vec.NewVec2(300, 300),
		vel:     vec.NewVec2(0, 0),
		ownerId: 0,
	})
	balls.Store(1, &RemoteBall{
		id:      1,
		pos:     vec.NewVec2(600, 300),
		vel:     vec.NewVec2(0, 0),
		ownerId: NoOwner,
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
		s.MoveBalls()
		s.SendServerUpdate()

		if time.Since(start) < tickTime {
			time.Sleep(tickTime - time.Since(start))
		}
	}
}

func (s *Server) CheckCollisions() {
	s.players.Range(func(key, value any) bool {
		remotePlayer := value.(*RemotePlayer)

		// Picking up balls
		s.balls.Range(func(key, value any) bool {
			ball := value.(*RemoteBall)
			if remotePlayer.pos.AddVecRet(vec.NewVec2(16, 16)).IsWithinRadius(ball.pos, 25) { // TODO Hardcoded
				isOwned := false
				s.balls.Range(func(key, value any) bool {
					innerBall := value.(*RemoteBall)
					if remotePlayer.clientId == innerBall.ownerId {
						isOwned = true
						return false
					}
					return true
				})
				if !isOwned {
					ball.ownerId = remotePlayer.clientId
				}
			}
			return true
		})

		return true
	})

	// Ball wall collisions
	s.balls.Range(func(key, value any) bool {
		remoteBall := value.(*RemoteBall)

		if remoteBall.pos.Y <= 0 || remoteBall.pos.Y+16 >= constants.ScreenHeight {
			remoteBall.vel.Y *= -1
		}
		if remoteBall.pos.X <= 0 || remoteBall.pos.X+16 >= constants.ScreenWidth {
			remoteBall.vel.X *= -1
		}

		return true
	})
}

func (s *Server) MoveBalls() {
	s.balls.Range(func(key, value any) bool {
		remoteBall := value.(*RemoteBall)
		if remoteBall.ownerId == 255 {
			remoteBall.pos.Add(remoteBall.vel.X, remoteBall.vel.Y)
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

		s.balls.Range(func(key, value any) bool {
			ball := value.(*RemoteBall)
			ballsData = append(ballsData, packets.BallData{
				Id:    ball.id,
				Owner: ball.ownerId,
				Pos:   ball.pos,
			})
			return true
		})

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
	firePacketData := data.(packets.FirePacketData)
	serverData := server.(*Server)

	serverData.balls.Range(func(key, value any) bool {
		ball := value.(*RemoteBall)
		if ball.ownerId == firePacketData.ClientId {
			value, ok := serverData.players.Load(firePacketData.ClientId)
			if !ok {
				log.Fatalf("Couldn't find player with given client id: %d from fire packet data\n", firePacketData.ClientId)
			}
			remotePlayer := value.(*RemotePlayer)
			newX := remotePlayer.pos.X + 16 - 7.5 + 40*math.Cos(remotePlayer.rotation) // TODO Hardcoded
			newY := remotePlayer.pos.Y + 16 - 7.5 + 40*math.Sin(remotePlayer.rotation)
			ball.pos.Set(newX, newY)

			// TODO Fix throw direction (prob problem where mouse points to where pos is [top left] - should be converted to center)
			ballPosVec := vec.NewVec2(newX, newY)
			playerPosVec := vec.NewVec2(remotePlayer.pos.X+16, remotePlayer.pos.Y+16)
			val := ballPosVec.SubVecRet(playerPosVec)
			val.Normalize()
			val.Mul(3)
			ball.vel = val

			ball.ownerId = NoOwner
		}
		return true
	})
}
