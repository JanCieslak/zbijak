package main

import (
	"fmt"
	"github.com/JanCieslak/Zbijak/client/player"
	"github.com/JanCieslak/Zbijak/client/utils"
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/netman"
	"github.com/JanCieslak/zbijak/common/vec"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image/color"
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
	RemotePlayers sync.Map // TODO Abstract - World struct ?
	RemoteBalls   sync.Map

	LastServerUpdate time.Time
	serverUpdates    []netman.ServerUpdatePacketData
}

func (g *Game) Update() error {
	g.Player.Update()

	// TODO Player update should be sending as little data as possible
	// Probably ideally:
	// - id
	// - mouse position and mouse + keyboard inputs
	// - rotation (?)
	// The rest should be tracked and evaluated by server (single source of truth)
	netman.SendUnreliable(netman.PlayerUpdate, netman.PlayerUpdatePacketData{
		ClientId: g.Id,
		Team:     g.Team,
		Name:     g.Name,
		Pos:      g.Player.Pos,
		Rotation: g.Player.Rotation,
		InDash:   reflect.TypeOf(g.Player.MovementState) == reflect.TypeOf(player.DashMovementState{}),
	})

	g.InterpolateRemoteObjectsPositions()

	return nil
}

func (g *Game) InterpolateRemoteObjectsPositions() {
	renderTime := time.Now().Add(-constants.InterpolationOffset)

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
					newX := utils.Lerp(playerOne.Pos.X, playerTwo.Pos.X, interpolationFactor)
					newY := utils.Lerp(playerOne.Pos.Y, playerTwo.Pos.Y, interpolationFactor)
					newRotation := utils.Slerp(playerOne.Rotation, playerTwo.Rotation, interpolationFactor)

					remotePlayer.pos.Set(newX, newY)
					remotePlayer.rotation = newRotation
				}

				return true
			})
		}
		// TODO Extrapolation (should be working, but it's not. It's not necessary, at least for now. The game looks and feels smooth)
		//} else if renderTime.After(g.serverUpdates[1].Timestamp) {
		//	fmt.Println("HEY") // TODO not calling
		//	extrapolationFactor := float64(renderTime.UnixMilli()-g.serverUpdates[0].Timestamp.UnixMilli())/float64(g.serverUpdates[1].Timestamp.UnixMilli()-g.serverUpdates[0].Timestamp.UnixMilli()) - 1.0
		//	g.RemotePlayers.Range(func(key, value any) bool {
		//		clientId := key.(uint8)
		//		remotePlayer := value.(*RemotePlayer)
		//
		//		playerOne, ok0 := g.serverUpdates[0].PlayersData[clientId]
		//		playerTwo, ok1 := g.serverUpdates[1].PlayersData[clientId]
		//
		//		if ok0 && ok1 {
		//			positionDelta := f64.Vec2{playerTwo.Pos.X - playerOne.Pos.X, playerTwo.Pos.Y - playerOne.Pos.Y}
		//			newX := playerTwo.Pos.X + (positionDelta[0] * extrapolationFactor)
		//			newY := playerTwo.Pos.Y + (positionDelta[1] * extrapolationFactor)
		//
		//			remotePlayer.pos.Set(newX, newY)
		//		}
		//
		//		return true
		//	})
		//}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// TODO Draw based on state ? (trail when in dash, don't draw when dead state, when charging draw charge bar)

	g.DebugDraw(screen)

	DrawNoGoZone(screen)
	DrawDivider(screen)
	DrawTeamNames(screen)
	DrawScore(screen)

	g.RemotePlayers.Range(func(key, value any) bool {
		clientId := key.(uint8)
		remotePlayer := value.(*RemotePlayer)
		if clientId != g.Id {
			utils.DrawCircle(screen, remotePlayer.pos.X, remotePlayer.pos.Y, constants.PlayerRadius, 1, utils.GetTeamColor(remotePlayer.team))
			utils.DrawText(screen, "jcs", remotePlayer.pos.X, remotePlayer.pos.Y+constants.PlayerRadius*3/4)
		}
		return true
	})

	g.RemoteBalls.Range(func(key, value any) bool {
		remoteBall := value.(*RemoteBall)

		remotePlayer, hasOwner := g.RemotePlayers.Load(remoteBall.OwnerId)
		if hasOwner {
			ballOwner := remotePlayer.(*RemotePlayer)
			// TODO Not The same ?
			if remoteBall.OwnerId == g.Id {
				bx := g.Player.Pos.X + constants.BallOrbitRadius*math.Cos(g.Player.Rotation)
				by := g.Player.Pos.Y + constants.BallOrbitRadius*math.Sin(g.Player.Rotation)
				utils.DrawCircle(screen, bx, by, constants.BallRadius, 0.5, utils.GetTeamColor(g.Team))
			} else {
				bx := ballOwner.pos.X + constants.BallOrbitRadius*math.Cos(ballOwner.rotation)
				by := ballOwner.pos.Y + constants.BallOrbitRadius*math.Sin(ballOwner.rotation)
				utils.DrawCircle(screen, bx, by, constants.BallRadius, 0.5, utils.GetTeamColor(ballOwner.team))
			}
		} else {
			utils.DrawCircle(screen, remoteBall.Pos.X, remoteBall.Pos.Y, constants.BallRadius, 0.5, color.White)
		}
		return true
	})

	g.Player.Draw(screen)
}

