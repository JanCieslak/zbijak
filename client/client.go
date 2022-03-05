package main

import (
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/math/f64"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

const (
	screenWidth  = 640
	screenHeight = 480
	tickRate     = 144
	speed        = 2.5
	dashSpeed    = 2 * speed
	dashDuration = 250 * time.Millisecond
	dashCooldown = time.Second
)

var (
	interpolationTicks = 144 / constants.ServerTickRate
)

type Player struct {
	x, y      float64
	dashVec   f64.Vec2
	inDash    bool
	startDash time.Time
	endDash   time.Time
}

type RemotePlayer struct {
	x, y, targetX, targetY float64
	inDash                 bool
}

type Game struct {
	id            uint8
	player        *Player
	conn          *net.UDPConn
	remotePlayers sync.Map
}

func (g *Game) Update() error {
	moveVector := f64.Vec2{0, 0}

	if !g.player.inDash {
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			moveVector[0] -= 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			moveVector[0] += 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
			moveVector[1] -= 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
			moveVector[1] += 1
		}
	}
	if !g.player.inDash && time.Since(g.player.endDash) > dashCooldown && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.player.startDash = time.Now()
		normalize(&moveVector)
		multiply(&moveVector, dashSpeed)
		g.player.dashVec = moveVector
		g.player.inDash = true
	}

	if g.player.inDash {
		if time.Since(g.player.startDash) > dashDuration {
			g.player.inDash = false
			g.player.endDash = time.Now()
		}

		g.player.x += g.player.dashVec[0]
		g.player.y += g.player.dashVec[1]
	} else {
		normalize(&moveVector)
		multiply(&moveVector, speed)

		g.player.x += moveVector[0]
		g.player.y += moveVector[1]
	}

	packets.Send(g.conn, packets.PlayerUpdate, packets.PlayerUpdateData{
		ClientId: g.id,
		X:        g.player.x,
		Y:        g.player.y,
		InDash:   g.player.inDash,
	})

	g.remotePlayers.Range(func(key, value any) bool {
		remotePlayer := value.(*RemotePlayer)
		moveVector := f64.Vec2{remotePlayer.targetX - remotePlayer.x, remotePlayer.targetY - remotePlayer.y}
		normalize(&moveVector)

		if remotePlayer.inDash {
			multiply(&moveVector, dashSpeed)
		} else {
			multiply(&moveVector, speed)
		}

		if math.Abs(remotePlayer.targetX-remotePlayer.x) > 5 || math.Abs(remotePlayer.targetY-remotePlayer.y) > 5 { // TODO 5??? (fix diagonal interpolation - missing frame ?)
			remotePlayer.x += moveVector[0]
			remotePlayer.y += moveVector[1]
		}

		return true
	})

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, g.player.x, g.player.y, 30, 30, color.White)

	g.remotePlayers.Range(func(key, value any) bool {
		clientId := key.(uint8)
		remotePlayer := value.(*RemotePlayer)
		if clientId != g.id {
			//ebitenutil.DrawLine(screen, remotePlayer.targetX, remotePlayer.targetY, remotePlayer.targetX+30, remotePlayer.targetY, color.RGBA{R: 255, G: 255, B: 255, A: 100})
			//ebitenutil.DrawLine(screen, remotePlayer.targetX+30, remotePlayer.targetY, remotePlayer.targetX+30, remotePlayer.targetY+30, color.RGBA{R: 255, G: 255, B: 255, A: 100})
			//ebitenutil.DrawLine(screen, remotePlayer.targetX+30, remotePlayer.targetY+30, remotePlayer.targetX, remotePlayer.targetY+30, color.RGBA{R: 255, G: 255, B: 255, A: 100})
			//ebitenutil.DrawLine(screen, remotePlayer.targetX, remotePlayer.targetY+30, remotePlayer.targetX, remotePlayer.targetY, color.RGBA{R: 255, G: 255, B: 255, A: 100})
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

	packets.Send(conn, packets.Hello, packets.HelloPacketData{})

	var welcomePacket packets.Packet[packets.WelcomePacketData]
	packets.ReceivePacket(true, conn, &welcomePacket)
	welcomePacketData := welcomePacket.Data

	log.Println("Client id", welcomePacketData.ClientId)

	game := &Game{
		id: welcomePacketData.ClientId,
		player: &Player{
			x:         250,
			y:         250,
			dashVec:   f64.Vec2{},
			inDash:    false,
			startDash: time.Now(),
			endDash:   time.Now(),
		},
		conn:          conn,
		remotePlayers: sync.Map{},
	}

	ebiten.SetWindowTitle("Zbijak")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(tickRate)

	go listen(game)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalln(err)
	}
}

func listen(game *Game) {
	for {
		var serverUpdatePacket packets.Packet[packets.ServerUpdateData]
		packets.ReceivePacket(true, game.conn, &serverUpdatePacket)
		serverUpdateData := serverUpdatePacket.Data

		for _, player := range serverUpdateData.PlayersData {
			value, present := game.remotePlayers.Load(player.ClientId)
			if present {
				remotePlayer := value.(*RemotePlayer)
				game.remotePlayers.Store(player.ClientId, &RemotePlayer{
					x:       remotePlayer.x,
					y:       remotePlayer.y,
					targetX: player.X,
					targetY: player.Y,
					inDash:  player.InDash,
				})
			} else {
				game.remotePlayers.Store(player.ClientId, &RemotePlayer{
					x:       player.X,
					y:       player.Y,
					targetX: player.X,
					targetY: player.Y,
					inDash:  player.InDash,
				})
			}
		}
	}
}

func normalize(vector *f64.Vec2) {
	length := math.Sqrt(math.Pow(vector[0], 2) + math.Pow(vector[1], 2))
	if length != 0 {
		vector[0] /= length
		vector[1] /= length
	}
}

func multiply(vector *f64.Vec2, mul float64) {
	vector[0] *= mul
	vector[1] *= mul
}
