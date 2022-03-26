package main

import (
	"fmt"
	"github.com/JanCieslak/Zbijak/client/game"
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"net"
	"sync"
	"time"
)

func main() {
	log.SetPrefix("Client - ")

	//name, ok := inputbox.InputBox("Enter your name", "Type 3 char name", "abc")
	//if !ok {
	//	log.Fatalln("No value entered")
	//}
	//if len(name) != 3 {
	//	log.Fatalln("Name should be constructed from 3 characters")
	//}
	name := "jcs"

	serverAddress, err := net.ResolveUDPAddr("udp", "127.0.0.1:8083")
	if err != nil {
		log.Fatalln("Udp address:", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddress)
	if err != nil {
		log.Fatalln("Dial creation:", err)
	}

	// TODO Use reliable connection
	clientId := game.Hello(conn)

	fmt.Println("Client id", clientId)

	g := &game.Game{
		Id:               clientId,
		Name:             name,
		Player:           game.NewPlayer(250, 250), // TODO Get from the server ?
		Conn:             conn,
		RemotePlayers:    sync.Map{},
		RemoteBalls:      sync.Map{},
		LastServerUpdate: time.Now(),
	}

	g.PacketListener = packets.NewPacketListener(g)

	ebiten.SetWindowTitle("Zbijak")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(constants.ScreenWidth, constants.ScreenHeight)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(constants.TickRate)

	g.RegisterCallbacks()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatalln(err)
	}

	g.ShutDown()
	// TODO find better way of waiting
	time.Sleep(time.Millisecond * 250)
	// TODO Use reliable connection
	game.Bye(g)
}
