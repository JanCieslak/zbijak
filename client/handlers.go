package main

import (
	"fmt"
	"github.com/JanCieslak/zbijak/common/netman"
	"net"
)

func handleServerUpdatePacket(_ netman.PacketKind, _ net.Addr, data interface{}, game interface{}) {
	serverUpdateData := data.(netman.ServerUpdatePacketData)
	gameData := game.(*Game)

	if gameData.LastServerUpdate.Before(serverUpdateData.Timestamp) {
		gameData.LastServerUpdate = serverUpdateData.Timestamp
		gameData.serverUpdates = append(gameData.serverUpdates, serverUpdateData)

		for _, b := range serverUpdateData.Balls {
			gameData.RemoteBalls.Store(b.Id, &RemoteBall{
				Id:      b.Id,
				OwnerId: b.Owner,
				Pos:     b.Pos,
			})
		}

		for _, p := range serverUpdateData.PlayersData {
			_, _ = gameData.RemotePlayers.LoadOrStore(p.ClientId, &RemotePlayer{
				pos:      p.Pos,
				team:     p.Team,
				name:     p.Name,
				rotation: p.Rotation,
				inDash:   p.InDash,
			})
		}
	}
}

func handleByeAckPacket(_ netman.PacketKind, _ net.Addr, data interface{}, game interface{}) {
	byeAckData := data.(netman.ByeAckPacketData)
	gameData := game.(*Game)
	fmt.Println("ClientId", byeAckData.ClientId)
	gameData.RemotePlayers.Delete(byeAckData.ClientId)
}
