package game

import (
	"bufio"
	"fmt"
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/packets"
	"github.com/JanCieslak/zbijak/common/vec"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/f64"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"net"
	"os"
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

type Game struct {
	Id            uint8
	Team          constants.Team
	Name          string
	Player        *Player
	Conn          *net.UDPConn // TODO Abstract - NetworkManager
	RemotePlayers sync.Map     // TODO Abstract - World
	RemoteBalls   sync.Map

	LastServerUpdate time.Time
	serverUpdates    []packets.ServerUpdatePacketData
	PacketListener   packets.PacketListener
}

func (g *Game) Update() error {
	g.Player.Update(g)

	packets.Send(g.Conn, packets.PlayerUpdate, packets.PlayerUpdatePacketData{
		ClientId: g.Id,
		Team:     g.Team,
		Name:     g.Name,
		Pos:      g.Player.Pos,
		Rotation: g.Player.Rotation,
		InDash:   reflect.TypeOf(g.Player.MovementState) == reflect.TypeOf(DashMovementState{}),
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
					newX := packets.Lerp(playerOne.Pos.X, playerTwo.Pos.X, interpolationFactor)
					newY := packets.Lerp(playerOne.Pos.Y, playerTwo.Pos.Y, interpolationFactor)
					newRotation := packets.Lerp(playerOne.Rotation, playerTwo.Rotation, interpolationFactor)

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
	circleOutlineImage *ebiten.Image
	circleImage        *ebiten.Image
)

func init() {
	circleOutlineImage = loadImage("resources/circle.png", 0.2)
	circleImage = loadImage("resources/filled_circle.png", 1.0)
}

func loadImage(path string, alpha float64) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(bufio.NewReader(f))
	if err != nil {
		log.Fatal(err)
	}
	origEbitenImage := ebiten.NewImageFromImage(img)

	w, h := origEbitenImage.Size()
	ebitenImage := ebiten.NewImage(w, h)

	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(1, 1, 1, alpha)
	ebitenImage.DrawImage(origEbitenImage, op)

	return ebitenImage
}

var (
	OrangeTeamColor = color.RGBA{R: 235, G: 131, B: 52, A: 255}
	BlueTeamColor   = color.RGBA{R: 52, G: 158, B: 235, A: 255}
)

func (g *Game) Draw(screen *ebiten.Image) {
	// TODO Draw based on state ? (trail when in dash, don't draw when dead state, when charging draw charge bar)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
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
		remoteBall := value.(*Ball)

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

func drawCircle(screen *ebiten.Image, x, y, s float64, c color.Color) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(x, y)
	r, g, b, _ := c.RGBA()
	rf := mapValue(float64(r), 0, 0xffff, 0, 1)
	gf := mapValue(float64(g), 0, 0xffff, 0, 1)
	bf := mapValue(float64(b), 0, 0xffff, 0, 1)
	op.ColorM.Scale(rf, gf, bf, 1)
	screen.DrawImage(circleImage, &op)
}

func mapValue(x, inMin, inMax, outMin, outMax float64) float64 {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}

func drawCircleOutline(screen *ebiten.Image, x, y, s float64) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(x, y)
	screen.DrawImage(circleOutlineImage, &op)
}

func drawRectOutline(screen *ebiten.Image, x, y, w, h float64) {
	ebitenutil.DrawLine(screen, x, y, x+w, y, color.RGBA{R: 255, G: 255, B: 255, A: 100})
	ebitenutil.DrawLine(screen, x+w, y, x+w, y+h, color.RGBA{R: 255, G: 255, B: 255, A: 100})
	ebitenutil.DrawLine(screen, x+w, y+h, x, y+h, color.RGBA{R: 255, G: 255, B: 255, A: 100})
	ebitenutil.DrawLine(screen, x, y+h, x, y, color.RGBA{R: 255, G: 255, B: 255, A: 100})
}

func (g *Game) Layout(_, _ int) (int, int) {
	return constants.ScreenWidth, constants.ScreenHeight
}

func (g *Game) RegisterCallbacks() {
	g.PacketListener.Register(packets.ServerUpdate, handleServerUpdatePacket)
	g.PacketListener.Register(packets.ByeAck, handleByeAckPacket)
	go g.PacketListener.Listen(g.Conn)
}

func (g *Game) ShutDown() {
	g.PacketListener.ShutDown()
}

func handleServerUpdatePacket(_ packets.PacketKind, _ net.Addr, data interface{}, game interface{}) {
	serverUpdateData := data.(packets.ServerUpdatePacketData)
	gameData := game.(*Game)

	if gameData.LastServerUpdate.Before(serverUpdateData.Timestamp) {
		gameData.LastServerUpdate = serverUpdateData.Timestamp
		gameData.serverUpdates = append(gameData.serverUpdates, serverUpdateData)

		for _, b := range serverUpdateData.Balls {
			gameData.RemoteBalls.Store(b.Id, &Ball{
				Id:      b.Id,
				OwnerId: b.Owner,
				Pos:     b.Pos,
			})
		}

		for _, p := range serverUpdateData.PlayersData {
			_, _ = gameData.RemotePlayers.LoadOrStore(p.ClientId, &RemotePlayer{
				pos:      p.Pos,
				team:     p.Team,
				name:     p.Name,
				rotation: p.Rotation,
				inDash:   p.InDash,
			})
		}
	}
}

func handleByeAckPacket(_ packets.PacketKind, _ net.Addr, data interface{}, game interface{}) {
	byeAckData := data.(packets.ByeAckPacketData)
	gameData := game.(*Game)
	fmt.Println("ClientId", byeAckData.ClientId)
	gameData.RemotePlayers.Delete(byeAckData.ClientId)
}

func Hello(conn *net.UDPConn) (uint8, constants.Team) {
	packets.Send(conn, packets.Hello, packets.HelloPacketData{})

	var welcomePacket packets.Packet[packets.WelcomePacketData]
	packets.ReceivePacket(true, conn, &welcomePacket)
	welcomePacketData := welcomePacket.Data

	return welcomePacketData.ClientId, welcomePacketData.Team
}

func Bye(game *Game) {
	packets.Send(game.Conn, packets.Bye, packets.ByePacketData{
		ClientId: game.Id,
	})

	var byeAckPacket packets.Packet[packets.ByeAckPacketData]
	packets.ReceivePacket(true, game.Conn, &byeAckPacket)
}
