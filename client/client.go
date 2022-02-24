package main

import (
	"encoding/json"
	"flag"
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

type GhostPlayer struct {
	x, y float64
}

type Game struct {
	id               uint8
	player           *Player
	conn             *net.UDPConn
	remotePlayerLock sync.Mutex
	remotePlayer     *GhostPlayer
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

	g.remotePlayerLock.Lock()
	ebitenutil.DrawRect(screen, g.remotePlayer.x+100, g.remotePlayer.y, 30, 30, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	g.remotePlayerLock.Unlock()
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	id := flag.Int("id", 0, "Client id")
	flag.Parse()

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

	log.Println("Client id", *id)

	game := &Game{
		id: uint8(*id),
		player: &Player{
			x: 250,
			y: 250,
		},
		conn:             conn,
		remotePlayerLock: sync.Mutex{},
		remotePlayer: &GhostPlayer{
			x: 0,
			y: 0,
		},
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

			game.remotePlayerLock.Lock()
			game.remotePlayer.x = serverUpdateData.PlayersData[0].X
			game.remotePlayer.y = serverUpdateData.PlayersData[0].Y
			game.remotePlayerLock.Unlock()
		}
	}()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalln(err)
	}
}
