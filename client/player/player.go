package player

import (
	"github.com/JanCieslak/Zbijak/client/utils"
	"github.com/JanCieslak/zbijak/common/constants"
	"github.com/JanCieslak/zbijak/common/vec"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
	"time"
)

const (
	NormalSpeed = 2.5

	FullChargeSpeed    = 0.2 * NormalSpeed
	FullChargeDuration = time.Second

	DashSpeed    = 2 * NormalSpeed
	DashDuration = 250 * time.Millisecond
	DashCooldown = time.Second
)

type Player struct {
	Id            uint8
	Team          constants.Team
	Pos           vec.Vec2
	Velocity      vec.Vec2
	MovementState State
	PlayerState   State
	Rotation      float64
}

func NewPlayer(id uint8, team constants.Team, x, y float64) *Player {
	return &Player{
		Id:       id,
		Team:     team,
		Pos:      vec.Vec2{X: x, Y: y},
		Velocity: vec.Vec2{},
		MovementState: NormalMovementState{
			lastDashTime: time.Now().Add(-DashCooldown),
		},
		PlayerState: NormalPlayerState{},
	}
}

func (p *Player) Update() {
	moveVector := vec.Vec2{}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		moveVector.Add(-1, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		moveVector.Add(1, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		moveVector.Add(0, -1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		moveVector.Add(0, 1)
	}

	mx, my := ebiten.CursorPosition()
	cursorVec := vec.NewIVec2(mx, my)
	if cursorVec.Y > p.Pos.Y {
		cursorVec.SubVec(p.Pos)
		p.Rotation = math.Acos(vec.Right.Dot(cursorVec) / (vec.Right.Len() * cursorVec.Len()))
	} else {
		cursorVec.SubVec(p.Pos)
		p.Rotation = math.Pi + math.Acos(-vec.Right.Dot(cursorVec)/(vec.Right.Len()*cursorVec.Len()))
	}

	moveVector.Normalize()
	p.Velocity = moveVector

	p.MovementState.Update(p)
	p.PlayerState.Update(p)

	p.Pos.AddVec(p.Velocity)

	// Wall collisions
	if p.Pos.X-constants.PlayerRadius <= constants.NoGoZonePadding {
		p.Pos.X = constants.PlayerRadius + constants.NoGoZonePadding
	}
	if p.Pos.X+constants.PlayerRadius >= constants.ScreenWidth-constants.NoGoZonePadding {
		p.Pos.X = constants.ScreenWidth - constants.PlayerRadius - constants.NoGoZonePadding
	}
	if p.Pos.Y-constants.PlayerRadius <= constants.NoGoZonePadding {
		p.Pos.Y = constants.PlayerRadius + constants.NoGoZonePadding
	}
	if p.Pos.Y+constants.PlayerRadius >= constants.ScreenHeight-constants.NoGoZonePadding {
		p.Pos.Y = constants.ScreenHeight - constants.PlayerRadius - constants.NoGoZonePadding
	}

	// TODO ball wall conditions (now one can throw ball outside of the screen) ;/
}

func (p *Player) Draw(screen *ebiten.Image) {
	utils.DrawCircle(screen, p.Pos.X, p.Pos.Y, constants.PlayerRadius, 1, utils.GetTeamColor(p.Team))
	utils.DrawText(screen, "jcs", p.Pos.X, p.Pos.Y+constants.PlayerRadius*3/4) // TODO name
}
