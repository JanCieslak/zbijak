package main

import (
	"github.com/JanCieslak/Zbijak/client/player"
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/netman"
	"github.com/JanCieslak/zbijak/common/vec"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"sync"
	"time"
)

func main() {
	log.SetPrefix("Client - ")

	netman.InitializeClientSockets("127.0.0.1:8083", "127.0.0.1:8084")

	clientId, team, initPos := hello()

	// TODO Uncomment when releasing
	//name, ok := inputbox.InputBox("Enter your name", "Type 3 char name", "abc")
	//if !ok {
	//	log.Fatalln("No value entered")
	//}
	//if len(name) != 3 {
	//	log.Fatalln("Name should be constructed from 3 characters")
	//}
	name := "jcs"
	game := &Game{
		Id:               clientId,
		Team:             team,
		Name:             name,
		Player:           player.NewPlayer(clientId, team, initPos.X, initPos.Y),
		RemotePlayers:    sync.Map{},
		RemoteBalls:      sync.Map{},
		LastServerUpdate: time.Now(),
	}

	netman.InitializeClientListener(game)
	netman.RegisterUDP(netman.ServerUpdate, handleServerUpdatePacket)
	netman.RegisterTCP(netman.ByeAck, handleByeAckPacket)

	log.Println("Client id:", clientId)

	ebiten.SetWindowTitle("Zbijak")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(constants.ScreenWidth, constants.ScreenHeight)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(constants.TickRate)

	go netman.ListenTCP()
	go netman.ListenUDP()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalln(err)
	}

	netman.ShutDown()
	time.Sleep(time.Millisecond * 250)
	// TODO Use reliable connection
	bye(game)
}

func hello() (uint8, constants.Team, vec.Vec2) {
	log.Println("Sending Hello packet")
	netman.SendReliable(netman.Hello, netman.HelloPacketData{})

	var welcomePacket netman.Packet[netman.WelcomePacketData]
	netman.ReceiveReliable(&welcomePacket)
	welcomePacketData := welcomePacket.Data

	return welcomePacketData.ClientId, welcomePacketData.Team, welcomePacketData.InitPos
}

func bye(game *Game) {
	log.Println("Sending Bye packet")
	netman.SendReliable(netman.Bye, netman.ByePacketData{
		ClientId: game.Id,
	})

	var byeAckPacket netman.Packet[netman.ByeAckPacketData]
	netman.ReceiveReliable(&byeAckPacket)
}
