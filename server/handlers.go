package main

import (
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/netman"
	"github.com/JanCieslak/zbijak/common/vec"
	"log"
	"math"
	"net"
	"sync/atomic"
)

func handleHelloPacket(_ netman.PacketKind, conn *net.TCPConn, _ interface{}, server interface{}) {
	serverData := server.(*Server)

	netman.SendReliableWithConn(conn, netman.Welcome, netman.WelcomePacketData{
		ClientId: uint8(serverData.nextClientId),
		Team:     serverData.nextTeam,
	})

	atomic.AddUint32(&serverData.nextClientId, 1)

	if serverData.nextTeam == constants.TeamB {
		serverData.nextTeam = constants.TeamA
	} else {
		serverData.nextTeam = constants.TeamB
	}

	// TODO Registering should be happening here ? (right now, it's in handlePlayerUpdatePacket)
}

func handlePlayerUpdatePacket(_ netman.PacketKind, addr net.Addr, data interface{}, server interface{}) {
	playerUpdatePacketData := data.(netman.PlayerUpdatePacketData)
	serverData := server.(*Server)

	serverData.players.Store(playerUpdatePacketData.ClientId, &RemotePlayer{
		clientId: playerUpdatePacketData.ClientId,
		team:     playerUpdatePacketData.Team,
		name:     playerUpdatePacketData.Name,
		addr:     addr,
		pos:      playerUpdatePacketData.Pos,
		rotation: playerUpdatePacketData.Rotation,
		inDash:   playerUpdatePacketData.InDash,
	})
}

func handleByePacket(_ netman.PacketKind, _ *net.TCPConn, data interface{}, server interface{}) {
	byePacketData := data.(netman.ByePacketData)
	serverData := server.(*Server)

	log.Println("Bye:", byePacketData.ClientId)
	serverData.players.Delete(byePacketData.ClientId)

	netman.BroadcastReliable(netman.ByeAck, netman.ByeAckPacketData{
		ClientId: byePacketData.ClientId,
	})
}

func handleFirePacket(_ netman.PacketKind, _ net.Addr, data interface{}, server interface{}) {
	firePacketData := data.(netman.FirePacketData)
	serverData := server.(*Server)

	serverData.balls.Range(func(key, value any) bool {
		ball := value.(*RemoteBall)
		if ball.ownerId == firePacketData.ClientId {
			value, ok := serverData.players.Load(firePacketData.ClientId)
			if !ok {
				log.Fatalf("Couldn't find player with given client id: %d from fire packet data\n", firePacketData.ClientId)
			}
			remotePlayer := value.(*RemotePlayer)
			newX := remotePlayer.pos.X + constants.BallOrbitRadius*math.Cos(remotePlayer.rotation)
			newY := remotePlayer.pos.Y + constants.BallOrbitRadius*math.Sin(remotePlayer.rotation)
			ball.pos.Set(newX, newY)

			ball.vel = vec.NewVec2(math.Cos(remotePlayer.rotation), math.Sin(remotePlayer.rotation)).
				Normalized().
				Muled(3) // TODO Vector builder ? (overall, better vec struct - package ?)

			ball.ownerId = constants.NoTeam
		}
		return true
	})
}