func DrawNoGoZone(screen *ebiten.Image) {
	ebitenutil.DrawLine(screen, constants.NoGoZonePadding, constants.NoGoZonePadding, constants.ScreenWidth-constants.NoGoZonePadding, constants.NoGoZonePadding, color.White)
	ebitenutil.DrawLine(screen, constants.NoGoZonePadding, constants.ScreenHeight-constants.NoGoZonePadding, constants.ScreenWidth-constants.NoGoZonePadding, constants.ScreenHeight-constants.NoGoZonePadding, color.White)
	ebitenutil.DrawLine(screen, constants.NoGoZonePadding, constants.NoGoZonePadding, constants.NoGoZonePadding, constants.ScreenHeight-constants.NoGoZonePadding, color.White)
	ebitenutil.DrawLine(screen, constants.ScreenWidth-constants.NoGoZonePadding, constants.NoGoZonePadding, constants.ScreenWidth-constants.NoGoZonePadding, constants.ScreenHeight-constants.NoGoZonePadding, color.White)
}

func DrawDivider(screen *ebiten.Image) {
	ebitenutil.DrawLine(screen, constants.ScreenWidth/2, 0, constants.ScreenWidth/2, constants.ScreenHeight, color.White)
}

func DrawTeamNames(screen *ebiten.Image) {
	utils.DrawText(screen, "Team A", constants.ScreenWidth/4, constants.NoGoZonePadding)
	utils.DrawText(screen, "Team B", constants.ScreenWidth*3/4, constants.NoGoZonePadding)
}

func DrawScore(screen *ebiten.Image) {
	utils.DrawText(screen, "15", constants.ScreenWidth/2-20, constants.NoGoZonePadding)
	utils.DrawText(screen, "19", constants.ScreenWidth/2+20, constants.NoGoZonePadding)
}

func (g *Game) DebugDraw(screen *ebiten.Image) {
	info := fmt.Sprintf("Fps: %f Tps: %f", ebiten.CurrentFPS(), ebiten.CurrentTPS())
	ebitenutil.DebugPrint(screen, info)

	for _, update := range g.serverUpdates {
		// TODO fix jumping ballz (this may be happening because ebiten.GetMousePosition() gets position from other windows from time to time ?)
		// TODO It's a interpolation thing (jump from almost full rotation value to 0 - it will try to interpolate whole circle)
		// Draw ball's buffered positions
		for _, b := range update.Balls {
			ballOwner, hasOwner := update.PlayersData[b.Owner]
			if hasOwner {
				bx := ballOwner.Pos.X + constants.BallOrbitRadius*math.Cos(ballOwner.Rotation)
				by := ballOwner.Pos.Y + constants.BallOrbitRadius*math.Sin(ballOwner.Rotation)
				utils.DrawCircleOutline(screen, bx, by, constants.BallRadius, 0.5)
			} else {
				utils.DrawCircleOutline(screen, b.Pos.X, b.Pos.Y, constants.BallRadius, 0.5)
			}
		}

		// Draw player's buffered positions
		for _, p := range update.PlayersData {
			utils.DrawCircleOutline(screen, p.Pos.X, p.Pos.Y, constants.PlayerRadius, 1)
		}
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return constants.ScreenWidth, constants.ScreenHeight
}
