package main

import (
	"github.com/JanCieslak/Zbijak/client/player"
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/netman"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"sync"
	"time"
)

func main() {
	log.SetPrefix("Client - ")

	// TODO Uncomment when releasing
	//name, ok := inputbox.InputBox("Enter your name", "Type 3 char name", "abc")
	//if !ok {
	//	log.Fatalln("No value entered")
	//}
	//if len(name) != 3 {
	//	log.Fatalln("Name should be constructed from 3 characters")
	//}
	name := "jcs"

	netman.InitializeClientSockets("127.0.0.1:8083")

	// TODO Use reliable connection
	clientId, team := hello()

	log.Println("Client id:", clientId)

	g := &Game{
		Id:               clientId,
		Team:             team,
		Name:             name,
		Player:           player.NewPlayer(clientId, team, 250, 250), // TODO Get position from the server
		RemotePlayers:    sync.Map{},
		RemoteBalls:      sync.Map{},
		LastServerUpdate: time.Now(),
	}

	packetListener := netman.NewPacketListener(g)

	ebiten.SetWindowTitle("Zbijak")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(constants.ScreenWidth, constants.ScreenHeight)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(constants.TickRate)

	packetListener.Register(netman.ServerUpdate, handleServerUpdatePacket)
	packetListener.Register(netman.ByeAck, handleByeAckPacket)
	go packetListener.Listen()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatalln(err)
	}

	packetListener.ShutDown()
	time.Sleep(time.Millisecond * 250)
	// TODO Use reliable connection
	bye(g)
}

func hello() (uint8, constants.Team) {
	log.Println("Sending Hello packet")
	netman.SendUnreliable(netman.Hello, netman.HelloPacketData{})

	var welcomePacket netman.Packet[netman.WelcomePacketData]
	netman.ReceiveUnreliable(&welcomePacket)
	welcomePacketData := welcomePacket.Data

	return welcomePacketData.ClientId, welcomePacketData.Team
}

func bye(game *Game) {
	log.Println("Sending Bye packet")
	netman.SendUnreliable(netman.Bye, netman.ByePacketData{
		ClientId: game.Id,
	})

	var byeAckPacket netman.Packet[netman.ByeAckPacketData]
	netman.ReceiveUnreliable(&byeAckPacket)
}
