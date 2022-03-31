package main

import (
	"fmt"
	"github.com/JanCieslak/Zbijak/client/player"
	"github.com/JanCieslak/zbijak/common"
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/netman"
	"github.com/JanCieslak/zbijak/common/vec"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/f64"
	"image/color"
	_ "image/png"
	"math"
	"reflect"
	"sync"
	"time"
)

type RemotePlayer struct {
	pos      vec.Vec2
	team     constants.Team
	name     string
	rotation float64
	inDash   bool
}

type RemoteBall struct {
	Id      uint8
	OwnerId uint8
	Pos     vec.Vec2
}

type Game struct {
	Id            uint8
	Team          constants.Team
	Name          string
	Player        *player.Player
	RemotePlayers sync.Map // TODO Abstract - World
	RemoteBalls   sync.Map

	LastServerUpdate time.Time
	serverUpdates    []netman.ServerUpdatePacketData
	PacketListener   netman.PacketListener
}

func (g *Game) Update() error {
	g.Player.Update()

	netman.SendUnreliable(netman.PlayerUpdate, netman.PlayerUpdatePacketData{
		ClientId: g.Id,
		Team:     g.Team,
		Name:     g.Name,
		Pos:      g.Player.Pos,
		Rotation: g.Player.Rotation,
		InDash:   reflect.TypeOf(g.Player.MovementState) == reflect.TypeOf(player.DashMovementState{}),
	})

	renderTime := time.Now().Add(-constants.InterpolationOffset * time.Millisecond)
	if len(g.serverUpdates) > 1 {
		for len(g.serverUpdates) > 2 && renderTime.After(g.serverUpdates[1].Timestamp) {
			g.serverUpdates = append(g.serverUpdates[:0], g.serverUpdates[1:]...)
		}

		// Interpolation
		if len(g.serverUpdates) > 2 {
			interpolationFactor := float64(renderTime.UnixMilli()-g.serverUpdates[0].Timestamp.UnixMilli()) / float64(g.serverUpdates[1].Timestamp.UnixMilli()-g.serverUpdates[0].Timestamp.UnixMilli())
			g.RemotePlayers.Range(func(key, value any) bool {
				clientId := key.(uint8)
				remotePlayer := value.(*RemotePlayer)

				playerOne, ok0 := g.serverUpdates[0].PlayersData[clientId]
				playerTwo, ok1 := g.serverUpdates[1].PlayersData[clientId]

				if ok0 && ok1 {
					newX := common.Lerp(playerOne.Pos.X, playerTwo.Pos.X, interpolationFactor)
					newY := common.Lerp(playerOne.Pos.Y, playerTwo.Pos.Y, interpolationFactor)
					newRotation := common.Lerp(playerOne.Rotation, playerTwo.Rotation, interpolationFactor)

					remotePlayer.pos.Set(newX, newY)
					remotePlayer.rotation = newRotation
				}

				return true
			})
			// Extrapolation TODO Test
		} else if renderTime.After(g.serverUpdates[1].Timestamp) {
			fmt.Println("HEY") // TODO not calling
			extrapolationFactor := float64(renderTime.UnixMilli()-g.serverUpdates[0].Timestamp.UnixMilli())/float64(g.serverUpdates[1].Timestamp.UnixMilli()-g.serverUpdates[0].Timestamp.UnixMilli()) - 1.0
			g.RemotePlayers.Range(func(key, value any) bool {
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

var (
	circleOutlineImage = loadImage("resources/circle.png", 0.2)
	circleImage        = loadImage("resources/filled_circle.png", 1.0)

	OrangeTeamColor = color.RGBA{R: 235, G: 131, B: 52, A: 255}
	BlueTeamColor   = color.RGBA{R: 52, G: 158, B: 235, A: 255}
)

func (g *Game) Draw(screen *ebiten.Image) {
	// TODO Draw based on state ? (trail when in dash, don't draw when dead state, when charging draw charge bar)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	//ebitenutil.DrawRect(screen, 0, 0, constants.ScreenWidth, constants.ScreenHeight, white)
	var teamColor color.Color
	if g.Team == constants.TeamOrange {
		teamColor = OrangeTeamColor
	} else {
		teamColor = BlueTeamColor
	}

	info := fmt.Sprintf("Fps: %f Tps: %f", ebiten.CurrentFPS(), ebiten.CurrentTPS())
	ebitenutil.DebugPrint(screen, info)

	for _, update := range g.serverUpdates {
		// TODO fix jumping ballz (this may be happening because ebiten.GetMousePosition() gets position from other windows from time to time ?)
		// TODO It's a interpolation thing (jump from almost full rotation value to 0 - it will try to interpolate whole circle)
		for _, b := range update.Balls {
			ballOwner, hasOwner := update.PlayersData[b.Owner]
			if hasOwner {
				bx := ballOwner.Pos.X + 16 - 7.5 + 40*math.Cos(ballOwner.Rotation)
				by := ballOwner.Pos.Y + 16 - 7.5 + 40*math.Sin(ballOwner.Rotation)
				drawCircleOutline(screen, bx, by, 0.5)
			} else {
				drawCircleOutline(screen, b.Pos.X, b.Pos.Y, 0.5)
			}
		}
		for _, p := range update.PlayersData {
			drawCircleOutline(screen, p.Pos.X, p.Pos.Y, 1)
		}
	}

	g.RemotePlayers.Range(func(key, value any) bool {
		clientId := key.(uint8)
		remotePlayer := value.(*RemotePlayer)
		if clientId != g.Id {
			if remotePlayer.team == constants.TeamOrange {
				drawCircle(screen, remotePlayer.pos.X, remotePlayer.pos.Y, 1, OrangeTeamColor)
			} else {
				drawCircle(screen, remotePlayer.pos.X, remotePlayer.pos.Y, 1, BlueTeamColor)
			}
			face := inconsolata.Bold8x16
			textBounds := text.BoundString(face, g.Name)
			text.Draw(screen, remotePlayer.name, face, int(remotePlayer.pos.X+constants.PlayerRadius)-textBounds.Dx()/2, int(remotePlayer.pos.Y+constants.PlayerRadius)+3, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
		return true
	})

	g.RemoteBalls.Range(func(key, value any) bool {
		remoteBall := value.(*RemoteBall)

		remotePlayer, hasOwner := g.RemotePlayers.Load(remoteBall.OwnerId)
		if hasOwner {
			ballOwner := remotePlayer.(*RemotePlayer)
			if remoteBall.OwnerId == g.Id {
				bx := g.Player.Pos.X + 16 - 8 + 40*math.Cos(g.Player.Rotation)
				by := g.Player.Pos.Y + 16 - 8 + 40*math.Sin(g.Player.Rotation)
				drawCircle(screen, bx, by, 0.5, teamColor)
			} else {
				bx := ballOwner.pos.X + 16 - 8 + 40*math.Cos(ballOwner.rotation)
				by := ballOwner.pos.Y + 16 - 8 + 40*math.Sin(ballOwner.rotation)
				if ballOwner.team == constants.TeamOrange {
					drawCircle(screen, bx, by, 0.5, OrangeTeamColor)
				} else {
					drawCircle(screen, bx, by, 0.5, BlueTeamColor)
				}
			}
		} else {
			drawCircle(screen, remoteBall.Pos.X, remoteBall.Pos.Y, 0.5, white)
		}
		return true
	})

	drawCircle(screen, g.Player.Pos.X, g.Player.Pos.Y, 1, teamColor)

	face := inconsolata.Bold8x16
	textBounds := text.BoundString(face, g.Name)
	text.Draw(screen, g.Name, face, int(g.Player.Pos.X+constants.PlayerRadius)-textBounds.Dx()/2, int(g.Player.Pos.Y+constants.PlayerRadius)+3, color.RGBA{R: 0, G: 0, B: 0, A: 255})
}

func (g *Game) Layout(_, _ int) (int, int) {
	return constants.ScreenWidth, constants.ScreenHeight
}

func (g *Game) RegisterCallbacks() {
	g.PacketListener.Register(netman.ServerUpdate, handleServerUpdatePacket)
	g.PacketListener.Register(netman.ByeAck, handleByeAckPacket)
	go g.PacketListener.Listen2()
}

func (g *Game) ShutDown() {
	g.PacketListener.ShutDown()
}

func Hello() (uint8, constants.Team) {
	netman.SendUnreliable(netman.Hello, netman.HelloPacketData{})

	var welcomePacket netman.Packet[netman.WelcomePacketData]
	netman.ReceiveUnreliable(&welcomePacket)
	welcomePacketData := welcomePacket.Data

	return welcomePacketData.ClientId, welcomePacketData.Team
}

func Bye(game *Game) {
	netman.SendUnreliable(netman.Bye, netman.ByePacketData{
		ClientId: game.Id,
	})

	var byeAckPacket netman.Packet[netman.ByeAckPacketData]
	netman.ReceiveUnreliable(&byeAckPacket)
}
