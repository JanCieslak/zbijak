package main

import (
	"encoding/json"
	"fmt"
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image/color"
	"io/ioutil"
	"log"
	"net"
	"sync"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type Player struct {
	x, y float64
}

type RemotePlayer struct {
	x, y float64
}

type Game struct {
	id            uint8
	player        *Player
	conn          *net.UDPConn
	remotePlayers sync.Map
}

func (g *Game) Update() error {
	var playerUpdatePacket packets.Packet[packets.PlayerUpdateData]
	playerUpdatePacket.Kind = packets.PlayerUpdate
	playerUpdatePacket.Data = packets.PlayerUpdateData{
		ClientId: g.id,
		X:        g.player.x,
		Y:        g.player.y,
	}
	log.Println("Sending:", playerUpdatePacket)
	packets.SendPacket(g.conn, nil, packets.Serialize(playerUpdatePacket))

	speed := 10.0

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.x -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.x += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.player.y -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.player.y += speed
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	ebitenutil.DrawRect(screen, g.player.x, g.player.y, 30, 30, color.White)

	g.remotePlayers.Range(func(key, value any) bool {
		clientId := key.(uint8)
		remotePlayer := value.(*RemotePlayer)
		if clientId != g.id {
			ebitenutil.DrawRect(screen, remotePlayer.x, remotePlayer.y, 30, 30, color.RGBA{R: 100, G: 0, B: 0, A: 255})
		}
		return true
	})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	log.SetPrefix("Client - ")
	log.SetOutput(ioutil.Discard)

	serverAddress, err := net.ResolveUDPAddr("udp", "127.0.0.1:8083")
	if err != nil {
		log.Fatalln("Udp address:", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddress)
	if err != nil {
		log.Fatalln("Dial creation:", err)
	}

	var helloPacket packets.Packet[packets.HelloPacketData]
	helloPacket.Kind = packets.Hello
	helloPacket.Data = packets.HelloPacketData{}
	packets.SendPacket(conn, nil, packets.Serialize(helloPacket))

	bytes := packets.ReceivePacket(true, conn)
	var serverUpdatePacket packets.Packet[packets.WelcomePacketData]
	err = json.Unmarshal(bytes, &serverUpdatePacket)
	if err != nil {
		log.Fatalln("Error when deserializing packet")
	}
	welcomePacket := serverUpdatePacket.Data

	log.Println("Client id", welcomePacket.ClientId)
	fmt.Println("Client id", welcomePacket.ClientId)

	game := &Game{
		id: welcomePacket.ClientId,
		player: &Player{
			x: 250,
			y: 250,
		},
		conn:          conn,
		remotePlayers: sync.Map{},
	}

	ebiten.SetWindowTitle("Zbijak")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(144)

	go func() {
		for {
			bytes := packets.ReceivePacket(true, game.conn)
			var serverUpdatePacket packets.Packet[packets.ServerUpdateData]
			err := json.Unmarshal(bytes, &serverUpdatePacket)
			if err != nil {
				log.Fatalln("Error when deserializing packet")
			}
			log.Println("Received:", serverUpdatePacket)
			serverUpdateData := serverUpdatePacket.Data

			for _, player := range serverUpdateData.PlayersData {
				game.remotePlayers.Store(player.ClientId, &RemotePlayer{
					x: player.X,
					y: player.Y,
				})
			}
		}
	}()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalln(err)
	}
}
