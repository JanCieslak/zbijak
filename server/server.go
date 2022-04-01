package main

import (
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/netman"
	"github.com/JanCieslak/zbijak/common/vec"
	"net"
	"sync"
	"time"
)

type RemotePlayer struct {
	clientId uint8
	team     constants.Team
	name     string
	addr     net.Addr
	pos      vec.Vec2
	rotation float64
	inDash   bool
}

type RemoteBall struct {
	id      uint8
	team    constants.Team
	pos     vec.Vec2
	vel     vec.Vec2
	ownerId uint8
}

type Server struct {
	players      sync.Map
	nextClientId uint32
	nextTeam     constants.Team
	conn         *net.UDPConn
	balls        sync.Map
}

func (s *Server) Update() {
	for {
		start := time.Now()

		s.checkCollisions()
		s.moveBalls()
		s.sendServerUpdate()

		if time.Since(start) < constants.TickTime {
			time.Sleep(constants.TickTime - time.Since(start))
		}
	}
}

func (s *Server) checkCollisions() {
	s.players.Range(func(key, value any) bool {
		remotePlayer := value.(*RemotePlayer)

		// TODO There's a bug where you can pick up a ball that somebody is holding

		// Picking up balls
		s.balls.Range(func(key, value any) bool {
			ball := value.(*RemoteBall)
			if remotePlayer.pos.IsWithinRadius(ball.pos, 25) { // TODO Hardcoded
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
					ball.team = remotePlayer.team
				}
			}
			return true
		})

		return true
	})

	// Ball wall collisions
	s.balls.Range(func(key, value any) bool {
		remoteBall := value.(*RemoteBall)

		if remoteBall.pos.Y <= 0 || remoteBall.pos.Y+constants.BallRadius >= constants.ScreenHeight {
			remoteBall.vel.Y *= -1
			remoteBall.team = constants.NoTeam
		}
		if remoteBall.pos.X <= 0 || remoteBall.pos.X+constants.BallRadius >= constants.ScreenWidth {
			remoteBall.vel.X *= -1
			remoteBall.team = constants.NoTeam
		}

		return true
	})
}

func (s *Server) moveBalls() {
	s.balls.Range(func(key, value any) bool {
		remoteBall := value.(*RemoteBall)
		if remoteBall.ownerId == 255 {
			remoteBall.pos.Add(remoteBall.vel.X, remoteBall.vel.Y)
		}
		return true
	})
}

func (s *Server) sendServerUpdate() {
	players := map[uint8]netman.PlayerData{}
	s.players.Range(func(key, value any) bool {
		clientId := key.(uint8)
		player := value.(*RemotePlayer)

		players[clientId] = netman.PlayerData{
			ClientId: clientId,
			Team:     player.team,
			Name:     player.name,
			Pos:      player.pos,
			Rotation: player.rotation,
			InDash:   player.inDash,
		}

		return true
	})

	if len(players) > 0 {
		timeStamp := time.Now()

		ballsData := make([]netman.BallData, 0)

		s.balls.Range(func(key, value any) bool {
			ball := value.(*RemoteBall)
			ballsData = append(ballsData, netman.BallData{
				Id:    ball.id,
				Owner: ball.ownerId,
				Pos:   ball.pos,
			})
			return true
		})

		s.players.Range(func(key, value any) bool {
			player := value.(*RemotePlayer)

			netman.SendToUnreliable(player.addr, netman.ServerUpdate, netman.ServerUpdatePacketData{
				PlayersData: players,
				Balls:       ballsData,
				Timestamp:   timeStamp,
			})

			return true
		})
	}
}
