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
	ScreenWidth         = 640
	ScreenHeight        = 480
	TickRate            = 144
	InterpolationOffset = 100
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
	serverUpdates    []packets.ServerUpdatePacketData
}

func (g *Game) Update() error {
	g.player.Update()

	packets.Send(g.conn, packets.PlayerUpdate, packets.PlayerUpdatePacketData{
		ClientId: g.id,
		Pos:      g.player.Pos,
		InDash:   reflect.TypeOf(g.player.State) == reflect.TypeOf(DashState{}),
	})

	renderTime := time.Now().Add(-InterpolationOffset * time.Millisecond)
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
	// TODO Draw based on state ? (trail when in dash, don't draw when dead state, when charging draw charge bar)
	ebitenutil.DrawRect(screen, g.player.Pos.X, g.player.Pos.Y, 30, 30, color.White)

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
	return ScreenWidth, ScreenHeight
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

	// TODO Use reliable connection
	clientId := hello(conn)

	fmt.Println("Client id", clientId)

	game := &Game{
		id: clientId,
		player: &Player{
			Pos:   vector.Vec2{X: 250, Y: 250},
			State: NormalState{},
		},
		conn:             conn,
		remotePlayers:    sync.Map{},
		lastServerUpdate: time.Now(),
	}

	ebiten.SetWindowTitle("Zbijak")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(TickRate)

	packetListener := packets.NewPacketListener(game)
	packetListener.Register(packets.ServerUpdate, handleServerUpdatePacket)
	packetListener.Register(packets.ByeAck, handleByeAckPacket)
	go packetListener.Listen(game.conn)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalln(err)
	}

	packetListener.ShutDown()
	// TODO find better way of waiting
	time.Sleep(time.Millisecond * 250)
	// TODO Use reliable connection
	bye(game)
}

func handleServerUpdatePacket(kind packets.PacketKind, addr net.Addr, data interface{}, game interface{}) {
	serverUpdateData := data.(packets.ServerUpdatePacketData)
	gameData := game.(*Game)

	if gameData.lastServerUpdate.Before(serverUpdateData.Timestamp) {
		gameData.lastServerUpdate = serverUpdateData.Timestamp
		gameData.serverUpdates = append(gameData.serverUpdates, serverUpdateData)

		for _, player := range serverUpdateData.PlayersData {
			_, _ = gameData.remotePlayers.LoadOrStore(player.ClientId, &RemotePlayer{
				pos:    player.Pos,
				inDash: player.InDash,
			})
		}
	}
}

func handleByeAckPacket(kind packets.PacketKind, addr net.Addr, data interface{}, game interface{}) {
	byeAckData := data.(packets.ByeAckPacketData)
	gameData := game.(*Game)
	fmt.Println("Clientid", byeAckData.ClientId)
	gameData.remotePlayers.Delete(byeAckData.ClientId)
}

func hello(conn *net.UDPConn) uint8 {
	packets.Send(conn, packets.Hello, packets.HelloPacketData{})

	var welcomePacket packets.Packet[packets.WelcomePacketData]
	packets.ReceivePacket(true, conn, &welcomePacket)
	welcomePacketData := welcomePacket.Data

	return welcomePacketData.ClientId
}

func bye(game *Game) {
	packets.Send(game.conn, packets.Bye, packets.ByePacketData{
		ClientId: game.id,
	})

	var byeAckPacket packets.Packet[packets.ByeAckPacketData]
	packets.ReceivePacket(true, game.conn, &byeAckPacket)
}
