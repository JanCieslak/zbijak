package main

import (
	"bytes"
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image/color"
	_ "image/png"
	"log"
	"net"
	"strconv"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type Player struct {
	x, y float64
}

type Game struct {
	id     uint8
	player Player
	conn   *net.UDPConn
	buffer *bytes.Buffer
}

func (g *Game) Update() error {
	g.buffer.Reset()

	speed := 3.0

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

	packets.SendPlayerUpdatePacket(g.conn, nil, g.player.x, g.player.y)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	ebitenutil.DrawRect(screen, g.player.x, g.player.y, 30, 30, color.White)
	ebitenutil.DebugPrint(screen, "Fps: "+strconv.FormatFloat(ebiten.CurrentFPS(), 'f', 2, 64))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	log.SetPrefix("Client - ")

	serverAddress, err := net.ResolveUDPAddr("udp", "127.0.0.1:8083")
	if err != nil {
		log.Fatalln("Udp address:", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddress)
	if err != nil {
		log.Fatalln("Dial creation:", err)
	}

	_, id := connect(conn)

	game := &Game{
		id: id,
		player: Player{
			x: 250,
			y: 250,
		},
		buffer: bytes.NewBuffer(make([]byte, 32)),
		conn:   conn,
	}

	log.Println("Player id assigned:", game.id)

	ebiten.SetWindowTitle("Zbijak")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(144)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalln(err)
	}

	err = conn.Close()
	if err != nil {
		log.Fatalln(err)
	}
}

func connect(conn *net.UDPConn) (*net.UDPConn, uint8) {
	packets.SendHelloPacket(conn, nil)
	welcomePacket := packets.ReceiveWelcomePacket(false, conn)

	return conn, welcomePacket.ClientId
}
