package game

import (
	"github.com/JanCieslak/zbijak/common/vec"
	"github.com/hajimehoshi/ebiten/v2"
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
	Pos           vec.Vec2
	Velocity      vec.Vec2
	MovementState State
	PlayerState   State
}

func NewPlayer(x, y float64) *Player {
	return &Player{
		Pos:      vec.Vec2{X: x, Y: y},
		Velocity: vec.Vec2{},
		MovementState: NormalMovementState{
			lastDashTime: time.Now().Add(-DashCooldown),
		},
		PlayerState: NormalPlayerState{},
	}
}

func (p *Player) Update(g *Game) {
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

	moveVector.Normalize()
	p.Velocity = moveVector

	p.MovementState.Update(g, p)
	p.PlayerState.Update(g, p)

	p.Pos.AddVec(p.Velocity)
}
