package main

import (
	"fmt"
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/JanCieslak/zbijak/common/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/math/f64"
	"image/color"
	"io/ioutil"
	"log"
	"net"
	"reflect"
	"sync"
	"time"
)

const (
	screenWidth         = 640
	screenHeight        = 480
	tickRate            = 144
	speed               = 2.5
	interpolationOffset = 100
)

type RemotePlayer struct {
	pos    vector.Vec2
	inDash bool
}

type Game struct {
	id            uint8
	player        *Player
	conn          *net.UDPConn // TODO Abstract - NetworkManager
	remotePlayers sync.Map     // TODO Abstract - World

	lastServerUpdate time.Time
	serverUpdates    []packets.ServerUpdateData
}

func (g *Game) Update() error {
	g.player.Update()

	packets.Send(g.conn, packets.PlayerUpdate, packets.PlayerUpdateData{
		ClientId: g.id,
		Pos:      g.player.pos,
		InDash:   reflect.TypeOf(g.player.state) == reflect.TypeOf(DashState{}),
	})

	renderTime := time.Now().Add(-interpolationOffset * time.Millisecond)
	if len(g.serverUpdates) > 1 {
		for len(g.serverUpdates) > 2 && renderTime.After(g.serverUpdates[1].Timestamp) {
			g.serverUpdates = append(g.serverUpdates[:0], g.serverUpdates[1:]...)
		}

		// Interpolation
		if len(g.serverUpdates) > 2 {
			interpolationFactor := float64(renderTime.UnixMilli()-g.serverUpdates[0].Timestamp.UnixMilli()) / float64(g.serverUpdates[1].Timestamp.UnixMilli()-g.serverUpdates[0].Timestamp.UnixMilli())
			g.remotePlayers.Range(func(key, value any) bool {
				clientId := key.(uint8)
				remotePlayer := value.(*RemotePlayer)

				playerOne, ok0 := g.serverUpdates[0].PlayersData[clientId]
				playerTwo, ok1 := g.serverUpdates[1].PlayersData[clientId]

				if ok0 && ok1 {
					newX := Lerp(playerOne.Pos.X, playerTwo.Pos.X, interpolationFactor)
					newY := Lerp(playerOne.Pos.Y, playerTwo.Pos.Y, interpolationFactor)

					remotePlayer.pos.Set(newX, newY)
				}

				return true
			})
			// Extrapolation TODO Test
		} else if renderTime.After(g.serverUpdates[1].Timestamp) {
			fmt.Println("HEY") // TODO not calling
			extrapolationFactor := float64(renderTime.UnixMilli()-g.serverUpdates[0].Timestamp.UnixMilli())/float64(g.serverUpdates[1].Timestamp.UnixMilli()-g.serverUpdates[0].Timestamp.UnixMilli()) - 1.0
			g.remotePlayers.Range(func(key, value any) bool {
				clientId := key.(uint8)
				remotePlayer := value.(*RemotePlayer)

				playerOne, ok0 := g.serverUpdates[0].PlayersData[clientId]
				playerTwo, ok1 := g.serverUpdates[1].PlayersData[clientId]

				if ok0 && ok1 {
					positionDelta := f64.Vec2{playerTwo.Pos.X - playerOne.Pos.X, playerTwo.Pos.Y - playerOne.Pos.Y}
					newX := playerTwo.Pos.X + (positionDelta[0] * extrapolationFactor)
					newY := playerTwo.Pos.Y + (positionDelta[1] * extrapolationFactor)

					remotePlayer.pos.Set(newX, newY)
				}

				return true
			})
		}
	}

	return nil
}

func Lerp(start, end, p float64) float64 {
	return start + (end-start)*p
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, g.player.pos.X, g.player.pos.Y, 30, 30, color.White)

	info := fmt.Sprintf("Fps: %f Tps: %f", ebiten.CurrentFPS(), ebiten.CurrentTPS())
	ebitenutil.DebugPrint(screen, info)

	for _, update := range g.serverUpdates {
		for _, player := range update.PlayersData {
			ebitenutil.DrawLine(screen, player.Pos.X, player.Pos.Y, player.Pos.X+30, player.Pos.Y, color.RGBA{R: 255, G: 255, B: 255, A: 100})
			ebitenutil.DrawLine(screen, player.Pos.X+30, player.Pos.Y, player.Pos.X+30, player.Pos.Y+30, color.RGBA{R: 255, G: 255, B: 255, A: 100})
			ebitenutil.DrawLine(screen, player.Pos.X+30, player.Pos.Y+30, player.Pos.X, player.Pos.Y+30, color.RGBA{R: 255, G: 255, B: 255, A: 100})
			ebitenutil.DrawLine(screen, player.Pos.X, player.Pos.Y+30, player.Pos.X, player.Pos.Y, color.RGBA{R: 255, G: 255, B: 255, A: 100})
		}
	}

	g.remotePlayers.Range(func(key, value any) bool {
		clientId := key.(uint8)
		remotePlayer := value.(*RemotePlayer)
		if clientId != g.id {
			ebitenutil.DrawRect(screen, remotePlayer.pos.X, remotePlayer.pos.Y, 30, 30, color.RGBA{R: 100, G: 0, B: 0, A: 255})
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
			pos:   vector.Vec2{X: 250, Y: 250},
			state: NormalState{},
		},
		conn:             conn,
		remotePlayers:    sync.Map{},
		lastServerUpdate: time.Now(),
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

		if game.lastServerUpdate.Before(serverUpdateData.Timestamp) {
			game.lastServerUpdate = serverUpdateData.Timestamp
			game.serverUpdates = append(game.serverUpdates, serverUpdateData)

			for _, player := range serverUpdateData.PlayersData {
				_, _ = game.remotePlayers.LoadOrStore(player.ClientId, &RemotePlayer{
					pos:    player.Pos,
					inDash: player.InDash,
				})
			}
		}
	}
}
