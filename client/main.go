package main

import (
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image/color"
	_ "image/png"
	"log"
	"net"
	"strconv"
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
	id          uint8
	player      Player
	conn        *net.UDPConn
	playersLock sync.Mutex
	players     map[uint8]*GhostPlayer
}

func (g *Game) Update() error {
	g.playersLock.Lock()
	ghostPlayers := packets.ReceiveServerUpdatePacket(g.conn)
	log.Println("GHOST PLAYERS", ghostPlayers)
	for i, ghostPlayer := range ghostPlayers {
		log.Println(i, "GHOST PLAYER", ghostPlayer)
		if g.players[ghostPlayer.ClientId] == nil {
			g.players[ghostPlayer.ClientId] = &GhostPlayer{
				x: ghostPlayer.X,
				y: ghostPlayer.Y,
			}
			continue
		}
		player := g.players[ghostPlayer.ClientId]
		player.x = ghostPlayer.X
		player.y = ghostPlayer.Y
	}
	g.playersLock.Unlock()

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

	packets.SendPlayerUpdatePacket(g.conn, nil, g.id, g.player.x, g.player.y)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	ebitenutil.DrawRect(screen, g.player.x, g.player.y, 30, 30, color.White)
	log.Println("Drawing player at", g.player.x, g.player.y)
	g.playersLock.Lock()
	for i, player := range g.players {
		log.Println("Drawing ghost player ", i, "at", player.x, player.y)
		ebitenutil.DrawRect(screen, player.x, player.y, 30, 30, color.White)
	}
	g.playersLock.Unlock()
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
		conn:        conn,
		players:     make(map[uint8]*GhostPlayer),
		playersLock: sync.Mutex{},
	}

	log.Println("Player id assigned:", game.id)

	ebiten.SetWindowTitle("Zbijak")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(1)

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
